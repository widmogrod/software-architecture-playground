package gm

import (
	"errors"
)

func NewGraphSchema() *GraphSchema {
	return &GraphSchema{
		VertexHasEdges: map[string][]UniqueNameOfVertexEntity{},
	}
}

type Type uint8

const (
	StringType Type = iota
	IntType
	BoolType
)

func TypeToString(t Type) string {
	switch t {
	case StringType:
		return "string"
	case IntType:
		return "int64"
	case BoolType:
		return "bool"
	}
	panic("unknown type")
}

type AttrType struct {
	//Guard      *Predicate
	T          Type
	Required   bool
	Identifier bool
}

type UniqueNameOfVertexEntity = string
type AttributesSchema = map[string]AttrType

type VertexEntity struct {
	Name       UniqueNameOfVertexEntity
	Attributes AttributesSchema
}

type EdgeRelationship struct {
	Edge       UniqueNameOfVertexEntity
	FromVertex UniqueNameOfVertexEntity
	ToVertex   UniqueNameOfVertexEntity
}

type EdgeEntity struct {
	Name       UniqueNameOfVertexEntity
	Attributes AttributesSchema
}

type EntityGroup struct {
	Name       UniqueNameOfVertexEntity
	Attributes AttributesSchema
}

type GraphSchema struct {
	Vertices []VertexEntity
	Edges    []EdgeEntity
	//VerticesEntityGroups []EntityGroup
	EdgeRelationships []EdgeRelationship
	VertexHasEdges    map[string][]UniqueNameOfVertexEntity
}

func (gs *GraphSchema) AddVertexEntity(id UniqueNameOfVertexEntity, attributes AttributesSchema) error {
	// check if id is unique and return error if not
	for _, v := range gs.Vertices {
		if v.Name == id {
			return errors.New("vertex with id " + id + " already exists")
		}
	}

	// create a new vertex entity
	// add it to the list of vertex entities
	gs.Vertices = append(gs.Vertices, VertexEntity{
		Name:       id,
		Attributes: attributes,
	})

	return nil
}

func (gs *GraphSchema) AddEdgeEntity(id UniqueNameOfVertexEntity, attributes AttributesSchema) error {
	// check if id is unique and return error if not
	for _, e := range gs.Edges {
		if e.Name == id {
			return errors.New("edge with id " + id + " already exists")
		}
	}

	// create a new edge entity
	// add it to the list of edge entities
	gs.Edges = append(gs.Edges, EdgeEntity{
		Name:       id,
		Attributes: attributes,
	})

	return nil
}

type AttributeMap = map[string]interface{}

// find vertex by name
func (gs *GraphSchema) FindVertexByName(name string) (*VertexEntity, error) {
	for _, v := range gs.Vertices {
		if v.Name == name {
			return &v, nil
		}
	}

	return nil, errors.New("vertex with name " + name + " not found")
}

func (gs *GraphSchema) FindEdgeByName(name string) (*EdgeEntity, error) {
	for _, e := range gs.Edges {
		if e.Name == name {
			return &e, nil
		}
	}

	return nil, errors.New("edge with name " + name + " not found")
}

func (gs *GraphSchema) ConnectByEdge(v1 string, e string, v2 string) {
	// check if vertex has edge
	if _, ok := gs.VertexHasEdges[v1]; !ok {
		gs.VertexHasEdges[v1] = []string{}
	}

	for i := range gs.VertexHasEdges[v1] {
		if gs.VertexHasEdges[v1][i] == e {
			panic("edge " + e + " already connected to vertex " + v1)
		}
	}

	gs.VertexHasEdges[v1] = append(gs.VertexHasEdges[v1], e)

	// add an edge relationship to the list of edge relationships
	gs.EdgeRelationships = append(gs.EdgeRelationships, EdgeRelationship{
		Edge:       e,
		FromVertex: v1,
		ToVertex:   v2,
	})
}
