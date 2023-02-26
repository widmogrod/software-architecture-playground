package tictactoe_game_server

import (
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
)

func NewQueryUsingStorage(store storage.Repository2[tictactoemanage.SessionStatsResult]) *QueryStorage {
	return &QueryStorage{
		store: store,
	}
}

type QueryStorage struct {
	store storage.Repository2[tictactoemanage.SessionStatsResult]
}

func (q *QueryStorage) Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error) {
	result, err := q.store.Get("session-stats:" + query.SessionID)
	if err != nil {
		return nil, err
	}

	return &result.Data, nil
}
