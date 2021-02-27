package continuations

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
	"testing"
)

func TestSatSolve(t *testing.T) {
	a := sat.MkBool()
	b := sat.MkBool()
	c := sat.MkBool()

	solve := sat.NewSolver()
	solve.And(c)
	solve.And(sat.Not(a))
	solve.And(sat.Imply(a, sat.Not(b)))

	assert.Equal(t, []bool{false, false, true}, solve.Solution(a, b, c))
}

func TestExactlyOnce(t *testing.T) {
	a := sat.MkBool()
	b := sat.MkBool()
	c := sat.MkBool()
	d := sat.MkBool()

	solve := sat.NewSolver()
	closures := sat.ExactlyOne([]*sat.BoolVar{a, b, c, d})
	solve.AddClosures(closures)
	solve.PrintCNF()

	assert.Equal(t, []bool{false, false, false, true}, solve.Solution(a, b, c, d))
}
