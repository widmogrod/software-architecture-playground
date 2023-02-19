package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"testing"
)

var latestGames = []tictactoemanage.State{
	&tictactoemanage.SessionInGame{
		ID:          "session-2",
		Players:     []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:      "game-1",
		GameState:   nil,
		GameProblem: nil,
	},
	&tictactoemanage.SessionInGame{
		ID:      "session-1",
		Players: []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:  "game-2",
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
		ID:      "session-1",
		Players: []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:  "game-3",
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
		ID:      "session-1",
		Players: []tictactoemanage.PlayerID{"player-1", "player-2"},
		GameID:  "game-4",
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

func TestIndexer(t *testing.T) {

	indexer := NewIndexer[tictactoemanage.State, *tictactoemanage.SessionStatsResult]()

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

	indexer.GroupByKey(func(data tictactoemanage.State) ([]string, *tictactoemanage.SessionStatsResult) {
		return tictactoemanage.MustMatchStateR2(
			data,
			func(x *tictactoemanage.SessionWaitingForPlayers) ([]string, *tictactoemanage.SessionStatsResult) {
				return []string{"session-stats", x.ID}, InitEmtpy(x.ID)
			},
			func(x *tictactoemanage.SessionReady) ([]string, *tictactoemanage.SessionStatsResult) {
				return []string{"session-stats", x.ID}, InitEmtpy(x.ID)
			},
			func(x *tictactoemanage.SessionInGame) ([]string, *tictactoemanage.SessionStatsResult) {
				if x.GameState == nil {
					return []string{"session-stats", x.ID}, InitEmtpy(x.ID)
				}

				return tictacstatemachine.MustMatchStateR2(
					x.GameState,
					func(y *tictacstatemachine.GameWaitingForPlayer) ([]string, *tictactoemanage.SessionStatsResult) {
						return []string{"session-stats", x.ID}, InitEmtpy(x.ID)
					},
					func(y *tictacstatemachine.GameProgress) ([]string, *tictactoemanage.SessionStatsResult) {
						return []string{"session-stats", x.ID}, InitEmtpy(x.ID)
					},
					func(y *tictacstatemachine.GameEndWithWin) ([]string, *tictactoemanage.SessionStatsResult) {
						return []string{"session-stats", x.ID}, &tictactoemanage.SessionStatsResult{
							ID:         x.ID,
							TotalGames: 1,
							TotalDraws: 0,
							PlayerWins: map[tictactoemanage.PlayerID]int{
								y.Winner: 1,
							},
						}
					},
					func(y *tictacstatemachine.GameEndWithDraw) ([]string, *tictactoemanage.SessionStatsResult) {
						return []string{"session-stats", x.ID}, &tictactoemanage.SessionStatsResult{
							ID:         x.ID,
							TotalGames: 1,
							TotalDraws: 1,
							PlayerWins: nil,
						}
					},
				)
			},
		)
	})

	/*
		CombineByKey is commutative, associative, and distributive.
			commutativity = a * b = b * a
			associativity = (a * b) * c = a * (b * c)
			distributivity = a * (b + c) = (a * b) + (a * c)
	*/
	indexer.CombineByKey(func(a, b *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
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
	})

	/*
		In Contrast to CombineByKey, UncombineByKey is not commutative, and not associative.
			anticommute = a - b = -b + a
			not associative = a - (b - c) = (a - b) + c

		Example:
			3 - (2 - 1) = (3 - 2) + 1 = 2

		Some operations should marked as "substractive"
	*/
	//indexer.UncombineByKey(func(a, b *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
	//	winds := a.PlayerWins
	//	for k, v := range b.PlayerWins {
	//		winds[k] += v
	//	}
	//
	//	return &tictactoemanage.SessionStatsResult{
	//		ID:         a.ID,
	//		TotalGames: a.TotalGames - b.TotalGames,
	//		TotalDraws: a.TotalDraws - b.TotalDraws,
	//		PlayerWins: winds,
	//	}, nil
	//})

	//indexer.Init(func(groupContext GroupContext2[tictactoemanage.SessionID, tictacstatemachine.State]) {
	//	sessionId, _ := groupContext.GroupKey()
	//	//groupContext.Unpack()
	//
	//	return &tictactoemanage.SessionStatsResult{
	//		ID:         sessionId,
	//		TotalGames: 0,
	//		TotalDraws: 0,
	//		PlayerWins: map[tictactoemanage.PlayerID]int{},
	//	}
	//})

	//indexer.OnInsert(func(groupContext GroupContext2[tictactoemanage.SessionID, tictacstatemachine.State], stats *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
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
		indexer.Append(game)
	}

	result := indexer.GetKey([]string{"session-stats", "session-1"})

	assert.Equal(t, &tictactoemanage.SessionStatsResult{
		ID:         "session-1",
		TotalGames: 3,
		TotalDraws: 1,
		PlayerWins: map[tictactoemanage.PlayerID]int{
			"player-1": 2,
		},
	}, result)

	repo := NewRepositoryInMemory2[tictactoemanage.State](
		indexer,
	)
	update := UpdateRecords[Record[tictactoemanage.State]]{
		Saving: map[string]Record[tictactoemanage.State]{},
	}

	for _, game := range latestGames {
		id := tictactoemanage.MustMatchState(
			game,
			func(x *tictactoemanage.SessionWaitingForPlayers) string {
				return x.ID
			},
			func(x *tictactoemanage.SessionReady) string {
				return x.ID
			},
			func(x *tictactoemanage.SessionInGame) string {
				return x.ID
			},
		)
		update.Saving["game:"+id] = Record[tictactoemanage.State]{
			ID:      "game" + id,
			Data:    game,
			Version: 1,
		}
		update.Saving["session:"+id] = Record[tictactoemanage.State]{
			ID:      "session:" + id,
			Data:    game,
			Version: 1,
		}
	}

	err := repo.UpdateRecords(update)
	assert.NoError(t, err)
	fmt.Printf("%+v", repo.store)
}
