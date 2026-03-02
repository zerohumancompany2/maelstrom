package tools

import (
	"testing"
)

func TestTools_NotFound(t *testing.T) {
	svc := NewToolsService()

	_, err := svc.Resolve("nonexistent-tool", "inner")
	if err != nil {
		t.Fatalf("Resolve for nonexistent tool failed: %v", err)
	}

	_, err = svc.Invoke("nonexistent-tool", map[string]any{}, "inner")
	if err != nil {
		t.Fatalf("Invoke for nonexistent tool failed: %v", err)
	}

	err = svc.Unregister("nonexistent-tool")
	if err != nil {
		t.Fatalf("Unregister nonexistent tool failed: %v", err)
	}
}
