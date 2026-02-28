package source

import "time"

// EventType represents the kind of change detected by a Source.
type EventType int

const (
	// Created indicates a new file was detected.
	Created EventType = iota
	// Updated indicates an existing file was modified.
	Updated
	// Deleted indicates a file was removed.
	Deleted
)

// SourceEvent represents a file change event from a Source.
type SourceEvent struct {
	Key       string    // Relative path, e.g., "gateway.yaml"
	Content   []byte    // Raw YAML content (empty for Deleted)
	Type      EventType // Created, Updated, or Deleted
	Timestamp time.Time
}

// Source decouples event producers from consumers.
// Implementations: FileSystemSource, HTTPSource, TestSource, etc.
type Source interface {
	// Events returns a receive-only channel of file changes.
	// The Source owns this channel and closes it on shutdown.
	Events() <-chan SourceEvent

	// Err returns any error after graceful shutdown.
	// Call this after the Events channel is closed.
	Err() error
}
