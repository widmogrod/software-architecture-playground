package tictactoe_game_server

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"log"
)

func UnmarshalCommand(msg []byte) (tictactoemanage.Command, error) {
	sch, err := schema.FromJSON(msg)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalCommand: %s", err)
	}

	goo := schema.ToGo(sch)

	cmd, ok := goo.(tictactoemanage.Command)
	if !ok {
		return nil, fmt.Errorf("UnmarshalCommand: %T not a command", goo)
	}

	return cmd, nil
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

func NewGame(b websockproto.Broadcaster, r Repository[tictactoemanage.State]) *Game {
	return &Game{
		broadcast:           b,
		gameStateRepository: r,
	}
}

type Game struct {
	broadcast           websockproto.Broadcaster
	gameStateRepository Repository[tictactoemanage.State]
}

func (g *Game) OnMessage(connectionID string, data []byte) error {
	log.Println("command", string(data))
	cmd, err := UnmarshalCommand(data)
	if err != nil {
		log.Println("command err", err)
		return err
	}

	fmt.Printf("command go %#v \n", cmd)

	sessionID := ExtractSessionID(cmd)
	g.broadcast.AssociateConnectionWithSession(connectionID, sessionID)

	state, err := storage.RetriveID[tictactoemanage.State](g.gameStateRepository, "session:"+sessionID)
	if err != nil && err != storage.ErrNotFound {
		log.Println("OnMessage: Get: err", err)
		return err
	}

	machine := tictactoemanage.NewMachineWithState(state)
	err = machine.Handle(cmd)
	if err != nil {
		log.Println("Handle error continued:", err)
		//return err
	}

	newState := machine.State()
	if newState != nil {
		// session has also latest state
		update := storage.UpdateRecords[any]{
			Saving: map[string]any{
				"session:" + sessionID: newState,
			},
		}

		// but pass game state are also valuable, for example to calculate leaderboards and stats
		if inGame, ok := newState.(*tictactoemanage.SessionInGame); ok {
			update.Saving["game:"+inGame.GameID] = inGame
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
	log.Println("state", string(msg))

	shouldBroadcast := true
	if shouldBroadcast {
		g.broadcast.BroadcastToSession(sessionID, msg)
	} else {
		g.broadcast.SendBackToSender(connectionID, msg)
	}

	log.Println("OnMessage: done")
	return nil

}
func (g *Game) OnConnect(connectionID string) error {
	return g.broadcast.RegisterConnectionID(connectionID)
}
func (g *Game) OnDisconnect(connectionID string) error {
	return g.broadcast.UnregisterConnectionID(connectionID)
}
