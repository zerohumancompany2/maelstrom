package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGateway_SSEStreaming(t *testing.T) {
	// Test: SSE chunk streaming
	svc := NewGatewayService()
	sse := &SSEAdapter{}
	svc.RegisterAdapter("sse", sse)

	// Create SSE endpoint handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Stream multiple chunks
		for i := 0; i < 3; i++ {
			mail := Mail{
				From:    "stream@example.com",
				To:      []string{"client@example.com"},
				Subject: fmt.Sprintf("Chunk %d", i),
				Body:    fmt.Sprintf("Message chunk %d", i),
			}

			data, _ := json.Marshal(mail)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		}
	})

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Connect client
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
