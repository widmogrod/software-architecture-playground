package sat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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

	err := tree.Backtrack()
	assert.NoError(t, err)
	err = tree.Backtrack()
	assert.NoError(t, err)
	tree.CreateDecisionBranch(c)
	tree.ActivateBranch(c.Not())
	err = tree.Backtrack()
	assert.NoError(t, err)
	tree.CreateDecisionBranch(d)
	tree.ActivateBranch(d.Not())
	err = tree.Backtrack()
	assert.NoError(t, err)
	tree.Print()

	assert.Equal(t, []Preposition{d, c, a.Not()}, tree.Breadcrumbs())
	assert.True(t, tree.ActiveBranch().prep == d)
}
