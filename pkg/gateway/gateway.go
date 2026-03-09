package gateway

import (
	"fmt"
	"net/http"
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

type GatewayService struct {
	endpoints map[string]http.Handler
	mu        sync.RWMutex
}

func NewGatewayService() *GatewayService {
	return &GatewayService{
		endpoints: make(map[string]http.Handler),
	}
}

func (g *GatewayService) RegisterHTTPEndpoint(path string, handler http.Handler) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.endpoints[path] = handler
	return nil
}

func (g *GatewayService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, handler := range g.endpoints {
		handler.ServeHTTP(w, r)
	}
}

func (g *GatewayService) GetOpenAPISpec() *OpenAPISpec {
	g.mu.RLock()
	defer g.mu.RUnlock()

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:   "Gateway API",
			Version: "1.0.0",
		},
		Paths: make(map[string]interface{}),
	}

	for path := range g.endpoints {
		spec.Paths[path] = map[string]interface{}{
			"get": map[string]interface{}{
				"summary": "Endpoint at " + path,
			},
		}
	}

	return spec
}

func (g *GatewayService) checkBoundaryExposure(boundary mail.BoundaryType) bool {
	switch boundary {
	case mail.InnerBoundary:
		return false
	case mail.DMZBoundary, mail.OuterBoundary:
		return true
	default:
		return false
	}
}
