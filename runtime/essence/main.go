package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/widmogrod/software-architecture-playground/runtime"
	"io"
	"net/http"
	"reflect"
	"time"
)

var (
	mappings = make(map[interface{}]map[string]interface{})
)

func init() {
	RegisterAggregateChanges(&OrderAggregateState{}, map[string]interface{}{
		"OrderCreated": &OrderCreated{},
		"ProductAdded": &ProductAdded{},
	})
}

type eventName = string
type aggregateState = interface{}
type aggregateEvent = interface{}
type doFunc = func() error

type whenApplyFunc = func(eventName, aggregateEvent, doFunc)
type hasEventFunc = func(aggregateState, ...func(whenApplyFunc))
type initAggregateFunc = func(hasEventFunc)

func x() {
	AggregateState(func(hasEvents hasEventFunc) {
		state := &OrderAggregateState{}
		hasEvents(
			state,
			func(whenApply whenApplyFunc) {
				event := &OrderCreated{}
				whenApply("OrderCreated", event, func() error {
					if state.isOrderCreated {
						return errors.New("order cannot be created twice, check your logic")
					}

					state.OrderID = event.OrderID
					state.OrderCreatedAt = event.CreatedAt
					state.isOrderCreated = true
					return nil
				})
			},
			func(whenApply whenApplyFunc) {
				event := &ProductAdded{}
				whenApply("ProductAdded", event, func() error {
					if !state.isOrderCreated {
						return errors.New("You cannot add products to not created order")
					}

					state.ProductID = event.ProductID
					state.ProductQuantity = event.Quantity
					return nil
				})
			})
	})
}

func AggregateState(init initAggregateFunc) {
	hasEvents := func(state aggregateState, addEvents ...func(whenApplyFunc)) {
		for _, addEvent := range addEvents {
			whenApply := func(name eventName, event aggregateEvent, do doFunc) {

			}

			addEvent(whenApply)
		}
	}

	init(hasEvents)
}

func Apply2(state agg, change interface{}) error {
	t0 := reflect.TypeOf(change).String()
	tName := ""
	fmt.Println(reflect.TypeOf(state).String(), mappings)
	for typ, c := range mappings[reflect.TypeOf(state).String()] {
		t1 := reflect.TypeOf(c).String()
		fmt.Println(t1, change)
		if t0 == t1 {
			tName = typ
			break
		}
	}

	if tName == "" {
		return errors.New(fmt.Sprintf("type %T cannot be found in mapping return by mapOfChanges()", change))
	}

	// todo delegate to diferent layer
	payload, err := json.Marshal(change)
	if err != nil {
		return errors.New(fmt.Sprintf("coudn't marchal change %T", change))
	}

	now := time.Now()
	c := &runtime.AggregateChange{
		Type:       tName,
		Payload:    payload,
		RecordTime: &now,
	}

	err = state.Handle(change)
	if err != nil {
		state.AggregateAppendChange(c)
	}

	return err
}

type AggregateApply struct {
	modifications []*runtime.AggregateChange
}

func (a *AggregateApply) AggregateModifications() []*runtime.AggregateChange {
	return a.modifications
}
func (a *AggregateApply) AggregateAppendChange(change *runtime.AggregateChange) {
	if a.modifications == nil {
		a.modifications = make([]*runtime.AggregateChange, 0)
	}
	a.modifications = append(a.modifications, change)
}

//func (a *AggregateApply) apply(change interface{}) error {
//	if a.modifications == nil {
//		a.modifications = make([]*runtime.AggregateChange, 0)
//	}
//
//	//convert domain object to aggregate change
//
//	t0 := reflect.TypeOf(change).String()
//	tName := ""
//	for typ, c := range a.mapOfChanges() {
//		t1 := reflect.TypeOf(c).String()
//		if t0 == t1 {
//			tName = typ
//			break
//		}
//	}
//
//	if tName == "" {
//		return errors.New(fmt.Sprintf("type %t cannot be found in mapping return by mapOfChanges()", change))
//	}
//
//	// todo delegate to diferent layer
//	payload, err := json.Marshal(change)
//	if err != nil {
//		return errors.New(fmt.Sprintf("coudn't marchal change %t", change))
//	}
//
//	now := time.Now()
//	c := &runtime.AggregateChange{
//		Type:       tName,
//		Payload:    payload,
//		RecordTime: &now,
//	}
//
//	a.modifications = append(a.modifications, c)
//
//	return nil
//}

type agg interface {
	Handle(interface{}) error
	AggregateAppendChange(*runtime.AggregateChange)
	AggregateModifications() []*runtime.AggregateChange
}

func Apply(state agg, change interface{}) error {
	t0 := reflect.TypeOf(change).String()
	tName := ""
	fmt.Println(reflect.TypeOf(state).String(), mappings)
	for typ, c := range mappings[reflect.TypeOf(state).String()] {
		t1 := reflect.TypeOf(c).String()
		fmt.Println(t1, change)
		if t0 == t1 {
			tName = typ
			break
		}
	}

	if tName == "" {
		return errors.New(fmt.Sprintf("type %T cannot be found in mapping return by mapOfChanges()", change))
	}

	// todo delegate to diferent layer
	payload, err := json.Marshal(change)
	if err != nil {
		return errors.New(fmt.Sprintf("coudn't marchal change %T", change))
	}

	now := time.Now()
	c := &runtime.AggregateChange{
		Type:       tName,
		Payload:    payload,
		RecordTime: &now,
	}

	err = state.Handle(change)
	if err != nil {
		state.AggregateAppendChange(c)
	}

	return err
}

func RegisterAggregateChanges(state interface{}, ma map[string]interface{}) {
	key := reflect.TypeOf(state).String()
	mappings[key] = ma
}

type OrderAggregateState struct {
	AggregateApply

	OrderID        string
	OrderCreatedAt *time.Time

	UserID string

	OrderTotalPrice string
	Created         *time.Time

	Updated          *time.Time
	ProductID        string
	ProductUnitPrice string

	ProductQuantity string
	WarehouseStatus string

	WarehouseReservationID string
	ShippingStatus         string

	ShippingID     string
	isOrderCreated bool
}

func (s *OrderAggregateState) Handle(change interface{}) error {
	switch c := change.(type) {
	case *OrderCreated:
		if s.isOrderCreated {
			return errors.New("order cannot be created twice, check your logic")
		}

		s.OrderID = c.OrderID
		s.OrderCreatedAt = c.CreatedAt

		s.isOrderCreated = true

	case *ProductAdded:
		s.ProductID = c.ProductID
		s.ProductQuantity = c.Quantity

	default:
		return errors.New(fmt.Sprintf("unsupported type to handle %T", change))
	}

	return nil
}
func (s *OrderAggregateState) CreateNew(userID string) error {
	now := time.Now()
	return Apply(s, &OrderCreated{
		OrderID:   ksuid.New().String(),
		UserId:    userID,
		CreatedAt: &now,
	})
}

type ProductAdded struct {
	ProductID string
	Quantity  string
}

func (s *OrderAggregateState) AddProduct(productID, quantity string) error {
	if !s.isOrderCreated {
		return errors.New("You cannot add products to not created order")
	}

	return Apply(s, &ProductAdded{
		ProductID: productID,
		Quantity:  quantity,
	})
}

type OrderCreateCMD struct {
	UserID    string
	ProductID string
	Quantity  string
}

type OrderCreated struct {
	OrderID   string
	UserId    string
	CreatedAt *time.Time
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

	app.
		HandleFunc("/order/create", AggregateChangeHandlerFunc(func(apply ApplyChangeFunc) {
			cmd := &OrderCreateCMD{}
			state := &OrderAggregateState{}
			apply(cmd, state, func() error {
				return state.CreateNew(cmd.UserID)
			}, func() error {
				return state.AddProduct(cmd.ProductID, cmd.Quantity)
			})
		})).
		AggregateChange("order", "create")

	//app.
	//	HandleFunc("/order/create", func(w http.ResponseWriter, rq *http.Request) {
	//		// what if response decides what this handler is about?
	//		// runtime.AggregateInvokeCommand()
	//		// - runtimes receives it and apply command on aggregate
	//		// - application of command is an API call that takes as input
	//		//		runtime.AggregateApplyCommand(cmd, agg_type, agg_id, snapshot)
	//		// - application of command is always on snapshot, snapshot is generated by invoking another API call
	//		//
	//		//         CLIENT                      RUNTIME					STORAGE		    APPLICATION
	//		//            |
	//		//            +---- POST /something  ----------------------------------------------> +
	//		//										  								  			 |]
	//		//                           			  | <- AggregateInvokeCommand() ------------ +
	//		//										  |	-- GetAggregate() --> |
	//		//										  |	-- ProdudeState -----------------------> +
	//		//										  								  			 |]
	//		//                           			  | <- state:=AggregateState()  ------------ +
	//		//										  |	-- AggregateApply(cmd,state) ----------> +
	//		//										  								  			 |]
	//		//                           			  | <- changes:=AggregateApplyChanges()----- +
	//		//
	//		//										  |	-- ApplyChanges(ch) > |
	//		//
	//		// runtime.ScheduleInvokeCommand()
	//
	//		input := &runtime.AggregateChangeCMD{}
	//		output := &runtime.AggregateChangeResult{}
	//		JsonRequestResponse(w, rq, input, output, func() {
	//			cmd := &OrderCreateCMD{}
	//			err := json.Unmarshal(input.Payload, cmd)
	//			if err != nil {
	//				output.Err = err.Error()
	//				output.Logs += "fail to unmarshall command, terminate command, err=" + err.Error()
	//				return
	//			}
	//
	//			// validate if product exists
	//			state := &OrderAggregateState{}
	//			handle2 := func(state interface{}, input *runtime.AggregateChangeCMD, do func()) {
	//				err = json.Unmarshal(input.Snapshot, state)
	//				if err != nil && err != io.EOF {
	//					output.Err = err.Error()
	//					output.Logs += "fail to unmarshall state aggregate, terminate command, err=" + err.Error()
	//					return
	//				}
	//			}
	//
	//			handle2(state, input, func() {
	//				state.CreateNew()
	//				state.AddProduct(cmd.ProductID, cmd.Quantity)
	//			})
	//
	//			retunr2 := func() {
	//				output.Changes = state.AggregateModifications()
	//
	//				data, err := json.Marshal(state)
	//				if err != nil {
	//					output.Err = err.Error()
	//					output.Logs += "fail to marshal change, terminate command, err=" + err.Error()
	//					return
	//				}
	//
	//				output.Snapshot = data
	//
	//			}
	//
	//			retunr2()
	//		})
	//	}).
	//	AggregateChange("order", "create")

	app.
		HandleFunc("/order/reduce", func(w http.ResponseWriter, rq *http.Request) {
			input := &runtime.AggregateReduceCMD{}
			output := &runtime.AggregateReduceResult{}
			JsonRequestResponse(w, rq, input, output, func() {
				aggregate := &OrderAggregateState{}
				err := json.Unmarshal(input.Snapshot, aggregate)
				if err != nil && input.Snapshot != nil {
					output.Err = err.Error()
					output.Logs += "snapshot(" + string(input.Snapshot) + ")"
					output.Logs += "failed on snapshot restore, end reduction, err=" + err.Error()
					return
				}

				for _, change := range input.Changes {
					output.Logs += fmt.Sprintf("change(%s)\n", change.Type)
					switch change.Type {
					case "OrderCreated":
						c := &OrderCreated{}
						err := json.Unmarshal(change.Payload, c)
						if err != nil {
							output.Err = err.Error()
							output.Logs += "failed on event, end changes application, err=" + err.Error()
							return
						}

						aggregate.UserID = c.UserId
						aggregate.OrderID = c.OrderID
						aggregate.OrderCreatedAt = c.CreatedAt
						// etc on other types of changes
					}
				}

				// This is a snapshot that will be serialise and available to other commands in aggregate
				data, err := json.Marshal(aggregate)
				if err != nil {
					output.Err = err.Error()
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
	//			Snapshot: &Order{},
	//			Changes: []interface{}{
	//				&OrderCreatedAt{
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

type result struct {
	Error error `json:"error"`
}

func JsonRequestResponse(w http.ResponseWriter, rq *http.Request, input, output interface{}, do func()) {
	err := json.NewDecoder(rq.Body).Decode(input)
	if err != nil && err != io.EOF {
		result := &result{
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

type ApplyChangeFunc = func(interface{}, AggregateHelper, func() error, ...func() error)

type AggregateHelper interface {
	AggregateModifications() []*runtime.AggregateChange
}

func AggregateChangeHandlerFunc(handle func(apply ApplyChangeFunc)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		input := &runtime.AggregateChangeCMD{}
		output := &runtime.AggregateChangeResult{}
		JsonRequestResponse(w, rq, input, output, func() {
			apply := func(cmd interface{}, state AggregateHelper, do func() error, dos ...func() error) {
				err := json.Unmarshal(input.Snapshot, state)
				if err != nil && err != io.EOF {
					output.Err = err.Error()
					output.Logs += "fail to unmarshall state aggregate, terminate command, err=" + err.Error()
					return
				}

				dos = append([]func() error{do}, dos...)
				for _, do := range dos {
					err = do()
					if err != nil {
						output.Err = err.Error()
						output.Logs += fmt.Sprintf("fail during application of command %T to aggregate %T", cmd, state)
						return
					}
				}

				output.Changes = state.AggregateModifications()

				data, err := json.Marshal(state)
				if err != nil {
					output.Err = err.Error()
					output.Logs += "fail to marshal change, terminate command, err=" + err.Error()
					return
				}

				output.Snapshot = data
			}

			handle(apply)
		})
	}
}
