package schemaless

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

func CountTotalSessionsStats(b Builder) Builder {
	return b.
		WithName("CountTotalSessionsStats:MapSessionsToStats").
		Map(&MapHandler[SessionsStats, int]{
			F: func(x SessionsStats, returning func(key string, value int)) error {
				returning("total", 1)
				return nil
			},
		}, WithRetraction()).
		WithName("CountTotalSessionsStats:Count").
		Merge(&MergeHandler[int]{
			Combine: func(base, x int) (int, error) {
				log.Debugln("counting(+)", base+x, base, x)
				return base + x, nil
			},
			DoRetract: func(base int, x int) (int, error) {
				log.Debugln("counting(-)", base+x, base, x)
				return base - x, nil
			},
		}, WithRetraction())
}

func TestProjection(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "",
		PadLevelText:    true,
	})
	store := schemaless.NewInMemoryRepository()
	sessionStatsRepo := typedful.NewTypedRepository[SessionsStats](store)
	totalRepo := typedful.NewTypedRepository[int](store)

	dag := NewBuilder()
	games := dag.
		WithName("GenerateData").
		Load(GenerateData())
	gameStats := games.
		WithName("MapGameToStats").
		Map(MapGameToStats())
	gameStatsBySession := gameStats.
		WithName("MergeSessionStats").
		Merge(MergeSessionStats(), IgnoreRetractions())

	_ = CountTotalSessionsStats(gameStatsBySession).
		WithName("Sink ⚽️TotalCount").
		Map(NewRepositorySink("total", store), IgnoreRetractions())

	end := gameStatsBySession.
		WithName("NewRepositorySink").
		Map(NewRepositorySink("session", store), IgnoreRetractions())

	interpretation := DefaultInMemoryInterpreter()
	err := interpretation.Run(end.Build())
	assert.NoError(t, err)

	<-time.After(1 * time.Second)

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
