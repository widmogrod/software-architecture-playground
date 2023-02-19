package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"log"
	"sync"
)

func NewRepositoryInMemory2[B, C any](
	indexer *Indexer[B, C],
) *RepositoryInMemory2[B, C] {
	return &RepositoryInMemory2[B, C]{
		store:   make(map[string]schema.Schema),
		indexer: indexer,
	}
}

type Record[A any] struct {
	ID      string
	Data    A
	Version uint64
}

type UpdateRecords2[B any] struct {
	Saving map[string]Record[B]
}

func (u *UpdateRecords2[B]) Save(x Record[B]) error {
	if u.Saving == nil {
		u.Saving = make(map[string]Record[B])
	}
	u.Saving[x.ID] = x
	return nil
}

type Repository2[B any, C any] interface {
	UpdateRecords(s UpdateRecords[Record[B]]) error
}

var _ Repository2[any, any] = &RepositoryInMemory2[any, any]{}

type RepositoryInMemory2[B any, C any] struct {
	store   map[string]schema.Schema
	mux     sync.Mutex
	indexer *Indexer[B, C]
}

//	func (r *RepositoryInMemory2[B, C]) GetAs(key string, x *A) error {
//		v, ok := r.store.Load(key)
//		if !ok {
//			return ErrNotFound
//		}
//
//		y, ok := v.(*A)
//		if !ok {
//			return fmt.Errorf("GetAs: %w want %T, got %T", ErrInvalidType, x, v)
//		}
//
//		x = y
//
//		return nil
//
// }
func (r *RepositoryInMemory2[B, C]) UpdateRecords(s UpdateRecords[Record[B]]) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for id, record := range s.Saving {
		log.Printf("saving %s %#v\n", id, record)
		r.indexer.Append(record.Data)
	}

	// Check if Version are correct
	for id, record := range s.Saving {
		stored, ok := r.store[id]
		if !ok {
			continue
		}

		object, err := schema.ToGo(stored)
		if err != nil {
			return fmt.Errorf("store.RepositoryInMemory2.UpdateRecords conversion error id=%s err=%s %w", id, err, ErrInternalError)
		}

		typed, ok := object.(Record[B])
		if !ok {
			return fmt.Errorf("store.RepositoryInMemory2.UpdateRecords conversion error id=%s %w", id, ErrInternalError)
		}

		if typed.Version > record.Version {
			return fmt.Errorf("store.RepositoryInMemory2.UpdateRecords id=%s %d > %d %w",
				id, typed.Version, record.Version, ErrVersionConflict)
		}
	}

	// Save
	for id, record := range s.Saving {
		schemed := schema.FromGo(record)
		r.store[id] = schemed
	}

	for index, unversionedData := range r.indexer.dataByKey {
		// Index don't need to be unversionedData, because it's constructed from unversionedData that are versioned
		// if save will be rejected for them, that means that index is not valid anymore
		// but if save will be accepted, then index is valid
		log.Printf("index %s %#v\n", index, unversionedData)
		schemed := schema.FromGo(unversionedData)
		r.store[index] = schemed
	}

	return nil
}

// ReindexAll is used to reindex all records with a provided indexer definition
// Example: when indexer is created, it's empty, so it needs to be filled with all records
// Example: when indexer definition is changed, it needs to be reindexed
// Example: when indexer is corrupted, it needs to be reindexed
//
// How it works?
// 1. It's called by the user
// 2. It's called by the system when it detects that indexer is corrupted
// 3. It's called by the system when it detects that indexer definition is changed
//
// How it's implemented?
//  1. Create index from snapshot of all records. Because it's snapshot, changes are not applied.
//  2. In parallel process stream of changes from give point of time.
//  3. Indexer must be idempotent, so same won't be indexed twice.
//  4. When indexer detects same record with new Version, it retracts old Version and accumulates new Version.
//  5. When it's done, it's ready to be used
//  6. When indices are set up as synchronous, then every change is indexed immediately.
//     But, because synchronous index is from point of time, it needs to trigger reindex.
//     Which imply that indexer myst know when index was created, so it can know when to stop rebuilding process.
//     This implies control plane. Versions of records should follow monotonically increasing order, that way it will be easier to detect when index is up to date.
func (r *RepositoryInMemory2[B, C]) ReindexAll() {
	panic("not implemented")
}
