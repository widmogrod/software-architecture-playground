package some

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"reflect"
)

func Do() {
	flow := data.Transition{
		T1: data.Activity{T1: data.AID("start"), T2: data.Start{}},
		T2: data.Transition{
			T1: data.Activity{
				T1: data.AID("run-echo"),
				T2: data.Invocation{
					T1: data.Fid("echo"),
				},
			},
			T2: data.Activity{
				T1: data.AID("is-ok?"),
				T2: data.Choose{
					T1: data.Eq{
						T1: data.Path([]string{"a"}),
						T2: 1,
					},
					T2: data.Activity{
						T1: data.AID("end"),
						T2: data.End{T1: data.Ok{}},
					},
					T3: data.Activity{
						T1: data.AID("error"),
						T2: data.End{T1: data.Err{}},
					},
				},
			},
		},
	}

	state := ExecutionState{
		Data: map[string]interface{}{"a": 1},
		//Next: "start",
	}

	output := ExecuteWorkflow(flow, state)
	fmt.Printf("result: %#v \n", output)
}

type ExecutionState struct {
	Data Data
	//Next data.AID
	End data.EndT
}

type TraverseResult struct {
	Value interface{}
	Stop  bool
}

func ExecuteWorkflow(w data.Workflow, state ExecutionState) ExecutionState {
	if state.End != nil {
		return state
	}

	switch x := w.(type) {
	case data.Activity:
		switch y := x.T2.(type) {
		case data.End:
			state.End = y.T1

		case data.Choose:
			if Match(y.T1, state.Data) {
				return ExecuteWorkflow(y.T2, state)
			} else {
				return ExecuteWorkflow(y.T3, state)
			}

		case data.Invocation:
			switch y.T1 {
			case "echo":
				state.Data["echoed"] = 12
			default:
				panic(fmt.Sprintf("unknow invocation: %#v", y.T1))
			}
		}

	default:
		panic(fmt.Sprintf("unknow type: %#v", x))

	case data.Transition:
		state = ExecuteWorkflow(x.T1, state)
		state = ExecuteWorkflow(x.T2, state)
	}

	return state
}

type Data = map[string]interface{}

func Match(p data.Predicate, d Data) bool {
	switch x := p.(type) {
	case data.Eq:
		if v, ok := ValueFrom(x.T1, d); ok {
			return reflect.DeepEqual(v, x.T2)
		}
		return false

	case data.Exists:
		_, found := ValueFrom(x.T1, d)
		return found

	case data.And:
		return Match(x.T1, d) && Match(x.T2, d)
	case data.Or:
		return Match(x.T1, d) || Match(x.T2, d)
	}

	panic(fmt.Sprintf("unknown predicate: %#v with: %#v", p, d))
}

func ValueFrom(p data.Path, d Data) (interface{}, bool) {
	if len(p) == 0 {
		return nil, false
	}

	for i, key := range p {
		isLast := i == len(p)-1
		v, found := d[key]
		if !found {
			return nil, false
		}

		if isLast {
			return v, true
		}

		if d2, ok := v.(Data); ok {
			// is nested
			d = d2
		} else {
			// unfortunately cannot follow path on structure that is not nested
			return nil, false
		}
	}

	return d, true
}
