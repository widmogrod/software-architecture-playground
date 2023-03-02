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

type Game struct {
	Players []string
	Winner  string
	IsDraw  bool
}

type SessionsStats struct {
	Wins  int
	Draws int
}

var generateData = []Message{
	&Combine{
		Key: "game:1",
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			Winner:  "a",
		}),
	},
	&Combine{
		Key: "game:2",
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			Winner:  "b",
		}),
	},
	&Combine{
		Key: "game:3",
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			IsDraw:  true,
		}),
	},
}

func GenerateData() Handler {
	return &GenerateHandler{
		load: func(returning func(message Message) error) error {
			for _, msg := range generateData {
				if err := returning(msg); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func MapGameToStats() Handler {
	m := func(x Game) (SessionsStats, error) {
		if x.IsDraw {
			return SessionsStats{
				Draws: 1,
			}, nil
		}

		if x.Winner == "" {
			return SessionsStats{}, nil
		}

		return SessionsStats{
			Wins: 1,
		}, nil
	}

	return &MapHandler[Game, SessionsStats]{
		onCombine: m,
		onRetract: m,
	}
}

func MergeSessionStats() Handler {
	return &MergeHandler[SessionsStats]{
		state: SessionsStats{},
		onCombine: func(base, x SessionsStats) (SessionsStats, error) {
			return SessionsStats{
				Wins:  base.Wins + x.Wins,
				Draws: base.Draws + x.Draws,
			}, nil
		},
		onRetract: func(base, x SessionsStats) (SessionsStats, error) {
			return SessionsStats{
				Wins:  base.Wins - x.Wins,
				Draws: base.Draws - x.Draws,
			}, nil
		},
	}
}

func TestProjection(t *testing.T) {
	store := schemaless.NewInMemoryRepository()
	typed := typedful.NewTypedRepository[SessionsStats](store)

	dag := NewBuilder()
	games := dag.Load(GenerateData())
	gameStats := games.Map(MapGameToStats())
	gameStatsBySession := gameStats.Merge(MergeSessionStats())

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
	assert.Len(t, result.Items, 1)

	for _, x := range result.Items {
		fmt.Printf("item: %#v\n", x)
	}
}
