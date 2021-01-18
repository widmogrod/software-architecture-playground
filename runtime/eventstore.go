package runtime

import (
	"container/list"
	"sync"
)

func NewEventStore() *EventStore {
	l := list.New()
	return &EventStore{
		log:           l,
		lock:          &sync.Mutex{},
		version:       0,
		lastReduction: l.Front(),
		err:           nil,
	}
}

func WithError(err error, a *EventStore) *EventStore {
	return &EventStore{
		log:           a.log,
		lock:          a.lock,
		version:       a.version,
		lastReduction: a.lastReduction,
		err:           err,
	}
}

type EventStore struct {
	lock          sync.Locker
	log           *list.List
	version       uint
	lastReduction *list.Element
	err           error
}

type Aggregate struct {
	ID       string
	Type     string
	Snapshot *Snapshot
	Changes  []*Change
}

/*
	creating path:
		req = {aggRef, cmd}
		state, changes, create()

		err = append(aggId, changes, state)

	updating path:
		// aggRef = {aggId, aggType}
		req = {aggRef, cmd}

		retry = 3
		white retry-- > 0 {
			changes = get(req.aggRef)
			initial_state := {}
			state = reduce(changes, initial_state) {
				match {
					"create"
					"turnUp"
					"rotateLeft"
					"moveEast":
				}

				return state
			}

			changes, state = apply(cmd, state)
			err = append(req.aggRef, changes, state)
			if err != VersionConflict {
				break
			}
		}
*/

//type Snapshot struct {4
//	Version uint
//	State interface{}
//}

type Change struct {
	//Type    string
	Payload interface{}
	Version uint
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

	a.version++

	el := a.log.PushBack(&Change{
		Payload: input,
		Version: a.version,
	})

	// TODO workaround, that is. Intoduce something like ReduceFromLatest()
	if a.lastReduction == nil {
		a.lastReduction = el
	}

	return &AggregateAppendResult{
		Ok: a,
	}
}

type Reduced struct {
	StopReduction bool
	Value         interface{}
}

func (a *EventStore) Reduce(f func(change interface{}, result *Reduced) *Reduced, init interface{}) *AggregateResultResult {
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
		result = f(e.Value.(*Change).Payload, result)
		if result.StopReduction {
			break
		}
	}

	return &AggregateResultResult{
		Ok: result.Value,
	}
}

type Reducer interface {
	Apply(change interface{}) error
}

func (a *EventStore) ReduceRecent(reducer Reducer) *AggregateResultResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.err != nil {
		return &AggregateResultResult{
			Err: a.err,
		}
	}

	for e := a.lastReduction; e != nil; e = e.Next() {
		err := reducer.Apply(e.Value.(*Change).Payload)
		if err != nil {
			return &AggregateResultResult{
				Err: err,
			}
		}
	}

	a.lastReduction = nil

	return &AggregateResultResult{
		Ok: reducer,
	}
}

type AggregateAppendResult struct {
	Ok  *EventStore
	Err error
}

type AggregateResultResult struct {
	Ok  interface{}
	Err error
}
