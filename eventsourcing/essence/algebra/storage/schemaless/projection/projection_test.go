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
		//onRetract: func(base, x SessionsStats) (SessionsStats, error) {
		//	return SessionsStats{
		//		Wins:  base.Wins - x.Wins,
		//		Draws: base.Draws - x.Draws,
		//	}, nil
		//},
	}
}

func CountTotalSessionsStats(b Builder) Builder {
	return b.
		Map(&MapHandler[SessionsStats, int]{
			F: func(x SessionsStats, returning func(key string, value int)) error {
				returning("total", 1)
				return nil
			},
		}).
		Merge(&MergeHandler[int]{
			Combine: func(base, x int) (int, error) {
				return base + x, nil
			},
		})
}

func TestProjection(t *testing.T) {
	store := schemaless.NewInMemoryRepository()
	typed := typedful.NewTypedRepository[SessionsStats](store)

	dag := NewBuilder()
	games := dag.Load(GenerateData())
	gameStats := games.Map(MapGameToStats()).Map(Log("after-map"))
	gameStatsBySession := gameStats.Merge(MergeSessionStats()).Map(Log("after-merge"))

	_ = CountTotalSessionsStats(gameStatsBySession)

	end := gameStatsBySession.Map(NewRepositorySink("session", store))
	// .Map(Log())

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
	err := interpretation.Run(end.Build())
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, len(interpretation.errors), "interpretation should be without errors")

	result, err := typed.FindingRecords(schemaless.FindingRecords[schemaless.Record[SessionsStats]]{})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 2)

	stats, err := typed.Get("session-stats-by-player:a", "session-stats-by-player")
	assert.NoError(t, err)
	assert.Equal(t, SessionsStats{
		Wins:  1,
		Loose: 1,
		Draws: 1,
	}, stats.Data)

	stats, err = typed.Get("session-stats-by-player:b", "session-stats-by-player")
	assert.NoError(t, err)
	assert.Equal(t, SessionsStats{
		Wins:  1,
		Loose: 1,
		Draws: 1,
	}, stats.Data)

	for _, x := range result.Items {
		v, err := schema.ToJSON(schema.FromGo(x.Data))
		assert.NoError(t, err)
		fmt.Printf("item: id=%s type-%s %s\n", x.ID, x.Type, string(v))
	}
}
