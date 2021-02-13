package churchencoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTree(t *testing.T) {
	/*

			       3					_Node(3,
			    /     \
		       2	   1 			_Node(2, 							_Node(1,
			    \    /  \		_Leaf(), _Node(6,			_Node(4,
				 6  4    5			_Leaf(), _Leaf()),	_Leaf(), _Leaf()),   ..
						/
			           7

			BFS(t, print) == 3,2,1,6,4,5,7
	*/
	var exampleTree Tree = _Node(
		3,
		_Node(2,
			_Leaf(),
			_Node(6, _Leaf(), _Leaf())),
		_Node(1,
			_Node(4, _Leaf(), _Leaf()),
			_Node(5,
				_Node(7, _Leaf(), _Leaf()), _Leaf())),
	)

	assert.Equal(t, []int{3, 2, 6, 1, 4, 5, 7}, preorder(exampleTree))
}
