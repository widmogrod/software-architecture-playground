package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"time"
)

func NewRepositorySink(recordType string, store schemaless.Repository[schema.Schema]) *RepositorySink {
	sink := &RepositorySink{
		flushWhenBatchSize: 10,
		flushWhenDuration:  1 * time.Second,

		store:      store,
		recordType: recordType,

		bufferSaving:   map[string]schemaless.Record[schema.Schema]{},
		bufferDeleting: map[string]schemaless.Record[schema.Schema]{},
	}

	sink.FlushOnTime()

	return sink
}

type RepositorySink struct {
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

func (s *RepositorySink) Process(msg Message, returning func(Message) error) error {
	err := MustMatchMessage(
		msg,
		func(x *Combine) error {
			s.bufferSaving[x.Key] = schemaless.Record[schema.Schema]{
				ID:      x.Key,
				Type:    s.recordType,
				Data:    x.Data,
				Version: 0,
			}
			return nil
		},
		func(x *Retract) error {
			s.bufferDeleting[x.Key] = schemaless.Record[schema.Schema]{
				ID:      x.Key,
				Type:    s.recordType,
				Data:    x.Data,
				Version: 0,
			}
			return nil
		},
		func(x *Both) error {
			s.bufferSaving[x.Key] = schemaless.Record[schema.Schema]{
				ID:      x.Key,
				Type:    s.recordType,
				Data:    x.Combine.Data,
				Version: 0,
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	if len(s.bufferSaving)+len(s.bufferDeleting) >= s.flushWhenBatchSize {
		return s.flush()
	}

	return nil

}

func (s *RepositorySink) flush() error {
	if len(s.bufferSaving)+len(s.bufferDeleting) == 0 {
		return nil
	}

	err := s.store.UpdateRecords(schemaless.UpdateRecords[schemaless.Record[schema.Schema]]{
		Saving:   s.bufferSaving,
		Deleting: s.bufferDeleting,
	})
	if err != nil {
		return err
	}

	s.bufferSaving = map[string]schemaless.Record[schema.Schema]{}
	s.bufferDeleting = map[string]schemaless.Record[schema.Schema]{}
	return nil

}
