package continuations

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/continuations/amb"
	"testing"
)

/*
Solution to a problem send+more == money as described here
[Alphametic Page](https://www.math.uni-bielefeld.de/~sillke/PUZZLES/ALPHAMETIC/alphametic-mike-keith.html)
*/
func TestSendMoreMoney(t *testing.T) {
	s := amb.MkRange(0, 9)
	e := amb.MkRange(0, 9)
	n := amb.MkRange(0, 9)
	d := amb.MkRange(0, 9)

	m := amb.MkRange(1, 1)
	o := amb.MkRange(0, 9)
	r := amb.MkRange(0, 9)

	y := amb.MkRange(0, 9)

	ctx := amb.NewRuntime()
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
