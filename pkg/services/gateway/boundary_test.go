package gateway

import (
	"testing"
)

func TestGateway_BoundaryOuter(t *testing.T) {
	// Test: Gateway only accepts outer boundary mail
	svc := NewGatewayService()

	// Register an adapter
	svc.RegisterAdapter("webhook", &WebhookAdapter{})

	// Simulate outer boundary mail (from external source)
	outerMail := GatewayMail{
		From:    "external@example.com",
		To:      []string{"gateway@maelstrom.io"},
		Subject: "Outer boundary message",
		Body:    "Message from outer layer",
	}

	// Publish should succeed for outer mail
	ack, err := svc.Publish(outerMail)
	if err != nil {
		t.Fatalf("Failed to publish outer mail: %v", err)
	}

	if ack.Status != "published" {
		t.Errorf("Expected status 'published', got '%s'", ack.Status)
	}

	// Publish mail and verify it's received
	err = svc.PublishTo(outerMail)
	if err != nil {
		t.Fatalf("Failed to publish to channel: %v", err)
	}

	// Subscribe and receive mail
	ch, err := svc.Subscribe("test-channel")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	received := <-ch
	if received.Subject != outerMail.Subject {
		t.Errorf("Expected subject '%s', got '%s'", outerMail.Subject, received.Subject)
	}
}
