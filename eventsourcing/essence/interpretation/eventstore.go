package interpretation

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/protoorder"
)

var _ protoorder.OrderAggregateServer = &EventStore{}

type EventStore struct {
}

func (e *EventStore) CreateOrder(ctx context.Context, request *protoorder.CreateOrderRequest) (*protoorder.OrderAggregateState, error) {
	panic("implement me")
}
