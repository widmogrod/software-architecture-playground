package schemaless

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

	repo := NewDynamoDBRepository2(d, "test-repo-record")

	// clean database
	err = repo.UpdateRecords(UpdateRecords[Record[schema.Schema]]{
		Deleting: exampleUpdateRecords.Saving,
	})
	assert.NoError(t, err, "while deleting records")

	err = repo.UpdateRecords(exampleUpdateRecords)
	assert.NoError(t, err, "while saving records")

	result, err := repo.FindingRecords(FindingRecords[Record[schema.Schema]]{
		RecordType: "exampleRecord",
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
	assert.NoError(t, err, "while finding records")

	foundRecords := []Record[schema.Schema]{}
	for {
		for _, item := range result.Items {
			foundRecords = append(foundRecords, item)
		}

		if result.HasNext() {
			result, err = repo.FindingRecords(*result.Next)
		} else {
			break
		}
	}

	if assert.Len(t, foundRecords, 3, "dynamo should scan all records") {
		// DynamoDB don't support sorting on attributes, that are not part of sort key
		//assert.Equal(t, "Alice", schema.As[string](schema.Get(result.Items[0].Data, "Name"), "no-name"))
		//assert.Equal(t, "Jane", schema.As[string](schema.Get(result.Items[1].Data, "Name"), "no-name"))

		//should be able to find by id
		for _, item := range result.Items {
			found, err := repo.Get(item.ID, item.Type)
			if assert.NoError(t, err, "while getting record by id") {
				assert.Equal(t, item.ID, found.ID, "should be able to find by id")
			}
		}
	}
}
