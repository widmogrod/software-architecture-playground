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
			"1234": {
				ID: "1234",
				Data: schema.FromGo(exampleRecord{
					Name: "Bob",
					Age:  40,
				}),
			},
			"3123": {
				ID: "3123",
				Data: schema.FromGo(exampleRecord{
					Name: "Zarlie",
					Age:  39,
				}),
			},
		},
	})
	assert.NoError(t, err)

	result, err := repo.FindingRecords(FindingRecords[Record[schema.Schema]]{
		Where: predicate.MustWhere(
			"Age > :age AND Age < :maxAge",
			predicate.ParamBinds{
				":age":    schema.MkInt(20),
				":maxAge": schema.MkInt(40),
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

	if assert.Len(t, result.Items, 2, "first page should have 2 items") {
		assert.Equal(t, "Alice", schema.As[string](schema.Get(result.Items[0].Data, "Name"), "no-name"))
		assert.Equal(t, "Jane", schema.As[string](schema.Get(result.Items[1].Data, "Name"), "no-name"))
	}

	if assert.True(t, result.HasNext(), "should have next page of results") {
		nextResult, err := repo.FindingRecords(*result.Next)

		assert.NoError(t, err)
		if assert.Len(t, nextResult.Items, 1, "second page should have 1 item") {
			assert.Equal(t, "Zarlie", schema.As[string](schema.Get(nextResult.Items[0].Data, "Name"), "no-name"))

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
