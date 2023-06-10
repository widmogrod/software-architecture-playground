package tictactoe_game_server

import (
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
)

func NewQueryUsingStorage(store schemaless.Repository[tictactoemanage.SessionStatsResult]) *QueryStorage {
	return &QueryStorage{
		store: store,
	}
}

type QueryStorage struct {
	store schemaless.Repository[tictactoemanage.SessionStatsResult]
}

func (q *QueryStorage) Query(query tictactoemanage.SessionStatsQuery) (*tictactoemanage.SessionStatsResult, error) {
	result, err := q.store.Get(query.SessionID, "session-stats")
	if err != nil {
		return nil, err
	}

	return &result.Data, nil
}
