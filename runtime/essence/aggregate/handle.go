package aggregate

import (
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"time"
)

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
				UserID:    c.UserID,
				CreatedAt: &now,
			}).Ok.
			Append(&ProductAdded{
				ProductID: c.ProductID,
				Quantity:  c.Quantity,
			}).Ok.
			Reducer(o).Err

	case *OrderCollectPaymentsCMD:
		// validate necessary condition
		if o.state == nil {
			return errors.New("Order dont exists!")
		}
		if c.OrderID != o.state.OrderID {
			return errors.New(fmt.Sprintf("Order missmatch %#v", c))
		}

		return o.changes.
			Append(&OrderPaymentsCollected{
				PaymentCollected: true,
			}).Ok.
			Reducer(o).Err
	}

	return errors.New(fmt.Sprintf("Invalid command: %T", cmd))
}
