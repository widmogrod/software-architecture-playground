package amb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPythagoreanEquation(t *testing.T) {
	a := MkRange(1, 7)
	b := MkRange(1, 7)
	c := MkRange(1, 7)

	ctx := NewRuntime()
	ctx.With(a, b, c)
	ctx.Until(func() bool {
		return (a.Val()*a.Val())+(b.Val()*b.Val()) == (c.Val() * c.Val())
	})
	ctx.Until(func() bool {
		return a.Val() > b.Val()
	})

	assert.Equal(t, 4, a.Val())
	assert.Equal(t, 3, b.Val())
	assert.Equal(t, 5, c.Val())
}
