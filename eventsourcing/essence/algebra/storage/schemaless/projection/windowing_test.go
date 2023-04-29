package projection

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func withTime(hour, minute int) int64 {
	return time.
		Date(2019, 1, 1, hour, minute, 0, 0, time.UTC).
		UnixNano()
}

func TestWindowing(t *testing.T) {
	list := []Item{
		{
			Key:       "a",
			Data:      nil,
			EventTime: withTime(10, 2),
		},
	}

	t.Run("assign session windows", func(t *testing.T) {
		result := AssignWindows(list, &SessionWindow{
			GapDuration: 30 * time.Minute,
		})
		expected := []Item{
			{
				Key:       "a",
				Data:      nil,
				EventTime: withTime(10, 2),
				Window: &Window{
					Start: withTime(10, 2),
					End:   withTime(10, 32),
				},
			},
		}

		//rules := WindowBuilder{}
		//rules.Trigger(&AtPeriod{Duration: 1 * time.Second})
		//
		//result := Process(list, rules)
		//expected := []Item{
		//	{"a", nil, 0, 0},
		//}

		assert.Equal(t, expected, result)
	})
	t.Run("assign sliding windows", func(t *testing.T) {
		result := AssignWindows(list, &SlidingWindow{
			Width:  2 * time.Minute,
			Period: 1 * time.Minute,
		})
		expected := []Item{
			{
				Key:       "a",
				Data:      nil,
				EventTime: withTime(10, 2),
				Window: &Window{
					Start: withTime(10, 1),
					End:   withTime(10, 3),
				},
			},
			{
				Key:       "a",
				Data:      nil,
				EventTime: withTime(10, 2),
				Window: &Window{
					Start: withTime(10, 2),
					End:   withTime(10, 4),
				},
			},
		}

		assert.Len(t, result, 2)
		assert.Equal(t, expected, result)
	})

}
