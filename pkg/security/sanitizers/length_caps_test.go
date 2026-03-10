package sanitizers

import (
	"strings"
	"testing"
)

func TestStreamLengthCaps_ChunkSize(t *testing.T) {
	// Given: A chunk of 15KB data with max chunk size set to 10KB
	maxChunkSize := 10 * 1024      // 10KB
	maxTotalSize := 100 * 1024     // 100KB
	chunk := make([]byte, 15*1024) // 15KB chunk

	sanitizer := NewStreamLengthCaps(maxChunkSize, maxTotalSize)

	// When: StreamLengthCaps sanitizer processes the chunk
	result, isTruncated, err := sanitizer.Sanitize(chunk, 0, 0)

	// Then: Chunk is truncated to 10KB with "[TRUNCATED]" marker appended
	if err != nil {
		t.Fatalf("Sanitize returned error: %v", err)
	}

	if !isTruncated {
		t.Errorf("Expected isTruncated to be true, got false")
	}

	expectedSize := maxChunkSize
	if len(result) != expectedSize {
		t.Errorf("Expected output chunk size to be %d bytes, got %d", expectedSize, len(result))
	}

	truncationMarker := "[TRUNCATED]"
	if !strings.HasSuffix(string(result), truncationMarker) {
		t.Errorf("Expected output to end with %q, got %q", truncationMarker, string(result[len(result)-20:]))
	}
}
