package stream

import (
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"testing"
)

func TestSubscriber(t *testing.T) {
	fr := invoker.NewInMemoryFunctionRegistry()
	fr.Register("saga:reserve-availability", &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return "ok"
		}})
	fr.Register("saga:wait-for-payment-or-cancel", &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return "ok"
		}})
	fr.Register("saga:complete-order", &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return "ok"
		}})
	fr.Register("sage:error-handler", &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return "ok"
		}})

	w := NewWorkflow()
	w.When("order:created", MkFunctionID("saga:reserve-availability"))
	w.When("warehouse:reserved", MkFunctionID("saga:wait-for-payment-or-cancel"))
	w.When("customer:charged", MkFunctionID("saga:wait-for-payment-or-cancel"))
	w.When("delivery:shipped", MkFunctionID("saga:complete-order"))
	w.When(MkMessageType("saga", "*", "error"), MkFunctionID("sage:error-handler"))

	s := NewRandomStream()
	cs := NewComposedStreamSubscriber()
	cs.Source("order:created", s)
	cs.Source("warehouse:reserved", s)
	cs.Source("customer:charged", s)
	cs.Source("delivery:shipped", s)

	cs.Execute(w, fr)
}
