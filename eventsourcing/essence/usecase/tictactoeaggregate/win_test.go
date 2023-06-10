package tictactoeaggregate

import (
	"github.com/stretchr/testify/assert"
	"strings"
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
				"1.1", "2.2", "1.2", "2.3", "1.3",
			},
			playerID: "1",
			want:     true,
		},
		"winning sequence 2": {
			sequence: []Move{
				"1.1", "2.2", "1.2", "3.1",
			},
			playerID: "1",
			want:     false,
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			_, win := CheckIfMoveWin(
				uc.sequence,
				Wining3x3(),
			)
			assert.Equal(t, uc.want, win)
		})
	}
}

func TestGeneratedPositions(t *testing.T) {
	assert.Equal(t, Wining3x3(), GenerateWiningPositions(3, 3, 3))

	rows, cols := 5, 5
	for _, seq := range GenerateWiningPositions(3, rows, cols) {
		mm := map[Move]PlayerID{}
		for _, m := range seq {
			mm[m] = "1"
		}
		buf := strings.Builder{}
		PrintGameRC(&buf, mm, rows, cols)
		t.Log(buf.String())
	}
}

func TestMoveOrder(t *testing.T) {
	moves := []Move{
		"1.1", "1.2", "1.3", "1.4", "1.5", "2.1", "2.2", "2.3", "2.4", "3.2",
	}
	mm := ToMovesTaken(moves)

	buf := strings.Builder{}
	PrintGameRC(&buf, mm, 5, 5)
	t.Log(buf.String())

	_, result := CheckIfMoveWin(moves, GenerateWiningPositions(3, 5, 5))
	assert.Equal(t, true, result)
}
