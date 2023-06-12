package projection

import (
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
	AtWatermark struct {
		Timestamp int64
	}
	AnyOf struct {
		Triggers []TriggerDescription
	}
	AllOf struct {
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

type TriggerHandler struct {
	wd WindowDescription
	fm WindowFlushMode
	td TriggerDescription

	wb *WindowBuffer

	wts BagOf[*WindowTrigger]

	lock sync.Mutex
}

var _ Handler = (*TriggerHandler)(nil)

func printTrigger(triggerType TriggerType) {
	MustMatchTriggerTypeR0(
		triggerType,
		func(x *AtPeriod) {
			fmt.Printf("AtPeriod(%v)", x.Duration)

		},
		func(x *AtCount) {
			fmt.Printf("AtCount(%v)", x.Number)
		},
		func(x *AtWatermark) {
			fmt.Printf("AtWatermark()")
		})
}

func (tm *TriggerHandler) Triggered(trigger TriggerType, returning func(Item)) error {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	tm.wb.EachItemGroupedByWindow(func(group *ItemGroupedByWindow) {
		wt, err := tm.wts.Get(WindowKey(group.Window))
		isError := err != nil && err != NotFound
		isFound := err == nil

		if isError {
			panic(err)
		}

		if !isFound {
			wt = NewWindowTrigger(group.Window, tm.td)
			err = tm.wts.Set(WindowKey(group.Window), wt)
			if err != nil {
				panic(err)
			}
		}

		wt.ReceiveEvent(trigger)
		wt.ReceiveEvent(&AtCount{Number: len(group.Data.Items)})

		if wt.ShouldTrigger() {
			returning(ToElement(group))
			tm.wb.RemoveItemGropedByWindow(group)
		}
	})

	return nil
}

func (tm *TriggerHandler) Process(x Item, returning func(Item)) error {
	tm.lock.Lock()
	tm.wb.Append(x)
	tm.lock.Unlock()
	return tm.Triggered(&AtCount{Number: 0}, returning)
}

func (tm *TriggerHandler) Retract(x Item, returning func(Item)) error {
	panic("implement me")
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

						// operation is in one messages, as one or nothing principle
						// which will help in transactional systems.
						returning(Item{
							Key: newAggregate.Key,
							Data: PackRetractAndAggregate(
								previous.Data,
								newAggregate.Data,
							),
							EventTime: newAggregate.EventTime,
							Window:    newAggregate.Window,
							Type:      ItemRetractAndAggregate,
						})
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

func NewTriggersManager() *TriggersManager {
	return &TriggersManager{
		tickers: map[TriggerDescription]*time.Ticker{},
	}
}

type TriggersManager struct {
	tickers map[TriggerDescription]*time.Ticker
}

func (tm *TriggersManager) Register(td TriggerDescription, onTrigger func(triggerType TriggerType)) {
	MustMatchTriggerDescriptionR0(
		td,
		func(x *AtPeriod) {
			go func() {
				tm.tickers[td] = time.NewTicker(x.Duration)
				for range tm.tickers[td].C {
					onTrigger(&AtPeriod{
						Duration: x.Duration,
					})
				}
			}()
		},
		func(x *AtCount) {},
		func(x *AtWatermark) {},
		func(x *AnyOf) {
			for _, td := range x.Triggers {
				tm.Register(td, onTrigger)
			}
		},
		func(x *AllOf) {
			for _, td := range x.Triggers {
				tm.Register(td, onTrigger)
			}
		},
	)
}

func (tm *TriggersManager) Unregister(td TriggerDescription) {
	MustMatchTriggerDescriptionR0(
		td,
		func(x *AtPeriod) {
			if ticker, ok := tm.tickers[td]; ok {
				ticker.Stop()
				delete(tm.tickers, td)
			}
		},
		func(x *AtCount) {},
		func(x *AtWatermark) {},
		func(x *AnyOf) {
			for _, td := range x.Triggers {
				tm.Unregister(td)
			}
		},
		func(x *AllOf) {
			for _, td := range x.Triggers {
				tm.Unregister(td)
			}
		},
	)
}
