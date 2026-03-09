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

func TestTaintEngine_AttachTaint_Map(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	result, err := engine.AttachTaint(data, []string{"PII", "SECRET"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	if resultMap["key1"] != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %v", resultMap["key1"])
	}

	if resultMap["key2"] != 123 {
		t.Errorf("Expected key2 to be 123, got %v", resultMap["key2"])
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected _taints key with []string value, got %T", resultMap["_taints"])
	}

	if len(taints) != 2 {
		t.Errorf("Expected 2 taints, got %d", len(taints))
	}

	expectedTaints := map[string]bool{"PII": true, "SECRET": true}
	for _, taint := range taints {
		if !expectedTaints[taint] {
			t.Errorf("Unexpected taint %q", taint)
		}
	}
}

func TestTaintEngine_AttachTaint_Nested(t *testing.T) {
	engine := NewTaintEngine()

	nested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"data": "secret",
			},
		},
	}

	result, err := engine.AttachTaint(nested, []string{"INNER_ONLY"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	if _, hasTaints := resultMap["_taints"]; !hasTaints {
		t.Error("Expected root map to have _taints key")
	}

	level1, ok := resultMap["level1"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected level1 to be map[string]interface{}, got %T", resultMap["level1"])
	}

	if _, hasTaints := level1["_taints"]; !hasTaints {
		t.Error("Expected level1 to have _taints key")
	}

	level2, ok := level1["level2"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected level2 to be map[string]interface{}, got %T", level1["level2"])
	}

	if _, hasTaints := level2["_taints"]; !hasTaints {
		t.Error("Expected level2 to have _taints key")
	}

	if level2["data"] != "secret" {
		t.Errorf("Expected data to be 'secret', got %v", level2["data"])
	}
}

func TestTaintEngine_AttachTaint_Empty(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"key": "value",
	}

	result, err := engine.AttachTaint(data, []string{})
	if err != nil {
		t.Fatalf("Expected no error for empty taints, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected _taints key with []string value, got %T", resultMap["_taints"])
	}

	if len(taints) != 0 {
		t.Errorf("Expected empty taints, got %v", taints)
	}

	nilResult, nilErr := engine.AttachTaint(nil, []string{"TAINT"})
	if nilErr == nil {
		t.Error("Expected error for nil object, got nil")
	}

	if nilResult != nil {
		t.Errorf("Expected nil result for nil input, got %v", nilResult)
	}
}

func TestTaintEngine_PropagateTaint_MapMerge(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"_taints": []string{"PII"},
		"key":     "value",
	}

	result, err := engine.Propagate(data, []string{"TOOL_OUTPUT"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected _taints key with []string value, got %T", resultMap["_taints"])
	}

	if len(taints) != 2 {
		t.Errorf("Expected 2 taints, got %d", len(taints))
	}

	expectedTaints := map[string]bool{"PII": true, "TOOL_OUTPUT": true}
	for _, taint := range taints {
		if !expectedTaints[taint] {
			t.Errorf("Unexpected taint %q", taint)
		}
	}
}

func TestTaintEngine_PropagateTaint_NestedMap(t *testing.T) {
	engine := NewTaintEngine()

	nested := map[string]interface{}{
		"_taints": []string{"PII"},
		"level1": map[string]interface{}{
			"_taints": []string{"PII"},
			"level2": map[string]interface{}{
				"_taints": []string{"PII"},
				"data":    "secret",
			},
		},
	}

	result, err := engine.Propagate(nested, []string{"SECRET"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	rootTaints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected root _taints to be []string, got %T", resultMap["_taints"])
	}

	if len(rootTaints) != 2 {
		t.Errorf("Expected root to have 2 taints, got %d", len(rootTaints))
	}

	level1, ok := resultMap["level1"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected level1 to be map[string]interface{}, got %T", resultMap["level1"])
	}

	level1Taints, ok := level1["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected level1 _taints to be []string, got %T", level1["_taints"])
	}

	if len(level1Taints) != 2 {
		t.Errorf("Expected level1 to have 2 taints, got %d", len(level1Taints))
	}

	level2, ok := level1["level2"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected level2 to be map[string]interface{}, got %T", level1["level2"])
	}

	level2Taints, ok := level2["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected level2 _taints to be []string, got %T", level2["_taints"])
	}

	if len(level2Taints) != 2 {
		t.Errorf("Expected level2 to have 2 taints, got %d", len(level2Taints))
	}

	expectedTaints := map[string]bool{"PII": true, "SECRET": true}
	for _, taint := range rootTaints {
		if !expectedTaints[taint] {
			t.Errorf("Unexpected root taint %q", taint)
		}
	}
}

func TestTaintEngine_PropagateTaint_Slice(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"_taints": []string{"PII"},
		"items": []interface{}{
			map[string]interface{}{
				"_taints": []string{"PII"},
				"name":    "item1",
			},
			map[string]interface{}{
				"_taints": []string{"PII"},
				"name":    "item2",
			},
			"non-map-element",
		},
	}

	result, err := engine.Propagate(data, []string{"WORKSPACE"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	rootTaints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected root _taints to be []string, got %T", resultMap["_taints"])
	}

	if len(rootTaints) != 2 {
		t.Errorf("Expected root to have 2 taints, got %d", len(rootTaints))
	}

	items, ok := resultMap["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be []interface{}, got %T", resultMap["items"])
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	item1, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected item1 to be map[string]interface{}, got %T", items[0])
	}

	item1Taints, ok := item1["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected item1 _taints to be []string, got %T", item1["_taints"])
	}

	if len(item1Taints) != 2 {
		t.Errorf("Expected item1 to have 2 taints, got %d", len(item1Taints))
	}

	item2, ok := items[1].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected item2 to be map[string]interface{}, got %T", items[1])
	}

	item2Taints, ok := item2["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected item2 _taints to be []string, got %T", item2["_taints"])
	}

	if len(item2Taints) != 2 {
		t.Errorf("Expected item2 to have 2 taints, got %d", len(item2Taints))
	}

	if items[2] != "non-map-element" {
		t.Errorf("Expected non-map element unchanged, got %v", items[2])
	}

	expectedTaints := map[string]bool{"PII": true, "WORKSPACE": true}
	for _, taint := range item1Taints {
		if !expectedTaints[taint] {
			t.Errorf("Unexpected item1 taint %q", taint)
		}
	}
}

func TestTaintEngine_PropagateTaint_Deduplication(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"_taints": []string{"TOOL_OUTPUT", "PII", "SECRET"},
		"key":     "value",
	}

	result, err := engine.Propagate(data, []string{"SECRET", "PII", "WORKSPACE"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected _taints key with []string value, got %T", resultMap["_taints"])
	}

	if len(taints) != 4 {
		t.Fatalf("Expected 4 taints (no duplicates), got %d: %v", len(taints), taints)
	}

	expectedTaints := []string{"TOOL_OUTPUT", "PII", "SECRET", "WORKSPACE"}
	for i, expectedTaint := range expectedTaints {
		if taints[i] != expectedTaint {
			t.Errorf("Expected taint at index %d to be %q, got %q (order must preserve first occurrence)", i, expectedTaint, taints[i])
		}
	}
}

func TestTaintEngine_PropagateTaint_EmptySource(t *testing.T) {
	engine := NewTaintEngine()

	dataNoTaints := map[string]interface{}{
		"key": "value",
	}

	result, err := engine.Propagate(dataNoTaints, []string{"NEW_TAINT"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected _taints key with []string value, got %T", resultMap["_taints"])
	}

	if len(taints) != 1 || taints[0] != "NEW_TAINT" {
		t.Errorf("Expected ['NEW_TAINT'], got %v", taints)
	}

	existingTaints := map[string]interface{}{
		"_taints": []string{"EXISTING"},
		"key":     "value",
	}

	result2, err := engine.Propagate(existingTaints, []string{})
	if err != nil {
		t.Fatalf("Expected no error for empty newTaints, got %v", err)
	}

	resultMap2, ok := result2.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result2)
	}

	taints2, ok := resultMap2["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected _taints key with []string value, got %T", resultMap2["_taints"])
	}

	if len(taints2) != 1 || taints2[0] != "EXISTING" {
		t.Errorf("Expected ['EXISTING'] preserved, got %v", taints2)
	}
}
