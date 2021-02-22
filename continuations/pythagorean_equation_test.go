package continuations

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/continuations/amb"
	"testing"
)

func TestPythagoreanEquation(t *testing.T) {
	a := amb.MkRange(1, 7)
	b := amb.MkRange(1, 7)
	c := amb.MkRange(1, 7)

	ctx := amb.NewRuntime()
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
