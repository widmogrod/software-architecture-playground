package tictactoe_game_server

import (
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
)

func InitEmtpy(id tictactoemanage.SessionID) tictactoemanage.SessionStatsResult {
	return tictactoemanage.SessionStatsResult{
		ID:         id,
		TotalGames: 0,
		TotalDraws: 0,
		PlayerWins: nil,
	}
}

func GroupByKey(data tictactoemanage.State) (string, tictactoemanage.SessionStatsResult) {
	return tictactoemanage.MustMatchStateR2(
		data,
		func(x *tictactoemanage.SessionWaitingForPlayers) (string, tictactoemanage.SessionStatsResult) {
			return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
		},
		func(x *tictactoemanage.SessionReady) (string, tictactoemanage.SessionStatsResult) {
			return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
		},
		func(x *tictactoemanage.SessionInGame) (string, tictactoemanage.SessionStatsResult) {
			if x.GameState == nil {
				return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
			}

			return tictacstatemachine.MustMatchStateR2(
				x.GameState,
				func(y *tictacstatemachine.GameWaitingForPlayer) (string, tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
				},
				func(y *tictacstatemachine.GameProgress) (string, tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, InitEmtpy(x.SessionID)
				},
				func(y *tictacstatemachine.GameEndWithWin) (string, tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, tictactoemanage.SessionStatsResult{
						ID:         x.SessionID,
						TotalGames: 1,
						TotalDraws: 0,
						PlayerWins: map[tictactoemanage.PlayerID]int{
							y.Winner: 1,
						},
					}
				},
				func(y *tictacstatemachine.GameEndWithDraw) (string, tictactoemanage.SessionStatsResult) {
					return "session-stats:" + x.SessionID, tictactoemanage.SessionStatsResult{
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
func CombineByKey(a, b tictactoemanage.SessionStatsResult) (tictactoemanage.SessionStatsResult, error) {
	winds := a.PlayerWins
	if winds == nil {
		winds = map[tictactoemanage.PlayerID]int{}
	}

	for k, v := range b.PlayerWins {
		winds[k] += v
	}

	return tictactoemanage.SessionStatsResult{
		ID:         a.ID,
		TotalGames: a.TotalGames + b.TotalGames,
		TotalDraws: a.TotalDraws + b.TotalDraws,
		PlayerWins: winds,
	}, nil
}

func NewTictactoeManageStateAggregate(repo storage.Repository2[schema.Schema]) *storage.KayedAggregate[tictactoemanage.State, tictactoemanage.SessionStatsResult] {
	return storage.NewKeyedAggregate[tictactoemanage.State, tictactoemanage.SessionStatsResult](
		"session-stats",
		[]string{"game"},
		GroupByKey,
		CombineByKey,
		repo,
	)
}
