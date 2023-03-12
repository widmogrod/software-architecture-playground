package projection

import "github.com/widmogrod/mkunion/x/schema"

//go:generate mkunion -name=Node
type (
	Map struct {
		Ctx   *DefaultContext
		OnMap Handler
		Input Node
	}
	// Merge implicitly means, merge by key
	Merge struct {
		Ctx     *DefaultContext
		OnMerge Handler
		Input   Node
	}
	Load struct {
		Ctx    *DefaultContext
		OnLoad Handler
	}
	Join struct {
		Ctx   *DefaultContext
		Input []Node
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
	Join(a, b Builder, opts ...ContextOptionFunc) Builder
	Build() []Node
}

type ContextOptionFunc func(c *DefaultContext)

func WithRetraction() ContextOptionFunc {
	return func(c *DefaultContext) {
		yes := true
		c.retracting = &yes
	}
}

func IgnoreRetractions() ContextOptionFunc {
	return func(c *DefaultContext) {
		no := false
		c.retracting = &no
	}
}

type DefaultContext struct {
	name       string
	retracting *bool
}

func (c *DefaultContext) ShouldRetract() bool {
	if c.retracting == nil {
		return false
	}

	return *c.retracting
}

func (c *DefaultContext) Scope(name string) *DefaultContext {
	return &DefaultContext{
		name: c.name + "." + name,
		// Should we copy this?
		//retracting: c.retracting,
	}
}

func (c *DefaultContext) Name() string {
	return c.name
}

type Message struct {
	Offset int
	// at some point of time i may need to pass type reference
	Key       string
	Aggregate *Item
	Retract   *Item
}
