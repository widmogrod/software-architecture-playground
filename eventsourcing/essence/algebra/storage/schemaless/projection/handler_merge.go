package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
)

type MergeHandler[A any] struct {
	onCombine func(base A, x A) (A, error)
	//onRetract func(base A, x A) (A, error)
}

func (h *MergeHandler[A]) Process2(a, b Message, returning func(Message)) error {
	return MustMatchMessage(
		a,
		func(x *Combine) error {
			dataA, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			return MustMatchMessage(
				b,
				func(y *Combine) error {
					dataB, err := ConvertAs[A](y.Data)
					if err != nil {
						return err
					}

					res, err := h.onCombine(dataA, dataB)
					if err != nil {
						return err
					}

					returning(&Combine{
						Key:  x.Key,
						Data: schema.FromGo(res),
					})
					return nil
				},
				func(x *Retract) error {
					return fmt.Errorf("MergeHandler: not implemented (1)")
				},
				func(x *Both) error {
					return fmt.Errorf("MergeHandler: not implemented (2)")
				},
			)
		},
		func(x *Retract) error {
			return fmt.Errorf("MergeHandler: not implemented (3)")
		},
		func(x *Both) error {
			return fmt.Errorf("MergeHandler: not implemented (4)")
		},
	)
}
