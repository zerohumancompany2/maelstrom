package observability

import (
	"github.com/maelstrom/v3/pkg/mail"
	"time"
)

type DeadLetterEntry struct {
	Mail   mail.Mail
	Reason string
	Logged time.Time
}

type DeadLetterFilters struct {
	RuntimeID string
	FromTime  time.Time
	ToTime    time.Time
	Reason    string
}
