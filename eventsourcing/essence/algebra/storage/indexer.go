package storage

import "strings"

func NewIndexer[T, R any]() *Indexer[T, R] {
	return &Indexer[T, R]{
		dataByKey: make(map[string]R),
	}
}

type Indexer[T, R any] struct {
	groupByKey   func(data T) ([]string, R)
	combineByKey func(a, b R) (R, error)

	dataByKey map[string]R
}

func (t *Indexer[T, R]) GroupByKey(f func(data T) ([]string, R)) {
	t.groupByKey = f
}

func (t *Indexer[T, R]) CombineByKey(f func(a, b R) (R, error)) {
	t.combineByKey = f
}

func (t *Indexer[T, R]) Append(data T) {
	key, result := t.groupByKey(data)

	index := t.indexName(key)
	if _, ok := t.dataByKey[index]; !ok {
		t.dataByKey[index] = result
	} else {
		t.dataByKey[index], _ = t.combineByKey(t.dataByKey[index], result)
	}
}

func (t *Indexer[T, R]) GetKey(key []string) R {
	index := t.indexName(key)
	return t.dataByKey[index]
}

func (t *Indexer[T, R]) indexName(key []string) string {
	return strings.Join(key, ":")
}
