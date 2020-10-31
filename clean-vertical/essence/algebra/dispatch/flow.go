package dispatch

func NewFlowAround(aggregate interface{}) *Flow {
	return &Flow{}
}

type Flow struct {
	Ok     *ActivityResult
	Err    *ActivityResult
	Invoke *ActivityResult
}

func (f *Flow) OnEffect(handler interface{}) {

}

func (f *Flow) OnFailure(handler interface{}) {

}

func (f *Flow) Run(cmdFactory interface{}) *FlowResult {
	return &FlowResult{}
}

func (f *Flow) If(predicate func() bool) *Flow {

	return f
}

func (f *Flow) Then(result *ActivityResult) *Flow {

	return f
}

func (f *Flow) Else(result *ActivityResult) *Flow {
	return f
}

func (f *Flow) Log() {

}

type ActivityResult struct {
}

func (r *ActivityResult) With(handler interface{}) *ActivityResult {
	return r
}

type FlowResult struct {
	id string
}

func (r *FlowResult) InvocationID() string {
	return r.id
}
