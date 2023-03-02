package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestMapHandler(t *testing.T) {
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
	h := &MapHandler[Game, SessionsStats]{
		onCombine: m,
		onRetract: m,
	}

	l := &ListAssert{
		t: t,
	}

	err := h.Process(&Combine{
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			Winner:  "a",
		}),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(0, &Combine{
		Data: schema.FromGo(SessionsStats{
			Wins: 1,
		}),
	})
}
