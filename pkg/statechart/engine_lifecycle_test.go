package statechart

import (
	"errors"
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Core Lifecycle Tests
// ============================================================================

func TestSpawn_CreatesRuntimeWithUniqueID(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id1, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if id1 == "" {
		t.Error("Spawn returned empty RuntimeID")
	}

	// Spawn second runtime - should have different ID
	id2, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Second Spawn failed: %v", err)
	}

	if id1 == id2 {
		t.Error("Spawn returned duplicate RuntimeID")
	}
}

func TestControl_StartTransitionsToRunning(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Before Start, runtime is in Created state
	// After Start, runtime should be Running
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify by attempting to dispatch an event (only works when running)
	err = engine.Dispatch(id, Event{Type: "test"})
	if err != nil {
		t.Errorf("Dispatch failed after Start: %v", err)
	}
}

func TestControl_PauseAndResume(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Start the runtime
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Pause
	err = engine.Control(id, CmdPause)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Dispatch should fail while paused
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch should fail when runtime is paused")
	}

	// Resume
	err = engine.Control(id, CmdResume)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	// Dispatch should work after resume
	err = engine.Dispatch(id, Event{Type: "test"})
	if err != nil {
		t.Errorf("Dispatch failed after Resume: %v", err)
	}
}

func TestControl_StopCleansUpRuntime(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Start first
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop
	err = engine.Control(id, CmdStop)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// After stop, operations should fail
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch should fail after Stop")
	}

	err = engine.Control(id, CmdStart)
	if err == nil {
		t.Error("Start should fail after Stop")
	}
}

func TestControl_InvalidCommand(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, "invalid-cmd")
	if err == nil {
		t.Error("Invalid control command should fail")
	}

	if !errors.Is(err, ErrInvalidControlCmd) {
		t.Errorf("Expected ErrInvalidControlCmd, got: %v", err)
	}
}

func TestControl_StartFromWrongState(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Start from Created should work
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start from Created failed: %v", err)
	}

	// Start from Running should fail
	err = engine.Control(id, CmdStart)
	if err == nil {
		t.Error("Start from Running should fail")
	}

	if !errors.Is(err, ErrInvalidState) {
		t.Errorf("Expected ErrInvalidState, got: %v", err)
	}
}

// ============================================================================
// Implied Lifecycle Behavior Tests
// ============================================================================

func TestImplied_LifecycleStateMachine(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Created -> Start -> Running
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Running -> Pause -> Paused
	err = engine.Control(id, CmdPause)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Paused -> Resume -> Running
	err = engine.Control(id, CmdResume)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	// Running -> Stop -> Stopped
	err = engine.Control(id, CmdStop)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify stopped by trying to dispatch
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch should fail after Stop")
	}
}

func TestImplied_IDUniqueness(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()

	ids := make(map[RuntimeID]bool)

	for i := 0; i < 100; i++ {
		mockCtx := testutil.NewMockApplicationContext()
		id, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn %d failed: %v", i, err)
		}

		if ids[id] {
			t.Fatalf("Duplicate ID generated: %s", id)
		}

		ids[id] = true
	}
}

func TestImplied_StoppedRuntimeCleanup(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = engine.Control(id, CmdStop)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// After stop, the runtime should be cleaned up
	// Operations on the ID should fail
	err = engine.Control(id, CmdStart)
	if err == nil {
		t.Error("Control after Stop should fail")
	}

	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch after Stop should fail")
	}

	_, err = engine.Snapshot(id)
	if err == nil {
		t.Error("Snapshot after Stop should fail")
	}
}
