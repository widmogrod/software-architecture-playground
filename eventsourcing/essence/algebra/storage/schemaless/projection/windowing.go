package projection

import "time"

func AssignWindows(x []Item, wd WindowDescription) []Item {
	return MustMatchWindowDescription(
		wd,
		func(wd *SessionWindow) []Item {
			return assignSessionWindows(x, wd)
		},
		func(wd *SlidingWindow) []Item {
			return assignSlidingWindows(x, wd)
		},
	)
}

func assignSlidingWindows(x []Item, wd *SlidingWindow) []Item {
	result := make([]Item, 0, len(x))
	for _, item := range x {
		eventTime := time.Unix(0, item.EventTime)
		// slicing window is [start, end)
		// left side inclusive, and right side exclusive,
		// so we need to add 1 period to the end
		for start := eventTime.Add(-wd.Width).Add(wd.Period); start.UnixNano() <= item.EventTime; start = start.Add(wd.Period) {
			result = append(result, Item{
				Key:       item.Key,
				Data:      item.Data,
				EventTime: item.EventTime,
				Window: &Window{
					Start: start.UnixNano(),
					End:   start.Add(wd.Width).UnixNano(),
				},
			})
		}
	}
	return result
}

func assignSessionWindows(x []Item, wd *SessionWindow) []Item {
	result := make([]Item, 0, len(x))
	for _, item := range x {
		result = append(result, Item{
			Key:       item.Key,
			Data:      item.Data,
			EventTime: item.EventTime,
			Window: &Window{
				Start: item.EventTime,
				End:   time.Unix(0, item.EventTime).Add(wd.GapDuration).UnixNano(),
			},
		})
	}
	return result
}

func Sliding()           {}
func Sessions()          {}
func DropTimestamps()    {}
func GroupByKey()        {}
func GroupAlsoByWindow() {}
func ExpandToElements()  {}

//go:generate mkunion -name=WindowDescription
type (
	SessionWindow struct {
		GapDuration time.Duration
	}
	SlidingWindow struct {
		Width  time.Duration
		Period time.Duration
	}
	//FixedWindow struct{}
)

//go:generate mkunion -name=TriggerDescription
type (
	AtPeriod struct {
		Duration time.Duration
	}
	AtCount     struct{}
	AtWatermark struct{}

	SequenceOf  struct{}
	RepeatUntil struct{}
)

type WindowBuilder struct{}

func (WindowBuilder) Trigger(td TriggerDescription) {}
