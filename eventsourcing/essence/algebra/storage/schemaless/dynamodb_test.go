package schemaless

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/awsutils"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"testing"
)

func TestNewDynamoDBRepository(t *testing.T) {
	tableName := "test-repo-record"

	cfg, err := awsutils.LoadLocalStackAwsConfig(context.Background())
	d := dynamodb.NewFromConfig(cfg)

	err = setupDynamoDB(d, tableName)
	assert.NoError(t, err, "while setting up dynamodb")

	repo := NewDynamoDBRepository(d, tableName)
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
				Descending: false,
			},
		},
		Limit: 2,
	})
	assert.NoError(t, err, "while finding records")

	var foundRecords []Record[schema.Schema]
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
		//assert.Equal(t, "Alice", schema.As[string](schema.Get(result.Items[0].Data, "Ctx"), "no-name"))
		//assert.Equal(t, "Jane", schema.As[string](schema.Get(result.Items[1].Data, "Ctx"), "no-name"))

		//should be able to find by id
		for _, item := range result.Items {
			found, err := repo.Get(item.ID, item.Type)
			if assert.NoError(t, err, "while getting record by id") {
				assert.Equal(t, item.ID, found.ID, "should be able to find by id")
			}
		}
	}
}

func setupDynamoDB(d *dynamodb.Client, tableName string) error {
	// clean database, if exists
	_, _ = d.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: &tableName,
	})

	_, err := d.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Type"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("Type"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		TableName:   &tableName,
	})

	return err
}
