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
	"strings"
	"time"
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

const queryTemplate = `{{
  "size": 0,
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "Data.M.SessionInGame.M.SessionID.S.keyword": {
              "value": "%s"
            }
          }
        },
        {
          "prefix": {
           "Type.S.keyword": "game"
          }
        }
      ]
    }
  },
  "aggs": {
    "wins": {
      "terms": {
        "field": "Data.M.SessionInGame.M.GameState.M.GameEndWithWin.M.Winner.S.keyword"
      }
    },
    "draws": {
      "terms": {
        "field": "Data.M.SessionInGame.M.GameState.M.GameEndWithDraw.M.TicTacToeBaseState.M.BoardCols.N.keyword"
      }
    }
  }
}`

func (os *OpenSearchStorage) SetupIndex() error {
	_, err := os.client.Indices.Create(
		os.indexName,
		func(request *opensearchapi.IndicesCreateRequest) {
			request.Body = strings.NewReader(`{
  "settings": {
    "index": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "refresh_interval": "100ms"
    }
  }
}
`)
		})

	return err
}

func (os *OpenSearchStorage) Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error) {
	response, err := os.client.Search(func(request *opensearchapi.SearchRequest) {
		request.Timeout = 100 * time.Microsecond
		request.Pretty = false
		request.Index = []string{os.indexName}
		request.Body = strings.NewReader(fmt.Sprintf(queryTemplate, query.SessionID))
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
		TotalDraws: schema.AsDefault[int](schema.Get(data, "aggregations.draws.buckets.[0].doc_count"), 0),
		PlayerWins: schema.Reduce[map[tictactoemanage.PlayerID]int](
			schema.Get(data, "aggregations.wins.buckets"),
			map[tictactoemanage.PlayerID]int{},
			func(x schema.Schema, agg map[tictactoemanage.PlayerID]int) map[tictactoemanage.PlayerID]int {
				playerID := schema.AsDefault[tictactoemanage.PlayerID](schema.Get(x, "key"), "n/a")
				winds := schema.AsDefault[int](schema.Get(x, "doc_count"), 0)
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
