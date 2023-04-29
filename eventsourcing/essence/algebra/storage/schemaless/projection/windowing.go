package projection

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"time"
)

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

func MergeWindows(x []ItemGroupedByKey, wd WindowDescription) []ItemGroupedByKey {
	return MustMatchWindowDescription(
		wd,
		func(wd *SessionWindow) []ItemGroupedByKey {
			return mergeSessionWindows(x, wd)
		},
		func(wd *SlidingWindow) []ItemGroupedByKey {
			return mergeSlidingWindows(x, wd)
		},
	)
}

func winNo(w *Window, min int64, wd *SessionWindow) int64 {
	return int64((time.Unix(0, w.Start).Sub(time.Unix(0, min)) + wd.GapDuration) / wd.GapDuration)
}

func mergeSessionWindows(x []ItemGroupedByKey, wd *SessionWindow) []ItemGroupedByKey {
	result := make([]ItemGroupedByKey, 0, len(x))
	for _, group := range x {
		var min int64
		for _, item := range group.Data {
			if min > item.Window.Start {
				min = item.Window.Start
			}
		}

		window := map[int64]*Window{}
		for _, item := range group.Data {
			// detect where in which session window item belongs
			// if in window session there are no items, then leave elemetn as is
			// when there are items, then merge them and set window to the min start and max end of elements in this window

			windowNo := winNo(item.Window, min, wd)
			if _, ok := window[windowNo]; !ok {
				window[windowNo] = &Window{
					Start: item.Window.Start,
					End:   item.Window.End,
				}
			} else {
				w := window[windowNo]
				if w.Start > item.Window.Start {
					w.Start = item.Window.Start
				}
				if w.End < item.Window.End {
					w.End = item.Window.End
				}
			}
		}

		newGroup := ItemGroupedByKey{
			Key:  group.Key,
			Data: make([]Item, 0, len(group.Data)),
		}
		for _, item := range group.Data {
			windowNo := winNo(item.Window, min, wd)
			newGroup.Data = append(newGroup.Data, Item{
				Key:       item.Key,
				Data:      item.Data,
				EventTime: item.EventTime,
				Window: &Window{
					Start: window[windowNo].Start,
					End:   window[windowNo].End,
				},
			})
		}

		result = append(result, newGroup)
	}

	return result
}

func printWindow(w *Window) {
	fmt.Printf("Window(%s, %s)\n",
		time.Unix(0, w.Start).Format("15:04"),
		time.Unix(0, w.End).Format("15:04"),
	)
}

func mergeSlidingWindows(x []ItemGroupedByKey, mwd *SlidingWindow) []ItemGroupedByKey {
	panic("implement me")
}

func DropTimestamps(x []Item) []Item {
	result := make([]Item, 0, len(x))
	for _, item := range x {
		result = append(result, Item{
			Key:       item.Key,
			Data:      item.Data,
			EventTime: 0,
			Window:    item.Window,
		})
	}
	return result
}

func GroupByKey(x []Item) []ItemGroupedByKey {
	result := make([]*ItemGroupedByKey, 0, 0)
	groups := map[string]*ItemGroupedByKey{}
	for _, item := range x {
		group, ok := groups[item.Key]
		if !ok {
			group = &ItemGroupedByKey{Key: item.Key}
			groups[item.Key] = group
			result = append(result, group)
		}
		group.Data = append(group.Data, item)
	}

	// yet another workaround for unordered maps in golang
	final := make([]ItemGroupedByKey, 0, len(result))
	for _, group := range result {
		final = append(final, *group)
	}

	return final
}

func GroupAlsoByWindow(x []ItemGroupedByKey) []ItemGroupedByWindow {
	result := make([]ItemGroupedByWindow, 0, len(x))
	windowGroups := map[int64]map[int64]*ItemGroupedByWindow{}

	for _, group := range x {
		for _, item := range group.Data {
			if _, ok := windowGroups[item.Window.Start]; !ok {
				windowGroups[item.Window.Start] = map[int64]*ItemGroupedByWindow{}
			}
			if _, ok := windowGroups[item.Window.Start][item.Window.End]; !ok {
				windowGroups[item.Window.Start][item.Window.End] = &ItemGroupedByWindow{
					Key:    group.Key,
					Data:   &schema.List{},
					Window: item.Window,
				}
			}

			windowGroups[item.Window.Start][item.Window.End].Data.Items =
				append(windowGroups[item.Window.Start][item.Window.End].Data.Items, item.Data)
		}

		// because golang maps are not ordered,
		// to create ordered result we need to iterate over data again in order to get ordered result
		for _, item := range group.Data {
			if _, ok := windowGroups[item.Window.Start]; !ok {
				continue
			}
			if _, ok := windowGroups[item.Window.Start][item.Window.End]; !ok {
				continue
			}

			result = append(result, *windowGroups[item.Window.Start][item.Window.End])
			delete(windowGroups[item.Window.Start], item.Window.End)
		}
	}

	return result
}

func ExpandToElements(x []ItemGroupedByWindow) []Item {
	result := make([]Item, 0, len(x))
	for _, group := range x {
		result = append(result, Item{
			Key:       group.Key,
			Data:      group.Data,
			EventTime: group.Window.End,
			Window:    group.Window,
		})
	}
	return result
}

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
