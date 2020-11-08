package dispatch

import (
	"fmt"
	"reflect"
)

func NewFlowAround(aggregate interface{}) *Flow {
	flow := &Flow{
		effect:    map[string]*Effect{},
		aggregate: aggregate,
	}

	flow.End = &ActivityResult{
		typ:  EndA,
		flow: flow,
	}
	flow.Invoke = &ActivityResult{
		typ:  InvokeA,
		flow: flow,
	}

	return flow
}

type Flow struct {
	End       *ActivityResult
	Invoke    *ActivityResult
	effect    map[string]*Effect
	aggregate interface{}

	run *ActivityResult
}

func (f *Flow) On(value interface{}) *Effect {
	result := &Effect{
		value: value,
	}

	// ASSUMPTION: result is a struct
	typ := reflect.TypeOf(value)

	f.effect[typ.String()] = result

	return result
}

func (f *Flow) OnFailure(handler interface{}) *Flow {
	return f
}

func (f *Flow) Run(cmdFactory interface{}) *FlowResult {
	// validate
	// - does it terminate?
	// - does it initiate?
	// - does it circulate?

	f.run = &ActivityResult{
		typ:     InvokeA,
		handler: cmdFactory,
		flow:    f,
	}

	// TODO execution

	return &FlowResult{}
}

func (f *Flow) If(predicate interface{}) *Condition {
	// ASSUMPTION: predicate is a function that operates on a value and returns bool
	// in: aggregate
	// out: bool

	return &Condition{
		predicate: predicate,
		flow:      f,
	}
}

func (f *Flow) Log() {
	f.run.Visit(log)
}

func log(node interface{}) {
	switch n := node.(type) {
	case *Condition:
		fmt.Printf("Condition{%#v}\n", n)
	case *Run:
		// TODO: Run & Invoke must be a tuple
		// then extract return type from factory
		// then get Command and Result types
		// from them I have initial tranformation (Command)
		// and also additional "pending connection" for Return type to be handled!
		fmt.Printf("Run{%#v}\n", n)
	}
	typ := reflect.TypeOf(node).String()
	fmt.Printf("%s{%v}\n", typ, node)
}

func (f *Flow) Visit(visitor func(node interface{})) {
	f.run.Visit(visitor)
}
func (c *Condition) Visit(visitor func(condition interface{})) {
	//visitor("[condition ...]")
	visitor(c)
	c.left.Visit(visitor)
	if c.right != nil {
		c.right.Visit(visitor)
	}
}
func (r *ActivityResult) Visit(visitor func(condition interface{})) {
	visitor(r)

	switch r.typ {
	case CondA:
		r.condition.Visit(visitor)

	case EndA:
		// TODO check that end is the same as start aggregate!

	case InvokeA:
		// Invoke like run must bind command with result type
		typ := reflect.TypeOf(r.handler)
		cmdTyp := typ.Out(0).String()
		returnTyp := typ.Out(1).String()

		visitor(cmdTyp + " -> " + returnTyp)

		// now travers on effect
		// find a affect that corresponds to returnType,
		// when not expecting handling this effect panic()
		if effect, ok := r.flow.effect[returnTyp]; ok {
			effect.Visit(visitor)
		} else {
			panic(fmt.Sprintf(
				"flow: InvokeA activity could not find effect handler on a return type %s that is bind to command %s",
				returnTyp,
				cmdTyp,
			))
		}
	}
}
func (e *Effect) Visit(visitor func(condition interface{})) {
	// TODO activity may be empty! so chack it!!!
	e.activity.Visit(visitor)
}
func (r *Run) Visit(visitor func(condition interface{})) {
	// ASSUMPTION: about Suspend being factory function building tuple
	// - command,
	// - and binding return type
	typ := reflect.TypeOf(r.suspend)
	cmdTyp := typ.Out(0).String()
	returnTyp := typ.Out(1).String()

	// now travers on effect
	// find a affect that corresponds to returnType,
	// when not expecting handling this effect panic()
	if effect, ok := r.flow.effect[returnTyp]; ok {
		effect.Visit(visitor)
	} else {
		panic(fmt.Sprintf(
			"flow: Run could not find effect handler on a return type %s that is bind to command %s",
			returnTyp,
			cmdTyp,
		))
	}
}

type Run struct {
	suspend interface{}
	flow    *Flow
}

type Effect struct {
	activity *ActivityResult
	value    interface{}
}

func (e *Effect) With(activity *ActivityResult) {
	e.activity = activity
}

type Condition struct {
	left      *ActivityResult
	right     *ActivityResult
	predicate interface{}
	flow      *Flow
}

func (c *Condition) Then(result *ActivityResult) *Condition {
	c.left = result
	c.left.flow = c.flow

	return c
}

func (c *Condition) Else(result *ActivityResult) *ActivityResult {
	if c.left == nil {
		c.left = result
		c.left.flow = c.flow
	} else {
		c.right = result
		c.right.flow = c.flow
	}

	return c.Activity()
}

func (c *Condition) Activity() *ActivityResult {
	return &ActivityResult{
		typ:       CondA,
		condition: c,
		flow:      c.flow,
	}
}

type activityTyp uint8

const (
	EndA activityTyp = iota
	InvokeA
	CondA
)

type ActivityResult struct {
	typ       activityTyp
	flow      *Flow
	condition *Condition
	handler   interface{}
}

func (r *ActivityResult) With(handler interface{}) *ActivityResult {
	return &ActivityResult{
		typ:     r.typ,
		flow:    r.flow,
		handler: handler,
	}
}

type FlowResult struct {
	id string
}

func (r *FlowResult) InvocationID() string {
	return r.id
}
