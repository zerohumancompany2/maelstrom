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

func TestStreamChunkFormat(t *testing.T) {
	chunk := StreamChunk{
		Data:     "Hello, ",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"USER_SUPPLIED"},
	}

	if chunk.Data != "Hello, " {
		t.Errorf("Expected Data 'Hello, ', got '%s'", chunk.Data)
	}
	if chunk.Sequence != 1 {
		t.Errorf("Expected Sequence 1, got %d", chunk.Sequence)
	}
	if chunk.IsFinal {
		t.Error("Expected IsFinal to be false")
	}
	if len(chunk.Taints) != 1 {
		t.Errorf("Expected 1 taint, got %d", len(chunk.Taints))
	}

	// Test final chunk
	finalChunk := StreamChunk{
		Data:     "world!",
		Sequence: 2,
		IsFinal:  true,
		Taints:   []string{},
	}

	if !finalChunk.IsFinal {
		t.Error("Expected IsFinal to be true")
	}
	if finalChunk.Sequence != 2 {
		t.Errorf("Expected Sequence 2, got %d", finalChunk.Sequence)
	}
}
