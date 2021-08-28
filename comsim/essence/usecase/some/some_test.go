package some

import (
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
						Path:  data.Path([]string{"a"}),
						Value: 1,
					},
					Then: data.Activity{
						Id:       data.AID("end"),
						Activity: data.End{Reason: data.Ok{}},
					},
					Else: data.Activity{
						Id:       data.AID("error"),
						Activity: data.End{Reason: data.Err{}},
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
