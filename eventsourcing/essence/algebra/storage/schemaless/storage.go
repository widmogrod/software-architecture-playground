package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
)

type RecordType = string
type Repository[T any] interface {
	Get(recordID string, recordType RecordType) (Record[T], error)
	UpdateRecords(command UpdateRecords[Record[T]]) error
	FindingRecords(query FindingRecords[Record[T]]) (PageResult[Record[T]], error)
}

var (
	ErrNotFound        = fmt.Errorf("not found")
	ErrEmptyCommand    = fmt.Errorf("empty command")
	ErrInvalidType     = fmt.Errorf("invalid type")
	ErrVersionConflict = fmt.Errorf("version conflict")
	ErrInternalError   = fmt.Errorf("internal error")
)

// Record could have two types (to think about it more):
// data records, which is current implementation
// index records, which is future implementation
//   - when two replicas have same aggregator rules, then during replication of logs, index can be reused
type Record[A any] struct {
	ID      string
	Type    string
	Data    A
	Version uint16
}

type UpdatingPolicy uint

const (
	PolicyIfServerNotChanged UpdatingPolicy = iota
	PolicyOverwriteServerChanges
)

type (
	UpdateRecords[T any] struct {
		UpdatingPolicy UpdatingPolicy
		Saving         map[string]T
		Deleting       map[string]T
	}
)

func (s UpdateRecords[T]) IsEmpty() bool {
	return len(s.Saving) == 0 && len(s.Deleting) == 0
}

type (
	FindingRecords[T any] struct {
		RecordType string
		Where      *predicate.WherePredicates
		Sort       []SortField
		Limit      uint8
		After      *Cursor
		//Before *Cursor
	}

	SortField struct {
		Field      string
		Descending bool
	}

	Cursor = string

	PageResult[A any] struct {
		Items []A
		Next  *FindingRecords[A]
	}
)

func (a PageResult[A]) HasNext() bool {
	return a.Next != nil
}

type Storage[T any] interface {
	GetAs(id string, x *T) error
}

func Save[T any](xs ...Record[T]) UpdateRecords[Record[T]] {
	m := make(map[string]Record[T])
	for _, x := range xs {
		m[x.ID+":"+x.Type] = x
	}

	return UpdateRecords[Record[T]]{
		Saving: m,
	}
}

func Delete[T any](xs ...Record[T]) UpdateRecords[Record[T]] {
	m := make(map[string]Record[T])
	for _, x := range xs {
		m[x.ID+":"+x.Type] = x
	}

	return UpdateRecords[Record[T]]{
		Deleting: m,
	}
}

func SaveAndDelete(saving, deleting UpdateRecords[Record[schema.Schema]]) UpdateRecords[Record[schema.Schema]] {
	return UpdateRecords[Record[schema.Schema]]{
		Saving:   saving.Saving,
		Deleting: deleting.Deleting,
	}
}

func RecordAs[A any](record Record[schema.Schema]) (Record[A], error) {
	typed, err := ConvertAs[A](record.Data)
	if err != nil {
		return Record[A]{}, err
	}

	return Record[A]{
		ID:      record.ID,
		Type:    record.Type,
		Data:    typed,
		Version: record.Version,
	}, nil
}

func ConvertAs[A any](x schema.Schema) (A, error) {
	var a A
	var result any
	var err error

	switch any(a).(type) {
	case int:
		result = schema.As[int](x, any(a).(int))
	case int8:
		result = schema.As[int8](x, any(a).(int8))
	case int16:
		result = schema.As[int16](x, any(a).(int16))
	case int32:
		result = schema.As[int32](x, any(a).(int32))
	case int64:
		result = schema.As[int64](x, any(a).(int64))
	case uint:
		result = schema.As[uint](x, any(a).(uint))
	case uint8:
		result = schema.As[uint8](x, any(a).(uint8))
	case uint16:
		result = schema.As[uint16](x, any(a).(uint16))
	case uint32:
		result = schema.As[uint32](x, any(a).(uint32))
	case uint64:
		result = schema.As[uint64](x, any(a).(uint64))
	case float32:
		result = schema.As[float32](x, any(a).(float32))
	case float64:
		result = schema.As[float64](x, any(a).(float64))
	case string:
		result = schema.As[string](x, any(a).(string))
	case bool:
		result = schema.As[bool](x, any(a).(bool))
	case []byte:
		result = schema.As[[]byte](x, any(a).([]byte))
	default:
		if any(a) == nil {
			result, err = schema.ToGo(x)
		} else {
			result, err = schema.ToGo(x, schema.WithExtraRules(schema.WhenPath(nil, schema.UseStruct(a))))
		}

		if err != nil {
			var a A
			return a, fmt.Errorf("store.RecordAs[%T] schema conversion failed. %s. %w", a, err, ErrInternalError)
		}
	}

	typed, ok := result.(A)
	if !ok {
		var a A
		return a, fmt.Errorf("store.RecordAs[%T] type assertion got %T. %w", a, result, ErrInternalError)
	}

	return typed, nil
}
