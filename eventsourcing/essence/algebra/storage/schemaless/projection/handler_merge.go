package schemaless

import (
	"github.com/widmogrod/mkunion/x/schema"
)

type MergeHandler[A any] struct {
	onCombine func(base A, x A) (A, error)
	//onRetract func(base A, x A) (A, error)
}

func (h *MergeHandler[A]) Process2(x, y Item, returning func(Item)) error {
	dataA, err := ConvertAs[A](x.Data)
	if err != nil {
		return err
	}

	dataB, err := ConvertAs[A](y.Data)
	if err != nil {
		return err
	}

	res, err := h.onCombine(dataA, dataB)
	if err != nil {
		return err
	}

	returning(Item{
		Key:  x.Key,
		Data: schema.FromGo(res),
	})
	return nil
}
