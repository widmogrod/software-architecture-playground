package sat

import "fmt"

var counter = 0

func MkBool() *BoolVar {
	counter += 1
	return &BoolVar{
		no: counter,
	}
}

func MkLit(no int) *BoolVar {
	return &BoolVar{
		no: no,
	}
}

func MkPrep(no int) Preposition {
	if no < 0 {
		return Not(&BoolVar{
			no: -no,
		})
	} else {
		return &BoolVar{
			no: no,
		}
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
	return b.No() == x.Unwrap().No()
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
