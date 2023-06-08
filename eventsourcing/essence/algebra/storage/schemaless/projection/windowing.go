package projection

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"time"
)

//go:generate mkunion -name=WindowDescription
type (
	SessionWindow struct {
		GapDuration time.Duration
	}
	SlidingWindow struct {
		Width  time.Duration
		Period time.Duration
	}
	FixedWindow struct {
		Width time.Duration
	}
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
		func(wd *FixedWindow) []Item {
			return assignFixedWindows(x, wd)
		},
	)
}

func assignFixedWindows(x []Item, wd *FixedWindow) []Item {
	result := make([]Item, 0, len(x))
	for _, item := range x {
		start := item.EventTime - item.EventTime%wd.Width.Nanoseconds()
		end := start + wd.Width.Nanoseconds()
		result = append(result, Item{
			Key:       item.Key,
			Data:      item.Data,
			EventTime: item.EventTime,
			Window: &Window{
				Start: start,
				End:   end,
			},
		})
	}
	return result
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
			// assumption here is that before calling MergeWindows,
			// items got assigned window using the same WindowDefinition,
			// so we can assume that all items in the group have the same value for sliding & fixed windows
			// that don't need to be adjusted, like in session windows
			return x
		},
		func(wd *FixedWindow) []ItemGroupedByKey {
			// assumption here is that before calling MergeWindows,
			// items got assigned window using the same WindowDefinition,
			// so we can assume that all items in the group have the same value for sliding & fixed windows
			// that don't need to be adjusted, like in session windows
			return x
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
			// if in window session there are no items, then leave element as is
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
		result = append(result, ToElement(&group))
	}
	return result
}

func ToElement(group *ItemGroupedByWindow) Item {
	return Item{
		Key:       group.Key,
		Data:      group.Data,
		EventTime: group.Window.End,
		Window:    group.Window,
	}
}

func WindowKey(window *Window) string {
	return fmt.Sprintf("%d.%d", window.Start, window.End)
}

func ItemKeyWindow(x Item) string {
	return fmt.Sprintf("%s:%s", x.Key, WindowKey(x.Window))
}

func NewWindowBuffer(wd WindowDescription) *WindowBuffer {
	return &WindowBuffer{
		wd: wd,

		windowGroups: map[int64]map[int64]*ItemGroupedByWindow{},
	}
}

type WindowBuffer struct {
	wd WindowDescription
	fm WindowFlushMode

	// Windows groups that haven't been flushed yet
	windowGroups map[int64]map[int64]*ItemGroupedByWindow
	// Window groups that have been flushed, and because flush mode is accumulating or accumulatingAndRetracting
	//windowGroupsLast map[int64]map[int64]*ItemGroupedByWindow
}

func (w *WindowBuffer) Append(x Item) {
	list1 := AssignWindows([]Item{x}, w.wd)
	list2 := DropTimestamps(list1)
	list3 := GroupByKey(list2)
	list4 := MergeWindows(list3, w.wd)
	w.GroupAlsoByWindow(list4)
}

// FlushItemGroupedByWindow makes sure that windows that needs to be flushed are delivered to the function f.
//
// Some operations that require aggregate or aggregateAndRetract are not expressed by window buffer,
// but by the function f that is responsible for grouping windows and knowing whenever value of window was calculated,
// and aggregation can add previous value, or retract previous value.
//
// Snapshotting process works as follows:
// - store information about last message that was successfully processed
// - store outbox of windows that were successfully processed and need to be flushed
//
// When process is restarted, it will:
// - restore information about last message that was successfully processed, and ask runtime to continue sending messages from that point
// - start emptying outbox
//
// Flush process works as follows:
// - for each window in outbox, call flush function, that function needs to return OK or error
// - if flush function returns OK, remove window from outbox
// - if flush function returns error, stop flushing, and retry on next flush
//
// Because each of the processes is independent by key, we can retry flushing only for windows that failed to flush.
// Because each of outbox pattern, we have order of windows guaranteed.
// _
// Because we make failure first class citizen, client can define failure stream and decide that after N retries,
// message should be sent to dead letter queue, or other error handling mechanism.
//
// Because we can model backfilling as a failure, we can use same mechanism to backfill windows that failed to flush,
// in the same way as we would backfill normal messages from time window
//
// Backfill is the same as using already existing DAG, but only with different input.
//func (w *WindowBuffer) FlushItemGroupedByWindow(shouldFlush func(group *ItemGroupedByWindow) bool, returning func(Item)) {
//	for _, windowGroups := range w.windowGroups {
//		for _, group := range windowGroups {
//			// flushing depends on type of window flush mode:
//			// - aggregate - computes the result of the window reusing the previous result
//			// - discard - discards the previous result and computes the result of the window from scratch
//			// - aggregate and retract - computes the result of the window reusing the previous result and removes the previous result from the state
//			MustMatchWindowFlushModeR0(
//				w.fm,
//				func(x *Accumulate) {
//					// if it's first flush, create map
//					//if _, ok := w.windowGroupsLast[group.Window.Start]; !ok {
//					//	w.windowGroupsLast[group.Window.Start] = map[int64]*ItemGroupedByWindow{}
//					//}
//					//
//					//// create last group, or merge results with previous group
//					//lastGroup, ok := w.windowGroupsLast[group.Window.Start][group.Window.End]
//					//if !ok {
//					//	lastGroup = &ItemGroupedByWindow{
//					//		Key:    group.Key,
//					//		Window: group.Window,
//					//		Data: &schema.List{
//					//			Items: group.Data.Items,
//					//		},
//					//	}
//					//} else {
//					//	lastGroup.Data.Items = append(
//					//		lastGroup.Data.Items,
//					//		group.Data.Items...,
//					//	)
//					//}
//
//					if !shouldFlush(group) {
//						return
//					}
//
//					returning(ToElement(*group))
//
//					// set last group
//					//w.windowGroupsLast[group.Window.Start][group.Window.End] = lastGroup
//
//					// since we have the last group, we can remove current group
//					// this will be useful, when late arrivals will pass grace period
//					w.RemoveItemGropedByWindow(group)
//				},
//				func(x *Discard) {
//					if !shouldFlush(group) {
//						return
//					}
//					returning(ToElement(*group))
//					w.RemoveItemGropedByWindow(group)
//				},
//				func(x *AccumulatingAndRetracting) {
//					// if it's first flush, create map
//					//if _, ok := w.windowGroupsLast[group.Window.Start]; !ok {
//					//	w.windowGroupsLast[group.Window.Start] = map[int64]*ItemGroupedByWindow{}
//					//}
//					//
//					//if !shouldFlush(group) {
//					//	return
//					//}
//					//
//					//// retract previous result
//					//if lastGroup, ok := w.windowGroupsLast[group.Window.Start][group.Window.End]; ok {
//					//	returning(ToElement(*lastGroup))
//					//}
//					//
//					//// set last group
//					//w.windowGroupsLast[group.Window.Start][group.Window.End] = &ItemGroupedByWindow{
//					//	Key:    group.Key,
//					//	Window: group.Window,
//					//	Data: &schema.List{
//					//		Items: group.Data.Items,
//					//	},
//					//}
//
//					// flush current result
//					returning(ToElement(*group))
//
//					// since we have the last group, we can remove current group
//					// this will be useful, when late arrivals will pass grace period
//					w.RemoveItemGropedByWindow(group)
//				},
//			)
//		}
//	}
//}

func (w *WindowBuffer) EachItemGroupedByWindow(f func(group *ItemGroupedByWindow)) {
	for _, windowGroups := range w.windowGroups {
		for _, group := range windowGroups {
			f(group)
		}
	}
}

func (w *WindowBuffer) RemoveItemGropedByWindow(window *ItemGroupedByWindow) {
	delete(w.windowGroups[window.Window.Start], window.Window.End)
	delete(w.windowGroups, window.Window.Start)
}

//func (w *WindowBuffer) GroupByKey(x []Item) {
//	for _, item := range x {
//		group, ok := w._keyGroups[item.Key]
//		if !ok {
//			group = &ItemGroupedByKey{Key: item.Key}
//			w._keyGroups[item.Key] = group
//		}
//		group.Data = append(group.Data, item)
//	}
//}

func (w *WindowBuffer) GroupAlsoByWindow(x []ItemGroupedByKey) {
	for _, group := range x {
		for _, item := range group.Data {
			if _, ok := w.windowGroups[item.Window.Start]; !ok {
				w.windowGroups[item.Window.Start] = map[int64]*ItemGroupedByWindow{}
			}
			if _, ok := w.windowGroups[item.Window.Start][item.Window.End]; !ok {
				w.windowGroups[item.Window.Start][item.Window.End] = &ItemGroupedByWindow{
					Key:    group.Key,
					Data:   &schema.List{},
					Window: item.Window,
				}
			}

			w.windowGroups[item.Window.Start][item.Window.End].Data.Items =
				append(w.windowGroups[item.Window.Start][item.Window.End].Data.Items, item.Data)
		}

		//// because golang maps are not ordered,
		//// to create ordered result we need to iterate over data again in order to get ordered result
		//for _, item := range group.Data {
		//	if _, ok := w.windowGroups[item.Window.Start]; !ok {
		//		continue
		//	}
		//	if _, ok := w.windowGroups[item.Window.Start][item.Window.End]; !ok {
		//		continue
		//	}
		//
		//	result = append(result, *w.windowGroups[item.Window.Start][item.Window.End])
		//	delete(w.windowGroups[item.Window.Start], item.Window.End)
		//}
	}
}
