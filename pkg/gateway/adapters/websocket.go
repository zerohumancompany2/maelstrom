package adapters

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/maelstrom/v3/pkg/mail"
)

type WebSocketAdapter struct {
	name     string
	server   *http.Server
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
}

func NewWebSocketAdapter() *WebSocketAdapter {
	return &WebSocketAdapter{name: "websocket"}
}

func (a *WebSocketAdapter) Name() string {
	return a.name
}

func (a *WebSocketAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return mail.Mail{}, err
	}

	return mail.Mail{
		ID:      generateID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:websocket",
		Content: payload,
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *WebSocketAdapter) NormalizeOutbound(mailObj mail.Mail) ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    mailObj.Type,
		"id":      mailObj.ID,
		"content": mailObj.Content,
		"source":  mailObj.Source,
		"stream":  mailObj.Metadata.StreamChunk != nil,
	})
}

func (a *WebSocketAdapter) StartServer(addr string) error {
	a.mu.Lock()
	a.clients = make(map[*websocket.Conn]bool)
	a.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", a.handleWebSocket)

	a.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Server error
		}
	}()

	a.mu.Unlock()
	return nil
}

func (a *WebSocketAdapter) StopServer() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		a.server.Close()
	}

	for client := range a.clients {
		client.Close()
	}
}

func (a *WebSocketAdapter) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	a.mu.Lock()
	a.clients[conn] = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.clients, conn)
		a.mu.Unlock()
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		_, err = a.NormalizeInbound(msg)
		if err != nil {
			break
		}
	}
}
