package datasource

import (
	"testing"
)

func TestInMemory_TagOnWrite(t *testing.T) {
	ds := NewInMemoryDataSource()

	path := "/workspace/test.txt"
	taints := []string{"WORKSPACE", "TOOL_OUTPUT"}

	err := ds.TagOnWrite(path, taints)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	im, ok := ds.(*inMemoryDataSource)
	if !ok {
		t.Fatal("Failed to cast to inMemoryDataSource")
	}

	stored, ok := im.taints[path]
	if !ok {
		t.Fatal("Expected taints to be stored for path")
	}

	if len(stored) != len(taints) {
		t.Errorf("Expected %d taints, got %d", len(taints), len(stored))
	}

	for i, expected := range taints {
		if stored[i] != expected {
			t.Errorf("Expected taint %s at index %d, got %s", expected, i, stored[i])
		}
	}
}

func TestInMemory_GetTaints(t *testing.T) {
	ds := NewInMemoryDataSource()

	knownPath := "/workspace/test.txt"
	unknownPath := "/workspace/other.txt"
	taints := []string{"WORKSPACE", "TOOL_OUTPUT"}

	err := ds.TagOnWrite(knownPath, taints)
	if err != nil {
		t.Fatalf("Expected no error on TagOnWrite, got %v", err)
	}

	retrieved, err := ds.GetTaints(knownPath)
	if err != nil {
		t.Fatalf("Expected no error on GetTaints, got %v", err)
	}

	if len(retrieved) != len(taints) {
		t.Errorf("Expected %d taints for known path, got %d", len(taints), len(retrieved))
	}

	for i, expected := range taints {
		if retrieved[i] != expected {
			t.Errorf("Expected taint %s at index %d, got %s", expected, i, retrieved[i])
		}
	}

	emptyTaints, err := ds.GetTaints(unknownPath)
	if err != nil {
		t.Fatalf("Expected no error for unknown path, got %v", err)
	}

	if len(emptyTaints) != 0 {
		t.Errorf("Expected empty taints for unknown path, got %v", emptyTaints)
	}
}
