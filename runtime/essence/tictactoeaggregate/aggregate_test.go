package tictactoeaggregate

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/aggssert"
	"math/rand"
	"testing"
)

var (
	p1 = "P1"
	p2 = "P2"
)

func TestTicTacToe_new_aggregate_has_empty_state(t *testing.T) {
	a := NewTicTacToeAggregate()
	aggssert.Empty(t, a)
}

func TestTicTacToe_aggregate_state_equal_to_new_replay_state(t *testing.T) {
	a := NewTicTacToeAggregate()

	// Two ways to start game
	if rand.Float32() < 0.5 {
		err := a.Handle(&CreateGameCMD{
			FirstPlayerID: p1,
		})
		assert.NoError(t, err)

		err = a.Handle(&JoinGameCMD{
			SecondPlayerID: p2,
		})
		assert.NoError(t, err)
	} else {
		err := a.Handle(&StartGameCMD{
			FirstPlayerID:  p1,
			SecondPlayerID: p2,
		})
		assert.NoError(t, err)
	}

	err := a.Handle(&MoveCMD{
		PlayerID: p1,
		Position: "1.1",
	})
	assert.NoError(t, err)

	err = a.Handle(&MoveCMD{
		PlayerID: p2,
		Position: "2.1",
	})
	assert.NoError(t, err)

	err = a.Handle(&MoveCMD{
		PlayerID: p1,
		Position: "1.2",
	})
	assert.NoError(t, err)

	err = a.Handle(&MoveCMD{
		PlayerID: p2,
		Position: "2.2",
	})
	assert.NoError(t, err)

	err = a.Handle(&MoveCMD{
		PlayerID: p1,
		Position: "1.3",
	})
	assert.NoError(t, err)

	// End of moves, p1 won on previous step
	err = a.Handle(&MoveCMD{
		PlayerID: p2,
		Position: "2.3",
	})
	assert.Error(t, err)

	aggssert.Reproducible(t, a, NewTicTacToeAggregate())

	aggssert.ChangesSequence(t, a.Changes(),
		&GameCreated{
			FirstPlayerID: p1,
		},
		&SecondPlayerJoined{
			SecondPlayerID: p2,
		},
		&Moved{
			PlayerID: p1,
			Position: "1.1",
		},
		&Moved{
			PlayerID: p2,
			Position: "2.1",
		}, &Moved{
			PlayerID: p1,
			Position: "1.2",
		},
		&Moved{
			PlayerID: p2,
			Position: "2.2",
		}, &Moved{
			PlayerID: p1,
			Position: "1.3",
		},
		&GameFinish{
			WinnerPlayerID: p1,
			Positions:      []string{"1.1", "1.2", "1.3"},
		},
	)
}
