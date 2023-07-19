package event

import (
	"time"

	"github.com/daichitakahashi/go-enum"
)

//go:generate go run github.com/daichitakahashi/go-enum/cmd/enumgen@latest

type (
	Event interface {
		ID() string
	}

	OrderPlaced struct {
		enum.MemberOf[Event]
		Items []string
	}

	PaymentReceived struct {
		enum.MemberOf[Event]
		Amount int
	}

	ItemShipped struct {
		enum.MemberOf[Event]
		ShippedAt time.Time
	}
)

// ID implements Event.
func (OrderPlaced) ID() string {
	return "orderPlaced"
}

// ID implements Event.
func (PaymentReceived) ID() string {
	return "paymentReceived"
}

// ID implements Event.
func (ItemShipped) ID() string {
	return "itemShipped"
}

var (
	_ Event = OrderPlaced{}
	_ Event = PaymentReceived{}
	_ Event = ItemShipped{}
)
