package amb

type Value struct {
	values []int
	len    int
	index  int
}

func (s *Value) Push(i int) {
	s.values = append(s.values, i)
	s.len++
}

func (s *Value) has() bool {
	return s.index < s.len
}

func (s *Value) Val() int {
	return s.values[s.index]
}

func (s *Value) next() {
	if s.has() {
		s.index++
	}
}

func (s *Value) end() bool {
	return s.index == s.len-1
}

func (s *Value) reset() {
	s.index = 0
}

type Permutations struct {
	el   []*Value
	len  int
	path []func() bool
}

func (t *Permutations) With(xs ...*Value) {
	t.el = xs
	t.len = len(xs)
	t.path = nil
}

func (t *Permutations) Val() []int {
	res := make([]int, t.len)
	for i, stack := range t.el {
		res[i] = stack.Val()
	}
	return res
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
	for i := 0; i < t.len; i++ {
		exhausted = exhausted && t.el[i].end()
		if !exhausted {
			break
		}
	}

	if exhausted {
		panic("cannot backtrack")
	}

	for i := t.len - 1; i >= 0; i-- {
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

func NewRuntime() *Permutations {
	return &Permutations{}
}

func MkRange(start, stop int) *Value {
	result := &Value{}
	for i := start; i <= stop; i++ {
		result.Push(i)
	}

	return result
}
