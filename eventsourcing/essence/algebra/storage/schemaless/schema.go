package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"sort"
	"sync"
)

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store: make(map[string]schema.Schema),
	}
}

var _ Repository[schema.Schema] = &InMemoryRepository{}

type InMemoryRepository struct {
	store map[string]schema.Schema
	mux   sync.Mutex
}

func (s *InMemoryRepository) Get(recordID, recordType string) (Record[schema.Schema], error) {
	result, err := s.FindingRecords(FindingRecords[Record[schema.Schema]]{
		RecordType: recordType,
		Where: predicate.MustWhere("ID = :id", predicate.ParamBinds{
			":id": schema.MkString(recordID),
		}),
		Limit: 1,
	})
	if err != nil {
		return Record[schema.Schema]{}, err
	}

	if len(result.Items) == 0 {
		return Record[schema.Schema]{}, ErrNotFound
	}

	return result.Items[0], nil
}

func (s *InMemoryRepository) UpdateRecords(x UpdateRecords[Record[schema.Schema]]) error {
	if x.IsEmpty() {
		return fmt.Errorf("store.InMemoryRepository.UpdateRecords: empty command %w", ErrEmptyCommand)
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	for _, record := range x.Saving {
		stored, ok := s.store[record.ID+record.Type]
		if !ok {
			// new record, should have version 1
			// and since few lines below we increment version
			// we need to set it to 0
			record.Version = 0
			continue
		}

		storedVersion := schema.As[uint16](schema.Get(stored, "Version"), 0)

		if storedVersion != record.Version {
			return fmt.Errorf("store.InMemoryRepository.UpdateRecords ID=%s Type=%s %d != %d %w",
				record.ID, record.Type, storedVersion, record.Version, ErrVersionConflict)
		}
	}

	for _, record := range x.Saving {
		record.Version += 1
		s.store[record.ID+record.Type] = s.fromTyped(record)
	}

	for _, id := range x.Deleting {
		delete(s.store, id.ID+id.Type)
	}

	return nil
}

func (s *InMemoryRepository) FindingRecords(query FindingRecords[Record[schema.Schema]]) (PageResult[Record[schema.Schema]], error) {
	records := make([]schema.Schema, 0)
	for _, v := range s.store {
		records = append(records, v)
	}

	if query.RecordType != "" {
		newRecords := make([]schema.Schema, 0)
		for _, record := range records {
			if predicate.EvaluateEqual(record, "Type", query.RecordType) {
				newRecords = append(newRecords, record)
			}
		}
		records = newRecords
	}

	if query.Where != nil {
		newRecords := make([]schema.Schema, 0)
		for _, record := range s.store {
			if predicate.Evaluate(query.Where.Predicate, record, query.Where.Params) {
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
		newRecords := make([]schema.Schema, 0)
		for _, record := range records {
			if predicate.EvaluateEqual(record, "ID", *query.After) {
				found = true
				continue // we're interested in records after this one
			}
			if found {
				newRecords = append(newRecords, record)
			}
		}
		records = newRecords
	}

	typedRecords := make([]Record[schema.Schema], 0)
	for _, record := range records {
		typed, err := s.toTyped(record)
		if err != nil {
			return PageResult[Record[schema.Schema]]{}, err
		}
		typedRecords = append(typedRecords, typed)
	}

	// Use limit to reduce number of records
	var next *FindingRecords[Record[schema.Schema]]
	if query.Limit > 0 {
		if len(typedRecords) > int(query.Limit) {
			typedRecords = typedRecords[:query.Limit]

			next = &FindingRecords[Record[schema.Schema]]{
				Where: query.Where,
				Sort:  query.Sort,
				Limit: query.Limit,
				After: &typedRecords[len(typedRecords)-1].ID,
			}
		}
	}

	result := PageResult[Record[schema.Schema]]{
		Items: typedRecords,
		Next:  next,
	}

	return result, nil
}

func (s *InMemoryRepository) fromTyped(record Record[schema.Schema]) *schema.Map {
	return schema.MkMap(
		schema.MkField("ID", schema.MkString(record.ID)),
		schema.MkField("Type", schema.MkString(record.Type)),
		schema.MkField("Data", record.Data),
		schema.MkField("Version", schema.MkInt(int(record.Version))),
	)
}

func (s *InMemoryRepository) toTyped(record schema.Schema) (Record[schema.Schema], error) {
	typed := Record[schema.Schema]{
		ID:      schema.As[string](schema.Get(record, "ID"), "record-id-corrupted"),
		Type:    schema.As[string](schema.Get(record, "Type"), "record-type-corrupted"),
		Data:    schema.Get(record, "Data"),
		Version: schema.As[uint16](schema.Get(record, "Version"), 0),
	}
	if typed.Type == "record-id-corrupted" &&
		typed.ID == "record-id-corrupted" &&
		typed.Version == 0 {
		return Record[schema.Schema]{}, fmt.Errorf("store.InMemoryRepository.FindingRecords corrupted record: %v", record)
	}
	return typed, nil
}

func sortRecords(records []schema.Schema, sortFields []SortField) []schema.Schema {
	sort.Slice(records, func(i, j int) bool {
		for _, sortField := range sortFields {
			fieldA := schema.Get(records[i], sortField.Field)
			fieldB := schema.Get(records[j], sortField.Field)
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
