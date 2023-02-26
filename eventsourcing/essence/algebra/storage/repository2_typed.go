package storage

import (
	"github.com/widmogrod/mkunion/x/schema"
)

func NewRepository2Typed[A any](
	storage Repository2[schema.Schema],
) *RepositoryWithAggregator[A, any] {
	return NewRepositoryWithAggregator[A, any](
		storage,
		func() Aggregator[A, any] {
			return NewNoopAggregator[A, any]()
		},
	)
}
