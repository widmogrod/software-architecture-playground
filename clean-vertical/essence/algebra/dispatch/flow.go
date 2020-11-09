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

type VisitorFunc = func(condition *ActivityResult)

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
		activity: &ActivityResult{
			typ:          EffectA,
			contextValue: value,
		},
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
		flow:         f,
		typ:          InvokeA,
		handler:      cmdFactory,
		contextValue: f.aggregate,
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
	f.DepthFirstSearch(func(_ *ActivityResult) {
		counter++
	})

	return counter
}

func (f *Flow) CountBFS() int {
	counter := 0
	f.BreadthFirstSearch(func(_ *ActivityResult) {
		counter++
	})

	return counter
}

func (f *Flow) Log() {
	f.DepthFirstSearch(func(node *ActivityResult) {
		fmt.Printf("%#v\n", node)
	})

	fmt.Println("Count()    = " + strconv.Itoa(f.Count()))
	fmt.Println("CountBFS() = " + strconv.Itoa(f.CountBFS()))

	//var prev *string
	//hasPrev := false
	//previous := func(v string) {
	//	prev = &v
	//	hasPrev = true
	//}

	//a := func(a, b *ActivityResult) {
	//	switch a.typ {
	//	//case EndA:
	//	//case EffectA:
	//	case InvokeA:
	//		next := reflect.TypeOf(node.handler).In(0).Name()
	//		fmt.Printf("%s -> %s: else \n", ctx, next)
	//		//case CondA:
	//
	//	}
	//}

	//f.run.DepthFirstSearch(func(n *ActivityResult) bool {
	//	switch n.typ {
	//	case EffectA:
	//
	//	case InvokeA:
	//		typ := reflect.TypeOf(n.handler)
	//		cmdTyp := "cmd_" + typ.Out(0).Name()
	//		returnTyp := typ.Out(1).Name()
	//
	//		//typ  = reflect.TypeOf(n.contextValue)
	//		//contextTyp := typ.String()
	//
	//		if !hasPrev {
	//			fmt.Printf("[*] -> %s \n", cmdTyp)
	//		}
	//
	//		fmt.Printf("%s -> %s \n", cmdTyp, returnTyp)
	//		previous(returnTyp)
	//
	//	case CondA:
	//		// ASSUMPTION on predicate
	//		// - first argument context value
	//		if hasPrev {
	//			ctx := reflect.TypeOf(n.condition.predicate).In(0).Name()
	//			fmt.Printf("%s -> %s \n", *prev, ctx)
	//			previous(ctx)
	//
	//			nextoo := func(node *ActivityResult) bool {
	//				switch node.typ {
	//				//case EndA:
	//				//case EffectA:
	//				case InvokeA:
	//					next := reflect.TypeOf(node.handler).In(0).Name()
	//					fmt.Printf("%s -> %s: else \n", ctx, next)
	//					//case CondA:
	//
	//				}
	//				return false
	//			}
	//
	//			n.condition.thenBranch.DepthFirstSearch(nextoo)
	//			if n.condition.elseBranch != nil {
	//				n.condition.elseBranch.DepthFirstSearch(nextoo)
	//			}
	//		}
	//
	//	case EndA:
	//		if hasPrev {
	//			fmt.Printf("%s -> [*] \n", *prev)
	//		}
	//	}
	//
	//	return true
	//})
}

func (r *ActivityResult) invokeEffectActivity(fn func(result *ActivityResult)) (found bool) {
	if r.typ != InvokeA {
		return
	}

	// Invoke like run must bind command with result type
	typ := reflect.TypeOf(r.handler)
	returnTyp := typ.Out(1).String()

	// now travers on effect
	// find a affect that corresponds to returnType,
	// when not expecting handling this effect panic()
	if effect, ok := r.flow.effect[returnTyp]; ok {
		if effect.activity != nil {
			found = true
			fn(effect.activity)
		}
	}

	return
}

func (f *Flow) DepthFirstSearch(visitor VisitorFunc) {
	DepthFirstSearch(f.run, visitor)
}

func DepthFirstSearch(r *ActivityResult, visitor func(*ActivityResult)) {
	visitor(r)

	switch r.typ {
	case CondA:
		DepthFirstSearch(r.condition.thenBranch, visitor)
		if r.condition.elseBranch != nil {
			DepthFirstSearch(r.condition.elseBranch, visitor)
		}

	case EffectA:
		if r.effectActivity != nil {
			DepthFirstSearch(r.effectActivity, visitor)
		}

	case EndA:
		// TODO check that end is the same as start aggregate!

	case InvokeA:
		if !r.invokeEffectActivity(func(a *ActivityResult) {
			DepthFirstSearch(a, visitor)
		}) {
			r.panicNoEffect()
		}
	}
}

func (r *ActivityResult) panicNoEffect() {
	// Invoke like run must bind command with result type
	typ := reflect.TypeOf(r.handler)
	cmdTyp := typ.Out(0).String()
	returnTyp := typ.Out(1).String()

	panic(fmt.Sprintf(
		"flow: InvokeA activity could not find effect handler on a return type %s that is bind to command %s",
		returnTyp,
		cmdTyp,
	))
}

func (f *Flow) BreadthFirstSearch(fn func(result *ActivityResult)) {
	BreadthFirstSearch(f.run, fn)
}

func BreadthFirstSearch(start *ActivityResult, fn func(*ActivityResult)) {
	visited := make(map[*ActivityResult]bool)
	for queue := []*ActivityResult{start}; len(queue) > 0; {
		activity := queue[0]
		queue = queue[1:]

		if _, ok := visited[activity]; ok {
			continue
		}

		fn(activity)
		visited[activity] = true

		switch activity.typ {
		case CondA:
			queue = append(queue, activity.condition.thenBranch)
			if activity.condition.elseBranch != nil {
				queue = append(queue, activity.condition.elseBranch)
			}
		case InvokeA:
			if !activity.invokeEffectActivity(func(activity *ActivityResult) {
				queue = append(queue, activity)
			}) {
				activity.panicNoEffect()
			}
		case EndA:

		case EffectA:
			queue = append(queue, activity.effectActivity)
		}
	}
}

type Effect struct {
	activity *ActivityResult
	value    interface{}
}

func (e *Effect) Activity(activity *ActivityResult) {
	e.activity.effectActivity = activity
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
	EffectA
)

type ActivityResult struct {
	typ            activityTyp
	flow           *Flow
	condition      *Condition
	handler        interface{}
	contextValue   interface{}
	effectActivity *ActivityResult
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
