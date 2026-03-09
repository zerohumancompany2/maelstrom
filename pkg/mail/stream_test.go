package mail

import "testing"

func stringPtr(s string) *string {
	return &s
}

func TestStreamUpgrade(t *testing.T) {
	sessionID := "session-001"
	lastEventID := stringPtr("event-100")

	ch, err := UpgradeToStream(sessionID, lastEventID)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ch == nil {
		t.Error("Expected non-nil channel")
	}

	// Verify channel is buffered
	select {
	case ch <- StreamChunk{Data: "test", Sequence: 1}:
		// Success
	default:
		t.Error("Expected channel to accept chunk without blocking")
	}
}
