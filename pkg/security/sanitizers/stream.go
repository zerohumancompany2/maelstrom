package sanitizers

import "strings"

// StreamChunk represents a chunk in a streaming response
type StreamChunk struct {
	Chunk    string
	Sequence int
	IsFinal  bool
	Taints   []string
}

// StreamSanitizer sanitizes stream chunks per-chunk with <50ms latency
type StreamSanitizer struct {
	Redactor          *PIIRedactor
	LengthCapper      *LengthCapper
	SchemaValidator   *SchemaValidator
	InnerDataStripper *InnerDataStripper
}

// NewStreamSanitizer creates a new StreamSanitizer
func NewStreamSanitizer() *StreamSanitizer {
	return &StreamSanitizer{
		Redactor:          NewPIIRedactor(),
		LengthCapper:      NewLengthCapper(1000),
		SchemaValidator:   NewSchemaValidator(),
		InnerDataStripper: NewInnerDataStripper(),
	}
}

// SanitizeChunk sanitizes a single chunk (stateless, <50ms latency)
func (ss *StreamSanitizer) SanitizeChunk(chunk StreamChunk) (StreamChunk, error) {
	result := chunk.Chunk

	if ss.Redactor != nil {
		result = ss.Redactor.Redact(result)
	}

	if ss.InnerDataStripper != nil && hasTaint(chunk.Taints, "INNER_ONLY") {
		result = ss.InnerDataStripper.Strip(result)
	}

	if ss.LengthCapper != nil {
		result = ss.LengthCapper.Cap(result)
	}

	return StreamChunk{
		Chunk:    result,
		Sequence: chunk.Sequence,
		IsFinal:  chunk.IsFinal,
		Taints:   chunk.Taints,
	}, nil
}

func hasTaint(taints []string, target string) bool {
	for _, t := range taints {
		if t == target {
			return true
		}
	}
	return false
}

func init() {
	_ = strings.Contains
}
