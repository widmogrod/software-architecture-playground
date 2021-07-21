package tictactoeaggregate

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/aggregate/aggssert"
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

func GenerateNextCMD(state *TicTacToeState) interface{} {
	// introspect from state what are next actions
	// generate next states - they may be random,
	// but some fields or information may depend on previous context
	//if state.OneOf.GameWaitingForPlayer != nil {
	//
	//} else if state.OneOf.GameProgress != nil {
	//
	//} else if state.OneOf.GameResult != nil {
	//
	//}

	switch rand.Int() % 4 {
	case 0:
		return &CreateGameCMD{
			FirstPlayerID: uuid.Must(uuid.NewUUID()).String(),
		}
	case 1:
		return &JoinGameCMD{
			SecondPlayerID: uuid.Must(uuid.NewUUID()).String(),
		}
	case 2:
		return &StartGameCMD{
			FirstPlayerID:  uuid.Must(uuid.NewUUID()).String(),
			SecondPlayerID: uuid.Must(uuid.NewUUID()).String(),
		}
	case 3:
		if state != nil && state.OneOf.GameProgress != nil {
			for move := range state.OneOf.GameProgress.AvailableMoves {
				return &MoveCMD{
					PlayerID: state.OneOf.GameProgress.NextMovePlayerID,
					Position: move,
				}
			}
		}

		return &MoveCMD{
			PlayerID: uuid.Must(uuid.NewUUID()).String(),
			Position: uuid.Must(uuid.NewUUID()).String(),
		}
	}

	panic("GenerateNextCMD undefined")
}

// TODO complete implementation
func TestTicTacToeAggregate_StateDynamic(t *testing.T) {
	// Generate States n-times
	// Collect probabilities of success transitions
	// Display causation graph
	// Assert What states are expected, and which are not

	for range rand.Perm(100) {
		var commands []interface{}

		a := NewTicTacToeAggregate()
		for {
			cmd := GenerateNextCMD(a.state)
			err := a.Handle(cmd)
			if err != nil {
				break
			}

			commands = append(commands, cmd)
		}

		fmt.Printf("Commands: %d %#v \n", len(commands), commands)
	}
}

// Dependent events
// P(A^B) = P(A) * P(B|A)

/*
	| A | B | C |
  ------------------
  A |
  B |
  C |
  ------------------
*/
