package eventsourcing

import (
	"container/list"
	"sync"
)

func NewEventStore() *EventStore {
	return &EventStore{
		log:  list.New(),
		lock: &sync.Mutex{},
		err:  nil,
	}
}

func WithError(err error, a *EventStore) *EventStore {
	return &EventStore{
		log:  a.log,
		lock: a.lock,
		err:  err,
	}
}

type EventStore struct {
	lock sync.Locker
	log  *list.List
	err  error
}

func (a *EventStore) Append(input interface{}) *AggregateAppendResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.err != nil {
		return &AggregateAppendResult{
			Ok:  a,
			Err: a.err,
		}
	}

	a.log.PushBack(input)

	return &AggregateAppendResult{
		Ok: a,
	}
}

type Reduced struct {
	StopReduction bool
	Value         interface{}
}

func (a *EventStore) Reduce(f func(cmd interface{}, result *Reduced) *Reduced, init interface{}) *AggregateResultResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	result := &Reduced{
		StopReduction: false,
		Value:         init,
	}

	if a.err != nil {
		return &AggregateResultResult{
			Ok:  result,
			Err: a.err,
		}
	}

	for e := a.log.Front(); e != nil; e = e.Next() {
		result = f(e.Value, result)
		if result.StopReduction {
			break
		}
	}

	return &AggregateResultResult{
		Ok: result,
	}
}

type AggregateAppendResult struct {
	Ok  *EventStore
	Err error
}

type AggregateResultResult struct {
	Ok  *Reduced
	Err error
}
