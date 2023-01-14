package main

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gofrs/uuid"
	"github.com/rs/cors"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/machine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

type Repository[A any] struct {
	store sync.Map
	new   func() A
}

var ErrNotFound = fmt.Errorf("not found")

func (r *Repository[A]) Get(key string) (A, error) {
	v, ok := r.store.Load(key)
	if !ok {
		var a A
		return a, ErrNotFound
	}
	return v.(A), nil
}

func (r *Repository[A]) Set(key string, value A) error {
	r.store.Store(key, value)
	return nil
}

func (r *Repository[A]) GetOrNew(s string) (A, error) {
	v, err := r.Get(s)
	if err == nil {
		return v, nil
	}

	if err != nil && err != ErrNotFound {
		var a A
		return a, err
	}

	v = r.new()

	err = r.Set(s, v)
	if err != nil {
		var a A
		return a, err
	}

	return v, nil
}

type SessionIDAware interface {
	SessionID() string
}
type SessionIDAwareCommand interface {
	tictactoemanage.Command
	SessionIDAware
}

type Protocol[C, S any] struct {
	connections         sync.Map
	sessionToConnection sync.Map
	sessions            Repository[*machine.Machine[C, S]]

	UnmarshalCommand func(msg []byte) (C, error)
	MarshalState     func(state S) ([]byte, error)
	ExtractSessionID func(cmd C) string
	publish          chan Item
}

func (s *Protocol[C, S]) ConnectionOpen(conn net.Conn) {
	s.connections.Store(conn, uuid.Must(uuid.NewV4()).String())
}

func (s *Protocol[C, S]) ConnectionClose(conn net.Conn) {
	s.connections.Delete(conn)
	s.sessionToConnection.Delete(conn)
}

func (s *Protocol[C, S]) ConnectionReceiveData(conn net.Conn) error {
	msg, _, err := wsutil.ReadClientData(conn)
	log.Println("msg", string(msg))
	if err != nil {
		return err
	}

	cmd, err := s.UnmarshalCommand(msg)
	if err != nil {
		return err
	}

	connectionAny, ok := s.connections.Load(conn)
	if !ok {
		return fmt.Errorf("connection not found")
	}
	connectionID := connectionAny.(string)

	sessionID := s.ExtractSessionID(cmd)
	s.AssociateConnectionWithSession(conn, sessionID)

	machine, err := s.sessions.GetOrNew(sessionID)
	if err != nil {
		return err
	}

	err = machine.Handle(cmd)
	if err != nil {
		log.Println("Handle error continued:", err)
		//return err
	}

	state := machine.State()

	msg, err = s.MarshalState(state)
	if err != nil {
		return err
	}
	log.Println("state", string(msg))

	shouldBroadcast := true
	if shouldBroadcast {
		s.BroadcastToSession(sessionID, msg)
	} else {
		s.SendBackToSender(connectionID, msg)
	}

	return nil
}

func (s *Protocol[C, S]) BroadcastToSession(sessionID string, msg []byte) {
	var conns []net.Conn
	s.sessionToConnection.Range(func(conn, connSessionID interface{}) bool {
		if connSessionID == sessionID {
			conns = append(conns, conn.(net.Conn))
		}
		return true
	})

	for _, conn := range conns {
		i := Item{
			Conn: conn,
			Op:   ws.OpText,
			Data: msg,
		}
		s.pub(i)
	}
}

func (s *Protocol[C, S]) SendBackToSender(connectionID string, msg []byte) {
	s.connections.Range(func(conn, connConnectionID interface{}) bool {
		if connConnectionID == connectionID {
			i := Item{
				Conn: conn.(net.Conn),
				Op:   ws.OpText,
				Data: msg,
			}
			s.pub(i)
		}
		return true
	})
}

func (s *Protocol[C, S]) pub(i Item) {
	log.Println("chan pub.Data == ", string(i.Data))
	s.publish <- i
}

func (s *Protocol[C, S]) ToPublish() chan Item {
	return s.publish
}

func (s *Protocol[C, S]) AssociateConnectionWithSession(conn net.Conn, sessionID string) {
	s.sessionToConnection.Store(conn, sessionID)
}

func (s *Protocol[C, S]) PublishLoop() {
	for {
		select {
		case pub := <-s.publish:
			err := wsutil.WriteServerMessage(pub.Conn, pub.Op, pub.Data)
			if err != nil {
				log.Println("WriteServerMessage:", err)
				s.ConnectionClose(pub.Conn)
			}
			log.Println("published:", string(pub.Data))
		}
	}
}

type Item struct {
	Conn net.Conn
	Op   ws.OpCode
	Data []byte
}

func main() {
	reg := Repository[*machine.Machine[tictactoemanage.Command, tictactoemanage.State]]{
		store: sync.Map{},
		new:   tictactoemanage.NewMachine,
	}

	proto := &Protocol[tictactoemanage.Command, tictactoemanage.State]{
		sessions: reg,
		publish:  make(chan Item),

		UnmarshalCommand: func(msg []byte) (tictactoemanage.Command, error) {
			sch, err := schema.JsonToSchema(msg)
			if err != nil {
				return nil, fmt.Errorf("UnmarshalCommand: %s", err)
			}

			goo := schema.SchemaToGo(
				sch,
				schema.WhenPath(
					[]string{},
					schema.UseStruct(&tictactoemanage.CommandOneOf{}),
				),
				schema.WhenPath(
					[]string{"CreateSessionCMD"},
					schema.UseStruct(&tictactoemanage.CreateSessionCMD{}),
				),
				schema.WhenPath(
					[]string{"JoinGameSessionCMD"},
					schema.UseStruct(&tictactoemanage.JoinGameSessionCMD{}),
				),
				schema.WhenPath(
					[]string{"GameSessionWithBotCMD"},
					schema.UseStruct(&tictactoemanage.GameSessionWithBotCMD{}),
				),
				schema.WhenPath(
					[]string{"NewGameCMD"},
					schema.UseStruct(&tictactoemanage.NewGameCMD{}),
				),
				schema.WhenPath(
					[]string{"LeaveGameSessionCMD"},
					schema.UseStruct(&tictactoemanage.LeaveGameSessionCMD{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD"},
					schema.UseStruct(&tictactoemanage.GameActionCMD{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD", "Action"},
					schema.UseStruct(&tictacstatemachine.CommandOneOf{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD", "Action", "CreateGameCMD"},
					schema.UseStruct(&tictacstatemachine.CreateGameCMD{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD", "Action", "JoinGameCMD"},
					schema.UseStruct(&tictacstatemachine.JoinGameCMD{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD", "Action", "StartGameCMD"},
					schema.UseStruct(&tictacstatemachine.StartGameCMD{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD", "Action", "MoveCMD"},
					schema.UseStruct(&tictacstatemachine.MoveCMD{}),
				),
				schema.WhenPath(
					[]string{"GameActionCMD", "Action", "GiveUpCMD"},
					schema.UseStruct(&tictacstatemachine.GiveUpCMD{}),
				),
			)

			cmd, ok := goo.(*tictactoemanage.CommandOneOf)
			if !ok {
				return nil, fmt.Errorf("UnmarshalCommand: %s", "not a command")
			}

			return cmd.Unwrap(), nil
		},
		MarshalState: func(state tictactoemanage.State) ([]byte, error) {
			stateOneOf := tictactoemanage.WrapStateOneOf(state)
			return json.Marshal(stateOneOf)
		},
		ExtractSessionID: func(cmd tictactoemanage.Command) string {
			return tictactoemanage.MustMatchCommand(cmd,
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
				})
		},
	}

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
