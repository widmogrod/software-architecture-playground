package orderaggregate

import (
	"errors"
	"fmt"
)

func (o *OrderAggregate) Apply(change interface{}) error {
	switch c := change.(type) {
	case *OrderCreated:
		if o.state != nil {
			return errors.New("order cannot be created twice, check your logic")
		}

		// when everything is ok, record changes that you want to make
		o.state = &OrderAggregateState{}
		o.state.UserID = c.UserID
		o.state.OrderID = c.OrderID
		o.state.OrderCreatedAt = c.CreatedAt

	case *ProductAdded:
		if o.state == nil {
			return errors.New("You cannot add products to not created order")
		}

		o.state.ProductID = c.ProductID
		o.state.ProductQuantity = c.Quantity

	case *OrderPaymentsCollected:
		if o.state == nil {
			return errors.New("You cannot collect payment for order that don't exists")
		}

		o.state.PaymentCollected = true

	default:
		return errors.New(fmt.Sprintf("unsupported type to handle %T", change))
	}

	return nil
}
