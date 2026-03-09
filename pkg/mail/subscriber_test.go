package mail

import (
	"testing"
)

func TestSubscriber_Receives(t *testing.T) {
	sub := NewChannelSubscriber(10)

	ch := sub.Receive()
	if ch == nil {
		t.Error("Expected non-nil channel")
	}

	// Send mail through internal channel via GetMail verification
	mail := Mail{ID: "msg-001", Type: MailTypeUser}
	sub.ch <- mail

	// Receive mail
	received, ok := sub.GetMail()
	if !ok {
		t.Error("Expected to receive mail")
	}

	if received.ID != "msg-001" {
		t.Errorf("Expected msg-001, got %s", received.ID)
	}
}