package projection

import "github.com/widmogrod/mkunion/x/schema"

type CountHandler struct {
	value int
}

func (h *CountHandler) Process(msg Item, returning func(Item)) error {
	h.value += schema.As[int](msg.Data, 0)
	returning(Item{
		Data: schema.MkInt(h.value),
	})
	return nil
}
