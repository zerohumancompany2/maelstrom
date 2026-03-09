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
