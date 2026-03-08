package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGateway_WebhookIngress(t *testing.T) {
	// Test: HTTP POST → mail_received event
	svc := NewGatewayService()
	webhook := &WebhookAdapter{}
	svc.RegisterAdapter("webhook", webhook)

	// Create a webhook endpoint handler
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

	// Create test request
	reqBody := GatewayMail{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test message",
		Body:    "Hello, world!",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
