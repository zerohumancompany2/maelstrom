package lifecycle

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestLifecycle_RuntimeInfoHasRequiredFields(t *testing.T) {
	info := RuntimeInfo{
		ID:           "test",
		DefinitionID: "def1",
		Boundary:     mail.InnerBoundary,
		ActiveStates: []string{"state1"},
		IsRunning:    true,
	}

	if info.ID != "test" {
		t.Errorf("expected ID to be 'test', got '%s'", info.ID)
	}
	if info.DefinitionID != "def1" {
		t.Errorf("expected DefinitionID to be 'def1', got '%s'", info.DefinitionID)
	}
	if info.Boundary != mail.InnerBoundary {
		t.Errorf("expected Boundary to be InnerBoundary, got '%s'", info.Boundary)
	}
	if len(info.ActiveStates) != 1 || info.ActiveStates[0] != "state1" {
		t.Errorf("expected ActiveStates to be ['state1'], got %v", info.ActiveStates)
	}
	if !info.IsRunning {
		t.Error("expected IsRunning to be true")
	}
}
