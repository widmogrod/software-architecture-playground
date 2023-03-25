package schemaless

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

// refactored exampleUpdateRecords that use Save
var exampleUpdateRecords = Save(
	Record[schema.Schema]{
		ID:   "123",
		Type: "exampleRecord",
		Data: schema.FromGo(exampleRecord{
			Name: "John",
			Age:  20,
		}),
	},
	Record[schema.Schema]{
		ID:   "124",
		Type: "exampleRecord",
		Data: schema.FromGo(exampleRecord{
			Name: "Jane",
			Age:  30,
		}),
	},
	Record[schema.Schema]{
		ID:   "313",
		Type: "exampleRecord",
		Data: schema.FromGo(exampleRecord{
			Name: "Alice",
			Age:  39,
		}),
	},
	Record[schema.Schema]{
		ID:   "1234",
		Type: "exampleRecord",
		Data: schema.FromGo(exampleRecord{
			Name: "Bob",
			Age:  40,
		}),
	},
	Record[schema.Schema]{
		ID:   "3123",
		Type: "exampleRecord",
		Data: schema.FromGo(exampleRecord{
			Name: "Zarlie",
			Age:  39,
		}),
	},
)

func TestNewRepository2WithSchema(t *testing.T) {
	repo := NewInMemoryRepository()
	assert.NotNil(t, repo)

	err := repo.UpdateRecords(exampleUpdateRecords)
	assert.NoError(t, err)

	result, err := repo.FindingRecords(FindingRecords[Record[schema.Schema]]{
		Where: predicate.MustWhere(
			"Data.Age > :age AND Data.Age < :maxAge",
			predicate.ParamBinds{
				":age":    schema.MkInt(20),
				":maxAge": schema.MkInt(40),
			}),
		Sort: []SortField{
			{
				Field:      "Data.Name",
				Descending: true,
			},
		},
		Limit: 2,
	})
	assert.NoError(t, err)

	if assert.Len(t, result.Items, 2, "first page should have 2 items") {
		assert.Equal(t, "Alice", schema.AsDefault[string](schema.Get(result.Items[0].Data, "Name"), "no-name"))
		assert.Equal(t, "Jane", schema.AsDefault[string](schema.Get(result.Items[1].Data, "Name"), "no-name"))
	}

	if assert.True(t, result.HasNext(), "should have next page of results") {
		nextResult, err := repo.FindingRecords(*result.Next)

		assert.NoError(t, err)
		if assert.Len(t, nextResult.Items, 1, "second page should have 1 item") {
			assert.Equal(t, "Zarlie", schema.AsDefault[string](schema.Get(nextResult.Items[0].Data, "Name"), "no-name"))

			//// find last before
			//if assert.True(t, nextResult.HasPrev(), "should have previous page of results") {
			//	beforeResult, err := repo.FindingRecords(*nextResult.Prev)
			//	assert.NoError(t, err)
			//
			//	if assert.Len(t, beforeResult.Items, 1, "before page should have 1 item") {
			//		assert.Equal(t, "Jane", schema.As[string](schema.Get(beforeResult.Items[0].Data, "Name"), "no-name"))
			//	}
			//}
		}
	}
}

func TestRepositoryWithSchema_UpdateRecords_Deletion(t *testing.T) {
	repo := NewInMemoryRepository()
	assert.NotNil(t, repo)

	err := repo.UpdateRecords(exampleUpdateRecords)
	assert.NoError(t, err)

	result, err := repo.FindingRecords(FindingRecords[Record[schema.Schema]]{})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 5, "should have 5 records")
	assert.False(t, result.HasNext(), "should not have next page of results")

	deleting := map[string]Record[schema.Schema]{}
	for _, item := range result.Items {
		deleting[item.ID] = item
	}

	err = repo.UpdateRecords(UpdateRecords[Record[schema.Schema]]{
		Deleting: deleting,
	})

	result, err = repo.FindingRecords(FindingRecords[Record[schema.Schema]]{})
	assert.NoError(t, err)
	for _, item := range result.Items {
		t.Log(item)
	}
	assert.Len(t, result.Items, 0, "should have 0 records")
}
