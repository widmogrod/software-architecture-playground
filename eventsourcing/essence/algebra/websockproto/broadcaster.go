package websockproto

import (
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
)

type Broadcaster interface {
	AssociateConnectionWithSession(connectionID string, sessionID string)
	BroadcastToSession(sessionID string, msg []byte)
	SendBackToSender(connectionID string, msg []byte)
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

func NewInMemoryBroadcaster(publisher Publisher) *InMemoryBroadcaster {
	return &InMemoryBroadcaster{
		repository: storage.NewRepositoryInMemory(func() ConnectionToSession {
			panic("not supported creation of ConnectionToSession")
		}),
		publisher: publisher,
	}
}

type InMemoryBroadcaster struct {
	publisher  Publisher
	repository *storage.RepositoryInMemory[ConnectionToSession]
}

func (i *InMemoryBroadcaster) AssociateConnectionWithSession(connectionID string, sessionID string) {
	err := i.repository.Set(connectionID, ConnectionToSession{
		ConnectionID: connectionID,
		SessionID:    sessionID,
	})
	if err != nil {
		panic(err)
	}
}

func (i *InMemoryBroadcaster) BroadcastToSession(sessionID string, msg []byte) {
	for {
		result, err := i.repository.FindAllKeyEqual("SessionID", sessionID)
		if err != nil {
			panic(err)
		}

		for _, item := range result.Items {
			err = i.publisher.Publish(item.ConnectionID, msg)
			// TODO handle this differently
			if err != nil {
				panic(err)
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
		panic(err)
	}

	err = i.publisher.Publish(item.ConnectionID, msg)
	if err != nil {
		panic(err)
	}
}
