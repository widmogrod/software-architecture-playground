package orderaggregate

import (
	"time"
)

type ProductAdded struct {
	ProductID string
	Quantity  string
}

type OrderCreateCMD struct {
	OrderID   string
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
	Method string
	Amount uint
	Date   *time.Time
}
type OrderPaymentsCollected struct {
	Method string
	Amount uint
	Date   *time.Time
}
