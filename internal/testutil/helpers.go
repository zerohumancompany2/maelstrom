package testutil

import (
	"context"
	"github.com/maelstrom/v3/pkg/mail"
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
func WaitForChannel(t *testing.T, ch <-chan mail.Mail, timeout time.Duration) (mail.Mail, bool) {
	t.Helper()

	select {
	case m := <-ch:
		return m, true
	case <-time.After(timeout):
		return mail.Mail{}, false
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
func MustReceiveMail(t *testing.T, ch <-chan mail.Mail, timeout time.Duration) mail.Mail {
	t.Helper()

	m, ok := WaitForChannel(t, ch, timeout)
	if !ok {
		t.Fatalf("timeout waiting for mail")
	}
	return m
}
