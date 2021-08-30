package some

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"testing"
)

func TestSome(t *testing.T) {
	flow := data.Transition{
		From: data.Activity{Id: data.AID("start"), Activity: data.Start{}},
		To: data.Transition{
			From: data.Activity{
				Id: data.AID("run-echo"),
				Activity: data.Invocation{
					T1: data.Fid("echo"),
				},
			},
			To: data.Activity{
				Id: data.AID("is-ok?"),
				Activity: data.Choose{
					If: data.Eq{
						Path:  data.Path([]string{"echoed"}),
						Value: 12,
					},
					Then: data.Transition{
						//From: data.Activity{
						//	Id:       "map",
						//	Activity: data.Select{Path: []string{"echoed"}},
						//},
						From: data.Activity{
							Id: "map",
							Activity: data.ReMap([]data.ReMapRecord{
								{
									Key:   data.Path([]string{"my", "nested"}),
									Value: data.Path([]string{"echoed"}),
								},
							}),
						},
						To: data.Activity{
							Id:       data.AID("end"),
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
				"nested": 12,
			},
		},
		End: data.Ok{},
	}

	state := &ExecutionState{
		Data: map[string]interface{}{"a": 1},
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
