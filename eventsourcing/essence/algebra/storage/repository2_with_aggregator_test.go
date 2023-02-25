package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"testing"
)

type User struct {
	Name string
	Age  int
}

type UsersCountByAge struct {
	Count int
}

func AgeRangeKey(age int) string {
	if age < 20 {
		return "byAge:0-20"
	} else if age < 30 {
		return "byAge:20-30"
	} else if age < 40 {
		return "byAge:30-40"
	} else {
		return "byAge:40+"
	}
}

var exampleUserRecords = Save(
	Record[User]{
		ID:   "1",
		Type: "user",
		Data: User{
			Name: "John",
			Age:  20,
		},
	},
	Record[User]{
		ID:   "2",
		Type: "user",
		Data: User{
			Name: "Jane",
			Age:  30,
		},
	},
	Record[User]{
		ID:   "3",
		Type: "user",
		Data: User{
			Name: "Alice",
			Age:  39,
		},
	},
)

func TestNewRepositoryInMemory(t *testing.T) {
	storage := NewRepository2WithSchema()
	aggregate := NewKeyedAggregate[User, UsersCountByAge](
		"byAge",
		[]string{"user"},
		func(data User) (string, UsersCountByAge) {
			return AgeRangeKey(data.Age), UsersCountByAge{
				Count: 1,
			}
		},
		func(a, b UsersCountByAge) (UsersCountByAge, error) {
			return UsersCountByAge{
				Count: a.Count + b.Count,
			}, nil
		},
		storage,
	)
	r := NewRepositoryWithIndexer[User, UsersCountByAge](
		storage,
		aggregate,
	)

	err := r.UpdateRecords(exampleUserRecords)
	assert.NoError(t, err)

	result, err := r.FindingRecords(FindingRecords[Record[User]]{
		Where: predicate.MustWhere(
			"Age > :age",
			predicate.ParamBinds{
				":age": schema.MkInt(20),
			},
		),
		Sort: []SortField{
			{
				Field:      "Name",
				Descending: true,
			},
		},
	})
	assert.NoError(t, err)

	if assert.Len(t, result.Items, 2) {
		assert.Equal(t, "Alice", result.Items[0].Data.Name)
		assert.Equal(t, "Jane", result.Items[1].Data.Name)
	}

	results, err := storage.FindingRecords(FindingRecords[Record[schema.Schema]]{
		RecordType: "byAge",
		Sort: []SortField{
			{
				Field:      "Count",
				Descending: true,
			},
		},
	})
	assert.NoError(t, err)

	if assert.Len(t, results.Items, 2) {
		r, err := RecordAs[UsersCountByAge](results.Items[0])
		assert.NoError(t, err)
		assert.Equal(t, "byAge:20-30", r.ID)
		assert.Equal(t, 1, r.Data.Count)

		r, err = RecordAs[UsersCountByAge](results.Items[1])
		assert.NoError(t, err)
		assert.Equal(t, "byAge:30-40", r.ID)
		assert.Equal(t, 2, r.Data.Count)
	}
}
