package churchencoding

import (
	"fmt"
	"math/rand"
)

type calc = interface{}

type (
	Lit  = func(int) calc
	Add  = func(calc, calc) calc
	Mul  = func(calc, calc) calc
	Calc = func(Lit, Add, Mul) calc
)

func _Lit(v int) Calc {
	return func(lit Lit, add Add, mul Mul) calc {
		return lit(v)
	}
}

func _Mul(a, b Calc) Calc {
	return func(lit Lit, add Add, mul Mul) calc {
		return mul(a(lit, add, mul), b(lit, add, mul))
	}
}

func _Add(a, b Calc) Calc {
	return func(lit Lit, add Add, mul Mul) calc {
		return add(a(lit, add, mul), b(lit, add, mul))
	}
}

func eval(c Calc) int {
	return c(func(i int) calc {
		return i
	}, func(a, b calc) calc {
		return a.(int) + b.(int)
	}, func(a, b calc) calc {
		return a.(int) * b.(int)
	}).(int)
}

func print(c Calc) string {
	return c(func(i int) calc {
		return fmt.Sprintf("%d", i)
	}, func(a, b calc) calc {
		return fmt.Sprintf("(%s + %s)", a, b)
	}, func(a, b calc) calc {
		return fmt.Sprintf("(%s * %s)", a, b)
	}).(string)
}

func generate(v int) Calc {
	if v > 100 || v < -100 {
		sub := 2 + rand.Int()%8
		if v%sub == 0 {
			return _Mul(_Lit(sub), generate(v/sub))
		}

		return _Add(generate(v-sub), _Lit(sub))
	}

	return _Lit(v)
}
