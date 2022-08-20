package kv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ddb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type MapAny = map[string]interface{}

type Key struct {
	PartitionKey string
	EntityKey    string
}

type AttrType struct {
	S *string
	I *int
	B *bool
}

func Default() *Store {
	os.Setenv("AWS_PROFILE", "gh-dev")
	s := session.Must(session.NewSession(
		aws.NewConfig().WithRegion("eu-west-1"),
	))
	d := ddb.New(s)
	tableName := "DevDatabaseStack-TableCD117FA1-W7FDIJ71IKFN"

	k := kinesis.New(s)
	streamName := "DevDatabaseStack-Stream790BDEE4-3WFV2JX877ao"

	return NewStore(d, k, tableName, streamName)
}

func NewStore(d *ddb.DynamoDB, k *kinesis.Kinesis, tableName, streamName string) *Store {
	return &Store{
		dynamo:     d,
		kinesis:    k,
		tableName:  &tableName,
		streamName: &streamName,
	}
}

type Store struct {
	dynamo     *ddb.DynamoDB
	tableName  *string
	kinesis    *kinesis.Kinesis
	streamName *string
}

func (s *Store) SetAttributes(key Key, attributes map[string]AttrType) error {
	item := map[string]*ddb.AttributeValue{}

	for k, v := range attributes {
		if v.S != nil {
			item[k] = &ddb.AttributeValue{
				S: v.S,
			}
		} else if v.I != nil {
			item[k] = &ddb.AttributeValue{
				N: aws.String(strconv.Itoa(*v.I)),
			}
		} else if v.B != nil {
			item[k] = &ddb.AttributeValue{
				BOOL: v.B,
			}
		} else {
			return errors.New("unsupported type")
		}
	}

	item["entity"] = &ddb.AttributeValue{
		S: aws.String(key.EntityKey),
	}
	item["id"] = &ddb.AttributeValue{
		S: aws.String(key.PartitionKey),
	}

	_, err := s.dynamo.PutItem(&ddb.PutItemInput{
		Item:                   item,
		TableName:              s.tableName,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		ReturnValues:           aws.String("ALL_OLD"),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAttributes(key Key) (map[string]AttrType, error) {
	return nil, nil
}

func (s *Store) Sync(ctx context.Context, sink func(Key, map[string]AttrType)) error {
	stream, err := s.kinesis.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: s.streamName,
		Limit:      aws.Int64(10000),
	})
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for _, shard := range stream.StreamDescription.Shards {
		wg.Add(1)
		it := &kinesis.GetShardIteratorInput{
			ShardId:    shard.ShardId,
			StreamName: s.streamName,
		}
		fmt.Println("starting from beginning", *shard.ShardId)
		//it.ShardIteratorType = aws.String("TRIM_HORIZON")
		//it.ShardIteratorType = aws.String("LATEST")
		it.ShardIteratorType = aws.String("AT_SEQUENCE_NUMBER")
		it.StartingSequenceNumber = shard.SequenceNumberRange.StartingSequenceNumber

		go func(shard *kinesis.Shard) {
			defer wg.Done()
			iterator, err := s.kinesis.GetShardIterator(it)
			if err != nil {
				log.Println("GetShardIterator:", err)
				return
			}

			nextIterator := iterator.ShardIterator
			err = s.consumeShard(ctx, sink, nextIterator)
			if err != nil {
				log.Println("consumeShard:", err)
				return
			}
		}(shard)
	}

	wg.Wait()

	return nil
}

func (s *Store) consumeShard(ctx context.Context, sink func(Key, map[string]AttrType), nextIterator *string) error {
	tickC := time.Tick(time.Millisecond * 100)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tickC:
			// continue
		}

		records, err := s.kinesis.GetRecords(&kinesis.GetRecordsInput{
			Limit:         aws.Int64(100),
			ShardIterator: nextIterator,
		})

		if err != nil {
			return fmt.Errorf("GetRecords: %w", err)
		}

		err2 := s.processRecords(records, sink)
		if err2 != nil {
			return fmt.Errorf("processRecords: %w", err)
		}
		nextIterator = records.NextShardIterator
	}

	return nil
}

func (s *Store) processRecords(records *kinesis.GetRecordsOutput, sink func(Key, map[string]AttrType)) error {
	for _, d := range records.Records {
		//state[*shard.ShardId] = *d.SequenceNumber

		m := make(map[string]interface{})
		err := json.Unmarshal(d.Data, &m)
		if err != nil {
			log.Println(err)
			continue
		}

		dat := m["dynamodb"].(MapAny)["NewImage"].(MapAny)
		res, err2 := s.deser(dat)
		if err2 != nil {
			return err2
		}

		sink(Key{
			PartitionKey: m["dynamodb"].(MapAny)["Keys"].(MapAny)["id"].(MapAny)["S"].(string),
			EntityKey:    m["dynamodb"].(MapAny)["Keys"].(MapAny)["entity"].(MapAny)["S"].(string),
		}, res)
	}

	//fmt.Println("records.NextShardIterator", *records.NextShardIterator)
	//fmt.Println("iterator.ShardIterator", *iterator.ShardIterator)
	//if *records.NextShardIterator == *iterator.ShardIterator {
	//	fmt.Println("no more records, shard equal", *shard.ShardId)
	//}
	return nil
}

func (s *Store) deser(dat MapAny) (map[string]AttrType, error) {
	res := map[string]AttrType{}
	for k, v := range dat {
		//if k != "id" && k != "entity" {
		if v.(MapAny)["S"] != nil {
			res[k] = AttrType{
				S: aws.String(v.(MapAny)["S"].(string)),
			}
		} else if v.(MapAny)["N"] != nil {
			val, err := strconv.Atoi(v.(MapAny)["BOOL"].(string))
			if err != nil {
				log.Println("deser: err", err)
				continue
			}
			res[k] = AttrType{
				I: aws.Int(val),
			}
		} else if v.(MapAny)["BOOL"] != nil {
			res[k] = AttrType{
				B: aws.Bool(v.(MapAny)["BOOL"].(bool)),
			}
		} else {
			vv, _ := json.Marshal(v)
			return nil, errors.New("sync: unsupported type; " + string(vv))
		}
		//}
	}
	return res, nil
}

// 49632387027563710669111247013022276959494921650972917762
