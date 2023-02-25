package storage

import "github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"

// Record could have two types (to think about it more):
// data records, which is current implementation
// index records, which is future implementation
//   - when two replicas have same aggregate rules, then during replication of logs, index can be reused
type Record[A any] struct {
	ID      string
	Data    A
	Version uint64
}

type FindingRecords[T any] struct {
	Where  *predicate.Where
	Sort   []SortField
	Limit  uint8
	Cursor *Cursor
}

type SortField struct {
	Field      string
	Descending bool
}

type Repository2[T any] interface {
	Get(key string) (Record[T], error)
	UpdateRecords(command UpdateRecords[Record[T]]) error
	FindingRecords(query FindingRecords[Record[T]]) (PageResult[Record[T]], error)
}
