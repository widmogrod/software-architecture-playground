package projection

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"sync"
	"testing"
)

func Each(x schema.Schema, f func(value schema.Schema)) {
	_ = schema.MustMatchSchema(
		x,
		func(x *schema.None) any {
			return nil
		},
		func(x *schema.Bool) any {
			f(x)
			return nil
		},
		func(x *schema.Number) any {
			f(x)
			return nil
		},
		func(x *schema.String) any {
			f(x)
			return nil
		},
		func(x *schema.Binary) any {
			f(x)
			return nil
		},
		func(x *schema.List) any {
			for _, v := range x.Items {
				f(v)
			}
			return nil
		},
		func(x *schema.Map) any {
			f(x)
			return nil
		},
	)
}

func NewDual() *Dual {
	return &Dual{}
}

type Dual struct {
	lock sync.Mutex
	list []*Message

	aggIdx int
	retIdx int
}

func (d *Dual) ReturningAggregate(msg Item) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.list = append(d.list, &Message{
		Key:       msg.Key,
		Aggregate: &msg,
	})

	d.aggIdx++
}

func (d *Dual) ReturningRetract(msg Item) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.retIdx <= len(d.list) {
		if d.list[d.retIdx].Key != msg.Key {
			panic("key mismatch")
		}

		d.list[d.retIdx].Retract = &msg
		d.retIdx++
	}
}

func (d *Dual) IsValid() bool {
	return d.aggIdx == d.retIdx
}

func (d *Dual) List() []*Message {
	return d.list
}

type ListAssert struct {
	t     *testing.T
	Items []Item
	Err   error
}

func (l *ListAssert) Returning(msg Item) {
	if l.Err != nil {
		panic(l.Err)
	}

	l.Items = append(l.Items, msg)
}

func (l *ListAssert) AssertLen(expected int) bool {
	return assert.Equal(l.t, expected, len(l.Items))
}

func (l *ListAssert) AssertAt(index int, expected Item) bool {
	return assert.Equal(l.t, expected, l.Items[index])
}

func (l *ListAssert) Contains(expected Item) bool {
	for _, item := range l.Items {
		if assert.Equal(l.t, expected, item) {
			return true
		}
	}

	l.t.Errorf("expected to find %v in result set but failed", expected)
	return false
}
