package projection

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
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

func TestMergeWindows(t *testing.T) {
	list := []Item{
		{
			Key:       "k1",
			Data:      schema.MkString("v1"),
			EventTime: withTime(13, 2),
		},
		{
			Key:       "k2",
			Data:      schema.MkString("v2"),
			EventTime: withTime(13, 14),
		},
		{
			Key:       "k1",
			Data:      schema.MkString("v3"),
			EventTime: withTime(13, 57),
		},
		{
			Key:       "k1",
			Data:      schema.MkString("v4"),
			EventTime: withTime(13, 20),
		},
	}

	list2 := AssignWindows(list, &SessionWindow{
		GapDuration: 30 * time.Minute,
	})
	assert.Equal(t, []Item{
		{
			Key:       "k1",
			Data:      schema.MkString("v1"),
			EventTime: withTime(13, 2),
			Window: &Window{
				Start: withTime(13, 2),
				End:   withTime(13, 32),
			},
		},
		{
			Key:       "k2",
			Data:      schema.MkString("v2"),
			EventTime: withTime(13, 14),
			Window: &Window{
				Start: withTime(13, 14),
				End:   withTime(13, 44),
			},
		},
		{
			Key:       "k1",
			Data:      schema.MkString("v3"),
			EventTime: withTime(13, 57),
			Window: &Window{
				Start: withTime(13, 57),
				End:   withTime(14, 27),
			},
		},
		{
			Key:       "k1",
			Data:      schema.MkString("v4"),
			EventTime: withTime(13, 20),
			Window: &Window{
				Start: withTime(13, 20),
				End:   withTime(13, 50),
			},
		},
	}, list2, "AssignWindows")

	list3 := DropTimestamps(list2)
	assert.Equal(t, []Item{
		{
			Key:       "k1",
			Data:      schema.MkString("v1"),
			EventTime: 0,
			Window: &Window{
				Start: withTime(13, 2),
				End:   withTime(13, 32),
			},
		},
		{
			Key:       "k2",
			Data:      schema.MkString("v2"),
			EventTime: 0,
			Window: &Window{
				Start: withTime(13, 14),
				End:   withTime(13, 44),
			},
		},
		{
			Key:       "k1",
			Data:      schema.MkString("v3"),
			EventTime: 0,
			Window: &Window{
				Start: withTime(13, 57),
				End:   withTime(14, 27),
			},
		},
		{
			Key:       "k1",
			Data:      schema.MkString("v4"),
			EventTime: 0,
			Window: &Window{
				Start: withTime(13, 20),
				End:   withTime(13, 50),
			},
		},
	}, list3, "DropTimestamps")

	list4 := GroupByKey(list3)
	assert.Equal(t, []ItemGroupedByKey{
		{
			Key: "k1",
			Data: []Item{
				{
					Key:       "k1",
					Data:      schema.MkString("v1"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 2),
						End:   withTime(13, 32),
					},
				},
				{
					Key:       "k1",
					Data:      schema.MkString("v3"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 57),
						End:   withTime(14, 27),
					},
				},
				{
					Key:       "k1",
					Data:      schema.MkString("v4"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 20),
						End:   withTime(13, 50),
					},
				},
			},
		},
		{
			Key: "k2",
			Data: []Item{
				{
					Key:       "k2",
					Data:      schema.MkString("v2"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 14),
						End:   withTime(13, 44),
					},
				},
			},
		},
	}, list4, "GroupByKey")

	list5 := MergeWindows(list4, &SessionWindow{
		GapDuration: 30 * time.Minute,
	})

	assert.Equal(t, []ItemGroupedByKey{
		{
			Key: "k1",
			Data: []Item{
				{
					Key:       "k1",
					Data:      schema.MkString("v1"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 2),
						End:   withTime(13, 50),
					},
				},
				{
					Key:       "k1",
					Data:      schema.MkString("v3"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 57),
						End:   withTime(14, 27),
					},
				},
				{
					Key:       "k1",
					Data:      schema.MkString("v4"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 2),
						End:   withTime(13, 50),
					},
				},
			},
		},
		{
			Key: "k2",
			Data: []Item{
				{
					Key:       "k2",
					Data:      schema.MkString("v2"),
					EventTime: 0,
					Window: &Window{
						Start: withTime(13, 14),
						End:   withTime(13, 44),
					},
				},
			},
		},
	}, list5, "MergeWindows")

	list6 := GroupAlsoByWindow(list5)
	assert.Equal(t, []ItemGroupedByWindow{
		{
			Key:  "k1",
			Data: schema.MkList(schema.MkString("v1"), schema.MkString("v4")),
			Window: &Window{
				Start: withTime(13, 2),
				End:   withTime(13, 50),
			},
		},
		{
			Key:  "k1",
			Data: schema.MkList(schema.MkString("v3")),
			Window: &Window{
				Start: withTime(13, 57),
				End:   withTime(14, 27),
			},
		},
		{
			Key:  "k2",
			Data: schema.MkList(schema.MkString("v2")),
			Window: &Window{
				Start: withTime(13, 14),
				End:   withTime(13, 44),
			},
		},
	}, list6, "GroupAlsoByWindow")

	list7 := ExpandToElements(list6)
	assert.Equal(t, []Item{
		{
			Key:       "k1",
			Data:      schema.MkList(schema.MkString("v1"), schema.MkString("v4")),
			EventTime: withTime(13, 50),
			Window: &Window{
				Start: withTime(13, 2),
				End:   withTime(13, 50),
			},
		},
		{
			Key:       "k1",
			Data:      schema.MkList(schema.MkString("v3")),
			EventTime: withTime(14, 27),
			Window: &Window{
				Start: withTime(13, 57),
				End:   withTime(14, 27),
			},
		},
		{
			Key:       "k2",
			Data:      schema.MkList(schema.MkString("v2")),
			EventTime: withTime(13, 44),
			Window: &Window{
				Start: withTime(13, 14),
				End:   withTime(13, 44),
			},
		},
	}, list7)

}
