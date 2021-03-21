package sat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoolVar(t *testing.T) {
	a := MkBool()

	assert.True(t, a.IsTrue())
	assert.True(t, Not(Not(a)).IsTrue())
	assert.False(t, Not(a).IsTrue())

	// Same variable that translate to the same result can be equal
	assert.True(t, a.Equal(a))
	assert.True(t, Not(a).Equal(Not(a)))
	assert.True(t, Not(Not(a)).Equal(a))
	assert.False(t, a.Equal(Not(a)))
	assert.False(t, Not(a).Equal(a))

	// Two different variables cannot be equal
	b := MkBool()
	assert.False(t, a.Equal(b))
	assert.False(t, Not(a).Equal(b))
	assert.False(t, Not(a).Equal(Not(b)))

	// Same
	assert.True(t, a.SameVar(a))
	assert.True(t, a.SameVar(Not(a)))
	assert.True(t, Not(a).SameVar(Not(a)))
	assert.True(t, Not(a).SameVar(a))

	assert.False(t, b.SameVar(a))
	assert.False(t, b.SameVar(Not(a)))
	assert.False(t, Not(b).SameVar(Not(a)))
	assert.False(t, Not(b).SameVar(a))
}

func TestSat1(t *testing.T) {
	a := MkBool()
	b := MkBool()

	sat := NewSolver()
	sat.And(a, Not(b))

	sat.PrintCNF()

	result := sat.Solution()
	assert.Equal(t, result, []Preposition{a})
}

//	a -b  c
//	a  b -c
//	   b -c
// 			 d
func TestSat2(t *testing.T) {
	a := MkBool()
	b := MkBool()
	c := MkBool()
	d := MkBool()

	sat := NewSolver()
	sat.And(a, Not(b), c)
	sat.And(a, b, Not(c))
	sat.And(b, Not(c))
	sat.And(d)

	sat.PrintCNF()

	result := sat.Solution()
	assert.Equal(t, result, []Preposition{d, b, a})
}

func TestSat3(t *testing.T) {
	a := MkBool()
	b := MkBool()
	c := MkBool()
	d := MkBool()

	sat := NewSolver()
	sat.AddClosures(ExactlyOne([]*BoolVar{a, b, c, d}))

	sat.PrintCNF()

	result := sat.Solution()
	assert.Equal(t, result, []Preposition{d.Not(), c.Not(), b.Not(), a})
}

func TestDecisionTree(t *testing.T) {
	a := MkBool()
	b := MkBool()
	c := MkBool()
	d := MkBool()

	tree := NewDecisionTree()
	n := tree.ActiveBranch()
	assert.True(t, tree.IsRoot(n))

	assert.False(t, tree.HasFromBranchToRoot(n, a))
	tree.CreateDecisionBranch(a)
	tree.ActivateBranch(a)
	tree.CreateDecisionBranch(b)
	tree.ActivateBranch(b.Not())

	n = tree.ActiveBranch()
	//tree.Print()
	assert.True(t, n.IsLeaf())
	assert.True(t, tree.HasFromBranchToRoot(n, a))
	assert.True(t, tree.HasFromBranchToRoot(n, b.Not()))

	tree.Backtrack()
	tree.Backtrack()
	tree.CreateDecisionBranch(c)
	tree.ActivateBranch(c.Not())
	tree.Backtrack()
	tree.CreateDecisionBranch(d)
	tree.ActivateBranch(d.Not())
	tree.Backtrack()
	tree.Print()

	assert.Equal(t, []Preposition{d, c, a.Not()}, tree.Breadcrumbs())
	assert.True(t, tree.ActiveBranch().Value() == d.String())
}
