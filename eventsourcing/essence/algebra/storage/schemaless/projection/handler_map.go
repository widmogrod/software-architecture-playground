package schemaless

import "github.com/widmogrod/mkunion/x/schema"

type MapHandler[A any, B any] struct {
	onCombine func(x A) (B, error)
	onRetract func(x A) (B, error)
}

func (h *MapHandler[A, B]) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			newState, err := h.onCombine(data)
			if err != nil {
				return err
			}

			return returning(&Combine{
				Data: schema.FromGo(newState),
			})
		},
		func(x *Retract) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			newState, err := h.onRetract(data)
			if err != nil {
				return err
			}

			return returning(&Combine{
				Data: schema.FromGo(newState),
			})
		},
		func(x *Both) error {
			data, err := ConvertAs[A](x.Combine.Data)
			if err != nil {
				return err
			}

			newState, err := h.onCombine(data)
			if err != nil {
				return err
			}

			data, err = ConvertAs[A](x.Retract.Data)
			if err != nil {
				return err
			}

			newState, err = h.onRetract(data)
			if err != nil {
				return err
			}

			return returning(&Both{
				Combine: Combine{
					Data: schema.FromGo(newState),
				},
				Retract: Retract{
					Data: schema.FromGo(newState),
				},
			})
		},
	)
}
