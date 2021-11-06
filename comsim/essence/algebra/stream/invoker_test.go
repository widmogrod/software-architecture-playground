package stream

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"testing"
	"time"
)

func TestStreamOfInvocation(t *testing.T) {
	fr := invoker.NewInMemoryFunctionRegistry()
	fr.Register(MkFunctionID("test-func"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return fmt.Sprintf("output of test func(%s)", input)
		}})

	fr.Register(MkFunctionID("test-func2"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return fmt.Sprintf("output of test func2(%s)", input)
		}})

	s := NewChannelStream()
	go s.Work()
	i := NewStreamInvoker(fr, s)
	go i.Work()

	err, r := i.Invoke(MkFunctionID("test-func"), "1312")
	assert.NoError(t, err)
	assert.Equal(t, "output of test func(1312)", r)

	err, r = i.Invoke(MkFunctionID("test-func2"), "32")
	assert.NoError(t, err)
	assert.Equal(t, "output of test func2(32)", r)

	time.Sleep(time.Second)

	AssertLogContains(t, s.Log(), []*Message{
		{
			Kind: "Invocation",
			Data: toBytes(Invocation{
				FID:   "test-func",
				Input: "1312",
			}),
		}, {
			Kind: "InvocationResult",
			Data: toBytes(InvocationResult{
				FID:    "test-func",
				Input:  "1312",
				Output: "output of test func(1312)",
			}),
		},
		{
			Kind: "Invocation",
			Data: toBytes(Invocation{
				FID:   "test-func2",
				Input: "32",
			}),
		}, {
			Kind: "InvocationResult",
			Data: toBytes(InvocationResult{
				FID:    "test-func2",
				Input:  "32",
				Output: "output of test func2(32)",
			}),
		},
	})
}

func AssertLogContains(t *testing.T, log, contains []*Message) {
	if !assert.Equal(t, len(contains), len(log), "Log lengths don't match") {
		t.Logf("log(%d)=%v\n", len(log), log)
		t.Logf("con(%d)=%v\n", len(contains), contains)
		return
	}

	for i, m := range log {
		assert.Equal(t, m.Kind, contains[i].Kind)

		var a, b map[string]interface{} = nil, nil

		//TODO fix assumption that message are JSON
		if m.Data != nil {
			err := json.Unmarshal(m.Data, &a)
			assert.NoError(t, err)
		}
		if contains[i].Data != nil {
			err := json.Unmarshal(contains[i].Data, &b)
			assert.NoError(t, err)
		}
		AssertMapSubset(t, a, b)
	}
}

func AssertMapSubset(t *testing.T, amap, subset map[string]interface{}) {
	for k, v := range subset {
		assert.Equal(t, v, amap[k])
	}
}
