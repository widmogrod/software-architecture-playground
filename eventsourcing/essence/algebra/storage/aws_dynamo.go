package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/widmogrod/mkunion/x/schema"
)

func NewDynamoDBRepository[A any](client *dynamodb.Client, tableName string, new func() A) *DynamoDBRepository[A] {
	return &DynamoDBRepository[A]{
		client:    client,
		tableName: tableName,
		new:       new,
	}
}

var _ Repository[any] = (*DynamoDBRepository[any])(nil)

type DynamoDBRepository[A any] struct {
	tableName string
	client    *dynamodb.Client
	new       func() A
}

func (d *DynamoDBRepository[A]) Get(key string) (A, error) {
	var a A

	item, err := d.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{
				Value: key,
			},
		},
		TableName:      &d.tableName,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return a, err
	}

	if len(item.Item) == 0 {
		return a, ErrNotFound
	}

	return d.toStruct(item.Item)
}

func (d *DynamoDBRepository[A]) toStruct(item map[string]types.AttributeValue) (A, error) {
	var a A
	delete(item, "key")

	i := &types.AttributeValueMemberM{
		Value: item,
	}
	sch, err := schema.FromDynamoDB(i)
	if err != nil {
		return a, err
	}

	var obj any

	// TODO fix me!!!
	if any(a) == nil {
		obj = schema.ToGo(sch)
	} else {
		obj = schema.ToGo(sch, schema.WhenPath(nil, schema.UseStruct(a)))
	}

	if result, ok := obj.(A); ok {
		return result, nil
	} else {
		return a, fmt.Errorf("could not convert object to type %T", a)
	}
}

func (d *DynamoDBRepository[A]) Set(key string, value A) error {
	sch := schema.FromGo(value)
	item := schema.ToDynamoDB(sch)

	if _, ok := item.(*types.AttributeValueMemberM); !ok {
		return fmt.Errorf("DynamoDBRepository.Set: unsupported type: %T", item)
	}

	final := item.(*types.AttributeValueMemberM)
	final.Value["key"] = &types.AttributeValueMemberS{
		Value: key,
	}

	_, err := d.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		Item:      final.Value,
		TableName: &d.tableName,
	})
	return err
}

func (d *DynamoDBRepository[A]) Delete(key string) error {
	_, err := d.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{
				Value: key,
			},
		},
		TableName: &d.tableName,
	})
	return err
}

func (d *DynamoDBRepository[A]) GetOrNew(s string) (A, error) {
	v, err := d.Get(s)
	if err == nil {
		return v, nil
	}

	if err != nil && err != ErrNotFound {
		var a A
		return a, err
	}

	v = d.new()

	err = d.Set(s, v)
	if err != nil {
		var a A
		return a, err
	}

	return v, nil
}

func (d *DynamoDBRepository[A]) FindAllKeyEqual(key string, value string) (PageResult[A], error) {
	items, err := d.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: &d.tableName,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":value": &types.AttributeValueMemberS{
				Value: value,
			},
		},
		FilterExpression: aws.String(fmt.Sprintf("%s = :value", key)),
		ConsistentRead:   aws.Bool(true),
	})

	if err != nil {
		return PageResult[A]{}, err
	}

	result := PageResult[A]{
		Items: []A{},
	}

	for _, item := range items.Items {
		r, err := d.toStruct(item)
		if err != nil {
			return PageResult[A]{}, err
		}
		result.Items = append(result.Items, r)
	}

	return result, nil
}