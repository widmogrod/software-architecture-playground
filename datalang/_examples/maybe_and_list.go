package _examples

type (
	Maybe struct { 
		Nothing0 *Nothing
		Just1 *Just
	}
	Nothing struct { 
	}
	Just struct { 
		A0 interface{}
	}
)

type (
	List struct { 
		Cons0 *Cons
		Nil1 *Nil
	}
	Cons struct { 
		A0 interface{}
		List1 *List
	}
	Nil struct { 
	}
)

type (
	Tree struct { 
		Branch0 *Branch
		Leaf1 *Leaf
	}
	Branch struct { 
		Tree0 *Tree
		Tree1 *Tree
	}
	Leaf struct { 
		A0 interface{}
	}
)


// package _examples



func MkNothing() *Maybe {
	return &Maybe {
		Nothing0: &Nothing { 
		},
	}
}

func MkJust(a0 interface{}) *Maybe {
	return &Maybe {
		Just1: &Just { 
			A0: a0,
		},
	}
}



func MkCons(a0 interface{}, list1 *List) *List {
	return &List {
		Cons0: &Cons { 
			A0: a0,
			List1: list1,
		},
	}
}

func MkNil() *List {
	return &List {
		Nil1: &Nil { 
		},
	}
}



func MkBranch(tree0 *Tree, tree1 *Tree) *Tree {
	return &Tree {
		Branch0: &Branch { 
			Tree0: tree0,
			Tree1: tree1,
		},
	}
}

func MkLeaf(a0 interface{}) *Tree {
	return &Tree {
		Leaf1: &Leaf { 
			A0: a0,
		},
	}
}


