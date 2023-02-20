package storage

import "strings"

func NewNoopIndexer[T, R any]() *NoopIndexer[T, R] {
	return &NoopIndexer[T, R]{}
}

var _ Indexerr[any, any] = (*NoopIndexer[any, any])(nil)

type NoopIndexer[T, R any] struct{}

func (n *NoopIndexer[T, R]) Append(data T) {}

func (n *NoopIndexer[T, R]) GetIndices() map[string]R {
	return nil
}

func NewIndexer[T, R any](
	groupByFunc func(data T) ([]string, R),
	combineByFunc func(a, b R) (R, error),
) *Indexer[T, R] {
	return &Indexer[T, R]{
		dataByKey:    make(map[string]R),
		groupByKey:   groupByFunc,
		combineByKey: combineByFunc,
	}
}

type Indexerr[T, R any] interface {
	Append(data T)
	GetIndices() map[string]R
}

var _ Indexerr[any, any] = (*Indexer[any, any])(nil)

type Indexer[T, R any] struct {
	groupByKey   func(data T) ([]string, R)
	combineByKey func(a, b R) (R, error)

	dataByKey map[string]R
}

func (t *Indexer[T, R]) Append(data T) {
	key, result := t.groupByKey(data)

	index := t.indexName(key)
	// TODO load from storage index, instead assuming it's empty
	// That way, if index is created by other process, it can be updated with new data.
	// Interestingly, when asychronouse process will update the same index, to not overwrite, it should
	// use the same combineByKey function, and it should be idempotent.
	// and it should have versioning to do this. One ove ways how to approach it is to inject latest version
	// and durign save when it's rejected, it should be retried with latest version.
	// Thanks to Combine operation, injecting latest state of index, can be done outside of indexer, in repository layer
	// This may may non intuitive from this class perspective, but it may be intuitive from repository perspective.
	if _, ok := t.dataByKey[index]; !ok {
		t.dataByKey[index] = result
	} else {
		t.dataByKey[index], _ = t.combineByKey(t.dataByKey[index], result)
	}
}

func (t *Indexer[T, R]) GetIndices() map[string]R {
	return t.dataByKey
}

func (t *Indexer[T, R]) GetIndexByKey(key []string) R {
	index := t.indexName(key)
	return t.dataByKey[index]
}

func (t *Indexer[T, R]) indexName(key []string) string {
	return strings.Join(key, ":")
}
