package schemaless

import "github.com/widmogrod/mkunion/x/schema"

type MergeHandler[A any] struct {
	state     A
	onCombine func(base A, x A) (A, error)
	onRetract func(base A, x A) (A, error)
}

func (h *MergeHandler[A]) Process(msg Message, returning func(Message)) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			oldState := h.state

			newState, err := h.onCombine(oldState, data)
			if err != nil {
				return err
			}
			h.state = newState

			h.returns(newState, oldState, x.Key, returning)
			return nil
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

			h.returns(newState, oldState, x.Key, returning)
			return nil
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
			newState, err := h.onRetract(oldState, retractData)
			if err != nil {
				return err
			}
			newState, err = h.onCombine(newState, combineData)
			if err != nil {
				return err
			}

			h.state = newState
			h.returns(newState, oldState, x.Key, returning)
			return nil
		},
	)
}

func (h *MergeHandler[A]) returns(newState, oldState A, key string, returning func(Message)) {
	//if any(oldState) == nil {
	//	returning(&Retract{
	//		Key:  key,
	//		Data: schema.FromGo(newState),
	//	})
	//} else {
	returning(&Both{
		Key: key,
		Retract: Retract{
			Key:  key,
			Data: schema.FromGo(oldState),
		},
		Combine: Combine{
			Key:  key,
			Data: schema.FromGo(newState),
		},
	})
	//}
}
