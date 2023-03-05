package schemaless

import "container/list"

var _ Builder = &DagBuilder{}

func NewBuilder() *DagBuilder {
	return &DagBuilder{
		nodes: make(map[DAG]*list.List),
	}
}

type DagBuilder struct {
	nodes map[DAG]*list.List
	dag   DAG
}

func (d *DagBuilder) addNode(node DAG) {
	d.nodes[node] = list.New()
}

func (d *DagBuilder) addDependency(from, to DAG) {
	if _, ok := d.nodes[from]; !ok {
		d.addNode(from)
	}
	if _, ok := d.nodes[to]; !ok {
		d.addNode(to)
	}
	d.nodes[from].PushBack(to)
}

func (b *DagBuilder) Map(ctx Context, handler Handler) Builder {
	node := &Map{
		Name:  ctx,
		OnMap: handler,
		Input: b.dag,
	}

	b.addDependency(node, b.dag)

	return &DagBuilder{
		nodes: b.nodes,
		dag:   node,
	}
}

func (b *DagBuilder) Merge(ctx Context, handler Handler) Builder {
	node := &Merge{
		Name:    ctx,
		OnMerge: handler,
		Input:   b.dag,
	}

	b.addDependency(node, b.dag)

	return &DagBuilder{
		nodes: b.nodes,
		dag:   node,
	}
}

func (b *DagBuilder) Load(ctx Context, data Handler) Builder {
	node := &Load{
		Name:   ctx,
		OnLoad: data,
	}

	b.addNode(node)

	return &DagBuilder{
		nodes: b.nodes,
		dag:   node,
	}
}

func (b *DagBuilder) Build() DAG {
	return b.dag
}

func (d *DagBuilder) Build2() []DAG {
	result := make([]DAG, 0, len(d.nodes))
	//visited := make(map[DAG]bool)
	//var visit func(node DAG)
	//visit = func(node DAG) {
	//	visited[node] = true
	//	for e := d.nodes[node].Front(); e != nil; e = e.Next() {
	//		child := e.Value.(DAG)
	//		if !visited[child] {
	//			visit(child)
	//		}
	//	}
	//	result = append(result, node)
	//}
	//for node := range d.nodes {
	//	if !visited[node] {
	//		visit(node)
	//	}
	//}
	//for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
	//	result[i], result[j] = result[j], result[i]
	//}
	for node := range d.nodes {
		result = append(result, node)
	}
	return result
}
