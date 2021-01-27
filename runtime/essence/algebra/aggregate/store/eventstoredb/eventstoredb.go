package eventstoredb

import (
	"context"
	"fmt"
	eventstore "github.com/EventStore/EventStore-Client-Go/client"
	"github.com/EventStore/EventStore-Client-Go/direction"
	"github.com/EventStore/EventStore-Client-Go/messages"
	"github.com/EventStore/EventStore-Client-Go/streamrevision"
	"github.com/gofrs/uuid"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate/store"
	"strings"
)

type SerDeSer interface {
	Register(typ interface{})
	RegisterName(name string, typ interface{})
	Name(typ interface{}) (string, error)
	Serialise(typ interface{}) ([]byte, error)
	DeSerialiseName(name string, data []byte) (interface{}, error)
}

func NewEventStoreDB(client *eventstore.Client, deser SerDeSer) *evenstordbimpl {
	return &evenstordbimpl{
		client: client,
		deser:  deser,
	}
}

type evenstordbimpl struct {
	client *eventstore.Client
	deser  SerDeSer
}

func (i *evenstordbimpl) ReadChanges(ctx context.Context, aggregateID string) ([]runtime.Change, error) {
	// TODO read til exhaustion, don't wait for limit 20, or snapshot?
	events, err := i.client.ReadStreamEvents(ctx, direction.Forwards, aggregateID, streamrevision.StreamRevisionStart, 20, false)
	if err != nil {
		if strings.Contains(err.Error(), "stream was not found") {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	changes := make([]runtime.Change, 0)
	for _, event := range events {
		change := runtime.Change{
			Payload: nil,
			// TODO uint64!
			Version: event.EventNumber,
		}

		payload, err := i.deser.DeSerialiseName(event.EventType, event.Data)
		change.Payload = payload

		if err != nil {
			return nil, fmt.Errorf("unnmarshall, %s - %w", event.EventType, err)
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (i *evenstordbimpl) AppendChanges(ctx context.Context, aggregateID string, version uint64, changes []runtime.Change) error {
	events := make([]messages.ProposedEvent, 0)

	for _, change := range changes {
		event := messages.ProposedEvent{
			ContentType: "application/json",
			EventID:     uuid.Must(uuid.DefaultGenerator.NewV4()),
		}

		var err error

		event.EventType, err = i.deser.Name(change.Payload)
		event.Data, err = i.deser.Serialise(change.Payload)

		if err != nil {
			return fmt.Errorf("AppendChanges, Marshall, %T - %w", change.Payload, err)
		}

		events = append(events, event)
	}
	var revision = streamrevision.StreamRevisionNoStream
	if version != ^uint64(0) {
		revision = streamrevision.NewStreamRevision(version)
	}

	fmt.Printf("version =%#v\n", version)
	fmt.Printf("revision =%#v\n", revision)
	res, err := i.client.AppendToStream(ctx, aggregateID, revision, events)
	fmt.Printf("res=%#v\n", res)
	fmt.Printf("err=%#v\n", err)
	if err != nil {
		return fmt.Errorf("AppendChanges %w", err)
	}

	return nil
}
