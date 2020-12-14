package dispatch

import (
	"bytes"
	"fmt"
	"reflect"
)

var flows map[string]*Flow

func init() {
	flows = make(map[string]*Flow)
}

// RetrieveFlow can rebuild workflow from a invocationID
// and thanks to that in theory make testing easier?
// TODO rethink this approach, and consider DI?
func RetrieveFlow(invocationID string) *Flow {
	return flows[invocationID]
}

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

	// TODO execution, and generation of UUID
	invocationID := "test"
	flows[invocationID] = f

	return &FlowResult{
		invocationID: invocationID,
	}
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

type plantTextState struct {
	isFirst bool
	buffer  *bytes.Buffer
}

// ToPlantText travers Flow's AST and generates plant text result
func ToPlantText(f *Flow) string {
	state := plantTextState{
		isFirst: true,
		buffer:  new(bytes.Buffer),
	}

	result := Para(func(start, next *ActivityResult, accumulator interface{}) interface{} {
		result := accumulator.(plantTextState)

		// When is first, it will be true
		isFirst := result.isFirst
		// otherwise it will always be false thanks to this line:
		result.isFirst = false

		if isFirst {
			fmt.Fprintf(state.buffer, "[*] --> %s: run \n", start.name())
		}

		switch start.typ {
		case EffectA:
			switch next.typ {
			case CondA:
				fmt.Fprintf(state.buffer, "%s --> if_%s  \n", start.name(), next.name())

			case InvokeA:
				fmt.Fprintf(state.buffer, "%s --> %s  \n", start.name(), next.name())

			case EndA:
				fmt.Fprintf(state.buffer, "%s --> [*]  \n", start.name())
			}

		case InvokeA:
			switch next.typ {
			case EffectA:
				fmt.Fprintf(state.buffer, "%s --> %s  \n", start.name(), next.name())
			}

		case CondA:
			isThen := start.condition.thenBranch == next
			switch next.typ {
			case InvokeA:
				if isThen {
					fmt.Fprintf(state.buffer, "if_%s --> %s: then \n", start.name(), next.name())
				} else {
					fmt.Fprintf(state.buffer, "if_%s --> %s: else \n", start.name(), next.name())
				}

			case EndA:
				if isThen {
					fmt.Fprintf(state.buffer, "if_%s --> [*]: then \n", start.name())
				} else {
					fmt.Fprintf(state.buffer, "if_%s --> [*]: else \n", start.name())
				}
			}
		}

		return result
	}, state, f.run)

	return result.(plantTextState).buffer.String()
}

//func ToWorkflowLog(f *Flow) []work {
//	log := make([]work, 2)
//
//	return log
//}
//
//func ExecuteWork(work work) work {
//
//}

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

//go:generate go run golang.org/x/tools/cmd/stringer -type=activityTyp
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

// With is a builder that help to build ActivityResult that compose Flow's AST
// TODO refactor to not clutter AST!
func (r *ActivityResult) With(handler interface{}) *ActivityResult {
	return &ActivityResult{
		typ:     r.typ,
		flow:    r.flow,
		handler: handler,
	}
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

type FlowResult struct {
	invocationID string
	workflowID   string
}

func (r *FlowResult) InvocationID() string {
	return r.invocationID
}

func (r *FlowResult) WorkflowID() string {
	return r.workflowID
}

// Transition can be trigger externally, schedule or invoke automatically.
// Automatic transition is default
type Transition struct {
	ID string

	// Schedule | Trigger | Invoke
	Schedule string
	Trigger  string
	Invoke   bool

	TransitionTrigger string
	FromActivityID    string
	ToActivityID      string
}

// Activity process operation, has clearly defined input and output type
// as well access to an Aggregate which represents state of whole Flow
type Activity struct {
	ID string

	// Initial - represent first state
	// Terminal - represent last state
	// Transitional - transitional state
	Initial  bool
	Terminal bool
	Regular  bool

	// Represents types, that should be mappable to runtime types
	InputType  string
	OutputType string
}

type ActivityLog struct {
	ID string

	ActivityID string

	// Pending | Processing | Ok | Err represent status of the activity
	Pending    bool
	Processing bool
	Ok         bool
	Err        bool

	// Payloads represent values that are pass to activity
	InputPayload  interface{}
	OutputPayload interface{}
}

type WorkflowState struct {
	transitions []Transition
	activities  []Activity
}

// Aggregate represents workflows, log of all transitions, retries
type Aggregate struct {
	ID string
}

func FlowToStorage(f *Flow) *WorkflowState {
	return &WorkflowState{
		transitions: []Transition{
			{
				Invoke:         true,
				FromActivityID: "initial",
				ToActivityID:   "MarkAccountActivationTokenAsUse",
			},
			{
				FromActivityID: "MarkAccountActivationTokenAsUse",
				ToActivityID:   "if_ResultOfMarkingAccountActivationTokenAsUsed",
			},
			{
				FromActivityID: "if_ResultOfMarkingAccountActivationTokenAsUsed",
				ToActivityID:   "GenerateSessionToken",
			},
			{
				FromActivityID: "if_ResultOfMarkingAccountActivationTokenAsUsed",
				ToActivityID:   "[*]",
			},
			{
				FromActivityID: "GenerateSessionToken",
				ToActivityID:   "[*]",
			},
		},
		activities: []Activity{
			{
				ID:       "initial",
				Initial:  true,
				Terminal: false,
				//Handler:  "HandleMarkAccountActivationTokenAsUse",
			},
		},
	}
}

func ToFlow(s *WorkflowState) *Flow {
	return &Flow{}
}
