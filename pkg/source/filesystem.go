package source

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileSystemSource watches a directory for YAML file changes.
type FileSystemSource struct {
	root      string
	debounce  time.Duration
	debouncers map[string]*time.Timer
	events    chan SourceEvent
	err       error
	done      chan struct{}
	watcher   *fsnotify.Watcher
	mu        sync.Mutex
	seen      map[string]bool // tracks files we've seen for update vs create detection
}

// NewFileSystemSource creates a new file system source.
func NewFileSystemSource(root string, debounce time.Duration) (*FileSystemSource, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileSystemSource{
		root:       root,
		debounce:   debounce,
		debouncers: make(map[string]*time.Timer),
		events:     make(chan SourceEvent, 10),
		done:       make(chan struct{}),
		watcher:    watcher,
		seen:       make(map[string]bool),
	}, nil
}

// Run starts watching the directory (blocking).
func (s *FileSystemSource) Run() error {
	// Do initial scan for existing files
	if err := s.initialScan(); err != nil {
		s.err = err
		close(s.events)
		return err
	}

	// Add the directory to watch
	if err := s.watcher.Add(s.root); err != nil {
		s.err = err
		close(s.events)
		return err
	}

	for {
		select {
		case <-s.done:
			close(s.events)
			return nil
		case evt, ok := <-s.watcher.Events:
			if !ok {
				close(s.events)
				return nil
			}
			s.handleEvent(evt)
		case err, ok := <-s.watcher.Errors:
			if !ok {
				close(s.events)
				return nil
			}
			s.err = err
			close(s.events)
			return err
		}
	}
}

func (s *FileSystemSource) initialScan() error {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}

		path := filepath.Join(s.root, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue // Skip files we can't read
		}

		s.seen[name] = true
		s.events <- SourceEvent{
			Key:       name,
			Content:   content,
			Type:      Created,
			Timestamp: time.Now(),
		}
	}
	return nil
}

func (s *FileSystemSource) handleEvent(evt fsnotify.Event) {
	// Only handle YAML files
	if !strings.HasSuffix(evt.Name, ".yaml") {
		return
	}

	key := filepath.Base(evt.Name)

	// Handle remove events (deletes)
	if evt.Op&fsnotify.Remove == fsnotify.Remove || evt.Op&fsnotify.Rename == fsnotify.Rename {
		s.debounceEvent(key, func() {
			_, err := os.Stat(evt.Name)
			if os.IsNotExist(err) {
				s.mu.Lock()
				delete(s.seen, key)
				s.mu.Unlock()

				s.events <- SourceEvent{
					Key:       key,
					Content:   nil,
					Type:      Deleted,
					Timestamp: time.Now(),
				}
			}
		})
		return
	}

	// Handle create or write events
	if evt.Op&fsnotify.Create == fsnotify.Create || evt.Op&fsnotify.Write == fsnotify.Write {
		s.debounceEvent(key, func() {
			content, err := os.ReadFile(evt.Name)
			if err != nil {
				return
			}

			s.mu.Lock()
			eventType := Created
			if s.seen[key] {
				eventType = Updated
			}
			s.seen[key] = true
			s.mu.Unlock()

			s.events <- SourceEvent{
				Key:       key,
				Content:   content,
				Type:      eventType,
				Timestamp: time.Now(),
			}
		})
	}
}

func (s *FileSystemSource) debounceEvent(key string, fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if timer, exists := s.debouncers[key]; exists {
		timer.Stop()
	}

	s.debouncers[key] = time.AfterFunc(s.debounce, func() {
		s.mu.Lock()
		delete(s.debouncers, key)
		s.mu.Unlock()
		fn()
	})
}

// Stop gracefully shuts down the watcher.
func (s *FileSystemSource) Stop() error {
	close(s.done)
	return s.watcher.Close()
}

// Events implements Source interface.
func (s *FileSystemSource) Events() <-chan SourceEvent {
	return s.events
}

// Err implements Source interface.
func (s *FileSystemSource) Err() error {
	return s.err
}
