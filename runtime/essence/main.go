package main

import (
	"encoding/json"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"io"
	"net/http"
	"time"
)

type OrderAggregateState struct {
	OrderID string
	UserID  string

	OrderTotalPrice string

	Created time.Time
	Updated time.Time

	ProductID        string
	ProductUnitPrice string
	ProductQuantity  string

	WarehouseStatus        string
	WarehouseReservationID string

	ShippingStatus string
	ShippingID     string
}

type OrderCreateCMD struct {
	UserID    string
	ProductID string
	Quantity  string
}

type OrderCreateResult struct {
	OrderID   string
	UserID    string
	ProductID string
	Quantity  string
}

type OrderUpdatePriceCMD struct {
	OrderID string
}
type OrderUpdatePriceResult struct {
	ProductID        string
	ProductUnitPrice string
}

type OrderCollectPaymentsCMD struct {
	OrderID string
}
type OrderCollectPaymentsResult struct {
	PaymentCollected bool
}

type OrderPrepareWarehouseReservationCMD struct {
}
type OrderPrepareWarehouseReservationResult struct {
	ReservationID      string
	ReservationTimeout time.Time
}

type OrderCommitWarehouseReservationCMD struct {
}
type OrderCommitWarehouseReservationResult struct {
	ReservationCommitted    bool
	ReservationCancelReason string
}

type OrderCompleteCMD struct {
}
type OrderCompleteResult struct {
}

func main() {
	// what if data is polymorfic
	// what if everything is an networks call
	// what if response contains data to be persisted, events to be publish, ...
	// what if storage is not a user defined concern?
	// what if external communication also is HTTP base, and persisting results must always happen in aggregates
	//  and external information can be only delivered as a webhook

	//mux := http.NewServeMux()
	//mux.HandleFunc("/entrypoint/schedule", func(rq http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(rq, "Hohohoho")
	//})
	app := runtime.NewMuxRuntimeClient()
	app.
		HandleFunc("/schedule/time", ScheduleInvokeHandlerFunc(func(cmd *runtime.ScheduleInvokeCMD, result *runtime.ScheduleInvokeResult) {
			result.Logs = time.Now().String()
		})).
		Schedule("* * * * *")

	//app.HandleFunc("/order/create2", AggregateChangeHandlerFunc(func(cmd *runtime.AggregateChangeCMD, result *runtime.AggregateChangeResult) {
	//	input := &OrderCreateCMD{}
	//	JsonAggregateChange(cmd, result, input, func() {
	//		// do some business logic, like validate inputs
	//		//if errors := input.Validate(); errors != nil {
	//		//	r.Result(errors)
	//		//	return
	//		//}
	//
	//		change := &OrderCreateResult{
	//			OrderID:   ksuid.New().String(),
	//			UserID:    cmd.UserID,
	//			ProductID: cmd.ProductID,
	//			Quantity:  cmd.Quantity,
	//		}
	//
	//		r.AggregateID(change.OrderID)
	//		r.Append(change)
	//	})
	//}))

	app.
		HandleFunc("/order/create", func(w http.ResponseWriter, rq *http.Request) {
			input := &runtime.AggregateChangeCMD{}
			output := &runtime.AggregateChangeResult{}
			JsonRequestResponse(w, rq, input, output, func() {
				cmd := &OrderCreateCMD{}
				err := json.Unmarshal(input.Payload, cmd)
				if err != nil {
					output.Err = err
					output.Logs += "fail to unmarshall command, terminate command, err=" + err.Error()
					return
				}

				// Do command specific logic, and when everything is ok then create change
				change := &OrderCreateResult{
					OrderID:   ksuid.New().String(),
					UserID:    cmd.UserID,
					ProductID: cmd.ProductID,
					Quantity:  cmd.Quantity,
				}

				data, err := json.Marshal(change)
				if err != nil {
					output.Err = err
					output.Logs += "fail to marshal change, terminate command, err=" + err.Error()
					return
				}

				output.AggregateID = change.OrderID
				output.Changes = append(output.Changes, &runtime.AggregateChange{
					Type:       "OrderCreateResult",
					Payload:    data,
					RecordTime: time.Now(),
				})
			})
		}).
		AggregateChange("order", "create")

	app.
		HandleFunc("/order/reduce", func(w http.ResponseWriter, rq *http.Request) {
			input := &runtime.AggregateReduceCMD{}
			output := &runtime.AggregateReduceResult{}
			JsonRequestResponse(w, rq, input, output, func() {
				aggregate := &OrderAggregateState{}
				err := json.Unmarshal(input.Snapshot, aggregate)
				if err != nil && input.Snapshot != nil {
					output.Err = err
					output.Logs += "snapshot(" + string(input.Snapshot) + ")"
					output.Logs += "failed on snapshot restore, end reduction, err=" + err.Error()
					return
				}

				for _, change := range input.Changes {
					output.Logs += fmt.Sprintf("change(%s)\n", change.Type)
					switch change.Type {
					case "OrderCreateResult":
						c := &OrderCreateResult{}
						err := json.Unmarshal(change.Payload, c)
						if err != nil {
							output.Err = err
							output.Logs += "failed on event, end changes application, err=" + err.Error()
							return
						}

						aggregate.ProductID = c.ProductID
						aggregate.ProductQuantity = c.Quantity
						aggregate.UserID = c.UserID
						aggregate.OrderID = c.OrderID
						// etc on other types of changes
					}
				}

				// This is a snapshot that will be serialise and available to other commands in aggregate
				data, err := json.Marshal(aggregate)
				if err != nil {
					output.Err = err
					output.Logs += err.Error()
				} else {
					output.Snapshot = data
				}
			})
		}).
		AggregateReducer("order")

	//type OrderAggregate struct {
	//	AggregateChange
	//}
	//
	//func (a OrderAggregate) Reducer()
	//
	//app.AggregateChange("order").
	//	Command("create").
	//	Func(func(w http.ResponseWriter, rq *http.Request){
	//		return &AggregateChange{
	//			Type: "order",
	//			ID: "asdasd",
	//			State: &Order{},
	//			Changes: []interface{}{
	//				&OrderCreated{
	//					ID: "",
	//					UserID:
	//				},
	//			},
	//		}
	//	})
	//
	//app.HandleFunc("/order/create", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/internal/collect-price-information")
	//})
	//app.HandleFunc("/order/status", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]},
	//})
	////app.HandleFunc("/order/internal/collect-price-information", func(w http.ResponseWriter, rq *http.Request) {
	////	return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]},
	////}).WhenCompleted("/order/create")
	//
	//app.HandleFunc("/order/internal/reserve-availability", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/internal/collect-payment")
	//})
	//app.HandleFunc("/order/internal/collect-payment", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/internal/commit-to-availability")
	//})
	//app.HandleFunc("/order/internal/commit-to-availability", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/ship")
	//})
	//app.HandleFunc("/order/internal/ship", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/internal/check-for-delivery-status")
	//})
	//app.HandleFunc("/order/internal/check-for-delivery-status", func(w http.ResponseWriter, rq *http.Request) {
	//	// wait-for(delivery).delivered
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/complete")
	//})
	//app.HandleFunc("/order/complete", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}
	//})
	//
	//app.HandleFunc("/shipping/webhook/delivered", func(w http.ResponseWriter, rq *http.Request) {
	//	//return result{aggregateType: order, aggregateID: 123, state: "", changes: [""]}, invoke("order/ship")
	//})

	http.ListenAndServe(":8080", app)
}

type Result struct {
	Error error `json:"error"`
}

func JsonRequestResponse(w http.ResponseWriter, rq *http.Request, input, output interface{}, do func()) {
	err := json.NewDecoder(rq.Body).Decode(input)
	if err != nil && err != io.EOF {
		result := &Result{
			Error: err,
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result)
	} else {
		do()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(output)
	}
}

func ScheduleInvokeHandlerFunc(handle func(*runtime.ScheduleInvokeCMD, *runtime.ScheduleInvokeResult)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		input := &runtime.ScheduleInvokeCMD{}
		output := &runtime.ScheduleInvokeResult{}
		JsonRequestResponse(w, rq, input, output, func() {
			handle(input, output)
		})
	}
}

//func AggregateChangeHandlerFunc(handle func()) http.HandlerFunc {
//	return func(w http.ResponseWriter, rq *http.Request) {
//		input := &runtime.AggregateReduceCMD{}
//		output := &runtime.AggregateReduceResult{}
//		JsonRequestResponse(w, rq, input, output, func() {
//			handle(input, output)
//		})
//	}
//}
