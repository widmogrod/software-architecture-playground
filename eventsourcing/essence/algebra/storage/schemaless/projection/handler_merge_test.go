package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestMergeHandler(t *testing.T) {
	h := &MergeHandler[SessionsStats]{
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

	l := &ListAssert{t: t}
	err := h.Process(&Combine{
		Data: schema.FromGo(SessionsStats{
			Wins:  1,
			Draws: 2,
		}),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Retract: Retract{
			Data: schema.FromGo(SessionsStats{}),
		},
		Combine: Combine{
			Data: schema.FromGo(SessionsStats{
				Wins:  1,
				Draws: 2,
			}),
		},
	})
	assert.Equal(t, SessionsStats{
		Wins:  1,
		Draws: 2,
	}, h.state)
}
