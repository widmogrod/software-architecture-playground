package projection

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
	"time"
)

func TestTriggers(t *testing.T) {
	useCases := map[string]struct {
		td       TriggerDescription
		wd       WindowDescription
		fm       WindowFlushMode
		expected []Item
	}{
		"should trigger window emitting once at period 100ms, and 10 items arrives as 1 item": {
			td: &AtPeriod{
				Duration: 100 * time.Millisecond,
			},
			wd: &FixedWindow{
				Width: 100 * time.Millisecond,
			},
			fm: &Discard{},
			expected: []Item{
				{
					Key: "key",
					Data: schema.MkList(
						schema.MkInt(0), schema.MkInt(1), schema.MkInt(2), schema.MkInt(3), schema.MkInt(4),
						schema.MkInt(5), schema.MkInt(6), schema.MkInt(7), schema.MkInt(8), schema.MkInt(9),
					),
					EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0),
						End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
					},
				},
				{
					Key: "key",
					Data: schema.MkList(
						schema.MkInt(10), schema.MkInt(11), schema.MkInt(12), schema.MkInt(13), schema.MkInt(14),
						schema.MkInt(15), schema.MkInt(16), schema.MkInt(17), schema.MkInt(18),
						// it should fit in 100ms window, but due timeouts being part of process time, not event time,
						// it's not guaranteed that when system will receive event at 10.1s, it will be processed at 10.2s
						// schema.MkInt(19),
					),
					EventTime: withTime(10, 0) + (200 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0) + (100 * int64(time.Millisecond)),
						End:   withTime(10, 0) + (200 * int64(time.Millisecond)),
					},
				},
			},
		},
		"should trigger window emitting first item arrives 1 item, and don't emmit when there are more events": {
			td: &AtCount{
				Number: 1,
			},
			wd: &FixedWindow{
				Width: 100 * time.Millisecond,
			},
			fm: &Discard{},
			expected: []Item{
				{
					Key: "key",
					Data: schema.MkList(
						schema.MkInt(0),
					),
					EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0),
						End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
					},
				},
			},
		},
		"should trigger window flush at watermark": {
			td: &AtWatermark{},
			wd: &FixedWindow{
				Width: 100 * time.Millisecond,
			},
			fm: &Discard{},
			expected: []Item{
				{
					Key: "key",
					Data: schema.MkList(
						schema.MkInt(0), schema.MkInt(1), schema.MkInt(2), schema.MkInt(3), schema.MkInt(4),
						schema.MkInt(5), schema.MkInt(6), schema.MkInt(7), schema.MkInt(8), schema.MkInt(9),
					),
					EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0),
						End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
					},
				},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			handler := &TriggerHandler{
				td: uc.td,
				wd: uc.wd,
				fm: uc.fm,
				wb: &WindowBuffer{
					wd:           uc.wd,
					fm:           uc.fm,
					windowGroups: map[int64]map[int64]*ItemGroupedByWindow{},
				},
				wts: NewInMemoryBagOf[*WindowTrigger](),
			}

			returning := &ListAssert{t: t}

			tickers := NewTriggersManager()
			defer tickers.Unregister(uc.td)
			tickers.Register(uc.td, func(triggerType TriggerType) {
				// propagate trigger to handler
				err := handler.Triggered(triggerType, returning.Returning)
				assert.NoError(t, err)
			})

			for item := range GenerateItemsEvery(withTime(10, 0), 20, 10*time.Millisecond) {
				err := handler.Process(item, returning.Returning)
				assert.NoError(t, err)

				// simulate watermark
				err = handler.Triggered(&AtWatermark{
					Timestamp: item.EventTime,
				}, returning.Returning)
				assert.NoError(t, err)
			}

			time.Sleep(100 * time.Millisecond)
			for i, expected := range uc.expected {
				returning.AssertAt(i, expected)
			}
		})
	}
}

func TestAggregate(t *testing.T) {
	// arithmetic sum of series 0..9, 10..19, 0 .. 19
	// 45, 145, 190
	useCases := map[string]struct {
		td       TriggerDescription
		wd       WindowDescription
		fm       WindowFlushMode
		expected []Item
	}{
		"should trigger window emitting evey period 100ms, and 10 items arrives as 1 item, late arrivals are new aggregations": {
			td: &AtPeriod{
				Duration: 100 * time.Millisecond,
			},
			wd: &FixedWindow{
				Width: 100 * time.Millisecond,
			},
			fm: &Discard{},
			expected: []Item{
				{
					Key:       "key",
					Data:      schema.MkInt(45), // arithmetic sum fo series 0..9
					EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0),
						End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
					},
				},
				{
					Key:       "key",
					Data:      schema.MkInt(126),
					EventTime: withTime(10, 0) + (200 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0) + (100 * int64(time.Millisecond)),
						End:   withTime(10, 0) + (200 * int64(time.Millisecond)),
					},
				},
			},
		},
		"should trigger window emitting evey period 100ms, and 10 items arrives as 1 item, late arrivals use past aggregation as base": {
			td: &AtPeriod{
				Duration: 100 * time.Millisecond,
			},
			wd: &FixedWindow{
				Width: 100 * time.Millisecond,
			},
			fm: &Accumulate{},
			expected: []Item{
				{
					Key:       "key",
					Data:      schema.MkInt(45), // arithmetic sum fo series 0..9
					EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0),
						End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
					},
				},
				// this window is incomplete, and will be remitted
				{
					Key:       "key",
					Data:      schema.MkInt(126),
					EventTime: withTime(10, 0) + (200 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0) + (100 * int64(time.Millisecond)),
						End:   withTime(10, 0) + (200 * int64(time.Millisecond)),
					},
				},
				// here is complete aggregation in effect.
				{
					Key:       "key",
					Data:      schema.MkInt(145), // arithmetic sum of series 10..19
					EventTime: withTime(10, 0) + (200 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0) + (100 * int64(time.Millisecond)),
						End:   withTime(10, 0) + (200 * int64(time.Millisecond)),
					},
				},
			},
		},
		"should trigger window emitting every period 100ms, and 10 items arrives as 1 item, late arrivals use past aggregation as base, and retract last change": {
			td: &AtPeriod{
				Duration: 100 * time.Millisecond,
			},
			wd: &FixedWindow{
				Width: 100 * time.Millisecond,
			},
			fm: &AccumulatingAndRetracting{},
			expected: []Item{
				{
					Key:       "key",
					Data:      schema.MkInt(45), // arithmetic sum fo series 0..9
					EventTime: withTime(10, 0) + (100 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0),
						End:   withTime(10, 0) + (100 * int64(time.Millisecond)),
					},
					Type: ItemAggregation,
				},
				// this window is incomplete, and will be remitted
				{
					Key:       "key",
					Data:      schema.MkInt(126),
					EventTime: withTime(10, 0) + (200 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0) + (100 * int64(time.Millisecond)),
						End:   withTime(10, 0) + (200 * int64(time.Millisecond)),
					},
					Type: ItemAggregation,
				},
				// here is retracting and aggregate in effect.
				{
					Key: "key",
					Data: schema.MkList(
						schema.MkInt(126), // retract previous
						schema.MkInt(145), // aggregate new
					),
					EventTime: withTime(10, 0) + (200 * int64(time.Millisecond)),
					Window: &Window{
						Start: withTime(10, 0) + (100 * int64(time.Millisecond)),
						End:   withTime(10, 0) + (200 * int64(time.Millisecond)),
					},
					Type: ItemRetractAndAggregate,
				},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			handler := &TriggerHandler{
				td: uc.td,
				wd: uc.wd,
				fm: uc.fm,
				wb: &WindowBuffer{
					wd:           uc.wd,
					fm:           uc.fm,
					windowGroups: map[int64]map[int64]*ItemGroupedByWindow{},
				},
				wts: NewInMemoryBagOf[*WindowTrigger](),
			}

			handler2 := &AccumulateDiscardRetractHandler{
				wf: uc.fm,
				mapf: &SimpleHandler{
					P: func(item Item, returning func(Item)) error {
						returning(Item{
							Key: item.Key,
							Data: schema.MkInt(schema.Reduce(
								item.Data,
								0,
								func(s schema.Schema, i int) int {
									x, err := schema.ToGoG[int](s)
									if err != nil {
										panic(err)
									}
									return x + i
								},
							)),
							EventTime: item.EventTime,
							Window:    item.Window,
						})
						return nil
					}},
				mergef: &MergeHandler[int]{
					Combine: func(a, b int) (int, error) {
						return a + b, nil
					},
				},
				bag: NewInMemoryBagOf[Item](),
			}

			returning := &ListAssert{t: t}
			returning2 := func(item Item) {
				err := handler2.Process(item, returning.Returning)
				assert.NoError(t, err)
			}

			tickers := NewTriggersManager()
			defer tickers.Unregister(uc.td)
			tickers.Register(uc.td, func(triggerType TriggerType) {
				// propagate trigger to handler
				err := handler.Triggered(triggerType, returning2)
				assert.NoError(t, err)
			})

			for item := range GenerateItemsEvery(withTime(10, 0), 20, 10*time.Millisecond) {
				err := handler.Process(item, returning2)
				assert.NoError(t, err)

				// simulate watermark
				err = handler.Triggered(&AtWatermark{
					Timestamp: item.EventTime,
				}, returning2)
				assert.NoError(t, err)
			}

			time.Sleep(100 * time.Millisecond)
			for i, expected := range uc.expected {
				returning.AssertAt(i, expected)
			}
		})
	}
}
