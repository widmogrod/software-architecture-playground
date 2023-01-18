package storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type ConnectionToSession struct {
	ConnectionID string
	SessionID    string
}

func TestDynamoDBRepository(t *testing.T) {
	//schema.RegisterTransformations([]schema.TransformFunc{
	//	schema.WrapStruct(ConnectionToSession{}, "ConnectionToSession"),
	//})
	//schema.RegisterRules([]schema.RuleMatcher{
	//	schema.UnwrapStruct(ConnectionToSession{}, "ConnectionToSession"),
	//})

	os.Setenv("AWS_PROFILE", "gh-dev")
	os.Setenv("AWS_DEFAULT_REGION", "eu-west-1")

	cfg, err := config.LoadDefaultConfig(context.Background())
	assert.NoError(t, err)

	d := dynamodb.NewFromConfig(cfg)

	repo := DynamoDBRepository[ConnectionToSession]{
		tableName: "test-repo",
		client:    d,
	}

	err = repo.Set("test", ConnectionToSession{
		ConnectionID: "test",
		SessionID:    "session-test",
	})
	assert.NoError(t, err)

	item, err := repo.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", item.ConnectionID)
	assert.Equal(t, "session-test", item.SessionID)

	result, err := repo.FindAllKeyEqual("SessionID", "session-test")
	assert.NoError(t, err)
	assert.False(t, result.HasNext())
	assert.Equal(t, 1, len(result.Items))
	assert.Equal(t, result.Items[0].ConnectionID, "test")
	assert.Equal(t, result.Items[0].SessionID, "session-test")

	err = repo.Delete("test")
	assert.NoError(t, err)
	_, err = repo.Get("test")
	assert.ErrorIs(t, err, ErrNotFound)
}
