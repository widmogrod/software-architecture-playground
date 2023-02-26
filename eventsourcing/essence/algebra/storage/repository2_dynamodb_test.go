package storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"testing"
)

func TestNewDynamoDBRepository2(t *testing.T) {
	//TODO inject name of the table!
	cfg, err := config.LoadDefaultConfig(context.Background())
	assert.NoError(t, err)

	d := dynamodb.NewFromConfig(cfg)

	repo := NewDynamoDBRepository2(d, "test-repo")

	// clean database
	err = repo.UpdateRecords(UpdateRecords[Record[schema.Schema]]{
		Deleting: exampleUpdateRecords.Saving,
	})
	assert.NoError(t, err, "while deleting records")

	err = repo.UpdateRecords(exampleUpdateRecords)
	assert.NoError(t, err, "while saving records")

	result, err := repo.FindingRecords(FindingRecords[Record[schema.Schema]]{
		Where: predicate.MustWhere(
			"Data.Age > :age AND Data.Age <= :maxAge",
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
	assert.NoError(t, err, "while finding records")

	if assert.Len(t, result.Items, 2, "first page should have 2 items") {
		// DynamoDB don't support sorting on attributes, that are not part of sort key
		//assert.Equal(t, "Alice", schema.As[string](schema.Get(result.Items[0].Data, "Name"), "no-name"))
		//assert.Equal(t, "Jane", schema.As[string](schema.Get(result.Items[1].Data, "Name"), "no-name"))
	}
}
