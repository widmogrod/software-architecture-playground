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




func BFS_Maybe(f1 func(*Just),l *Maybe) {
	visited := map[*Maybe]bool{}
	queue := []*Maybe{l}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		if visited[i] {
			continue
		}
		visited[i] = true

		
		if i.Nothing0 != nil {
			continue
		}
		
		if i.Just1 != nil {
			f1(i.Just1)
			continue
		}
		
		panic("non-exhaustive")
	}
}

func BFS_List(f0 func(*Cons),l *List) {
	visited := map[*List]bool{}
	queue := []*List{l}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		if visited[i] {
			continue
		}
		visited[i] = true

		
		if i.Cons0 != nil {
			f0(i.Cons0)
			queue = append(queue, i.Cons0.List1)
			continue
		}
		
		if i.Nil1 != nil {
			continue
		}
		
		panic("non-exhaustive")
	}
}

func BFS_Tree(f1 func(*Leaf),l *Tree) {
	visited := map[*Tree]bool{}
	queue := []*Tree{l}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		if visited[i] {
			continue
		}
		visited[i] = true

		
		if i.Branch0 != nil {
			queue = append(queue, i.Branch0.Tree0)
			queue = append(queue, i.Branch0.Tree1)
			continue
		}
		
		if i.Leaf1 != nil {
			f1(i.Leaf1)
			continue
		}
		
		panic("non-exhaustive")
	}
}



func DFS_Maybe(f1 func(*Just),i *Maybe) {
	if i.Nothing0 != nil {
		return	
	}
	
	if i.Just1 != nil {
		f1(i.Just1)
		return	
	}
	
	panic("non-exhaustive")
}

func DFS_List(f0 func(*Cons),i *List) {
	if i.Cons0 != nil {
		f0(i.Cons0)
		DFS_List(f0,i.Cons0.List1)
		return	
	}
	
	if i.Nil1 != nil {
		return	
	}
	
	panic("non-exhaustive")
}

func DFS_Tree(f1 func(*Leaf),i *Tree) {
	if i.Branch0 != nil {
		DFS_Tree(f1,i.Branch0.Tree0)
		DFS_Tree(f1,i.Branch0.Tree1)
		return	
	}
	
	if i.Leaf1 != nil {
		f1(i.Leaf1)
		return	
	}
	
	panic("non-exhaustive")
}

