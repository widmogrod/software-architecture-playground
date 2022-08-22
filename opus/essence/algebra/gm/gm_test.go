package gm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Sourcable(in map[string]AttrType) map[string]AttrType {
	// add id
	//in["id"] = AttrType{T: StringType, Required: true, Identifier: true}

	// add sourceType and sourceId to the input map
	//in["sourceType"] = AttrType{T: StringType, Required: true}
	//in["sourceId"] = AttrType{T: StringType, Required: true}

	// add tenantId to the input map
	//in["tenantId"] = AttrType{T: StringType, Required: true}

	// add schema version to the input map
	//in["schema"] = AttrType{T: IntType, Required: true}

	// add version to the input map
	//in["version"] = AttrType{T: IntType, Required: true}
	return in
}

func TestGraphSchema(t *testing.T) {
	schema := NewGraphSchema()
	err := schema.AddVertexEntity("question", Sourcable(map[string]AttrType{
		"content": {T: StringType, Required: true},
	}))
	assert.NoError(t, err)

	err = schema.AddVertexEntity("answer", Sourcable(map[string]AttrType{
		"content": {T: StringType, Required: true},
	}))
	assert.NoError(t, err)

	err = schema.AddVertexEntity("comment", Sourcable(map[string]AttrType{
		"content": {T: StringType, Required: true},
	}))
	assert.NoError(t, err)

	err = schema.AddVertexEntity("user", Sourcable(map[string]AttrType{
		"name":     {T: StringType, Required: true},
		"age":      {T: IntType, Required: true},
		"is admin": {T: BoolType, Required: false},
	}))
	assert.NoError(t, err)

	err = schema.AddEdgeEntity("authoredBy", Sourcable(map[string]AttrType{}))
	assert.NoError(t, err)

	err = schema.AddEdgeEntity("isCommentedBy", Sourcable(map[string]AttrType{}))
	assert.NoError(t, err)

	schema.ConnectByEdge("question", "authoredBy", "user")
	schema.ConnectByEdge("answer", "authoredBy", "user")
	schema.ConnectByEdge("comment", "authoredBy", "user")

	schema.ConnectByEdge("question", "isCommentedBy", "comment")
	schema.ConnectByEdge("answer", "isCommentedBy", "comment")

	var queue []UniqueNameOfVertexEntity
	for _, vertex := range schema.Vertices {
		queue = append(queue, vertex.Name)
	}

	iterationCount := 0
	maxIterations := 10

	for {
		if len(queue) == 0 {
			break
		}
		if iterationCount >= maxIterations {
			break
		}

		iterationCount++
		id := queue[0]
		queue = queue[1:]

		v, err := schema.FindVertexByName(id)
		assert.NoError(t, err)

		// for each edge relationship
		for _, r := range schema.EdgeRelationships {
			if r.FromVertex != v.Name {
				continue
			}

			// find the edge entity
			e, err := schema.FindEdgeByName(r.Edge)
			assert.NoError(t, err)
			// find the vertex entity
			to, err := schema.FindVertexByName(r.ToVertex)
			assert.NoError(t, err)

			queue = append(queue, to.Name)

			fmt.Println(v.Name, e.Name, to.Name)
		}
	}
}
