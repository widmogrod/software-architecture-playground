package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"log"
	"sync"
)

func NewRepositoryWithIndexer[B, C any](
	storage Repository2[schema.Schema],
	indexer Aggregator[B, C],
) *RepositoryWithAggregator[B, C] {
	return &RepositoryWithAggregator[B, C]{
		storage:   storage,
		aggregate: indexer,
	}
}

var _ Repository2[any] = &RepositoryWithAggregator[any, any]{}

type RepositoryWithAggregator[B any, C any] struct {
	mux       sync.Mutex
	storage   Repository2[schema.Schema]
	aggregate Aggregator[B, C]
}

func (r *RepositoryWithAggregator[B, C]) Get(key string) (Record[B], error) {
	v, err := r.storage.Get(key)
	if err != nil {
		return Record[B]{}, fmt.Errorf("store.RepositoryWithAggregator.Get storage error id=%s %w", key, err)
	}

	typed, err := RecordAs[B](v)
	if err != nil {
		return Record[B]{}, fmt.Errorf("store.RepositoryWithAggregator.Get type assertion error id=%s %w", key, err)
	}

	return typed, nil
}

func (r *RepositoryWithAggregator[B, C]) UpdateRecords(s UpdateRecords[Record[B]]) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	schemas := UpdateRecords[Record[schema.Schema]]{
		Saving: make(map[string]Record[schema.Schema]),
	}

	for id, record := range s.Saving {
		err := r.aggregate.Append(record)
		if err != nil {
			return fmt.Errorf("store.RepositoryWithAggregator.UpdateRecords aggregate.Append %w", err)
		}

		schemed := schema.FromGo(record.Data)
		schemas.Saving[id] = Record[schema.Schema]{
			ID:      record.ID,
			Type:    record.Type,
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
		return fmt.Errorf("store.RepositoryWithAggregator.UpdateRecords schemas store err %w", err)
	}

	return nil
}

func (r *RepositoryWithAggregator[B, C]) FindingRecords(query FindingRecords[Record[B]]) (PageResult[Record[B]], error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	found, err := r.storage.FindingRecords(FindingRecords[Record[schema.Schema]]{
		Where: query.Where,
		Sort:  query.Sort,
		Limit: query.Limit,
		After: query.After,
	})
	if err != nil {
		return PageResult[Record[B]]{}, fmt.Errorf("store.RepositoryWithAggregator.FindingRecords storage error %w", err)
	}

	result := PageResult[Record[B]]{
		Items: nil,
		Next:  nil,
	}

	if found.HasNext() {
		result.Next = &FindingRecords[Record[B]]{
			Where: query.Where,
			Sort:  query.Sort,
			Limit: query.Limit,
			After: found.Next.After,
		}
	}

	for _, item := range found.Items {
		typed, err := RecordAs[B](item)
		if err != nil {
			return PageResult[Record[B]]{}, fmt.Errorf("store.RepositoryWithAggregator.FindingRecords RecordAs error id=%s %w", item.ID, err)
		}

		result.Items = append(result.Items, typed)
	}

	return result, nil
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
func (r *RepositoryWithAggregator[B, C]) ReindexAll() {
	panic("not implemented")
}
