package some

import (
	"encoding/json"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"strings"
)

func WorkflowToAWSStateMachine(flow data.Workflow) (string, error) {
	state := &FlowToAwsState{
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
	Spec           MapStrAny
	NextPrefetched *string
}

func FlowToAws(flow data.Workflow, state *FlowToAwsState) *FlowToAwsState {
	switch x := flow.(type) {
	case data.Activity:
		switch y := x.Activity.(type) {
		case data.Start:
			r := MapStrAny{
				"Type":       "Pass",
				"ResultPath": "$.__vars__." + y.Var + ".var_value",
				"Next":       *state.NextPrefetched,
			}

			state.Spec["Comment"] = fmt.Sprintf("flow (%s)", y.Var)
			state.Spec["StartAt"] = x.Id
			state.Spec["States"].(MapStrAny)[x.Id] = r

			return state

		case data.End:
			switch z := y.(type) {
			case data.Ok:
				if z.T1 != nil {
					r := MapStrAny{
						"Type":                   "Pass",
						RetrunResherperKey(z.T1): ReshaperToAWSDataFlow(z.T1),
						"End":                    true,
					}
					state.Spec["States"].(MapStrAny)[x.Id] = r
				} else {
					r := MapStrAny{
						"Type": "Succeed",
					}
					state.Spec["States"].(MapStrAny)[x.Id] = r
				}

			case data.Err:
				if z.T1 != nil {
					r := MapStrAny{
						"Type":                   "Pass",
						"End":                    true,
						RetrunResherperKey(z.T1): ReshaperToAWSDataFlow(z.T1),
					}
					state.Spec["States"].(MapStrAny)[x.Id] = r
				} else {
					r := MapStrAny{
						"Type": "Fail",
					}
					state.Spec["States"].(MapStrAny)[x.Id] = r
				}
			default:
				panic(fmt.Sprintf("unhandled End: %#v", flow))
			}
			return state

		case data.Choose:
			// (1) IF else and then branch terminate then every transition after branch will never happen
			// (2) But if any of branches don't terminate, then
			//   "Next" for last Activity in IF-THEN. branch
			//  		should be next activity outside scope of if-then-else
			//   "Default" for last Activity in ELSE. branch
			// 		    should be next activity outside scope of if-then-else
			// To handle those situations, first should one-Evaluate activity after to retrieve its ID

			choise := MapPredicateToAWS(y.If)
			if next := getNextActivityId(y.Then); next != nil {
				choise["Next"] = *next
			} else if state.NextPrefetched != nil {
				choise["Next"] = *state.NextPrefetched
			} else {
				panic(fmt.Sprintf("if-then-x does not terminate efter else statment in condition: %#v", y))
			}

			r := MapStrAny{
				"Type": "Choice",
				"Choices": []interface{}{
					choise,
				},
			}

			if y.Else != nil {
				if next := getNextActivityId(y.Else); next != nil {
					r["Default"] = *next
				} else if state.NextPrefetched != nil {
					r["Default"] = *state.NextPrefetched
				} else {
					panic(fmt.Sprintf("x-else does not terminate efter else statment in condition: %#v", y))
				}
			} else if state.NextPrefetched != nil {
				r["Default"] = *state.NextPrefetched
			} else {
				// TODO looks like flow can terminate un unknown state or in AWS lingo Cancelled!
				// What to do with this?
				// (1) I can leave it as it in hands of programmer, or or potential linter
				//		that can catch such situations
				// (2) or because I create a language that detect inconsistencies
				// 	   I should prompt programmer that something is wrong, like in this example
				//	   if statement handles only IF-THEN branch, but not ELSE and this is end of program!
				//		Function does not terminate for all cases!
				panic(fmt.Sprintf("if-then-else does not terminate efter else statment in condition: %#v", y))
			}

			state.Spec["States"].(MapStrAny)[x.Id] = r

			state = FlowToAws(y.Then, state)
			if y.Else != nil {
				state = FlowToAws(y.Else, state)
			}
			return state

		case data.Assign:
			// TODO add here check of uniqueness of assignments,
			// or to WorkparToWorkflow function
			if a, ok := y.Flow.(data.Activity); ok {
				if i, ok := a.Activity.(data.Invocation); ok {
					next := getNextActivityId(y.Flow)
					if next != nil {
						r := buildInvocation(i)

						//r["Comment"] = ToString(x)

						//  When variable is "_" then ignore result
						// TODO Because it's implicit and not express in AST
						// as well _ is valid value in paths, that means there can be "bugs"
						// that you can return(_) which should be equivalent to return()
						// I should rethink an move it to AST
						if y.Var != "_" {
							r["ResultPath"] = "$.__vars__." + y.Var
							r["ResultSelector"] = MapStrAny{
								"var_value.$": "$.Payload",
							}
						} else {
							r["ResultPath"] = "$.__vars__." + y.Var
							r["ResultSelector"] = MapStrAny{}
						}

						if state.NextPrefetched != nil {
							r["Next"] = *state.NextPrefetched
						} else {
							r["End"] = true
						}

						state.Spec["States"].(MapStrAny)[x.Id] = r
						return state
					} else {
						panic(fmt.Sprintf("nil nex-prefetch! but given: %#v", y))
					}
				}
			}

			panic(fmt.Sprintf("you can assing only result of function invocation, but given: %#v", y))

		default:
			panic(fmt.Sprintf("unhandled Activity: %#v", x.Activity))
		}

	case data.Transition:
		// In case of nested transitions, that are quite common
		prevPrefetch := state.NextPrefetched

		state.NextPrefetched = getNextActivityId(x.To)
		state = FlowToAws(x.From, state)

		state.NextPrefetched = prevPrefetch
		state = FlowToAws(x.To, state)

		return state

	default:
		panic(fmt.Sprintf("unhandled Workflow: %#v", flow))
	}
}

func buildInvocation(y data.Invocation) MapStrAny {
	// TODO make it dynamic
	parameters := MapStrAny{
		"FunctionName": "arn:aws:lambda:eu-west-1:483648412454:function:" + y.T1,
	}

	if y.T2 != nil {
		parameters["Payload"] = ReshaperToAWSDataFlow(y.T2)
	}

	r := MapStrAny{
		"Type":       "Task",
		"Resource":   "arn:aws:states:::lambda:invoke",
		"Parameters": parameters,
		"Retry": []interface{}{
			MapStrAny{
				"ErrorEquals": []interface{}{
					"Lambda.ServiceException",
					"Lambda.AWSLambdaException",
					"Lambda.SdkClientException",
				},
				"IntervalSeconds": 2,
				"MaxAttempts":     1,
				"BackoffRate":     2,
			},
		},
	}

	return r
}

func getNextActivityId(flow data.Workflow) *string {
	switch x := flow.(type) {
	case data.Activity:
		return &x.Id
	case data.Transition:
		id := getNextActivityId(x.From)
		if id != nil {
			return id
		}
		return getNextActivityId(x.To)
	default:
		panic(fmt.Sprintf("unhandled Workflow: %#v", flow))
	}

	return nil
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

func ReshaperToAWSDataFlow(shape data.Reshaper) interface{} {
	switch x := shape.(type) {
	case data.SetValue:
		return ValuesToAWSDataFlow(x.T1)
	case data.GetValue:
		return strings.TrimRight("$.__vars__."+x.T1[0]+".var_value."+strings.Join(x.T1[1:], "."), ".")

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
	case data.Or:
		return MapStrAny{
			"Or": []MapStrAny{
				MapPredicateToAWS(x.T1),
				MapPredicateToAWS(x.T2),
			},
		}
	case data.And:
		return MapStrAny{
			"And": []MapStrAny{
				MapPredicateToAWS(x.T1),
				MapPredicateToAWS(x.T2),
			},
		}
	case data.Eq:
		types := []string{"Boolean", "Numeric", "String", "Timestamp"}
		if _, ok := x.Left.(data.SetValue); ok {
			x = data.Eq{
				Left:  x.Right,
				Right: x.Left,
			}
		}

		if _, ok := x.Left.(data.GetValue); ok {
			if set, ok := x.Right.(data.SetValue); ok {
				// Here for linter’s sake I must detect type to compare
				// but I can only support types that step function supports
				switch set.T1.(type) {
				case data.VInt, data.VFloat:
					return MapStrAny{
						"Variable":      ReshaperToAWSDataFlow(x.Left),
						"NumericEquals": ReshaperToAWSDataFlow(x.Right),
					}
				case data.VBool:
					return MapStrAny{
						"Variable":      ReshaperToAWSDataFlow(x.Left),
						"BooleanEquals": ReshaperToAWSDataFlow(x.Right),
					}
				case data.VString:
					return MapStrAny{
						"Variable":     ReshaperToAWSDataFlow(x.Left),
						"StringEquals": ReshaperToAWSDataFlow(x.Right),
					}
				default:
					panic(fmt.Sprintf("not possible comparison of variable: %#v with value: %#v", x.Right, ReshaperToAWSDataFlow(x.Right)))
				}
			} else {
				// Here should be list of ors, since at this point of time
				// I cannot infer type of value to compare
				var ors []MapStrAny
				for _, typ := range types {
					ors = append(ors, MapStrAny{
						"Variable":                       ReshaperToAWSDataFlow(x.Left),
						fmt.Sprintf("%sEqualsPath", typ): ReshaperToAWSDataFlow(x.Right),
					})
				}
				return MapStrAny{
					"Or": ors,
				}
			}
		}
	}

	panic(fmt.Sprintf("unhandled Predicate: %#v", predicate))
}
