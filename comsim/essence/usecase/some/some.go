package some

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"reflect"
)

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
		switch y := x.Activity.(type) {
		case data.End:
			state.End = y.Reason

		case data.Choose:
			if Match(y.If, state.Data) {
				return ExecuteWorkflow(y.Then, state)
			} else {
				return ExecuteWorkflow(y.Else, state)
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
		state = ExecuteWorkflow(x.From, state)
		state = ExecuteWorkflow(x.To, state)
	}

	return state
}

type Data = map[string]interface{}

func Match(p data.Predicate, d Data) bool {
	switch x := p.(type) {
	case data.Eq:
		if v, ok := ValueFrom(x.Path, d); ok {
			return reflect.DeepEqual(v, x.Value)
		}
		return false

	case data.Exists:
		_, found := ValueFrom(x.Path, d)
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
