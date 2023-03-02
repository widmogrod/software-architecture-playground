package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func TestMergeHandler(t *testing.T) {
	h := MergeSessionStats()

	l := &ListAssert{t: t}
	err := h.Process(&Combine{
		Key: "session-stats-by-player:a",
		Data: schema.FromGo(SessionsStats{
			Wins:  1,
			Draws: 2,
		}),
	}, l.Returning)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Key: "session-stats-by-player:a",
		Retract: Retract{
			Key:  "session-stats-by-player:a",
			Data: schema.FromGo(SessionsStats{}),
		},
		Combine: Combine{
			Key: "session-stats-by-player:a",
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
