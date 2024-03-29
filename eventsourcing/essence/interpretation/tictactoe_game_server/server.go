package tictactoe_game_server

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"time"
)

func init() {
	schema.SetDefaultUnionTypeFormatter(schema.FormatUnionNameUsingTypeName)
}

func UnmarshalQueryOrCommand(
	data []byte,
	onCommand func(tictactoemanage.Command) error,
	onQuery func(tictactoemanage.Query) error,
	onSubscription func(tictactoemanage.Subscription) error,
) error {
	sch, err := schema.FromJSON(data)
	if err != nil {
		return fmt.Errorf("JsonUnmarshal: %s", err)
	}

	goo := schema.MustToGo(sch)

	switch x := goo.(type) {
	case tictactoemanage.Command:
		return onCommand(x)
	case tictactoemanage.Query:
		return onQuery(x)
	case tictactoemanage.Subscription:
		return onSubscription(x)
	}

	return fmt.Errorf("JsonUnmarshal: %T not a command or query", goo)
}

func MarshalState(state tictactoemanage.State) ([]byte, error) {
	result := schema.FromGo(state)
	return schema.ToJSON(result)
}

func ExtractSessionID(cmd tictactoemanage.Command) string {
	return tictactoemanage.MustMatchCommand(
		cmd,
		func(x *tictactoemanage.CreateSessionCMD) string {
			return x.SessionID
		},
		func(x *tictactoemanage.JoinGameSessionCMD) string {
			return x.SessionID
		},
		func(x *tictactoemanage.GameSessionWithBotCMD) string {
			return x.SessionID
		},
		func(x *tictactoemanage.LeaveGameSessionCMD) string {
			return x.SessionID
		},
		func(x *tictactoemanage.NewGameCMD) string {
			return x.SessionID
		},
		func(x *tictactoemanage.GameActionCMD) string {
			return x.SessionID
		},
		func(x *tictactoemanage.SequenceCMD) string {
			for _, cmd := range x.Commands {
				return ExtractSessionID(cmd)
			}

			// TODO convert to error
			return "SequenceCMD(no session id)"
		},
	)
}

type Repository[A any] interface {
	GetAs(key string, x *A) error
	UpdateRecords(s schemaless.UpdateRecords[any]) error
}

type SessionWithGame struct {
	SessionID     string
	CurrentGameID string
}

func NewGame(b websockproto.Broadcaster, r schemaless.Repository[tictactoemanage.State], q Query, liveSelect *LiveSelect) *Game {
	return &Game{
		broadcast:           b,
		gameStateRepository: r,
		query:               q,
		liveSelect:          liveSelect,
	}
}

type Query interface {
	Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error)
}

type LiveSelectI interface {
	Process(ctx context.Context, sessionID string) error
}

type Game struct {
	broadcast           websockproto.Broadcaster
	gameStateRepository schemaless.Repository[tictactoemanage.State]
	query               Query
	liveSelect          LiveSelectI
}

func (g *Game) OnMessage(connectionID string, data []byte) error {
	return UnmarshalQueryOrCommand(
		data,
		func(cmd tictactoemanage.Command) error {
			log.Printf("OnMessage: command %#v \n", cmd)
			sessionID := ExtractSessionID(cmd)
			g.broadcast.AssociateConnectionWithSession(connectionID, sessionID)

			stateRecord, err := g.gameStateRepository.Get(sessionID, "session")
			if err != nil && !errors.Is(err, schemaless.ErrNotFound) {
				log.Errorln("OnMessage: Get: err", err)
				return err
			}

			machine := tictactoemanage.NewMachineWithState(stateRecord.Data)
			err = machine.Handle(cmd)
			if err != nil {
				msg, err := schema.ToJSON(schema.MkMap(
					schema.MkField("error", schema.MkString(err.Error())),
				))
				if err != nil {
					log.Warnf("machine.Handle() %s \n", err)
					return err
				} else {
					g.broadcast.SendBackToSender(connectionID, msg)
					msg, err := MarshalState(stateRecord.Data)
					if err != nil {
						log.Warnf("MarshalState() %s \n", err)
						return err
					} else {
						g.broadcast.SendBackToSender(connectionID, msg)
					}
				}
				return nil
			}

			newState := machine.State()
			if newState == nil {
				// TODO convert to error
				log.Warnln("OnMessage: newState is nil")
				return nil
			}

			// session has also latest stateRecord
			update := schemaless.UpdateRecords[schemaless.Record[tictactoemanage.State]]{
				Saving: map[string]schemaless.Record[tictactoemanage.State]{
					sessionID: {
						ID:      sessionID,
						Type:    "session",
						Data:    newState,
						Version: stateRecord.Version,
					},
				},
			}

			// but pass game stateRecord are also valuable, for example to calculate leaderboards and stats
			if inGame, ok := newState.(*tictactoemanage.SessionInGame); ok {
				update.Saving[inGame.GameID] = schemaless.Record[tictactoemanage.State]{
					ID:      inGame.GameID,
					Type:    "game",
					Data:    newState,
					Version: stateRecord.Version,
				}
			}

			err = g.gameStateRepository.UpdateRecords(update)
			if err != nil {
				log.Println("OnMessage: Set: err", err)
				return err
			}

			msg, err := MarshalState(newState)
			if err != nil {
				log.Errorln("OnMessage: MarshalState: err", err)
				return err
			}
			log.Println("stateRecord", string(msg))

			shouldBroadcast := true
			if shouldBroadcast {
				g.broadcast.BroadcastToSession(sessionID, msg)
			} else {
				g.broadcast.SendBackToSender(connectionID, msg)
			}

			log.Println("OnMessage: done")
			return nil
		},
		func(q tictactoemanage.Query) error {
			log.Printf("OnMessage: query %#v \n", q)
			return tictactoemanage.MustMatchQuery(
				q,
				func(x *tictactoemanage.SessionStatsQuery) error {
					log.Printf("OnMessage(query): SessionStatsQuery %#v \n", *x)
					var result tictactoemanage.QueryResult
					var err error
					result, err = g.query.Query(*x)
					if err != nil {
						log.Errorln("OnMessage(query): g.query.Query: err", err)
						return err
					}

					sch := schema.FromGo(result)
					msg, err := schema.ToJSON(sch)
					if err != nil {
						log.Errorln("OnMessage(query): schema.ToJSON", err)
						return err
					}

					g.broadcast.SendBackToSender(connectionID, msg)
					return nil
				},
			)
		},
		func(s tictactoemanage.Subscription) error {
			log.Printf("OnMessage: subscription %#v \n", s)
			return tictactoemanage.MustMatchSubscription(
				s,
				func(x *tictactoemanage.SessionStatsSubscription) error {
					log.Infof("OnMessage(subscription): SessionStatsSubscription %#v \n", *x)
					ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
					err := g.liveSelect.Process(ctx, x.SessionID)
					if err != nil {
						log.Errorf("SessionStatsSubscription(): liveSelect.Process: %v", err)
					}

					log.Infof("OnMessage(subscription) [return]: SessionStatsSubscription %#v \n", *x)
					return nil
				},
			)
		},
	)
}
func (g *Game) OnConnect(connectionID string) error {
	return g.broadcast.RegisterConnectionID(connectionID)
}
func (g *Game) OnDisconnect(connectionID string) error {
	return g.broadcast.UnregisterConnectionID(connectionID)
}
