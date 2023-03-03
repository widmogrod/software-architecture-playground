package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
)

var _ Handler = &MergeHandler[any]{}

type MergeHandler[A any] struct {
	Combine func(base A, x A) (A, error)
	//onRetract func(base A, x A) (A, error)
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

		elem, err = ConvertAs[A](value)
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

	returning(Item{
		Key:  x.Key,
		Data: schema.FromGo(result),
	})

	return nil
}

func (h *MergeHandler[A]) Retract(x Item, returning func(Item)) error {
	//TODO implement me
	panic("implement me")
}
