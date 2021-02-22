package continuations

import (
	"fmt"
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

	fmt.Println(solve.Solution(a, b, c))
}
