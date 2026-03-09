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
