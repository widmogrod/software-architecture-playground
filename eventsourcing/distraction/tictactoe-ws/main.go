package main

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/rs/cors"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/websockproto"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"io"
	"log"
	"net/http"
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

	state, err := g.stateRepository.GetOrNew(sessionID)
	if err != nil {
		return err
	}

	machine := tictactoemanage.NewMachineWithState(state)
	err = machine.Handle(cmd)
	if err != nil {
		log.Println("Handle error continued:", err)
		//return err
	}

	newState := machine.State()
	err = g.stateRepository.Set(sessionID, newState)
	if err != nil {
		return err
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
	return nil
}
func (g *Game) OnDisconnect(connectionID string) error {
	return nil
}

func main() {
	reg := storage.NewRepositoryInMemory(func() tictactoemanage.State {
		return nil
	})

	storage := storage.NewRepositoryInMemory(func() websockproto.ConnectionToSession {
		panic("not supported creation of ConnectionToSession")
	})
	proto := websockproto.NewInMemoryProtocol()
	broadcaster := websockproto.NewBroadcaster(proto, storage)

	game := &Game{
		broadcast:       broadcaster,
		stateRepository: reg,
	}

	proto.OnMessage = game.OnMessage
	proto.OnConnect = game.OnConnect
	proto.OnDisconnect = game.OnDisconnect

	go proto.PublishLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.FileServer(http.Dir("../tictactoe-app/build")).ServeHTTP)
	mux.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		proto.ConnectionOpen(conn)

		go func() {
			defer proto.ConnectionClose(conn)
			defer conn.Close()

			for {
				err = proto.ConnectionReceiveData(conn)
				if err != nil {
					if err == io.EOF {
						log.Println("ConnectionReceiveData: CLOSED")
						break
					}
					log.Println("ConnectionReceiveData:", err)
					continue
				}
			}
		}()
	})

	handler := cors.AllowAll().Handler(mux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
