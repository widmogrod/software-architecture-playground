package projection

import (
	"fmt"
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

	/*
		mermaid
		graph TD
			a[Load]
			b[Map]
			c[Load]
			d[Map]
			e[Join]
			f[Map]
			a --> b
			c --> d
			b --> e
			d --> e
			e --> f
	*/
	mapped1 := dag.
		Load(log, WithName("a")).
		Map(m, WithName("b"))

	mapped2 := dag.
		Load(log, WithName("c")).
		Map(m, WithName("d"))

	dag.
		Join(mapped1, mapped2, WithName("e")).
		Map(m, WithName("f"))

	found, err = dag.GetByName("a")
	assert.NoError(t, err)
	assert.Equal(t, log, found.dag.(*Load).OnLoad)

	found, err = dag.GetByName("b")
	assert.NoError(t, err)
	assert.Equal(t, m, found.dag.(*Map).OnMap)

	nodes := dag.Build()
	assert.Equal(t, 6, len(nodes))

	//assert.Equal(t, "a", GetCtx(nodesFromTo[0]).Name())
	//assert.Equal(t, "b", GetCtx(nodesFromTo[1]).Name())

	fmt.Println(ToMermaidGraph(dag))

	fmt.Println(ToMermaidGraphWithOrder(dag, ReverseSort(Sort(dag))))
}
