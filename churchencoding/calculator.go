package churchencoding

type calc = interface{}

type (
	Add  = func(calc, calc) calc
	Mul  = func(calc, calc) calc
	Calc = func(Add, Mul) interface{}
)

func _Lit(v int) Calc {
	return func(add Add, mul Mul) calc {
		return v
	}
}

func _Mul(a, b Calc) Calc {
	return func(add Add, mul Mul) calc {
		return mul(a(add, mul), b(add, mul))
	}
}

func _Add(a, b Calc) Calc {
	return func(add Add, mul Mul) calc {
		return add(a(add, mul), b(add, mul))
	}
}

func eval(c Calc) int {
	return c(func(a, b calc) calc {
		return a.(int) + b.(int)
	}, func(a, b calc) calc {
		return a.(int) * b.(int)
	}).(int)
}

func generate(v int) Calc {
	if v > 1 || v < -1 {
		if v%5 == 0 {
			return _Mul(_Lit(5), generate(v/5))
		}

		if v%2 == 0 {
			return _Mul(_Lit(2), generate(v/2))
		}
	}

	return _Lit(int(v))
}
