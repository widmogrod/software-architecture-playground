package storage

import (
	"errors"
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
)

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

func NewKeyedAggregate[T, R any](
	groupByFunc func(data T) (string, R),
	combineByFunc func(a, b R) (R, error),
	storage Repository2[schema.Schema],
) *KayedAggregate[T, R] {
	return &KayedAggregate[T, R]{
		dataByKey:    make(map[string]Record[R]),
		groupByKey:   groupByFunc,
		combineByKey: combineByFunc,
		storage:      storage,
	}
}

type Aggregator[T, R any] interface {
	Append(data T) error
	GetVersionedIndices() map[string]Record[schema.Schema]
}

var _ Aggregator[any, any] = (*KayedAggregate[any, any])(nil)

type KayedAggregate[T, R any] struct {
	groupByKey   func(data T) (string, R)
	combineByKey func(a, b R) (R, error)

	dataByKey map[string]Record[R]

	storage Repository2[schema.Schema]
}

func (t *KayedAggregate[T, R]) Append(data T) error {
	var err error

	index, result := t.groupByKey(data)
	if _, ok := t.dataByKey[index]; !ok {
		initial, err := t.loadIndex(index)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return err
			}

			t.dataByKey[index] = Record[R]{
				ID:      index,
				Data:    result,
				Version: 0,
			}
			return nil
		}

		t.dataByKey[index] = initial
	}

	result, err = t.combineByKey(t.dataByKey[index].Data, result)
	if err != nil {
		return err
	}

	existing := t.dataByKey[index]
	existing.Data = result
	t.dataByKey[index] = existing

	return nil
}

func (t *KayedAggregate[T, R]) GetVersionedIndices() map[string]Record[schema.Schema] {
	var result = make(map[string]Record[schema.Schema])
	for k, v := range t.dataByKey {
		schemed := schema.FromGo(v.Data)
		result[k] = Record[schema.Schema]{
			ID:      v.ID,
			Data:    schemed,
			Version: v.Version,
		}
	}

	return result
}

func (t *KayedAggregate[T, R]) GetIndexByKey(key string) R {
	return t.dataByKey[key].Data
}

func (t *KayedAggregate[T, R]) loadIndex(index string) (Record[R], error) {
	var r Record[R]
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

	return indexValue, nil
}
