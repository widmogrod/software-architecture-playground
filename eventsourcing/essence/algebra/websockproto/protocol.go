package websockproto

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
)

func NewProtocol[C, S any]() *Protocol[C, S] {
	return &Protocol[C, S]{
		publish: make(chan Item),
	}
}

type ConnectionID = string

type Protocol[C, S any] struct {
	publish             chan Item
	connections         sync.Map
	sessionToConnection sync.Map

	UnmarshalCommand func(msg []byte) (C, error)
	MarshalState     func(state S) ([]byte, error)
	ExtractSessionID func(cmd C) string
	OnMessage        func(connectionID string, data []byte) error
}

func (s *Protocol[C, S]) ConnectionOpen(conn net.Conn) {
	s.connections.Store(conn, uuid.Must(uuid.NewUUID()).String())
}

func (s *Protocol[C, S]) ConnectionID(conn net.Conn) (string, error) {
	connectionID, ok := s.connections.Load(conn)
	if !ok {
		return "", fmt.Errorf("connection not found")
	}
	return connectionID.(string), nil
}

func (s *Protocol[C, S]) ConnectionClose(conn net.Conn) {
	connectionID, err := s.ConnectionID(conn)
	if err != nil {
		return
	}

	s.connections.Delete(conn)
	s.sessionToConnection.Delete(connectionID)
}

func (s *Protocol[C, S]) ConnectionReceiveData(conn net.Conn) error {
	msg, _, err := wsutil.ReadClientData(conn)
	log.Println("msg", string(msg))
	if err != nil {
		return err
	}

	connectionID, err := s.ConnectionID(conn)
	if err != nil {
		return err
	}

	return s.OnMessage(connectionID, msg)
}

func (s *Protocol[C, S]) BroadcastToSession(sessionID string, msg []byte) {
	var conns []ConnectionID
	s.sessionToConnection.Range(func(connectionID, connSessionID interface{}) bool {
		if connSessionID == sessionID {
			conns = append(conns, connectionID.(ConnectionID))
		}
		return true
	})

	for _, connectionID := range conns {
		i := Item{
			ConnectionID: connectionID,
			Data:         msg,
		}
		s.pub(i)
	}
}

func (s *Protocol[C, S]) SendBackToSender(connectionID string, msg []byte) {
	i := Item{
		ConnectionID: connectionID,
		Data:         msg,
	}
	s.pub(i)
}

func (s *Protocol[C, S]) pub(i Item) {
	log.Println("chan pub.Data == ", string(i.Data))
	s.publish <- i
}

func (s *Protocol[C, S]) ToPublish() chan Item {
	return s.publish
}

func (s *Protocol[C, S]) AssociateConnectionWithSession(connectionID, sessionID string) {
	s.sessionToConnection.Store(connectionID, sessionID)
}

func (s *Protocol[C, S]) PublishLoop() {
	for {
		select {
		case pub := <-s.publish:
			conn, err := s.ConnByID(pub.ConnectionID)
			if err != nil {
				log.Println("PublishLoop: ConnByID error", err)
				continue
			}

			err = wsutil.WriteServerMessage(conn, ws.OpText, pub.Data)
			if err != nil {
				log.Println("PublishLoop: WriteServerMessage:", err)
				s.ConnectionClose(conn)
			}

			log.Println("PublishLoop: published:", string(pub.Data))
		}
	}
}

func (s *Protocol[C, S]) ConnByID(connectionID ConnectionID) (net.Conn, error) {
	var conn net.Conn
	s.connections.Range(func(key, value interface{}) bool {
		if value == connectionID {
			conn = key.(net.Conn)
			return false
		}
		return true
	})

	if conn == nil {
		return nil, fmt.Errorf("connection not found")
	}

	return conn, nil

}

type Item struct {
	ConnectionID ConnectionID
	Data         []byte
}
