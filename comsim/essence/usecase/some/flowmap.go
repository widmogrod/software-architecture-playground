package some

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/wokpar"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"strings"
)

func WorkparToWorkflow(in []byte) data.Workflow {
	ast := wokpar.MustParse(in)
	state := &WState{No: 0}
	return MapAstToWorkflow(*ast, state)
}

type ReduceState struct {
	workflow data.Workflow
}

type WState struct {
	No int
}

func MapAstToWorkflow(in interface{}, state *WState) data.Workflow {
	state.No++
	switch x := in.(type) {
	case wokpar.Ast:
		var result data.Workflow = data.Activity{
			Id: fmt.Sprintf("Start%d", state.No),
			Activity: data.Start{
				Var: x.Input,
			},
		}
		result = data.Transition{
			From: result,
			To:   MapAstToWorkflow(x.Body, state),
		}
		return result

	case []wokpar.Expr:
		result := MapAstToWorkflow(x[0], state)
		for i := 1; i < len(x); i++ {
			expr := x[i]
			result = data.Transition{
				From: result,
				To:   MapAstToWorkflow(expr, state),
			}
		}
		return result

	case wokpar.Expr:
		if x.Apply != nil {
			return MapAstToWorkflow(*x.Apply, state)
		}
		if x.Choose != nil {
			return MapAstToWorkflow(*x.Choose, state)
		}
		if x.Assign != nil {
			return MapAstToWorkflow(*x.Assign, state)
		}
		if x.End != nil {
			return MapAstToWorkflow(*x.End, state)
		}

	case wokpar.Assign:
		name := "_"
		if !x.Name.Ignore {
			name = *x.Name.Name
		}

		return data.Activity{
			Id: fmt.Sprintf("Assign%d", state.No),
			Activity: data.Assign{
				Var:  name,
				Flow: MapAstToWorkflow(x.Expr, state),
			},
		}
	case wokpar.Apply:
		return data.Activity{
			Id: fmt.Sprintf("Apply%d", state.No),
			Activity: data.Invocation{
				T1: x.Name,
				T2: MapSelectToReshape(x.Args),
			},
		}
	case wokpar.Choose:

		var elseFlow data.Workflow
		if len(x.Else) > 0 {
			elseFlow = MapAstToWorkflow(x.Else, state)
		}

		return data.Activity{
			Id: fmt.Sprintf("Choose%d", state.No),
			Activity: data.Choose{
				If:   MapPredicateToPredicate(x.Predicate),
				Then: MapAstToWorkflow(x.Then, state),
				Else: elseFlow,
			},
		}
	case wokpar.End:
		if x.Result != nil {
			return data.Activity{
				Id: fmt.Sprintf("Ok%d", state.No),
				Activity: data.Ok{
					T1: MapSelectToReshape(x.Result.Args),
				},
			}
		}
		if x.Fail != nil {
			return data.Activity{
				Id: fmt.Sprintf("Err%d", state.No),
				Activity: data.Err{
					T1: MapSelectToReshape(x.Fail.Args),
				},
			}
		}
		panic(fmt.Sprintf("MapAstToWorkflow. unknow end behaviour: %#v", x))
	}

	panic(fmt.Sprintf("MapAstToWorkflow. unhandled type reach: %#v", in))
}

func MapPredicateToPredicate(predicate wokpar.Predicate) data.Predicate {
	if predicate.And != nil {
		return data.And{
			T1: MapPredicateToPredicate(predicate.And.Left),
			T2: MapPredicateToPredicate(predicate.And.Right),
		}
	}
	if predicate.Or != nil {
		return data.Or{
			T1: MapPredicateToPredicate(predicate.Or.Left),
			T2: MapPredicateToPredicate(predicate.Or.Right),
		}
	}
	if predicate.Eq != nil {
		return data.Eq{
			Left: MapSelectToReshape(&predicate.Eq.Left),
			// TODO figure out how it should map
			Right: MapSelectToReshape(&predicate.Eq.Right),
		}
	}
	if predicate.Exists != nil {
		return data.Exists{
			Path: data.Path(predicate.Exists.GetValue),
		}
	}

	panic(fmt.Sprintf("MapPredicateToPredicate. unhandled type reach: %#v", predicate))
}

func MapSelectToReshape(args *wokpar.Selector) data.Reshaper {
	if args == nil {
		return nil
	}

	if args.GetValue != nil {
		return data.GetValue{
			T1: data.Path(args.GetValue),
		}
	}
	if args.SetValue != nil {
		return data.SetValue{
			T1: MapValueToValues(args.SetValue),
		}
	}

	panic(fmt.Sprintf("MapSelectToReshape. unknow type to handle %+v", args))
}

func MapValueToValues(value *wokpar.Value) data.Values {
	if value.Bool != nil {
		return data.VBool{T1: bool(*value.Bool)}
	}
	if value.Int != nil {
		return data.VInt{T1: *value.Int}
	}
	if value.Float != nil {
		return data.VFloat{T1: *value.Float}
	}
	if value.String != nil {
		// because string is `"asd"`
		// TODO figure out if creating custom laxer rule can help?
		return data.VString{T1: strings.Trim(*value.String, `"`)}
	}
	if value.List != nil {
		var result data.VList
		for i := 0; i < len(value.List); i++ {
			result = append(result, MapSelectToReshape(&value.List[i]))
		}
		return result
	}
	if value.Map != nil {
		var result data.VMap
		for i := 0; i < len(value.Map); i++ {
			result = append(result, data.VMapRecord{
				Key:   MapSelectToReshape(&value.Map[i].Key),
				Value: MapSelectToReshape(&value.Map[i].Value),
			})
		}
		return result
	}

	panic(fmt.Sprintf("MapValueToValues. unknow type to handle %+v", value))
}
