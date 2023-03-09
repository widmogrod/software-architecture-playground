package schemaless

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/predicate"
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
		// mkunion should be able to deduce this type
		// todo add this feature!
		schema.WhenPath([]string{"Data"}, schema.UseStruct(Game{})),
	})

	// This is example that is aiming to explore concept of live select.
	// Example use case in context of tic-tac-toe game:
	// - As a player I want to see my stats in session in real time
	// - As a player I want to see tic-tac-toe game updates in real time
	//   (I wonder how live select would compete vs current implementation on websockets)
	// - (some other) As a player I want to see achivements, in-game messages, in real time
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
	// ---
	// This example works, and there are few things to solve
	// - detect when there is no updates
	//   - let's data producer send a signal that it finished, and no frther updates will sent
	//   - add watermarking, to detect, what are latest events in system
	// - closing live select, when connection is closed
	// - optimiastion of DAGs, few edges in line, withotu forks, can be executed in memory, without need of streams between them
	// - what if different partitions needs to merge? like count total,
	//   data from different counting nodes, should be send to one selected node
	//   - How to sove such partitioning? would RabbitMQ help or make things harder?
	// - DynamoDB loader, can have information on how many RUs to use, like 5% percent
	// - when system is on production, and there will be more live select DAGs,
	//   - loading subset of records from db, may be fine for live select
	//   - but what if there will be a lot of new DAGs, that need to process all data fron whole db?
	//      my initial assumption, was that DAGs can be lightwaight, so that I can add new forks on runtime,
	//      but fork on "joined" will be from zero oldest offset, and may not have data from DB, so it's point in time
	//      maybe this means that instead of having easy way of forking, just DAGs can be deployed with full load from DB
	//      since such situation can happen multiple times, that would mean that database needs to be optimised for massive parallel reads
	//
	//      	Premature optimisation: In context of DDB, this will consume a lot of RCUs,
	//       	so that could be solved by creating a data (delta) lake on object storage like S3,
	//      	Where there is DAG that use DDB and stream to keep S3 data up to date, and always with the latest representation
	//
	//		Thinking in a way that each DAG is separate deployment, that tracks it's process
	// 		Means that change is separates, deployments can be separate, scaling needs can be separate, blast radius and ownership as well
	//      More teams can work in parallel, and with uniform language of describing DAGs, means that domain concepts can be included as library
	//
	//		From that few interesing patterns can happed, (some described in Data Architecture at Scale)
	//		- Read-only Data Stores. Sharing read RDS, each team can gen a database that other team has,
	//	      deployed to their account, and keep up to date by data system (layer)
	//		  which means, each system, can do reads as much as they can with close proximity to data (different account can be in different geo regions)
	//		  which means, each system, can share libraries that perform domain specific queries, and those libraries can use RDS in their account
	//		  which means, that those libraries, can have also catching, and catch layer can be deployed on reader account,
	//
	//	How live select architecture can be decomposed?
	//  - Fast message and reliable message delivery platform
	//  - Fast change detection
	//
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
						Version: 3,
						Data: Game{
							SessionID: "session-1",
							Players:   []string{"a", "b"},
							Winner:    "a",
						},
					}),
				})
				push(Item{
					Key: "game-2",
					Data: schema.FromGo(schemaless.Record[Game]{
						ID:      "game-2",
						Version: 3,
						Data: Game{
							SessionID: "session-2",
							Players:   []string{"a", "b"},
							Winner:    "a",
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
							SessionID: "session-1",
							Players:   []string{"a", "b"},
							Winner:    "a",
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
		Map(&FilterHandler{
			Where: predicate.MustWhere(
				"Data.SessionID = :sessionID",
				predicate.ParamBinds{
					":sessionID": schema.MkString("session-1"),
				}),
		}).
		// Joining by key and producing a new key is like merging!
		Merge(&JoinHandler[schemaless.Record[Game]]{
			F: func(a, b schemaless.Record[Game], returning func(schemaless.Record[Game])) error {
				if a.Version < b.Version {
					returning(b)
				}
				return nil
			},
		})

	gameStats := joined.
		WithName("MapGameToStats").
		Map(Log("gameStats")).
		Map(&MapHandler[schemaless.Record[Game], SessionsStats]{
			F: func(x schemaless.Record[Game], returning func(key string, value SessionsStats)) error {
				y := x.Data
				for _, player := range y.Players {
					wins := 0
					draws := 0
					loose := 0

					if y.IsDraw {
						draws = 1
					} else if y.Winner == player {
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
		})

	gameStatsBySession := gameStats.
		WithName("MergeSessionStats").
		Merge(MergeSessionStats())

	//// Storing in database those updates is like creating materialized view
	//// For live select this can be skipped.
	//store := schemaless.NewInMemoryRepository()
	//gameStatsBySession.
	//	WithName("Store in database").
	//	Map(NewRepositorySink("session", store), IgnoreRetractions())

	gameStatsBySession.
		WithName("Publish to websocket").
		Map(Log("publish-web-socket"))
	//Map(NewWebsocketSink())

	interpretation := DefaultInMemoryInterpreter()
	err := interpretation.Run(dag.Build())
	assert.NoError(t, err)
	interpretation.WaitForDone()
}
