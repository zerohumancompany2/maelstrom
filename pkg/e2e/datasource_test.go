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

func TestE2E_DataSource_ContextMapBoundaryEnforcement(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	dmzAgent := runtime.CreateAgent("dmz-agent", mail.DMZBoundary, security.TaintPolicy{})
	if dmzAgent == nil {
		t.Fatal("Failed to create DMZ agent")
	}

	ds := datasource.NewInMemoryDataSource()
	runtime.RegisterDataSource("inmemory", ds)

	filePath := "/agents/dmz-agent/workspace/restricted_data.txt"
	forbiddenTaints := []string{"INNER_ONLY"}

	err := runtime.WriteFileViaDataSourceWithDS(ds, filePath, forbiddenTaints)
	if err != nil {
		t.Fatalf("Failed to write file via DataSource: %v", err)
	}

	readTaints, err := ds.GetTaints(filePath)
	if err != nil {
		t.Fatalf("Failed to get taints from DataSource: %v", err)
	}

	hasInnerOnly := false
	for _, t := range readTaints {
		if t == "INNER_ONLY" {
			hasInnerOnly = true
			break
		}
	}
	if !hasInnerOnly {
		t.Error("Expected INNER_ONLY taint on file after write")
	}

	toolResultMessage, err := runtime.ReadFileViaDataSource(dmzAgent.ID, filePath)
	if err != nil {
		t.Fatalf("Failed to read file via DataSource: %v", err)
	}

	if toolResultMessage == nil {
		t.Fatal("Tool result message is nil")
	}

	hasInnerOnlyInMessage := false
	for _, t := range toolResultMessage.Metadata.Taints {
		if t == "INNER_ONLY" {
			hasInnerOnlyInMessage = true
			break
		}
	}
	if !hasInnerOnlyInMessage {
		t.Error("Expected INNER_ONLY taint attached to ContextBlock Message")
	}

	contextBlock := &security.ContextBlock{
		Name:    "workspace-file",
		Source:  "memoryService",
		Content: "restricted content from file",
		TaintPolicy: security.TaintPolicy{
			RedactMode:         "dropBlock",
			AllowedForBoundary: []security.BoundaryType{security.InnerBoundary},
		},
		Taints: security.TaintSet{"INNER_ONLY": true},
	}

	runtime.RegisterContextBlock(dmzAgent.ID, contextBlock)

	err = runtime.PrepareContextForBoundary(dmzAgent.ID, mail.DMZBoundary)
	if err != nil {
		t.Logf("prepareContextForBoundary returned error (expected for forbidden taint): %v", err)
	}

	_, _ = runtime.GetContextMap(dmzAgent.ID)

	blockedBlocks := runtime.GetBlockedContextBlocks(dmzAgent.ID)
	if len(blockedBlocks) == 0 {
		t.Error("Expected ContextBlock to be dropped per dropBlock policy for INNER_ONLY taint")
	}

	violations := runtime.GetViolations()
	violationFound := false
	for _, v := range violations {
		if v.Type == mail.MailTypeTaintViolation {
			violationFound = true
			if content, ok := v.Content.(map[string]interface{}); ok {
				if taints, ok := content["taints"].([]string); ok {
					for _, t := range taints {
						if t == "INNER_ONLY" {
							break
						}
					}
				}
			}
		}
	}
	if !violationFound {
		t.Error("Expected violation logged to sys:observability for INNER_ONLY taint at DMZ boundary")
	}

	assembledPrompt := runtime.AssemblePromptForLLM(dmzAgent.ID)
	if assembledPrompt == "" {
		t.Log("Prompt is empty as expected when all blocks are dropped")
	} else {
		if containsString(assembledPrompt, "restricted content") {
			t.Error("INNER_ONLY data should not be in assembled prompt")
		}
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
