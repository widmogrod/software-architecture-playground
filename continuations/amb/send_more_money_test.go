package amb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
Solution to a problem send+more == money as described here
[Alphametic Page](https://www.math.uni-bielefeld.de/~sillke/PUZZLES/ALPHAMETIC/alphametic-mike-keith.html)
*/
func TestSendMoreMoney(t *testing.T) {
	s := MkRange(0, 9)
	e := MkRange(0, 9)
	n := MkRange(0, 9)
	d := MkRange(0, 9)

	m := MkRange(1, 1)
	o := MkRange(0, 9)
	r := MkRange(0, 9)

	y := MkRange(0, 9)

	ctx := NewRuntime()
	ctx.With(s, e, n, d, m, o, r, y)

	ctx.Until(func() bool {
		send := (s.Val() * 1000) + (e.Val() * 100) + (n.Val() * 10) + d.Val()
		more := (m.Val() * 1000) + (o.Val() * 100) + (r.Val() * 10) + e.Val()
		money := (m.Val() * 10000) + (o.Val() * 1000) + (n.Val() * 100) + (e.Val() * 10) + y.Val()
		return send+more == money
	})

	// Ensure each numbers are unique
	ctx.Until(func() bool {
		unique := map[int]bool{}
		for _, i := range ctx.Val() {
			if _, found := unique[i]; found {
				return false
			}
			unique[i] = true
		}

		return true
	})

	assert.Equal(t, 1, m.Val())
	assert.Equal(t, 0, o.Val())
	assert.Equal(t, 6, n.Val())
	assert.Equal(t, 5, e.Val())
	assert.Equal(t, 2, y.Val())
}
