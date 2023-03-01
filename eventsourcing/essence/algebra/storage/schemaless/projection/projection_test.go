package schemaless

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func NewBuilder() Builder {
	return nil
}

type handler struct{}

func (h *handler) Process(msg Message) (Message, error) {
	return nil, nil
}

func (h *handler) InputType() TypeDef {
	return TypeDef{}
}

func (h *handler) OutputType() TypeDef {
	return TypeDef{}
}

func Log() Handler {
	return &handler{}
}

func GenerateData() Handler {
	return &handler{}
}

func MapGameToStats() Handler {
	return &handler{}
}

func MapGameStatsToSession() Handler {
	return &handler{}
}

func TestProjection(t *testing.T) {
	t.Skip("not implemented yet")
	// Given
	// When
	// Then

	dag := NewBuilder()
	games := dag.Map(GenerateData())
	gameStats := games.Map(MapGameToStats())
	gameStatsBySession := gameStats.Merge(MapGameStatsToSession())
	gameStatsBySession.Map(Log())

	expected := &Map{
		OnMap: Log(),
		Input: &Merge{
			OnMerge: MapGameStatsToSession(),
			Input: []DAG{
				&Map{
					OnMap: MapGameToStats(),
					Input: &Map{
						OnMap: GenerateData(),
						Input: nil,
					},
				},
			},
		},
	}
	assert.Equal(t, expected, dag.Build())
}

type CountHandler struct {
	value int
}

func (h *CountHandler) Process(msg Message) (Message, error) {
	return MustMatchMessageR2(
		msg,
		func(x *Combine) (Message, error) {
			oldValue := h.value
			h.value += schema.As[int](x.Data, 0)
			return &Both{
				Retract: Retract{
					Data: schema.MkInt(oldValue),
				},
				Combine: Combine{
					Data: schema.MkInt(h.value),
				},
			}, nil
		},
		func(x *Retract) (Message, error) {
			oldValue := h.value
			h.value -= schema.As[int](x.Data, 0)
			return &Both{
				Retract: Retract{
					Data: schema.MkInt(oldValue),
				},
				Combine: Combine{
					Data: schema.MkInt(h.value),
				},
			}, nil
		},
		func(x *Both) (Message, error) {
			oldValue := h.value
			h.value -= schema.As[int](x.Retract.Data, 0)
			h.value += schema.As[int](x.Combine.Data, 0)

			return &Both{
				Retract: Retract{
					Data: schema.MkInt(oldValue),
				},
				Combine: Combine{
					Data: schema.MkInt(h.value),
				},
			}, nil
		},
	)
}

func TestCountHandler(t *testing.T) {
	h := &CountHandler{}
	assert.Equal(t, 0, h.value)

	message, err := h.Process(&Combine{
		Data: schema.MkInt(1),
	})
	assert.NoError(t, err)
	assert.Equal(t, &Both{
		Retract: Retract{
			Data: schema.MkInt(0),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	}, message)
	assert.Equal(t, 1, h.value)

	message, err = h.Process(&Combine{
		Data: schema.MkInt(2),
	})
	assert.NoError(t, err)
	assert.Equal(t, &Both{
		Retract: Retract{
			Data: schema.MkInt(1),
		},
		Combine: Combine{
			Data: schema.MkInt(3),
		},
	}, message)
	assert.Equal(t, 3, h.value)

	message, err = h.Process(&Retract{
		Data: schema.MkInt(1),
	})
	assert.NoError(t, err)
	assert.Equal(t, &Both{
		Retract: Retract{
			Data: schema.MkInt(3),
		},
		Combine: Combine{
			Data: schema.MkInt(2),
		},
	}, message)
	assert.Equal(t, 2, h.value)

	message, err = h.Process(&Both{
		Retract: Retract{
			Data: schema.MkInt(2),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, &Both{
		Retract: Retract{
			Data: schema.MkInt(2),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	}, message)
	assert.Equal(t, 1, h.value)
}

type AvgHandler struct {
	avg   float64
	count int
}

func (h *AvgHandler) Process(msg Message) (Message, error) {
	return MustMatchMessageR2(
		msg,
		func(x *Combine) (Message, error) {
			oldValue := schema.Number(h.avg)

			// avg = (avg * count + x) / (count + 1)
			h.count += 1
			h.avg = (h.avg*float64(h.count) + schema.As[float64](x.Data, 0)) / (float64(h.count) + 1)

			newValue := schema.Number(h.avg)

			return &Both{
				Retract: Retract{
					Data: &oldValue,
				},
				Combine: Combine{
					Data: &newValue,
				},
			}, nil
		},
		func(x *Retract) (Message, error) {
			oldValue := schema.Number(h.avg)

			// avg = (avg * count - x) / (count - 1)
			h.count -= 1
			h.avg = (h.avg*float64(h.count) - schema.As[float64](x.Data, 0)) / (float64(h.count) - 1)

			newValue := schema.Number(h.avg)

			return &Both{
				Retract: Retract{
					Data: &oldValue,
				},
				Combine: Combine{
					Data: &newValue,
				},
			}, nil

		},
		func(x *Both) (Message, error) {
			oldValue := schema.Number(h.avg)

			// avg = (avg * count - x) / (count - 1)
			h.count -= 1
			h.avg = (h.avg*float64(h.count) - schema.As[float64](x.Retract.Data, 0)) / (float64(h.count) - 1)

			// avg = (avg * count + x) / (count + 1)
			h.count += 1
			h.avg = (h.avg*float64(h.count) + schema.As[float64](x.Combine.Data, 0)) / (float64(h.count) + 1)

			newValue := schema.Number(h.avg)

			return &Both{
				Retract: Retract{
					Data: &oldValue,
				},
				Combine: Combine{
					Data: &newValue,
				},
			}, nil
		},
	)
}

func TestAvgHandler(t *testing.T) {
	h := &AvgHandler{}
	assert.Equal(t, float64(0), h.avg)

	message, err := h.Process(&Combine{
		Data: schema.MkInt(1),
	})
	assert.NoError(t, err)
	assert.Equal(t, &Both{
		Retract: Retract{
			Data: schema.MkFloat(0),
		},
		Combine: Combine{
			Data: schema.MkFloat(0.5),
		},
	}, message)
}
