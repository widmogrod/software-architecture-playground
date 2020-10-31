package dispatch

func NewFlowAround(aggregate interface{}) *Flow {
	return &Flow{}
}

type Flow struct {
	End    *ActivityResult
	Invoke *ActivityResult
}

func (f *Flow) OnEffect(handler interface{}) *Flow {
	return f
}

func (f *Flow) OnFailure(handler interface{}) *Flow {
	return f
}

func (f *Flow) Run(cmdFactory interface{}) *FlowResult {
	return &FlowResult{}
}

func (f *Flow) If(predicate func() bool) *Condition {
	return &Condition{}
}

func (f *Flow) Log() {

}

type Condition struct {
}

func (c *Condition) Then(result *ActivityResult) *Condition {
	return c
}

func (c *Condition) Else(result *ActivityResult) *ActivityResult {
	return result
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
