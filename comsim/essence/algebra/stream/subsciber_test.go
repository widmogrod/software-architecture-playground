package stream

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"testing"
)

func TestStreamOfInvocation(t *testing.T) {
	fr := invoker.NewInMemoryFunctionRegistry()
	fr.Register(MkFunctionID("test-func"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return fmt.Sprintf("output of test func(%s)", input)
		}})

	fr.Register(MkFunctionID("test-func2"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return fmt.Sprintf("output of test func2(%s)", input)
		}})

	s := NewChannelStream()
	go s.Work()
	i := NewStreamInvoker(fr, s)
	go i.Work()

	err, r := i.Invoke(MkFunctionID("test-func"), "1312")
	assert.NoError(t, err)
	assert.Equal(t, "output of test func(1312)", r)

	err, r = i.Invoke(MkFunctionID("test-func2"), "32")
	assert.NoError(t, err)
	assert.Equal(t, "output of test func2(32)", r)

	AssertLogContains(t, s.Log(), []*Message{
		{
			Kind: "Invocation",
			Data: toBytes(Invocation{
				FID:   "test-func",
				Input: "1312",
			}),
		}, {
			Kind: "InvocationResult",
			Data: toBytes(InvocationResult{
				FID:    "test-func",
				Input:  "1312",
				Output: "output of test func(1312)",
			}),
		},
		{
			Kind: "Invocation",
			Data: toBytes(Invocation{
				FID:   "test-func2",
				Input: "32",
			}),
		}, {
			Kind: "InvocationResult",
			Data: toBytes(InvocationResult{
				FID:    "test-func2",
				Input:  "32",
				Output: "output of test func2(32)",
			}),
		},
	})
}

func AssertLogContains(t *testing.T, log, contains []*Message) {
	if !assert.Equal(t, len(contains), len(log), "Log lengths don't match") {
		return
	}

	for i, m := range log {
		assert.Equal(t, m.Kind, contains[i].Kind)

		var a, b map[string]interface{} = nil, nil

		// TODO fix assumption that message are JSON
		err := json.Unmarshal(m.Data, &a)
		assert.NoError(t, err)
		err = json.Unmarshal(contains[i].Data, &b)
		assert.NoError(t, err)
		AssertMapSubset(t, a, b)
	}
}

func AssertMapSubset(t *testing.T, amap, subset map[string]interface{}) {
	for k, v := range subset {
		assert.Equal(t, v, amap[k])
	}
}

//
//func TestSubscriber(t *testing.T) {
//	fr := invoker.NewInMemoryFunctionRegistry()
//	fr.Register("saga:reserve-availability", &invoker.FunctionInMemory{
//		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
//			return "ok"
//		}})
//	fr.Register("saga:wait-for-payment-or-cancel", &invoker.FunctionInMemory{
//		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
//			return "ok"
//		}})
//	fr.Register("saga:complete-order", &invoker.FunctionInMemory{
//		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
//			return "ok"
//		}})
//	fr.Register("sage:error-handler", &invoker.FunctionInMemory{
//		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
//			return "ok"
//		}})
//
//	w := NewWorkflow()
//	w.When("order:created", MkFunctionID("saga:reserve-availability"))
//	w.When("warehouse:reserved", MkFunctionID("saga:wait-for-payment-or-cancel"))
//	w.When("customer:charged", MkFunctionID("saga:wait-for-payment-or-cancel"))
//	w.When("delivery:shipped", MkFunctionID("saga:complete-order"))
//	w.When(MkMessageType("saga", "*", "error"), MkFunctionID("sage:error-handler"))
//
//	s := NewRandomStream()
//	cs := NewComposedStreamSubscriber()
//	cs.Source("order:created", s)
//	cs.Source("warehouse:reserved", s)
//	cs.Source("customer:charged", s)
//	cs.Source("delivery:shipped", s)
//
//	cs.Execute(w, fr)
//}
//
func MkFunctionID(s string) invoker.FunctionID {
	return s
}

//
//func NewComposedStreamSubscriber() *ComposedStream {
//	return &ComposedStream{
//		streams: make(map[string]Streamer),
//	}
//}
//
//type ComposedStream struct {
//	streams map[MessageTypeID]Streamer
//}
//
//func (s ComposedStream) Source(name string, stream Streamer) {
//	s.streams[name] = stream
//}
//
//type WorkflowContext struct {
//	//Invocations []struct{}
//	Message *Message
//}
//
////func (s ComposedStream) Execute(w *Workflow, fr invoker.FunctionRegistry) {
////	for name, s := range s.streams {
////		for _, m := range s.Fetch(1) {
////			fid := w.Flow[name]
////			_, f := fr.Get(fid)
////
////			p := WorkflowContext{
////				Message: m,
////			}
////
////			f.Call(string(toBytes(p)))
////		}
////	}
////}
//
//type (
//	MessageTypeID = string
//	Namespace     = string
//	ReturnType    = string
//)
//
//func MkMessageType(ns Namespace, fid invoker.FunctionID, rt ReturnType) MessageTypeID {
//	return ns + fid + rt
//}
//
//type Workflow struct {
//	Flow map[MessageTypeID]invoker.FunctionID
//}
//
//func (w *Workflow) When(message MessageTypeID, f invoker.FunctionID) {
//	w.Flow[message] = f
//}
//
//func NewWorkflow() *Workflow {
//	return &Workflow{
//		Flow: make(map[MessageTypeID]invoker.FunctionID),
//	}
//}
