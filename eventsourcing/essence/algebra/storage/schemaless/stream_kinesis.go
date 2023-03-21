package schemaless

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	types2 "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
	"time"
)

func NewKinesisStream(k *kinesis.Client, streamName string) *KinesisStream {
	ctx := context.Background()
	stream, err := k.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
		Limit:      aws.Int32(1000),
	})
	if err != nil {
		panic(err)
	}

	return &KinesisStream{
		kinesis:    k,
		stream:     stream,
		streamName: streamName,
	}
}

type KinesisStream struct {
	kinesis    *kinesis.Client
	stream     *kinesis.DescribeStreamOutput
	streamName string

	lock        sync.RWMutex
	subscribers []func(Change[schema.Schema])
	done        []chan struct{}
	once        sync.Once
}

func (s *KinesisStream) Pull() chan Change[schema.Schema] {
	result := make(chan Change[schema.Schema])

	ctx := context.Background()
	for _, shard := range s.stream.StreamDescription.Shards {
		var shardIterator *string = nil
		if shardIterator == nil {
			it := &kinesis.GetShardIteratorInput{
				ShardId:           shard.ShardId,
				StreamName:        aws.String(s.streamName),
				ShardIteratorType: types.ShardIteratorTypeLatest,
				//ShardIteratorType:      types.ShardIteratorTypeAtSequenceNumber,
				//StartingSequenceNumber: shard.SequenceNumberRange.StartingSequenceNumber,
			}
			iterator, err := s.kinesis.GetShardIterator(ctx, it)
			if err != nil {
				panic(err)
			}
			shardIterator = iterator.ShardIterator
		}

		go s.processShard(ctx, shardIterator, result)
	}

	return result
}

func (s *KinesisStream) processShard(ctx context.Context, shardIterator *string, resultC chan Change[schema.Schema]) {
	lastRequest := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// continue
		}

		if diff := time.Now().Sub(lastRequest); diff < time.Second/5 {
			log.Debugf("🗺Sleeping for %s", time.Second/5-diff)
			time.Sleep(time.Second/5 - diff)
		}

		lastRequest = time.Now()
		records, err := s.kinesis.GetRecords(ctx, &kinesis.GetRecordsInput{
			ShardIterator: shardIterator,
			Limit:         aws.Int32(100),
			//StreamARN:     shard.ShardId,
		})
		if err != nil {
			log.Errorln("🗺GetRecords:", err)
			panic(err)
		}

		for _, record := range records.Records {
			schemed, err := schema.FromJSON(record.Data)
			if err != nil {
				panic(err)
			}

			// potentially change, can be just state. And data pipeline can detect it
			// groub by  key
			// no initial state, not created
			// ther is state, and there is new - updated
			// there is state, with deleted flag set to true, delete
			// this implice that soft delete, or other options can happen.
			// but when is deleted, key could be closed? this would require some instruction
			// or maybe as with imposibility of distributed consensus,
			// data flush is important, windowing and triggers, etc?
			result := Change[schema.Schema]{
				Before:  nil,
				After:   nil,
				Deleted: false,
			}

			switch schema.As[string](schema.Get(schemed, "eventName"), "") {
			case "MODIFY":
				// has both NewImage and OldImage
				old := schema.Get(schemed, "dynamodb.OldImage")
				before, err := s.toTyped(old)
				if err != nil {
					panic(err)
				}
				result.Before = &before

				new := schema.Get(schemed, "dynamodb.NewImage")
				after, err := s.toTyped(new)
				if err != nil {
					panic(err)
				}
				result.After = &after

			case "INSERT":
				// has only NewImage
				new := schema.Get(schemed, "dynamodb.NewImage")
				after, err := s.toTyped(new)
				if err != nil {
					panic(err)
				}
				result.After = &after
			case "REMOVE":
				// has only OldImage
				old := schema.Get(schemed, "dynamodb.OldImage")
				before, err := s.toTyped(old)
				if err != nil {
					panic(err)
				}
				result.Before = &before
				result.Deleted = true

			default:
				panic(fmt.Errorf("unknown event name: %s", schema.As[string](schema.Get(schemed, "eventName"), "")))
			}

			resultC <- result
		}

		if records.NextShardIterator == nil {
			break
		}
		shardIterator = records.NextShardIterator
	}
}

func (s *KinesisStream) Subscribe(ctx context.Context, fromOffset int, f func(Change[schema.Schema])) error {
	// TODO review this, to something more robust
	s.once.Do(func() {
		defer func() {
			s.lock.RLock()
			for _, done := range s.done {
				done <- struct{}{}
			}
			s.lock.RUnlock()
		}()

		for result := range s.Pull() {
			s.lock.RLock()
			for _, f := range s.subscribers {
				f(result)
			}
			s.lock.RUnlock()
		}

	})

	done := make(chan struct{})

	s.lock.Lock()
	s.subscribers = append(s.subscribers, f)
	s.done = append(s.done, done)
	s.lock.Unlock()

	<-done

	return nil
}

func (s *KinesisStream) toTyped(record schema.Schema) (Record[schema.Schema], error) {
	normalised, err := UnwrapDynamoDB(record)
	if err != nil {
		data, err := schema.ToJSON(record)
		log.Errorln("🗺store.KinesisStream corrupted record:", string(data), err)
		return Record[schema.Schema]{}, fmt.Errorf("store.KinesisStream unwrap DynamoDB record: %v", record)
	}

	typed := Record[schema.Schema]{
		ID:      schema.As[string](schema.Get(normalised, "ID"), "record-id-corrupted"),
		Type:    schema.As[string](schema.Get(normalised, "Type"), "record-id-corrupted"),
		Data:    schema.Get(normalised, "Data"),
		Version: schema.As[uint16](schema.Get(normalised, "Version"), 0),
	}
	if typed.Type == "record-id-corrupted" &&
		typed.ID == "record-id-corrupted" &&
		typed.Version == 0 {
		data, err := schema.ToJSON(normalised)
		log.Errorln("🗺store.KinesisStream corrupted record:", string(data), err)
		return Record[schema.Schema]{}, fmt.Errorf("store.KinesisStream corrupted record: %v", normalised)
	}
	return typed, nil
}

func UnwrapDynamoDB(data schema.Schema) (schema.Schema, error) {
	switch x := data.(type) {
	case *schema.Map:
		if len(x.Field) == 1 {
			for _, field := range x.Field {
				switch field.Name {
				case "S":
					value := schema.As[string](field.Value, "")
					return schema.FromDynamoDB(&types2.AttributeValueMemberS{
						Value: value,
					})
				case "N":
					value := schema.As[string](field.Value, "")
					return schema.FromDynamoDB(&types2.AttributeValueMemberN{
						Value: value,
					})
				case "B":
					value := schema.As[[]byte](field.Value, nil)
					return schema.FromDynamoDB(&types2.AttributeValueMemberB{
						Value: value,
					})
				case "BOOL":
					value := schema.As[bool](field.Value, false)
					return schema.FromDynamoDB(&types2.AttributeValueMemberBOOL{
						Value: value,
					})
				case "NULL":
					return &schema.None{}, nil
				case "M":
					switch y := field.Value.(type) {
					case *schema.Map:
						return assumeMap(y)
					default:
						return nil, fmt.Errorf("unknown type (1): %T", field.Value)
					}

				case "L":
					result := &schema.List{}
					switch field.Value.(type) {
					case *schema.List:
						for _, item := range field.Value.(*schema.List).Items {
							value, err := UnwrapDynamoDB(item)
							if err != nil {
								return nil, err
							}
							result.Items = append(result.Items, value)
						}
					}
					return result, nil

				default:
					return nil, fmt.Errorf("unknown type (3): %T", field.Value)
				}
			}
		} else {
			return assumeMap(x)
		}
	}

	return nil, fmt.Errorf("unknown type (2): %T", data)
}

func assumeMap(x *schema.Map) (schema.Schema, error) {
	result := &schema.Map{}
	for _, field := range x.Field {
		value, err := UnwrapDynamoDB(field.Value)
		if err != nil {
			return nil, err
		}
		result.Field = append(result.Field, schema.Field{
			Name:  field.Name,
			Value: value,
		})
	}
	return result, nil
}
