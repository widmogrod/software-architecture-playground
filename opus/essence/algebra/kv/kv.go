package kv

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ddb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io/ioutil"
	"log"
	"net/http"
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

func (k Key) String() string {
	return fmt.Sprintf("pk=%s#e=%s", k.PartitionKey, k.EntityKey)
}

////go:generate mkunion -name=AttrType2 -types=S,I,B,DT
//type (
//	S  string
//	I  int
//	B  bool
//	DT time.Time
//)

type AttrType struct {
	S  *string    `json:"S,omitempty"`
	I  *int64     `json:"I,omitempty"`
	B  *bool      `json:"B,omitempty"`
	DT *time.Time `json:"DT,omitempty"`
}

func Default() *Store {
	os.Setenv("AWS_PROFILE", "gh-dev")
	s := session.Must(session.NewSession(
		aws.NewConfig().WithRegion("eu-west-1"),
	))
	d := ddb.New(s)
	tableName := "opus-table-dev"

	k := kinesis.New(s)
	streamName := "opus-stream-dev"

	oss := opensearchservice.New(s)
	domain, err := oss.DescribeDomain(&opensearchservice.DescribeDomainInput{
		DomainName: aws.String("opus-domain-dev"),
	})

	if err != nil {
		log.Fatal(err)
	}

	o, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			"https://" + *domain.DomainStatus.Endpoint,
		},
		//Addresses: []string{
		//	"http://localhost:9200",
		//},
		Username: "admin",
		//Password: "admin",
		Password: "nile!DISLODGE5clause",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	indexName := "opus-index-dev"
	if err != nil {
		panic(err)
	}

	return NewStore(d, k, o, tableName, streamName, indexName)
}

func NewStore(
	d *ddb.DynamoDB, k *kinesis.Kinesis, o *opensearch.Client,
	tableName, streamName, indexName string,
) *Store {
	return &Store{
		dynamo:             d,
		tableName:          &tableName,
		kinesis:            k,
		streamName:         &streamName,
		opensearch:         o,
		indexName:          &indexName,
		dynamoPartitionKey: "id",
		dynamoEntityKey:    "entity",
	}
}

type Store struct {
	dynamo             *ddb.DynamoDB
	tableName          *string
	kinesis            *kinesis.Kinesis
	streamName         *string
	opensearch         *opensearch.Client
	indexName          *string
	dynamoPartitionKey string
	dynamoEntityKey    string
}

func (s *Store) SetAttributes(key Key, attributes map[string]AttrType) error {
	item, err2 := s.toDynamoAttributeValue(key, attributes)
	if err2 != nil {
		return err2
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

func (s *Store) toDynamoAttributeValue(key Key, attributes map[string]AttrType) (map[string]*ddb.AttributeValue, error) {
	item := map[string]*ddb.AttributeValue{}

	for k, v := range attributes {
		if v.S != nil {
			item[k] = &ddb.AttributeValue{
				S: v.S,
			}
		} else if v.I != nil {
			item[k] = &ddb.AttributeValue{
				N: aws.String(strconv.FormatInt(*v.I, 10)),
			}
		} else if v.B != nil {
			item[k] = &ddb.AttributeValue{
				BOOL: v.B,
			}
		} else if v.DT != nil {
			dv, err := dynamodbattribute.Marshal(v)
			if err != nil {
				return nil, err
			}
			item[k] = dv
		} else {
			return nil, errors.New("unsupported type")
		}
	}

	item[s.dynamoEntityKey] = &ddb.AttributeValue{
		S: aws.String(key.EntityKey),
	}
	item[s.dynamoPartitionKey] = &ddb.AttributeValue{
		S: aws.String(key.PartitionKey),
	}
	return item, nil
}

func (s *Store) GetAttributes(key Key) (map[string]AttrType, error) {
	return nil, nil
}

func (s *Store) EtlDynamoAndSync(ctx context.Context, sink func(Key, map[string]AttrType)) error {
	// measure start time of etl to use it as time of sync from kinestis
	// measure end time of etl, and if is took longer than kinesis window
	// then this becomes a problem, so maybe kinesis stream and dynamo processes should be interleaved
	// using some heuristic like
	// - when average time of sync shows that it will take longer than kinesis window,
	// - then start kinesis stream and dynamo processes in parallel
	// - when dynamo processs is done, close it, and continue only with kinesis

	s.Count()

	//s.dynamo.QueryPages(&ddb.QueryInput{})
	err := s.dynamo.ScanPagesWithContext(ctx, &ddb.ScanInput{
		TableName: s.tableName,
		//ConsistentRead: aws.Bool(true),
		// for parallelism
		//TotalSegments: 3,
		//Segment: 3,
	}, func(output *ddb.ScanOutput, b bool) bool {
		fmt.Println(output.LastEvaluatedKey)
		for _, item := range output.Items {
			data := map[string]AttrType{}
			for k, v := range item {
				val, err := s.dynamoToKv(v)
				if err != nil {
					panic(err)
				}
				data[k] = *val
			}

			key := Key{
				PartitionKey: *data[s.dynamoPartitionKey].S,
				EntityKey:    *data[s.dynamoEntityKey].S,
			}

			delete(data, s.dynamoPartitionKey)
			delete(data, s.dynamoEntityKey)

			sink(key, data)
		}

		return true
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Count() int64 {
	res, err2 := s.dynamo.Scan(&ddb.ScanInput{
		TableName:              s.tableName,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		//ConsistentRead:         aws.Bool(true),
		Select: aws.String("COUNT"),
	})
	fmt.Println("count", res, err2)
	return *res.Count
}

func (s *Store) Sync(ctx context.Context, sink func(Key, map[string]AttrType)) error {
	stream, err := s.kinesis.DescribeStreamWithContext(ctx, &kinesis.DescribeStreamInput{
		StreamName: s.streamName,
		Limit:      aws.Int64(10000),
	})
	if err != nil {
		return err
	}

	// TODO stream.StreamDescription.HasMoreShards

	//stream.StreamDescription.RetentionPeriodHours

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
			iterator, err := s.kinesis.GetShardIteratorWithContext(ctx, it)
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

		records, err := s.kinesis.GetRecordsWithContext(ctx, &kinesis.GetRecordsInput{
			Limit:         aws.Int64(100),
			ShardIterator: nextIterator,
		})

		if err != nil {
			return fmt.Errorf("GetRecords: %w", err)
		}

		err2 := s.processRecords(records, sink)
		if err2 != nil {
			return fmt.Errorf("processRecords: %s", err)
		}
		nextIterator = records.NextShardIterator
	}
}

func (s *Store) processRecords(records *kinesis.GetRecordsOutput, sink func(Key, map[string]AttrType)) error {
	for _, d := range records.Records {
		m := make(map[string]interface{})
		err := json.Unmarshal(d.Data, &m)
		if err != nil {
			log.Println("processRecords: err=", err)
			continue
		}

		dat := m["dynamodb"].(MapAny)["NewImage"].(MapAny)
		res, err2 := s.deser(dat)
		if err2 != nil {
			return err2
		}

		delete(res, s.dynamoPartitionKey)
		delete(res, s.dynamoEntityKey)

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
		if v.(MapAny)["S"] != nil {
			res[k] = AttrType{
				S: aws.String(v.(MapAny)["S"].(string)),
			}
		} else if v.(MapAny)["N"] != nil {
			val, err := strconv.ParseInt(v.(MapAny)["N"].(string), 10, 64)
			if err != nil {
				log.Println("deser: err", err)
				continue
			}
			res[k] = AttrType{
				I: aws.Int64(val),
			}
		} else if v.(MapAny)["BOOL"] != nil {
			res[k] = AttrType{
				B: aws.Bool(v.(MapAny)["BOOL"].(bool)),
			}
		} else if v.(MapAny)["M"].(MapAny)["DT"] != nil {
			val := v.(MapAny)["M"].(MapAny)["DT"].(MapAny)["S"].(string)
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				log.Println("deser: time err", err)
				continue
			}
			res[k] = AttrType{
				DT: aws.Time(t),
			}
		} else {
			vv, _ := json.Marshal(v)
			return nil, errors.New("sync: unsupported type; " + string(vv))
		}
	}
	return res, nil
}

func (s *Store) dynamoToKv(in *ddb.AttributeValue) (*AttrType, error) {
	if in.S != nil {
		return &AttrType{
			S: in.S,
		}, nil
	} else if in.N != nil {
		val, err := strconv.ParseInt(*in.N, 10, 64)
		if err != nil {
			return nil, err
		}
		return &AttrType{
			I: aws.Int64(val),
		}, nil
	} else if in.BOOL != nil {
		return &AttrType{
			B: in.BOOL,
		}, nil
	} else if in.M != nil && in.M["DT"].S != nil {
		t, err := time.Parse(time.RFC3339, *in.M["DT"].S)
		if err != nil {
			return nil, err
		}
		return &AttrType{
			DT: aws.Time(t),
		}, nil
	} else {
		vv, _ := json.Marshal(in)
		return nil, errors.New("sync: unsupported type; " + string(vv))
	}
}

func (s *Store) IndexDocument(ctx context.Context, key Key, attrs map[string]AttrType) error {
	doc := ToOpenSearchIndex(key, attrs)
	body, err := json.Marshal(&doc)
	if err != nil {
		return err
	}

	req := opensearchapi.IndexRequest{
		Index:        *s.indexName,
		DocumentID:   doc["docId"].(string),
		DocumentType: "_doc",
		Body:         bytes.NewReader(body),
	}

	res, err := req.Do(ctx, s.opensearch)
	if err != nil {
		return err
	}

	rbody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	fmt.Println("body", string(rbody))
	if err != nil {
		return err
	}
	return nil
}

/*
	{
		"docId": "textbook_solution#13123/question#666"
	    "entityId": "question#us123",
		"entityType": "question#666"
		"attributes": [
			{"name": "edgeName", "value": { "string": "is_canonical_to}}
			{"name": "created",  "value": { "date": 2020-08-15 10:10:10}}
			{"name": "update",   "value": { "date": 2020-08-15 10:10:10}
		]
	}
*/
func ToOpenSearchIndex(key Key, attrs map[string]AttrType) map[string]interface{} {
	fmt.Println(key, attrs)
	result := make(map[string]interface{})

	params := []map[string]interface{}{}
	for k, v := range attrs {
		params = append(params, map[string]interface{}{
			"name":  k,
			"value": v,
		})
	}

	params = append(params, map[string]interface{}{
		"name":  "entityId",
		"value": AttrType{S: PtrString(key.PartitionKey)},
	})
	params = append(params, map[string]interface{}{
		"name":  "entityType",
		"value": AttrType{S: PtrString(key.EntityKey)},
	})

	// sha1(docId)
	docId := key.PartitionKey + "#" + key.EntityKey
	docIdHash := sha1.Sum([]byte(docId))
	docIdHashStr := hex.EncodeToString(docIdHash[:])

	result["docId"] = docIdHashStr
	result["entityId"] = key.PartitionKey
	result["entityType"] = key.EntityKey
	result["attributes"] = params
	return result
}

func (s *Store) Find() error {
	response, err := s.opensearch.Search(func(request *opensearchapi.SearchRequest) {
		request.Index = []string{*s.indexName}
		request.Query = "*:*"
		request.Pretty = true
	})
	if err != nil {
		return err
	}

	fmt.Println(response.StatusCode)
	return nil
}

func SyncStrategySequence(ctx context.Context, store *Store, sink func(key Key, attrs map[string]AttrType)) error {
	err := store.EtlDynamoAndSync(ctx, sink)
	if err != nil {
		return err
	}
	err = store.Sync(ctx, sink)
	if err != nil {
		return err
	}
	return nil

}

func PtrString(s string) *string {
	return &s
}

func PtrInt64(nano int64) *int64 {
	return &nano
}

func PtrTime(now time.Time) *time.Time {
	return &now
}

func PtrBool(b bool) *bool {
	return &b
}

func MkString(s string) AttrType {
	return AttrType{
		S: &s,
	}
}

func MkInt64(i int64) AttrType {
	return AttrType{
		I: &i,
	}
}

func MkTime(t time.Time) AttrType {
	return AttrType{
		DT: &t,
	}
}

func MkBool(b bool) AttrType {
	return AttrType{
		B: &b,
	}
}
