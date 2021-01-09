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

func NewOrderAggregate() *OrderAggregate {
	store := runtime.NewEventStore()
	aggregate := &OrderAggregate{
		state:   nil,
		changes: store,
		ref: &runtime.AggregateRef{
			ID:   "",
			Type: "order",
		},
	}

	return aggregate
}

type OrderAggregate struct {
	state   *OrderAggregateState
	changes *runtime.EventStore
	ref     *runtime.AggregateRef
}

func (o *OrderAggregate) Ref() *runtime.AggregateRef {
	return o.ref
}

func (o *OrderAggregate) State() interface{} {
	return o.state
}

func (o *OrderAggregate) Changes() *runtime.EventStore {
	return o.changes
}

func (o *OrderAggregate) Handle(cmd interface{}) error {
	switch c := cmd.(type) {
	case *OrderCreateCMD:
		// validate necessary condition
		if o.state != nil {
			return errors.New("Order already exists!")
		}
		if c.Quantity == "" {
			return errors.New(fmt.Sprintf("Given quantity is to low %v", c))
		}

		now := time.Now()
		return o.changes.
			Append(&OrderCreated{
				OrderID:   ksuid.New().String(),
				UserId:    c.UserID,
				CreatedAt: &now,
			}).Ok.
			Append(&ProductAdded{
				ProductID: c.ProductID,
				Quantity:  c.Quantity,
			}).Ok.
			Reducer(o).Err
	}

	return nil
}

func (o *OrderAggregate) Apply(change interface{}) error {
	switch c := change.(type) {
	case *OrderCreated:
		if o.state != nil {
			return errors.New("order cannot be created twice, check your logic")
		}

		o.ref.ID = c.OrderID

		// when everything is ok, record changes that you want to make
		o.state = &OrderAggregateState{}
		o.state.OrderID = c.OrderID
		o.state.OrderCreatedAt = c.CreatedAt

		o.state.isOrderCreated = true

	case *ProductAdded:
		if !o.state.isOrderCreated {
			return errors.New("You cannot add products to not created order")
		}

		o.state.ProductID = c.ProductID
		o.state.ProductQuantity = c.Quantity

	default:
		return errors.New(fmt.Sprintf("unsupported type to handle %T", change))
	}

	return nil
}

type OrderAggregateState struct {
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

type ProductAdded struct {
	ProductID string
	Quantity  string
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

func NewMarchaller() *ChangesRegistry {
	return &ChangesRegistry{
		changes: map[string]interface{}{
			"OrderCreated": &OrderCreated{},
			"ProductAdded": &ProductAdded{},
		},
	}
}

type ChangesRegistry struct {
	changes map[string]interface{}
}

func (m *ChangesRegistry) SetChange(name string, change interface{}) {
	m.changes[name] = change
}

type RuntimeChangeStorage struct {
	input  *runtime.AggregateChangeCMD
	output *runtime.AggregateChangeResult
	//registry *ChangesRegistry
}

type Aggregate interface {
	Changes() *runtime.EventStore
	State() interface{}
	Ref() *runtime.AggregateRef
}

func (s *RuntimeChangeStorage) Persist(a Aggregate) error {
	_ = a.Changes().Reduce(func(change interface{}, result *runtime.Reduced) *runtime.Reduced {
		data, _ := json.Marshal(change)

		now := time.Now()
		s.output.Changes = append(s.output.Changes, &runtime.AggregateChange{
			// TODO change name can be provided by Event
			Type:       reflect.TypeOf(change).Name(),
			Payload:    data,
			RecordTime: &now,
			// TODO introduce version of change
			//Version: 1
		})

		return result
	}, nil).Err

	data, err := json.Marshal(a.State())
	if err != nil {
		return err
	}

	s.output.AggregateType = a.Ref().Type
	s.output.AggregateID = a.Ref().ID
	s.output.Snapshot = data
	return nil
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
		HandleFunc("/order/create", func(w http.ResponseWriter, rq *http.Request) {
			input := &runtime.AggregateChangeCMD{}
			output := &runtime.AggregateChangeResult{}
			store := &RuntimeChangeStorage{
				input:  input,
				output: output,
			}
			JsonRequestResponse(w, rq, input, output, func() {
				cmd := &OrderCreateCMD{}
				_ = json.Unmarshal(input.Payload, cmd)

				aggregate := NewOrderAggregate()
				err := aggregate.Handle(cmd)
				if err != nil {
					output.Err = err.Error()
					output.Logs += "error from ApplyCommand"
					return
				}

				err = store.Persist(aggregate)
				if err != nil {
					output.Err = err.Error()
					output.Logs += "error from ApplyCommand"
				}
			})
		}).
		AggregateChange("order", "create")

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
