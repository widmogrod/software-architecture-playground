package websockproto

import (
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
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

func NewBroadcaster(publisher Publisher, repository storage.Repository2[ConnectionToSession]) *InMemoryBroadcaster {
	return &InMemoryBroadcaster{
		publisher:  publisher,
		repository: repository,
	}
}

type InMemoryBroadcaster struct {
	publisher  Publisher
	repository storage.Repository2[ConnectionToSession]
}

func (i *InMemoryBroadcaster) RegisterConnectionID(connectionID string) error {
	return i.repository.UpdateRecords(storage.Save(storage.Record[ConnectionToSession]{
		ID:   connectionID,
		Type: "connectionToSession",
		Data: ConnectionToSession{
			ConnectionID: connectionID,
		},
	}))
}

func (i *InMemoryBroadcaster) UnregisterConnectionID(connectionID string) error {
	return i.repository.UpdateRecords(storage.Delete(storage.Record[ConnectionToSession]{
		ID: connectionID,
	}))
}

func (i *InMemoryBroadcaster) AssociateConnectionWithSession(connectionID string, sessionID string) {
	record, err := i.repository.Get(connectionID)
	if err != nil {
		log.Println("InMemoryBroadcaster.AssociateConnectionWithSession i.repository.Get() err:", err)
		return
	}

	record.Data.SessionID = sessionID

	err = i.repository.UpdateRecords(storage.Save(record))
	if err != nil {
		log.Println("InMemoryBroadcaster.AssociateConnectionWithSession error:", err)
	}
}

func (i *InMemoryBroadcaster) BroadcastToSession(sessionID string, msg []byte) {
	cursor := storage.FindingRecords[storage.Record[ConnectionToSession]]{
		Where: predicate.MustWhere(
			"Type = :type AND Data.SessionID = :sessionID",
			predicate.ParamBinds{
				":type":      schema.MkString("connectionToSession"),
				":sessionID": schema.MkString(sessionID),
			},
		),
	}

	for {
		result, err := i.repository.FindingRecords(cursor)
		if err != nil {
			log.Println("InMemoryBroadcaster.BroadcastToSession FindingRecords error:", err)
		}

		for _, item := range result.Items {
			log.Println("BroadcastToSession connectionID:", item.Data.ConnectionID)
			err = i.publisher.Publish(item.Data.ConnectionID, msg)
			// TODO handle this differently
			if err != nil {
				log.Println("InMemoryBroadcaster.BroadcastToSession Publish error:", err)
			}
		}

		if !result.HasNext() {
			break
		}

		cursor = *result.Next
	}
}

func (i *InMemoryBroadcaster) SendBackToSender(connectionID string, msg []byte) {
	item, err := i.repository.Get(connectionID)
	if err != nil {
		log.Println("InMemoryBroadcaster.SendBackToSender error:", err)
	}

	err = i.publisher.Publish(item.Data.ConnectionID, msg)
	if err != nil {
		log.Println("InMemoryBroadcaster.SendBackToSender error:", err)
	}
}
