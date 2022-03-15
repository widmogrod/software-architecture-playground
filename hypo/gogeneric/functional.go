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
