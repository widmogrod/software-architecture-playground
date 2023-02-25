package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"sort"
	"sync"
)

func NewRepository2WithSchema() *RepositoryWithSchema {
	return &RepositoryWithSchema{
		store: make(map[string]Record[schema.Schema]),
	}
}

var _ Repository2[schema.Schema] = &RepositoryWithSchema{}

type RepositoryWithSchema struct {
	store map[string]Record[schema.Schema]
	mux   sync.Mutex
}

func (s *RepositoryWithSchema) Get(key string) (Record[schema.Schema], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	v, ok := s.store[key]
	if !ok {
		return Record[schema.Schema]{}, ErrNotFound
	}

	return v, nil
}

func (s *RepositoryWithSchema) UpdateRecords(x UpdateRecords[Record[schema.Schema]]) error {
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
			return fmt.Errorf("store.RepositoryWithSchema.UpdateRecords id=%s %d != %d %w",
				id, stored.Version, record.Version, ErrVersionConflict)
		}
	}

	for id, record := range x.Saving {
		record.Version += 1
		s.store[id] = record
	}

	return nil
}

func (s *RepositoryWithSchema) FindingRecords(query FindingRecords[Record[schema.Schema]]) (PageResult[Record[schema.Schema]], error) {
	records := make([]Record[schema.Schema], 0)
	for _, v := range s.store {
		records = append(records, v)
	}

	if query.Where != nil {
		newRecords := make([]Record[schema.Schema], 0)
		for _, record := range s.store {
			if predicate.Evaluate(query.Where.Predicate, record.Data, query.Where.Params) {
				newRecords = append(newRecords, record)
			}
		}
		records = newRecords
	}

	if len(query.Sort) > 0 {
		records = sortRecords(records, query.Sort)
	}

	if query.After != nil {
		found := false
		newRecords := make([]Record[schema.Schema], 0)
		for _, record := range records {
			if !found && record.ID == *query.After {
				found = true
				continue // we're interested in records after this one
			}
			if found {
				newRecords = append(newRecords, record)
			}
		}
		records = newRecords
		//} else if query.Before != nil {
		//	found := false
		//	newRecords := make([]Record[schema.Schema], 0)
		//	for _, record := range records {
		//		if !found && record.ID == *query.Before {
		//			found = true
		//			break // we accumulated records before this one, and now we're done
		//		}
		//
		//		if !found {
		//			newRecords = append(newRecords, record)
		//		}
		//	}
		//	records = newRecords
	}

	// Use limit to reduce number of records
	var next *FindingRecords[Record[schema.Schema]]
	if query.Limit > 0 {
		if len(records) > int(query.Limit) {
			//if query.Before != nil {
			//	records = records[len(records)-int(query.Limit):]
			//} else {
			records = records[:query.Limit]
			//}

			next = &FindingRecords[Record[schema.Schema]]{
				Where: query.Where,
				Sort:  query.Sort,
				Limit: query.Limit,
				After: &records[len(records)-1].ID,
				//Before: nil,
			}

		}
	}

	result := PageResult[Record[schema.Schema]]{
		Items: records,
		Next:  next,
	}

	return result, nil
}

func sortRecords(records []Record[schema.Schema], sortFields []SortField) []Record[schema.Schema] {
	sort.Slice(records, func(i, j int) bool {
		for _, sortField := range sortFields {
			fieldA := schema.Get(records[i].Data, sortField.Field)
			fieldB := schema.Get(records[j].Data, sortField.Field)
			cmp := schema.Compare(fieldA, fieldB)
			if !sortField.Descending {
				cmp = -cmp
			}
			if cmp != 0 {
				return cmp < 0
			}
		}
		return false
	})
	return records
}
