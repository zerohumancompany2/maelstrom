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
