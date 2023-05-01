package projection

import (
	"context"
	"sync"
	"time"
)

//go:generate mkunion -name=TriggerDescription
type (
	AtPeriod struct {
		Duration time.Duration
	}
	AtCount     struct{}
	AtWatermark struct{}

	SequenceOf struct {
		Triggers []TriggerDescription
	}
	RepeatUntil struct{}
)

type TriggerHandler struct {
	td TriggerDescription
	wd WindowDescription

	lock   sync.Mutex
	buffer []Item

	groups map[string]*ItemGroupedByKey
}

var _ Handler = (*TriggerHandler)(nil)

func (tm *TriggerHandler) Triggered(returning func(Item)) error {
	// if trigger fires, return buffered data
	tm.lock.Lock()
	list0 := tm.buffer
	tm.buffer = make([]Item, 0)
	tm.lock.Unlock()

	list1 := AssignWindows(list0, tm.wd)
	list2 := DropTimestamps(list1)
	list3 := GroupByKey(list2)
	list4 := MergeWindows(list3, tm.wd)
	list5 := GroupAlsoByWindow(list4)
	list6 := ExpandToElements(list5)

	for _, item := range list6 {
		returning(item)
	}
	return nil
}

func (tm *TriggerHandler) Process(x Item, returning func(Item)) error {
	// buffer data until trigger fires
	tm.lock.Lock()
	tm.buffer = append(tm.buffer, x)
	tm.lock.Unlock()

	return nil
}

func (tm *TriggerHandler) Retract(x Item, returning func(Item)) error {
	panic("implement me")
}

func Trigger(ctx context.Context, trigger func(), td TriggerDescription) {
	MustMatchTriggerDescription(
		td,
		func(x *AtPeriod) any {
			for range time.NewTicker(x.Duration).C {
				trigger()
			}
			return nil
		},
		func(x *AtCount) any {
			return nil
		},
		func(x *AtWatermark) any {
			return nil
		},
		func(x *SequenceOf) any {
			return nil
		},
		func(x *RepeatUntil) interface{} {
			return nil
		},
	)
}
