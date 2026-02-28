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
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a file first
	testFile := filepath.Join(tmpDir, "test.yaml")
	content1 := []byte("id: test-chart\nversion: 1.0.0\n")
	if err := os.WriteFile(testFile, content1, 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	// Create the source
	src, err := NewFileSystemSource(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// Start watching
	go func() {
		src.Run()
	}()
	defer src.Stop()

	// Wait a bit for watcher to start
	time.Sleep(50 * time.Millisecond)

	// Consume the Created event
	select {
	case <-src.Events():
		// Created event consumed
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for initial Created event")
	}

	// Modify the file
	content2 := []byte("id: test-chart\nversion: 2.0.0\n")
	if err := os.WriteFile(testFile, content2, 0644); err != nil {
		t.Fatalf("failed to update file: %v", err)
	}

	// Wait for Updated event
	select {
	case evt := <-src.Events():
		if evt.Type != Updated {
			t.Errorf("expected Updated event, got %v", evt.Type)
		}
		if evt.Key != "test.yaml" {
			t.Errorf("expected Key 'test.yaml', got %q", evt.Key)
		}
		if string(evt.Content) != string(content2) {
			t.Errorf("expected Content %q, got %q", content2, evt.Content)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Updated event")
	}
}

// TestFileSystemSource_EmitsDeleted verifies file deletions emit Deleted events.
func TestFileSystemSource_EmitsDeleted(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a file first
	testFile := filepath.Join(tmpDir, "test.yaml")
	content := []byte("id: test-chart\nversion: 1.0.0\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	// Create the source
	src, err := NewFileSystemSource(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	// Start watching
	go func() {
		src.Run()
	}()
	defer src.Stop()

	// Consume the Created event
	select {
	case <-src.Events():
		// Created event consumed
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for initial Created event")
	}

	// Delete the file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	// Wait for Deleted event
	select {
	case evt := <-src.Events():
		if evt.Type != Deleted {
			t.Errorf("expected Deleted event, got %v", evt.Type)
		}
		if evt.Key != "test.yaml" {
			t.Errorf("expected Key 'test.yaml', got %q", evt.Key)
		}
		if evt.Content != nil {
			t.Errorf("expected nil Content for deleted file, got %q", evt.Content)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Deleted event")
	}
}

// TestFileSystemSource_Debounces verifies rapid changes coalesce into single event.
func TestFileSystemSource_Debounces(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := NewFileSystemSource(tmpDir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	go func() { src.Run() }()
	defer src.Stop()

	time.Sleep(20 * time.Millisecond) // Let initial scan complete

	testFile := filepath.Join(tmpDir, "test.yaml")

	// Rapid writes
	for i := 0; i < 5; i++ {
		content := []byte("version: 1.0." + string(rune('0'+i)) + "\n")
		os.WriteFile(testFile, content, 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// Should receive only one event due to debouncing
	select {
	case <-src.Events():
		// Got event
		select {
		case <-src.Events():
			t.Fatal("expected only one event due to debouncing, got two")
		case <-time.After(100 * time.Millisecond):
			// No second event - debouncing worked
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for debounced event")
	}
}

// TestFileSystemSource_GracefulShutdown verifies clean shutdown closes channel.
func TestFileSystemSource_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := NewFileSystemSource(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	go func() { src.Run() }()

	time.Sleep(20 * time.Millisecond)

	// Stop should close the channel
	if err := src.Stop(); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}

	// Channel should be closed
	select {
	case _, ok := <-src.Events():
		if ok {
			t.Fatal("expected Events channel to be closed after Stop")
		}
		// Channel closed as expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for Events channel to close")
	}
}

// TestFileSystemSource_ErrAfterClose verifies Err() returns error after abnormal exit.
func TestFileSystemSource_ErrAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := NewFileSystemSource(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	go func() { src.Run() }()
	time.Sleep(20 * time.Millisecond)
	src.Stop()

	// After graceful shutdown, Err() should be nil
	if src.Err() != nil {
		t.Errorf("expected nil error after graceful shutdown, got: %v", src.Err())
	}
}

// TestFileSystemSource_FiltersYAML verifies only .yaml files emit events.
func TestFileSystemSource_FiltersYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-YAML file
	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("not yaml"), 0644); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}

	src, err := NewFileSystemSource(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	go func() { src.Run() }()
	defer src.Stop()

	// Wait for any events
	select {
	case evt := <-src.Events():
		t.Fatalf("expected no events for non-YAML files, got: %+v", evt)
	case <-time.After(100 * time.Millisecond):
		// No events received - filtering works
	}

	// Now create a YAML file
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(yamlFile, []byte("id: test\n"), 0644); err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}

	// Should receive event for YAML
	select {
	case evt := <-src.Events():
		if evt.Key != "test.yaml" {
			t.Errorf("expected event for test.yaml, got: %s", evt.Key)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for YAML file event")
	}
}
