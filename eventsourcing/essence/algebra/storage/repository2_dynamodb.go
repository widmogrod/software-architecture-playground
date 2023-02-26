package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"strings"
)

func NewDynamoDBRepository2(client *dynamodb.Client, tableName string) *DynamoDBRepository2 {
	return &DynamoDBRepository2{
		client:    client,
		tableName: tableName,
	}
}

var _ Repository2[schema.Schema] = (*DynamoDBRepository2)(nil)

type DynamoDBRepository2 struct {
	client    *dynamodb.Client
	tableName string
}

func (d *DynamoDBRepository2) Get(key, recordType string) (Record[schema.Schema], error) {
	item, err := d.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{
				Value: key,
			},
			"Type": &types.AttributeValueMemberS{
				Value: recordType,
			},
		},
		TableName:      &d.tableName,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return Record[schema.Schema]{}, fmt.Errorf("DynamoDBRepository2.Get error=%s. %w", err, ErrInternalError)
	}

	if len(item.Item) == 0 {
		return Record[schema.Schema]{}, fmt.Errorf("DynamoDBRepository2.Get not found. %w", ErrNotFound)
	}

	i := &types.AttributeValueMemberM{
		Value: item.Item,
	}

	schemed, err := schema.FromDynamoDB(i)
	if err != nil {
		return Record[schema.Schema]{}, fmt.Errorf("DynamoDBRepository2.Get schema conversion error=%s. %w", err, ErrInternalError)
	}

	typed, err := d.toTyped(schemed)
	if err != nil {
		return Record[schema.Schema]{}, fmt.Errorf("DynamoDBRepository2.Get type conversion error=%s. %w", err, ErrInvalidType)
	}

	return typed, nil
}

func (d *DynamoDBRepository2) UpdateRecords(command UpdateRecords[Record[schema.Schema]]) error {
	if command.IsEmpty() {
		return fmt.Errorf("DynamoDBRepository2.UpdateRecords: empty command %w", ErrEmptyCommand)
	}

	var transact []types.TransactWriteItem
	for _, value := range command.Saving {
		originalVersion := value.Version
		value.Version++
		sch := d.fromTyped(value)
		item := schema.ToDynamoDB(sch)
		if _, ok := item.(*types.AttributeValueMemberM); !ok {
			return fmt.Errorf("DynamoDBRepository2.UpdateRecords: unsupported type: %T", item)
		}

		final, ok := item.(*types.AttributeValueMemberM)
		if !ok {
			return fmt.Errorf("DynamoDBRepository2.UpdateRecords: expected map as item. %w", ErrInternalError)
		}

		transact = append(transact, types.TransactWriteItem{
			Put: &types.Put{
				TableName:           aws.String(d.tableName),
				Item:                final.Value,
				ConditionExpression: aws.String("Version = :version OR attribute_not_exists(Version)"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":version": &types.AttributeValueMemberN{
						Value: fmt.Sprintf("%d", originalVersion),
					},
				},
			},
		})
	}

	for _, id := range command.Deleting {
		transact = append(transact, types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: aws.String(d.tableName),
				Key: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberS{
						Value: id.ID,
					},
					"Type": &types.AttributeValueMemberS{
						Value: id.Type,
					},
				},
			},
		})
	}

	_, err := d.client.TransactWriteItems(context.Background(), &dynamodb.TransactWriteItemsInput{
		TransactItems: transact,
	})

	if err != nil {
		respErr := &http.ResponseError{}
		if errors.As(err, &respErr) {
			conditional := &types.TransactionCanceledException{}
			if errors.As(respErr.ResponseError.Err, &conditional) {
				for _, reason := range conditional.CancellationReasons {
					if *reason.Code == "ConditionalCheckFailed" {
						return fmt.Errorf("storage.DynamoDBRepository.UpdateRecords: %w", ErrVersionConflict)
					}
				}
			}
		}
		return err
	}

	return nil
}

func (d *DynamoDBRepository2) FindingRecords(query FindingRecords[Record[schema.Schema]]) (PageResult[Record[schema.Schema]], error) {
	filterExpression, paramsExpression, expressionNames, err := d.buildFilterExpression(query)
	if err != nil {
		return PageResult[Record[schema.Schema]]{}, err
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 &d.tableName,
		ExpressionAttributeNames:  expressionNames,
		ExpressionAttributeValues: paramsExpression,
		FilterExpression:          aws.String(filterExpression),
		//ConsistentRead:            aws.Bool(true),
	}

	if query.After != nil {
		schemed, err := schema.FromJSON([]byte(*query.After))
		if err != nil {
			return PageResult[Record[schema.Schema]]{}, err
		}

		scanInput.ExclusiveStartKey = map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{
				Value: schema.As[string](schema.Get(schemed, "ID"), ""),
			},
			"Type": &types.AttributeValueMemberS{
				Value: schema.As[string](schema.Get(schemed, "Type"), ""),
			},
		}
	}

	// Be aware that DynamoDB limit is scan limit, not page limit!
	if query.Limit > 0 {
		scanInput.Limit = aws.Int32(int32(query.Limit))
	}

	items, err := d.client.Scan(context.Background(), scanInput)
	if err != nil {
		return PageResult[Record[schema.Schema]]{}, err
	}

	result := PageResult[Record[schema.Schema]]{
		Items: nil,
	}

	for _, item := range items.Items {
		// normalize input for further processing
		i := &types.AttributeValueMemberM{
			Value: item,
		}

		schemed, err := schema.FromDynamoDB(i)
		if err != nil {
			return PageResult[Record[schema.Schema]]{}, err
		}

		typed, err := d.toTyped(schemed)
		if err != nil {
			return PageResult[Record[schema.Schema]]{}, err
		}
		result.Items = append(result.Items, typed)
	}

	if items.LastEvaluatedKey != nil {
		after := &types.AttributeValueMemberM{
			Value: items.LastEvaluatedKey,
		}
		schemed, err := schema.FromDynamoDB(after)
		if err != nil {
			return PageResult[Record[schema.Schema]]{}, fmt.Errorf("DynamoDBRepository2.FindingRecords: error calculating after cursor %s. %w", err, ErrInternalError)
		}
		json, err := schema.ToJSON(schemed)
		if err != nil {
			return PageResult[Record[schema.Schema]]{}, fmt.Errorf("DynamoDBRepository2.FindingRecords: error serializing after cursor %s. %w", err, ErrInternalError)
		}
		cursor := string(json)
		result.Next = &FindingRecords[Record[schema.Schema]]{
			Where: query.Where,
			Sort:  query.Sort,
			Limit: query.Limit,
			After: &cursor,
		}
	}

	return result, nil
}

func (d *DynamoDBRepository2) fromTyped(record Record[schema.Schema]) *schema.Map {
	return schema.MkMap(
		schema.MkField("ID", schema.MkString(record.ID)),
		schema.MkField("Type", schema.MkString(record.Type)),
		schema.MkField("Data", record.Data),
		schema.MkField("Version", schema.MkInt(int(record.Version))),
	)
}

func (d *DynamoDBRepository2) toTyped(record schema.Schema) (Record[schema.Schema], error) {
	typed := Record[schema.Schema]{
		ID:      schema.As[string](schema.Get(record, "ID"), "record-id-corrupted"),
		Type:    schema.As[string](schema.Get(record, "Type"), "record-type-corrupted"),
		Data:    schema.Get(record, "Data"),
		Version: schema.As[uint16](schema.Get(record, "Version"), 0),
	}
	if typed.Type == "record-id-corrupted" &&
		typed.ID == "record-id-corrupted" &&
		typed.Version == 0 {
		return Record[schema.Schema]{}, fmt.Errorf("store.DynamoDBRepository2.FindingRecords corrupted record: %v", record)
	}
	return typed, nil
}

func (d *DynamoDBRepository2) buildFilterExpression(query FindingRecords[Record[schema.Schema]]) (string, map[string]types.AttributeValue, map[string]string, error) {
	var where predicate.Predicate
	var binds predicate.ParamBinds = map[predicate.BindValue]schema.Schema{}
	var names map[string]string = map[string]string{}

	if query.RecordType != "" {
		names["Type"] = "#Type"
		where = &predicate.Compare{
			Location:  "Type",
			Operation: "=",
			BindValue: ":Type",
		}
		binds[":Type"] = schema.MkString(query.RecordType)
	}

	if query.Where != nil {
		if where == nil {
			where = query.Where.Predicate
			binds = query.Where.Params
		} else {
			where = &predicate.And{
				L: []predicate.Predicate{where, query.Where.Predicate},
			}

			for k, v := range query.Where.Params {
				if _, ok := binds[k]; ok {
					return "", nil, nil, fmt.Errorf("store.DynamoDBRepository2.FindingRecords: duplicated bind value: %s", k)
				}

				binds[k] = v
			}
		}
	}

	if where == nil {
		return "", nil, nil, nil
	}

	expression := toExpression(where, names)

	// reverse names
	reverser := map[string]string{}
	for k, v := range names {
		reverser[v] = k
	}

	return expression, toAttributes(binds), reverser, nil
}

func toExpression(where predicate.Predicate, names map[string]string) string {
	return predicate.MustMatchPredicate(
		where,
		func(x *predicate.And) string {
			var result []string
			for _, v := range x.L {
				result = append(result, toExpression(v, names))
			}

			return strings.Join(result, " AND ")
		},
		func(x *predicate.Or) string {
			var result []string
			for _, v := range x.L {
				result = append(result, toExpression(v, names))
			}

			return strings.Join(result, " OR ")

		},
		func(x *predicate.Not) string {
			return "NOT " + toExpression(x.P, names)
		},
		func(x *predicate.Compare) string {
			// Because of https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.ExpressionAttributeNames.html
			// we need to make sure that all names are not reserved keyword, so we add a counter to the end of the name in case of collision
			var named []string
			var parts []string = strings.Split(x.Location, ".")

			for _, part := range parts {
				if _, ok := names[part]; !ok {
					names[part] = "#" + part
				}
				named = append(named, names[part])
			}

			return strings.Join(named, ".") + " " + x.Operation + " " + x.BindValue
		},
	)
}

func toAttributes(binds predicate.ParamBinds) map[string]types.AttributeValue {
	result := map[string]types.AttributeValue{}
	for k, v := range binds {
		result[k] = schema.ToDynamoDB(v)
	}

	return result
}
