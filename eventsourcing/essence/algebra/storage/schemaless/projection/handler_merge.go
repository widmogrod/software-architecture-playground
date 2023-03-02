package schemaless

import "github.com/widmogrod/mkunion/x/schema"

type MergeHandler[A any] struct {
	state     A
	onCombine func(base A, x A) (A, error)
	onRetract func(base A, x A) (A, error)
}

func (h *MergeHandler[A]) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			oldState := h.state

			newState, err := h.onCombine(h.state, data)
			if err != nil {
				return err
			}
			h.state = newState

			return h.returns(newState, oldState, returning)
		},
		func(x *Retract) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			oldState := h.state

			newState, err := h.onRetract(h.state, data)
			if err != nil {
				return err
			}
			h.state = newState

			return h.returns(newState, oldState, returning)
		},
		func(x *Both) error {
			combineData, err := ConvertAs[A](x.Combine.Data)
			if err != nil {
				return err
			}

			retractData, err := ConvertAs[A](x.Retract.Data)
			if err != nil {
				return err
			}

			oldState := h.state

			newState, err := h.onCombine(h.state, combineData)
			if err != nil {
				return err
			}
			newState, err = h.onRetract(newState, retractData)
			if err != nil {
				return err
			}
			h.state = newState

			return h.returns(newState, oldState, returning)
		},
	)
}

func (h *MergeHandler[A]) returns(newState, oldState A, returning func(Message) error) error {
	if any(newState) == nil {
		return returning(&Retract{
			Data: schema.FromGo(oldState),
		})
	} else {
		return returning(&Both{
			Retract: Retract{
				Data: schema.FromGo(oldState),
			},
			Combine: Combine{
				Data: schema.FromGo(newState),
			},
		})
	}
}
