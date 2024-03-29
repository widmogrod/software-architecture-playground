package some

import (
	"errors"
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"reflect"
	"strings"
)

type Invoker interface {
	Invoke(functionName string, input interface{}) (interface{}, error)
}

type ExecutionState struct {
	Data  interface{}
	Scope MapAnyAny
	// TODO remove Invoker from state
	Invoker Invoker
	// TODO introduce status of execution
}

type TraverseResult struct {
	Value interface{}
	Stop  bool
}

func ExecuteWorkflow(w data.Workflow, state *ExecutionState) *ExecutionState {
	switch x := w.(type) {
	case data.Activity:
		switch y := x.Activity.(type) {
		case data.Start:
			// noop
		case data.End:
			state.Data = nil
			switch z := y.(type) {
			case data.Ok:
				if z.T1 != nil {
					newData, err := DoReshape(z.T1, state.Scope)
					if err != nil {
						panic(err)
					}
					state.Data = newData
				}
			case data.Err:
				if z.T1 != nil {
					newData, err := DoReshape(z.T1, state.Scope)
					if err != nil {
						panic(err)
					}
					state.Data = newData
				}
			default:
				panic(fmt.Sprintf("unknow End type: %#v", z))
			}

		case data.Choose:
			if Match(y.If, state.Scope) {
				// Values in scope should not escape up
				// TODO: refactor state from pointer to make sure it works
				_ = ExecuteWorkflow(y.Then, state)
			} else {
				_ = ExecuteWorkflow(y.Else, state)
			}

		case data.Reshaper:
			newData, err := DoReshape(y, state.Scope)
			if err != nil {
				panic(err)
			}
			state.Data = newData

		case data.Assign:
			result := ExecuteWorkflow(y.Flow, state)
			if y.Var == "_" {
				break
			}
			if _, exists := state.Scope[y.Var]; exists {
				fmt.Println(y.Var)
				fmt.Println(state.Scope)
				panic(fmt.Sprintf("cannot resuse variable '%s'", y.Var))
			}
			state.Scope[y.Var] = result.Data

		case data.Invocation:
			var value interface{}
			if y.T2 != nil {
				var err error
				value, err = DoReshape(y.T2, state.Scope)
				if err != nil {
					panic(fmt.Sprintf("cannot reshape value: %#v in invocation: %#v; err=%s", y.T2, y.T1, err))
				}
			}
			result, err := state.Invoker.Invoke(y.T1, value)
			if err != nil {
				panic(fmt.Sprintf("invocation error: %s", err))
			}
			state.Data = result
		default:
			panic(fmt.Sprintf("unknow ActivityT type: %#v", y))
		}

	case data.Transition:
		state = ExecuteWorkflow(x.From, state)
		state = ExecuteWorkflow(x.To, state)
	}

	return state
}

type Data = MapAnyAny

func Match(p data.Predicate, scope MapAnyAny) bool {
	switch x := p.(type) {
	case data.Eq:
		left, err := DoReshape(x.Left, scope)
		if err != nil {
			fmt.Println("Match:err(1)", err)
			return false
		}

		right, err := DoReshape(x.Right, scope)
		if err != nil {
			fmt.Println("Match:err(2)", err)
			return false
		}

		eq := reflect.DeepEqual(left, right)
		fmt.Printf("Match:equality eq(%t,%t)test = %v \n", left, right, eq)
		return eq

	case data.Exists:
		_, found := ValueFrom(x.Path, scope)
		return found

	case data.And:
		return Match(x.T1, scope) && Match(x.T2, scope)
	case data.Or:
		return Match(x.T1, scope) || Match(x.T2, scope)
	}

	panic(fmt.Sprintf("unknown predicate: %#v with: %#v", p, scope))
}

type MapStrAny = map[string]interface{}
type MapAnyAny = map[interface{}]interface{}

func ValueFrom(p data.Path, d interface{}) (interface{}, bool) {
	if len(p) == 0 {
		return nil, false
	}

	tval := reflect.ValueOf(d)
	if tval.Kind() != reflect.Map {
		return nil, false
	}

	for i, key := range p {
		isLast := i == len(p)-1
		v := tval.MapIndex(reflect.ValueOf(key))
		if v.IsZero() {
			// not found
			return nil, false
		}

		if isLast {
			return v.Interface(), true
		}

		if !v.CanInterface() {
			return nil, false
		} else {
			v = v.Elem()
		}

		if v.Kind() == reflect.Map {
			// is nested
			tval = v
		} else {
			// unfortunately cannot follow path on structure that is not nested
			return nil, false
		}
	}

	return tval.Interface(), true
}

func DoReshape(shape data.Reshaper, scope MapAnyAny) (interface{}, error) {
	to := &simpleDataShaper{
		scope: scope,
	}

	result := data.MapReshaper(shape, to)
	if to.err != nil {
		return nil, to.err
	}

	return result, nil
}

var _ data.ReshaperVisitor = &simpleDataShaper{}

type simpleDataShaper struct {
	err   error
	scope MapAnyAny
}

func (m *simpleDataShaper) VisitSetValue(x data.SetValue) interface{} {
	value, err := ValuesToGo(x.T1, m.scope)
	if err != nil {
		m.err = err
		return nil
	}

	return value
}

func (m *simpleDataShaper) VisitGetValue(x data.GetValue) interface{} {
	value, found := ValueFrom(x.T1, m.scope)
	if !found {
		m.err = fmt.Errorf("cannot match path: %v", x.T1)
		return nil
	}

	return value
}

var ErrInputVarNotFound = errors.New("input var name not found")

func FindInputVar(flow data.Workflow) (string, error) {
	switch x := flow.(type) {
	case data.Activity:
		// Because start must be first activity
		// this is enough
		switch y := x.Activity.(type) {
		case data.Start:
			return y.Var, nil
		default:
			return "", ErrInputVarNotFound
		}

	case data.Transition:
		result, err := FindInputVar(x.From)
		if err != nil {
			result, err = FindInputVar(x.To)
		}
		return result, err
	default:
		panic(fmt.Sprintf("unknow Workflow type: %#v", x))
	}
}

func ValuesToGo(v data.Values, scope MapAnyAny) (interface{}, error) {
	visitor := &valuesT{
		scope: scope,
		err:   nil,
	}

	result := data.MapValues(v, visitor)
	return result, visitor.err
}

func MustGoValToValues(v interface{}) data.Values {
	result, err := GoValToValues(v)
	if err != nil {
		panic(err)
	}
	return result
}

func GoValToValues(v interface{}) (data.Values, error) {
	switch x := v.(type) {
	case int64:
		// TODO exloding bomb?!
		return data.VInt{T1: int(x)}, nil
	case int:
		return data.VInt{T1: x}, nil
	case bool:
		return data.VBool{T1: x}, nil
	case string:
		return data.VString{T1: x}, nil
	case []interface{}:
		result := make(data.VList, len(x))
		for i := range x {
			val, err := GoValToValues(x[i])
			if err != nil {
				return nil, fmt.Errorf("GoValToValues: %w", err)
			}
			result[i] = data.SetValue{T1: val}
		}
		return result, nil

	case map[interface{}]interface{}:
		result := data.VMap{}
		for key, val := range x {
			key2, err := GoValToValues(key)
			if err != nil {
				return nil, fmt.Errorf("GoValToValues: %w", err)
			}
			val2, err := GoValToValues(val)
			if err != nil {
				return nil, fmt.Errorf("GoValToValues: %w", err)
			}
			result = append(result, data.VMapRecord{
				Key:   data.SetValue{T1: key2},
				Value: data.SetValue{T1: val2},
			})
		}

		return result, nil
	default:
		panic(fmt.Sprintf("don't know how to convert type: %+v to data.Values", v))
	}
}

var _ data.ValuesVisitor = &valuesT{}

type valuesT struct {
	scope MapAnyAny
	err   error
}

func (v *valuesT) VisitVFloat(x data.VFloat) interface{} {
	return x.T1
}

func (v *valuesT) VisitVInt(x data.VInt) interface{} {
	return x.T1
}

func (v *valuesT) VisitVString(x data.VString) interface{} {
	return x.T1
}

func (v *valuesT) VisitVBool(x data.VBool) interface{} {
	return x.T1
}

func (v *valuesT) VisitVMap(x data.VMap) interface{} {
	if v.err != nil {
		return nil
	}

	result := make(MapAnyAny)
	for i := 0; i < len(x); i++ {
		key, err := DoReshape(x[i].Key, v.scope)
		if err != nil {
			v.err = err
			return nil
		}

		val, err := DoReshape(x[i].Value, v.scope)
		if err != nil {
			v.err = err
			return nil
		}

		if value, ok := key.(data.Values); ok {
			key = data.MapValues(value, v)
		}
		if value, ok := val.(data.Values); ok {
			val = data.MapValues(value, v)
		}

		result[key] = val
	}
	return result
}

func (v *valuesT) VisitVList(x data.VList) interface{} {
	if v.err != nil {
		return nil
	}

	result := make([]interface{}, len(x))
	for i := 0; i < len(x); i++ {
		val, err := DoReshape(x[i], v.scope)
		if err != nil {
			v.err = err
			return nil
		}
		if value, ok := val.(data.Values); ok {
			val = data.MapValues(value, v)
		}

		result[i] = val
	}

	return result
}

func ToString(value interface{}) string {
	switch x := value.(type) {
	case data.End:
		return data.MapEnd(x, &toString{}).(string)
	case data.Reshaper:
		return data.MapReshaper(x, &toString{}).(string)
	case data.Values:
		return data.MapValues(x, &toString{}).(string)
	case data.ActivityT:
		return data.MapActivityT(x, &toString{}).(string)
	case data.Predicate:
		return data.MapPredicate(x, &toString{}).(string)
	case data.Workflow:
		return data.MapWorkflow(x, &toString{}).(string)
	}

	return fmt.Sprintf("unknown(%v)", value)
}

var _ data.ActivityTVisitor = &toString{}
var _ data.EndVisitor = &toString{}
var _ data.ReshaperVisitor = &toString{}
var _ data.ValuesVisitor = &toString{}
var _ data.PredicateVisitor = &toString{}
var _ data.WorkflowVisitor = &toString{}

type toString struct{}

func (t *toString) VisitActivity(x data.Activity) interface{} {
	return fmt.Sprintf("%s", data.MapActivityT(x.Activity, t))
}

func (t *toString) VisitTransition(x data.Transition) interface{} {
	return fmt.Sprintf("%s; %s;", data.MapWorkflow(x.From, t), data.MapWorkflow(x.To, t))
}

func (t *toString) VisitEq(x data.Eq) interface{} {
	return fmt.Sprintf("eq(%s,%s)", data.MapReshaper(x.Left, t), data.MapReshaper(x.Right, t))
}

func (t *toString) VisitExists(x data.Exists) interface{} {
	return fmt.Sprintf("exists(%s)", fmt.Sprintf("$.%s", strings.Join(x.Path, ".")))
}

func (t *toString) VisitAnd(x data.And) interface{} {
	return fmt.Sprintf("and(%s,%s)", data.MapPredicate(x.T1, t), data.MapPredicate(x.T2, t))
}

func (t *toString) VisitOr(x data.Or) interface{} {
	return fmt.Sprintf("or(%s,%s)", data.MapPredicate(x.T1, t), data.MapPredicate(x.T2, t))
}

func (t *toString) VisitVFloat(x data.VFloat) interface{} {
	return fmt.Sprintf("%f", x.T1)
}

func (t *toString) VisitVInt(x data.VInt) interface{} {
	return fmt.Sprintf("%d", x.T1)
}

func (t *toString) VisitVString(x data.VString) interface{} {
	return x.T1
}

func (t *toString) VisitVBool(x data.VBool) interface{} {
	return fmt.Sprintf("%t", x.T1)
}

func (t *toString) VisitVMap(x data.VMap) interface{} {
	result := "{"
	for i := range x {
		result += fmt.Sprintf(`%s: %s, `, ToString(x[i].Key), ToString(x[i].Value))
	}
	result += "}"

	return result
}

func (t *toString) VisitVList(x data.VList) interface{} {
	result := "["
	for i := range x {
		result += ToString(x[i])
		result += ", "
	}
	result += "]"

	return result
}

func (t *toString) VisitGetValue(x data.GetValue) interface{} {
	return fmt.Sprintf("$.%s", strings.Join(x.T1, "."))
}

func (t *toString) VisitSetValue(x data.SetValue) interface{} {
	return fmt.Sprintf("Ok(%s)", data.MapValues(x.T1, t))
}

func (t *toString) VisitOk(x data.Ok) interface{} {
	return fmt.Sprintf("Ok(%s)", data.MapReshaper(x.T1, t))
}

func (t *toString) VisitErr(x data.Err) interface{} {
	return fmt.Sprintf("Err(%s)", data.MapReshaper(x.T1, t))
}

func (t *toString) VisitStart(x data.Start) interface{} {
	return fmt.Sprintf("Start(%s)", x.Var)
}

func (t *toString) VisitEnd(x data.End) interface{} {
	return data.MapEnd(x, t)
}

func (t *toString) VisitChoose(x data.Choose) interface{} {
	return fmt.Sprintf("IF(%s)", data.MapPredicate(x.If, t))
}

func (t *toString) VisitAssign(x data.Assign) interface{} {
	return fmt.Sprintf("Assign(%s,%s)", x.Var, data.MapWorkflow(x.Flow, t))
}

func (t *toString) VisitReshaper(x data.Reshaper) interface{} {
	return data.MapReshaper(x, t)
}

func (t *toString) VisitInvocation(x data.Invocation) interface{} {
	if x.T2 != nil {
		return fmt.Sprintf("%s(%s)", x.T1, data.MapReshaper(x.T2, t))
	}
	return fmt.Sprintf("%s()", x.T1)
}
