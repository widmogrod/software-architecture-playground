package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
)

type MergeHandler[A any] struct {
	onCombine func(base A, x A) (A, error)
	//onRetract func(base A, x A) (A, error)
}

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

		result, err = h.onCombine(result, elem)
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
