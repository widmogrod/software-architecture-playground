package tictactoeaggregate

import (
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate"
)

func NewTicTacToeAggregate() *TicTacToeAggregate {
	store := aggregate.NewEventStore()
	return &TicTacToeAggregate{
		state:   nil,
		changes: store,
	}
}

type TicTacToeAggregate struct {
	state   *TicTacToeState
	changes *aggregate.EventStore
}

func (o *TicTacToeAggregate) State() interface{} {
	return o.state
}

func (o *TicTacToeAggregate) Changes() *aggregate.EventStore {
	return o.changes
}
