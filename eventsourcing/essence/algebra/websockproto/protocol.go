package websockproto

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/machine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"log"
	"net"
	"sync"
)

func NewProtocol[C, S any](reg *storage.RepositoryInMemory[*machine.Machine[C, S]]) *Protocol[C, S] {
	return &Protocol[C, S]{
		sessions: reg,
		publish:  make(chan Item),
	}
}

type Protocol[C, S any] struct {
	publish             chan Item
	connections         sync.Map
	sessionToConnection sync.Map
	sessions            *storage.RepositoryInMemory[*machine.Machine[C, S]]

	UnmarshalCommand func(msg []byte) (C, error)
	MarshalState     func(state S) ([]byte, error)
	ExtractSessionID func(cmd C) string
}

func (s *Protocol[C, S]) ConnectionOpen(conn net.Conn) {
	s.connections.Store(conn, uuid.Must(uuid.NewUUID()).String())
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
