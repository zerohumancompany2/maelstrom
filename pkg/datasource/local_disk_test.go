package datasource

import (
	"encoding/json"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestLocalDisk_TagOnWrite(t *testing.T) {
	tmpDir := t.TempDir()
	ds, err := NewLocalDisk(map[string]any{
		"path": tmpDir,
	})
	if err != nil {
		t.Fatalf("NewLocalDisk failed: %v", err)
	}

	testFile := tmpDir + "/test.txt"
	taints := []string{"PII", "TOOL_OUTPUT"}

	err = ds.TagOnWrite(testFile, taints)
	if err != nil {
		t.Fatalf("TagOnWrite failed: %v", err)
	}

	dest := make([]byte, 256)
	n, err := unix.Lgetxattr(testFile, "user.maelstrom.taints", dest)
	if err != nil {
		t.Fatalf("Lgetxattr failed: %v", err)
	}
	value := dest[:n]

	expected := `["PII","TOOL_OUTPUT"]`
	if string(value) != expected {
		t.Errorf("xattr value mismatch: got %q, want %q", string(value), expected)
	}
}

func TestLocalDisk_GetTaints(t *testing.T) {
	tmpDir := t.TempDir()
	ds, err := NewLocalDisk(map[string]any{
		"path": tmpDir,
	})
	if err != nil {
		t.Fatalf("NewLocalDisk failed: %v", err)
	}

	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	expectedTaints := []string{"PII", "SECRET"}
	jsonData, _ := json.Marshal(expectedTaints)
	if err := unix.Lsetxattr(testFile, "user.maelstrom.taints", jsonData, 0); err != nil {
		t.Fatalf("Lsetxattr failed: %v", err)
	}

	taints, err := ds.GetTaints(testFile)
	if err != nil {
		t.Fatalf("GetTaints failed: %v", err)
	}

	if len(taints) != len(expectedTaints) {
		t.Errorf("taint length mismatch: got %d, want %d", len(taints), len(expectedTaints))
	}

	for i, expected := range expectedTaints {
		if taints[i] != expected {
			t.Errorf("taint[%d] mismatch: got %q, want %q", i, taints[i], expected)
		}
	}
}

func TestLocalDisk_SidecarFallback(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := tmpDir + "/test.txt"
	taints := []string{"PII", "TOOL_OUTPUT"}

	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	jsonData, _ := json.Marshal(taints)
	sidecarPath := testFile + ".maelstrom"
	if err := os.WriteFile(sidecarPath, jsonData, 0644); err != nil {
		t.Fatalf("WriteFile sidecar failed: %v", err)
	}

	ds, err := NewLocalDisk(map[string]any{
		"path": tmpDir,
	})
	if err != nil {
		t.Fatalf("NewLocalDisk failed: %v", err)
	}

	retrieved, err := ds.GetTaints(testFile)
	if err != nil {
		t.Fatalf("GetTaints failed: %v", err)
	}

	if len(retrieved) != len(taints) {
		t.Errorf("taint length mismatch: got %d, want %d", len(retrieved), len(taints))
	}

	for i, expected := range taints {
		if retrieved[i] != expected {
			t.Errorf("taint[%d] mismatch: got %q, want %q", i, retrieved[i], expected)
		}
	}
}
