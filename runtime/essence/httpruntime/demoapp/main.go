package main

import (
	"encoding/json"
	"github.com/widmogrod/software-architecture-playground/runtime/essence/httpruntime"
	"io"
	"net/http"
	"time"
)

//func NewOrderAggregate() *OrderAggregate {
//	store := aggregate.NewEventStore()
//	aggregate := &OrderAggregate{
//		state:   nil,
//		changes: store,
//		ref: &runtime.AggregateRef{
//			ID:   "",
//			Type: "order",
//		},
//	}
//
//	return aggregate
//}
//
//type OrderAggregate struct {
//	state   *OrderAggregateState
//	changes *aggregate.EventStore
//	ref     *runtime.AggregateRef
//}
//
//func (o *OrderAggregate) Ref() *runtime.AggregateRef {
//	return o.ref
//}
//
//func (o *OrderAggregate) State() interface{} {
//	return o.state
//}
//
//func (o *OrderAggregate) Changes() *aggregate.EventStore {
//	return o.changes
//}
//func (o *OrderAggregate) Hydrate(state interface{}, ref *runtime.AggregateRef) error {
//	o.state = state.(*OrderAggregateState)
//	o.ref = ref
//
//	return nil
//}
//
//func (o *OrderAggregate) Handle(cmd interface{}) error {
//	switch c := cmd.(type) {
//	case *OrderCreateCMD:
//		// validate necessary condition
//		if o.state != nil {
//			return errors.New("Order already exists!")
//		}
//		if c.Quantity == "" {
//			return errors.New(fmt.Sprintf("Given quantity is to low %v", c))
//		}
//
//		now := time.Now()
//		return o.changes.
//			Append(&OrderCreated{
//				OrderID:   ksuid.New().String(),
//				UserId:    c.UserID,
//				CreatedAt: &now,
//			}).Ok.
//			Append(&ProductAdded{
//				ProductID: c.ProductID,
//				Quantity:  c.Quantity,
//			}).Ok.
//			ReduceRecent(o).Err
//
//	case *OrderCollectPaymentsCMD:
//		// validate necessary condition
//		if o.state == nil {
//			return errors.New("Order dont exists!")
//		}
//		if c.OrderID != o.state.OrderID {
//			return errors.New(fmt.Sprintf("Order missmatch %v", c))
//		}
//
//		return o.changes.
//			Append(&OrderCollectPaymentsResult{
//				PaymentCollected: true,
//			}).Ok.
//			ReduceRecent(o).Err
//	}
//
//	return nil
//}
//
//func (o *OrderAggregate) Apply(change interface{}) error {
//	switch c := change.(type) {
//	case *OrderCreated:
//		if o.state != nil {
//			return errors.New("order cannot be created twice, check your logic")
//		}
//
//		o.ref.ID = c.OrderID
//
//		// when everything is ok, record changes that you want to make
//		o.state = &OrderAggregateState{}
//		o.state.OrderID = c.OrderID
//		o.state.OrderCreatedAt = c.CreatedAt
//
//	case *ProductAdded:
//		if o.state == nil {
//			return errors.New("You cannot add products to not created order")
//		}
//
//		o.state.ProductID = c.ProductID
//		o.state.ProductQuantity = c.Quantity
//
//	case *OrderCollectPaymentsResult:
//		if o.state == nil {
//			return errors.New("You cannot collect payment for order that don't exists")
//		}
//
//		o.state.PaymentCollected = c.PaymentCollected
//
//	default:
//		return errors.New(fmt.Sprintf("unsupported type to handle %T", change))
//	}
//
//	return nil
//}
//
//type OrderAggregateState struct {
//	OrderID        string
//	OrderCreatedAt *time.Time
//
//	UserID string
//
//	OrderTotalPrice string
//
//	Created *time.Time
//	Updated *time.Time
//
//	ProductID        string
//	ProductUnitPrice string
//
//	ProductQuantity string
//	WarehouseStatus string
//
//	WarehouseReservationID string
//	ShippingStatus         string
//
//	ShippingID       string
//	isOrderCreated   bool
//	PaymentCollected bool
//}
//
//type ProductAdded struct {
//	ProductID string
//	Quantity  string
//}
//
//type OrderCreateCMD struct {
//	UserID    string
//	ProductID string
//	Quantity  string
//}
//
//type OrderCreated struct {
//	OrderID   string
//	UserId    string
//	CreatedAt *time.Time
//}
//
//type OrderUpdatePriceCMD struct {
//	OrderID string
//}
//
//type OrderUpdatePriceResult struct {
//	ProductID        string
//	ProductUnitPrice string
//}
//
//type OrderCollectPaymentsCMD struct {
//	OrderID string
//}
//type OrderCollectPaymentsResult struct {
//	PaymentCollected bool
//}
//
//type OrderPrepareWarehouseReservationCMD struct {
//}
//type OrderPrepareWarehouseReservationResult struct {
//	ReservationID      string
//	ReservationTimeout time.Time
//}
//
//type OrderCommitWarehouseReservationCMD struct {
//}
//type OrderCommitWarehouseReservationResult struct {
//	ReservationCommitted    bool
//	ReservationCancelReason string
//}
//
//type OrderCompleteCMD struct {
//}
//type OrderCompleteResult struct {
//}
//
//func NewMarchaller() *ChangesRegistry {
//	return &ChangesRegistry{
//		changes: map[string]interface{}{
//			"OrderCreated": &OrderCreated{},
//			"ProductAdded": &ProductAdded{},
//		},
//	}
//}
//
//func NewChangesRegistry() *ChangesRegistry {
//	return &ChangesRegistry{
//		state:   &OrderAggregateState{},
//		changes: nil,
//	}
//}
//
//type ChangesRegistry struct {
//	state   interface{}
//	changes map[string]interface{}
//}
//
//func (m *ChangesRegistry) Set(name string, change interface{}) {
//	m.changes[name] = change
//}
//
//type RuntimeChangeStorage struct {
//	input  *runtime.AggregateChangeCMD
//	output *runtime.AggregateChangeResult
//	//registry *ChangesRegistry
//}
//
//type Aggregate interface {
//	Changes() *aggregate.EventStore
//	State() interface{}
//	Ref() *runtime.AggregateRef
//	Hydrate(state interface{}, ref *runtime.AggregateRef) error
//}
//
//func (s *RuntimeChangeStorage) Persist(a Aggregate) error {
//	version := s.input.SnapshotVersion
//	_ = a.Changes().Reduce(func(change interface{}, result *aggregate.Reduced) *aggregate.Reduced {
//		data, _ := json.Marshal(change)
//
//		version++
//
//		now := time.Now()
//		s.output.Changes = append(s.output.Changes, &runtime.AggregateChange{
//			// TODO change name can be provided by Event
//			Type:       reflect.TypeOf(change).Elem().Name(),
//			Payload:    data,
//			RecordTime: &now,
//			Version:    version,
//		})
//
//		return result
//	}, nil).Err
//
//	data, err := json.Marshal(a.State())
//	if err != nil {
//		return err
//	}
//
//	s.output.AggregateType = a.Ref().Type
//	s.output.AggregateID = a.Ref().ID
//	s.output.Snapshot = data
//	return nil
//}
//
//func (s *RuntimeChangeStorage) Retrieve(ref *runtime.AggregateRef, aggregate Aggregate) error {
//	if !(s.input.AggregateType == ref.Type && s.input.AggregateID == ref.ID) {
//		return errors.New(fmt.Sprintf("Aggregate not found"))
//	}
//
//	// TODO get read of ths explicit type
//	state := &OrderAggregateState{}
//	err := json.Unmarshal(s.input.Snapshot, state)
//	if err != nil {
//		return err
//	}
//
//	return aggregate.Hydrate(state, ref)
//}

func main() {
	// what if data is polymorfic
	// what if everything is an networks call
	// what if response contains data to be persisted, events to be publish, ...
	// what if storage is not a user defined concern?
	// what if external communication also is HTTP base, and persisting results must always happen in aggregates
	// and external information can be only delivered as a webhook

	app := httpruntime.NewMuxRuntimeClient()
	app.
		HandleFunc("/schedule/time", ScheduleInvokeHandlerFunc(func(cmd *httpruntime.ScheduleInvokeCMD, result *httpruntime.ScheduleInvokeResult) {
			result.Logs = time.Now().String()
		})).
		Schedule("* * * * *")

	//app.
	//	HandleFunc("/order/create", func(w http.ResponseWriter, rq *http.Request) {
	//		input := &runtime.AggregateChangeCMD{}
	//		output := &runtime.AggregateChangeResult{}
	//		store := &RuntimeChangeStorage{
	//			input:  input,
	//			output: output,
	//		}
	//		JsonRequestResponse(w, rq, input, output, func() {
	//			cmd := &OrderCreateCMD{}
	//			_ = json.Unmarshal(input.Payload, cmd)
	//
	//			aggregate := NewOrderAggregate()
	//			err := aggregate.Handle(cmd)
	//			if err != nil {
	//				output.Err = err.Error()
	//				output.Logs += "error from ApplyCommand"
	//				return
	//			}
	//
	//			err = store.Persist(aggregate)
	//			if err != nil {
	//				output.Err = err.Error()
	//				output.Logs += "error from ApplyCommand"
	//			}
	//		})
	//	}).
	//	AggregateChange("order", "create")
	//
	//app.
	//	HandleFunc("/order/colpay", func(w http.ResponseWriter, rq *http.Request) {
	//		input := &runtime.AggregateChangeCMD{}
	//		output := &runtime.AggregateChangeResult{}
	//		store := &RuntimeChangeStorage{
	//			input:  input,
	//			output: output,
	//		}
	//		JsonRequestResponse(w, rq, input, output, func() {
	//			aggregate := NewOrderAggregate()
	//			err := store.Retrieve(&runtime.AggregateRef{
	//				input.AggregateID,
	//				input.AggregateType,
	//			}, aggregate)
	//
	//			if err != nil {
	//				output.Err = err.Error()
	//				output.Logs += "error when store.Retrieve"
	//				return
	//			}
	//
	//			cmd := &OrderCollectPaymentsCMD{}
	//			_ = json.Unmarshal(input.Payload, cmd)
	//
	//			err = aggregate.Handle(cmd)
	//			if err != nil {
	//				output.Err = err.Error()
	//				output.Logs += fmt.Sprintf("cmd=%v", cmd)
	//				output.Logs += string(input.Snapshot)
	//				output.Logs += "error from ApplyCommand"
	//				return
	//			}
	//
	//			err = store.Persist(aggregate)
	//			if err != nil {
	//				output.Err = err.Error()
	//				output.Logs += "error from ApplyCommand"
	//			}
	//		})
	//	}).
	//	AggregateChange("order", "collect-payment")

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

func ScheduleInvokeHandlerFunc(handle func(*httpruntime.ScheduleInvokeCMD, *httpruntime.ScheduleInvokeResult)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		input := &httpruntime.ScheduleInvokeCMD{}
		output := &httpruntime.ScheduleInvokeResult{}
		JsonRequestResponse(w, rq, input, output, func() {
			handle(input, output)
		})
	}
}
