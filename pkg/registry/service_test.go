package registry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/source"
)

// TestService_ProcessesEvents verifies service processes source events into registry.
func TestService_ProcessesEvents(t *testing.T) {
	manualSrc := source.NewManualSource()
	reg := New()
	svc := NewService(manualSrc, reg)

	// Set up a simple hydrator that prepends "hydrated:"
	svc.SetHydrator(func(content []byte) (interface{}, error) {
		return "hydrated:" + string(content), nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	// Send a Created event
	manualSrc.Send(source.SourceEvent{
		Key:     "test.yaml",
		Content: []byte("test content"),
		Type:    source.Created,
	})

	// Give time to process
	time.Sleep(50 * time.Millisecond)

	// Event should be processed into registry
	val, err := reg.Get("test.yaml")
	if err != nil {
		t.Fatalf("value not stored in registry: %v", err)
	}
	if val != "hydrated:test content" {
		t.Errorf("expected 'hydrated:test content', got %v", val)
	}
}

// TestService_ObserverNotifications verifies observers receive updates.
func TestService_ObserverNotifications(t *testing.T) {
	manualSrc := source.NewManualSource()
	reg := New()
	svc := NewService(manualSrc, reg)

	var receivedKey string
	var receivedValue interface{}
	svc.OnChange(func(key string, value interface{}) {
		receivedKey = key
		receivedValue = value
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	manualSrc.Send(source.SourceEvent{
		Key:     "test.yaml",
		Content: []byte("content"),
		Type:    source.Created,
	})

	time.Sleep(50 * time.Millisecond)

	if receivedKey != "test.yaml" {
		t.Errorf("expected key 'test.yaml', got %q", receivedKey)
	}
	if receivedValue != "content" {
		t.Errorf("expected value 'content', got %v", receivedValue)
	}
}

// TestService_ContextCancellation verifies clean shutdown on context cancel.
func TestService_ContextCancellation(t *testing.T) {
	manualSrc := source.NewManualSource()
	reg := New()
	svc := NewService(manualSrc, reg)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for Run to return")
	}
}

// TestService_SourceErrorHandling verifies errors from source are handled.
func TestService_SourceErrorHandling(t *testing.T) {
	manualSrc := source.NewManualSource()
	reg := New()
	svc := NewService(manualSrc, reg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	// Close source with an error
	srcErr := errors.New("source error")
	manualSrc.Close(srcErr)

	select {
	case err := <-done:
		// Service should return the source error
		if err != srcErr {
			t.Errorf("expected source error %v, got %v", srcErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for Run to return")
	}
}

// TestService_MultipleEvents verifies sequential processing of events.
func TestService_MultipleEvents(t *testing.T) {
	manualSrc := source.NewManualSource()
	reg := New()
	svc := NewService(manualSrc, reg)

	var events []string
	svc.OnChange(func(key string, value interface{}) {
		events = append(events, key+":"+value.(string))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	// Send multiple events
	for i := 1; i <= 3; i++ {
		manualSrc.Send(source.SourceEvent{
			Key:     fmt.Sprintf("file%d.yaml", i),
			Content: []byte(fmt.Sprintf("content%d", i)),
			Type:    source.Created,
		})
	}

	time.Sleep(100 * time.Millisecond)

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d: %v", len(events), events)
	}

	expected := []string{"file1.yaml:content1", "file2.yaml:content2", "file3.yaml:content3"}
	for i, exp := range expected {
		if events[i] != exp {
			t.Errorf("event %d: expected %q, got %q", i, exp, events[i])
		}
	}
}
