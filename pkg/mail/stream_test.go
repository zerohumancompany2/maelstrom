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

func TestTaintStripping(t *testing.T) {
	chunk := StreamChunk{
		Data:   "test data",
		Taints: []string{"USER_SUPPLIED", "TOOL_OUTPUT", "INNER_BOUNDARY"},
	}

	// Test with allowed taints
	allowed := []string{"USER_SUPPLIED", "TOOL_OUTPUT"}
	stripped := StripForbiddenTaints(chunk, allowed)

	if len(stripped.Taints) != 2 {
		t.Errorf("Expected 2 taints, got %d", len(stripped.Taints))
	}

	// Verify correct taints preserved
	hasUserSupplied := false
	hasToolOutput := false
	for _, t := range stripped.Taints {
		if t == "USER_SUPPLIED" {
			hasUserSupplied = true
		}
		if t == "TOOL_OUTPUT" {
			hasToolOutput = true
		}
	}
	if !hasUserSupplied {
		t.Error("Expected USER_SUPPLIED to be preserved")
	}
	if !hasToolOutput {
		t.Error("Expected TOOL_OUTPUT to be preserved")
	}

	// Test with empty allowed list (strip all)
	chunk2 := StreamChunk{
		Data:   "test data",
		Taints: []string{"USER_SUPPLIED"},
	}
	stripped2 := StripForbiddenTaints(chunk2, []string{})
	if len(stripped2.Taints) != 0 {
		t.Errorf("Expected 0 taints with empty allowed list, got %d", len(stripped2.Taints))
	}
}

func TestStreamChunk_IsFinal(t *testing.T) {
	// Create a sequence of chunks
	chunks := []StreamChunk{
		{Data: "Part 1", Sequence: 1, IsFinal: false},
		{Data: "Part 2", Sequence: 2, IsFinal: false},
		{Data: "Part 3", Sequence: 3, IsFinal: true},
	}

	// Verify only last chunk is final
	for i, chunk := range chunks {
		expectedFinal := (i == len(chunks)-1)
		if chunk.IsFinal != expectedFinal {
			t.Errorf("Chunk %d: Expected IsFinal=%v, got %v", i, expectedFinal, chunk.IsFinal)
		}
	}

	// Test stream completion detection
	isComplete := false
	for _, chunk := range chunks {
		if chunk.IsFinal {
			isComplete = true
			break
		}
	}
	if !isComplete {
		t.Error("Expected to detect stream completion via IsFinal")
	}
}

func TestStreamChunk_Sequence(t *testing.T) {
	// Test sequence uniqueness
	chunks := []StreamChunk{
		{Data: "A", Sequence: 1},
		{Data: "B", Sequence: 2},
		{Data: "C", Sequence: 3},
	}

	seen := make(map[int]bool)
	for _, chunk := range chunks {
		if seen[chunk.Sequence] {
			t.Errorf("Duplicate sequence number: %d", chunk.Sequence)
		}
		seen[chunk.Sequence] = true
	}

	// Test sequence ordering
	for i := 1; i < len(chunks); i++ {
		if chunks[i].Sequence <= chunks[i-1].Sequence {
			t.Errorf("Sequence not in order: %d <= %d", chunks[i].Sequence, chunks[i-1].Sequence)
		}
	}

	// Test sequence starts at 1
	if chunks[0].Sequence != 1 {
		t.Errorf("Expected first sequence to be 1, got %d", chunks[0].Sequence)
	}
}

func TestStreamSession_Send(t *testing.T) {
	session := NewStreamSession("test-session", nil)
	chunk := StreamChunk{Data: "test data", Sequence: 1, IsFinal: false}

	err := session.Send(chunk)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	received, ok := <-session.Chunks
	if !ok {
		t.Error("Expected channel to be open and contain chunk")
	}

	if received.Data != chunk.Data {
		t.Errorf("Expected data 'test data', got '%s'", received.Data)
	}
	if received.Sequence != chunk.Sequence {
		t.Errorf("Expected sequence 1, got %d", received.Sequence)
	}
}

func TestStreamSession_SendMultiple(t *testing.T) {
	session := NewStreamSession("test-session", nil)
	chunks := []StreamChunk{
		{Data: "Part 1", Sequence: 1, IsFinal: false},
		{Data: "Part 2", Sequence: 2, IsFinal: false},
		{Data: "Part 3", Sequence: 3, IsFinal: true},
	}

	for _, chunk := range chunks {
		err := session.Send(chunk)
		if err != nil {
			t.Errorf("Expected nil error for chunk %d, got %v", chunk.Sequence, err)
		}
	}

	received := make([]StreamChunk, 0, 3)
	for i := 0; i < 3; i++ {
		receivedChunk, ok := <-session.Chunks
		if !ok {
			t.Errorf("Expected to receive chunk %d but channel was closed", i+1)
			break
		}
		received = append(received, receivedChunk)
	}

	if len(received) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(received))
	}

	for i, chunk := range chunks {
		if received[i].Data != chunk.Data {
			t.Errorf("Chunk %d: Expected data '%s', got '%s'", i+1, chunk.Data, received[i].Data)
		}
		if received[i].Sequence != chunk.Sequence {
			t.Errorf("Chunk %d: Expected sequence %d, got %d", i+1, chunk.Sequence, received[i].Sequence)
		}
	}
}

func TestStreamSession_Close(t *testing.T) {
	session := NewStreamSession("test-session", nil)

	err := session.Close()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if !session.Closed {
		t.Error("Expected session to be closed")
	}

	select {
	case _, ok := <-session.Chunks:
		if ok {
			t.Error("Expected channel to be closed")
		}
	default:
		// Channel may have no data, check if closed
		if !session.Closed {
			t.Error("Expected session to be marked as closed")
		}
	}
}

func TestStreamSession_CloseAfterSend(t *testing.T) {
	session := NewStreamSession("test-session", nil)

	chunks := []StreamChunk{
		{Data: "Part 1", Sequence: 1, IsFinal: false},
		{Data: "Part 2", Sequence: 2, IsFinal: true},
	}

	for _, chunk := range chunks {
		err := session.Send(chunk)
		if err != nil {
			t.Errorf("Expected nil error for chunk %d, got %v", chunk.Sequence, err)
		}
	}

	err := session.Close()
	if err != nil {
		t.Errorf("Expected nil error on Close, got %v", err)
	}

	if !session.Closed {
		t.Error("Expected session to be closed after Close()")
	}

	received := make([]StreamChunk, 0)
	for chunk := range session.Chunks {
		received = append(received, chunk)
	}

	if len(received) != 2 {
		t.Errorf("Expected 2 chunks before close, got %d", len(received))
	}

	err = session.Send(StreamChunk{Data: "after close", Sequence: 3})
	if err == nil {
		t.Error("Expected error when sending to closed session")
	}
}
