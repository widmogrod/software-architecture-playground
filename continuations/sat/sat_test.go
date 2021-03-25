package sat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMkLit(t *testing.T) {
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
	assert.Equal(t, result, []Preposition{d, a, b})
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
	assert.Equal(t, result, []Preposition{a, b.Not(), c.Not(), d.Not()})
}

func TestSat4(t *testing.T) {
	sat := NewSolver()
	sat.AddClosures(ExactlyOne(Num(1, 2, 3, 4)))

	result, err := sat.Solution()
	assert.NoError(t, err)
	assert.Equal(t, result, Num(1, -2, -3, -4))
}
