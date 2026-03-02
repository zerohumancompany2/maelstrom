package tools

import (
	"testing"
)

func TestTools_Unregister(t *testing.T) {
	svc := NewToolsService()

	svc.Register(ToolDescriptor{Name: "unregister-test", Boundary: "inner", Isolation: "process"})

	_, err := svc.Resolve("unregister-test", "inner")
	if err != nil {
		t.Fatalf("Resolve before unregister failed: %v", err)
	}

	err = svc.Unregister("unregister-test")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	_, err = svc.Resolve("unregister-test", "inner")
	if err != nil {
		t.Fatalf("Resolve after unregister failed: %v", err)
	}
}
