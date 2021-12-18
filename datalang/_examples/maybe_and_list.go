package _examples

type (
	Maybe struct { 
		Nothing *Nothing
		Just *Just
	}
	Nothing struct { 
	}
	Just struct { 
		A interface{}
	}
)

type (
	List struct { 
		Cons *Cons
		Nil *Nil
	}
	Cons struct { 
		A interface{}
		List List
	}
	Nil struct { 
	}
)

