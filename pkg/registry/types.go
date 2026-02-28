package registry

import (
	"errors"
	"sync"
	"time"
)

// Registry stores and retrieves values by key with version tracking.
// The stored values are interface{} but hydrators guarantee type consistency.
type Registry struct {
	mu           sync.RWMutex
	entries      map[string]*entry
	preLoadHooks  []PreLoadHook
	postLoadHooks []PostLoadHook
}

// entry stores the current value and all previous versions.
type entry struct {
	versions []Version
	current  int // index into versions slice
}

// Version represents a single version of a registry entry.
type Version struct {
	Data      interface{}
	Timestamp int64
}

// ErrNotFound is returned when a key doesn't exist in the registry.
var ErrNotFound = errors.New("key not found in registry")

// ErrVersionNotFound is returned when a specific version doesn't exist.
var ErrVersionNotFound = errors.New("version not found")

// PreLoadHook transforms raw YAML before hydration.
type PreLoadHook func(key string, content []byte) ([]byte, error)

// PostLoadHook validates/processes values after hydration.
type PostLoadHook func(key string, value interface{}) error

// New creates a new empty Registry.
func New() *Registry {
	return &Registry{
		entries: make(map[string]*entry),
	}
}

// Set stores a value in the registry as the current version.
func (r *Registry) Set(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, exists := r.entries[key]
	if !exists {
		e = &entry{}
		r.entries[key] = e
	}

	e.versions = append(e.versions, Version{
		Data:      value,
		Timestamp: time.Now().UnixNano(),
	})
	e.current = len(e.versions) - 1
}

// Get retrieves the current version of a key.
func (r *Registry) Get(key string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, exists := r.entries[key]
	if !exists {
		return nil, ErrNotFound
	}

	return e.versions[e.current].Data, nil
}

// GetVersion retrieves a specific version of a key.
func (r *Registry) GetVersion(key string, version int) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, exists := r.entries[key]
	if !exists {
		return nil, ErrNotFound
	}

	if version < 0 || version >= len(e.versions) {
		return nil, ErrVersionNotFound
	}

	return e.versions[version].Data, nil
}

// CloneUnderLock executes fn with a read-locked snapshot of the registry.
func (r *Registry) CloneUnderLock(fn func(map[string]interface{})) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make(map[string]interface{}, len(r.entries))
	for key, e := range r.entries {
		snapshot[key] = e.versions[e.current].Data
	}

	fn(snapshot)
}

// AddPreLoadHook registers a hook to transform raw YAML before hydration.
func (r *Registry) AddPreLoadHook(hook PreLoadHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preLoadHooks = append(r.preLoadHooks, hook)
}

// AddPostLoadHook registers a hook to validate/process values after hydration.
func (r *Registry) AddPostLoadHook(hook PostLoadHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.postLoadHooks = append(r.postLoadHooks, hook)
}

// SetWithHooks stores a value after applying all pre-load and post-load hooks.
func (r *Registry) SetWithHooks(key string, content []byte, hydrator func([]byte) (interface{}, error)) error {
	// Copy hooks under lock to avoid race conditions
	r.mu.RLock()
	preHooks := make([]PreLoadHook, len(r.preLoadHooks))
	copy(preHooks, r.preLoadHooks)
	postHooks := make([]PostLoadHook, len(r.postLoadHooks))
	copy(postHooks, r.postLoadHooks)
	r.mu.RUnlock()

	// Apply pre-load hooks
	for _, hook := range preHooks {
		var err error
		content, err = hook(key, content)
		if err != nil {
			return err
		}
	}

	// Hydrate
	value, err := hydrator(content)
	if err != nil {
		return err
	}

	// Apply post-load hooks
	for _, hook := range postHooks {
		if err := hook(key, value); err != nil {
			return err
		}
	}

	// Store
	r.Set(key, value)
	return nil
}
