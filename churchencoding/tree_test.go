package churchencoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTree(t *testing.T) {
	var exampleTree Tree = _Node(
		1,
		_Node(2, _Leaf(), _Leaf()),
		_Node(3, _Leaf(), _Leaf()),
	)

	assert.Equal(t, []int{1, 2, 3}, preorder(exampleTree))
}
