package datasource

import (
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
