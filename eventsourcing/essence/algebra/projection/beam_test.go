package projection

//import apache beam for go and process data in a streaming fashion

import (
	"context"
	"fmt"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/runners/direct"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictacstatemachine"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/usecase/tictactoemanage"
	"testing"
)

func TestProcess(t *testing.T) {
	beam.Init()
	p, root := beam.NewPipelineWithRoot()

	enchanced := beam.ParDo(root, func(x []byte, emit func(string)) {
		emit(string(x) + "word1")
		emit(string(x) + "word2")
	}, beam.Impulse(root))

	beam.ParDo0(root, func(x string) {
		t.Log(x)
	}, enchanced)

	res, err := direct.Execute(context.Background(), p)
	if assert.NoError(t, err) {
		t.Logf("%v", res.Metrics())
	}
}

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

func TestLatestGames(t *testing.T) {
	beam.Init()
	p, root := beam.NewPipelineWithRoot()

	//enchanced := beam.Create(root, latestGames...)
	enchanced := beam.ParDo(root.Scope(".Generate data"), func(x []byte, emit func([]byte)) error {
		for _, game := range latestGames {
			gameSchema := schema.FromGo(game)
			bytes, err := schema.ToJSON(gameSchema)
			//return fmt.Errorf("error while marshaling game to json")
			if err != nil {
				return err
			}
			emit(bytes)
		}
		return nil
	}, beam.Impulse(root))

	bySession := beam.ParDo(root.Scope(".BySession"), func(x []byte) (string, []byte) {
		xSchema, err := schema.FromJSON(x)
		if err != nil {
			panic(err)
		}

		xValue := schema.ToGo(xSchema)
		state, ok := xValue.(tictactoemanage.State)
		if !ok {
			panic(fmt.Errorf("error while converting to tictactoemanage.State"))
		}

		return tictactoemanage.MustMatchStateR2(
			state,
			func(y *tictactoemanage.SessionWaitingForPlayers) (string, []byte) {
				return y.ID, x
			},
			func(y *tictactoemanage.SessionReady) (string, []byte) {
				return y.ID, x
			},
			func(y *tictactoemanage.SessionInGame) (string, []byte) {
				return y.ID, x
			})
	}, enchanced)

	groupped := beam.GroupByKey(root.Scope(".Group"), bySession)

	beam.ParDo0(root.Scope(".Log"), func(key string, values func(x *[]byte) bool) {
		t.Logf("Key: %s \n", key)
		var r []byte
		for values(&r) {
			t.Logf("\tElement: \n\t %s\n", string(r))
		}
	}, groupped)

	res, err := direct.Execute(context.Background(), p)
	if assert.NoError(t, err) {
		t.Logf("%v", res.Metrics())
	}
}

func TestStatsInMemory(t *testing.T) {
	result := tictactoemanage.SessionStatsResult{
		ID:         "",
		TotalGames: 0,
		TotalDraws: 0,
		PlayerWins: nil,
	}

	for _, game := range latestGames {
		inProgress, ok := game.(*tictactoemanage.SessionInGame)
		if !ok {
			continue
		}
		result.ID = inProgress.ID
		result.TotalGames++
		if _, ok := inProgress.GameState.(*tictacstatemachine.GameEndWithDraw); ok {
			result.TotalDraws++
		} else if win, ok := inProgress.GameState.(*tictacstatemachine.GameEndWithWin); ok {
			if result.PlayerWins == nil {
				result.PlayerWins = make(map[tictactoemanage.PlayerID]float64)
			}
			result.PlayerWins[win.Winner]++
		}
	}

	assert.Equal(t, tictactoemanage.SessionStatsResult{
		ID:         "session-1",
		TotalGames: 4, // TODO: 3
		TotalDraws: 1,
		PlayerWins: map[tictactoemanage.PlayerID]float64{
			"player-1": 2,
		},
	}, result)
}

func TestStatsInBeam(t *testing.T) {
	beam.Init()
	p, root := beam.NewPipelineWithRoot()

	//enchanced := beam.Create(root, latestGames...)
	enchanced := beam.ParDo(root.Scope(".Generate data"), func(x []byte, emit func([]byte)) error {
		for _, game := range latestGames {
			gameSchema := schema.FromGo(game)
			bytes, err := schema.ToJSON(gameSchema)
			//return fmt.Errorf("error while marshaling game to json")
			if err != nil {
				return err
			}
			emit(bytes)
		}
		return nil
	}, beam.Impulse(root))

	bySession := beam.ParDo(root.Scope(".BySession"), func(x []byte, emit func(string, []byte)) error {
		xSchema, err := schema.FromJSON(x)
		if err != nil {
			return err
		}

		xValue := schema.ToGo(xSchema)

		inProgress, ok := xValue.(*tictactoemanage.SessionInGame)
		if !ok {
			return fmt.Errorf("error while converting to tictactoemanage.SessionInGame")
		}
		result := tictactoemanage.SessionStatsResult{
			ID:         inProgress.ID,
			TotalGames: 1,
		}
		if _, ok := inProgress.GameState.(*tictacstatemachine.GameEndWithDraw); ok {
			result.TotalDraws++
		} else if win, ok := inProgress.GameState.(*tictacstatemachine.GameEndWithWin); ok {
			if result.PlayerWins == nil {
				result.PlayerWins = make(map[tictactoemanage.PlayerID]float64)
			}
			result.PlayerWins[win.Winner]++
		}

		data, err := schema.ToJSON(schema.FromGo(result))
		if err != nil {
			return err
		}

		emit(inProgress.ID, data)

		return nil
	}, enchanced)

	statsBySession := beam.CombinePerKey(root.Scope(".Combine"), func(x []byte, y []byte) []byte {
		xSch, err := schema.FromJSON(x)
		if err != nil {
			panic(err)
		}

		ySch, err := schema.FromJSON(y)
		if err != nil {
			panic(err)
		}

		xVal := schema.ToGo(xSch, schema.WhenPath(nil, schema.UseStruct(&tictactoemanage.SessionStatsResult{})))
		yVal := schema.ToGo(ySch, schema.WhenPath(nil, schema.UseStruct(&tictactoemanage.SessionStatsResult{})))

		wins := xVal.(*tictactoemanage.SessionStatsResult).PlayerWins
		if wins == nil {
			wins = make(map[tictactoemanage.PlayerID]float64)
		}

		for k, v := range yVal.(*tictactoemanage.SessionStatsResult).PlayerWins {
			wins[k] = v
		}

		result := tictactoemanage.SessionStatsResult{
			ID:         xVal.(*tictactoemanage.SessionStatsResult).ID,
			TotalGames: xVal.(*tictactoemanage.SessionStatsResult).TotalGames + yVal.(*tictactoemanage.SessionStatsResult).TotalGames,
			TotalDraws: xVal.(*tictactoemanage.SessionStatsResult).TotalDraws + yVal.(*tictactoemanage.SessionStatsResult).TotalDraws,
			PlayerWins: wins,
		}

		data, err := schema.ToJSON(schema.FromGo(result))
		if err != nil {
			panic(err)
		}

		return data
	}, bySession)

	beam.ParDo0(root.Scope(".Log"), func(key string, x []byte) {
		t.Logf("Key: %s \n", key)
		t.Logf("\tElement: \n\t %s\n", string(x))
	}, statsBySession)

	res, err := direct.Execute(context.Background(), p)
	if assert.NoError(t, err) {
		t.Logf("%v", res.Metrics())
	}
}

//func TestStatsInBeamSeparateFields(t *testing.T) {
//	beam.Init()
//	p, root := beam.NewPipelineWithRoot()
//
//	//enchanced := beam.Create(root, latestGames...)
//	enchanced := beam.ParDo(root.Scope(".Generate data"), func(x []byte, emit func([]byte)) error {
//		for _, game := range latestGames {
//			gameSchema := schema.FromGo(game)
//			bytes, err := schema.ToJSON(gameSchema)
//			//return fmt.Errorf("error while marshaling game to json")
//			if err != nil {
//				return err
//			}
//			emit(bytes)
//		}
//		return nil
//	}, beam.Impulse(root))
//
//	drawBySession, winnerBySession := beam.ParDo2(
//		root.Scope(".BySession"),
//		func(
//			x []byte,
//			emitDraw func(string, int),
//			emitWin func(string, map[tictactoemanage.PlayerID]float64),
//		) {
//			xSchema, err := schema.FromJSON(x)
//			if err != nil {
//				panic(err)
//			}
//
//			xValue := schema.ToGo(xSchema)
//
//			inProgress, ok := xValue.(*tictactoemanage.SessionInGame)
//			if !ok {
//				panic(fmt.Errorf("error while converting to tictactoemanage.SessionInGame"))
//			}
//
//			if _, ok := inProgress.GameState.(*tictacstatemachine.GameEndWithDraw); ok {
//				emitDraw(inProgress.ID, 1)
//			} else if win, ok := inProgress.GameState.(*tictacstatemachine.GameEndWithWin); ok {
//				emitWin(inProgress.ID, map[tictactoemanage.PlayerID]float64{win.Winner: 1})
//			}
//
//		}, enchanced)
//
//	allDrawsBySession := beam.CombinePerKey(root.Scope(".CombineWind"), func(x, y int) int {
//		return x + y
//	}, drawBySession)
//
//	allWinersBySession := beam.CombinePerKey(
//		root.Scope(".CombineWind"),
//		func(x, y map[tictactoemanage.PlayerID]float64) map[tictactoemanage.PlayerID]float64 {
//			result := make(map[tictactoemanage.PlayerID]float64)
//			for k, v := range x {
//				result[k] += v
//			}
//			for k, v := range y {
//				result[k] += v
//			}
//			return result
//		}, winnerBySession)
//
//	rr := beam.CoGroupByKey(root.Scope(".CoGroupByKey"), allDrawsBySession, allWinersBySession)
//
//	statsBySession := beam.ParDo(
//		root.Scope(".Stats"),
//		func(
//			// CoGBK<string,int,map[string]float64>
//			session string,
//			drawsI func(*int) bool,
//			winsI func(*map[tictactoemanage.PlayerID]float64) bool,
//			//iterate func(
//			//session *string,
//			//draws int,
//			//winners *map[tictactoemanage.PlayerID]float64,
//			//),
//		) []byte {
//			//var session string
//			var draws int
//			var wins map[tictacstatemachine.PlayerID]float64
//
//			//for iterate(&session, &draws, &wins) {
//			//
//			//}
//
//			result := tictactoemanage.SessionStatsResult{
//				TotalGames: draws + len(wins),
//				TotalDraws: draws,
//				PlayerWins: wins,
//			}
//
//			data, err := schema.ToJSON(schema.FromGo(result))
//			if err != nil {
//				panic(err)
//			}
//
//			return data
//		}, rr)
//	//}, allDrawsBySession, beam.SideInput{Input: allWinersBySession})
//
//	beam.ParDo0(root.Scope(".Log"), func(key string, x []byte) {
//		t.Logf("Key: %s \n", key)
//		t.Logf("\tElement: \n\t %s\n", string(x))
//	}, statsBySession)
//
//	err := beamx.Run(context.Background(), p)
//	assert.NoError(t, err)
//}
