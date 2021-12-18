package _examples

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTraversingList(t *testing.T) {
	list := MkCons("a", MkCons("b", MkNil()))
	//list := &List{
	//	Cons0: &Cons{A0: "a",
	//		List1: &List{Cons0: &Cons{
	//			A0: "b", List1: &List{Nil1: &Nil{}}}}}}

	BFS(func(cons *Cons) {
		fmt.Printf("BFS:Cons=%v\n", cons.A0)
	}, func(n *Nil) {
		fmt.Printf("BFS:Nil\n")
	}, list)

	DFS(func(cons *Cons) {
		fmt.Printf("DFS:Cons=%v\n", cons.A0)
	}, func(n *Nil) {
		fmt.Printf("DFS:Nil\n")
	}, list)

	res, err := json.Marshal(list)
	assert.NoError(t, err)
	fmt.Println(string(res))
}

func TestTraversingTree(t *testing.T) {
	tree := MkBranch(
		MkBranch(
			MkLeaf("a"),
			MkBranch(
					MkLeaf("a.1"),
					MkLeaf("a.2")),
		),
		MkBranch(
			MkLeaf("b"),
			MkBranch(
				MkLeaf("b.1"),
				MkLeaf("b.2")),
		),
	)
	//tree := &Tree{
	//	Branch0: &Branch{
	//		Tree0: &Tree{
	//			Branch0: &Branch{
	//				Tree0: &Tree{Leaf1: &Leaf{A0: "a"}},
	//				Tree1: &Tree{
	//					Branch0: &Branch{
	//						Tree0: &Tree{Leaf1: &Leaf{A0: "a.1"}},
	//						Tree1: &Tree{Leaf1: &Leaf{A0: "a.2"}},
	//					},
	//					Leaf1: nil,
	//				},
	//			},
	//			Leaf1: nil,
	//		},
	//		Tree1: &Tree{
	//			Branch0: &Branch{
	//				Tree0: &Tree{Leaf1: &Leaf{A0: "b"}},
	//				Tree1: &Tree{
	//					Branch0: &Branch{
	//						Tree0: &Tree{Leaf1: &Leaf{A0: "b.1"}},
	//						Tree1: &Tree{Leaf1: &Leaf{A0: "b.2"}},
	//					},
	//					Leaf1: nil,
	//				},
	//			},
	//			Leaf1: nil,
	//		},
	//	},
	//	Leaf1: nil,
	//}

	BFS_Tree(func(n *Leaf) {
		fmt.Printf("BFS_Tree:Leaf=%v\n", n.A0)
	}, tree)

	DFS_Tree(func(n *Leaf) {
		fmt.Printf("DFS_Tree:Leaf=%v\n", n.A0)
	}, tree)

	res, err := json.Marshal(tree)
	assert.NoError(t, err)
	fmt.Println(string(res))
}
