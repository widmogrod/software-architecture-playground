package storage

// Record could have two types (to think about it more):
// data records, which is current implementation
// index records, which is future implementation
//   - when two replicas have same aggregate rules, then during replication of logs, index can be reused
type Record[A any] struct {
	ID      string
	Data    A
	Version uint64
}

type Repository2[B any] interface {
	Get(key string) (Record[B], error)
	UpdateRecords(s UpdateRecords[Record[B]]) error
}
