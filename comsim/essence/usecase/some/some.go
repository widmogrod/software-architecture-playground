package some

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"reflect"
)

type ExecutionState struct {
	Data Data
	//Next data.AID
	End data.End
}

type TraverseResult struct {
	Value interface{}
	Stop  bool
}

func ExecuteWorkflow(w data.Workflow, state *ExecutionState) *ExecutionState {
	if state.End != nil {
		return state
	}

	switch x := w.(type) {
	case data.Activity:
		switch y := x.Activity.(type) {
		case data.Start:
			// noop
		case data.End:
			state.End = y

		case data.Choose:
			if Match(y.If, state.Data) {
				return ExecuteWorkflow(y.Then, state)
			} else {
				return ExecuteWorkflow(y.Else, state)
			}

		case data.Reshape:
			newData, err := DoReshape(y, state.Data)
			if err != nil {
				panic(err)
			}
			state.Data = newData

		case data.Invocation:
			switch y.T1 {
			case "echo":
				state.Data = MapStrAny{
					"echoed": 12,
				}
			default:
				panic(fmt.Sprintf("unknow invocation: %#v", y.T1))
			}
		default:
			panic(fmt.Sprintf("unknow ActivityT type: %#v", y))
		}

	case data.Transition:
		state = ExecuteWorkflow(x.From, state)
		state = ExecuteWorkflow(x.To, state)
	}

	return state
}

type Data = interface{}

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

type MapStrAny = map[string]interface{}

func ValueFrom(p data.Path, d Data) (interface{}, bool) {
	if len(p) == 0 {
		return nil, false
	}

	dmap, ok := d.(MapStrAny)
	if !ok {
		return nil, false
	}

	for i, key := range p {
		isLast := i == len(p)-1
		v, found := dmap[key]
		if !found {
			return nil, false
		}

		if isLast {
			return v, true
		}

		if d2, ok := v.(MapStrAny); ok {
			// is nested
			dmap = d2
		} else {
			// unfortunately cannot follow path on structure that is not nested
			return nil, false
		}
	}

	return dmap, true
}

func ValueTo(path data.Path, result MapStrAny, value interface{}) MapStrAny {
	breadcrumbs := make([]string, 0)
	dmap := result
	for i, key := range path {
		isLast := i == len(path)-1
		breadcrumbs = append(breadcrumbs, key)

		v1, foundValue := dmap[key]
		m1, foundMap := v1.(MapStrAny)

		if foundMap {
			if isLast {
				dmap[key] = value
			} else {
				dmap = m1
			}
		} else if foundValue {
			panic(fmt.Sprintf("cannot ReMap map to map. Under key '%#v' there is value that is not map: %#v", breadcrumbs, value))
		} else {
			if isLast {
				dmap[key] = value
			} else {
				m2 := make(MapStrAny)
				dmap[key] = m2
				dmap = m2
			}
		}
	}

	return result
}

func DoReshape(shape data.Reshape, value interface{}) (interface{}, error) {
	to := &simpleDataShaper{
		data: value,
	}

	_ = data.MapReshape(shape, to)

	if to.err != nil {
		return nil, to.err
	}

	return to.data, nil
}

var _ data.ReshapeVisitor = &simpleDataShaper{}

type simpleDataShaper struct {
	data interface{}
	err  error
}

func (m *simpleDataShaper) VisitSelect(x data.Select) interface{} {
	if v, ok := ValueFrom(x.Path, m.data); ok {
		m.data = v
	} else {
		m.err = fmt.Errorf("cannot match path: %v", x.Path)
		return nil
	}

	return nil
}

func (m *simpleDataShaper) VisitReMap(x data.ReMap) interface{} {
	result := make(MapStrAny)
	for i := range x {
		mapping := x[i]
		if value, ok := ValueFrom(mapping.Value, m.data); ok {
			result = ValueTo(mapping.Key, result, value)
		} else {
			m.err = fmt.Errorf("cannot map: %v", mapping)
			return nil
		}
	}
	m.data = result

	return nil
}
