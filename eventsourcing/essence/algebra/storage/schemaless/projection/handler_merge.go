package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
)

var _ Handler = &MergeHandler[any]{}

type MergeHandler[A any] struct {
	Combine   func(base A, x A) (A, error)
	DoRetract func(base A, x A) (A, error)
}

func (h *MergeHandler[A]) Process(x Item, returning func(Item)) error {
	var result A
	var first bool = true
	var err error
	Each(x.Data, func(value schema.Schema) {
		var elem A
		if err != nil {
			return
		}

		elem, err = schemaless.ConvertAs[A](value)
		if err != nil {
			return
		}

		if first {
			first = false
			result = elem
			return
		}

		result, err = h.Combine(result, elem)
		if err != nil {
			return
		}
	})

	if err != nil {
		d, err2 := schema.ToJSON(x.Data)
		return fmt.Errorf("mergeHandler:Process(%s, err=%s) err %s", string(d), err, err2)
	}

	returning(Item{
		Key:  x.Key,
		Data: schema.FromGo(result),
	})

	return nil
}

func (h *MergeHandler[A]) Retract(x Item, returning func(Item)) error {
	var result A
	var first bool = true
	var err error
	Each(x.Data, func(value schema.Schema) {
		var elem A
		if err != nil {
			return
		}

		elem, err = schemaless.ConvertAs[A](value)
		if err != nil {
			return
		}

		if first {
			first = false
			result = elem
			return
		}

		result, err = h.DoRetract(result, elem)
		if err != nil {
			return
		}
	})

	if err != nil {
		d, err2 := schema.ToJSON(x.Data)
		return fmt.Errorf("mergeHandler:Retract(%s, err=%s) err %s", string(d), err, err2)
	}

	returning(Item{
		Key:  x.Key,
		Data: schema.FromGo(result),
	})

	return nil
}
