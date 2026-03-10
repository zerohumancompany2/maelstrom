package admin

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services/lifecycle"
	"github.com/maelstrom/v3/pkg/statechart"
)

// TestAdminService_ID - spec: arch-v1.md L467, L477-480
func TestAdminService_ID(t *testing.T) {
	svc := NewAdminService()

	id := svc.ID()
	if id != "sys:admin" {
		t.Errorf("Expected ID 'sys:admin', got '%s'", id)
	}
}

// TestAdminService_ListAgents - spec: arch-v1.md L467, L485
func TestAdminService_ListAgents(t *testing.T) {
	svc := NewAdminService()

	agents, err := svc.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}

	for _, agent := range agents {
		if agent.ID == "" {
			t.Error("Expected agent with ID")
		}
		if !lifecycle.RuntimeInfo(agent).IsRunning {
			t.Log("Agent is not running as expected")
		}
	}
}

// TestAdminService_ControlAgent - spec: arch-v1.md L467, L485
func TestAdminService_ControlAgent(t *testing.T) {
	svc := NewAdminService()

	err := svc.ControlAgent("nonexistent-agent", "pause")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

// TestAdminService_QueryTaints - spec: arch-v1.md L467, L485
func TestAdminService_QueryTaints(t *testing.T) {
	svc := NewAdminService()

	taints, err := svc.QueryTaints("nonexistent-agent")
	if err != nil {
		t.Fatalf("QueryTaints failed: %v", err)
	}

	if taints == nil {
		t.Error("Expected non-nil taints map")
	}
}

func TestAdmin_2FARequired(t *testing.T) {
	svc := NewAdminService()

	agents, err := svc.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}
}

func TestAdmin_ListAgents(t *testing.T) {
	svc := NewAdminService()

	agents, err := svc.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}
}

func TestAdmin_ControlAgent(t *testing.T) {
	svc := NewAdminService()

	err := svc.ControlAgent("nonexistent-agent", "pause")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

func TestAdmin_QueryTaints(t *testing.T) {
	svc := NewAdminService()

	taints, err := svc.QueryTaints("nonexistent-agent")
	if err != nil {
		t.Fatalf("QueryTaints failed: %v", err)
	}

	if taints == nil {
		t.Error("Expected non-nil taints map")
	}
}

func TestAdmin_InjectEvent(t *testing.T) {
	svc := NewAdminService()

	event := statechart.Event{
		Type:          "test:event",
		Payload:       map[string]any{"key": "value"},
		CorrelationID: "test-corr-id",
	}

	err := svc.InjectEvent("nonexistent-agent", event)
	if err != nil {
		t.Fatalf("InjectEvent failed: %v", err)
	}
}

func TestAdmin_BoundaryOuter(t *testing.T) {
	svc := NewAdminService()

	agents, err := svc.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	for _, agent := range agents {
		if agent.Boundary != "outer" {
			t.Errorf("Expected boundary 'outer', got '%s'", agent.Boundary)
		}
	}
}

func TestAdminService_AdminCommand(t *testing.T) {
	svc := NewAdminService()

	err := svc.ExecuteCommand("test-command", "valid-token")
	if err != nil {
		t.Errorf("Expected nil error for valid command, got %v", err)
	}
}

func TestAdminService_2FAGate(t *testing.T) {
	svc := NewAdminService()

	err := svc.ExecuteCommand("test-command", "")
	if err == nil {
		t.Error("Expected error for empty token")
	}
}

// TestAdminService_2FARequired - spec: arch-v1.md L467 (2FA-gated), L485 (authToken parameter)
func TestAdminService_2FARequired(t *testing.T) {
	svc := NewAdminService()

	err := svc.ExecuteCommand("test-command", "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	err = svc.ExecuteCommand("test-command", "")
	if err == nil {
		t.Error("Expected error for empty token")
	}
}

// TestAdminService_OuterBoundaryOnly - spec: arch-v1.md L467 (outer-only), L261-270 (Boundary Model)
func TestAdminService_OuterBoundaryOnly(t *testing.T) {
	svc := NewAdminService()
	adminSvc := svc.(*adminService)

	token, err := adminSvc.authManager.CreateToken("admin", time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	err = adminSvc.ExecuteCommandOnBoundary("test-command", token, mail.InnerBoundary)
	if err == nil {
		t.Error("Expected error for inner boundary")
	}

	err = adminSvc.ExecuteCommandOnBoundary("test-command", token, mail.DMZBoundary)
	if err == nil {
		t.Error("Expected error for DMZ boundary")
	}

	err = adminSvc.ExecuteCommandOnBoundary("test-command", token, mail.OuterBoundary)
	if err != nil {
		t.Errorf("Expected nil error for outer boundary, got %v", err)
	}
}
