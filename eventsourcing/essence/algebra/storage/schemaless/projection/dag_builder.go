package projection

import (
	"container/list"
	"fmt"
)

var _ Builder = &DAGBuilder{}

func NewDAGBuilder() *DAGBuilder {
	return &DAGBuilder{
		nodes: make(map[Node]*list.List),
		dag:   nil,
		ctx: &DefaultContext{
			name: "root",
		},
	}
}

type DAGBuilder struct {
	nodes map[Node]*list.List
	dag   Node
	ctx   *DefaultContext
}

func (d *DAGBuilder) nextNumber() int {
	return len(d.nodes)
}

func (d *DAGBuilder) addNode(node Node) {

	// check if node name is already in use, yes - fail
	for n := range d.nodes {
		if n == nil || node == nil {
			panic("node is nil")
		}

		if GetCtx(n).Name() == GetCtx(node).Name() {
			panic(fmt.Sprintf("node name %s is already in use", GetCtx(node).Name()))
		}
	}

	d.nodes[node] = list.New()
}

//func (d *DAGBuilder) addDependency(from, to Node) {
//	if _, ok := d.nodes[from]; !ok {
//		d.addNode(from)
//	}
//	if _, ok := d.nodes[to]; !ok {
//		d.addNode(to)
//	}
//	d.nodes[from].PushBack(to)
//}

func (d *DAGBuilder) Load(f Handler, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope(fmt.Sprintf("Load%d", d.nextNumber()))
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Load{
		Ctx:    ctx,
		OnLoad: f,
	}

	d.addNode(node)

	return &DAGBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}

func (d *DAGBuilder) Map(f Handler, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope(fmt.Sprintf("Map%d", d.nextNumber()))
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Map{
		Ctx:   ctx,
		OnMap: f,
		Input: d.dag,
	}

	d.addNode(node)
	//d.addDependency(node, d.dag)

	return &DAGBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}

func (d *DAGBuilder) Merge(f Handler, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope(fmt.Sprintf("Merge%d", d.nextNumber()))
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Merge{
		Ctx:     ctx,
		OnMerge: f,
		Input:   d.dag,
	}

	d.addNode(node)
	//d.addDependency(node, d.dag)

	return &DAGBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}

func (d *DAGBuilder) Join(a, b Builder, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope(fmt.Sprintf("Join%d", d.nextNumber()))
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Join{
		Ctx: ctx,
		Input: []Node{
			a.(*DAGBuilder).dag,
			b.(*DAGBuilder).dag,
		},
	}

	d.addNode(node)
	//d.addDependency(node, d.dag)

	return &DAGBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}
func (d *DAGBuilder) Build() []Node {
	result := make([]Node, 0, len(d.nodes))
	for node := range d.nodes {
		result = append(result, node)
	}
	return result
}

func (d *DAGBuilder) GetByName(name string) (*DAGBuilder, error) {
	//TODO fix me!

	for node := range d.nodes {

		if node == nil {
			//continue
			panic("node is nil")
		}

		if GetCtx(node).Name() == name {
			return &DAGBuilder{
				nodes: d.nodes,
				dag:   node,
				ctx:   GetCtx(node),
			}, nil
		}
	}
	return nil, ErrNotFound
}
