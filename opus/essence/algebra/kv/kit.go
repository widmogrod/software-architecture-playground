package kv

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Record struct {
	Key        Key
	Attributes map[string]AttrType
}

type SaveRecords struct {
	Saving   []Record
	Deleting []Record
}

func (s *Store) SaveRecords(cmd *SaveRecords) error {
	var transact []*dynamodb.TransactWriteItem

	for i := range cmd.Saving {
		item, err := s.toDynamoAttributeValue(cmd.Saving[i].Key, cmd.Saving[i].Attributes)
		if err != nil {
			return err
		}

		transact = append(transact, &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				TableName: s.tableName,
				Item:      item,
			},
		})
	}

	// delete transact items in dynamodb
	for i := range cmd.Deleting {
		item, err := s.toDynamoAttributeValue(cmd.Deleting[i].Key, nil)
		if err != nil {
			return err
		}
		transact = append(transact, &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: s.tableName,
				Key:       item,
			},
		})
	}

	if len(transact) == 0 {
		return errors.New("no records to save")
	}

	_, err := s.dynamo.TransactWriteItemsWithContext(context.Background(), &dynamodb.TransactWriteItemsInput{
		TransactItems: transact,
	})

	if err != nil {
		return err
	}

	return nil
}

type RetrieveRecords struct {
	Keys []Key
}

type RetrievedRecordsResult struct {
	Records map[Key]*Record
}

func (s *Store) RetrieveRecords(cmd RetrieveRecords) error {
	var items map[string]*dynamodb.KeysAndAttributes
	for i := range cmd.Keys {
		item, err := s.toDynamoAttributeValue(cmd.Keys[i], nil)
		if err != nil {
			return err
		}

		key := cmd.Keys[i].String()
		items[key] = &dynamodb.KeysAndAttributes{
			Keys: []map[string]*dynamodb.AttributeValue{item},
		}
	}
	s.dynamo.BatchGetItem(&dynamodb.BatchGetItemInput{
		RequestItems:           items,
		ReturnConsumedCapacity: nil,
	})

	return nil
}
