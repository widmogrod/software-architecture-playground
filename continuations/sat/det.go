package sat

import (
	"errors"
	"fmt"
	"strings"
)

func NewDecisionTree() *DecisionTree {
	root := &Branch{}
	return &DecisionTree{
		root:   root,
		active: root,
	}
}

type Branch struct {
	prep Preposition

	parent      *Branch
	left, right *Branch
	dontVisit   bool
}

func (b *Branch) IsLeaf() bool {
	return b.left == nil && b.right == nil
}

type DecisionTree struct {
	root   *Branch
	active *Branch
}

// ActiveBranch return currently visited node
func (t *DecisionTree) ActiveBranch() *Branch {
	return t.active
}

func (t *DecisionTree) IsRoot(n *Branch) bool {
	return t.root == n
}

func (t *DecisionTree) Breadcrumbs() []Preposition {
	var result []Preposition
	n := t.active

	for !t.IsRoot(n) {
		result = append(result, n.prep)
		n = n.parent
	}

	return result
}

func (t *DecisionTree) HasFromBranchToRoot(n *Branch, prep Preposition) bool {
	for !t.IsRoot(n) {
		if n.prep.Equal(prep) {
			return true
		}

		n = n.parent
	}

	return false
}

func (t *DecisionTree) CreateDecisionBranch(prep Preposition) {
	if !t.active.IsLeaf() {
		panic(fmt.Sprintf("CreateDecisionBranch: branch '%s' has decision", t.fmtPath(t.active)))
	}

	lb := &Branch{
		prep:   prep,
		parent: t.active,
	}
	rb := &Branch{
		prep:   prep.Not(),
		parent: t.active,
	}

	t.active.left = lb
	t.active.right = rb
}
func (t *DecisionTree) ActivateBranch(prep Preposition) {
	if t.active.left != nil {
		if t.active.left.prep.Equal(prep) {
			t.active = t.active.left
			return
		}
	}
	if t.active.right != nil {
		if t.active.right.prep.Equal(prep) {
			t.active = t.active.right
			return
		}
	}

	panic(fmt.Sprintf(
		"ActivateBranch: active branch '%s' don't lead to next decision %s",
		t.fmtPath(t.active), prep.String()))
}

var ErrBacktrackExhausted = errors.New("backtrack: reach root of decision tree, cannot backtrack more")

func (t *DecisionTree) Backtrack() error {
	if t.IsRoot(t.active) {
		return ErrBacktrackExhausted
	}

	t.DontVisitAnymore(t.active)

	if !t.active.IsLeaf() && !t.active.dontVisit {
		if t.active.left.dontVisit && t.active.right.dontVisit {
			return t.Backtrack()
		}
	}

	if t.active.parent.left == t.active {
		if t.active.parent.right.dontVisit {
			t.active = t.active.parent
			return t.Backtrack()
		} else {
			t.active = t.active.parent.right
		}
	} else {
		if t.active.parent.left.dontVisit {
			t.active = t.active.parent
			return t.Backtrack()
		} else {
			t.active = t.active.parent.left
		}
	}

	return nil
}

func (t *DecisionTree) DontVisitAnymore(n *Branch) {
	n.dontVisit = true
}

func (t *DecisionTree) Print() {
	fmt.Println(t.fmtBranch(t.root, 0))
}

func (t *DecisionTree) fmtBranch(branch *Branch, level int) string {
	if branch == nil {
		return ""
	}

	indent := "\n  "
	if t.active == branch {
		indent = "\n* "
	}
	indent += strings.Repeat(" ", level)
	name := ""
	if !t.IsRoot(branch) {
		indent += "∟"
		name = branch.prep.String()
	}

	if branch.dontVisit {
		name += " ˣ"
	}

	if branch.IsLeaf() {
		name += " 🍂"
	}

	return fmt.Sprintf(
		"%s %s %s %s",
		indent,
		name,
		t.fmtBranch(branch.left, level+1),
		t.fmtBranch(branch.right, level+1),
	)
}

func (t *DecisionTree) fmtPath(n *Branch) string {
	if t.IsRoot(n) {
		return "root"
	}

	path := ""
	for !t.IsRoot(n) {
		path = n.prep.String() + " > " + path
		n = n.parent
	}

	return path
}
