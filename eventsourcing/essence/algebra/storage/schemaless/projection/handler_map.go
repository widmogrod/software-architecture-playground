package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
)

type MapHandler[A any, B any] struct {
	F func(x A, returning func(key string, value B)) error
}

func (h *MapHandler[A, B]) Process(x Item, returning func(Item)) error {
	mapCombineReturning := func(key string, value B) {
		returning(Item{
			Key:  key,
			Data: schema.FromGo(value),
		})
	}
	data, err := ConvertAs[A](x.Data)
	if err != nil {
		return err
	}

	return h.F(data, mapCombineReturning)
}
