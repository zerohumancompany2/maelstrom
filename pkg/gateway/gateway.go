package gateway

import (
	"fmt"
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
)

type Adapter interface {
	Name() string
	NormalizeInbound(data []byte) (mail.Mail, error)
	NormalizeOutbound(mail mail.Mail) ([]byte, error)
}

type Gateway struct {
	adapters map[string]Adapter
	mu       sync.RWMutex
}

func NewGateway() *Gateway {
	return &Gateway{
		adapters: make(map[string]Adapter),
	}
}

func (g *Gateway) RegisterAdapter(adapter Adapter) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.adapters[adapter.Name()] = adapter
	return nil
}

func (g *Gateway) GetAdapter(name string) (Adapter, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	adapter, exists := g.adapters[name]
	if !exists {
		return nil, fmt.Errorf("adapter not found: %s", name)
	}
	return adapter, nil
}

func (g *Gateway) ListAdapters() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	names := make([]string, 0, len(g.adapters))
	for name := range g.adapters {
		names = append(names, name)
	}
	return names
}
