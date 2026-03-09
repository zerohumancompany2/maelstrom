package mail

import "testing"

func TestMail_StreamChunkHasRequiredFields(t *testing.T) {
	chunk := StreamChunk{
		Data:     "test",
		Sequence: 42,
		IsFinal:  true,
		Taints:   []string{"test"},
	}

	if chunk.Data != "test" {
		t.Errorf("expected Data to be 'test', got '%s'", chunk.Data)
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

func TestMail_StreamChunkMigrationComplete(t *testing.T) {
	// Test 1: mail.StreamChunk type exists and is usable
	chunk := StreamChunk{
		Data:     "test",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"test-taint"},
	}
	if chunk.Data != "test" {
		t.Error("mail.StreamChunk should be usable")
	}

	// Test 2: MailMetadata.Stream field type is *StreamChunk (not bool)
	metadata := MailMetadata{}
	metadata.Stream = &chunk
	if metadata.Stream == nil {
		t.Error("MailMetadata.Stream should accept *StreamChunk")
	}

	// Test 3: Can assign nil to MailMetadata.Stream
	metadata.Stream = nil
	if metadata.Stream != nil {
		t.Error("MailMetadata.Stream should accept nil")
	}

	// Test 4: Verify Stream field is pointer type by checking assignment works
	metadata.Stream = &StreamChunk{Data: "final", Sequence: 2, IsFinal: true}
	if metadata.Stream.Data != "final" {
		t.Error("MailMetadata.Stream should be *StreamChunk type")
	}
}
