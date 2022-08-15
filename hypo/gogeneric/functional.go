package gogeneric

func Partial2[A, B, C any](f func(A, B) C, x A) func(B) C {
	return func(y B) C {
		return f(x, y)
	}
}

func Curry2[A, B, C any](f func(A, B) C) func(A) func(B) C {
	return func(x A) func(B) C {
		return func(y B) C {
			return f(x, y)
		}
	}
}

func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C {
		return g(f(a))
	}
}

func Identity[T any](x T) T {
	return x
}

// cannot convert x (variable of type T1 constrained by any) to type T2
//func Identity2[T1, T2 any](x T1) T2 {
//	return T2(x)
//}

type Container[T any] interface {
	*T
}

type Functor[T any] interface {
	Join() T
}

// err: generic type cannot be alias
//type MapFunc[T1, T2 any] = func(T1) T2

type Functor2[T1, T2 any] interface {
	Return(T2) Functor[T2]
	Map(func(T1) T2) Functor[T2]
}

type Functor3[T any, T1 Functor[T], T2 any] interface {
	*T1
	Map(func(T1) T2) Functor[T2]
}

//func Map[T1, T2 any](f func(T1) T2, x Functor[T1]) Functor[T2] {
//	y := Functor2[T1, T2]{x}
//return y.Return(f(x.Join()))
//}

type IdentityM[T1 any] struct {
	v T1
}

func (i *IdentityM[T1]) Return(t T1) Functor[T1] {
	return &IdentityM[T1]{v: t}
}

func (i *IdentityM[T1]) Join() T1 {
	return i.v
}

//var _ Functor[int] = IdentityM[int]

type ProxyF[T any, T1 Functor[T], T2 any] struct {
	v T1
}

func (p ProxyF[T, T1, T2]) Map(f func(T) T2) Functor[T2] {
	return &IdentityM[T2]{v: f(p.v.Join())}
}
