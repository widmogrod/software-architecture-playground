package schemaless

var _ Builder = &DagBuilder{}

func NewBuilder() *DagBuilder {
	return &DagBuilder{}
}

type DagBuilder struct {
	dag DAG
}

func (b *DagBuilder) Map(handler Handler) Builder {
	return &DagBuilder{
		dag: &Map{
			OnMap: handler,
			Input: b.dag,
		},
	}
}

func (b *DagBuilder) Merge(handler Handler) Builder {
	return &DagBuilder{
		dag: &Merge{
			OnMerge: handler,
			Input:   []DAG{b.dag},
		},
	}
}

func (b *DagBuilder) Build() DAG {
	return b.dag
}

func (b *DagBuilder) Load(data Handler) Builder {
	return &DagBuilder{
		dag: &Load{
			OnLoad: data,
		},
	}
}
