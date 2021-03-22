package sat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepositions(t *testing.T) {
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
