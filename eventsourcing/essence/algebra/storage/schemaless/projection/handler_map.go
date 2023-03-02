package schemaless

import (
	"fmt"
	"github.com/widmogrod/mkunion/x/schema"
)

type MapHandler[A any, B any] struct {
	F func(x A, returning func(key string, value B)) error
}

func (h *MapHandler[A, B]) Process(msg Message, returning func(Message)) error {
	mapCombineReturning := func(key string, value B) {
		returning(&Combine{
			Key:  key,
			Data: schema.FromGo(value),
		})
	}
	mapRetractReturning := func(key string, value B) {
		returning(&Retract{
			Key:  key,
			Data: schema.FromGo(value),
		})
	}
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			return h.F(data, mapCombineReturning)
		},
		func(x *Retract) error {
			data, err := ConvertAs[A](x.Data)
			if err != nil {
				return err
			}

			return h.F(data, mapRetractReturning)
		},
		func(x *Both) error {
			data, err := ConvertAs[A](x.Combine.Data)
			if err != nil {
				return err
			}

			result := []*Both{}
			err = h.F(data, func(s string, b B) {
				result = append(result, &Both{
					Combine: Combine{
						Key:  s,
						Data: schema.FromGo(b),
					},
				})
			})
			if err != nil {
				return err
			}

			data, err = ConvertAs[A](x.Retract.Data)
			if err != nil {
				return err
			}

			idx := 0
			err = h.F(data, func(s string, b B) {
				result[idx].Retract = Retract{
					Key:  s,
					Data: schema.FromGo(b),
				}
				idx++
			})
			if err != nil {
				return err
			}

			for _, r := range result {
				if r.Combine.Key != r.Retract.Key {
					return fmt.Errorf("MapHandler: key mismatch: %s != %s", r.Combine.Key, r.Retract.Key)
				}

				returning(r)
			}

			return nil
		},
	)
}
