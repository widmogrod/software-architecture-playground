package tictactoeaggregate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWining(t *testing.T) {
	useCases := map[string]struct {
		sequence []Move
		move     Move
		playerID PlayerID
		want     bool
	}{
		"winning sequence": {
			sequence: []Move{
				"1.1", "2.2", "1.2", "2.3",
			},
			move:     "1.3",
			playerID: "1",
			want:     true,
		},
		"winning sequence 2": {
			sequence: []Move{
				"1.1", "2.2", "1.2", "3.1",
			},
			move:     "",
			playerID: "1",
			want:     false,
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			_, win := CheckIfMoveWin(
				uc.sequence,
				uc.move,
				uc.playerID,
			)
			assert.Equal(t, uc.want, win)
		})
	}
}
