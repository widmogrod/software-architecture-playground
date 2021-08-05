package stream

import (
	"encoding/json"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"math/rand"
	"sync"
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
		Data MessageData
		Kind MessageKind
		//ID   MessageID
	}
)

var _ Streamer = &ChannelStream{}

func NewChannelStream() *ChannelStream {
	return &ChannelStream{

		ch:  make(chan *Message),
		log: make([]*Message, 0, 0),

		selectors: make(map[*SelectOnceCMD]chan []*Message),

		probabilityOfRedelivery: 0.5,
	}
}

type ChannelStream struct {
	loc sync.Mutex

	ch     chan *Message
	log    []*Message
	cursor int

	selectors map[*SelectOnceCMD]chan []*Message

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
	}
)

//type selectorResult struct {
//	S SelectOnceCMD
//	R chan *Message
//}

func (c *ChannelStream) SelectOnce(s SelectOnceCMD) []*Message {
	c.loc.Lock()
	c.selectors[&s] = make(chan []*Message, 1)
	c.loc.Unlock()

	return <-c.selectors[&s]
}

func (c *ChannelStream) Work() {
	results := make(map[*SelectOnceCMD][]*Message)

	for {
		for i := c.cursor; i < len(c.log); i++ {
			m := c.log[i]

			c.loc.Lock()
			sel := c.selectors
			c.loc.Unlock()

			for s, res := range sel {
				if match(m, s) {
					results[s] = append(results[s], m)
					if len(results[s]) >= s.MaxFetchSize {
						select {
						case res <- results[s]:
						default:
						}

						c.loc.Lock()
						close(c.selectors[s])
						delete(c.selectors, s)
						c.loc.Unlock()
					}
				}
			}

			c.cursor = i
		}
	}
}

func match(m *Message, s *SelectOnceCMD) bool {
	if m == nil || s == nil {
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
	c.log = append(c.log, &message)
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
