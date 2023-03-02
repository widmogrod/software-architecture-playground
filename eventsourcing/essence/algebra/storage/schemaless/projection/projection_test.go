package schemaless

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/mkunion/x/schema"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless"
	"github.com/widmogrod/software-architecture-playground/eventsourcing/essence/algebra/storage/schemaless/typedful"
	"sync"
	"testing"
	"time"
)

var generateData = []Message{
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

func GenerateData() Handler {
	return &GenerateHandler{
		load: func(returning func(message Message) error) error {
			for _, msg := range generateData {
				if err := returning(msg); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func MapGameToStats() Handler {
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

	return &MapHandler[Game, SessionsStats]{
		onCombine: m,
		onRetract: m,
	}
}

func MapGameStatsToSession() Handler {
	return &MergeHandler[SessionsStats]{
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
}

func TestProjection(t *testing.T) {
	store := schemaless.NewInMemoryRepository()
	typed := typedful.NewTypedRepository[SessionsStats](store)

	dag := NewBuilder()
	games := dag.Load(GenerateData())
	gameStats := games.Map(MapGameToStats())
	gameStatsBySession := gameStats.Merge(MapGameStatsToSession())

	end := gameStatsBySession.Map(NewRepositorySink("session", store))

	//end := gameStatsBySession.Map(Log())

	//expected := &Map{
	//	OnMap: Log(),
	//	Input: &Merge{
	//		OnMerge: MapGameStatsToSession(),
	//		Input: []DAG{
	//			&Map{
	//				OnMap: MapGameToStats(),
	//				Input: &Map{
	//					OnMap: GenerateData(),
	//					Input: nil,
	//				},
	//			},
	//		},
	//	},
	//}
	//assert.Equal(t, expected, end.Build())

	interpretation := NewInMemoryInterpreter()
	err := interpretation.Run(end.Build())
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, len(interpretation.errors), "interpretation should be without errors")

	result, err := typed.FindingRecords(schemaless.FindingRecords[schemaless.Record[SessionsStats]]{})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 1)

	for _, x := range result.Items {
		fmt.Printf("item: %#v\n", x)
	}
}

type InMemoryInterpreter struct {
	lock     sync.Mutex
	channels map[DAG]chan Message
	errors   map[DAG]error
}

func (i *InMemoryInterpreter) Run(dag DAG) error {
	if dag == nil {
		return nil
	}

	return MustMatchDAG(
		dag,
		func(x *Map) error {
			go func() {
				//fmt.Printf("Map: gorutine starting %T\n", x)
				for {
					select {
					case msg := <-i.channelForNode(x.Input):
						//fmt.Printf("Map: recieved %T msg=%v\n", x, msg)
						if err := x.OnMap.Process(msg, i.returning(x)); err != nil {
							i.recordError(x, err)
							return
						}
						//case <-time.After(1 * time.Second):
						//	fmt.Printf("Map: timeout for %T\n", x)
					}
				}
			}()
			return i.Run(x.Input)
		},
		func(x *Merge) error {
			go func() {
				//fmt.Printf("Merge: gorutine starting %T\n", x)
				for {
					select {
					case msg := <-i.channelForNode(x.Input[0]):
						//fmt.Printf("Merge: recieved %T msg=%v\n", x, msg)
						if err := x.OnMerge.Process(msg, i.returning(x)); err != nil {
							i.recordError(x, err)
							return
						}
						//case <-time.After(1 * time.Second):
						//	fmt.Printf("Merge: timeout for %T\n", x)
					}
				}
			}()
			return i.Run(x.Input[0])
		},
		func(x *Load) error {
			go func() {
				//fmt.Printf("Load: gorutine starting %T\n", x)
				if err := x.OnLoad.Process(&Combine{}, i.returning(x)); err != nil {
					i.recordError(x, err)
					return
				}
			}()

			return nil
		},
	)
}

func (i *InMemoryInterpreter) returning(x DAG) func(Message) error {
	return func(msg Message) error {
		i.channelForNode(x) <- msg
		return nil
	}
}

func (i *InMemoryInterpreter) channelForNode(x DAG) chan Message {
	i.lock.Lock()
	defer i.lock.Unlock()
	if _, ok := i.channels[x]; !ok {
		i.channels[x] = make(chan Message)
	}
	return i.channels[x]
}

func (i *InMemoryInterpreter) recordError(x DAG, err error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	//fmt.Printf("element %v error %s", x, err)
	i.errors[x] = err
}

func NewInMemoryInterpreter() *InMemoryInterpreter {
	return &InMemoryInterpreter{
		channels: make(map[DAG]chan Message),
		errors:   make(map[DAG]error),
	}
}

func NewBuilder() *DagBuilder {
	return &DagBuilder{}
}

type DagBuilder struct {
	dag DAG
}

func (b *DagBuilder) Map(handler Handler) *DagBuilder {
	return &DagBuilder{
		dag: &Map{
			OnMap: handler,
			Input: b.dag,
		},
	}
}

func (b *DagBuilder) Merge(handler Handler) *DagBuilder {
	return &DagBuilder{
		dag: &Merge{
			OnMerge: handler,
			Input:   []DAG{b.dag},
		},
	}
}

func (b *DagBuilder) Build() DAG {
	return b.dag
}

func (b *DagBuilder) Load(data Handler) *DagBuilder {
	return &DagBuilder{
		dag: &Load{
			OnLoad: data,
		},
	}
}

type LogHandler struct{}

func (l *LogHandler) Process(msg Message, returning func(Message) error) error {
	return MustMatchMessage(
		msg,
		func(x *Combine) error {
			res, err := schema.ToJSON(x.Data)
			if err != nil {
				return err
			}
			fmt.Printf("Log: Combine(%s) \n", res)
			return nil
		},
		func(x *Retract) error {
			res, err := schema.ToJSON(x.Data)
			if err != nil {
				return err
			}
			fmt.Printf("Log: Retract(%s) \n", res)
			return nil

		},
		func(x *Both) error {
			fmt.Printf("Log: Both(\n")
			fmt.Printf("\t")
			_ = l.Process(&x.Retract, returning)
			fmt.Printf("\t")
			_ = l.Process(&x.Combine, returning)
			fmt.Printf(") Both end\n")
			return nil
		},
	)
}

func Log() Handler {
	return &LogHandler{}
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	}, l.Returning)
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
	err := h.Process(&Combine{}, l.Returning)
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

func (l *ListAssert) Returning(msg Message) error {
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

func NewRepositorySink(recordType string, store schemaless.Repository[schema.Schema]) *RepositorySink {
	sink := &RepositorySink{
		flushWhenBatchSize: 10,
		flushWhenDuration:  1 * time.Second,

		store:      store,
		recordType: recordType,

		bufferSaving:   map[string]schemaless.Record[schema.Schema]{},
		bufferDeleting: map[string]schemaless.Record[schema.Schema]{},
	}

	sink.FlushOnTime()

	return sink
}

type RepositorySink struct {
	flushWhenBatchSize int
	flushWhenDuration  time.Duration

	bufferSaving   map[string]schemaless.Record[schema.Schema]
	bufferDeleting map[string]schemaless.Record[schema.Schema]

	store      schemaless.Repository[schema.Schema]
	recordType string
}

func (s *RepositorySink) FlushOnTime() {
	go func() {
		ticker := time.NewTicker(s.flushWhenDuration)
		for range ticker.C {
			s.flush()
		}
	}()
}

func (s *RepositorySink) Process(msg Message, returning func(Message) error) error {
	err := MustMatchMessage(
		msg,
		func(x *Combine) error {
			s.bufferSaving[x.Key] = schemaless.Record[schema.Schema]{
				ID:      x.Key,
				Type:    s.recordType,
				Data:    x.Data,
				Version: 0,
			}
			return nil
		},
		func(x *Retract) error {
			s.bufferDeleting[x.Key] = schemaless.Record[schema.Schema]{
				ID:      x.Key,
				Type:    s.recordType,
				Data:    x.Data,
				Version: 0,
			}
			return nil
		},
		func(x *Both) error {
			s.bufferSaving[x.Key] = schemaless.Record[schema.Schema]{
				ID:      x.Key,
				Type:    s.recordType,
				Data:    x.Combine.Data,
				Version: 0,
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	if len(s.bufferSaving)+len(s.bufferDeleting) >= s.flushWhenBatchSize {
		return s.flush()
	}

	return nil

}

func (s *RepositorySink) flush() error {
	if len(s.bufferSaving)+len(s.bufferDeleting) == 0 {
		return nil
	}

	err := s.store.UpdateRecords(schemaless.UpdateRecords[schemaless.Record[schema.Schema]]{
		Saving:   s.bufferSaving,
		Deleting: s.bufferDeleting,
	})
	if err != nil {
		return err
	}

	s.bufferSaving = map[string]schemaless.Record[schema.Schema]{}
	s.bufferDeleting = map[string]schemaless.Record[schema.Schema]{}
	return nil

}
