package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"log"
	"strings"
	"sync"
)

func RecordAs[A any](record Record[schema.Schema]) (Record[A], error) {
	object, err := schema.ToGo(record.Data)
	if err != nil {
		var a A
		return Record[A]{}, fmt.Errorf("store.GetSchemaAs[%T] schema conversion failed. %s. %w", a, err, ErrInternalError)
	}

	typed, ok := object.(A)
	if !ok {
		var a A
		return Record[A]{}, fmt.Errorf("store.GetSchemaAs[%T] type assertion got %T. %w", a, object, ErrInternalError)
	}

	return Record[A]{
		ID:      record.ID,
		Data:    typed,
		Version: record.Version,
	}, nil
}

func NewInMemorySchemaStore() *InMemorySchemaStore {
	return &InMemorySchemaStore{
		store: make(map[string]Record[schema.Schema]),
	}
}

var _ Repository2[schema.Schema] = &InMemorySchemaStore{}

type InMemorySchemaStore struct {
	store map[string]Record[schema.Schema]
	mux   sync.Mutex
}

func (s *InMemorySchemaStore) Get(key string) (Record[schema.Schema], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	v, ok := s.store[key]
	if !ok {
		return Record[schema.Schema]{}, ErrNotFound
	}

	return v, nil
}

func (s *InMemorySchemaStore) UpdateRecords(x UpdateRecords[Record[schema.Schema]]) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for id, record := range x.Saving {
		stored, ok := s.store[id]
		if !ok {
			// new record, should have version 1
			// and since few lines below we increment version
			// we need to set it to 0
			record.Version = 0
			continue
		}

		if stored.Version != record.Version {
			return fmt.Errorf("store.InMemorySchemaStore.UpdateRecords id=%s %d != %d %w",
				id, stored.Version, record.Version, ErrVersionConflict)
		}
	}

	for id, record := range x.Saving {
		record.Version += 1
		s.store[id] = record
	}

	return nil
}

func NewRepositoryInMemory2[B, C any](
	storage Repository2[schema.Schema],
	indexer Aggregator[B, C],
) *RepositoryInMemory2[B, C] {
	return &RepositoryInMemory2[B, C]{
		storage:   storage,
		aggregate: indexer,
	}
}

// Record could have two types (to think about it more):
// data records, which is current implementation
// index records, which is future implementation
//   - when two replicas have same aggregate rules, then during replication of logs, index can be reused
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

type Repository2[B any] interface {
	Get(key string) (Record[B], error)
	UpdateRecords(s UpdateRecords[Record[B]]) error
}

var _ Repository2[any] = &RepositoryInMemory2[any, any]{}

type RepositoryInMemory2[B any, C any] struct {
	mux       sync.Mutex
	storage   Repository2[schema.Schema]
	aggregate Aggregator[B, C]
}

func (r *RepositoryInMemory2[B, C]) Get(key string) (Record[B], error) {
	v, err := r.storage.Get(key)
	if err != nil {
		return Record[B]{}, fmt.Errorf("store.RepositoryInMemory2.Get storage error id=%s %w", key, err)
	}

	object, err := schema.ToGo(v.Data)
	if err != nil {
		return Record[B]{}, fmt.Errorf("store.RepositoryInMemory2.Get schema conversion error id=%s err=%s %w", key, err, ErrInternalError)
	}

	typed, ok := object.(B)
	if !ok {
		return Record[B]{}, fmt.Errorf("store.RepositoryInMemory2.Get conversion error id=%s %w", key, ErrInternalError)

	}

	return Record[B]{
		ID:      v.ID,
		Data:    typed,
		Version: v.Version,
	}, nil
}

func (r *RepositoryInMemory2[B, C]) UpdateRecords(s UpdateRecords[Record[B]]) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	schemas := UpdateRecords[Record[schema.Schema]]{
		Saving: make(map[string]Record[schema.Schema]),
	}

	for id, record := range s.Saving {

		// TODO fix me
		if strings.HasPrefix(id, "game:") {
			log.Printf("saving %s %#v\n", id, record)
			err := r.aggregate.Append(record.Data)
			if err != nil {
				return fmt.Errorf("store.RepositoryInMemory2.UpdateRecords aggregate.Append %w", err)
			}
		}

		schemed := schema.FromGo(record)

		schemas.Saving[id] = Record[schema.Schema]{
			ID:      record.ID,
			Data:    schemed,
			Version: record.Version,
		}
	}

	for index, versionedData := range r.aggregate.GetVersionedIndices() {
		log.Printf("index %s %#v\n", index, versionedData)
		schemas.Saving[versionedData.ID] = versionedData
	}

	err := r.storage.UpdateRecords(schemas)
	if err != nil {
		return fmt.Errorf("store.RepositoryInMemory2.UpdateRecords schemas store err %w", err)
	}

	return nil
}

// ReindexAll is used to reindex all records with a provided aggregate definition
// Example: when aggregate is created, it's empty, so it needs to be filled with all records
// Example: when aggregate definition is changed, it needs to be reindexed
// Example: when aggregate is corrupted, it needs to be reindexed
//
// How it works?
// 1. It's called by the user
// 2. It's called by the system when it detects that aggregate is corrupted
// 3. It's called by the system when it detects that aggregate definition is changed
//
// How it's implemented?
//  1. Create index from snapshot of all records. Because it's snapshot, changes are not applied.
//  2. In parallel process stream of changes from give point of time.
//  3. KayedAggregate must be idempotent, so same won't be indexed twice.
//  4. When aggregate detects same record with new Version, it retracts old Version and accumulates new Version.
//  5. When it's done, it's ready to be used
//  6. When indices are set up as synchronous, then every change is indexed immediately.
//     But, because synchronous index is from point of time, it needs to trigger reindex.
//     Which imply that aggregate myst know when index was created, so it can know when to stop rebuilding process.
//     This implies control plane. Versions of records should follow monotonically increasing order, that way it will be easier to detect when index is up to date.
func (r *RepositoryInMemory2[B, C]) ReindexAll() {
	panic("not implemented")
}
