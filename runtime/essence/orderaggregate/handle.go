package orderaggregate

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
			ReduceRecent(o).Err

	case *OrderCollectPaymentsCMD:
		// validate necessary condition
		if o.state == nil {
			return errors.New("Order dont exists!")
		}
		if c.Method != "apple" {
			return errors.New(fmt.Sprintf("Accept ApplePay only %#v", c))
		}

		if c.Amount <= 0 {
			return errors.New(fmt.Sprintf("Payment must be positive amount %#v", c))
		}

		return o.changes.
			Append(&OrderPaymentsCollected{
				Method: c.Method,
				Amount: c.Amount,
				Date:   c.Date,
			}).Ok.
			ReduceRecent(o).Err
	}

	return errors.New(fmt.Sprintf("Invalid command: %T", cmd))
}
