package websockproto

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"sync"
)

type OnMessageFunc func(id ConnectionID, data []byte) error
type OnConnectFunc func(id ConnectionID) error
type OnDisconnectFunc func(id ConnectionID) error

func NewInMemoryProtocol() *InMemoryProtocol {
	return &InMemoryProtocol{
		publish:     make(chan Item),
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
	s.connections.Store(conn, connectionID)

	err := s.OnConnect(connectionID)
	if err != nil {
		log.Warnf("ConnectionOpen: OnConnect(%s) error: %v ", connectionID, err)
		return
	}
}

func (s *InMemoryProtocol) getConnectionIDFromConn(conn net.Conn) (string, error) {
	connectionID, ok := s.connections.Load(conn)
	if !ok {
		return "", fmt.Errorf("connection not found")
	}
	return connectionID.(string), nil
}

func (s *InMemoryProtocol) ConnectionClose(conn net.Conn) {
	connectionID, err := s.getConnectionIDFromConn(conn)
	if err != nil {
		return
	}

	s.connections.Delete(conn)

	err = s.OnDisconnect(connectionID)
	if err != nil {
		log.Warnf("ConnectionClose: OnDisconnect(%s) error %v \n", connectionID, err)
	}
}

func (s *InMemoryProtocol) ConnectionReceiveData(conn net.Conn) error {
	msg, _, err := wsutil.ReadClientData(conn)
	log.Debugln("msg", string(msg))
	if err != nil {
		return err
	}

	connectionID, err := s.getConnectionIDFromConn(conn)
	if err != nil {
		return err
	}

	return s.OnMessage(connectionID, msg)
}

func (s *InMemoryProtocol) PublishLoop() {
	for {
		select {
		case pub := <-s.publish:
			log.Debugf("PublishLoop: %s %s \n", pub.ConnectionID, string(pub.Data))
			conn, err := s.connByID(pub.ConnectionID)
			if err != nil {
				log.Warnln("PublishLoop: connByID error", err)
				continue
			}

			err = wsutil.WriteServerMessage(conn, ws.OpText, pub.Data)
			if err != nil {
				log.Warnln("PublishLoop: WriteServerMessage:", err)
				s.ConnectionClose(conn)
			}

			log.Debugln("PublishLoop: published:", string(pub.Data))
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
	log.Debugln("Publish:", connectionID, string(msg))
	i := Item{
		ConnectionID: connectionID,
		Data:         msg,
	}
	s.publish <- i
	return nil
}

func (s *InMemoryProtocol) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	go func() {
		s.ConnectionOpen(conn)
		defer s.ConnectionClose(conn)
		defer conn.Close()

		for {
			log.Infof("ConnectionReceiveData: Start")
			err = s.ConnectionReceiveData(conn)
			log.Infof("ConnectionReceiveData: End")
			if err != nil {
				if err == io.EOF {
					log.Infof("ConnectionReceiveData: CLOSED")
					break
				}
				log.Errorln("ConnectionReceiveData error:", err)
				continue
			}
		}
	}()
}
