package churchencoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"testing/quick"
)

func TestCalculator(t *testing.T) {
	var calculation Calc = _Mul(
		_Lit(2),
		_Add(
			_Lit(2),
			_Lit(3),
		),
	)

	assert.Equal(t, 10, eval(calculation))

	_ = quick.Check(func(v int) bool {
		t.Log(print(generate(v)))
		return assert.Equal(t, v, eval(generate(v)))
	}, nil)
}
