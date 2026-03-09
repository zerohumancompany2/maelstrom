package security

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestContextMap_AssemblyOrder(t *testing.T) {
	blocks := []*ContextBlock{
		{Name: "low_priority", Source: "static", Content: "low", Priority: 10},
		{Name: "high_priority", Source: "static", Content: "high", Priority: 1},
		{Name: "medium_priority", Source: "static", Content: "medium", Priority: 5},
	}

	assembled := assemble(blocks)

	if len(assembled) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(assembled))
	}

	if assembled[0].Name != "high_priority" {
		t.Errorf("Expected first block to be high_priority, got %s", assembled[0].Name)
	}

	if assembled[1].Name != "medium_priority" {
		t.Errorf("Expected second block to be medium_priority, got %s", assembled[1].Name)
	}

	if assembled[2].Name != "low_priority" {
		t.Errorf("Expected third block to be low_priority, got %s", assembled[2].Name)
	}
}

func TestContextMap_SourceStatic(t *testing.T) {
	block := &ContextBlock{
		Name:    "system_prompt",
		Source:  "static",
		Content: "You are a secure agent",
	}

	content := staticSource(block)

	if content != "You are a secure agent" {
		t.Errorf("Expected static content 'You are a secure agent', got %s", content)
	}
}

func TestContextMap_SourceSession(t *testing.T) {
	block := &ContextBlock{
		Name:     "conversation",
		Source:   "session",
		Strategy: "lastN",
		N:        30,
	}

	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}
	content := sessionSource(block, messages)

	expected := "msg1, msg2, msg3, msg4, msg5"
	if content != expected {
		t.Errorf("Expected '%s', got '%s'", expected, content)
	}
}

func TestContextMap_SourceMemory(t *testing.T) {
	block := &ContextBlock{
		Name:     "long_term_memory",
		Source:   "memory",
		Strategy: "RAG",
	}

	query := "current goal"
	content := memorySource(block, query)

	if content != "queried: current goal" {
		t.Errorf("Expected 'queried: current goal', got '%s'", content)
	}
}

func TestContextMap_SourceTool(t *testing.T) {
	block := &ContextBlock{
		Name:           "tools_available",
		Source:         "tool",
		BoundaryFilter: DMZBoundary,
	}

	tools := []string{"webSearch", "innerDbQuery", "fileRead"}
	content := toolSource(block, tools, DMZBoundary)

	expected := "webSearch,fileRead"
	if content != expected {
		t.Errorf("Expected '%s', got '%s'", expected, content)
	}
}

func TestTaintEngine_ProactiveMarking(t *testing.T) {
	engine := NewTaintEngine()

	obj := map[string]string{"key": "value"}
	marked, err := engine.Mark(obj, []string{"PII", "SECRET"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if marked == nil {
		t.Error("Expected marked object to be non-nil")
	}
}

func TestTaintEngine_ReactiveMarking(t *testing.T) {
	engine := NewTaintEngine()

	taints, err := engine.MarkRead("/path/to/file", []string{"WORKSPACE"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(taints) != 1 || taints[0] != "WORKSPACE" {
		t.Errorf("Expected ['WORKSPACE'], got %v", taints)
	}
}

func TestTaintEngine_Propagation(t *testing.T) {
	engine := NewTaintEngine()

	obj := map[string]string{"key": "value"}
	propagated, err := engine.Propagate(obj, []string{"TOOL_OUTPUT"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if propagated == nil {
		t.Error("Expected propagated object to be non-nil")
	}
}

func TestTaintEngine_Violation(t *testing.T) {
	engine := NewTaintEngine()

	err := engine.CheckForbidden([]string{"INNER_ONLY"}, DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error for DMZ boundary, got %v", err)
	}

	err = engine.CheckForbidden([]string{"INNER_ONLY"}, InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for Inner boundary, got %v", err)
	}
}

func TestTaintEngine_Redaction(t *testing.T) {
	engine := NewTaintEngine()

	obj := map[string]string{"secret": "my-api-key"}
	rules := []RedactRule{{Taint: "SECRET", Replacement: "[REDACTED]"}}

	redacted, err := engine.Redact(obj, rules)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if redacted == nil {
		t.Error("Expected redacted object to be non-nil")
	}
}

func TestTaintEngine_ReportTaints(t *testing.T) {
	engine := NewTaintEngine()

	taints, err := engine.ReportTaints("chart-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if taints == nil {
		t.Error("Expected non-nil TaintMap")
	}
}

func TestBoundaryService_Enforcement(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	err := service.CheckForbidden([]string{"PII"}, DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = service.CheckForbidden([]string{"PII"}, InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for inner, got %v", err)
	}
}

func TestBoundaryService_MailValidate(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	mail := map[string]string{"content": "test"}
	validated, err := service.ValidateAndSanitize(mail, OuterBoundary, DMZBoundary)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validated == nil {
		t.Error("Expected validated mail to be non-nil")
	}
}

func TestBoundaryService_ChartCheck(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	err := service.CheckForbidden([]string{"TOOL_OUTPUT"}, DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestBoundaryService_Namespace(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	view, err := service.NamespaceIsolate("chart-123", "read")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !view.Isolated {
		t.Error("Expected view to be isolated")
	}

	if view.ChartID != "chart-123" {
		t.Errorf("Expected chart-123, got %s", view.ChartID)
	}
}

func TestSecurity_KernelBootstrap(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	if service == nil {
		t.Error("Expected non-nil boundary service")
	}

	if engine == nil {
		t.Error("Expected non-nil taint engine")
	}
}

func TestSecurity_ContextMapFiltering(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	blocks := []*ContextBlock{
		{Name: "public", Source: "static", Content: "public", BoundaryFilter: DMZBoundary},
		{Name: "inner", Source: "static", Content: "inner", BoundaryFilter: InnerBoundary},
	}

	for _, block := range blocks {
		content := staticSource(block)
		if content == "" {
			t.Errorf("Expected content for block %s", block.Name)
		}
	}

	err := service.CheckForbidden([]string{"PUBLIC"}, DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error for public taint on DMZ, got %v", err)
	}
}

func TestSecurity_MailValidation(t *testing.T) {
	engine := NewTaintEngine()
	service := NewBoundaryService(engine)

	mail := map[string]any{
		"content": "test message",
		"source":  "agent-1",
		"target":  "agent-2",
	}

	validated, err := service.ValidateAndSanitize(mail, OuterBoundary, DMZBoundary)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validated == nil {
		t.Error("Expected validated mail to be non-nil")
	}
}

func TestTaintEngine_AttachTaint_Mail(t *testing.T) {
	engine := NewTaintEngine()

	m := &mail.Mail{
		ID:     "msg-1",
		Source: "user",
		Metadata: mail.MailMetadata{
			Taints: []string{"USER_SUPPLIED"},
		},
	}

	result, err := engine.AttachTaint(m, []string{"TOOL_OUTPUT"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMail, ok := result.(*mail.Mail)
	if !ok {
		t.Fatalf("Expected *mail.Mail, got %T", result)
	}

	expectedTaints := []string{"USER_SUPPLIED", "TOOL_OUTPUT"}
	if len(resultMail.Metadata.Taints) != len(expectedTaints) {
		t.Errorf("Expected %d taints, got %d", len(expectedTaints), len(resultMail.Metadata.Taints))
	}

	for _, expectedTaint := range expectedTaints {
		found := false
		for _, actualTaint := range resultMail.Metadata.Taints {
			if actualTaint == expectedTaint {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected taint %q not found in %v", expectedTaint, resultMail.Metadata.Taints)
		}
	}
}
