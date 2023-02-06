package tictactoe_game_server

import (
	"crypto/tls"
	"fmt"
	opensearch "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func NewQuery(endpoint, index string) (*OpenSearchStorage, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			endpoint,
		},
		Username: "admin",
		Password: "nile!DISLODGE5clause",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &OpenSearchStorage{
		client:    client,
		indexName: index,
	}, nil
}

type OpenSearchStorage struct {
	client    *opensearch.Client
	indexName string
}

const queryTemplate = `{
  "size": 0,
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "SessionInGame.M.ID.S.keyword": {
              "value": "%s"
            }
          }
        }
      ]
    }
  },
  "aggs": {
    "wins": {
      "terms": {
        "field": "SessionInGame.M.GameState.M.GameEndWithWin.M.Winner.S.keyword"
      }
    },
    "draws": {
      "terms": {
        "field": "SessionInGame.M.GameState.M.GameEndWithDraw.M.TicTacToeBaseState.M.BoardCols.N.keyword"
      }
    }
  }
}`

func (os *OpenSearchStorage) Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error) {
	response, err := os.client.Search(func(request *opensearchapi.SearchRequest) {
		request.Index = []string{os.indexName}

		request.Body = strings.NewReader(fmt.Sprintf(queryTemplate, query.SessionID))
		request.Pretty = true
	})
	if err != nil {
		return nil, err
	}

	result, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	data, err := schema.FromJSON(result)
	if err != nil {
		return nil, err
	}

	stats := &tictactoemanage.SessionStatsResult{
		ID:         query.SessionID,
		TotalDraws: As[int](Get(data, "aggregations.draws.buckets.[0].doc_count"), 0),
		PlayerWins: Reduce[map[tictactoemanage.PlayerID]float64](
			Get(data, "aggregations.wins.buckets"),
			func(x schema.Schema, agg map[tictactoemanage.PlayerID]float64) map[tictactoemanage.PlayerID]float64 {
				playerID := As[tictactoemanage.PlayerID](Get(x, "key"), "n/a")
				winds := As[float64](Get(x, "doc_count"), 0)

				if agg == nil {
					agg = map[tictactoemanage.PlayerID]float64{}
				}

				agg[playerID] = winds

				return agg
			}),
	}

	stats.TotalGames += stats.TotalDraws
	for _, v := range stats.PlayerWins {
		stats.TotalGames += int(v)
	}

	return stats, nil
}

func As[A int | float64 | bool | string](x schema.Schema, def A) A {
	if x == nil {
		return def
	}

	return schema.MustMatchSchema(
		x,
		func(x *schema.None) A {
			return def
		},
		func(x *schema.Bool) A {
			switch any(def).(type) {
			case bool:
				return any(bool(*x)).(A)
			}

			return def
		},
		func(x *schema.Number) A {
			switch any(def).(type) {
			case float64:
				return any(float64(*x)).(A)
			case int:
				return any(int(*x)).(A)
			}

			return def
		},
		func(x *schema.String) A {
			switch any(def).(type) {
			case string:
				return any(string(*x)).(A)
			}

			return def
		},
		func(x *schema.List) A {
			return def
		},
		func(x *schema.Map) A {
			return def
		})
}

func Get(data schema.Schema, location string) schema.Schema {
	path := strings.Split(location, ".")
	for _, p := range path {
		if p == "" {
			return nil
		}

		if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
			idx := strings.TrimPrefix(p, "[")
			idx = strings.TrimSuffix(idx, "]")
			i, err := strconv.Atoi(idx)
			if err != nil {
				return nil
			}

			listData, ok := data.(*schema.List)
			if ok && len(listData.Items) > i {
				data = listData.Items[i]
				continue
			}

			return nil
		}

		mapData, ok := data.(*schema.Map)
		if !ok {
			return nil
		}

		var found bool
		for _, item := range mapData.Field {
			if item.Name == p {
				found = true
				data = item.Value
				break
			}
		}

		if !found {
			return nil
		}
	}

	return data
}

func Reduce[B any](data schema.Schema, fn func(schema.Schema, B) B) B {
	var agg B

	if data == nil {
		return agg
	}

	return schema.MustMatchSchema(
		data,
		func(x *schema.None) B {
			return agg
		},
		func(x *schema.Bool) B {
			return fn(x, agg)
		},
		func(x *schema.Number) B {
			return fn(x, agg)

		},
		func(x *schema.String) B {
			return fn(x, agg)
		},
		func(x *schema.List) B {
			for _, y := range x.Items {
				agg = fn(y, agg)
			}

			return agg
		},
		func(x *schema.Map) B {
			for _, y := range x.Field {
				agg = fn(y.Value, agg)
			}

			return agg
		})
}
