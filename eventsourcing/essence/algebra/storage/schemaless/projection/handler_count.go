package schemaless

import "github.com/widmogrod/mkunion/x/schema"

type CountHandler struct {
	value int
}

func (h *CountHandler) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			oldValue := h.value
			h.value += schema.As[int](x.Data, 0)
			return returning(&Both{
				Retract: Retract{
					Data: schema.MkInt(oldValue),
				},
				Combine: Combine{
					Data: schema.MkInt(h.value),
				},
			})
		},
		func(x *Retract) error {
			oldValue := h.value
			h.value -= schema.As[int](x.Data, 0)
			return returning(&Both{
				Retract: Retract{
					Data: schema.MkInt(oldValue),
				},
				Combine: Combine{
					Data: schema.MkInt(h.value),
				},
			})
		},
		func(x *Both) error {
			oldValue := h.value
			h.value -= schema.As[int](x.Retract.Data, 0)
			h.value += schema.As[int](x.Combine.Data, 0)

			return returning(&Both{
				Retract: Retract{
					Data: schema.MkInt(oldValue),
				},
				Combine: Combine{
					Data: schema.MkInt(h.value),
				},
			})
		},
	)
}
