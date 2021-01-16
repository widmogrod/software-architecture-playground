package aggregate

import (
	"time"
)

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
	UserID    string
	CreatedAt *time.Time
}

type OrderCollectPaymentsCMD struct {
	OrderID string
}
type OrderPaymentsCollected struct {
	PaymentCollected bool
}
