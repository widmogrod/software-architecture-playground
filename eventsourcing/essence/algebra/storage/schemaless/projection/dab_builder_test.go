package projection

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDabBuilderTest(t *testing.T) {

	dag := NewDAGBuilder()
	found, err := dag.GetByName("a")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, found)

	//found, err = dag.GetByName("root")
	//assert.NoError(t, err)
	//assert.Equal(t, dag, found)

	log := &LogHandler{}
	m := &LogHandler{}

	dag.
		Load(log, WithName("a")).
		Map(m, WithName("b"))

	found, err = dag.GetByName("a")
	assert.NoError(t, err)
	assert.Equal(t, log, found.dag.(*Load).OnLoad)

	found, err = dag.GetByName("b")
	assert.NoError(t, err)
	assert.Equal(t, m, found.dag.(*Map).OnMap)
}
