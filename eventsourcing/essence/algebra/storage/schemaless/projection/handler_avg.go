package schemaless

import "github.com/widmogrod/mkunion/x/schema"

type AvgHandler struct {
	avg   float64
	count int
}

func (h *AvgHandler) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			oldValue := schema.Number(h.avg)

			h.avg = (h.avg*float64(h.count) + schema.As[float64](x.Data, 0)) / (float64(h.count) + 1)
			// avg = (avg * count + x) / (count + 1)
			h.count += 1

			newValue := schema.Number(h.avg)

			return returning(&Both{
				Retract: Retract{
					Data: &oldValue,
				},
				Combine: Combine{
					Data: &newValue,
				},
			})
		},
		func(x *Retract) error {
			oldValue := schema.Number(h.avg)

			h.avg = (h.avg*float64(h.count) - schema.As[float64](x.Data, 0)) / (float64(h.count) - 1)
			// avg = (avg * count - x) / (count - 1)
			h.count -= 1

			newValue := schema.Number(h.avg)

			return returning(&Both{
				Retract: Retract{
					Data: &oldValue,
				},
				Combine: Combine{
					Data: &newValue,
				},
			})
		},
		func(x *Both) error {
			oldValue := schema.Number(h.avg)

			h.avg = (h.avg*float64(h.count) - schema.As[float64](x.Retract.Data, 0)) / (float64(h.count) - 1)
			// avg = (avg * count - x) / (count - 1)
			h.count -= 1

			h.avg = (h.avg*float64(h.count) + schema.As[float64](x.Combine.Data, 0)) / (float64(h.count) + 1)
			// avg = (avg * count + x) / (count + 1)
			h.count += 1

			newValue := schema.Number(h.avg)

			return returning(&Both{
				Retract: Retract{
					Data: &oldValue,
				},
				Combine: Combine{
					Data: &newValue,
				},
			})
		},
	)
}
