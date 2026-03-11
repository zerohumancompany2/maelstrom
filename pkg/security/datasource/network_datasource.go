package datasource

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/maelstrom/v3/pkg/security/taint"
)

type networkDataSource struct {
	mu          sync.RWMutex
	taintEngine *taint.TaintEngine
	cache       map[string]any
	taints      map[string][]string
	httpClient  *http.Client
}

func NewNetworkDataSource(cfg *Config) *networkDataSource {
	return &networkDataSource{
		taintEngine: cfg.TaintEngine,
		cache:       make(map[string]any),
		taints:      make(map[string][]string),
		httpClient:  &http.Client{},
	}
}

func (n *networkDataSource) Fetch(url string) (any, []string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if cached, ok := n.cache[url]; ok {
		storedTaints := n.taints[url]
		return cached, storedTaints, nil
	}

	resp, err := n.httpClient.Get(url)
	if err != nil {
		return nil, []string{string(taint.TaintExternal)}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, []string{string(taint.TaintExternal)}, err
	}

	var content any
	if err := json.Unmarshal(body, &content); err != nil {
		content = string(body)
	}

	n.cache[url] = content
	n.taints[url] = []string{string(taint.TaintExternal)}

	return content, []string{string(taint.TaintExternal)}, nil
}
