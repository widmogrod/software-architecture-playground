package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"testing"
)

type User struct {
	ID   string
	Name string
	Age  int
}

type UsersCountByAge struct {
	Count int
}

func AgeRangeKey(age int) string {
	if age < 20 {
		return "0-20"
	} else if age < 30 {
		return "20-30"
	} else if age < 40 {
		return "30-40"
	} else {
		return "40+"
	}
}

func TestNewRepositoryInMemory(t *testing.T) {
	storage := NewRepository2WithSchema()
	aggregate := NewKeyedAggregate[User, UsersCountByAge](
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

	err := r.UpdateRecords(UpdateRecords[Record[User]]{
		Saving: map[string]Record[User]{
			"1": {
				ID: "1",
				Data: User{
					ID:   "1",
					Name: "John",
					Age:  20,
				},
			},
			"2": {
				ID: "2",
				Data: User{
					ID:   "2",
					Name: "Jane",
					Age:  30,
				},
			},
			"3": {
				ID: "3",
				Data: User{
					ID:   "3",
					Name: "Alice",
					Age:  39,
				},
			},
		},
	})
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
}
