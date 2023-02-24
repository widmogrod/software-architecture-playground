package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"testing"
)

type exampleRecord struct {
	Name string
	Age  int
}

func TestNewRepository2WithSchema(t *testing.T) {
	repo := NewRepository2WithSchema()
	assert.NotNil(t, repo)

	err := repo.UpdateRecords(UpdateRecords[Record[schema.Schema]]{
		Saving: map[string]Record[schema.Schema]{
			"123": {
				ID: "123",
				Data: schema.FromGo(exampleRecord{
					Name: "John",
					Age:  20,
				}),
			},
			"12": {
				ID: "124",
				Data: schema.FromGo(exampleRecord{
					Name: "Jane",
					Age:  30,
				}),
			},
			"313": {
				ID: "313",
				Data: schema.FromGo(exampleRecord{
					Name: "Alice",
					Age:  39,
				}),
			},
		},
	})
	assert.NoError(t, err)

	result, err := repo.FindingRecords(FindingRecords[Record[schema.Schema]]{
		Where: predicate.MustQuery(
			"Age > :age",
			map[string]schema.Schema{
				":age": schema.MkInt(20),
			}),
		Sort: []SortField{
			{
				Field:      "Name",
				Descending: true,
			},
		},
		Limit: 2,
	})
	assert.NoError(t, err)
	assert.False(t, result.HasNext())
	if assert.Len(t, result.Items, 2) {
		assert.Equal(t, "Alice", schema.As[string](schema.Get(result.Items[0].Data, "Name"), "no-name"))
		assert.Equal(t, "Jane", schema.As[string](schema.Get(result.Items[1].Data, "Name"), "no-name"))
	}
}
