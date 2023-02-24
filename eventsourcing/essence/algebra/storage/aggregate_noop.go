package storage

import "github.com/widmogrod/mkunion/x/schema"

func NewNoopAggregator[T, R any]() *NoopAggregator[T, R] {
	return &NoopAggregator[T, R]{}
}

var _ Aggregator[any, any] = (*NoopAggregator[any, any])(nil)

type NoopAggregator[T, R any] struct{}

func (n *NoopAggregator[T, R]) Append(data T) error {
	return nil
}

func (n *NoopAggregator[T, R]) GetVersionedIndices() map[string]Record[schema.Schema] {
	return nil
}
