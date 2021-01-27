package inmemorystore

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/algebra/aggregate/store"
	"sync"
)

func NewInMemoryStore() *inmemdatastor {
	return &inmemdatastor{data: sync.Map{}}
}

type inmemdatastor struct {
	data sync.Map
}

func (i *inmemdatastor) ReadChanges(_ context.Context, aggregateID string) ([]runtime.Change, error) {
	data, found := i.data.Load(aggregateID)
	if !found {
		return nil, store.ErrNotFound
	}

	return data.([]runtime.Change), nil
}

func (i *inmemdatastor) AppendChanges(_ context.Context, aggregateID string, version uint64, changes []runtime.Change) error {
	i.data.Store(aggregateID, changes)
	return nil
}
