package schemaless

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/typedful"
	"testing"
	"time"
)

var generateData = []Item{
	Item{
		Key: "game:1",
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			Winner:  "a",
		}),
	},
	Item{
		Key: "game:2",
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			Winner:  "b",
		}),
	},
	Item{
		Key: "game:3",
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			IsDraw:  true,
		}),
	},
}

func GenerateData() *GenerateHandler {
	return &GenerateHandler{
		load: func(returning func(message Item)) error {
			for _, msg := range generateData {
				returning(msg)
			}
			return nil
		},
	}
}

func MapGameToStats() *MapHandler[Game, SessionsStats] {
	return &MapHandler[Game, SessionsStats]{
		F: func(x Game, returning func(key string, value SessionsStats)) error {
			for _, player := range x.Players {
				wins := 0
				draws := 0
				loose := 0

				if x.IsDraw {
					draws = 1
				} else if x.Winner == player {
					wins = 1
				} else {
					loose = 1
				}

				returning("session-stats-by-player:"+player, SessionsStats{
					Wins:  wins,
					Draws: draws,
					Loose: loose,
				})
			}

			return nil
		},
	}
}

func MergeSessionStats() *MergeHandler[SessionsStats] {
	return &MergeHandler[SessionsStats]{
		Combine: func(base, x SessionsStats) (SessionsStats, error) {
			return SessionsStats{
				Wins:  base.Wins + x.Wins,
				Draws: base.Draws + x.Draws,
				Loose: base.Loose + x.Loose,
			}, nil
		},

		DoRetract: func(base, x SessionsStats) (SessionsStats, error) {
			return SessionsStats{
				Wins:  base.Wins - x.Wins,
				Draws: base.Draws - x.Draws,
				Loose: base.Loose - x.Loose,
			}, nil
		},
	}
}

func CountTotalSessionsStats(ctx Context, b Builder) Builder {
	return b.
		Map(
			ctx.Scope("Map SessionsStats -> TotalCount").WithRetracting(),
			&MapHandler[SessionsStats, int]{
				F: func(x SessionsStats, returning func(key string, value int)) error {
					returning("total", 1)
					return nil
				},
			},
		).
		Merge(
			ctx.Scope("Merge TotalCount").WithRetracting(),
			&MergeHandler[int]{
				Combine: func(base, x int) (int, error) {
					fmt.Println("counting(+)", base+x, base, x)
					return base + x, nil
				},
				DoRetract: func(base int, x int) (int, error) {
					fmt.Println("counting(-)", base+x, base, x)
					return base - x, nil
				},
			},
		)
}

func TestProjection(t *testing.T) {
	store := schemaless.NewInMemoryRepository()
	sessionStatsRepo := typedful.NewTypedRepository[SessionsStats](store)
	totalRepo := typedful.NewTypedRepository[int](store)

	root := &DefaultContext{name: "root"}

	dag := NewBuilder()
	games := dag.Load(root.Scope("GenerateData"), GenerateData())
	gameStats := games.Map(root.Scope("MapGameToStats"), MapGameToStats()) //.Map(Log("after-map"))
	gameStatsBySession := gameStats.Merge(
		root.Scope("MergeSessionStats").NoRetracting(),
		MergeSessionStats(),
	)
	//Map(Log("after-merge"))

	_ = CountTotalSessionsStats(root, gameStatsBySession).
		//Map(root.Scope("Log ⚽️TotalCount"), Log("count-total")).
		Map(
			root.Scope("Sink ⚽️TotalCount").NoRetracting(),
			NewRepositorySink("total", store))

	end := gameStatsBySession.Map(
		root.Scope("NewRepositorySink").NoRetracting(),
		NewRepositorySink("session", store))

	//expected := &Map{
	//	OnMap: Log(),
	//	Input: &Merge{
	//		OnMerge: MergeSessionStats(),
	//		Input: []DAG{
	//			&Map{
	//				OnMap: MapGameToStats(),
	//				Input: &Map{
	//					OnMap: GenerateData(),
	//					Input: nil,
	//				},
	//			},
	//		},
	//	},
	//}
	//assert.Equal(t, expected, end.Build())

	interpretation := NewInMemoryInterpreter()
	err := interpretation.Run2(end.Build2())
	//err := interpretation.Run(end.Build())
	assert.NoError(t, err)

	<-time.After(1 * time.Second)
	//interpretation.WaitUntilFinished()

	assert.Equal(t, 0, len(interpretation.errors), "interpretation should be without errors")

	result, err := sessionStatsRepo.FindingRecords(schemaless.FindingRecords[schemaless.Record[SessionsStats]]{
		RecordType: "session",
	})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 2)
	for _, x := range result.Items {
		v, err := schema.ToJSON(schema.FromGo(x.Data))
		assert.NoError(t, err)
		fmt.Printf("item: id=%s type-%s %s\n", x.ID, x.Type, string(v))
	}

	stats, err := sessionStatsRepo.Get("session-stats-by-player:a", "session")
	assert.NoError(t, err)
	assert.Equal(t, SessionsStats{
		Wins:  1,
		Loose: 1,
		Draws: 1,
	}, stats.Data)

	stats, err = sessionStatsRepo.Get("session-stats-by-player:b", "session")
	assert.NoError(t, err)
	assert.Equal(t, SessionsStats{
		Wins:  1,
		Loose: 1,
		Draws: 1,
	}, stats.Data)

	total, err := totalRepo.Get("total", "total")
	assert.NoError(t, err)
	assert.Equal(t, 2, total.Data)

}
