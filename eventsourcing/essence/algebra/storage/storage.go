package storage

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
)

type Storage[T any] interface {
	GetAs(id string, x *T) error
}

func RetriveID[T any](s Storage[T], id string) (T, error) {
	var x T
	err := s.GetAs(id, &x)
	if err != nil {
		return x, err
	}
	return x, nil
}

type UpdateRecords[T any] struct {
	Saving map[string]T
}

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
