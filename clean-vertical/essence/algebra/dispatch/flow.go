package dispatch

import (
	"fmt"
	"reflect"
)

func NewFlowAround(aggregate interface{}) *Flow {
	flow := &Flow{
		//graph:  NewGraph(),
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
	End    *ActivityResult
	Invoke *ActivityResult
	effect map[string]*Effect
	//graph  *Graph
	aggregate interface{}

	run *Run
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

//func (f *Flow) OnEffect(handler interface{}) *Flow {
//	// ASSUMPTION: about effect being function that accepts result of computation a.k.a. event
//
//	typ := reflect.TypeOf(handler)
//	resultTyp := typ.In(0)
//
//	// create stub of a event to alo evaluation of activity logic
//	reflect.ValueOf(resultTyp)
//
//	// Assigning effect should also trigger evaluation of it [?]
//	f.effect[resultTyp.Name()] = handler()
//
//	effectOriginType := "usecase.MarkAccountActivationTokenAsUse"
//	effectType := "usecase.ResultOfMarkingAccountActivationTokenAsUsed"
//	f.graph.AddEdge("effect", effectOriginType, effectType)
//
//	conditionOrigin := "usecase.ResultOfMarkingAccountActivationTokenAsUsed"
//	conditionThen := "end(usecase.ResultOfConfirmationOfAccountActivation)"
//	conditionElse := "invoke(usecase.GenerateSessionToken)"
//	conditionNode := "condition(ResultOfMarkingAccountActivationTokenAsUsed)"
//
//	f.graph.AddEdge("condition", conditionOrigin, conditionNode)
//	f.graph.AddEdge("condition-then", conditionNode, conditionThen)
//	f.graph.AddEdge("condition-else", conditionNode, conditionElse)
//
//	// What alternative way I can see is with constructing AST
//	//
//
//	return f
//}

func (f *Flow) OnFailure(handler interface{}) *Flow {
	return f
}

func (f *Flow) Run(cmdFactory interface{}) *FlowResult {
	// in theory we have computation graph at this stage, now we need to fill last bits
	// validate it
	// - does it terminate?
	// - does it initiate?
	// - does it circulate?
	// and execute

	//// extract cmd type
	//cmdType := "usecase.MarkAccountActivationTokenAsUse"
	//// extract result type of cmd
	//cmdResultType := "usecase.ResultOfMarkingAccountActivationTokenAsUsed"
	//f.graph.AddEdge("cmd-result", cmdType, cmdResultType)
	//
	//f.graph.AddEdge("init", "start", cmdType)

	f.run = &Run{
		suspend: cmdFactory,
		flow:    f,
	}

	return &FlowResult{}
}

func (f *Flow) If(predicate interface{}) *Condition {
	// ASSUMPTION: predicate is a function that operates on a value and returns bool
	// in: aggregate
	// out: bool

	return &Condition{
		flow: f,
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
	c.right.Visit(visitor)
}
func (c *ConditionBranch) Visit(visitor func(condition interface{})) {
	visitor(c)
	c.activity.Visit(visitor)
}
func (r *ActivityResult) Visit(visitor func(condition interface{})) {
	//if r == nil {
	//	visitor("ERR! empty activity!")
	//	return
	//}

	//visitor("[activity " + strconv.Itoa(int(r.typ)) + "] ... ")
	visitor(r)

	switch r.typ {
	case CondA:
		r.condition.Visit(visitor)

	case EndA:
		//visitor("[end]")
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
	//typ := reflect.TypeOf(e.value)
	//eventTyp := typ.String()
	//visitor(eventTyp + " -> [activity ...]")

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

	//visitor(cmdTyp + " -> " + returnTyp)

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

type ConditionBranch struct {
	activity *ActivityResult
	flow     *Flow
}

type Condition struct {
	left  *ConditionBranch
	right *ConditionBranch
	flow  *Flow
}

func (c *Condition) Then(result *ActivityResult) *Condition {
	c.left = &ConditionBranch{
		activity: result,
		flow:     c.flow,
	}

	return c
}

func (c *Condition) Else(result *ActivityResult) *ActivityResult {
	if c.left == nil {
		panic(fmt.Sprintf("flow: to use Else() you must use Then() before"))
	}

	c.right = &ConditionBranch{
		activity: result,
		flow:     c.flow,
	}

	return &ActivityResult{
		typ:       CondA,
		condition: c,
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
