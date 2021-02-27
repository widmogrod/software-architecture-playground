package sat

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/amb"
)

func NewSolver() *solver {
	return &solver{
		counter: 1,
		indexes: make(map[*BoolVar]int),
	}
}

func MkBool() *BoolVar {
	result := &amb.Value{}
	result.Push(-1)
	result.Push(1)
	return &BoolVar{
		result,
	}
}

type Preposition interface {
	Not() Preposition
	IsTrue() bool
}

var _ Preposition = &BoolVar{}

type BoolVar struct {
	v *amb.Value
}

func (b *BoolVar) Not() Preposition {
	return &negation{b}
}

func (b *BoolVar) IsTrue() bool {
	return isTrue(b.v.Val())
}

type negation struct {
	b Preposition
}

func (n *negation) Not() Preposition {
	return n.b
}

func (n *negation) IsTrue() bool {
	return !n.b.IsTrue()
}

type implication struct {
	p, q Preposition
}

func (i *implication) Not() Preposition {
	return &negation{i}
}

func (i *implication) IsTrue() bool {
	if i.q.IsTrue() {
		return true
	}

	if !i.p.IsTrue() && !i.q.IsTrue() {
		return true
	}

	return false
}

type Closure = [][]Preposition

type solver struct {
	closures Closure
	counter  int
	indexes  map[*BoolVar]int
}

func (s *solver) And(ors ...Preposition) {
	s.closures = append(s.closures, ors)
}

func (s *solver) AddClosures(closures Closure) {
	for _, ors := range closures {
		s.closures = append(s.closures, ors)
		for _, or := range ors {
			// create variable number
			if x, ok := or.(*BoolVar); ok {
				if _, found := s.indexes[x]; !found {
					s.indexes[x] = s.counter
					s.counter++
				}
			}
		}

	}
}

func (s *solver) PrintCNF() {
	fmt.Printf("p cnf %d %d\n", s.counter-1, len(s.closures))
	for _, ors := range s.closures {
		fmt.Printf("%s 0\n", s.printPrepositions(ors))
	}
}

func (s *solver) printPrepositions(ors []Preposition) string {
	result := ""
	count := len(ors)
	for i := 0; i < count; i++ {
		if i > 0 && i < count {
			result += " "
			//result += " \u2228 "
		}

		result += s.printPreposution(ors[i])
		//result += "(" + s.printPreposution(ors[i]) + ")"
	}

	return result
}

func (s *solver) printPreposution(prep Preposition) string {
	switch or := prep.(type) {
	case *BoolVar:
		return fmt.Sprintf("%d", s.indexes[or])
		//return fmt.Sprintf("x%d", s.indexes[or])
	case *negation:
		return fmt.Sprintf("-%s", s.printPreposution(or.b))
	case *implication:
		return fmt.Sprintf("%s → %s", s.printPreposution(or.p), s.printPreposution(or.q))
	}

	return "unknown"
}

func (s *solver) Solution(values ...*BoolVar) []bool {
	ctx := amb.NewRuntime()

	var xs []*amb.Value
	for _, v := range values {
		xs = append(xs, v.v)
	}

	ctx.With(xs...)

	for _, ors := range s.closures {
		ctx.Until(oneOfORs(ors))
	}

	solutions := ctx.Val()
	result := make([]bool, len(values))
	for i, value := range solutions {
		result[i] = isTrue(value)
	}
	return result
}

func oneOfORs(ors []Preposition) func() bool {
	return func() bool {
		for _, value := range ors {
			if value.IsTrue() {
				return true
			}
		}

		return false
	}
}

func Not(c Preposition) Preposition {
	return c.Not()
}

func Imply(p, q Preposition) Preposition {
	return &implication{p, q}
}

func isTrue(x int) bool {
	return x == 1
}

func OneOf(vars []*BoolVar) []Preposition {
	var result = make([]Preposition, len(vars))
	for i := 0; i < len(vars); i++ {
		result[i] = vars[i]
	}

	return result
}

// X1 or X2 and (!X1 or 1X2)
func ExactlyOne(vars []*BoolVar) Closure {
	var closures Closure
	closures = append(closures, OneOf(vars))

	size := len(vars) - 1
	for i := -1; i < size; i++ {
		if i == -1 {
			pair := []Preposition{
				Not(vars[size]),
				Not(vars[0]),
			}
			closures = append(closures, pair)
		} else {
			pair := []Preposition{
				Not(vars[i]),
				Not(vars[i+1]),
			}
			closures = append(closures, pair)
		}

	}

	return closures
}

func Take(vars []*BoolVar, index int, len int) []*BoolVar {
	var result []*BoolVar
	for i := index; i < index+len; i++ {
		result = append(result, vars[i])
	}

	return result
}
