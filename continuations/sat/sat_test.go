package sat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoolVar(t *testing.T) {
	a := MkBool()

	assert.True(t, a.IsTrue())
	assert.True(t, Not(Not(a)).IsTrue())
	assert.False(t, Not(a).IsTrue())

	// Same variable that translate to the same result can be equal
	assert.True(t, a.Equal(a))
	assert.True(t, Not(a).Equal(Not(a)))
	assert.True(t, Not(Not(a)).Equal(a))
	assert.False(t, a.Equal(Not(a)))
	assert.False(t, Not(a).Equal(a))

	// Two different variables cannot be equal
	b := MkBool()
	assert.False(t, a.Equal(b))
	assert.False(t, Not(a).Equal(b))
	assert.False(t, Not(a).Equal(Not(b)))

	// Same
	assert.True(t, a.SameVar(a))
	assert.True(t, a.SameVar(Not(a)))
	assert.True(t, Not(a).SameVar(Not(a)))
	assert.True(t, Not(a).SameVar(a))

	assert.False(t, b.SameVar(a))
	assert.False(t, b.SameVar(Not(a)))
	assert.False(t, Not(b).SameVar(Not(a)))
	assert.False(t, Not(b).SameVar(a))
}

func TestBoolVarC(t *testing.T) {
	a := MkLit(1)
	b := MkLit(1)
	c := MkLit(2)

	assert.True(t, a.Equal(b))
	assert.True(t, Not(a).Equal(Not(b)))

	assert.True(t, a.SameVar(b))
	assert.False(t, a.SameVar(c))
	assert.False(t, Not(a).SameVar(c))
}

func TestSat1(t *testing.T) {
	a := MkBool()
	b := MkBool()

	sat := NewSolver()
	sat.And(a, Not(b))

	result, err := sat.Solution()
	assert.NoError(t, err)
	assert.Equal(t, result, []Preposition{a})
}

//	a -b  c
//	a  b -c
//	   b -c
// 			 d
func TestSat2(t *testing.T) {
	a := MkBool()
	b := MkBool()
	c := MkBool()
	d := MkBool()

	sat := NewSolver()
	sat.And(a, Not(b), c)
	sat.And(a, b, Not(c))
	sat.And(b, Not(c))
	sat.And(d)

	result, err := sat.Solution()
	assert.NoError(t, err)
	assert.Equal(t, result, []Preposition{b, a, d})
}

func TestSat3(t *testing.T) {
	a := MkBool()
	b := MkBool()
	c := MkBool()
	d := MkBool()

	sat := NewSolver()
	sat.AddClosures(ExactlyOne([]Preposition{a, b, c, d}))

	result, err := sat.Solution()
	assert.NoError(t, err)
	assert.Equal(t, result, []Preposition{d.Not(), c.Not(), b.Not(), a})
}

func TestSat4(t *testing.T) {
	sat := NewSolver()
	sat.AddClosures(ExactlyOne(Num(1, 2, 3, 4)))

	result, err := sat.Solution()
	assert.NoError(t, err)
	assert.Equal(t, result, Num(-4, -3, -2, 1))
}
