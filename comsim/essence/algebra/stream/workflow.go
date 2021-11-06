package stream

import "github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"

func MkFunctionID(s string) invoker.FunctionID {
	return s
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
	Flow map[*SelectOnceCMD]invoker.FunctionID
}

func (w *Workflow) When(s SelectOnceCMD, f invoker.FunctionID) {
	w.Flow[&s] = f
}

func NewWorkflow() *Workflow {
	return &Workflow{
		Flow: make(map[*SelectOnceCMD]invoker.FunctionID),
	}
}
