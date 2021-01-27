package tictactoeaggregate

import (
	"github.com/widmogrod/software-architecture-playground/runtime"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate"
)

func NewTicTacToeAggregate() *TicTacToeAggregate {
	store := aggregate.NewEventStore()
	aggregate := &TicTacToeAggregate{
		state:   nil,
		changes: store,
		ref: &runtime.AggregateRef{
			ID:   "",
			Type: "tictactoe",
		},
	}

	return aggregate
}

type TicTacToeAggregate struct {
	state   *TicTacToeState
	changes *aggregate.EventStore
	ref     *runtime.AggregateRef
}

func (o *TicTacToeAggregate) Ref() *runtime.AggregateRef {
	return o.ref
}

func (o *TicTacToeAggregate) State() interface{} {
	return o.state
}

func (o *TicTacToeAggregate) Changes() *aggregate.EventStore {
	return o.changes
}
func (o *TicTacToeAggregate) Hydrate(state interface{}, ref *runtime.AggregateRef) error {
	o.state = state.(*TicTacToeState)
	o.ref = ref

	return nil
}
