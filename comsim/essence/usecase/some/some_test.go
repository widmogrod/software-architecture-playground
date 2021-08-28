package some

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/usecase/data"
	"testing"
)

func TestSome(t *testing.T) {
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

	expected := ExecutionState{
		Data: map[string]interface{}{"a": 1, "echoed": 12},
		End:  data.Ok{},
	}

	state := ExecutionState{
		Data: map[string]interface{}{"a": 1},
	}

	output := ExecuteWorkflow(flow, state)

	assert.Equal(t, expected, output)
}
