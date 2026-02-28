package registry

import (
	"context"

	"github.com/maelstrom/v3/pkg/source"
)

// Service orchestrates a Source and Registry, processing events and notifying observers.
type Service struct {
	src       source.Source
	registry  *Registry
	observers []func(key string, value interface{})
	hydrator  func([]byte) (interface{}, error)
}

// NewService creates a new Service with the given source and registry.
func NewService(src source.Source, reg *Registry) *Service {
	return &Service{
		src:      src,
		registry: reg,
		hydrator: func(b []byte) (interface{}, error) { return string(b), nil }, // default
	}
}

// SetHydrator sets the hydrator function (for testing).
func (s *Service) SetHydrator(fn func([]byte) (interface{}, error)) {
	s.hydrator = fn
}

// OnChange registers a callback for registry updates.
func (s *Service) OnChange(fn func(key string, value interface{})) {
	s.observers = append(s.observers, fn)
}

// Run processes source events until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt, ok := <-s.src.Events():
			if !ok {
				return s.src.Err()
			}
			s.processEvent(evt)
		}
	}
}

func (s *Service) processEvent(evt source.SourceEvent) {
	if evt.Type == source.Deleted {
		// For deletes, we could remove from registry or mark as deleted
		// For now, just store nil to indicate deletion
		s.registry.Set(evt.Key, nil)
		s.notify(evt.Key, nil)
		return
	}

	// Created or Updated: hydrate and store
	value, err := s.hydrator(evt.Content)
	if err != nil {
		// In production, log error and continue
		return
	}

	s.registry.Set(evt.Key, value)
	s.notify(evt.Key, value)
}

func (s *Service) notify(key string, value interface{}) {
	for _, fn := range s.observers {
		fn(key, value)
	}
}
