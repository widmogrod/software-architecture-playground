package churchencoding

type tree = interface{}
type (
	Node = func(int, tree, tree) tree
	Leaf = func() tree
	Tree = func(Node, Leaf) tree
)

func _Node(v int, left Tree, right Tree) Tree {
	return func(node Node, leaf Leaf) tree {
		return node(v, left(node, leaf), right(node, leaf))
	}
}

func _Leaf() Tree {
	return func(_ Node, leaf Leaf) tree {
		return leaf()
	}
}

func preorder(t Tree) []int {
	return t(func(i int, left, right tree) tree {
		result := []int{i}
		result = append(result, left.([]int)...)
		result = append(result, right.([]int)...)
		return result
	}, func() tree {
		return []int{}
	}).([]int)
}
