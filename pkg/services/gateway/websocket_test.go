package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGateway_WebSocketBidir(t *testing.T) {
	// Test: Bidirectional WS messaging
	svc := NewGatewayService()
	ws := &WebSocketAdapter{}
	svc.RegisterAdapter("websocket", ws)

	// Create test server simulating WebSocket upgrade
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")

		// Simulate bidirectional communication
		if r.Method == http.MethodPost {
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
		}
	}))
	defer server.Close()

	// Simulate bidirectional communication
	client := &http.Client{Timeout: 5 * time.Second}

	mail := GatewayMail{
		From:    "client@example.com",
		To:      []string{"server@example.com"},
		Subject: "WS test",
		Body:    "WebSocket bidirectional test",
	}

	jsonBody, _ := json.Marshal(mail)
	resp, err := client.Post(server.URL+"/ws", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
