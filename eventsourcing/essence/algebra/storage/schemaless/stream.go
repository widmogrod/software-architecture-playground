package schemaless

import (
	"container/list"
	"sync"
)

type Change[T any] struct {
	Before  *Record[T]
	After   *Record[T]
	Deleted bool
}

func NewAppendLog[T any]() *AppendLog[T] {
	res := &AppendLog[T]{
		log: list.List{},
	}
	res.cond = sync.NewCond(&res.mux)
	return res
}

// AppendLog is a stream of events, and in context of schemaless, it is a stream of changes to records, or deleted record with past state
type AppendLog[T any] struct {
	log  list.List
	mux  sync.Mutex
	cond *sync.Cond
}

func (a *AppendLog[T]) Change(from, to Record[T]) error {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.push(Change[T]{
		Before:  &from,
		After:   &to,
		Deleted: false,
	})
	a.cond.Signal()
	return nil
}

func (a *AppendLog[T]) Delete(data Record[T]) error {
	a.mux.Lock()
	defer a.mux.Unlock()

	a.push(Change[T]{
		Before:  &data,
		Deleted: true,
	})
	a.cond.Signal()
	return nil
}

func (a *AppendLog[T]) push(x Change[T]) {
	a.log.PushBack(x)
}

func (a *AppendLog[T]) Append(b *AppendLog[T]) {
	a.mux.Lock()
	defer a.mux.Unlock()

	b.mux.Lock()
	defer b.mux.Unlock()

	for e := b.log.Front(); e != nil; e = e.Next() {
		a.push(e.Value.(Change[T]))
	}
	a.cond.Signal()
}

func (a *AppendLog[T]) Subscribe2() <-chan Change[T] {
	changes := make(chan Change[T], 1)

	go func() {
		defer close(changes)

		var prev *list.Element = nil
		for {
			// Wait for new changes to be available
			a.cond.L.Lock()
			for prev == a.log.Back() {
				a.cond.Wait()
			}
			prev = a.log.Back()
			a.cond.L.Unlock()

			var next *list.Element = prev
			for {
				changes <- next.Value.(Change[T])
				next = prev.Next()
				if next == nil {
					break
				}
				prev = next
			}
		}
	}()

	return changes
}

func (a *AppendLog[T]) Subscribe(fromOffset int) (Change[T], int) {
	a.mux.Lock()
	defer a.mux.Unlock()

	if a.log.Len() <= fromOffset {
		return Change[T]{}, fromOffset
	}

	var i int
	var msg Change[T]
	var found bool
	for e := a.log.Front(); e != nil; e = e.Next() {
		if i == fromOffset {
			found = true
			msg = e.Value.(Change[T])
			break
		}
		i++
	}

	if !found {
		return Change[T]{}, fromOffset
	}

	return msg, i + 1
}
