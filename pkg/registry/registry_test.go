package registry

import (
	"errors"
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
	r := New()

	// Set multiple versions
	r.Set("key1", "v1")
	r.Set("key1", "v2")
	r.Set("key1", "v3")

	// Get should return latest
	val, err := r.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "v3" {
		t.Errorf("expected 'v3' for Get, got %v", val)
	}

	// GetVersion should return specific versions
	v0, err := r.GetVersion("key1", 0)
	if err != nil {
		t.Fatalf("GetVersion(0) failed: %v", err)
	}
	if v0 != "v1" {
		t.Errorf("expected 'v1' for version 0, got %v", v0)
	}

	v1, err := r.GetVersion("key1", 1)
	if err != nil {
		t.Fatalf("GetVersion(1) failed: %v", err)
	}
	if v1 != "v2" {
		t.Errorf("expected 'v2' for version 1, got %v", v1)
	}

	v2, err := r.GetVersion("key1", 2)
	if err != nil {
		t.Fatalf("GetVersion(2) failed: %v", err)
	}
	if v2 != "v3" {
		t.Errorf("expected 'v3' for version 2, got %v", v2)
	}
}

// TestRegistry_GetVersionNotFound verifies error for missing versions.
func TestRegistry_GetVersionNotFound(t *testing.T) {
	r := New()
	r.Set("key1", "v1")

	// Valid version should work
	_, err := r.GetVersion("key1", 0)
	if err != nil {
		t.Errorf("GetVersion(0) should succeed, got: %v", err)
	}

	// Invalid version should return error
	_, err = r.GetVersion("key1", 5)
	if err != ErrVersionNotFound {
		t.Errorf("expected ErrVersionNotFound, got: %v", err)
	}

	// Negative version should return error
	_, err = r.GetVersion("key1", -1)
	if err != ErrVersionNotFound {
		t.Errorf("expected ErrVersionNotFound for negative version, got: %v", err)
	}
}

// TestRegistry_CloneUnderLock verifies snapshot consistency.
func TestRegistry_CloneUnderLock(t *testing.T) {
	r := New()
	r.Set("key1", "value1")
	r.Set("key2", "value2")

	var snapshot map[string]interface{}
	r.CloneUnderLock(func(s map[string]interface{}) {
		snapshot = s
	})

	// Verify snapshot has all entries
	if len(snapshot) != 2 {
		t.Errorf("expected 2 entries in snapshot, got %d", len(snapshot))
	}
	if snapshot["key1"] != "value1" {
		t.Errorf("expected 'value1' for key1, got %v", snapshot["key1"])
	}
	if snapshot["key2"] != "value2" {
		t.Errorf("expected 'value2' for key2, got %v", snapshot["key2"])
	}

	// Modifying snapshot should not affect registry
	snapshot["key1"] = "modified"
	val, _ := r.Get("key1")
	if val != "value1" {
		t.Error("CloneUnderLock snapshot was not a copy")
	}
}

// TestRegistry_PreLoadHooks verifies pre-load hook execution.
func TestRegistry_PreLoadHooks(t *testing.T) {
	r := New()

	// Track hook execution
	var hookCalled bool
	hook := func(key string, content []byte) ([]byte, error) {
		hookCalled = true
		// Add prefix to content
		return []byte("prefix:" + string(content)), nil
	}
	r.AddPreLoadHook(hook)

	hydrator := func(content []byte) (interface{}, error) {
		return string(content), nil
	}

	err := r.SetWithHooks("test", []byte("value"), hydrator)
	if err != nil {
		t.Fatalf("SetWithHooks failed: %v", err)
	}

	if !hookCalled {
		t.Error("pre-load hook was not called")
	}

	val, _ := r.Get("test")
	if val != "prefix:value" {
		t.Errorf("expected transformed content 'prefix:value', got %v", val)
	}
}

// TestRegistry_PreLoadHookError verifies errors propagate from pre-load hooks.
func TestRegistry_PreLoadHookError(t *testing.T) {
	r := New()

	hookErr := errors.New("pre-load error")
	hook := func(key string, content []byte) ([]byte, error) {
		return nil, hookErr
	}
	r.AddPreLoadHook(hook)

	hydrator := func(content []byte) (interface{}, error) {
		return string(content), nil
	}

	err := r.SetWithHooks("test", []byte("value"), hydrator)
	if err != hookErr {
		t.Errorf("expected hook error, got: %v", err)
	}

	// Value should not be stored on error
	_, err = r.Get("test")
	if err != ErrNotFound {
		t.Error("value should not be stored when hook fails")
	}
}

// TestRegistry_PostLoadHooks verifies post-load hook execution.
func TestRegistry_PostLoadHooks(t *testing.T) {
	r := New()

	var hookCalled bool
	hook := func(key string, value interface{}) error {
		hookCalled = true
		// Validate value
		if value != "hydrated" {
			return errors.New("invalid value")
		}
		return nil
	}
	r.AddPostLoadHook(hook)

	hydrator := func(content []byte) (interface{}, error) {
		return "hydrated", nil
	}

	err := r.SetWithHooks("test", []byte("raw"), hydrator)
	if err != nil {
		t.Fatalf("SetWithHooks failed: %v", err)
	}

	if !hookCalled {
		t.Error("post-load hook was not called")
	}
}

// TestRegistry_PostLoadHookMultiple verifies multiple post-load hooks execute.
func TestRegistry_PostLoadHookMultiple(t *testing.T) {
	r := New()

	var hookOrder []string

	hook1 := func(key string, value interface{}) error {
		hookOrder = append(hookOrder, "hook1")
		return nil
	}
	hook2 := func(key string, value interface{}) error {
		hookOrder = append(hookOrder, "hook2")
		return nil
	}

	r.AddPostLoadHook(hook1)
	r.AddPostLoadHook(hook2)

	hydrator := func(content []byte) (interface{}, error) {
		return "value", nil
	}

	err := r.SetWithHooks("test", []byte("raw"), hydrator)
	if err != nil {
		t.Fatalf("SetWithHooks failed: %v", err)
	}

	if len(hookOrder) != 2 {
		t.Fatalf("expected 2 hooks to run, got %d", len(hookOrder))
	}
	if hookOrder[0] != "hook1" || hookOrder[1] != "hook2" {
		t.Errorf("hooks executed in wrong order: %v", hookOrder)
	}
}
