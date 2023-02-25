package storage

import (
	"github.com/widmogrod/mkunion/x/schema"
)

type Aggregator[T, R any] interface {
	Append(data Record[T]) error
	GetVersionedIndices() map[string]Record[schema.Schema]
}
