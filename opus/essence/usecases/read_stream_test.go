package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestPopulateOpenSearch(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	store := kv.Default()

	err := kv.SyncStrategySequence(ctx, store, func(key kv.Key, m map[string]kv.AttrType) {
		r, _ := json.Marshal(m)
		fmt.Println(key, string(r))
		err := store.IndexDocument(context.Background(), key, m)
		assert.NoError(t, err)
	})
	assert.NoError(t, err)

	<-ctx.Done()
}

func TestETLData(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	store := kv.Default()

	initialRecordsCount := store.Count()
	insertRecords := 0

	lock := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range GenerateData(5) {
			err := store.SetAttributes(data.Key, data.Attr)
			assert.NoError(t, err)
			insertRecords++
		}
	}()

	uniqueIds := make(map[string]bool)

	etlCount := 0
	etlDupCout := 0
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := store.EtlDynamoAndSync(ctx, func(key kv.Key, m map[string]kv.AttrType) {
			lock.Lock()
			if _, ok := uniqueIds[key.String()]; !ok {
				uniqueIds[key.String()] = true
				etlCount++
			} else {
				etlDupCout++
			}
			lock.Unlock()

		})
		assert.NoError(t, err)
	}()

	fromStreamCount := 0
	fromStreamDupCount := 0

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := store.Sync(ctx, func(key kv.Key, m map[string]kv.AttrType) {
			lock.Lock()
			if _, ok := uniqueIds[key.String()]; !ok {
				uniqueIds[key.String()] = true
				fromStreamCount++
			} else {
				fromStreamDupCount++
			}
			lock.Unlock()
		})
		assert.NoError(t, err)
	}()

	wg.Wait()

	fmt.Println("initialRecordsCount", initialRecordsCount)
	fmt.Println("insertRecords", insertRecords)
	fmt.Println("fromStreamCount", fromStreamCount)
	fmt.Println("fromStreamDupCount", fromStreamDupCount)
	fmt.Println("etlCount", etlCount)
	fmt.Println("etlDupCout", etlDupCout)

	endCount := store.Count()

	//assert.Eventually(t, func() bool {
	//	return true
	//}, time.Second*5, time.Second)
	assert.Equal(t, endCount, int64(len(uniqueIds)))
}

type Generic struct {
	Key  kv.Key                 `json:"key"`
	Attr map[string]kv.AttrType `json:"attr"`
}

func GenerateData(max int) chan *Generic {
	ch := make(chan *Generic)
	go func() {
		defer close(ch)
		for i := 0; i < max; i++ {
			entity := gofakeit.Verb()
			key := kv.Key{
				PartitionKey: entity + "#" + strconv.Itoa(i) + "#" + time.Now().Format("20060102150405"),
				EntityKey:    entity,
			}
			attr := map[string]kv.AttrType{
				"content": {S: kv.PtrString(gofakeit.HipsterSentence(30))},
				"created": {DT: kv.PtrTime(time.Now())},
				"version": {I: kv.PtrInt64(time.Now().UnixNano())},
			}

			for i := 0; i < rand.Intn(10); i++ {
				name := gofakeit.Noun()
				if rand.Float64() < 0.5 {
					attr[name] = kv.AttrType{S: kv.PtrString(gofakeit.HipsterSentence(30))}
				} else {
					attr[name] = kv.AttrType{I: kv.PtrInt64(rand.Int63())}
				}
			}

			ch <- &Generic{
				Key:  key,
				Attr: attr,
			}
		}
	}()

	return ch
}

//func TestAddRelations(t *testing.T) {
//	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
//	store := kv.Default()
//
//	store.SetAttributes(kv.Key{PartitionKey: "1", EntityKey: "1"}, map[string]kv.AttrType{})
//}
