package projection

import (
	"context"
	"errors"
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
	"time"
)

//go:generate mkunion -name=TriggerType -variants=AtPeriod,AtCount,AtWatermark

//go:generate mkunion -name=TriggerDescription
type (
	AtPeriod struct {
		Duration time.Duration
	}
	AtCount struct {
		Number int
	}
	AtWatermark struct{}

	SequenceOf struct {
		Triggers []TriggerDescription
	}
	Repeat struct {
		Trigger TriggerDescription
	}
	RepeatUntil struct {
		Triggers []TriggerDescription
	}
)

//go:generate mkunion -name=WindowFlushMode
type (
	Accumulate struct {
		AllowLateArrival time.Duration
	}
	Discard                   struct{}
	AccumulatingAndRetracting struct {
		AllowLateArrival time.Duration
	}
)

type signal struct {
	kind     signalType
	duration time.Duration
}

type signalType int

const (
	signalCount signalType = iota
	signalPeriod
	signalWatermark
)

type TriggerHandler struct {
	wd WindowDescription
	fm WindowFlushMode
	td TriggerDescription
	tc TriggerContext

	wb *WindowBuffer

	lock sync.Mutex
	//buffer  []Item
	signals chan signal

	groups  map[string]*ItemGroupedByKey
	tickers map[TriggerDescription]*time.Ticker
}

var _ Handler = (*TriggerHandler)(nil)

func printTrigger(triggerType TriggerType) {
	MustMatchTriggerTypeR0(
		triggerType,
		func(x *AtPeriod) {
			fmt.Printf("AtPeriod(%v) \n", x.Duration)

		},
		func(x *AtCount) {
			fmt.Printf("AtCount(%v) \n", x.Number)
		},
		func(x *AtWatermark) {
			fmt.Printf("AtWatermark() \n")
		})
}

func (tm *TriggerHandler) Triggered(trigger TriggerType, returning func(Item)) error {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	//printTrigger(trigger)

	MustMatchTriggerTypeR0(
		trigger,
		func(x *AtPeriod) {
			tm.wb.EachItemGroupedByWindow(func(group *ItemGroupedByWindow) {
				returning(ToElement(group))
				tm.wb.RemoveItemGropedByWindow(group)
			})
		},
		func(x *AtCount) {
			tm.wb.EachItemGroupedByWindow(func(group *ItemGroupedByWindow) {
				returning(ToElement(group))
				tm.wb.RemoveItemGropedByWindow(group)
			})
		},
		func(x *AtWatermark) {
			tm.wb.EachItemGroupedByWindow(func(group *ItemGroupedByWindow) {
				returning(ToElement(group))
				if group.Window.End < tm.tc.Watermark {
					returning(ToElement(group))
					tm.wb.RemoveItemGropedByWindow(group)
				}
			})
		},
	)

	return nil
}

func (tm *TriggerHandler) Process(x Item, returning func(Item)) error {
	// buffer data until trigger fires
	tm.lock.Lock()
	tm.wb.Append(x)

	//tm.buffer = append(tm.buffer, x)
	tm.tc.Count++
	tm.lock.Unlock()

	if tm.tc.Watermark < x.EventTime {
		tm.lock.Lock()
		tm.tc.Watermark = x.EventTime
		tm.lock.Unlock()

		tm.signals <- signal{
			kind: signalWatermark,
		}
	}

	tm.signals <- signal{
		kind: signalCount,
	}

	return nil
}

func (tm *TriggerHandler) Retract(x Item, returning func(Item)) error {
	panic("implement me")
}

func (tm *TriggerHandler) Background(ctx context.Context, returning func(Item)) {
	// start ticker for each trigger
	tm.register(tm.td)

	for {
		select {
		case <-ctx.Done():
			return

		case sig := <-tm.signals:
			switch sig.kind {
			case signalPeriod:
				tm.tc.Durations[sig.duration] = struct{}{}
			}

			if found := tm.evaluate(tm.td, tm.tc, false); found != nil {
				err := tm.Triggered(found, returning)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (tm *TriggerHandler) evaluate(td TriggerDescription, tc TriggerContext, shouldRepeat bool) TriggerType {
	return MustMatchTriggerDescription(
		td,
		func(x *AtPeriod) TriggerType {
			// was any of the durations reached?
			for durations := range tc.Durations {
				if durations == x.Duration {
					if shouldRepeat {
						delete(tm.tc.Durations, x.Duration)
					} else {
						delete(tm.tc.Durations, x.Duration)
						tm.unregister(x)
					}

					return x
				}
			}
			return nil
		},
		func(x *AtCount) TriggerType {
			// was item counts reached?
			if tc.Count%x.Number == 0 {
				if shouldRepeat {
					tc.Count = 0
				} else {
					tm.unregister(x)
				}

				return x
			}

			return nil
		},
		func(x *AtWatermark) TriggerType {
			return x
		},
		func(x *SequenceOf) TriggerType {
			// was sequence of triggers reached?
			for _, trigger := range x.Triggers {
				if found := tm.evaluate(trigger, tc, shouldRepeat); found != nil {
					return found
				}
			}

			return nil
		},
		func(x *Repeat) TriggerType {
			return tm.evaluate(x.Trigger, tc, true)
		},
		func(x *RepeatUntil) TriggerType {
			for _, trigger := range x.Triggers {
				if found := tm.evaluate(trigger, tc, shouldRepeat); found != nil {
					// unregister all triggers
					tm.unregister(x)
					return found
				}
			}
			return nil
		},
	)
}

func (tm *TriggerHandler) register(td TriggerDescription) {
	MustMatchTriggerDescriptionR0(
		td,
		func(x *AtPeriod) {
			tm.tickers[x] = time.NewTicker(x.Duration)
			go func() {
				for range tm.tickers[x].C {
					tm.signals <- signal{
						kind:     signalPeriod,
						duration: x.Duration,
					}
				}
			}()

		},
		func(x *AtCount) {},
		func(x *AtWatermark) {},
		func(x *SequenceOf) {},
		func(x *Repeat) {
			tm.register(x.Trigger)
		},
		func(x *RepeatUntil) {},
	)
}

func (tm *TriggerHandler) unregister(x TriggerDescription) {
	MustMatchTriggerDescriptionR0(
		x,
		func(y *AtPeriod) {
			if tm.tickers[y] == nil {
				return
			}

			tm.tickers[y].Stop()
			delete(tm.tickers, y)

		},
		func(y *AtCount) {},
		func(y *AtWatermark) {},
		func(y *SequenceOf) {},
		func(y *Repeat) {
			tm.unregister(y.Trigger)
		},
		func(y *RepeatUntil) {
			for _, trigger := range y.Triggers {
				tm.unregister(trigger)
			}
		},
	)
}

type TriggerContext struct {
	Count     int
	Durations map[time.Duration]struct{}
	Watermark int64
}

var NotFound = errors.New("not found")

type BagOf[A any] interface {
	Set(key string, value A) error
	Get(key string) (A, error)
}

type InMemoryBagOf[A any] struct {
	m map[string]A
}

func NewInMemoryBagOf[A any]() *InMemoryBagOf[A] {
	return &InMemoryBagOf[A]{
		m: make(map[string]A),
	}
}

func (b *InMemoryBagOf[A]) Set(key string, value A) error {
	b.m[key] = value
	return nil
}

func (b *InMemoryBagOf[A]) Get(key string) (A, error) {
	if value, ok := b.m[key]; ok {
		return value, nil
	}

	var a A
	return a, NotFound
}

type AccumulateDiscardRetractHandler struct {
	wf     WindowFlushMode
	mapf   Handler
	mergef Handler

	bag BagOf[Item]
}

var _ Handler = (*AccumulateDiscardRetractHandler)(nil)

func printItem(x Item, sx ...string) {
	data, _ := schema.ToJSON(x.Data)
	fmt.Println(fmt.Sprintf("Item(%v)", sx), x.Key, x.Window, string(data), x.EventTime)
}

func (a *AccumulateDiscardRetractHandler) Process(x Item, returning func(Item)) error {
	return MustMatchWindowFlushMode(
		a.wf,
		func(y *Accumulate) error {
			key := ItemKeyWindow(x)
			previous, err := a.bag.Get(key)

			isError := err != nil && err != NotFound
			isFound := err == nil
			if isError {
				panic(err)
			}

			if isFound {
				//printItem(previous, "previous")
				//printItem(x, "x")
				return a.mapf.Process(x, func(item Item) {
					z := Item{
						Key:    item.Key,
						Window: item.Window,
						Data: schema.MkList(
							previous.Data,
							item.Data,
						),
						EventTime: item.EventTime,
					}

					err := a.mergef.Process(z, func(item Item) {
						err := a.bag.Set(key, item)
						if err != nil {
							panic(err)
						}

						returning(item)
					})
					if err != nil {
						panic(err)
					}
				})
			}

			return a.mapf.Process(x, func(item Item) {
				err := a.bag.Set(key, item)
				//printItem(item, "set")
				if err != nil {
					panic(err)
				}
				returning(item)
			})
		},
		func(y *Discard) error {
			return a.mapf.Process(x, returning)
		},
		func(y *AccumulatingAndRetracting) error {
			key := ItemKeyWindow(x)
			previous, err := a.bag.Get(key)
			isError := err != nil && err != NotFound
			isFound := err == nil
			if isError {
				panic(err)
			}

			if isFound {
				return a.mapf.Process(x, func(item Item) {
					z := Item{
						Key:    item.Key,
						Window: item.Window,
						Data: schema.MkList(
							previous.Data,
							item.Data,
						),
						EventTime: item.EventTime,
					}

					err := a.mergef.Process(z, func(newAggregate Item) {
						err := a.bag.Set(key, newAggregate)
						if err != nil {
							panic(err)
						}

						returning(previous)     // retract previous
						returning(newAggregate) // emit new aggregate
					})
					if err != nil {
						panic(err)
					}
				})
			}

			return a.mapf.Process(x, func(item Item) {
				err := a.bag.Set(key, item)
				if err != nil {
					panic(err)
				}
				returning(item) // emit aggregate
			})
		},
	)
}

func (a *AccumulateDiscardRetractHandler) Retract(x Item, returning func(Item)) error {
	//TODO implement me
	panic("implement me")
}
