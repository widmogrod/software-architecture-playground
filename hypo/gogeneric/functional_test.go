package gogeneric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCurry(t *testing.T) {
	add := Curry2(func(a, b int) int {
		return a + b
	})
	mul := Curry2(func(a, b int) int {
		return a * b
	})

	com := Compose(add(2), mul(10))

	assert.Equal(t, com(3), (3+2)*10)
}
