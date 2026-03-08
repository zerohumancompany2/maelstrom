package testutil

import (
	"fmt"
	"github.com/maelstrom/v3/pkg/mail"
	"time"
)

// GenerateMailID generates a deterministic mail ID.
func GenerateMailID(index int) string {
	return fmt.Sprintf("mail-%d", index)
}

// GenerateCorrelationID generates a deterministic correlation ID.
func GenerateCorrelationID(index int) string {
	return fmt.Sprintf("corr-%d", index)
}

// GenerateMail creates a mail with deterministic values.
func GenerateMail(index int) mail.Mail {
	return mail.Mail{
		ID:            GenerateMailID(index),
		CorrelationID: GenerateCorrelationID(index),
		Type:          mail.Heartbeat,
		Source:        fmt.Sprintf("agent-%d", index%3),
		Target:        fmt.Sprintf("agent-%d", (index+1)%3),
		Content:       []byte(fmt.Sprintf("content-%d", index)),
		CreatedAt:     time.Now(),
	}
}

// GenerateServiceName generates a deterministic service name.
func GenerateServiceName(index int) string {
	return fmt.Sprintf("service-%d", index)
}

// GenerateService creates a service with deterministic values.
func GenerateService(index int) *MockService {
	return NewMockService(GenerateServiceName(index))
}

// TimeNow returns current time for testing.
func TimeNow() time.Time {
	return time.Now()
}
