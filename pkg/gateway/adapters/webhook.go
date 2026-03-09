package adapters

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
)

type WebhookAdapter struct {
	name       string
	server     *http.Server
	listenAddr string
	mu         sync.RWMutex
}

func NewWebhookAdapter() *WebhookAdapter {
	return &WebhookAdapter{name: "webhook"}
}

func (a *WebhookAdapter) Name() string {
	return a.name
}

func (a *WebhookAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return mail.Mail{}, err
	}

	return mail.Mail{
		ID:      generateID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:webhook",
		Target:  "agent:default",
		Content: payload,
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *WebhookAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    mail.Type,
		"content": mail.Content,
		"source":  mail.Source,
	})
}

func (a *WebhookAdapter) StartServer(addr string) error {
	a.mu.Lock()
	a.listenAddr = addr

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", a.handleWebhook)

	a.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Webhook server error: %v\n", err)
		}
	}()

	a.mu.Unlock()
	return nil
}

func (a *WebhookAdapter) StopServer() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		a.server.Close()
	}
}

func (a *WebhookAdapter) Addr() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.listenAddr
}

func (a *WebhookAdapter) handleWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("webhook-%x", b)
}
