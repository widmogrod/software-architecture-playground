package aggregate

import "time"

type OrderAggregateState struct {
	OrderID        string
	OrderCreatedAt *time.Time

	UserID string

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
