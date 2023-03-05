package schemaless

import "github.com/widmogrod/mkunion/x/schema"

//go:generate mkunion -name=Node
type (
	Map struct {
		Ctx   *DefaultContext
		OnMap Handler
		Input Node
	}
	Merge struct {
		Ctx     *DefaultContext
		OnMerge Handler
		Input   Node
	}
	Load struct {
		Ctx    *DefaultContext
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

type Builder interface {
	WithName(string) Builder
	Load(f Handler, opts ...ContextOptionFunc) Builder
	Map(f Handler, opts ...ContextOptionFunc) Builder
	Merge(f Handler, opts ...ContextOptionFunc) Builder
	Build() []Node
}

type ContextOptionFunc func(c *DefaultContext)

func WithRetraction() ContextOptionFunc {
	return func(c *DefaultContext) {
		c.retracting = true
	}
}

func IgnoreRetractions() ContextOptionFunc {
	return func(c *DefaultContext) {
		c.retracting = false
	}
}

type DefaultContext struct {
	name       string
	retracting bool
}

func (c *DefaultContext) NoRetracting() *DefaultContext {
	c.retracting = false
	return c
}

func (c *DefaultContext) WithRetracting() *DefaultContext {
	c.retracting = true
	return c
}

func (c *DefaultContext) ShouldRetract() bool {
	return c.retracting
}

func (c *DefaultContext) Scope(name string) *DefaultContext {
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
