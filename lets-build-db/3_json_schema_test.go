package lets_build_db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ Identifiable = &Person{}

func TestJsonSchema(t *testing.T) {
	// Because schema management was "delegated" to know formats like JSON or Protobuf
	// it makes implementation somewhat simpler at the beginning...
	// reflection of the schema is necessary to formulate access to record,
	// but that can be changed by introducing predicate DSL that is based on [path]
	//
	// Assumption about this simple driver is that
	// - Client is responsible for validation
	// - Driver may be instrumented by schema registry that validates whenever schema was created
	// - Finding records is much harder, because now value bytes, to to combad this
	//   Creation of indexes and dedicated API that expects DSL
	//   and will prevent any search (except lookup) that don't have index
	db := NewDbValueSchema()
	record := &Person{
		Id:   "123",
		Name: "Gaha",
	}
	err := db.Insert(FromJson(record))
	assert.NoError(t, err)

	//db.CreateIndex(PathSelector{
	//	Path: []string{"Name"},
	//})
	//
	//result, err := db.Select(FromJson(&Person{}), Predicate2{
	//	Path:  []string{"Id"},
	//	Value: Value{Eq: "123"},
	//})
	//
	//assert.NoError(t, err)
	//assert.Equal(t, record, result)
}
