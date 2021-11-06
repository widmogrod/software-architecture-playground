package stream

import (
	"encoding/json"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"math/rand"
	"sync"
	"time"
)

type Streamer interface {
	Push(message Message)
	Fetch(size int) []*Message
}

type SelectStreamer interface {
	Streamer
	SelectOnce(s SelectOnceCMD) []*Message
}

type (
	//MessageID   = string
	MessageKind = string
	MessageData = []byte

	Message struct {
		Data   MessageData
		Kind   MessageKind
		Cursor *Cursor
		//ID   MessageID
	}
)

var _ Streamer = &ChannelStream{}

func NewChannelStream() *ChannelStream {
	return &ChannelStream{
		ch:                      make(chan *Message),
		log:                     make([]*Message, 0, 0),
		next:                    make(chan struct{}),
		probabilityOfRedelivery: 0.5,
	}
}

type ChannelStream struct {
	ch     chan *Message
	log    []*Message
	next   chan struct{}
	cursor int

	selectors sync.Map

	probabilityOfRedelivery float64
}

type (
	SelectConditions struct {
		// selector   = Every [match]
		// match      = KeyValue (string, match) | KeyExists(string) | conditions |
		// conditions = eq | gt | lt | ...
		Eq        interface{}
		KeyExists string
		KeyValue  map[string]SelectConditions
	}

	SelectOnceCMD struct {
		Kind         MessageKind
		Selector     *SelectConditions
		MaxFetchSize int

		// Position oriented traversing of logO
		From *Cursor
		Skip int
	}
)

type Cursor struct {
	Position int
}

func (c *ChannelStream) SelectOnce(s SelectOnceCMD) []*Message {
	ch := make(chan []*Message)
	defer close(ch)

	c.selectors.Store(s, ch)
	defer c.selectors.Delete(s)

	c.nextIteration()

	return <-ch
}

type (
	ReduceCMD struct {
	}
)

func (c *ChannelStream) Reduce(r ReduceCMD) {

}

func (c *ChannelStream) Work() {
	tick := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-c.next:
		case <-tick.C:
		}

		c.selectors.Range(func(key, value interface{}) bool {
			s := key.(SelectOnceCMD)
			res := value.(chan []*Message)

			maxFetch := s.MaxFetchSize
			if s.MaxFetchSize == 0 {
				maxFetch = 1
			}

			cursor := 0
			if s.From != nil {
				cursor = s.From.Position + s.Skip
			}

			results := make([]*Message, 0, 0)
			for i := cursor; i < len(c.log); i++ {
				m := c.log[i]

				if match(m, s) {
					results = append(results, m)

					if len(results) >= maxFetch {
						select {
						case res <- results[0:maxFetch]:
						default:
						}

						// Clear
						c.selectors.Delete(key)
						break
					}
				}
			}

			return true
		})
	}
}

func match(m *Message, s SelectOnceCMD) bool {
	if m == nil {
		return false
	}

	if m.Kind != s.Kind {
		return false
	}

	if s.Selector == nil {
		return true
	}

	if m.Data == nil {
		return false
	}

	var a map[string]interface{} = nil
	err := json.Unmarshal(m.Data, &a)
	if err != nil {
		fmt.Printf("select: Unmarshal... selector(%v) message(%v) err = %s \n", s, m, err)
		return false
	}

	if s.Selector.KeyExists != "" {
		_, ok := a[s.Selector.KeyExists]
		return ok
	}

	return matchNested(s.Selector, a)
}

func matchNested(s *SelectConditions, a map[string]interface{}) bool {
	var found bool = true
	for key, cond := range s.KeyValue {
		if !found {
			break
		}
		if value, ok := a[key]; ok {
			if cond.Eq != nil {
				found = cond.Eq == value
				continue
			}
			if cond.KeyValue != nil {
				if v, ok := value.(map[string]interface{}); ok {
					found = matchNested(&cond, v)
					continue
				}
			}
			if cond.KeyExists != "" && cond.KeyExists != key {
				found = false
				continue
			}
		}
	}

	return found
}

func (c *ChannelStream) Fetch(size int) []*Message {
	if len(c.log) >= size {
		// Simulate message re-delivery
		if rand.Float64() < c.probabilityOfRedelivery {
			return c.log[len(c.log)-size:]
		}
	}

	return c.log[len(c.log)-size:]
}

func (c *ChannelStream) Push(message Message) {
	m := &Message{
		Data:   message.Data,
		Kind:   message.Kind,
		Cursor: &Cursor{Position: len(c.log)},
	}
	c.log = append(c.log, m)

	c.nextIteration()
}

func (c *ChannelStream) nextIteration() {
	select {
	case c.next <- struct{}{}:
	default:
	}
}

func (c *ChannelStream) Log() []*Message {
	return c.log
}

type (
	InvocationID = string

	Invocation struct {
		IID   InvocationID          `json:"iid,omitempty"`
		FID   invoker.FunctionID    `json:"fid"`
		Input invoker.FunctionInput `json:"input"`
	}

	InvocationResult struct {
		IID     InvocationID `json:"iid,omitempty"`
		FID     invoker.FunctionID
		Input   invoker.FunctionInput
		Output  invoker.FunctionOutput
		Failure error
	}
)

func toBytes(p interface{}) []byte {
	res, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return res
}
