package schemaless

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	log "github.com/sirupsen/logrus"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
	"io"
)

func NewOpenSearchRepository(client *opensearch.Client, index string) *OpenSearchRepository {
	return &OpenSearchRepository{
		client:    client,
		indexName: index,
	}
}

var _ Repository[schema.Schema] = (*OpenSearchRepository)(nil)

type OpenSearchRepository struct {
	client    *opensearch.Client
	indexName string
}

func (os *OpenSearchRepository) Get(recordID string, recordType RecordType) (Record[schema.Schema], error) {
	response, err := os.client.Get(os.indexName, os.recordID(recordType, recordID))
	if err != nil {
		return Record[schema.Schema]{}, err
	}

	result, err := io.ReadAll(response.Body)
	if err != nil {
		return Record[schema.Schema]{}, err
	}

	log.Println("OpenSearchRepository.Get result=", string(result))

	schemed, err := schema.FromJSON(result)
	if err != nil {
		return Record[schema.Schema]{}, err
	}

	typed, err := os.toTyped(schema.Get(schemed, "_source"))
	if err != nil {
		return Record[schema.Schema]{}, fmt.Errorf("DynamoDBRepository.Get type conversion error=%s. %w", err, ErrInvalidType)
	}

	return typed, nil
}

func (os *OpenSearchRepository) UpdateRecords(command UpdateRecords[Record[schema.Schema]]) error {
	for _, record := range command.Saving {
		data, err := schema.ToJSON(os.fromTyped(record))
		if err != nil {
			panic(err)
		}
		_, err = os.client.Index(os.indexName, bytes.NewReader(data), func(request *opensearchapi.IndexRequest) {
			request.DocumentID = os.toDocumentID(record)
		})
		if err != nil {
			panic(err)
		}
	}

	for _, record := range command.Deleting {
		_, err := os.client.Delete(os.indexName, os.toDocumentID(record))
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (os *OpenSearchRepository) FindingRecords(query FindingRecords[Record[schema.Schema]]) (PageResult[Record[schema.Schema]], error) {
	filters, sorters := os.toFiltersAndSorters(query)

	queryTemplate := map[string]any{}
	if query.Limit > 0 {
		if query.Limit > 0 {
			// add as last sorter _id, so that we can use search_after
			sorters = append(sorters, map[string]any{
				"_id": map[string]any{
					"order": "asc",
				},
			})
		}
		queryTemplate["size"] = query.Limit
	}

	if query.After != nil {
		schemed, err := schema.FromJSON([]byte(*query.After))
		if err != nil {
			panic(err)
		}

		list, ok := schemed.(*schema.List)
		if !ok {
			panic(fmt.Errorf("expected list, got %T", schemed))
		}

		afterSearch := make([]string, len(list.Items))
		for i, item := range list.Items {
			str, ok := schema.As[string](item)
			if !ok {
				panic(fmt.Errorf("expected string, got %T", item))
			}
			afterSearch[i] = str
		}

		queryTemplate["search_after"] = afterSearch
	}

	if len(filters) > 0 {
		queryTemplate["query"] = filters
	}
	if len(sorters) > 0 {
		queryTemplate["sort"] = sorters
	}

	response, err := os.client.Search(func(request *opensearchapi.SearchRequest) {
		request.Index = []string{
			os.indexName,
		}
		body, err := json.Marshal(queryTemplate)
		if err != nil {
			panic(err)
		}

		log.Println("OpenSearchRepository) FindingRecords ", string(body))

		request.Body = bytes.NewReader(body)
	})
	if err != nil {
		panic(err)
		//return PageResult[Record[schema.Schema]]{}, err
	}
	result, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
		//return PageResult[Record[schema.Schema]]{}, err
	}

	schemed, err := schema.FromJSON(result)
	if err != nil {
		panic(err)
		//return PageResult[Record[schema.Schema]]{}, err
	}

	hists := schema.Get(schemed, "hits.hits")
	var lastSort schema.Schema

	items := schema.Reduce(
		hists,
		[]Record[schema.Schema]{},
		func(s schema.Schema, result []Record[schema.Schema]) []Record[schema.Schema] {
			typed, err := os.toTyped(schema.Get(s, "_source"))
			if err != nil {
				panic(err)
			}
			result = append(result, typed)

			lastSort = schema.Get(s, "sort")

			return result
		})

	if len(items) == int(query.Limit) && lastSort != nil {
		// has next page of results
		next := query

		data, err := schema.ToJSON(lastSort)
		if err != nil {
			panic(err)
		}
		after := string(data)
		next.After = &after

		return PageResult[Record[schema.Schema]]{
			Items: items,
			Next:  &next,
		}, nil
	}

	return PageResult[Record[schema.Schema]]{
		Items: items,
		Next:  nil,
	}, nil
}

func (os *OpenSearchRepository) fromTyped(record Record[schema.Schema]) *schema.Map {
	return schema.MkMap(
		schema.MkField("ID", schema.MkString(record.ID)),
		schema.MkField("Type", schema.MkString(record.Type)),
		schema.MkField("Data", record.Data),
		schema.MkField("Version", schema.MkInt(int(record.Version))),
	)
}

func (os *OpenSearchRepository) toTyped(record schema.Schema) (Record[schema.Schema], error) {
	typed := Record[schema.Schema]{
		ID:      schema.AsDefault[string](schema.Get(record, "ID"), "record-id-corrupted"),
		Type:    schema.AsDefault[string](schema.Get(record, "Type"), "record-id-corrupted"),
		Data:    schema.Get(record, "Data"),
		Version: schema.AsDefault[uint16](schema.Get(record, "Version"), 0),
	}
	if typed.Type == "record-id-corrupted" &&
		typed.ID == "record-id-corrupted" &&
		typed.Version == 0 {
		return Record[schema.Schema]{}, fmt.Errorf("store.DynamoDBRepository.FindingRecords corrupted record: %v", record)
	}
	return typed, nil
}

func (os *OpenSearchRepository) toDocumentID(record Record[schema.Schema]) string {
	return os.recordID(record.Type, record.ID)
}

func (os *OpenSearchRepository) recordID(recordType, recordID string) string {
	return fmt.Sprintf("%s-%s", recordType, recordID)
}

func (os *OpenSearchRepository) toFiltersAndSorters(query FindingRecords[Record[schema.Schema]]) (filters map[string]any, sorters []any) {
	filters = os.toFilters(
		predicate.Optimize(query.Where.Predicate),
		query.Where.Params,
	)

	if query.RecordType != "" {
		if filters["bool"] == nil {
			filters["bool"] = map[string]any{}
		}
		if filters["bool"].(map[string]any)["must"] == nil {
			filters["bool"].(map[string]any)["must"] = []any{}
		}
		filters["bool"].(map[string]any)["must"] = append(filters["bool"].(map[string]any)["must"].([]any), map[string]any{
			"term": map[string]any{
				"Type.keyword": query.RecordType,
			},
		})
	}

	sorters = os.ToSorters(query.Sort)

	return
}

var mapOfOperationToOpenSearchQuery = map[string]string{
	">":  "gt",
	">=": "gte",
	"<":  "lt",
	"<=": "lte",
}

func (os *OpenSearchRepository) toFilters(p predicate.Predicate, params predicate.ParamBinds) map[string]any {
	return predicate.MustMatchPredicate(
		p,
		func(x *predicate.And) map[string]any {
			var must []any
			for _, pred := range x.L {
				must = append(must, os.toFilters(pred, params))
			}
			return map[string]any{
				"bool": map[string]any{
					"must": must,
				},
			}
		},
		func(x *predicate.Or) map[string]any {
			var should []any
			for _, pred := range x.L {
				should = append(should, os.toFilters(pred, params))
			}
			return map[string]any{
				"bool": map[string]any{
					"should": should,
				},
			}
		},
		func(x *predicate.Not) map[string]any {
			return map[string]any{
				"bool": map[string]any{
					"must_not": os.toFilters(x.P, params),
				},
			}
		},
		func(x *predicate.Compare) map[string]any {
			switch x.Operation {
			case "=":
				return map[string]any{
					"term": map[string]any{
						fmt.Sprintf("%s.keyword", x.Location): params[x.BindValue],
					},
				}

			case "!=":
				return map[string]any{
					"bool": map[string]any{
						"must_not": map[string]any{
							"term": map[string]any{
								fmt.Sprintf("%s.keyword", x.Location): params[x.BindValue],
							},
						},
					},
				}

			case ">", ">=", "<", "<=":
				return map[string]any{
					"range": map[string]any{
						x.Location: map[string]any{
							mapOfOperationToOpenSearchQuery[x.Operation]: params[x.BindValue],
						},
					},
				}
			}

			panic(fmt.Errorf("store.OpenSearchRepository.toFilters: unknown operation %s", x.Operation))
		},
	)
}

func (os *OpenSearchRepository) ToSorters(sort []SortField) []any {
	var sorters []any
	for _, s := range sort {
		if s.Descending {
			sorters = append(sorters, map[string]any{
				fmt.Sprintf("%s.keyword", s.Field): map[string]any{
					"order": "desc",
				},
			})
		} else {
			sorters = append(sorters, map[string]any{
				fmt.Sprintf("%s.keyword", s.Field): map[string]any{
					"order": "asc",
				},
			})
		}
	}

	return sorters
}

//func (os *OpenSearchRepository) Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error) {
//	response, err := os.client.Search(func(request *opensearchapi.SearchRequest) {
//		request.Timeout = 100 * time.Microsecond
//		request.Pretty = false
//		request.Index = []string{os.indexName}
//		request.Body = strings.NewReader(fmt.Sprintf(queryTemplate, query.SessionID))
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	result, err := io.ReadAll(response.Body)
//	if err != nil {
//		return nil, err
//	}
//
//	data, err := schema.FromJSON(result)
//	if err != nil {
//		return nil, err
//	}
//
//}
