package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestMergeHandler(t *testing.T) {
	h := MergeSessionStats()

	l := &ListAssert{t: t}
	err := h.Process2(&Combine{
		Key: "session-stats-by-player:a",
		Data: schema.FromGo(SessionsStats{
			Wins:  1,
			Draws: 2,
		}),
	}, &Combine{
		Key: "session-stats-by-player:a",
		Data: schema.FromGo(SessionsStats{
			Wins:  3,
			Draws: 4,
		}),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(0, &Combine{
		Key: "session-stats-by-player:a",
		Data: schema.FromGo(SessionsStats{
			Wins:  4,
			Draws: 6,
		}),
	})
}
