package orderaggregate

import "time"

type OrderAggregateState struct {
	OrderID        string `uri:"corp:product:order:id"`
	OrderCreatedAt *time.Time

	UserID string `uri:"corp:product:user:id"`

	OrderTotalPrice string

	Created *time.Time
	Updated *time.Time

	ProductID        string
	ProductUnitPrice string

	ProductQuantity string
	WarehouseStatus string

	WarehouseReservationID string
	ShippingStatus         string

	ShippingID       string
	isOrderCreated   bool
	PaymentCollected bool
}
