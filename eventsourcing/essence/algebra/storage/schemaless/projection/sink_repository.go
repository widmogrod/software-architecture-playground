package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"sync"
	"time"
)

func NewRepositorySink(recordType string, store schemaless.Repository[schema.Schema]) *RepositorySink {
	sink := &RepositorySink{
		flushWhenBatchSize: 0,
		flushWhenDuration:  1 * time.Second,

		store:      store,
		recordType: recordType,

		bufferSaving:   map[string]schemaless.Record[schema.Schema]{},
		bufferDeleting: map[string]schemaless.Record[schema.Schema]{},
	}

	// TODO consider to move it outside, for explicit managment
	// or make hooks on records, that will wait for flush
	sink.FlushOnTime()

	return sink
}

type RepositorySink struct {
	lock sync.Mutex

	flushWhenBatchSize int
	flushWhenDuration  time.Duration

	bufferSaving   map[string]schemaless.Record[schema.Schema]
	bufferDeleting map[string]schemaless.Record[schema.Schema]

	store      schemaless.Repository[schema.Schema]
	recordType string
}

func (s *RepositorySink) FlushOnTime() {
	go func() {
		ticker := time.NewTicker(s.flushWhenDuration)
		for range ticker.C {
			s.flush()
		}
	}()
}

func (s *RepositorySink) Process(x Item, returning func(Item)) error {
	s.lock.Lock()
	s.bufferSaving[x.Key] = schemaless.Record[schema.Schema]{
		ID:      x.Key,
		Type:    s.recordType,
		Data:    x.Data,
		Version: 0,
	}
	s.lock.Unlock()
	if len(s.bufferSaving)+len(s.bufferDeleting) >= s.flushWhenBatchSize {
		return s.flush()
	}

	return nil
}

func (s *RepositorySink) Retract(x Item, returning func(Item)) error {
	s.lock.Lock()
	s.bufferDeleting[x.Key] = schemaless.Record[schema.Schema]{
		ID:      x.Key,
		Type:    s.recordType,
		Data:    x.Data,
		Version: 0,
	}
	s.lock.Unlock()

	if len(s.bufferSaving)+len(s.bufferDeleting) >= s.flushWhenBatchSize {
		return s.flush()
	}

	return nil
}

func (s *RepositorySink) flush() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.bufferSaving)+len(s.bufferDeleting) == 0 {
		fmt.Println("nothing to flush")
		return nil
	}

	err := s.store.UpdateRecords(schemaless.UpdateRecords[schemaless.Record[schema.Schema]]{
		UpdatingPolicy: schemaless.PolicyOverwriteServerChanges,
		Saving:         s.bufferSaving,
		Deleting:       s.bufferDeleting,
	})
	if err != nil {
		return err
	}
	fmt.Println("flushed:")
	for id, record := range s.bufferSaving {
		fmt.Println("- saved", id, record)
	}
	for id, record := range s.bufferDeleting {
		fmt.Println("- deleted", id, record)
	}

	s.bufferSaving = map[string]schemaless.Record[schema.Schema]{}
	s.bufferDeleting = map[string]schemaless.Record[schema.Schema]{}

	return nil
}
