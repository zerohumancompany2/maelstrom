package datasource

import (
	"github.com/maelstrom/v3/pkg/security"
	"testing"
)

func TestDataSource_Register(t *testing.T) {
	registry := NewRegistry()

	registry.Register("testSource", func(config map[string]any) (DataSource, error) {
		return &localDisk{path: "/test"}, nil
	})

	names := registry.List()
	if len(names) != 1 {
		t.Errorf("Expected 1 source, got %d", len(names))
	}

	source, err := registry.Get("testSource", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if source == nil {
		t.Error("Expected non-nil source")
	}
}

func TestDataSource_LocalDisk(t *testing.T) {
	registry := NewRegistry()

	registry.Register("localDisk", func(config map[string]any) (DataSource, error) {
		return &localDisk{path: "/tmp/test-ds"}, nil
	})

	source, err := registry.Get("localDisk", map[string]any{"path": "/tmp/test-ds"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if source == nil {
		t.Error("Expected non-nil source")
	}

	err = source.TagOnWrite("/tmp/test-ds/file.txt", []string{"TEST"})
	if err != nil {
		t.Errorf("Expected no error on tag, got %v", err)
	}
}

func TestDataSource_GetTaints(t *testing.T) {
	registry := NewRegistry()

	registry.Register("localDisk", func(config map[string]any) (DataSource, error) {
		return &localDisk{path: "/tmp/test-ds"}, nil
	})

	source, _ := registry.Get("localDisk", map[string]any{"path": "/tmp/test-ds"})

	taints, err := source.GetTaints("/tmp/test-ds/file2.txt")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(taints) != 0 {
		t.Errorf("Expected empty taints, got %v", taints)
	}
}

func TestDataSource_ValidateAccess(t *testing.T) {
	registry := NewRegistry()

	registry.Register("localDisk", func(config map[string]any) (DataSource, error) {
		return &localDisk{path: "/tmp/test-ds"}, nil
	})

	source, _ := registry.Get("localDisk", map[string]any{"path": "/tmp/test-ds"})

	err := source.ValidateAccess(security.InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for inner, got %v", err)
	}

	err = source.ValidateAccess(security.DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error for dmz, got %v", err)
	}

	err = source.ValidateAccess(security.OuterBoundary)
	if err != nil {
		t.Errorf("Expected no error for outer, got %v", err)
	}
}
