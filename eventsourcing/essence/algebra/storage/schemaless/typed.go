package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
)

func NewRepository2Typed[A any](
	store Repository2[schema.Schema],
) *RepositoryWithAggregator[A, any] {
	return NewRepositoryWithAggregator[A, any](
		store,
		func() Aggregator[A, any] {
			return NewNoopAggregator[A, any]()
		},
	)
}
