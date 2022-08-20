package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ddb "github.com/aws/aws-sdk-go/service/dynamodb"
	kcl "github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestReadStream(t *testing.T) {
	os.Setenv("AWS_PROFILE", "gh-dev")
	fmt.Println(os.Getenv("AWS_PROFILE"))
	mySession := session.Must(session.NewSession(
		aws.NewConfig().WithRegion("eu-west-1"),
	))
	kclClient := kcl.New(mySession)

	streamName := aws.String("DevDatabaseStack-Stream790BDEE4-3WFV2JX877ao")
	stream, err := kclClient.DescribeStream(&kcl.DescribeStreamInput{
		StreamName: streamName,
		Limit:      aws.Int64(100),
	})
	assert.NoError(t, err)

	for _, shard := range stream.StreamDescription.Shards {
		fmt.Println("Shard: ", *shard.ShardId)
		iterator, err := kclClient.GetShardIterator(&kcl.GetShardIteratorInput{
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String("TRIM_HORIZON"),
			StreamName:        streamName,
			//StartingSequenceNumber: shard.SequenceNumberRange.StartingSequenceNumber,
		})
		assert.NoError(t, err)

		records, err := kclClient.GetRecords(&kcl.GetRecordsInput{
			Limit:         aws.Int64(5),
			ShardIterator: iterator.ShardIterator,
		})
		assert.NoError(t, err)

		for _, d := range records.Records {
			m := make(map[string]interface{})
			fmt.Println(string(d.Data))

			_ = `{
  "awsRegion": "eu-west-1",
  "eventID": "7df367f7-826b-438c-9fd6-712c32f45a5e",
  "eventName": "INSERT",
  "userIdentity": null,
  "recordFormat": "application/json",
  "tableName": "DevDatabaseStack-TableCD117FA1-W7FDIJ71IKFN",
  "dynamodb": {
    "ApproximateCreationDateTime": 1660598951097,
    "Keys": {
      "entity": {
        "S": "question"
      },
      "id": {
        "S": "question#1"
      }
    },
    "NewImage": {
      "content": {
        "S": "Napoleon?"
      },
      "entity": {
        "S": "question"
      },
      "id": {
        "S": "question#1"
      }
    },
    "SizeBytes": 68
  },
  "eventSource": "aws:dynamodb"
}
`

			_ = `{
  "awsRegion": "eu-west-1",
  "eventID": "fd3993fb-b46d-43f6-a817-e2270e727ef9",
  "eventName": "MODIFY",
  "userIdentity": null,
  "recordFormat": "application/json",
  "tableName": "DevDatabaseStack-TableCD117FA1-W7FDIJ71IKFN",
  "dynamodb": {
    "ApproximateCreationDateTime": 1660599091658,
    "Keys": {
      "entity": {
        "S": "question"
      },
      "id": {
        "S": "question#1"
      }
    },
    "NewImage": {
      "content": {
        "S": "Napoleon? 2"
      },
      "entity": {
        "S": "question"
      },
      "id": {
        "S": "question#1"
      }
    },
    "OldImage": {
      "content": {
        "S": "Napoleon?"
      },
      "entity": {
        "S": "question"
      },
      "id": {
        "S": "question#1"
      }
    },
    "SizeBytes": 112
  },
  "eventSource": "aws:dynamodb"
}
`

			err := json.Unmarshal(d.Data, &m)
			if err != nil {
				log.Println(err)
				continue
			}
			//log.Printf("GetRecords Data: %v\n", m)
		}

		//fmt.Println(records)
	}
}

func TestWriteToDynamoDb(t *testing.T) {
	os.Setenv("AWS_PROFILE", "gh-dev")
	fmt.Println(os.Getenv("AWS_PROFILE"))
	mySession := session.Must(session.NewSession(
		aws.NewConfig().WithRegion("eu-west-1"),
	))

	dynamoC := ddb.New(mySession)
	result, err := dynamoC.PutItem(&ddb.PutItemInput{
		Item: map[string]*ddb.AttributeValue{
			"entity": {
				S: aws.String("question"),
			},
			"id": {
				S: aws.String("question#123"),
			},
			"content": {
				S: aws.String("Napoleon? frome code 3"),
			},
			"msp": {
				BOOL: aws.Bool(true),
			},
		},
		TableName:              aws.String("DevDatabaseStack-TableCD117FA1-W7FDIJ71IKFN"),
		ReturnConsumedCapacity: aws.String("TOTAL"),
		ReturnValues:           aws.String("ALL_OLD"),
	})
	assert.NoError(t, err)
	fmt.Println(result.Attributes)
	fmt.Println(result.ConsumedCapacity)
	fmt.Println(result.GoString())

	result2, err := dynamoC.UpdateItem(&ddb.UpdateItemInput{
		Key: map[string]*ddb.AttributeValue{
			"entity": {
				S: aws.String("question"),
			},
			"id": {
				S: aws.String("question#333"),
			},
		},
		AttributeUpdates: map[string]*ddb.AttributeValueUpdate{
			"content": {
				Action: aws.String("PUT"),
				Value: &ddb.AttributeValue{
					S: aws.String("question update from code"),
				},
			},
		},
		TableName:              aws.String("DevDatabaseStack-TableCD117FA1-W7FDIJ71IKFN"),
		ReturnConsumedCapacity: aws.String("TOTAL"),
		ReturnValues:           aws.String("ALL_OLD"),
	})
	assert.NoError(t, err)
	fmt.Println(result2.Attributes)
	fmt.Println(result2.ConsumedCapacity)
	fmt.Println(result2.GoString())
}

func TestKVStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*101)
	store := kv.Default()

	go func() {
		err := store.Sync(ctx, func(key kv.Key, m map[string]kv.AttrType) {
			r, _ := json.Marshal(m)
			fmt.Println(key, string(r))
		})
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second * 20)

	go func() {
		for i := 0; i < 100; i++ {
			err := store.SetAttributes(kv.Key{
				PartitionKey: "question#e" + strconv.Itoa(i) + "#" + time.Now().Format("20060102150405"),
				EntityKey:    "question",
			}, map[string]kv.AttrType{
				"content": {S: PtrString("ohoho")},
				"created": {S: PtrString(time.Now().String())},
			})
			assert.NoError(t, err)
		}
	}()

	time.Sleep(time.Second * 10)
	//assert.Eventually(t, func() bool {
	//	return true
	//}, time.Second*5, time.Second)

	cancel()
	<-ctx.Done()
}

func PtrString(s string) *string {
	return &s
}
