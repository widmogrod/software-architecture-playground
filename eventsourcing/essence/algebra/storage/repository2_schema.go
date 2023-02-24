package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
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
