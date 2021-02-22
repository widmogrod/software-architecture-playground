package amb

type Values struct {
	values []int
	len    int
	index  int
}

func (s *Values) Push(i int) {
	s.values = append(s.values, i)
	s.len++
}

func (s *Values) has() bool {
	return s.index < s.len
}

func (s *Values) Val() int {
	return s.values[s.index]
}

func (s *Values) next() {
	if s.has() {
		s.index++
	}
}

func (s *Values) end() bool {
	return s.index == s.len-1
}

func (s *Values) reset() {
	s.index = 0
}

type Permutations struct {
	el    []*Values
	len   int
	state int
	path  []func() bool
}

func (t *Permutations) With(xs ...*Values) {
	t.el = xs
	t.len = len(xs)
	t.state = 0
	t.path = nil
}

func (t *Permutations) Val() []int {
	res := make([]int, t.len)
	for i, stack := range t.el {
		res[i] = stack.Val()
	}
	return res
}

func (t *Permutations) reset() {
	t.state = t.len - 1
}

func (t *Permutations) Until(f func() bool) {
	t.path = append(t.path, f)
	for {
		found := true
		for _, f := range t.path {
			found = found && f()
		}

		if found {
			return
		}

		t.backtrack()
	}
}

func (t *Permutations) backtrack() {
	exhausted := true
	for i := t.state; i < t.len; i++ {
		exhausted = exhausted && t.el[i].end()
	}

	if exhausted && t.state == 0 {
		panic("cannot backtrack")
	}

	if exhausted {
		t.el[t.state-1].next()
		for i := t.state; i < t.len; i++ {
			t.el[i].reset()
		}
		t.state--
	} else {
		for i := t.len - 1; i >= t.state; i-- {
			if t.el[i].end() {
				for j := i; j < t.len; j++ {
					t.el[j].reset()
				}
			} else {
				t.el[i].next()
				return
			}
		}
	}
}

func NewRuntime() *Permutations {
	return &Permutations{}
}

func MkRange(start, stop int) *Values {
	result := &Values{}
	for i := start; i <= stop; i++ {
		result.Push(i)
	}

	return result
}
