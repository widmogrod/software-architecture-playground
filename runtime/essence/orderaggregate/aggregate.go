package orderaggregate

import (
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate"
)

func NewOrderAggregate() *OrderAggregate {
	store := aggregate.NewEventStore()
	return &OrderAggregate{
		state:   nil,
		changes: store,
	}
}

type OrderAggregate struct {
	state   *OrderAggregateState
	changes *aggregate.EventStore
}

func (o *OrderAggregate) State() interface{} {
	return o.state
}

func (o *OrderAggregate) Changes() *aggregate.EventStore {
	return o.changes
}
