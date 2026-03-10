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

func TestStreamLengthCaps_TotalStreamSize(t *testing.T) {
	// Given: A stream of 3 chunks (5KB each) with max total size set to 12KB
	maxChunkSize := 10 * 1024 // 10KB
	maxTotalSize := 12 * 1024 // 12KB
	chunkSize := 5 * 1024     // 5KB

	sanitizer := NewStreamLengthCaps(maxChunkSize, maxTotalSize)

	var totalSize int
	var results [][]byte

	// Process 3 chunks
	for i := 0; i < 3; i++ {
		chunk := make([]byte, chunkSize)
		result, _, err := sanitizer.Sanitize(chunk, i, totalSize)
		if err != nil {
			t.Fatalf("Sanitize chunk %d returned error: %v", i, err)
		}
		results = append(results, result)
		totalSize += len(result)
	}

	// Then: Total stream size is exactly 12KB, truncation marker in final chunk
	if totalSize != maxTotalSize {
		t.Errorf("Expected total stream size to be %d bytes, got %d", maxTotalSize, totalSize)
	}

	// Third chunk should be truncated
	truncationMarker := "[TRUNCATED]"
	if !strings.HasSuffix(string(results[2]), truncationMarker) {
		t.Errorf("Expected third chunk to end with %q, got %q", truncationMarker, string(results[2][len(results[2])-20:]))
	}
}
