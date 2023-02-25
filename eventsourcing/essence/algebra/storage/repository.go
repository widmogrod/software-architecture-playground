package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
)

var (
	ErrNotFound        = fmt.Errorf("not found")
	ErrInvalidType     = fmt.Errorf("invalid type")
	ErrVersionConflict = fmt.Errorf("Version conflict")
	ErrInternalError   = fmt.Errorf("internal error")
)

type Repository[A any] interface {
	GetAs(key string, x *A) error
	UpdateRecords(s UpdateRecords[any]) error

	Get(key string) (A, error)
	Set(key string, value A) error
	Delete(key string) error
	GetOrNew(s string) (A, error)

	FindAllKeyEqual(key string, value string) (PageResult[A], error)
}

func NewRepositoryInMemory[A any](new func() A) *RepositoryInMemory[A] {
	return &RepositoryInMemory[A]{
		store: sync.Map{},
		new:   new,
	}
}

var _ Repository[any] = (*RepositoryInMemory[any])(nil)

type RepositoryInMemory[A any] struct {
	store sync.Map
	new   func() A
}

func (r *RepositoryInMemory[A]) GetAs(key string, x *A) error {
	v, ok := r.store.Load(key)
	if !ok {
		return ErrNotFound
	}

	y, ok := v.(*A)
	if !ok {
		return fmt.Errorf("GetAs: %w want %T, got %T", ErrInvalidType, x, v)
	}

	x = y

	return nil

}

func (r *RepositoryInMemory[A]) UpdateRecords(s UpdateRecords[any]) error {
	for id, record := range s.Saving {
		err := r.Set(id, record.(A))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RepositoryInMemory[A]) Get(key string) (A, error) {
	var a A
	err := r.GetAs(key, &a)
	return a, err
}

func (r *RepositoryInMemory[A]) Set(key string, value A) error {
	r.store.Store(key, value)
	return nil
}

func (r *RepositoryInMemory[A]) Delete(key string) error {
	r.store.Delete(key)
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

type Cursor = string

type PageResult[A any] struct {
	Items []A
	Next  *FindingRecords[A]
	//Prev  *FindingRecords[A]
}

func (a PageResult[A]) HasNext() bool {
	return a.Next != nil
}

//func (a PageResult[A]) HasPrev() bool {
//	return a.Prev != nil
//}

func (r *RepositoryInMemory[A]) FindAllKeyEqual(key string, value string) (PageResult[A], error) {
	result := PageResult[A]{
		Next: nil,
	}

	r.store.Range(func(k, v interface{}) bool {
		sch := schema.FromGo(v)
		if m, ok := sch.(*schema.Map); ok {
			for _, kv := range m.Field {
				valueOfKey := schema.MustToGo(kv.Value)
				if kv.Name == key && valueOfKey == value {
					result.Items = append(result.Items, v.(A))
				}
			}
		}

		return true
	})

	return result, nil
}
