package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
)

var _ Handler = &MapHandler[any, any]{}

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
	data, err := schemaless.ConvertAs[A](x.Data)
	if err != nil {
		return err
	}

	return h.F(data, mapCombineReturning)
}

func (h *MapHandler[A, B]) Retract(x Item, returning func(Item)) error {
	return h.Process(x, returning)
}
