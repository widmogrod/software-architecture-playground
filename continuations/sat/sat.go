package sat

func NewSolver() *solver {
	return &solver{}
}

type Closures = [][]Preposition

type solver struct {
	closures Closures
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
}

// Solution SAT with DPLL algorithm
func (s *solver) Solution() ([]Preposition, error) {
	breadcrumbs, err := s.reduceSolution(s.selectCandidate(s.closures), s.closures, nil)
	if err != nil {
		return nil, err
	}

	return breadcrumbs, nil
}

func (s *solver) reduceSolution(prep Preposition, st Closures, agg []Preposition) ([]Preposition, error) {
	if len(st) == 0 {
		return agg, nil
	}

	next, emptyLine := s.filterOut(prep, st)
	if emptyLine {
		// Backtrack, take another route
		return s.reduceSolution(prep.Not(), st, agg)
	}

	return s.reduceSolution(
		s.selectCandidate(next),
		next,
		append(agg, prep),
	)
}

func (s *solver) selectCandidate(closures Closures) Preposition {
	// This line plays very important role,
	// it sorts and select candidates that are easiest to prove
	// which is - single preposition in a line -
	// always must be true whole constraing be true
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

func (s *solver) filterOut(prep Preposition, st Closures) (Closures, bool) {
	result := Closures{}
	for _, line := range st {
		var emptyLine bool
		var newLines []Preposition
		for _, prep2 := range line {
			if prep2.Equal(prep) {
				newLines = nil
				break
			}

			if !prep2.SameVar(prep) {
				newLines = append(newLines, prep2)
			} else {
				emptyLine = true
			}
		}

		if newLines != nil {
			result = append(result, newLines)
		} else if emptyLine {
			return nil, true
		}
	}

	return result, false
}
