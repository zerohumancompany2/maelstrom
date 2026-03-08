package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGateway_SMTPIngress(t *testing.T) {
	// Test: Email → mail_received
	svc := NewGatewayService()
	smtp := &SMTPAdapter{}
	svc.RegisterAdapter("smtp", smtp)

	// Create SMTP ingress endpoint handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var mail GatewayMail
		if err := json.NewDecoder(r.Body).Decode(&mail); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ack, err := svc.Publish(mail)
		if err != nil {
			http.Error(w, "Failed to publish", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ack)
	})

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Simulate email ingress with POST request containing mail data
	mail := GatewayMail{
		From:    "sender@external.com",
		To:      []string{"agent@internal.com"},
		Subject: "Email from external",
		Body:    "This is an email message",
	}

	jsonBody, _ := json.Marshal(mail)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
