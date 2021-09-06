package some

import (
	"encoding/json"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"strings"
)

func WorkflowToAWSStateMachine(flow data.Workflow) (string, error) {
	state := &FlowToAwsState{
		TaskNo: 0,
		Spec: MapStrAny{
			"Comment": "",
			"StartAt": "",
			"States":  MapStrAny{},
		},
	}

	state = FlowToAws(flow, state)

	value, err := json.Marshal(state.Spec)
	return string(value), err
}

type FlowToAwsState struct {
	TaskNo      int
	Spec        MapStrAny
	CurrentTask string
	NextTask    string
	PrevTask    string
}

func FlowToAws(flow data.Workflow, state *FlowToAwsState) *FlowToAwsState {
	state.CurrentTask = ""
	state.TaskNo++

	switch x := flow.(type) {
	case data.Activity:
		switch y := x.Activity.(type) {
		case data.Start:
			state.CurrentTask = x.Id
			r := MapStrAny{
				"Type":   "Pass",
				"ResultPath": "$."+y.Var,
			}

			state.Spec["Comment"] = fmt.Sprintf("flow (%s)", y.Var)
			state.Spec["StartAt"] = state.CurrentTask
			state.Spec["States"].(MapStrAny)[state.CurrentTask] = r

			return state

		case data.End:
			switch z := y.(type) {
			case data.Ok:
				state.CurrentTask = x.Id
				//state.CurrentTask = fmt.Sprintf("Ok%d", state.TaskNo)
				if z.T1 != nil {
					r := MapStrAny{
						"Type":   "Pass",
						RetrunResherperKey(z.T1): ReshaperToAWSDataFlow(z.T1),
						"End":    true,
					}
					state.Spec["States"].(MapStrAny)[state.CurrentTask] = r
				} else {
					r := MapStrAny{
						"Type": "Succeed",
					}
					state.Spec["States"].(MapStrAny)[state.CurrentTask] = r
				}

			case data.Err:
				state.CurrentTask = x.Id
				//state.CurrentTask = fmt.Sprintf("Err%d", state.TaskNo)
				if z.T1 != nil {
					r := MapStrAny{
						"Type": "Pass",
						"End":    true,
						RetrunResherperKey(z.T1): ReshaperToAWSDataFlow(z.T1),
					}
					state.Spec["States"].(MapStrAny)[state.CurrentTask] = r
				} else {
					r := MapStrAny{
						"Type": "Fail",
					}
					state.Spec["States"].(MapStrAny)[state.CurrentTask] = r
				}
			}
			return state

		case data.Choose:
			state = FlowToAws(y.Then, state)
			choise := MapPredicateToAWS(y.If)
			choise["Next"] = state.CurrentTask

			r := MapStrAny{
				"Type": "Choice",
				"Choices": []interface{}{
					choise,
				},
			}

			// TODO
			if y.Else != nil {
				state = FlowToAws(y.Else, state)
				r["Default"] = state.CurrentTask
			}

			state.CurrentTask = x.Id
			//state.CurrentTask = fmt.Sprintf("Choise%d", state.TaskNo)
			state.Spec["States"].(MapStrAny)[state.CurrentTask] = r
			return state
		default:
			panic(fmt.Sprintf("unhandled Activity: %#v", x))
		}

	case data.Transition:
		// TODO currently state returns mutated value
		// figure out whenever we should simplify it
		state = FlowToAws(x.From, state)
		//if a, ok := x.From.(data.Activity); ok {
		//	state.CurrentTask = a.Id
		//}
		propagateCurrentState(state)

		state = FlowToAws(x.To, state)
		propagateCurrentState(state)

		if state.Spec["StartAt"] == "" {
			state.Spec["StartAt"] = state.PrevTask
		}

		if state.PrevTask != "" && state.NextTask != "" {
			state.Spec["States"].(MapStrAny)[state.PrevTask].(MapStrAny)["Next"] = state.NextTask
			state.PrevTask = state.NextTask
			state.NextTask = ""
		}

		return state

	default:
		panic(fmt.Sprintf("unhandled Workflow: %#v", flow))
	}
}

func RetrunResherperKey(t1 data.Reshaper) string {
	if isGetter(t1) {
		return "OutputPath"
	}

	if hasGetter(t1) {
		return "Parameters"
	}

	return "Parameters"
}

func ResherperKey(t1 data.Reshaper) string {
	if isGetter(t1) {
		return "InputPath"
	}

	if hasGetter(t1) {
		return "Parameters"
	}

	return "Result"
}

func propagateCurrentState(state *FlowToAwsState) {
	if state.CurrentTask != "" {
		if state.PrevTask == "" {
			state.PrevTask = state.CurrentTask
		} else if state.NextTask == "" {
			state.NextTask = state.CurrentTask
		} else {
			panic("should never reach this state")
		}
	}
}

func ReshaperToAWSDataFlow(shape data.Reshaper) interface{} {
	switch x := shape.(type) {
	case data.SetValue:
		return ValuesToAWSDataFlow(x.T1)
	case data.GetValue:
		return "$." + strings.Join(x.T1, ".")

	default:
		panic(fmt.Sprintf("unhandled Reshaper: %#v", shape))
	}
}

func ValuesToAWSDataFlow(values data.Values) interface{} {
	switch x := values.(type) {
	case data.VInt:
		return x.T1
	case data.VFloat:
		return x.T1
	case data.VBool:
		return x.T1
	case data.VString:
		return x.T1
	case data.VMap:
		result := MapStrAny{}
		for i := 0; i < len(x); i++ {
			// TODO panic ahead!
			key := ReshaperToAWSDataFlow(x[i].Key).(string)
			value := ReshaperToAWSDataFlow(x[i].Value)
			if isGetter(x[i].Value) {
				key = key + ".$"
			}

			result[key] = value
		}
		return result
	case data.VList:
		result := make([]interface{}, len(x))
		for i := 0; i < len(x); i++ {
			value := x[i]
			result[i] = ReshaperToAWSDataFlow(value)
		}
		return result
	default:
		panic(fmt.Sprintf("unhandled Values: %#v", values))
	}
}
func isGetter(value data.Reshaper) bool {
	_, ok := value.(data.GetValue)
	return ok
}

func hasGetter(value data.Reshaper) bool {
	switch x := value.(type) {
	case data.GetValue:
		return true
	case data.SetValue:
		switch y := x.T1.(type) {
		case data.VMap:
			for i := 0; i < len(y); i++ {
				if hasGetter(y[i].Value) {
					return true
				}
			}
		case data.VList:
			for i := 0; i < len(y); i++ {
				if hasGetter(y[i]) {
					return true
				}
			}
		}
	}
	return false
}

func MapPredicateToAWS(predicate data.Predicate) MapStrAny {
	switch x := predicate.(type) {
	case data.Eq:
		if _, ok := x.Left.(data.GetValue); ok {
			if _, ok := x.Right.(data.SetValue); ok {
				// Here I could do type detection,
				// since value is created, it still can reference
				return MapStrAny{
					"Variable":      ReshaperToAWSDataFlow(x.Left),
					"NumericEquals": ReshaperToAWSDataFlow(x.Right),
				}
			} else {
				// Here should be list of ors, since at this point of time
				// I cannot infer type of value to compare
				return MapStrAny{
					"Variable":          ReshaperToAWSDataFlow(x.Left),
					"NumericEqualsPath": ReshaperToAWSDataFlow(x.Right),
				}
			}
		}
		panic(fmt.Sprintf("unhandled Eq!: %#v", predicate))
	default:
		panic(fmt.Sprintf("unhandled Predicate: %#v", predicate))
	}
}
