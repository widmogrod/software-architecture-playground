package some

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/stream"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"testing"
)

var _ Invoker = &jsonInvoker{}

type FunctionID = string
type FunctionInput = string
type FunctionOutput = string

type jsonInvoker struct {
	Call func(name FunctionID, input FunctionInput) (FunctionOutput, error)
}

func (j *jsonInvoker) Invoke(functionName string, input interface{}) (interface{}, error) {
	in, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("invoke. Cannot conver input to JSON: %w", err)
	}

	out, err := j.Call(functionName, string(in))
	if err != nil {
		return nil, err
	}

	var res interface{}
	err = json.Unmarshal([]byte(out), &res)
	if err != nil {
		return nil, fmt.Errorf("invoke. Cannot conver output to JSON: %w", err)
	}

	return res, nil
}

func TestExecuteWorkflow(t *testing.T) {
	fr := invoker.NewInMemoryFunctionRegistry()
	fr.Register(stream.MkFunctionID("echo"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			var i int64
			err := json.Unmarshal([]byte(input), &i)
			if err != nil {
				panic(err)
			}

			result := MapStrAny{
				"echoed": i,
			}

			output, err := json.Marshal(result)
			if err != nil {
				panic(err)
			}
			return string(output)
		}})

	// TODO create JSON invoker
	invoke := invoker.NewInvoker(fr)
	jsonInvoke := &jsonInvoker{
		Call: invoke.Invoke,
	}

	flow := data.Transition{
		From: data.Activity{
			Id: data.AID("start"),
			Activity: data.Start{
				Var: "input",
			},
		},
		To: data.Transition{
			From: data.Activity{
				Id: data.AID("Passing-to-echo"),
				Activity: data.Assign{
					Var: "res1",
					Flow: data.Activity{
						Id: data.AID("run-echo"),
						Activity: data.Invocation{
							T1: data.Fid("echo"),
							T2: data.GetValue{
								T1: data.Path{"input", "some", "payload"},
							},
						},
					},
				},
			},
			To: data.Activity{
				Id: data.AID("is-ok?"),
				Activity: data.Choose{
					If: data.Eq{
						Left:  data.GetValue{T1: data.Path{"res1", "echoed"}},
						Right: data.SetValue{T1: data.VFloat{T1: 12}},
					},
					Then: data.Transition{
						From: data.Activity{
							Id: "map",
							Activity: data.SetValue{T1: data.VMap{
								data.VMapRecord{
									Key: data.SetValue{T1: data.VString{T1: "my"}},
									Value: data.SetValue{T1: data.VMap{
										data.VMapRecord{
											Key:   data.SetValue{T1: data.VString{T1: "nested"}},
											Value: data.GetValue{T1: data.Path{"res1", "echoed"}},
										},
									}},
								},
							}},
						},
						To: data.Activity{
							Id: data.AID("end"),
							Activity: data.Ok{
								T1: data.SetValue{T1: MustGoValToValues(MapAnyAny{"ok": true})},
							},
						},
					},
					Else: data.Activity{
						Id: data.AID("error"),
						Activity: data.Err{
							T1: data.SetValue{T1: MustGoValToValues(MapAnyAny{"ok": false})},
						},
					},
				},
			},
		},
	}

	expected := &ExecutionState{
		Data: MapAnyAny{
			"ok": true,
		},
		Scope: MapAnyAny{
			"input": MapAnyAny{"some": MapAnyAny{"payload": 12}},
			"res1":  MapStrAny{"echoed": float64(12)},
		},
		// TODO remove from state
		Invoker: jsonInvoke,
	}

	inputVar, err := FindInputVar(flow)
	assert.NoError(t, err)

	state := &ExecutionState{
		Scope: MapAnyAny{
			inputVar: MapAnyAny{"some": MapAnyAny{"payload": 12}},
		},
		// TODO remove from state
		Invoker: jsonInvoke,
	}

	output := ExecuteWorkflow(flow, state)

	assert.Equal(t, expected, output)
}

func TestReshape(t *testing.T) {
	useCases := map[string]struct {
		shape         data.Reshaper
		data          MapAnyAny
		expectedValue interface{}
		expectedErr   error
	}{
		"select element form map": {
			shape:         data.GetValue{T1: data.Path{"a", "b"}},
			data:          MapAnyAny{"a": MapAnyAny{"b": 3}},
			expectedValue: 3,
		},
		"select element form map that don't exists": {
			shape:       data.GetValue{T1: []string{"a", "b", "x", "y", "z"}},
			data:        MapAnyAny{"a": MapAnyAny{"b": 3}},
			expectedErr: errors.New("cannot match path: [a b x y z]"),
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				result, err := DoReshape(uc.shape, uc.data)
				if uc.expectedErr != nil {
					assert.Equal(t, uc.expectedErr, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, uc.expectedValue, result)
				}
			})
		})
	}
}

func TestValuesConversion(t *testing.T) {
	useCases := map[string]struct {
		native interface{}
		values data.Values
		scope  MapAnyAny
	}{
		"int64": {
			native: 123,
			values: data.VInt{T1: 123},
		},
		"string": {
			native: "ola boga",
			values: data.VString{T1: "ola boga"},
		},
		"bool": {
			native: true,
			values: data.VBool{T1: true},
		},
		"map": {
			native: MapAnyAny{"a": 1, "b": MapAnyAny{"c": true}},
			values: data.VMap([]data.VMapRecord{
				{
					Key:   data.SetValue{T1: data.VString{T1: "a"}},
					Value: data.SetValue{T1: data.VInt{T1: 1}},
				},
				{
					Key: data.SetValue{T1: data.VString{T1: "b"}},
					Value: data.SetValue{T1: data.VMap{
						data.VMapRecord{
							Key:   data.SetValue{T1: data.VString{T1: "c"}},
							Value: data.SetValue{T1: data.VBool{T1: true}},
						},
					}},
				},
			}),
		},
		"list": {
			native: []interface{}{"b", true, MapAnyAny{"a": 1}},
			values: data.VList{
				data.SetValue{T1: data.VString{T1: "b"}},
				data.SetValue{T1: data.VBool{T1: true}},
				data.SetValue{T1: data.VMap{
					data.VMapRecord{
						Key:   data.SetValue{T1: data.VString{T1: "a"}},
						Value: data.SetValue{T1: data.VInt{T1: 1}},
					},
				}},
			},
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			native, err := ValuesToGo(uc.values, uc.scope)
			assert.NoError(t, err)
			assert.Equal(t, uc.native, native)

			values, err := GoValToValues(uc.native)
			assert.NoError(t, err)
			assert.Equal(t, uc.values, values)
		})
	}
}
