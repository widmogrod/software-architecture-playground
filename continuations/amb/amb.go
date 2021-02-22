package amb

type stack struct {
	values []int
	len    int
	index  int
}

func (s *stack) push(i int) {
	s.values = append(s.values, i)
	s.len++
}

func (s *stack) has() bool {
	return s.index < s.len
}

func (s *stack) Val() int {
	return s.values[s.index]
}

func (s *stack) next() {
	if s.has() {
		s.index++
	}
}

func (s *stack) end() bool {
	return s.index == s.len-1
}

func (s *stack) reset() {
	s.index = 0
}

type timemachine2 struct {
	el    []*stack
	len   int
	state int
	path  []func() bool
}

func (t *timemachine2) With(xs ...*stack) {
	t.el = xs
	t.len = len(xs)
	t.state = 0
	t.path = nil
}

func (t *timemachine2) Val() []int {
	res := make([]int, t.len)
	for i, stack := range t.el {
		res[i] = stack.Val()
	}
	return res
}

func (t *timemachine2) reset() {
	t.state = t.len - 1
}

func (t *timemachine2) Until(f func() bool) {
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

func (t *timemachine2) backtrack() {
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

func NewRuntime() *timemachine2 {
	return &timemachine2{}
}

func MkRange(start, stop int) *stack {
	result := &stack{}
	for i := start; i <= stop; i++ {
		result.push(i)
	}

	return result
}
