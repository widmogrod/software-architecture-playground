package websockproto

import (
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"log"
)

type Broadcaster interface {
	AssociateConnectionWithSession(connectionID string, sessionID string)
	BroadcastToSession(sessionID string, msg []byte)
	SendBackToSender(connectionID string, msg []byte)
	RegisterConnectionID(connectionID string) error
	UnregisterConnectionID(connectionID string) error
}

type Publisher interface {
	Publish(connectionID string, msg []byte) error
}

var (
	_ Broadcaster = (*InMemoryBroadcaster)(nil)
)

type ConnectionToSession struct {
	ConnectionID string
	SessionID    string
}

func NewBroadcaster(publisher Publisher, repository storage.Repository[ConnectionToSession]) *InMemoryBroadcaster {
	return &InMemoryBroadcaster{
		publisher:  publisher,
		repository: repository,
	}
}

type InMemoryBroadcaster struct {
	publisher  Publisher
	repository storage.Repository[ConnectionToSession]
}

func (i *InMemoryBroadcaster) RegisterConnectionID(connectionID string) error {
	return i.repository.Set(connectionID, ConnectionToSession{
		ConnectionID: connectionID,
	})
}

func (i *InMemoryBroadcaster) UnregisterConnectionID(connectionID string) error {
	err := i.repository.Delete(connectionID)
	if err == storage.ErrNotFound {
		return nil
	}
	return err
}

func (i *InMemoryBroadcaster) AssociateConnectionWithSession(connectionID string, sessionID string) {
	err := i.repository.Set(connectionID, ConnectionToSession{
		ConnectionID: connectionID,
		SessionID:    sessionID,
	})
	if err != nil {
		log.Println("InMemoryBroadcaster.AssociateConnectionWithSession error:", err)
	}
}

func (i *InMemoryBroadcaster) BroadcastToSession(sessionID string, msg []byte) {
	for {
		result, err := i.repository.FindAllKeyEqual("SessionID", sessionID)
		if err != nil {
			log.Println("InMemoryBroadcaster.BroadcastToSession FindAllKeyEqual error:", err)
		}

		for _, item := range result.Items {
			err = i.publisher.Publish(item.ConnectionID, msg)
			// TODO handle this differently
			if err != nil {
				log.Println("InMemoryBroadcaster.BroadcastToSession Publish error:", err)
			}
		}

		if !result.HasNext() {
			break
		}
	}
}

func (i *InMemoryBroadcaster) SendBackToSender(connectionID string, msg []byte) {
	item, err := i.repository.Get(connectionID)
	if err != nil {
		log.Println("InMemoryBroadcaster.SendBackToSender error:", err)
	}

	err = i.publisher.Publish(item.ConnectionID, msg)
	if err != nil {
		log.Println("InMemoryBroadcaster.SendBackToSender error:", err)
	}
}
