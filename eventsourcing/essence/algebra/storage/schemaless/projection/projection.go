package schemaless

import "github.com/widmogrod/mkunion/x/schema"

//go:generate mkunion -name=DAG
type (
	Map struct {
		Name  Context
		OnMap Handler
		Input DAG
	}
	Merge struct {
		Name    Context
		OnMerge Handler
		Input   DAG
	}
	Load struct {
		Name   Context
		OnLoad Handler
	}
)

type Item struct {
	Key  string
	Data schema.Schema
}

//type TypeDef struct {}

//type Context interface {
//	KV() KVStore
//}

type Handler interface {
	//InputType() TypeDef
	//OutputType() TypeDef
	Process(x Item, returning func(Item)) error
	Retract(x Item, returning func(Item)) error
}

type Context interface {
	Scope(name string) Context
	Name() string

	WithRetracting() Context
	ShouldRetract() bool
	NoRetracting() Context
}

type Builder interface {
	Load(ctx Context, f Handler) Builder
	Map(ctx Context, f Handler) Builder
	Merge(ctx Context, f Handler) Builder
	Build() DAG
	Build2() []DAG
}

type DefaultContext struct {
	name       string
	retracting bool
}

func (c *DefaultContext) NoRetracting() Context {
	c.retracting = false
	return c
}

func (c *DefaultContext) WithRetracting() Context {
	c.retracting = true
	return c
}

func (c *DefaultContext) ShouldRetract() bool {
	return c.retracting
}

func (c *DefaultContext) Scope(name string) Context {
	return &DefaultContext{
		name:       c.name + "." + name,
		retracting: c.retracting,
	}
}

func (c *DefaultContext) Name() string {
	return c.name
}

type Message struct {
	Offset    int
	Key       string
	Aggregate *Item
	Retract   *Item
}
