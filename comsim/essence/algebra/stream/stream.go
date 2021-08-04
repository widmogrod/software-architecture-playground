package stream

import (
	"encoding/json"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"sync"
)

type Streamer interface {
	Fetch(size int) []*Message
}

type SelectStreamer interface {
	Streamer
	Select(s SelectCMD) []*Message
}

func NewRandomStream() *RandomStream {
	return &RandomStream{}
}

type RandomStream struct {
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

func (p *RandomStream) Fetch(size int) []*Message {
	result := make([]*Message, 0, size)
	for i := 0; i < size; i++ {
		result = append(result, GenerateMessage())
	}
	return result
}

func GenerateMessage() *Message {
	return &Message{
		Data: ksuid.New().Bytes(),
	}
}

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
	panic("not implemented")
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

func NewStreamInvoker(fr invoker.FunctionRegistry, s *ChannelStream) *StreamInvoke {
	return &StreamInvoke{fr: fr, s: s}
}

type StreamInvoke struct {
	fr invoker.FunctionRegistry
	s  *ChannelStream
}

func (i *StreamInvoke) Get(name invoker.FunctionID) (error, invoker.Function) {
	return i.fr.Get(name)
}

func (i *StreamInvoke) Invoke(name invoker.FunctionID, input invoker.FunctionInput) (error, invoker.FunctionOutput) {
	ik := Invocation{
		IID:   ksuid.New().String(),
		FID:   name,
		Input: input,
	}

	m := Message{
		Kind: "Invocation",
		Data: toBytes(&ik),
	}

	// TODO push must succeed, it can be async
	// in case of failure of persisting, error must be returned
	i.s.Push(m)

	return i.Result(ik.IID)
}

func (i *StreamInvoke) Work() {
	for {
		// Here is assumption that fetch guarantees order, but don't guarantee only-once delivery
		// (1) When worker will be in the same process as invoker, that simplify few things
		// - Invocation failure can be reported in runtime that issue function invocation
		// - at-least once delivery of invocations can be de-dup in process (let's say stream had network issue and message was not-ack, and is now re-delivered)
		//
		// (2) When worker will be in different process, maybe on different machine.
		// There could me many workers that can process invocations.
		// Everything from above apply in this case as well.
		// - Invocation failure will be reported by [InvocationResult:Failure and Result method will cause runtime error
		// - de-dup may be problematic, because even with consistent hashing different worker may pick up message when processing worker fails
		//
		// - There could be situation when for one invocation message was re-delivered but to other worker, because first restared, but may process invocation.
		// 	 Which means that process that reads InvocationResult, reads only first one and discards others?
		//	 It's sign of hell getting loose, most likely for synchronous invocations (sync coordination of process) solution (1) is preferred.
		//
		//   But how does it looks for asynchronous coordination, when a process waits or return to message When(InvocationResult{FID: "order:create"}
		//   duplication of processing can happen,... but in a way invocation processing is When(Invocation{FID}}, then...
		//   that is interesing thing to sort out.
		//
		//	 Solution to this is when functions are idempotent

		for _, m := range i.s.Select(SelectCMD{
			Kind: "Invocation",
			Size: 1,
		}) {
			var mm Invocation
			err := json.Unmarshal(m.Data, &mm)
			if err != nil {
				// TODO to figure out whenever not log it as failure, and carry forward?
				panic(fmt.Sprintf("work: Invocation unmarshall; err = %v", err))
			}

			err, f := i.fr.Get(mm.FID)
			if err != nil {
				panic(fmt.Sprintf("work: function don't exists; err = %v", err))
			}

			// TODO Call may fail, so Invocation can have failure
			// Now question whenever it should be in InvocationResult{Failure:
			// or new Kind should be added?
			result := f.Call(mm.Input)

			// TODO Push also can fail, so now the question what with function that was executed?
			// Most likely such function should be idempotent,
			// that would mean that FunctionRegistry or Function per see would need to have some guarantees
			// that the same request will always land to the same "instance", end even then
			// it may be hard to lift de-duplication on runtime level, to ensure transactionality

			irk := InvocationResult{
				IID:    mm.IID,
				FID:    mm.FID,
				Input:  mm.Input,
				Output: result,
			}
			me := Message{
				Kind: "InvocationResult",
				Data: toBytes(&irk),
			}
			i.s.Push(me)

			// Fetch message needs to be ACK, otherwise when i.s.Push fails, it may not be retried
			// ACK could be also nothing else like explicit moving cursor forward
			//i.s.CursorAt(m.OffsetID)
		}
	}
}

func (i *StreamInvoke) Result(iid InvocationID) (error, invoker.FunctionOutput) {
	// TODO should control offset, or offset should be manage by stream but then consumer needs to be identified
	// TODO Fetch should allow to listen on InvocationResult but for a function that was invoke +
	// i.s.Aggregate(IID)
	// Invocation
	for _, m := range i.s.Select(SelectCMD{
		Kind: "InvocationResult",
		JSONKeyValue: map[string]SelectConditions{
			"iid": {Eq: iid},
		},
		Size: 1,
	}) {
		var ir InvocationResult
		err := json.Unmarshal(m.Data, &ir)
		if err != nil {
			return fmt.Errorf("result: InvocationResult unmarshall; err = %v", err), ""
		}

		return nil, ir.Output
	}

	return fmt.Errorf("result: no result?!!!!! iid = %v", iid), ""
}
