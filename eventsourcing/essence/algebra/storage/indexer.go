package storage

import (
	"errors"
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"strings"
)

func NewNoopAggregator[T, R any]() *NoopAggregator[T, R] {
	return &NoopAggregator[T, R]{}
}

var _ Aggregator[any, any] = (*NoopAggregator[any, any])(nil)

type NoopAggregator[T, R any] struct{}

func (n *NoopAggregator[T, R]) Append(data T) error {
	return nil
}

func (n *NoopAggregator[T, R]) GetIndices() map[string]R {
	return nil
}

func NewAggregateInMemory[T, R any](
	groupByFunc func(data T) ([]string, R),
	combineByFunc func(a, b R) (R, error),
	storage Repository2[schema.Schema],
) *AggregateInMemory[T, R] {
	return &AggregateInMemory[T, R]{
		dataByKey:    make(map[string]R),
		groupByKey:   groupByFunc,
		combineByKey: combineByFunc,
		storage:      storage,
	}
}

type Aggregator[T, R any] interface {
	Append(data T) error
	GetIndices() map[string]R
}

var _ Aggregator[any, any] = (*AggregateInMemory[any, any])(nil)

type AggregateInMemory[T, R any] struct {
	groupByKey   func(data T) ([]string, R)
	combineByKey func(a, b R) (R, error)

	dataByKey map[string]R

	storage Repository2[schema.Schema]
}

func (t *AggregateInMemory[T, R]) Append(data T) error {
	var err error
	key, result := t.groupByKey(data)

	index := t.indexName(key)
	if _, ok := t.dataByKey[index]; !ok {
		initial, err := t.loadIndex(index)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return err
			}

			t.dataByKey[index] = result
			return nil
		}

		t.dataByKey[index] = initial
	}

	t.dataByKey[index], err = t.combineByKey(t.dataByKey[index], result)
	return err
}

func (t *AggregateInMemory[T, R]) GetIndices() map[string]R {
	return t.dataByKey
}

func (t *AggregateInMemory[T, R]) GetIndexByKey(key []string) R {
	index := t.indexName(key)
	return t.dataByKey[index]
}

func (t *AggregateInMemory[T, R]) indexName(key []string) string {
	return strings.Join(key, ":")
}

func (t *AggregateInMemory[T, R]) loadIndex(index string) (R, error) {
	var r R
	// load index state from storage
	// if index is found, then concat with unversionedData
	// otherwise just use unversionedData.
	initial, err := t.storage.Get(index)
	if err != nil {
		return r, fmt.Errorf("store.RepositoryInMemory2.UpdateRecords index(1)=%s %w", index, err)
	}

	indexValue, err := RecordAs[R](initial)
	if err != nil {
		return r, fmt.Errorf("store.RepositoryInMemory2.UpdateRecords index(2)=%s %w", index, err)
	}

	return indexValue.Data, nil
}
