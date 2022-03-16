package gogeneric

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/hypo/gogeneric/x/lists"
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
	var l lists.List[int]
	l.Push(1)

	//f := func(x int) string {
	//	return strconv.Itoa(x) + "+ 2"
	//}
	//g := func(x string) string {
	//	return x + "4"
	//}
	f := func(x int) int {
		return x + 2
	}
	g := func(x int) int {
		return x + 4
	}

	Compose(f, g)(1)

	m := &IdentityM[int]{v: 1}
	p := ProxyF[int, *IdentityM[int], int]{v: m}
	// identity: fmap id  ==  id

	assert.Equal(t, m, p.Map(Identity[int]))

	// composition: fmap (f . g)  ==  fmap f . fmap g

	//assert.Equal(t, p.Map(Compose(f, g)), p.Map(f).Map(g)))
}
