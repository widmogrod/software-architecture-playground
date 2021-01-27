package orderaggregate

import (
	"github.com/widmogrod/software-architecture-playground/runtime"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate"
)

func NewOrderAggregate() *OrderAggregate {
	store := aggregate.NewEventStore()
	aggregate := &OrderAggregate{
		state:   nil,
		changes: store,
		ref: &runtime.AggregateRef{
			ID:   "",
			Type: "order",
		},
	}

	return aggregate
}

type OrderAggregate struct {
	state   *OrderAggregateState
	changes *aggregate.EventStore
	ref     *runtime.AggregateRef
}

func (o *OrderAggregate) Ref() *runtime.AggregateRef {
	return o.ref
}

func (o *OrderAggregate) State() interface{} {
	return o.state
}

func (o *OrderAggregate) Changes() *aggregate.EventStore {
	return o.changes
}
func (o *OrderAggregate) Hydrate(state interface{}, ref *runtime.AggregateRef) error {
	o.state = state.(*OrderAggregateState)
	o.ref = ref

	return nil
}

type Aggregate interface {
	State() interface{}
	Changes() *aggregate.EventStore
	Ref() *runtime.AggregateRef
	Apply(change interface{}) error
	Handle(cmd interface{}) error
}
