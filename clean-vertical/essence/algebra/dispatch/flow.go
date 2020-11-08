package dispatch

import (
	"fmt"
	"reflect"
	"strconv"
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

func (f *Flow) OnEffect(value interface{}) *Effect {
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

func (f *Flow) Count() int {
	counter := 0
	f.run.Visit(func(node interface{}) {
		counter++
	})

	return counter
}
func (f *Flow) Log() {
	f.run.Visit(log)

	fmt.Println("nodes=" + strconv.Itoa(f.Count()))

	f.run.Visit(func(node interface{}) {
		switch n := node.(type) {
		case *ActivityResult:
			switch n.typ {
			case InvokeA:
			case CondA:
				// ASSUMPTION on predicate
				// - first argument context value
				ctx := reflect.TypeOf(n.condition.predicate).In(0).String()
				fmt.Println(ctx)

			case EndA:
			}
		}
	})
}

func log(node interface{}) {
	typ := reflect.TypeOf(node).String()
	fmt.Printf("%s{%v}\n", typ, node)
}

func (f *Flow) Visit(visitor func(node interface{})) {
	f.run.Visit(visitor)
}

func (r *ActivityResult) Visit(visitor func(condition interface{})) {
	visitor(r)

	switch r.typ {
	case CondA:
		r.condition.thenBranch.Visit(visitor)
		if r.condition.elseBranch != nil {
			r.condition.elseBranch.Visit(visitor)
		}

	case EndA:
		// TODO check that end is the same as start aggregate!

	case InvokeA:
		// Invoke like run must bind command with result type
		typ := reflect.TypeOf(r.handler)
		cmdTyp := typ.Out(0).String()
		returnTyp := typ.Out(1).String()

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

type Effect struct {
	activity *ActivityResult
	value    interface{}
}

func (e *Effect) Activity(activity *ActivityResult) {
	e.activity = activity
}

type Condition struct {
	thenBranch *ActivityResult
	elseBranch *ActivityResult
	predicate  interface{}
	flow       *Flow
}

func (c *Condition) Then(result *ActivityResult) *Condition {
	c.thenBranch = result
	c.thenBranch.flow = c.flow

	return c
}

func (c *Condition) Else(result *ActivityResult) *ActivityResult {
	if c.thenBranch == nil {
		c.thenBranch = result
		c.thenBranch.flow = c.flow
	} else {
		c.elseBranch = result
		c.elseBranch.flow = c.flow
	}

	return c.ToActivity()
}

func (c *Condition) ToActivity() *ActivityResult {
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
