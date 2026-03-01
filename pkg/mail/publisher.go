package mail

import "time"

type Publisher interface {
	Publish(mail Mail) (Ack, error)
}

type Ack struct {
	CorrelationID string
	DeliveredAt   time.Time
}
