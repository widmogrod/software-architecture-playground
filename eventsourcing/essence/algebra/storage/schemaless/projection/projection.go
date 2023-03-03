package schemaless

import "github.com/widmogrod/mkunion/x/schema"

//go:generate mkunion -name=DAG
type (
	Map struct {
		OnMap Handler
		Input DAG
	}
	Merge struct {
		OnMerge Handler2
		Input   []DAG
	}
	Load struct {
		OnLoad Handler
	}
)

type Item struct {
	Key  string
	Data schema.Schema
}

//type TypeDef struct {}

type Handler interface {
	//InputType() TypeDef
	//OutputType() TypeDef
	Process(x Item, returning func(Item)) error
}

type Handler2 interface {
	//InputType() TypeDef
	//OutputType() TypeDef
	Process2(a, b Item, returning func(Item)) error
}

type Builder interface {
	Load(f Handler) Builder
	Map(f Handler) Builder
	Merge(f Handler2) Builder
	Build() DAG
}
