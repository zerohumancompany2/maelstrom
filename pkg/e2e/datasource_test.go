package e2e

import (
	"testing"

	"github.com/maelstrom/v3/pkg/datasource"
	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
)

func TestE2E_DataSource_WriteReadTaintCycle(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	agent := runtime.CreateAgent("testagent", mail.DMZBoundary, security.TaintPolicy{})
	if agent == nil {
		t.Fatal("Failed to create test agent")
	}

	ds := datasource.NewInMemoryDataSource()
	runtime.RegisterDataSource("inmemory", ds)

	filePath := "/agents/testagent/workspace/sensitive_data.txt"
	taintsToWrite := []string{"PII", "TOOL_OUTPUT"}

	err := runtime.WriteFileViaDataSourceWithDS(ds, filePath, taintsToWrite)
	if err != nil {
		t.Fatalf("Failed to write file via DataSource: %v", err)
	}

	readTaints, err := ds.GetTaints(filePath)
	if err != nil {
		t.Fatalf("Failed to get taints from DataSource: %v", err)
	}

	if len(readTaints) != 2 {
		t.Errorf("Expected 2 taints on file, got %d", len(readTaints))
	}

	hasPII := false
	hasToolOutput := false
	for _, t := range readTaints {
		if t == "PII" {
			hasPII = true
		}
		if t == "TOOL_OUTPUT" {
			hasToolOutput = true
		}
	}
	if !hasPII {
		t.Error("Expected PII taint on file")
	}
	if !hasToolOutput {
		t.Error("Expected TOOL_OUTPUT taint on file")
	}

	toolResultMessage, err := runtime.ReadFileViaDataSource(agent.ID, filePath)
	if err != nil {
		t.Fatalf("Failed to read file via DataSource: %v", err)
	}

	if toolResultMessage == nil {
		t.Fatal("Tool result message is nil")
	}

	if toolResultMessage.Type != mail.MailTypeToolResult {
		t.Errorf("Expected tool_result message type, got %s", toolResultMessage.Type)
	}

	hasPIIInMessage := false
	hasToolOutputInMessage := false
	for _, t := range toolResultMessage.Metadata.Taints {
		if t == "PII" {
			hasPIIInMessage = true
		}
		if t == "TOOL_OUTPUT" {
			hasToolOutputInMessage = true
		}
	}

	if !hasPIIInMessage {
		t.Error("Expected PII taint in tool_result Message")
	}
	if !hasToolOutputInMessage {
		t.Error("Expected TOOL_OUTPUT taint in tool_result Message")
	}

	if len(toolResultMessage.Metadata.Taints) != 2 {
		t.Errorf("Expected 2 taints in Message.Metadata.Taints (PII and TOOL_OUTPUT union), got %d: %v", len(toolResultMessage.Metadata.Taints), toolResultMessage.Metadata.Taints)
	}
}
