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

type OnMessageFunc func(id ConnectionID, data []byte) error
type OnConnectFunc func(id ConnectionID) error
type OnDisconnectFunc func(id ConnectionID) error

func NewInMemoryProtocol() *InMemoryProtocol {
	return &InMemoryProtocol{
		publish: make(chan Item),

		connections: sync.Map{},
	}
}

type ConnectionID = string

type Item struct {
	ConnectionID ConnectionID
	Data         []byte
}

type InMemoryProtocol struct {
	publish     chan Item
	connections sync.Map

	OnMessage    OnMessageFunc
	OnConnect    OnConnectFunc
	OnDisconnect OnDisconnectFunc
}

func (s *InMemoryProtocol) ConnectionOpen(conn net.Conn) {
	connectionID := uuid.Must(uuid.NewUUID()).String()
	err := s.OnConnect(connectionID)

	if err != nil {
		log.Printf("ConnectionOpen: OnConnect(%s) error: %v ", connectionID, err)
		return
	}

	s.connections.Store(conn, connectionID)
}

func (s *InMemoryProtocol) GetConnectionIDFromConn(conn net.Conn) (string, error) {
	connectionID, ok := s.connections.Load(conn)
	if !ok {
		return "", fmt.Errorf("connection not found")
	}
	return connectionID.(string), nil
}

func (s *InMemoryProtocol) ConnectionClose(conn net.Conn) {
	s.connections.Delete(conn)

	connectionID, err := s.GetConnectionIDFromConn(conn)
	if err != nil {
		return
	}
	err = s.OnDisconnect(connectionID)
	if err != nil {
		log.Printf("ConnectionClose: OnDisconnect(%s) error %v \n", connectionID, err)
	}
}

func (s *InMemoryProtocol) ConnectionReceiveData(conn net.Conn) error {
	msg, _, err := wsutil.ReadClientData(conn)
	log.Println("msg", string(msg))
	if err != nil {
		return err
	}

	connectionID, err := s.GetConnectionIDFromConn(conn)
	if err != nil {
		return err
	}

	return s.OnMessage(connectionID, msg)
}

func (s *InMemoryProtocol) ToPublish() chan Item {
	return s.publish
}

func (s *InMemoryProtocol) PublishLoop() {
	for {
		select {
		case pub := <-s.publish:
			conn, err := s.connByID(pub.ConnectionID)
			if err != nil {
				log.Println("PublishLoop: connByID error", err)
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

func (s *InMemoryProtocol) connByID(connectionID ConnectionID) (net.Conn, error) {
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

func (s *InMemoryProtocol) Publish(connectionID string, msg []byte) error {
	i := Item{
		ConnectionID: connectionID,
		Data:         msg,
	}
	s.publish <- i
	return nil
}
