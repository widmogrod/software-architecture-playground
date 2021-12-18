package _examples

// tr :: (a -> b) ->
func DFS(c func(*Cons), n func(*Nil), l *List) {
	if l.Cons0 != nil {
		c(l.Cons0)
		DFS(c, n, l.Cons0.List1)
	} else if l.Nil1 != nil {
		n(l.Nil1)
	} else {
		panic("non-exhaustive")
	}
}

func BFS(c func(*Cons), n func(*Nil), l *List) {
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
			c(i.Cons0)
			queue = append(queue, i.Cons0.List1)
		} else if i.Nil1 != nil {
			n(i.Nil1)
		} else {
			panic("non-exhaustive")
		}
	}
}

//type paramorphism = func(a, b *ActivityResult, accumulator interface{}) interface{}
//
//// Para is a paramorphism that will reduce  Flow AST to new algebra,
//// and during reduction provide context as well with accumulator
//// ```haskell
//// para :: (a -> ([a], b) -> b) -> b -> [a] ->  b
//// ```
//func Para(fn paramorphism, accumulator interface{}, start *ActivityResult) interface{} {



// tr :: (a -> b) ->
func DFS_Tree(n func(*Leaf), t *Tree) {
	if t.Branch0 != nil {
		DFS_Tree(n, t.Branch0.Tree0)
		DFS_Tree(n, t.Branch0.Tree1)
	} else if t.Leaf1 != nil {
		n(t.Leaf1)
	} else {
		panic("non-exhaustive")
	}
}

func BFS_Tree(n func(*Leaf), l *Tree) {
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
		} else if i.Leaf1 != nil {
			n(i.Leaf1)
		} else {
			panic("non-exhaustive")
		}
	}
}