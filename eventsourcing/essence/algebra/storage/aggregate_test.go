package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"testing"
)

var latestGames = []tictactoemanage.State{
	&tictactoemanage.SessionInGame{
		SessionID:   "session-2",
		Players:     []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:      "game-2.1",
		GameState:   nil,
		GameProblem: nil,
	},
	&tictactoemanage.SessionInGame{
		SessionID: "session-2",
		Players:   []tictactoemanage.PlayerID{"player-1", "player-666"},
		GameID:    "game-2.2",
		GameState: &tictacstatemachine.GameEndWithWin{
			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
				FirstPlayerID:  "player-1",
				SecondPlayerID: "player-666",
				BoardRows:      3,
				BoardCols:      3,
				WinningLength:  3,
			},
			Winner:         "player-666",
			WiningSequence: []tictacstatemachine.Move{"1.1", "1.2", "1.3"},
			MovesTaken:     map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
		},
	},
	&tictactoemanage.SessionInGame{
		SessionID: "session-1",
		Players:   []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:    "game-2",
		GameState: &tictacstatemachine.GameEndWithDraw{
			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
				FirstPlayerID:  "player-1",
				SecondPlayerID: "player-2",
				BoardRows:      3,
				BoardCols:      3,
				WinningLength:  3,
			},
			MovesTaken: map[tictacstatemachine.Move]tictacstatemachine.PlayerID{
				"1.1": "player-1",
				"1.2": "player-2",
			},
		},
		GameProblem: nil,
	},
	&tictactoemanage.SessionInGame{
		SessionID: "session-1",
		Players:   []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:    "game-3",
		GameState: &tictacstatemachine.GameEndWithWin{
			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
				FirstPlayerID:  "player-1",
				SecondPlayerID: "player-2",
				BoardRows:      3,
				BoardCols:      3,
				WinningLength:  3,
			},
			Winner:         "player-1",
			WiningSequence: []tictacstatemachine.Move{"1.1", "1.2", "1.3"},
			MovesTaken:     map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
		},
		GameProblem: nil,
	},
	&tictactoemanage.SessionInGame{
		SessionID: "session-1",
		Players:   []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:    "game-4",
		GameState: &tictacstatemachine.GameEndWithWin{
			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
				FirstPlayerID:  "player-1",
				SecondPlayerID: "player-2",
				BoardRows:      3,
				BoardCols:      3,
				WinningLength:  3,
			},
			Winner:         "player-1",
			WiningSequence: []tictacstatemachine.Move{"1.1", "1.2", "1.3"},
			MovesTaken:     map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
		},
		GameProblem: nil,
	},
}

func InitEmtpy(id tictactoemanage.SessionID) *tictactoemanage.SessionStatsResult {
	return &tictactoemanage.SessionStatsResult{
		ID:         id,
		TotalGames: 0,
		TotalDraws: 0,
		PlayerWins: nil,
	}
}

func GroupByKey(data tictactoemanage.State) (string, *tictactoemanage.SessionStatsResult) {
	return tictactoemanage.MustMatchStateR2(
		data,
		func(x *tictactoemanage.SessionWaitingForPlayers) (string, *tictactoemanage.SessionStatsResult) {
			return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
		},
		func(x *tictactoemanage.SessionReady) (string, *tictactoemanage.SessionStatsResult) {
			return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
		},
		func(x *tictactoemanage.SessionInGame) (string, *tictactoemanage.SessionStatsResult) {
			if x.GameState == nil {
				return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
			}

			return tictacstatemachine.MustMatchStateR2(
				x.GameState,
				func(y *tictacstatemachine.GameWaitingForPlayer) (string, *tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
				},
				func(y *tictacstatemachine.GameProgress) (string, *tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
				},
				func(y *tictacstatemachine.GameEndWithWin) (string, *tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, &tictactoemanage.SessionStatsResult{
						ID:         x.SessionID,
						TotalGames: 1,
						TotalDraws: 0,
						PlayerWins: map[tictactoemanage.PlayerID]int{
							y.Winner: 1,
						},
					}
				},
				func(y *tictacstatemachine.GameEndWithDraw) (string, *tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, &tictactoemanage.SessionStatsResult{
						ID:         x.SessionID,
						TotalGames: 1,
						TotalDraws: 1,
						PlayerWins: nil,
					}
				},
			)
		},
	)
}

/*
CombineByKey is commutative, associative, and distributive.

	commutativity = a * b = b * a
	associativity = (a * b) * c = a * (b * c)
	distributivity = a * (b + c) = (a * b) + (a * c)
*/
func CombineByKey(a, b *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
	winds := a.PlayerWins
	if winds == nil {
		winds = map[tictactoemanage.PlayerID]int{}
	}

	for k, v := range b.PlayerWins {
		winds[k] += v
	}

	return &tictactoemanage.SessionStatsResult{
		ID:         a.ID,
		TotalGames: a.TotalGames + b.TotalGames,
		TotalDraws: a.TotalDraws + b.TotalDraws,
		PlayerWins: winds,
	}, nil
}

func TestIndexer(t *testing.T) {
	storage := NewRepository2WithSchema()

	indexer := NewKeyedAggregate[tictactoemanage.State, *tictactoemanage.SessionStatsResult](
		GroupByKey,
		CombineByKey,
		storage,
	)

	_ = `CREATE QUERY "session-stats" ON games as g WITH 
		PRIMARY_KEY session-stats.sessionID
		SELECT 
			sessionID
     		, count(types.GameEndWithWin) as winds
     		, count(types.GameEndWithDraw) as draws
     		, winds + draws as total 
		GROUP BY g.sessionID, g.GameState FIELD_NAME ['GameEndWithWin', 'GameEndWithDraw'] as types
		WHERE g.GameState HAS_FIELD_IN ['GameEndWithWin', 'GameEndWithDraw']; 
`

	/*
		In Contrast to CombineByKey, UncombineByKey is not commutative, and not associative.
			anticommute = a - b = -b + a
			not associative = a - (b - c) = (a - b) + c

		Example:
			3 - (2 - 1) = (3 - 2) + 1 = 2

		Some operations should marked as "substractive"
	*/
	//aggregate.UncombineByKey(func(a, b *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
	//	winds := a.PlayerWins
	//	for k, v := range b.PlayerWins {
	//		winds[k] += v
	//	}
	//
	//	return &tictactoemanage.SessionStatsResult{
	//		SessionID:         a.SessionID,
	//		TotalGames: a.TotalGames - b.TotalGames,
	//		TotalDraws: a.TotalDraws - b.TotalDraws,
	//		PlayerWins: winds,
	//	}, nil
	//})

	//aggregate.Init(func(groupContext GroupContext2[tictactoemanage.SessionID, tictacstatemachine.State]) {
	//	sessionId, _ := groupContext.GroupKey()
	//	//groupContext.Unpack()
	//
	//	return &tictactoemanage.SessionStatsResult{
	//		SessionID:         sessionId,
	//		TotalGames: 0,
	//		TotalDraws: 0,
	//		PlayerWins: map[tictactoemanage.PlayerID]int{},
	//	}
	//})

	//aggregate.OnInsert(func(groupContext GroupContext2[tictactoemanage.SessionID, tictacstatemachine.State], stats *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
	//	_, gamestate := groupContext.GroupKey()
	//	return tictacstatemachine.MustMatchStateR2(
	//		gamestate,
	//		func(y *tictacstatemachine.GameWaitingForPlayer) (*tictactoemanage.SessionStatsResult, error) {
	//			return stats, nil
	//		},
	//		func(y *tictacstatemachine.GameProgress) (*tictactoemanage.SessionStatsResult, error) {
	//			return stats, nil
	//		},
	//		func(y *tictacstatemachine.GameEndWithWin) (*tictactoemanage.SessionStatsResult, error) {
	//			stats.TotalGames += 1
	//			stats.PlayerWins[y.Winner] += 1
	//			return stats, nil
	//		},
	//		func(y *tictacstatemachine.GameEndWithDraw) (*tictactoemanage.SessionStatsResult, error) {
	//			stats.TotalGames += 1
	//			stats.TotalDraws += 1
	//			return stats, nil
	//		},
	//	)
	//})

	for _, game := range latestGames {
		err := indexer.Append(game)
		assert.NoError(t, err)
	}

	result := indexer.GetIndexByKey("session-stats:session-1")

	assert.Equal(t, &tictactoemanage.SessionStatsResult{
		ID:         "session-1",
		TotalGames: 3,
		TotalDraws: 1,
		PlayerWins: map[tictactoemanage.PlayerID]int{
			"player-1": 2,
		},
	}, result)
}

func TestIndexingWithRepository(t *testing.T) {
	storage := NewRepository2WithSchema()

	// Simulate, that we have a sessions stats already
	err := storage.UpdateRecords(UpdateRecords[Record[schema.Schema]]{
		Saving: map[string]Record[schema.Schema]{
			"session-stats:session-2": {
				ID: "session-stats:session-2",
				Data: schema.FromGo(&tictactoemanage.SessionStatsResult{
					ID:         "session-2",
					TotalGames: 665,
					TotalDraws: 665,
					PlayerWins: map[tictactoemanage.PlayerID]int{
						"player-666": 665,
					},
				}),
				Version: 0,
			},
		},
	})
	assert.NoError(t, err)

	_, err = storage.Get("session-stats:session-2")
	assert.NoError(t, err)

	indexer := NewKeyedAggregate[tictactoemanage.State, *tictactoemanage.SessionStatsResult](
		GroupByKey,
		CombineByKey,
		storage,
	)

	repo := NewRepositoryWithIndexer[tictactoemanage.State, *tictactoemanage.SessionStatsResult](
		storage,
		indexer,
	)
	update := UpdateRecords[Record[tictactoemanage.State]]{
		Saving: map[string]Record[tictactoemanage.State]{},
	}

	for _, game := range latestGames {
		id := tictactoemanage.MustMatchState(
			game,
			func(x *tictactoemanage.SessionWaitingForPlayers) string {
				return x.SessionID
			},
			func(x *tictactoemanage.SessionReady) string {
				return x.SessionID
			},
			func(x *tictactoemanage.SessionInGame) string {
				return x.SessionID + "-" + x.GameID
			},
		)
		update.Saving["game:"+id] = Record[tictactoemanage.State]{
			ID:      "game:" + id,
			Data:    game,
			Version: 1,
		}
		update.Saving["session:"+id] = Record[tictactoemanage.State]{
			ID:      "session:" + id,
			Data:    game,
			Version: 1,
		}
	}

	err = repo.UpdateRecords(update)
	assert.NoError(t, err)
	fmt.Printf("storage: %+v \n", storage)

	indexedRepo := NewRepositoryWithIndexer[*tictactoemanage.SessionStatsResult, any](
		storage,
		NewNoopAggregator[*tictactoemanage.SessionStatsResult, any](),
	)

	result2, err := indexedRepo.Get("session-stats:session-1")
	assert.NoError(t, err)
	assert.Equal(t, Record[*tictactoemanage.SessionStatsResult]{
		ID: "session-stats:session-1",
		Data: &tictactoemanage.SessionStatsResult{
			ID:         "session-1",
			TotalGames: 3,
			TotalDraws: 1,
			PlayerWins: map[tictactoemanage.PlayerID]int{
				"player-1": 2,
			},
		},
		Version: 1,
	}, result2)

	result3, err := indexedRepo.Get("session-stats:session-2")
	assert.NoError(t, err)
	assert.Equal(t, Record[*tictactoemanage.SessionStatsResult]{
		ID: "session-stats:session-2",
		Data: &tictactoemanage.SessionStatsResult{
			ID:         "session-2",
			TotalGames: 666,
			TotalDraws: 665,
			PlayerWins: map[tictactoemanage.PlayerID]int{
				"player-666": 666,
			},
		},
		Version: 2,
	}, result3)
}
