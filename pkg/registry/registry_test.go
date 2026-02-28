package registry

import (
	"testing"
)

// TestRegistry_SetGet verifies basic Set and Get operations.
func TestRegistry_SetGet(t *testing.T) {
	r := New()
	r.Set("key1", "value1")

	val, err := r.Get("key1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

// TestRegistry_GetNotFound verifies error for missing keys.
func TestRegistry_GetNotFound(t *testing.T) {
	r := New()

	_, err := r.Get("nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestRegistry_VersionTracking verifies version history is maintained.
func TestRegistry_VersionTracking(t *testing.T) {
	// TODO: implement after SetGet passes
}

// TestRegistry_GetVersionNotFound verifies error for missing versions.
func TestRegistry_GetVersionNotFound(t *testing.T) {
	// TODO: implement after VersionTracking passes
}

// TestRegistry_CloneUnderLock verifies snapshot consistency.
func TestRegistry_CloneUnderLock(t *testing.T) {
	// TODO: implement after VersionTracking passes
}

// TestRegistry_PreLoadHooks verifies pre-load hook execution.
func TestRegistry_PreLoadHooks(t *testing.T) {
	// TODO: implement after basic registry tests pass
}

// TestRegistry_PreLoadHookError verifies errors propagate from pre-load hooks.
func TestRegistry_PreLoadHookError(t *testing.T) {
	// TODO: implement after PreLoadHooks passes
}

// TestRegistry_PostLoadHooks verifies post-load hook execution.
func TestRegistry_PostLoadHooks(t *testing.T) {
	// TODO: implement after PreLoadHooks passes
}

// TestRegistry_PostLoadHookMultiple verifies multiple post-load hooks execute.
func TestRegistry_PostLoadHookMultiple(t *testing.T) {
	// TODO: implement after PostLoadHooks passes
}
