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

//go:generate mkunion -name=Message
type (
	Combine struct {
		Key  string
		Data schema.Schema
	}
	Retract struct {
		Key  string
		Data schema.Schema
	}
	Both struct {
		Key     string
		Retract Retract
		Combine Combine
	}
)

type TypeDef struct {
}

type Handler interface {
	//InputType() TypeDef
	//OutputType() TypeDef
	Process(msg Message, returning func(Message)) error
}

type Handler2 interface {
	//InputType() TypeDef
	//OutputType() TypeDef
	Process2(a, b Message, returning func(Message)) error
}

type Builder interface {
	Load(f Handler) Builder
	Map(f Handler) Builder
	Merge(f Handler2) Builder
	Build() DAG
}
