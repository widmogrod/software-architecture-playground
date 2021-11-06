package stream

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/comsim/essence/algebra/invoker"
	"testing"
	"time"
)

type (
	OrderCreateCMD struct {
		UserID    int `json:"user_id"`
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	OrderCreateSuccessfully struct {
		OrderID int `json:"order_id"`
	}

	OrderCreateResult struct {
		Success *OrderCreateSuccessfully `json:"success,omitempty"`
	}
)

type (
	SageResult struct {
		AggregateID   int    `json:"aggregate_id"`
		AggregateType string `json:"aggregate_type"`
		StepName      string `json:"step_name"`
		Success       string `json:"success"`
	}
)

func TestSubscriber(t *testing.T) {
	fr := invoker.NewInMemoryFunctionRegistry()
	fr.Register(MkFunctionID("order:create"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			cmd := &OrderCreateCMD{}
			err := json.Unmarshal([]byte(input), cmd)
			if err != nil {
				panic(err)
			}

			return string(toBytes(&OrderCreateResult{
				Success: &OrderCreateSuccessfully{
					OrderID: 66,
				},
			}))
		}})
	fr.Register(MkFunctionID("saga:reserve-availability"), &invoker.FunctionInMemory{
		F: func(input invoker.FunctionInput) invoker.FunctionOutput {
			cmd := OrderCreateResult{}
			err := json.Unmarshal([]byte(input), &cmd)
			if err != nil {
				panic(err)
			}

			return string(toBytes(&SageResult{
				AggregateID:   cmd.Success.OrderID,
				AggregateType: "order",
				StepName:      "reserve-availability",
				Success:       "???",
			}))
		}})
	//fr.Register(MkFunctionID("sage:error-handler"), &invoker.FunctionInMemory{
	//	F: func(input invoker.FunctionInput) invoker.FunctionOutput {
	//		return "ok"
	//	}})

	s := NewChannelStream()
	go s.Work()
	i := NewStreamInvoker(fr, s)
	go i.Work()
	//s.SelectOnce(SelectOnceCMD{
	//	Kind: "InvocationResult",
	//	KeyValue: map[string]SelectConditions{
	//		"fid":    {Eq: "order:created"},
	//		"output": {KeyExists: "success"},
	//	},
	//	MaxFetchSize: 1,
	//})

	w := NewWorkflow()
	w.When(MkFunctionSuccessful("order:create"), MkFunctionID("saga:reserve-availability"))
	//w.When(MkMessageType("saga", "*", "error"), MkFunctionID("sage:error-handler"))

	go ExecuteOnce(s, i, w)

	err, _ := i.Invoke("order:create", string(toBytes(OrderCreateCMD{
		UserID:    666,
		ProductID: 8,
		Quantity:  100,
	})))
	assert.NoError(t, err)

	time.Sleep(time.Second * 5)

	AssertLogContains(t, s.Log(), []*Message{
		{
			Kind: "Invocation",
			Data: toBytes(Invocation{
				FID:   "order:create",
				Input: `{"user_id":666,"product_id":8,"quantity":100}`,
			}),
		}, {
			Kind: "InvocationResult",
			Data: toBytes(InvocationResult{
				FID:    "order:create",
				Input:  `{"user_id":666,"product_id":8,"quantity":100}`,
				Output: `{"success":{"order_id":66}}`,
			}),
		},
		{
			Kind: "Invocation",
			Data: toBytes(Invocation{
				FID:   "saga:reserve-availability",
				Input: `{"success":{"order_id":66}}`,
			}),
		},
		{
			Kind: "InvocationResult",
			Data: toBytes(InvocationResult{
				FID:    "saga:reserve-availability",
				Input:  `{"success":{"order_id":66}}`,
				Output: `{"aggregate_id":66,"aggregate_type":"order","step_name":"reserve-availability","success":"???"}`,
			}),
		},
	})
}

func ExecuteOnce(s2 *ChannelStream, i *StreamInvoke, w *Workflow) {
	//wg := sync.WaitGroup{}
	for s, f := range w.Flow {
		//wg.Add(1)
		//go func(s *SelectOnceCMD, f invoker.FunctionID) {
		//	defer wg.Done()
		fmt.Printf("ExecuteOnce: SelectOnce %#v \n", *s)
		for _, m := range s2.SelectOnce(*s) {
			fmt.Println("ExecuteOnce: Invoke ...")
			ir := InvocationResult{}
			err := json.Unmarshal(m.Data, &ir)
			if err != nil {
				panic(err)
			}

			err, _ = i.Invoke(f, ir.Output)
			if err != nil {
				panic(err)
			}
			continue
		}
		//}(s, f)
	}

	//wg.Wait()
}

func MkFunctionSuccessful(fid invoker.FunctionID) SelectOnceCMD {
	return SelectOnceCMD{
		Kind: "InvocationResult",
		Selector: &SelectConditions{KeyValue: map[string]SelectConditions{
			"fid":    {Eq: fid},
			"output": {KeyExists: "success"},
		}},
		MaxFetchSize: 1,
	}
}
