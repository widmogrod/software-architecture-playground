package schemaless

import "container/list"

var _ Builder = &DagBuilder{}

func NewBuilder() *DagBuilder {
	return &DagBuilder{
		nodes: make(map[Node]*list.List),
		dag:   nil,
		ctx: &DefaultContext{
			name: "root",
		},
	}
}

type DagBuilder struct {
	nodes map[Node]*list.List
	dag   Node
	ctx   *DefaultContext
}

func (d *DagBuilder) WithName(s string) Builder {
	return &DagBuilder{
		nodes: d.nodes,
		dag:   d.dag,
		ctx:   d.ctx.Scope(s),
	}
}

func (d *DagBuilder) addNode(node Node) {
	d.nodes[node] = list.New()
}

func (d *DagBuilder) addDependency(from, to Node) {
	if _, ok := d.nodes[from]; !ok {
		d.addNode(from)
	}
	if _, ok := d.nodes[to]; !ok {
		d.addNode(to)
	}
	d.nodes[from].PushBack(to)
}

func (d *DagBuilder) Load(f Handler, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope("Load")
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Load{
		Ctx:    ctx,
		OnLoad: f,
	}

	d.addNode(node)

	return &DagBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}

func (d *DagBuilder) Map(f Handler, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope("Map")
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Map{
		Ctx:   ctx,
		OnMap: f,
		Input: d.dag,
	}

	d.addDependency(node, d.dag)

	return &DagBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}

func (d *DagBuilder) Merge(f Handler, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope("Merge")
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Merge{
		Ctx:     ctx,
		OnMerge: f,
		Input:   d.dag,
	}

	d.addDependency(node, d.dag)

	return &DagBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}

func (d *DagBuilder) Join(a, b Builder, opts ...ContextOptionFunc) Builder {
	ctx := d.ctx.Scope("Join")
	for _, opt := range opts {
		opt(ctx)
	}

	node := &Join{
		Ctx: ctx,
		Input: []Node{
			a.(*DagBuilder).dag,
			b.(*DagBuilder).dag,
		},
	}

	d.addDependency(node, d.dag)

	return &DagBuilder{
		nodes: d.nodes,
		dag:   node,
		ctx:   ctx,
	}
}
func (d *DagBuilder) Build() []Node {
	result := make([]Node, 0, len(d.nodes))
	for node := range d.nodes {
		result = append(result, node)
	}
	return result
}
