package source

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileSystemSource watches a directory for YAML file changes.
type FileSystemSource struct {
	root     string
	debounce time.Duration
	events   chan SourceEvent
	err      error
	done     chan struct{}
	watcher  *fsnotify.Watcher
}

// NewFileSystemSource creates a new file system source.
func NewFileSystemSource(root string, debounce time.Duration) (*FileSystemSource, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileSystemSource{
		root:     root,
		debounce: debounce,
		events:   make(chan SourceEvent, 10),
		done:     make(chan struct{}),
		watcher:  watcher,
	}, nil
}

// Run starts watching the directory (blocking).
func (s *FileSystemSource) Run() error {
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

func (s *FileSystemSource) handleEvent(evt fsnotify.Event) {
	// Only handle YAML files
	if !strings.HasSuffix(evt.Name, ".yaml") {
		return
	}

	key := filepath.Base(evt.Name)

	// Handle remove events (deletes)
	if evt.Op&fsnotify.Remove == fsnotify.Remove || evt.Op&fsnotify.Rename == fsnotify.Rename {
		_, err := os.Stat(evt.Name)
		if os.IsNotExist(err) {
			s.events <- SourceEvent{
				Key:       key,
				Content:   nil,
				Type:      Deleted,
				Timestamp: time.Now(),
			}
			return
		}
	}

	// Handle create events
	if evt.Op&fsnotify.Create == fsnotify.Create {
		content, err := os.ReadFile(evt.Name)
		if err != nil {
			return // Skip files we can't read
		}

		s.events <- SourceEvent{
			Key:       key,
			Content:   content,
			Type:      Created,
			Timestamp: time.Now(),
		}
		return
	}

	// Handle write events (updates)
	if evt.Op&fsnotify.Write == fsnotify.Write {
		content, err := os.ReadFile(evt.Name)
		if err != nil {
			return
		}

		s.events <- SourceEvent{
			Key:       key,
			Content:   content,
			Type:      Updated,
			Timestamp: time.Now(),
		}
	}
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
