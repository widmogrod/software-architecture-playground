package some

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/stream"
	"testing"
)

func TestWorkparToWorkflow(t *testing.T) {
	fr := invoker.NewInMemoryFunctionRegistry()
	fr.Register(stream.MkFunctionID("ReserveAvailability"), &invoker.FunctionInMemory{
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

	fr.Register(stream.MkFunctionID("ProcessPayment"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			return `"payment-ok"`
		}})

	// TODO create JSON invoker
	invoke := invoker.NewInvoker(fr)
	jsonInvoke := &jsonInvoker{
		Call: invoke.Invoke,
	}

	flow := WorkparToWorkflow([]byte(`flow start(input) {
	a = ReserveAvailability(input.Id)
	if eq(input.do, 7) {
		b = ProcessPayment(input.Id)
		return({"b": b, "ok": true})
	} else {
		fail({"ok": false})
	}
}`))

	fmt.Printf("%+v\n", flow)

	inputVar, err := FindInputVar(flow)
	assert.NoError(t, err)

	state := &ExecutionState{
		Scope: MapAnyAny{
			inputVar: MapAnyAny{
				"Id": 34,
				"do": false,
			},
		},
		Invoker: jsonInvoke,
	}

	output := ExecuteWorkflow(flow, state)
	fmt.Printf("%+v\n", output)

	assert.Equal(t, MapAnyAny{"ok": false}, output.Data)
}
