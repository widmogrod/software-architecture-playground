package websockproto

import (
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
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

func NewBroadcaster(publisher Publisher, repository schemaless.Repository[ConnectionToSession]) *InMemoryBroadcaster {
	return &InMemoryBroadcaster{
		publisher:  publisher,
		repository: repository,
	}
}

type InMemoryBroadcaster struct {
	publisher  Publisher
	repository schemaless.Repository[ConnectionToSession]
}

func (i *InMemoryBroadcaster) RegisterConnectionID(connectionID string) error {
	return i.repository.UpdateRecords(schemaless.Save(schemaless.Record[ConnectionToSession]{
		ID:   connectionID,
		Type: "connectionToSession",
		Data: ConnectionToSession{
			ConnectionID: connectionID,
		},
	}))
}

func (i *InMemoryBroadcaster) UnregisterConnectionID(connectionID string) error {
	return i.repository.UpdateRecords(schemaless.Delete(schemaless.Record[ConnectionToSession]{
		ID:   connectionID,
		Type: "connectionToSession",
	}))
}

func (i *InMemoryBroadcaster) AssociateConnectionWithSession(connectionID string, sessionID string) {
	record, err := i.repository.Get(connectionID, "connectionToSession")
	if err != nil {
		log.Errorln("InMemoryBroadcaster.AssociateConnectionWithSession i.repository.Get() err:", err)
		return
	}

	record.Data.SessionID = sessionID

	err = i.repository.UpdateRecords(schemaless.Save(record))
	if err != nil {
		log.Errorln("InMemoryBroadcaster.AssociateConnectionWithSession error:", err)
	}
}

func (i *InMemoryBroadcaster) BroadcastToSession(sessionID string, msg []byte) {
	log.Infoln("InMemoryBroadcaster.BroadcastToSession sessionID [start]:", sessionID, "msg:", string(msg))
	defer log.Infoln("InMemoryBroadcaster.BroadcastToSession sessionID [end]:", sessionID, "msg:", string(msg))

	cursor := schemaless.FindingRecords[schemaless.Record[ConnectionToSession]]{
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
		log.Infoln("InMemoryBroadcaster.BroadcastToSession FindingRecords result:", result, "err:", err)
		if err != nil {
			log.Errorln("InMemoryBroadcaster.BroadcastToSession FindingRecords error:", err)
			break
		}

		for _, item := range result.Items {
			log.Debugln("BroadcastToSession connectionID: publish", item.Data.ConnectionID)
			err = i.publisher.Publish(item.Data.ConnectionID, msg)
			// TODO handle this differently
			if err != nil {
				log.Errorln("InMemoryBroadcaster.BroadcastToSession Publish error:", err)
			}
		}

		if !result.HasNext() {
			log.Infoln("InMemoryBroadcaster.BroadcastToSession FindingRecords no more connections:")
			break
		}

		cursor = *result.Next
	}
}

func (i *InMemoryBroadcaster) SendBackToSender(connectionID string, msg []byte) {
	item, err := i.repository.Get(connectionID, "connectionToSession")
	if err != nil {
		log.Errorln("InMemoryBroadcaster.SendBackToSender error:", err)
		return
	}

	err = i.publisher.Publish(item.Data.ConnectionID, msg)
	if err != nil {
		log.Errorln("InMemoryBroadcaster.SendBackToSender error:", err)
	}
}
