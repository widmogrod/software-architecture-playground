package tictactoeaggregate

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestMinmax(t *testing.T) {
	moves := []Move{
		MkMove(1, 1), MkMove(1, 2),
		MkMove(3, 3), MkMove(2, 1),
	}

	buf := strings.Builder{}
	PrintGameRC(&buf, ToMovesTaken(moves), 3, 3)
	t.Log(buf.String())

	move := NextMoveNaive(ToMovesTaken(moves), 3, 3, 3)

	buf2 := strings.Builder{}
	PrintGameRC(&buf2, ToMovesTaken(append(moves, move)), 3, 3)
	t.Log(buf2.String())

	move2 := NextMoveMinMax(moves, 3, 3)
	assert.NotEqual(t, "", move2)

	buf3 := strings.Builder{}
	PrintGameRC(&buf3, ToMovesTaken(append(moves, move2)), 3, 3)
	t.Log(buf3.String())

}

func TestMinmax2(t *testing.T) {
	moves := []Move{
		MkMove(1, 1), MkMove(1, 2),
		MkMove(3, 3), MkMove(2, 2),
	}

	buf := strings.Builder{}
	PrintGameRC(&buf, ToMovesTaken(moves), 3, 3)
	t.Log(buf.String())

	move2 := NextMoveMinMax(moves, 3, 3)

	buf3 := strings.Builder{}
	PrintGameRC(&buf3, ToMovesTaken(append(moves, move2)), 3, 3)
	t.Log(buf3.String())

	assert.Equal(t, "3.2", move2)
}
