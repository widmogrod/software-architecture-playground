package stream

import (
	"encoding/json"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"sync"
)

type Streamer interface {
	Push(message Message)
	Fetch(size int) []*Message
}

type SelectStreamer interface {
	Streamer
	Select(s SelectCMD) []*Message
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

		selectors: make([]*SelectCMD, 0),
		results:   make(map[*SelectCMD]chan *Message),

		probabilityOfRedelivery: 0.5,
	}
}

type ChannelStream struct {
	loc sync.Mutex

	ch  chan *Message
	log []*Message

	selectors []*SelectCMD
	results   map[*SelectCMD]chan *Message

	probabilityOfRedelivery float64
}

type (
	SelectConditions struct {
		Eq interface{}
	}

	SelectCMD struct {
		Kind         MessageKind
		JSONKeyValue map[string]SelectConditions
		Size         int
	}
)

func (c *ChannelStream) Select(s SelectCMD) []*Message {
	c.loc.Lock()
	c.selectors = append(c.selectors, &s)
	c.results[&s] = make(chan *Message)
	c.loc.Unlock()

	var result []*Message
	for {
		select {
		case m := <-c.results[&s]:
			result = append(result, m)

			if len(result) >= s.Size {
				c.loc.Lock()
				close(c.results[&s])
				delete(c.results, &s)
				c.loc.Unlock()

				return result
			}
		}
	}
}

func (c *ChannelStream) Work() {
	for {
		select {
		case m := <-c.ch:
			for _, s := range c.selectors {
				if m.Kind != s.Kind {
					fmt.Printf("select: Kind... selector(%v) message(%v) \n", s, m)
					continue
				}

				if s.JSONKeyValue == nil {
					go c.funcName(s, m)
					fmt.Printf("select: FOUND.JSONKeyValue... selector(%v) message(%v) \n", s, m)
					continue
				}

				var a map[string]interface{} = nil
				err := json.Unmarshal(m.Data, &a)
				if err != nil {
					fmt.Printf("select: Unmarshal... selector(%v) message(%v) err = %s \n", s, m, err)
					continue
				}

				var found bool = true
				for key, cond := range s.JSONKeyValue {
					if !found {
						break
					}
					if value, ok := a[key]; ok {
						if cond.Eq != nil {
							found = cond.Eq == value
						}
					}
				}

				if found {
					fmt.Printf("select: FOUND ... selector(%v) message(%v) \n", s, m)
					go c.funcName(s, m)
				}
			}
		}
	}
}

func (c *ChannelStream) funcName(s *SelectCMD, m *Message) {
	c.results[s] <- m
}

func (c *ChannelStream) Fetch(size int) []*Message {
	return c.log[len(c.log)-size:]

	//panic("not implemented")
	//if len(c.log) >= size {
	//	// Simulate message re-delivery
	//	if rand.Float64() < c.probabilityOfRedelivery {
	//		return c.log[len(c.log)-size:]
	//	}
	//}
	//
	//var result []*Message
	//for {
	//	select {
	//	case m := <-c.ch:
	//		result = append(result, m)
	//		if len(result) == size {
	//			return result
	//		}
	//	}
	//}
}

func (c *ChannelStream) Push(message Message) {
	c.log = append(c.log, &message)
	c.ch <- &message
}

func (c *ChannelStream) Log() []*Message {
	return c.log
}

type (
	InvocationID = string

	Invocation struct {
		IID   InvocationID `json:"iid,omitempty"`
		FID   invoker.FunctionID
		Input invoker.FunctionInput
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
	res, _ := json.Marshal(p)
	return res
}
