package testutil

import (
	"fmt"
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
func GenerateMail(index int) Mail {
	return Mail{
		ID:            GenerateMailID(index),
		Type:          "COMMAND",
		From:          fmt.Sprintf("agent-%d", index%3),
		To:            fmt.Sprintf("agent-%d", (index+1)%3),
		Content:       []byte(fmt.Sprintf("content-%d", index)),
		CorrelationID: GenerateCorrelationID(index),
		Timestamp:     time.Now(),
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
