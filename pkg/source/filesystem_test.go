package source

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFileSystemSource_EmitsCreated verifies that creating a YAML file
// emits a Created event with the correct key and content.
func TestFileSystemSource_EmitsCreated(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create the source with 100ms debounce
	src, err := NewFileSystemSource(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// Start watching in a goroutine
	go func() {
		if err := src.Run(); err != nil {
			// Run may return error on shutdown, that's OK
		}
	}()

	// Ensure cleanup
	defer src.Stop()

	// Wait a bit for the watcher to start
	time.Sleep(50 * time.Millisecond)

	// Write a new file
	testFile := filepath.Join(tmpDir, "test.yaml")
	content := []byte("id: test-chart\nversion: 1.0.0\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Wait for the event with timeout
	select {
	case evt := <-src.Events():
		if evt.Type != Created {
			t.Errorf("expected Created event, got %v", evt.Type)
		}
		if evt.Key != "test.yaml" {
			t.Errorf("expected Key 'test.yaml', got %q", evt.Key)
		}
		if string(evt.Content) != string(content) {
			t.Errorf("expected Content %q, got %q", content, evt.Content)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Created event")
	}
}

// TestFileSystemSource_EmitsUpdated verifies file modifications emit Updated events.
func TestFileSystemSource_EmitsUpdated(t *testing.T) {
	// TODO: implement after EmitsCreated passes
}

// TestFileSystemSource_EmitsDeleted verifies file deletions emit Deleted events.
func TestFileSystemSource_EmitsDeleted(t *testing.T) {
	// TODO: implement after EmitsCreated passes
}

// TestFileSystemSource_Debounces verifies rapid changes coalesce into single event.
func TestFileSystemSource_Debounces(t *testing.T) {
	// TODO: implement after EmitsCreated passes
}

// TestFileSystemSource_GracefulShutdown verifies clean shutdown closes channel.
func TestFileSystemSource_GracefulShutdown(t *testing.T) {
	// TODO: implement after EmitsCreated passes
}

// TestFileSystemSource_ErrAfterClose verifies Err() returns error after abnormal exit.
func TestFileSystemSource_ErrAfterClose(t *testing.T) {
	// TODO: implement after EmitsCreated passes
}

// TestFileSystemSource_FiltersYAML verifies only .yaml files emit events.
func TestFileSystemSource_FiltersYAML(t *testing.T) {
	// TODO: implement after EmitsCreated passes
}
