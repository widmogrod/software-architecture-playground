package storage

//
//import (
//	"github.com/stretchr/testify/assert"
//	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/interpretation/tictactoe_game_server"
//	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
//	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
//	"testing"
//)
//
//var latestGames = []tictactoemanage.State{
//	&tictactoemanage.SessionInGame{
//		SessionID:   "session-2",
//		Players:     []tictactoemanage.PlayerID{"player-1", "player-2"},
//		GameID:      "game-2.1",
//		GameState:   nil,
//		GameProblem: nil,
//	},
//	&tictactoemanage.SessionInGame{
//		SessionID: "session-2",
//		Players:   []tictactoemanage.PlayerID{"player-1", "player-666"},
//		GameID:    "game-2.2",
//		GameState: &tictacstatemachine.GameEndWithWin{
//			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
//				FirstPlayerID:  "player-1",
//				SecondPlayerID: "player-666",
//				BoardRows:      3,
//				BoardCols:      3,
//				WinningLength:  3,
//			},
//			Winner:         "player-666",
//			WiningSequence: []tictacstatemachine.Move{"1.1", "1.2", "1.3"},
//			MovesTaken:     map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
//		},
//	},
//	&tictactoemanage.SessionInGame{
//		SessionID: "session-1",
//		Players:   []tictactoemanage.PlayerID{"player-1", "player-2"},
//		GameID:    "game-2",
//		GameState: &tictacstatemachine.GameEndWithDraw{
//			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
//				FirstPlayerID:  "player-1",
//				SecondPlayerID: "player-2",
//				BoardRows:      3,
//				BoardCols:      3,
//				WinningLength:  3,
//			},
//			MovesTaken: map[tictacstatemachine.Move]tictacstatemachine.PlayerID{
//				"1.1": "player-1",
//				"1.2": "player-2",
//			},
//		},
//		GameProblem: nil,
//	},
//	&tictactoemanage.SessionInGame{
//		SessionID: "session-1",
//		Players:   []tictactoemanage.PlayerID{"player-1", "player-2"},
//		GameID:    "game-3",
//		GameState: &tictacstatemachine.GameEndWithWin{
//			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
//				FirstPlayerID:  "player-1",
//				SecondPlayerID: "player-2",
//				BoardRows:      3,
//				BoardCols:      3,
//				WinningLength:  3,
//			},
//			Winner:         "player-1",
//			WiningSequence: []tictacstatemachine.Move{"1.1", "1.2", "1.3"},
//			MovesTaken:     map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
//		},
//		GameProblem: nil,
//	},
//	&tictactoemanage.SessionInGame{
//		SessionID: "session-1",
//		Players:   []tictactoemanage.PlayerID{"player-1", "player-2"},
//		GameID:    "game-4",
//		GameState: &tictacstatemachine.GameEndWithWin{
//			TicTacToeBaseState: tictacstatemachine.TicTacToeBaseState{
//				FirstPlayerID:  "player-1",
//				SecondPlayerID: "player-2",
//				BoardRows:      3,
//				BoardCols:      3,
//				WinningLength:  3,
//			},
//			Winner:         "player-1",
//			WiningSequence: []tictacstatemachine.Move{"1.1", "1.2", "1.3"},
//			MovesTaken:     map[tictacstatemachine.Move]tictacstatemachine.PlayerID{},
//		},
//		GameProblem: nil,
//	},
//}
//
//func TestIndexer(t *testing.T) {
//	storage := NewRepository2WithSchema()
//	indexer := tictactoe_game_server.NewTictactoeManageStateAggregate(storage)
//
//	_ = `CREATE QUERY "session-stats" ON games as g WITH
//		PRIMARY_KEY session-stats.sessionID
//		SELECT
//			sessionID
//     		, count(types.GameEndWithWin) as winds
//     		, count(types.GameEndWithDraw) as draws
//     		, winds + draws as total
//		GROUP BY g.sessionID, g.GameState FIELD_NAME ['GameEndWithWin', 'GameEndWithDraw'] as types
//		WHERE g.GameState HAS_FIELD_IN ['GameEndWithWin', 'GameEndWithDraw'];
//`
//
//	/*
//		In Contrast to CombineByKey, UncombineByKey is not commutative, and not associative.
//			anticommute = a - b = -b + a
//			not associative = a - (b - c) = (a - b) + c
//
//		Example:
//			3 - (2 - 1) = (3 - 2) + 1 = 2
//
//		Some operations should marked as "substractive"
//	*/
//	//aggregator.UncombineByKey(func(a, b *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
//	//	winds := a.PlayerWins
//	//	for k, v := range b.PlayerWins {
//	//		winds[k] += v
//	//	}
//	//
//	//	return &tictactoemanage.SessionStatsResult{
//	//		SessionID:         a.SessionID,
//	//		TotalGames: a.TotalGames - b.TotalGames,
//	//		TotalDraws: a.TotalDraws - b.TotalDraws,
//	//		PlayerWins: winds,
//	//	}, nil
//	//})
//
//	//aggregator.Init(func(groupContext GroupContext2[tictactoemanage.SessionID, tictacstatemachine.State]) {
//	//	sessionId, _ := groupContext.GroupKey()
//	//	//groupContext.Unpack()
//	//
//	//	return &tictactoemanage.SessionStatsResult{
//	//		SessionID:         sessionId,
//	//		TotalGames: 0,
//	//		TotalDraws: 0,
//	//		PlayerWins: map[tictactoemanage.PlayerID]int{},
//	//	}
//	//})
//
//	//aggregator.OnInsert(func(groupContext GroupContext2[tictactoemanage.SessionID, tictacstatemachine.State], stats *tictactoemanage.SessionStatsResult) (*tictactoemanage.SessionStatsResult, error) {
//	//	_, gamestate := groupContext.GroupKey()
//	//	return tictacstatemachine.MustMatchStateR2(
//	//		gamestate,
//	//		func(y *tictacstatemachine.GameWaitingForPlayer) (*tictactoemanage.SessionStatsResult, error) {
//	//			return stats, nil
//	//		},
//	//		func(y *tictacstatemachine.GameProgress) (*tictactoemanage.SessionStatsResult, error) {
//	//			return stats, nil
//	//		},
//	//		func(y *tictacstatemachine.GameEndWithWin) (*tictactoemanage.SessionStatsResult, error) {
//	//			stats.TotalGames += 1
//	//			stats.PlayerWins[y.Winner] += 1
//	//			return stats, nil
//	//		},
//	//		func(y *tictacstatemachine.GameEndWithDraw) (*tictactoemanage.SessionStatsResult, error) {
//	//			stats.TotalGames += 1
//	//			stats.TotalDraws += 1
//	//			return stats, nil
//	//		},
//	//	)
//	//})
//
//	for _, game := range latestGames {
//		err := indexer.Append(Record[tictactoemanage.State]{
//			Type: "game",
//			Data: game,
//		})
//		assert.NoError(t, err)
//	}
//
//	result := indexer.GetIndexByKey("session-stats:session-1")
//
//	assert.Equal(t, tictactoemanage.SessionStatsResult{
//		ID:         "session-1",
//		TotalGames: 3,
//		TotalDraws: 1,
//		PlayerWins: map[tictactoemanage.PlayerID]int{
//			"player-1": 2,
//		},
//	}, result)
//}
