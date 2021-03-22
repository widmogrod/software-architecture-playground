package sat

import (
	"errors"
)

func NewSolver() *solver {
	return &solver{
		counter: 1,
		indexes: make(map[*BoolVar]int),
	}
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
	if len(ors) == 0 {
		return
	}

	s.closures = append(s.closures, ors)
	for _, prep := range ors {
		x := prep.Unwrap()
		if _, found := s.indexes[x]; !found {
			s.indexes[x] = s.counter
			s.counter++
		}
	}
}

func (s *solver) Solution() ([]Preposition, error) {
	t := NewDecisionTree()

	st := s.closures

	// TODO add findingout paradoxes like -7 or -7 (the same prep twice)

	n := 0
	candidate := s.candidatePrep(st)
	for {
		n += 1
		next, err := s.assumeThatSolves(candidate, t, st)
		if next != nil && len(next) == 0 {
			break
		}

		if err != nil {
			if err := t.Backtrack(); err != nil {
				return nil, err
			}
			candidate = t.ActiveBranch().prep

		} else {
			candidate = s.candidatePrep(next)
			st = next
		}
	}

	return t.Breadcrumbs(), nil
}

func (s *solver) assumeThatSolves(prep Preposition, t *DecisionTree, st Closures) (Closures, error) {
	if t.IsRoot(t.ActiveBranch()) || !prep.SameVar(t.ActiveBranch().prep) {
		t.CreateDecisionBranch(prep)
		t.ActivateBranch(prep)
	}

	next, err := s.filterLinesWith(prep, st)
	if err != nil {
		return st, err
	}

	return next, nil
}

func (s *solver) candidatePrep(closures Closures) Preposition {
	for _, line := range closures {
		if len(line) == 1 {
			return line[0]
		}
	}

	for _, line := range closures {
		for _, prep := range line {
			return prep
		}
	}

	return nil
}

var ErrFilter = errors.New("filterLinesWith")

func (s *solver) filterLinesWith(prep Preposition, st Closures) (Closures, error) {
	result := Closures{}
	for _, line := range st {
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
			result = append(result, newLines)
		} else if filterSim {
			return nil, ErrFilter
			//return nil, fmt.Errorf("filterLinesWith: in line=%d after filtering our similar, there is no other options to satisfy!  Backtrack (%s)!", no, prep.String())
		}
	}

	return result, nil
}
