package sat

import (
	"github.com/widmogrod/software-architecture-playground/continuations/amb"
)

func NewSolver() *solver {
	return &solver{
		ands: nil,
	}
}

func MkBool() *boolean {
	result := &amb.Values{}
	result.Push(-1)
	result.Push(1)
	return &boolean{
		result,
	}
}

type Booler interface {
	Not() Booler
	IsTrue() bool
}

type boolean struct {
	v *amb.Values
}

func (b *boolean) IsTrue() bool {
	return isTrue(b.v.Val())
}

func (b *boolean) Not() Booler {
	return &negation{b}
}

type negation struct {
	b Booler
}

func (n *negation) Not() Booler {
	return n.b
}

func (n *negation) IsTrue() bool {
	return !n.b.IsTrue()
}

type implication struct {
	p, q Booler
}

func (i *implication) Not() Booler {
	return &negation{i}
}

func (i implication) IsTrue() bool {
	if i.q.IsTrue() {
		return true
	}

	if !i.p.IsTrue() && !i.q.IsTrue() {
		return true
	}

	return false
}

type solver struct {
	ands []func() bool
}

//func getValue(value Booler) []*amb.Values {
//	var result []*amb.Values
//	switch v := value.(type) {
//	case *boolean:
//		result = append(result, v.v)
//	case *negation:
//		result = append(result, getValue(v.b)...)
//	case *implication:
//		result = append(result, getValue(v.p)...)
//		result = append(result, getValue(v.q)...)
//	default:
//		panic(fmt.Sprintf("unknow type %T", value))
//	}
//
//	return result
//}

func (s *solver) And(or ...Booler) {
	fn := func() bool {
		for _, value := range or {
			if value.IsTrue() {
				return true
			}
		}

		return false
	}

	s.ands = append(s.ands, fn)
}

func (s *solver) Solution(values ...*boolean) []bool {
	ctx := amb.NewRuntime()

	var xs []*amb.Values
	for _, v := range values {
		xs = append(xs, v.v)
	}

	ctx.With(xs...)

	for _, and := range s.ands {
		ctx.Until(and)
	}

	solutions := ctx.Val()
	result := make([]bool, len(values))
	for i, value := range solutions {
		result[i] = isTrue(value)
	}
	return result
}

func Not(c Booler) Booler {
	return c.Not()
}

func Imply(p, q Booler) Booler {
	return &implication{p, q}
}

func isTrue(x int) bool {
	return x == 1
}
