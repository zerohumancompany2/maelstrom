package mail

import "testing"

func TestMail_StreamChunkHasRequiredFields(t *testing.T) {
	chunk := StreamChunk{
		Chunk:     "test",
		Sequence: 42,
		IsFinal:  true,
		Taints:   []string{"test"},
	}

	if chunk.Chunk != "test" {
		t.Errorf("expected Data to be 'test', got '%s'", chunk.Chunk)
	}
	if chunk.Sequence != 42 {
		t.Errorf("expected Sequence to be 42, got %d", chunk.Sequence)
	}
	if !chunk.IsFinal {
		t.Error("expected IsFinal to be true")
	}
	if len(chunk.Taints) != 1 || chunk.Taints[0] != "test" {
		t.Errorf("expected Taints to be ['test'], got %v", chunk.Taints)
	}
}

func TestMailMetadata_StreamBool(t *testing.T) {
	metadata := MailMetadata{}
	metadata.Stream = true
	if !metadata.Stream {
		t.Error("MailMetadata.Stream should be bool type and settable to true")
	}
	metadata.Stream = false
	if metadata.Stream {
		t.Error("MailMetadata.Stream should be bool type and settable to false")
	}
}

func TestMail_StreamChunkMigrationComplete(t *testing.T) {
	// Test 1: mail.StreamChunk type exists and is usable
	chunk := StreamChunk{
		Chunk:     "test",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"test-taint"},
	}
	if chunk.Chunk != "test" {
		t.Error("mail.StreamChunk should be usable")
	}

	// Test 2: MailMetadata.StreamChunk field type is *StreamChunk
	metadata := MailMetadata{}
	metadata.StreamChunk = &chunk
	if metadata.StreamChunk == nil {
		t.Error("MailMetadata.StreamChunk should accept *StreamChunk")
	}

	// Test 3: Can assign nil to MailMetadata.StreamChunk
	metadata.StreamChunk = nil
	if metadata.StreamChunk != nil {
		t.Error("MailMetadata.StreamChunk should accept nil")
	}

	// Test 4: Verify StreamChunk field is pointer type by checking assignment works
	metadata.StreamChunk = &StreamChunk{Chunk: "final", Sequence: 2, IsFinal: true}
	if metadata.StreamChunk.Chunk != "final" {
		t.Error("MailMetadata.StreamChunk should be *StreamChunk type")
	}
}

func TestMailMetadata_TypeAlignment(t *testing.T) {
	metadata := MailMetadata{
		Tokens:      100,
		Model:       "gpt-4",
		Cost:        0.05,
		Boundary:    InnerBoundary,
		Taints:      []string{"TEST"},
		Stream:      true,
		StreamChunk: &StreamChunk{Chunk: "chunk", Sequence: 1},
		IsFinal:     false,
	}

	if metadata.Tokens != 100 {
		t.Error("Tokens field mismatch")
	}
	if metadata.Model != "gpt-4" {
		t.Error("Model field mismatch")
	}
	if metadata.Cost != 0.05 {
		t.Error("Cost field mismatch")
	}
	if metadata.Boundary != InnerBoundary {
		t.Error("Boundary field mismatch")
	}
	if len(metadata.Taints) != 1 {
		t.Error("Taints field mismatch")
	}
	if !metadata.Stream {
		t.Error("Stream field mismatch")
	}
	if metadata.StreamChunk.Chunk != "chunk" {
		t.Error("StreamChunk field mismatch")
	}
	if metadata.IsFinal {
		t.Error("IsFinal field mismatch")
	}
}
