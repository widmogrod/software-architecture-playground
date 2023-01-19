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
	)
}

type Repository[A any] interface {
	Get(key string) (A, error)
	GetOrNew(s string) (A, error)
	Set(key string, value A) error
}

func NewGame(b websockproto.Broadcaster, r Repository[tictactoemanage.State]) *Game {
	return &Game{
		broadcast:       b,
		stateRepository: r,
	}
}

type Game struct {
	broadcast       websockproto.Broadcaster
	stateRepository Repository[tictactoemanage.State]
}

func (g *Game) OnMessage(connectionID string, data []byte) error {
	cmd, err := UnmarshalCommand(data)
	if err != nil {
		return err
	}

	sessionID := ExtractSessionID(cmd)
	g.broadcast.AssociateConnectionWithSession(connectionID, sessionID)

	state, err := g.stateRepository.Get(sessionID)
	if err != nil && err != storage.ErrNotFound {
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
		err = g.stateRepository.Set(sessionID, newState)
		if err != nil {
			return err
		}
	}

	msg, err := MarshalState(newState)
	if err != nil {
		return err
	}
	log.Println("state", string(msg))

	shouldBroadcast := true
	if shouldBroadcast {
		g.broadcast.BroadcastToSession(sessionID, msg)
	} else {
		g.broadcast.SendBackToSender(connectionID, msg)
	}

	return nil

}
func (g *Game) OnConnect(connectionID string) error {
	return g.broadcast.RegisterConnectionID(connectionID)
}
func (g *Game) OnDisconnect(connectionID string) error {
	return g.broadcast.UnregisterConnectionID(connectionID)
}
