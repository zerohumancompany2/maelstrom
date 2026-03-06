package testutil

import (
	"context"
	"testing"
	"time"
)

// WaitForCondition waits for a condition to become true with timeout.
// Uses polling with ticker instead of time.Sleep for deterministic testing.
func WaitForCondition(t *testing.T, cond func() bool, timeout time.Duration) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if cond() {
					done <- true
					return
				}
			}
		}
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

// WaitForChannel waits for a channel to receive a value with timeout.
func WaitForChannel(t *testing.T, ch <-chan Mail, timeout time.Duration) (Mail, bool) {
	t.Helper()

	select {
	case mail := <-ch:
		return mail, true
	case <-time.After(timeout):
		return Mail{}, false
	}
}

// WaitForGoroutines waits for all goroutines to complete.
func WaitForGoroutines(t *testing.T, doneChans []chan struct{}, timeout time.Duration) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, ch := range doneChans {
		select {
		case <-ch:
		case <-ctx.Done():
			return false
		}
	}
	return true
}

// MustReceiveMail receives mail from channel or fails test.
func MustReceiveMail(t *testing.T, ch <-chan Mail, timeout time.Duration) Mail {
	t.Helper()

	mail, ok := WaitForChannel(t, ch, timeout)
	if !ok {
		t.Fatalf("timeout waiting for mail")
	}
	return mail
}
