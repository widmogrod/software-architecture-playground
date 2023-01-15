package storage

import (
	"fmt"
	"sync"
)

func NewRepositoryInMemory[A any](new func() A) *RepositoryInMemory[A] {
	return &RepositoryInMemory[A]{
		store: sync.Map{},
		new:   new,
	}
}

type RepositoryInMemory[A any] struct {
	store sync.Map
	new   func() A
}

var ErrNotFound = fmt.Errorf("not found")

func (r *RepositoryInMemory[A]) Get(key string) (A, error) {
	v, ok := r.store.Load(key)
	if !ok {
		var a A
		return a, ErrNotFound
	}
	return v.(A), nil
}

func (r *RepositoryInMemory[A]) Set(key string, value A) error {
	r.store.Store(key, value)
	return nil
}

func (r *RepositoryInMemory[A]) GetOrNew(s string) (A, error) {
	v, err := r.Get(s)
	if err == nil {
		return v, nil
	}

	if err != nil && err != ErrNotFound {
		var a A
		return a, err
	}

	v = r.new()

	err = r.Set(s, v)
	if err != nil {
		var a A
		return a, err
	}

	return v, nil
}
