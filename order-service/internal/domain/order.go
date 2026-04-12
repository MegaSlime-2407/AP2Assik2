package domain

import (
	"errors"
	"time"
)

type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64
	Status     string
	CreatedAt  time.Time
}

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidAmount    = errors.New("amount must be greater than zero")
	ErrCancelNotAllowed = errors.New("only pending orders can be cancelled")
	ErrPaymentFailed    = errors.New("payment service is unavailable")
)

func NewOrder(id, customerID, itemName string, amount int64) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	return &Order{
		ID:         id,
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     "Pending",
	}, nil
}

func (o *Order) Cancel() error {
	if o.Status != "Pending" {
		return ErrCancelNotAllowed
	}
	o.Status = "Cancelled"
	return nil
}

func (o *Order) MarkPaid() {
	o.Status = "Paid"
}

func (o *Order) MarkFailed() {
	o.Status = "Failed"
}
