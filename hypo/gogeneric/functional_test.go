package gogeneric

import (
	"github.com/stretchr/testify/assert"
	"strconv"
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

//func FunctorLaw[T1, T2, T3 any](t *testing.T, f func(T1) T2, g func(T2) T3, m Functor[T1, T3]) {
//	// identity: fmap id  ==  id
//	assert.Equal(t, m.Map(Identity[T1]), m)
//
//	// composition: fmap (f . g)  ==  fmap f . fmap g
//
//	assert.Equal(t, m.Map(Compose(f, g)), m.Map(f).Map(g))
//}

func TestMap(t *testing.T) {
	f := func(x int) string {
		return strconv.Itoa(x) + "+ 2"
	}
	g := func(x string) float64 {
		return float64(len(x)) + 0.2
	}
	//f := func(x int) int {
	//	return x + 2
	//}
	//g := func(x int) int {
	//	return x + 4
	//}

	Compose(f, g)(1)

	m := &IdentityM[int]{v: 1}
	p := ProxyF[int, *IdentityM[int], int]{v: m}
	// identity: fmap id  ==  id

	assert.Equal(t, m, p.Map(func(x int) int {
		return x
	}))

	// composition: fmap (f . g)  ==  fmap f . fmap g
	assert.Equal(t,
		ProxyF[int, Functor[int], float64]{v: m}.Map(Compose(f, g)),
		ProxyF[string, Functor[string], float64]{v: ProxyF[int, *IdentityM[int], string]{m}.Map(f)}.Map(g),
	)
}
