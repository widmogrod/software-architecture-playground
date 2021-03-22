package sat

func Not(c Preposition) Preposition {
	return c.Not()
}

func OneOf(vars []*BoolVar) []Preposition {
	var result = make([]Preposition, len(vars))
	for i := 0; i < len(vars); i++ {
		result[i] = vars[i]
	}

	return result
}

// X1 or X2 and (!X1 or !X2)
func ExactlyOne(vars []Preposition) Closures {
	var closures Closures
	closures = append(closures, vars)

	size := len(vars)
	for i := 0; i < size-1; i++ {
		for j := i + 1; j < size; j++ {
			closures = append(closures, []Preposition{
				Not(vars[i]),
				Not(vars[j]),
			})
		}
	}

	return closures
}

func Num(xs ...int) []Preposition {
	preps := make([]Preposition, len(xs))
	for i, v := range xs {
		preps[i] = MkPrep(v)
	}
	return preps
}
