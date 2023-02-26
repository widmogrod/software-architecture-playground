package tictactoe_game_server

import (
	"errors"
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"log"
)

func init() {
	schema.SetDefaultUnionTypeFormatter(schema.FormatUnionNameUsingTypeName)
}

func UnmarshalQueryOrCommand(
	data []byte,
	onCommand func(tictactoemanage.Command) error,
	onQuery func(tictactoemanage.Query) error,
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
	UpdateRecords(s storage.UpdateRecords[any]) error
}

type SessionWithGame struct {
	SessionID     string
	CurrentGameID string
}

func NewGame(b websockproto.Broadcaster, r storage.Repository2[tictactoemanage.State], q *OpenSearchStorage) *Game {
	return &Game{
		broadcast:           b,
		gameStateRepository: r,
		query:               q,
	}
}

type Query interface {
	Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error)
}

type Game struct {
	broadcast           websockproto.Broadcaster
	gameStateRepository storage.Repository2[tictactoemanage.State]
	query               Query
}

func (g *Game) OnMessage(connectionID string, data []byte) error {
	return UnmarshalQueryOrCommand(
		data,
		func(cmd tictactoemanage.Command) error {
			sessionID := ExtractSessionID(cmd)
			g.broadcast.AssociateConnectionWithSession(connectionID, sessionID)

			stateRecord, err := g.gameStateRepository.Get(sessionID, "session")
			if err != nil && !errors.Is(err, storage.ErrNotFound) {
				log.Println("OnMessage: Get: err", err)
				return err
			}

			machine := tictactoemanage.NewMachineWithState(stateRecord.Data)
			err = machine.Handle(cmd)
			if err != nil {
				log.Println("Handle error continued:", err)
				//return err
			}

			newState := machine.State()
			if newState != nil {
				// session has also latest stateRecord
				update := storage.UpdateRecords[storage.Record[tictactoemanage.State]]{
					Saving: map[string]storage.Record[tictactoemanage.State]{
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
					update.Saving[inGame.GameID] = storage.Record[tictactoemanage.State]{
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
			}

			msg, err := MarshalState(newState)
			if err != nil {
				log.Println("OnMessage: MarshalState: err", err)
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
			log.Println("OnMessage: query", q)
			return tictactoemanage.MustMatchQuery(
				q,
				func(x *tictactoemanage.SessionStatsQuery) error {
					log.Printf("OnMessage(query): SessionStatsQuery %#v \n", *x)
					var result tictactoemanage.QueryResult
					var err error
					result, err = g.query.Query(*x)
					if err != nil {
						log.Println("OnMessage(query): g.query.Query: err", err)
						return err
					}

					sch := schema.FromGo(result)
					msg, err := schema.ToJSON(sch)
					if err != nil {
						log.Println("OnMessage(query): schema.ToJSON", err)
						return err
					}

					g.broadcast.SendBackToSender(connectionID, msg)
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
