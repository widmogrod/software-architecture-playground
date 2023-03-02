package schemaless

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"testing"
)

func NewBuilder() Builder {
	return nil
}

type handler struct{}

func (h *handler) Process(msg Message, next func(Message) error) error {
	return fmt.Errorf("not implemented yet")
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

func TestCountHandler(t *testing.T) {
	h := &CountHandler{}
	assert.Equal(t, 0, h.value)

	l := &ListAssert{t: t}
	err := h.Process(&Combine{
		Data: schema.MkInt(1),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Retract: Retract{
			Data: schema.MkInt(0),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	})
	assert.Equal(t, 1, h.value)

	err = h.Process(&Combine{
		Data: schema.MkInt(2),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(1, &Both{
		Retract: Retract{
			Data: schema.MkInt(1),
		},
		Combine: Combine{
			Data: schema.MkInt(3),
		},
	})
	assert.Equal(t, 3, h.value)

	err = h.Process(&Retract{
		Data: schema.MkInt(1),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(2, &Both{
		Retract: Retract{
			Data: schema.MkInt(3),
		},
		Combine: Combine{
			Data: schema.MkInt(2),
		},
	})
	assert.Equal(t, 2, h.value)

	err = h.Process(&Both{
		Retract: Retract{
			Data: schema.MkInt(2),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(3, &Both{
		Retract: Retract{
			Data: schema.MkInt(2),
		},
		Combine: Combine{
			Data: schema.MkInt(1),
		},
	})
	assert.Equal(t, 1, h.value)
}

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

func TestAvgHandler(t *testing.T) {
	h := &AvgHandler{}
	assert.Equal(t, float64(0), h.avg)

	l := ListAssert{t: t}

	err := h.Process(&Combine{
		Data: schema.MkInt(1),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Retract: Retract{
			Data: schema.MkFloat(0),
		},
		Combine: Combine{
			Data: schema.MkFloat(1),
		},
	})
	assert.Equal(t, float64(1), h.avg)
	assert.Equal(t, 1, h.count)

	err = h.Process(&Combine{
		Data: schema.MkInt(11),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(1, &Both{
		Retract: Retract{
			Data: schema.MkFloat(1),
		},
		Combine: Combine{
			Data: schema.MkFloat(6),
		},
	})
	assert.Equal(t, float64(6), h.avg)
	assert.Equal(t, 2, h.count)

	err = h.Process(&Combine{
		Data: schema.MkInt(3),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(2, &Both{
		Retract: Retract{
			Data: schema.MkFloat(6),
		},
		Combine: Combine{
			Data: schema.MkFloat(5),
		},
	})
	assert.Equal(t, float64(5), h.avg)
	assert.Equal(t, 3, h.count)

	err = h.Process(&Retract{
		Data: schema.MkInt(1),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(3, &Both{
		Retract: Retract{
			Data: schema.MkFloat(5),
		},
		Combine: Combine{
			Data: schema.MkFloat(7),
		},
	})
	assert.Equal(t, float64(7), h.avg)
	assert.Equal(t, 2, h.count)

	err = h.Process(&Both{
		Retract: Retract{
			Data: schema.MkInt(3),
		},
		Combine: Combine{
			Data: schema.MkInt(10),
		},
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(4, &Both{
		Retract: Retract{
			Data: schema.MkFloat(7),
		},
		Combine: Combine{
			Data: schema.MkFloat(10.5),
		},
	})
	assert.Equal(t, float64(10.5), h.avg)
	assert.Equal(t, 2, h.count)
}

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

func ConvertAs[A any](x schema.Schema) (A, error) {
	var a A
	var ret any
	var err error
	if any(a) == nil {
		ret, err = schema.ToGo(x)
	} else {
		ret, err = schema.ToGo(x, schema.WithExtraRules(schema.WhenPath(nil, schema.UseStruct(a))))
	}

	if err != nil {
		return a, err
	}

	result, ok := ret.(A)
	if !ok {
		return a, fmt.Errorf("cannot convert %T to %T", ret, a)
	}

	return result, nil
}

type SessionsStats struct {
	Wins  int
	Draws int
}

func TestGenericHandler(t *testing.T) {
	h := &MergeHandler[SessionsStats]{
		state: SessionsStats{},
		onCombine: func(base, x SessionsStats) (SessionsStats, error) {
			return SessionsStats{
				Wins:  base.Wins + x.Wins,
				Draws: base.Draws + x.Draws,
			}, nil
		},
		onRetract: func(base, x SessionsStats) (SessionsStats, error) {
			return SessionsStats{
				Wins:  base.Wins - x.Wins,
				Draws: base.Draws - x.Draws,
			}, nil
		},
	}

	l := &ListAssert{t: t}
	err := h.Process(&Combine{
		Data: schema.FromGo(SessionsStats{
			Wins:  1,
			Draws: 2,
		}),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(0, &Both{
		Retract: Retract{
			Data: schema.FromGo(SessionsStats{}),
		},
		Combine: Combine{
			Data: schema.FromGo(SessionsStats{
				Wins:  1,
				Draws: 2,
			}),
		},
	})
	assert.Equal(t, SessionsStats{
		Wins:  1,
		Draws: 2,
	}, h.state)
}

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

type Game struct {
	Players []string
	Winner  string
	IsDraw  bool
}

func TestMapHandler(t *testing.T) {
	m := func(x Game) (SessionsStats, error) {
		if x.IsDraw {
			return SessionsStats{
				Draws: 1,
			}, nil
		}

		if x.Winner == "" {
			return SessionsStats{}, nil
		}

		return SessionsStats{
			Wins: 1,
		}, nil
	}
	h := &MapHandler[Game, SessionsStats]{
		onCombine: m,
		onRetract: m,
	}

	l := &ListAssert{
		t: t,
	}

	err := h.Process(&Combine{
		Data: schema.FromGo(Game{
			Players: []string{"a", "b"},
			Winner:  "a",
		}),
	}, l.Append)
	assert.NoError(t, err)
	l.AssertAt(0, &Combine{
		Data: schema.FromGo(SessionsStats{
			Wins: 1,
		}),
	})
}

type GenerateHandler struct {
	load func(push func(message Message) error) error
}

func (h *GenerateHandler) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			return h.load(returning)
		},
		func(x *Retract) error {
			return fmt.Errorf("generator cannot retract")
		},
		func(x *Both) error {
			return fmt.Errorf("generator cannot bot retract and combine")
		},
	)
}

func TestGenerateHandler(t *testing.T) {
	generate := []Message{
		&Combine{
			Data: schema.FromGo(Game{
				Players: []string{"a", "b"},
				Winner:  "a",
			}),
		},
		&Combine{
			Data: schema.FromGo(Game{
				Players: []string{"a", "b"},
				Winner:  "b",
			}),
		},
		&Combine{
			Data: schema.FromGo(Game{
				Players: []string{"a", "b"},
				IsDraw:  true,
			}),
		},
	}

	h := &GenerateHandler{
		load: func(returning func(message Message) error) error {
			for idx, msg := range generate {
				err := returning(msg)
				assert.NoError(t, err, "failed to returning message at index=%d", idx)
			}
			return nil
		},
	}

	l := &ListAssert{
		t: t,
	}
	err := h.Process(&Combine{}, l.Append)
	assert.NoError(t, err)

	l.AssertLen(3)

	for idx, msg := range generate {
		l.AssertAt(idx, msg)
	}
}

type ListAssert struct {
	t     *testing.T
	Items []Message
	Err   error
}

func (l *ListAssert) Append(msg Message) error {
	if l.Err != nil {
		return l.Err
	}

	l.Items = append(l.Items, msg)
	return nil
}

func (l *ListAssert) AssertLen(expected int) bool {
	return assert.Equal(l.t, expected, len(l.Items))
}

func (l *ListAssert) AssertAt(index int, expected Message) bool {
	return assert.Equal(l.t, expected, l.Items[index])
}

func (l *ListAssert) Contains(expected Message) bool {
	for _, item := range l.Items {
		if assert.Equal(l.t, expected, item) {
			return true
		}
	}

	l.t.Errorf("expected to find %v in result set but failed", expected)
	return false
}
