package sat

import (
	"fmt"
)

func NewSolver() *solver {
	return &solver{
		counter: 1,
		indexes: make(map[*BoolVar]int),
	}
}

var counter = 0

func MkBool() *BoolVar {
	counter += 1
	return &BoolVar{
		no: counter,
	}
}

func MkBoolC(no int) *BoolVar {
	return &BoolVar{
		no: no,
	}
}

type Preposition interface {
	Not() Preposition
	IsTrue() bool
	Unwrap() *BoolVar
	Equal(prep Preposition) bool
	SameVar(x Preposition) bool
	String() string
	No() int
}

var _ Preposition = &BoolVar{}

type BoolVar struct {
	no int
}

func (b *BoolVar) No() int {
	return b.no
}

func (b *BoolVar) String() string {
	return fmt.Sprintf("%d", b.no)
}

func (b *BoolVar) SameVar(x Preposition) bool {
	return b == x.Unwrap()
}

func (b *BoolVar) Not() Preposition {
	return &negation{b}
}

func (b *BoolVar) IsTrue() bool {
	return true
}

func (b *BoolVar) Unwrap() *BoolVar {
	return b
}

func (b *BoolVar) Equal(prep Preposition) bool {
	return b.SameVar(prep) && b.IsTrue() == prep.IsTrue()
}

type negation struct {
	b Preposition
}

func (n *negation) No() int {
	return n.Unwrap().No()
}

func (n *negation) String() string {
	if n.IsTrue() != n.Unwrap().IsTrue() {
		return "-" + n.Unwrap().String()
	}

	return n.Unwrap().String()
}

func (n *negation) SameVar(x Preposition) bool {
	return n.Unwrap().SameVar(x)
}

func (n *negation) Not() Preposition {
	return n.b
}

func (n *negation) IsTrue() bool {
	return !n.b.IsTrue()
}

func (n *negation) Unwrap() *BoolVar {
	return n.b.Unwrap()
}

func (n *negation) Equal(prep Preposition) bool {
	return n.SameVar(prep) && n.IsTrue() == prep.IsTrue()
}

type Closures = [][]Preposition

type solver struct {
	closures Closures
	counter  int
	indexes  map[*BoolVar]int
}

func (s *solver) AddClosures(c Closures) {
	for _, line := range c {
		s.And(line...)
	}
}

func (s *solver) And(ors ...Preposition) {
	s.closures = append(s.closures, ors)
	for _, prep := range ors {
		x := prep.Unwrap()
		if _, found := s.indexes[x]; !found {
			s.indexes[x] = s.counter
			s.counter++
		}
	}
}

func (s *solver) PrintCNF() {
	//fmt.Printf("p cnf %d %d\n", s.counter-1, len(s.closures))
	fmt.Print(s.printClosures(s.closures))
}
func (s *solver) printClosures(closures Closures) string {
	result := ""
	for _, line := range closures {
		result += fmt.Sprintf("%s \n", s.printPrepositions(line))
	}

	return result
}

func (s *solver) printPrepositions(line []Preposition) string {
	result := ""
	count := len(line)
	for i := 0; i < count; i++ {
		if i > 0 && i < count {
			result += " "
		}

		result += s.printPreposition(line[i])
	}

	return result
}

func (s *solver) printPreposition(prep Preposition) string {
	return prep.String()
}

func (s *solver) Solution() []Preposition {
	t := NewDecisionTree()

	st := &State{
		closures: s.closures,
	}

	candidate := s.candidatePrep(st)
	s.assumeThatSolves(candidate, t, st)

	fmt.Println("Solutions:")
	t.Print()

	return t.Breadcrumbs()
}

type State struct {
	closures Closures
}

// lets remove variable from lines
//
// Prep that we're filtering out must satisfy!
//
// On input: -2
// 	1 -2 3
// 	2 3
// 	3
//
// Result should be
//	1 _ 3
//  _ 3
//  3

// -1 solved for variable 1=false
// 1			 variable 1=true
func (s *solver) candidatePrep(st *State) Preposition {
	for _, line := range st.closures {
		for _, prep := range line {
			return prep
		}
	}

	return nil
}

func (s *solver) filterLinesWith(prep Preposition, st *State) (*State, error) {
	result := &State{}
	for no, line := range st.closures {
		var filterSim bool
		var newLines []Preposition
		for _, prep2 := range line {
			if prep2.Equal(prep) {
				newLines = nil
				break
			}

			if !prep2.SameVar(prep) {
				newLines = append(newLines, prep2)
			} else {
				filterSim = true
			}
		}

		if newLines != nil {
			result.closures = append(result.closures, newLines)
		} else if filterSim {
			return nil, fmt.Errorf("filterLinesWith: in line=%d after filtering our similar, there is no other options to satisfy!  Backtrack (%s)!", no, prep.String())
		}
	}

	return result, nil
}

func (s *solver) isUnsat(prep Preposition, st *State) bool {
	if !prep.IsTrue() {
		panic(fmt.Sprintf("checkUnsat: preposition %s id not true", s.printPreposition(prep)))
	}

	for _, line := range st.closures {
		l := len(line)
		for _, prep2 := range line {
			// belong to the same variable
			if prep2.SameVar(prep2) && !prep2.IsTrue() {
				if l == 1 {
					return true
				}
			}
		}
	}

	return false
}

func (s *solver) assumeThatSolves(prep Preposition, t *DecisionTree, st *State) {
	if len(st.closures) == 0 {
		return
	}

	t.CreateDecisionBranch(prep)
	t.ActivateBranch(prep)

	//fmt.Println("PATH:", t.Breadcrumbs())
	//t.Print()

	next, err := s.filterLinesWith(prep, st)

	//fmt.Println("AFTER:")
	//fmt.Println(s.printClosures(next.closures))

	if err != nil {
		t.Backtrack()
		candidate := t.ActiveBranch().prep
		s.assumeThatSolves(candidate, t, st)
		return
	}

	candidate := s.candidatePrep(next)
	s.assumeThatSolves(candidate, t, next)
}

func (s *solver) lineHasPrep(line []Preposition, prep Preposition) bool {
	for _, prep2 := range line {
		if prep2.Equal(prep) {
			return true
		}
	}

	return false
}

func Not(c Preposition) Preposition {
	return c.Not()
}

func OneOf(vars []*BoolVar) []Preposition {
	var result = make([]Preposition, len(vars))
	for i := 0; i < len(vars); i++ {
		result[i] = vars[i]
	}

	return result
}

// X1 or X2 and (!X1 or !X2)
func ExactlyOne(vars []*BoolVar) Closures {
	var closures Closures
	closures = append(closures, OneOf(vars))

	size := len(vars)
	for i := 0; i < size; i++ {
		for j := 1; j < size; j++ {
			closures = append(closures, []Preposition{
				Not(vars[i]),
				Not(vars[j]),
			})
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
