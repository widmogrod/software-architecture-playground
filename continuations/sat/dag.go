package sat

import (
	"fmt"
	"hash/crc32"
	"image/color/palette"
)

func NewDAG() *DAG {
	return &DAG{
		vlist:       make([]VertexID, 0),
		vertices:    make(map[VertexID]struct{}),
		edgesFromTo: make(map[VertexID][]*Edge),
	}
}

type VertexID = int

type Edge struct {
	label    string
	from, to VertexID
}

func (e *Edge) Equal(e2 *Edge) bool {
	if e.to != e2.to {
		return false
	}
	if e.from != e2.from {
		return false
	}
	if e.label != e2.label {
		return false
	}

	return true
}

type DAG struct {
	vlist       []VertexID
	vertices    map[VertexID]struct{}
	edgesFromTo map[VertexID][]*Edge
}

func (d *DAG) SetVertex(v VertexID) {
	if _, found := d.vertices[v]; !found {
		d.vertices[v] = struct{}{}
		d.vlist = append(d.vlist, v)
	}
}

func (d *DAG) HasVertex(id VertexID) bool {
	_, found := d.vertices[id]
	return found
}

func (d *DAG) SetEdge(e *Edge) {
	if !(d.HasVertex(e.from) && d.HasVertex(e.to)) {
		panic(fmt.Sprintf("cannot conect vertices that don't exists %d %d by edge %#v", e.from, e.to, e))
	}

	if e.from == e.to {
		panic(fmt.Sprintf("cannot conect one vertice to itself %d by edge %#v", e.to, e))
	}

	// check if there is no circle
	if d.HasEdge(e) {
		return
		//fmt.Printf("\n\n %#v \n\n", d.edgesFromTo)
		//panic(fmt.Sprintf("edge %#v already exist", e))
	}

	if d.HasEdge(&Edge{
		label: e.label,
		from:  e.to,
		to:    e.from,
	}) {
		panic(fmt.Sprintf("circular reference! connecting by edge %#v will result in cirlular dependency", e))
	}

	d.edgesFromTo[e.from] = append(d.edgesFromTo[e.from], e)
}
func (d *DAG) HasEdge(e *Edge) bool {
	if edges, found := d.edgesFromTo[e.from]; found {
		for _, edge := range edges {
			if edge.Equal(e) {
				return true
			}
		}
	}

	return false
}

func (d *DAG) Edges(a VertexID) []*Edge {
	return d.edgesFromTo[a]
}

func (d *DAG) Print() {
	for _, tos := range d.edgesFromTo {
		for _, e := range tos {
			r, g, b, _ := palette.Plan9[crc32.ChecksumIEEE([]byte(e.label))%254].RGBA()
			fmt.Printf("x%d -> x%d [label=\"%s\", color=\"#%x%x%x\"] \n", e.from, e.to, e.label, r, g, b)
		}
	}
}

func (d *DAG) Vertices() []VertexID {
	return d.vlist
}
