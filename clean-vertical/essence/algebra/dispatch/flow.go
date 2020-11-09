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

func (a *ActivityResult) name() string {
	switch a.typ {
	case EffectA:
		return reflect.TypeOf(a.contextValue).Name()
	case InvokeA:
		return reflect.TypeOf(a.handler).Out(0).Name()
	case CondA:
		return reflect.TypeOf(a.condition.predicate).In(0).Name()
	case EndA:
		return "[*]"
	}

	return "unknown activity name!"
}

func (f *Flow) Log() {
	fmt.Println("Count()    = " + strconv.Itoa(f.Count()))
	fmt.Println("CountBFS() = " + strconv.Itoa(f.CountBFS()))

	isFirst := true
	Para(func(start, next *ActivityResult, accumulator interface{}) interface{} {
		isFirst := accumulator.(bool)
		if isFirst {
			fmt.Printf("[*] -> %s: run \n", start.name())
		}

		switch start.typ {
		case EffectA:
			name := reflect.TypeOf(start.contextValue).Name()
			switch next.typ {
			case CondA:
				to := reflect.TypeOf(next.condition.predicate).In(0).Name()
				fmt.Printf("%s -> if_%s  \n", name, to)

			case InvokeA:
				to := reflect.TypeOf(start.handler).Out(0).Name()
				fmt.Printf("%s -> %s  \n", name, to)

			case EndA:
				fmt.Printf("%s -> [*]  \n", name)
			}

		case InvokeA:
			name := reflect.TypeOf(start.handler).Out(0).Name()

			switch next.typ {
			case EffectA:
				to := reflect.TypeOf(next.contextValue).Name()
				fmt.Printf("%s -> %s  \n", name, to)
			}

		case CondA:
			name := "if_" + reflect.TypeOf(start.condition.predicate).In(0).Name()
			isThen := start.condition.thenBranch == next

			switch next.typ {
			case InvokeA:
				to := reflect.TypeOf(next.handler).Out(0).Name()
				if isThen {
					fmt.Printf("%s -> %s: then \n", name, to)
				} else {
					fmt.Printf("%s -> %s: else \n", name, to)
				}

			case EndA:
				if isThen {
					fmt.Printf("%s -> [*]: then \n", name)
				} else {
					fmt.Printf("%s -> [*]: else \n", name)
				}
			}
		}

		return false
	}, isFirst, f.run)
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

func (f *Flow) DepthFirstSearch(visitor VisitorFunc) {
	DepthFirstSearch(f.run, visitor)
}

func DepthFirstSearch(r *ActivityResult, visitor func(*ActivityResult)) {
	// TODO introduce check to visited nodes
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

type paramorphism = func(a, b *ActivityResult, accumulator interface{}) interface{}

// Para is a paramorphism that will reduce  Flow AST to new algebra,
// and during reduction provide context as well with accumulator
// ```haskell
// para :: (a -> ([a], b) -> b) -> b -> [a] ->  b
// ```
func Para(fn paramorphism, accumulator interface{}, start *ActivityResult) interface{} {
	res := accumulator
	switch start.typ {
	case CondA:
		res = Para(fn, fn(start, start.condition.thenBranch, accumulator), start.condition.thenBranch)
		if start.condition.elseBranch != nil {
			res = Para(fn, fn(start, start.condition.elseBranch, accumulator), start.condition.elseBranch)
		}

	case EffectA:
		if start.effectActivity != nil {
			res = Para(fn, fn(start, start.effectActivity, accumulator), start.effectActivity)
		}

	case EndA:
		// noop

	case InvokeA:
		if !start.invokeEffectActivity(func(a *ActivityResult) {
			res = Para(fn, fn(start, a, accumulator), a)
		}) {
			start.panicNoEffect()
		}
	}

	return res
}

// Effect is a builder that help to build ActivityResult that compose Flow's AST
// TODO consider refactoring to be the builder, right now part is setup in method OnEffect() in Flow
type Effect struct {
	activity *ActivityResult
	value    interface{}
}

func (e *Effect) Activity(activity *ActivityResult) {
	e.activity.effectActivity = activity
}

// Condition is a builder that help to build ActivityResult that compose Flow's AST
// TODO make it separate more clearly, now AST knows about Condition!
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

func (t activityTyp) String() string {
	switch t {
	case EndA:
		return "EndA"
	case InvokeA:
		return "InvokeA"
	case EffectA:
		return "EffectA"
	case CondA:
		return "CondA"
	}

	return "unknown activity type!"
}

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

// With is a builder that help to build ActivityResult that compose Flow's AST
// TODO refactor to not clutter AST!
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
