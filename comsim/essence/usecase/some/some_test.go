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
			var i int
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
							T2: data.Select{
								Path: data.Path{"input", "some", "payload"},
							},
						},
					},
				},
			},
			To: data.Activity{
				Id: data.AID("is-ok?"),
				Activity: data.Choose{
					If: data.Eq{
						Path:  data.Path{"res1", "echoed"},
						Value: float64(12),
					},
					Then: data.Transition{
						From: data.Activity{
							Id: "map",
							Activity: data.ReMap([]data.ReMapRecord{
								{
									Key:   data.Path{"my", "nested"},
									Value: data.Path{"res1", "echoed"},
								},
							}),
						},
						To: data.Activity{
							Id: data.AID("end"),
							// TODO replace with data.Return(Select{Path: []string{}})
							// to signal what data to return explicitly, not by implicitly by flow
							// ???
							Activity: data.Ok{},
						},
					},
					Else: data.Activity{
						Id:       data.AID("error"),
						Activity: data.Err{},
					},
				},
			},
		},
	}

	expected := &ExecutionState{
		//Data: 12,
		Data: map[string]interface{}{
			"my": map[string]interface{}{
				"nested": float64(12),
			},
		},
		Scope: MapStrAny{
			"input": MapStrAny{"some": MapStrAny{"payload": 12}},
			"res1":  MapStrAny{"echoed": float64(12)},
		},
		End: data.Ok{},
		// TODO remove from state
		Invoker: jsonInvoke,
	}

	inputVar, err := FindInputVar(flow)
	assert.NoError(t, err)

	state := &ExecutionState{
		Scope: MapStrAny{
			inputVar: MapStrAny{"some": MapStrAny{"payload": 12}},
		},
		// TODO remove from state
		Invoker: jsonInvoke,
	}

	output := ExecuteWorkflow(flow, state)

	assert.Equal(t, expected, output)
}

func TestReshape(t *testing.T) {
	useCases := map[string]struct {
		shape         data.Reshape
		data          interface{}
		expectedValue interface{}
		expectedErr   error
	}{
		"select element form map": {
			shape:         data.Select{Path: []string{"a", "b"}},
			data:          MapStrAny{"a": MapStrAny{"b": 3}},
			expectedValue: 3,
		},
		"select element form map that don't exists": {
			shape:       data.Select{Path: []string{"a", "b", "x", "y", "z"}},
			data:        MapStrAny{"a": MapStrAny{"b": 3}},
			expectedErr: errors.New("cannot match path: [a b x y z]"),
		},
		"reshape map values": {
			shape: data.ReMap([]data.ReMapRecord{
				{Key: []string{"k1"}, Value: []string{"a.1", "b.1"}},
				{Key: []string{"k2"}, Value: []string{"a.1", "b.2", "c.1"}},
				{Key: []string{"nest", "nest", "a"}, Value: []string{"a.1", "b.2", "c.2"}},
			}),
			data: MapStrAny{
				"a.1": MapStrAny{
					"b.1": 3,
					"b.2": MapStrAny{
						"c.1": false,
						"c.2": nil,
					},
				},
			},
			expectedValue: MapStrAny{
				"k1": 3,
				"k2": false,
				"nest": MapStrAny{
					"nest": MapStrAny{
						"a": nil,
					},
				},
			},
		},
		"reshape map values that don't match": {
			shape: data.ReMap([]data.ReMapRecord{
				{Key: []string{"k1"}, Value: []string{"a.1", "b.1", "x", "y"}},
			}),
			data: MapStrAny{
				"a.1": MapStrAny{
					"b.1": 3,
					"b.2": MapStrAny{
						"c.1": false,
						"c.2": nil,
					},
				},
			},
			expectedErr: errors.New("cannot map: {[k1] [a.1 b.1 x y]}"),
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
