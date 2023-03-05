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
			panic("retraction on SessionStats should not happen")
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
		Merge(MergeSessionStats())

	_ = CountTotalSessionsStats(gameStatsBySession).
		WithName("Sink ⚽️TotalCount").
		Map(NewRepositorySink("total", store))

	_ = gameStatsBySession.
		WithName("NewRepositorySink").
		Map(NewRepositorySink("session", store))

	interpretation := DefaultInMemoryInterpreter()
	err := interpretation.Run(dag.Build())
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

func TestLiveSelect(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "",
		PadLevelText:    true,
	})

	// setup type registry
	schema.RegisterRules([]schema.RuleMatcher{
		schema.WhenPath(nil, schema.UseStruct(&schemaless.Record[Game]{})),
		schema.WhenPath([]string{"Data"}, schema.UseStruct(Game{})),
	})

	// This is example that is aiming to explore concept of live select.
	// Example use case in context of tic-tac-toe game:
	// - As a player I want to see my stats in session in real time
	// - As a player I want to see tic-tac-toe game updates in real time
	//   (I wonder how live select would compete vs current implementation on websockets)
	//
	// 	LIVE SELECT
	//      sessionID,
	//		COUNT_BY_VALUE(winner, WHERE winnerID NOT NULL) as wins, // {"a": 1, "b": 2}
	//		COUNT(WHERE isDraw = TRUE) as draws,
	//      COUNT() as total
	//  GROUP BY sessionID as group
	//  WHERE sessionID = :sessionID
	//    AND gameState = "GameFinished"
	//
	// Solving live select model with DAG, can solve also MATERIALIZED VIEW problem with ease.
	//
	// At some point it would be nice to have benchmarks that show where is breaking point of
	// doing ad hock live selects vs precalculating materialized view.
	dag := NewBuilder()
	// Only latest records from database that match live select criteria are used
	lastState := dag.
		WithName("DynamoDB LastState Filtered").
		Load(&GenerateHandler{
			load: func(push func(message Item)) error {
				push(Item{
					Key: "game-1",
					Data: schema.FromGo(schemaless.Record[Game]{
						ID:      "game-1",
						Version: 1,
						Data: Game{
							Players: []string{"a", "b"},
							Winner:  "a",
						},
					}),
				})

				return nil
			},
		})
	// Only streamed records that match live select criteria are used
	streamState := dag.
		WithName("DynamoDB Filtered Stream").
		Load(&GenerateHandler{
			load: func(push func(message Item)) error {
				// This is where we would get data from stream
				push(Item{
					Key: "game-1",
					Data: schema.FromGo(schemaless.Record[Game]{
						ID:      "game-1",
						Version: 2,
						Data: Game{
							Players: []string{"a", "b"},
							Winner:  "a",
						},
					}),
				})
				return nil
			},
		})
	// Joining make sure that newest version is published

	joined := dag.
		WithName("Join").
		// Join by key, so if db and stream has the same key, then it will be joined.
		Join(lastState, streamState).
		// Joining by key and producing a new key is like merging!
		Merge(&MergeHandler[schemaless.Record[Game]]{
			Combine: func(a, b schemaless.Record[Game]) (schemaless.Record[Game], error) {
				if a.Version > b.Version {
					return a, nil
				} else if a.Version < b.Version {
					return b, nil
				} else {
					return a, nil
				}
			},
			DoRetract: nil,
		})

	gameStats := joined.
		WithName("MapGameToStats").
		Map(Log("gameStats"))
	//Map(MapGameToStats())

	_ = gameStats

	//gameStatsBySession := gameStats.
	//	WithName("MergeSessionStats").
	//	Merge(MergeSessionStats())

	//// Storing in database those updates is like creating materialized view
	//// For live select this can be skipped.
	//store := schemaless.NewInMemoryRepository()
	//gameStatsBySession.
	//	WithName("Store in database").
	//	Map(NewRepositorySink("session", store), IgnoreRetractions())

	//gameStatsBySession.
	//	WithName("Publish to websocket").
	//	Map(NewWebsocketSink())

	interpretation := DefaultInMemoryInterpreter()
	err := interpretation.Run(dag.Build())
	assert.NoError(t, err)

	<-time.After(1 * time.Second)
}
