package stream

import "github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"

func MkFunctionID(s string) invoker.FunctionID {
	return s
}

func NewComposedStreamSubscriber() *ComposedStream {
	return &ComposedStream{
		streams: make(map[string]Streamer),
	}
}

type ComposedStream struct {
	streams map[MessageTypeID]Streamer
}

func (s ComposedStream) Source(name string, stream Streamer) {
	s.streams[name] = stream
}

type WorkflowContext struct {
	//Invocations []struct{}
	Message *Message
}

func (s ComposedStream) Execute(w *Workflow, fr invoker.FunctionRegistry) {
	for name, s := range s.streams {
		for _, m := range s.Fetch(1) {
			fid := w.Flow[name]
			_, f := fr.Get(fid)

			p := WorkflowContext{
				Message: m,
			}

			f.Call(string(toBytes(p)))
		}
	}
}

type (
	MessageTypeID = string
	Namespace     = string
	ReturnType    = string
)

func MkMessageType(ns Namespace, fid invoker.FunctionID, rt ReturnType) MessageTypeID {
	return ns + fid + rt
}

type Workflow struct {
	Flow map[MessageTypeID]invoker.FunctionID
}

func (w *Workflow) When(message MessageTypeID, f invoker.FunctionID) {
	w.Flow[message] = f
}

func NewWorkflow() *Workflow {
	return &Workflow{
		Flow: make(map[MessageTypeID]invoker.FunctionID),
	}
}
