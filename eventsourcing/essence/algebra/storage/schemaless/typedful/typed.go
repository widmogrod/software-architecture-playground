package typedful

import (
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
)

func NewTypedRepository[A any](
	store schemaless.Repository[schema.Schema],
) *TypedRepoWithAggregator[A, any] {
	return NewTypedRepoWithAggregator[A, any](
		store,
		func() schemaless.Aggregator[A, any] {
			return schemaless.NewNoopAggregator[A, any]()
		},
	)
}
