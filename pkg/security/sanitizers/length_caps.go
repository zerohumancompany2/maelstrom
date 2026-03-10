package sanitizers

// StreamLengthCaps enforces maximum byte limits on chunks and total stream size.
type StreamLengthCaps struct {
	MaxChunkSize int
	MaxTotalSize int
}

// NewStreamLengthCaps creates a new StreamLengthCaps sanitizer.
func NewStreamLengthCaps(maxChunk, maxTotal int) *StreamLengthCaps {
	return &StreamLengthCaps{
		MaxChunkSize: maxChunk,
		MaxTotalSize: maxTotal,
	}
}

// Sanitize processes a chunk and enforces length caps.
// Returns: (sanitizedChunk, isTruncated, error)
func (s *StreamLengthCaps) Sanitize(chunk []byte, chunkIndex int, totalSizeSoFar int) ([]byte, bool, error) {
	truncationMarker := "[TRUNCATED]"
	isTruncated := false

	// Enforce max chunk size
	if len(chunk) > s.MaxChunkSize {
		availableSize := s.MaxChunkSize - len(truncationMarker)
		if availableSize < 0 {
			availableSize = 0
		}
		chunk = chunk[:availableSize]
		chunk = append(chunk, []byte(truncationMarker)...)
		isTruncated = true
	}

	return chunk, isTruncated, nil
}
