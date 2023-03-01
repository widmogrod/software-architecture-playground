package schemaless

import "github.com/widmogrod/mkunion/x/schema"

//go:generate mkunion -name=DAG
type (
	Map struct {
		OnMap Handler
		Input DAG
	}
	Merge struct {
		OnMerge Handler
		Input   []DAG
	}
)

//go:generate mkunion -name=Message
type (
	Combine struct {
		// Key string
		Data schema.Schema
	}
	Retract struct {
		Key  string
		Data schema.Schema
	}
	Both struct {
		Retract Retract
		Combine Combine
	}
)

type TypeDef struct {
}

type Handler interface {
	InputType() TypeDef
	OutputType() TypeDef
	Process(msg Message) (Message, error)
}

type Builder interface {
	Map(f Handler) Builder
	Merge(f Handler) Builder
	Build() DAG
}
