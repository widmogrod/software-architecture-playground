package projection

import (
	"errors"
	"github.com/widmogrod/mkunion/x/schema"
)

var ErrNotFound = errors.New("node not found")

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
	//TODO add for completness Split
	//Split struct {}
)

func GetCtx(node Node) *DefaultContext {
	return MustMatchNode(
		node,
		func(node *Map) *DefaultContext { return node.Ctx },
		func(node *Merge) *DefaultContext { return node.Ctx },
		func(node *Load) *DefaultContext { return node.Ctx },
		func(node *Join) *DefaultContext { return node.Ctx },
	)
}

func NodeToString(node Node) string {
	return MustMatchNode(
		node,
		func(node *Map) string { return "Map" },
		func(node *Merge) string { return "Merge" },
		func(node *Load) string { return "Load" },
		func(node *Join) string { return "Join" },
	)
}

type EventTime = int64

type Item struct {
	Key       string
	Data      schema.Schema
	EventTime EventTime
	Window    *Window

	//finished bool
}

type Window struct {
	Start int64
	End   int64
}

//type TypeDef struct {}

//type Context interface {
//	KV() KVStore
//}

//type Consumer interface {
//	Consume(x Item) (hasNext bool)
//}
//
//type Producer interface {
//	Return(x Item)
//	Finish()
//}

//type ProduceAdapter struct{}
//
//func (p *ProduceAdapter) Return(x Item) {}
//func (p *ProduceAdapter) Finish()       {}

type Handler interface {
	//InputType() TypeDef
	//OutputType() TypeDef
	Process(x Item, returning func(Item)) error
	Retract(x Item, returning func(Item)) error
}

type Builder interface {
	Load(f Handler, opts ...ContextOptionFunc) Builder
	Map(f Handler, opts ...ContextOptionFunc) Builder
	Merge(f Handler, opts ...ContextOptionFunc) Builder
	Join(a, b Builder, opts ...ContextOptionFunc) Builder
	Build() []Node
}

type ContextOptionFunc func(c *DefaultContext)

func WithName(name string) ContextOptionFunc {
	return func(c *DefaultContext) {
		c.name = name
	}
}
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
	name        string
	contextName string
	retracting  *bool
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

	finished bool
}

type Stats = map[string]int
