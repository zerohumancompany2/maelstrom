package persistence

import (
	"testing"
	"time"
)

// TestStorageBackend_SaveSnapshot
// Spec Reference: arch-v1.md L468 (Snapshots), L486 (snapshot(runtimeId))
// Given: A StorageBackend instance and a snapshot to save
// When: SaveSnapshot() is called with snapshot
// Then: Snapshot saved with unique ID
// Expected Result: Snapshot saved successfully, retrievable by ID
func TestStorageBackend_SaveSnapshot(t *testing.T) {
	backend := NewStorageBackend().(*storageBackend)

	state := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	snap := Snapshot{
		ID:        "test-snap-1",
		RuntimeID: "runtime-1",
		State:     state,
		Taints:    []string{"taint1", "taint2"},
		Timestamp: time.Now(),
	}

	err := backend.SaveSnapshot(snap)

	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	// Verify snapshot was saved
	stored, ok := backend.snapshots[snap.ID]
	if !ok {
		t.Error("Expected snapshot to be saved with ID")
	}

	if stored.ID != snap.ID {
		t.Errorf("Expected snapshot ID %s, got %s", snap.ID, stored.ID)
	}

	if stored.RuntimeID != snap.RuntimeID {
		t.Errorf("Expected RuntimeID %s, got %s", snap.RuntimeID, stored.RuntimeID)
	}

	if len(stored.Taints) != len(snap.Taints) {
		t.Errorf("Expected %d taints, got %d", len(snap.Taints), len(stored.Taints))
	}
}

// TestStorageBackend_LoadSnapshot
// Spec Reference: arch-v1.md L468 (Snapshots), L486 (restore(snapshotId))
// Given: A StorageBackend instance with a previously saved snapshot
// When: LoadSnapshot() is called with snapshot ID
// Then: Snapshot loaded correctly with all data preserved
// Expected Result: All snapshot data preserved, runtime state intact, taints included
func TestStorageBackend_LoadSnapshot(t *testing.T) {
	backend := NewStorageBackend().(*storageBackend)

	state := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	originalSnap := Snapshot{
		ID:        "test-snap-load",
		RuntimeID: "runtime-load-1",
		State:     state,
		Taints:    []string{"taint1", "taint2"},
		Timestamp: time.Now(),
	}

	err := backend.SaveSnapshot(originalSnap)
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	loaded, err := backend.LoadSnapshot(originalSnap.ID)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	if loaded.ID != originalSnap.ID {
		t.Errorf("Expected snapshot ID %s, got %s", originalSnap.ID, loaded.ID)
	}

	if loaded.RuntimeID != originalSnap.RuntimeID {
		t.Errorf("Expected RuntimeID %s, got %s", originalSnap.RuntimeID, loaded.RuntimeID)
	}

	if len(loaded.Taints) != len(originalSnap.Taints) {
		t.Errorf("Expected %d taints, got %d", len(originalSnap.Taints), len(loaded.Taints))
	}

	for i, taint := range originalSnap.Taints {
		if loaded.Taints[i] != taint {
			t.Errorf("Expected taint %s at index %d, got %s", taint, i, loaded.Taints[i])
		}
	}
}
