package tictactoe_game_server

import "github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"

type OpenSearchStorage struct {
}

func (os *OpenSearchStorage) Query(query tictactoemanage.SessionStatsQuery) tictactoemanage.SessionStats {
	_ = `GET lambda-index/_search
{
  "size": 0, 
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "SessionInGame.M.ID.S.keyword": {
              "value": "605e54ac-1d84-4ccf-9004-df4a21c98d5f"
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

	return tictactoemanage.SessionStats{
		ID:         "",
		TotalGames: 0,
		TotalDraws: 0,
		PlayerWins: map[tictactoemanage.PlayerID]float64{},
	}
}
