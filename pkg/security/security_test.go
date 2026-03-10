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

	err := engine.CheckForbidden([]string{"INNER_ONLY"}, InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for INNER_ONLY on Inner boundary, got %v", err)
	}
}

func TestTaintEngine_CheckViolation_INNER_ONLY_on_DMZ(t *testing.T) {
	engine := NewTaintEngine()

	err := engine.CheckForbidden([]string{"INNER_ONLY"}, DMZBoundary)
	if err == nil {
		t.Error("Expected error for INNER_ONLY taint on DMZ boundary, got nil")
	}

	expectedMsg := "taint INNER_ONLY is forbidden on boundary dmz"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestTaintEngine_CheckViolation_SECRET_on_Outer(t *testing.T) {
	engine := NewTaintEngine()

	err := engine.CheckForbidden([]string{"SECRET"}, OuterBoundary)
	if err == nil {
		t.Error("Expected error for SECRET taint on Outer boundary, got nil")
	}

	expectedMsg := "taint SECRET is forbidden on boundary outer"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestTaintEngine_CheckViolation_AllowedTaints(t *testing.T) {
	engine := NewTaintEngine()

	err := engine.CheckForbidden([]string{"TOOL_OUTPUT"}, DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error for TOOL_OUTPUT on DMZ boundary, got %v", err)
	}

	err = engine.CheckForbidden([]string{"PUBLIC"}, OuterBoundary)
	if err != nil {
		t.Errorf("Expected no error for PUBLIC taint on Outer boundary, got %v", err)
	}

	err = engine.CheckForbidden([]string{"INNER_ONLY"}, InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for INNER_ONLY on Inner boundary, got %v", err)
	}

	err = engine.CheckForbidden([]string{}, DMZBoundary)
	if err != nil {
		t.Errorf("Expected no error for empty taints, got %v", err)
	}

	err = engine.CheckForbidden([]string{"SECRET"}, InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for SECRET on Inner boundary, got %v", err)
	}

	err = engine.CheckForbidden([]string{"PII"}, OuterBoundary)
	if err == nil {
		t.Error("Expected error for PII taint on Outer boundary, got nil")
	}

	expectedMsg := "taint PII is forbidden on boundary outer"
	if err != nil && err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
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

	if view.RuntimeID != "chart-123" {
		t.Errorf("Expected chart-123, got %s", view.RuntimeID)
	}

	if view.Boundary != DMZBoundary {
		t.Errorf("Expected DMZBoundary, got %s", view.Boundary)
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

func TestTaintSet_AddHas(t *testing.T) {
	var set TaintSet

	set.Add("PII")

	if !set.Has("PII") {
		t.Error("Expected Has(\"PII\") to return true after Add(\"PII\")")
	}

	if set.Has("SECRET") {
		t.Error("Expected Has(\"SECRET\") to return false")
	}
}

func TestTaintSet_Union(t *testing.T) {
	set1 := TaintSet{"PII": true, "SECRET": true}
	set2 := TaintSet{"SECRET": true, "TOOL_OUTPUT": true}

	result := set1.Union(set2)

	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected union to have %d elements, got %d", expectedCount, len(result))
	}

	if !result.Has("PII") {
		t.Error("Expected union to contain PII")
	}

	if !result.Has("SECRET") {
		t.Error("Expected union to contain SECRET")
	}

	if !result.Has("TOOL_OUTPUT") {
		t.Error("Expected union to contain TOOL_OUTPUT")
	}
}

func TestTaintSet_EmptyOperations(t *testing.T) {
	var emptySet TaintSet

	if emptySet.Has("PII") {
		t.Error("Expected Has on empty set to return false")
	}

	emptySet.Add("PII")
	if !emptySet.Has("PII") {
		t.Error("Expected Add to work on empty set")
	}

	setWithTaints := TaintSet{"SECRET": true, "TOOL_OUTPUT": true}
	var anotherEmpty TaintSet

	union1 := setWithTaints.Union(anotherEmpty)
	if len(union1) != 2 {
		t.Errorf("Expected union with empty set to return original, got %d elements", len(union1))
	}

	union2 := anotherEmpty.Union(setWithTaints)
	if len(union2) != 2 {
		t.Errorf("Expected union of empty with set to return set, got %d elements", len(union2))
	}

	nilSetUnion := TaintSet(nil).Union(TaintSet{"PII": true})
	if len(nilSetUnion) != 1 {
		t.Errorf("Expected union with nil set to work, got %d elements", len(nilSetUnion))
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

func TestTaintEngine_StripTaint_Basic(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"_taints": []string{"PII", "SECRET", "TOOL_OUTPUT"},
		"key":     "value",
	}

	forbidden := []string{"PII", "SECRET"}
	result, _, err := engine.StripTaint(data, forbidden)
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

	if len(taints) != 1 || taints[0] != "TOOL_OUTPUT" {
		t.Errorf("Expected ['TOOL_OUTPUT'], got %v", taints)
	}

	if resultMap["key"] != "value" {
		t.Errorf("Expected key to be 'value', got %v", resultMap["key"])
	}
}

func TestTaintEngine_StripTaint_ReturnsStripped(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"_taints": []string{"PII", "SECRET", "TOOL_OUTPUT"},
		"key":     "value",
	}

	forbidden := []string{"PII"}
	_, stripped, err := engine.StripTaint(data, forbidden)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(stripped) != 1 || stripped[0] != "PII" {
		t.Errorf("Expected stripped list ['PII'], got %v", stripped)
	}
}

func TestTaintEngine_StripTaint_Nested(t *testing.T) {
	engine := NewTaintEngine()

	nested := map[string]interface{}{
		"_taints": []string{"PII", "INNER_ONLY"},
		"level1": map[string]interface{}{
			"_taints": []string{"SECRET", "INNER_ONLY"},
			"level2": map[string]interface{}{
				"_taints": []string{"INNER_ONLY", "TOOL_OUTPUT"},
				"data":    "secret",
			},
		},
	}

	forbidden := []string{"INNER_ONLY"}
	result, stripped, err := engine.StripTaint(nested, forbidden)
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

	if len(rootTaints) != 1 || rootTaints[0] != "PII" {
		t.Errorf("Expected root to have ['PII'], got %v", rootTaints)
	}

	level1, ok := resultMap["level1"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected level1 to be map[string]interface{}, got %T", resultMap["level1"])
	}

	level1Taints, ok := level1["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected level1 _taints to be []string, got %T", level1["_taints"])
	}

	if len(level1Taints) != 1 || level1Taints[0] != "SECRET" {
		t.Errorf("Expected level1 to have ['SECRET'], got %v", level1Taints)
	}

	level2, ok := level1["level2"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected level2 to be map[string]interface{}, got %T", level1["level2"])
	}

	level2Taints, ok := level2["_taints"].([]string)
	if !ok {
		t.Fatalf("Expected level2 _taints to be []string, got %T", level2["_taints"])
	}

	if len(level2Taints) != 1 || level2Taints[0] != "TOOL_OUTPUT" {
		t.Errorf("Expected level2 to have ['TOOL_OUTPUT'], got %v", level2Taints)
	}

	if len(stripped) != 3 {
		t.Errorf("Expected 3 stripped taints, got %d: %v", len(stripped), stripped)
	}
}

func TestTaintEngine_StripTaint_NoForbidden(t *testing.T) {
	engine := NewTaintEngine()

	data := map[string]interface{}{
		"_taints": []string{"TOOL_OUTPUT", "EXTERNAL"},
		"key":     "value",
	}

	forbidden := []string{"PII", "SECRET"}
	result, stripped, err := engine.StripTaint(data, forbidden)
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
		t.Errorf("Expected 2 taints unchanged, got %d: %v", len(taints), taints)
	}

	if len(stripped) != 0 {
		t.Errorf("Expected empty stripped list, got %v", stripped)
	}

	if resultMap["key"] != "value" {
		t.Errorf("Expected key to be 'value', got %v", resultMap["key"])
	}
}

func TestToolTaintOutput_AutoAttach(t *testing.T) {
	registry := NewToolRegistry()
	registry.RegisterTool(&ToolConfig{
		Name:        "webSearch",
		Boundary:    mail.DMZBoundary,
		TaintOutput: []string{"TOOL_OUTPUT"},
	})

	resultMail := &mail.Mail{
		ID:     "result-1",
		Type:   mail.MailTypeToolResult,
		Source: "sys:tools",
		Target: "agent:user",
		Content: map[string]interface{}{
			"query": "test",
			"data":  "search results",
		},
		Metadata: mail.MailMetadata{
			Taints: []string{},
		},
	}

	result, err := AttachToolTaints("webSearch", resultMail, registry)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMail, ok := result.(*mail.Mail)
	if !ok {
		t.Fatalf("Expected *mail.Mail, got %T", result)
	}

	if len(resultMail.Metadata.Taints) != 1 {
		t.Errorf("Expected 1 taint, got %d: %v", len(resultMail.Metadata.Taints), resultMail.Metadata.Taints)
	}

	if resultMail.Metadata.Taints[0] != "TOOL_OUTPUT" {
		t.Errorf("Expected taint 'TOOL_OUTPUT', got %v", resultMail.Metadata.Taints)
	}
}

func TestToolTaintOutput_InheritBoundary(t *testing.T) {
	registry := NewToolRegistry()
	registry.RegisterTool(&ToolConfig{
		Name:        "innerDbQuery",
		Boundary:    mail.InnerBoundary,
		TaintOutput: []string{"TOOL_OUTPUT", "INNER_ONLY"},
	})

	resultMail := &mail.Mail{
		ID:     "result-2",
		Type:   mail.MailTypeToolResult,
		Source: "sys:tools",
		Target: "agent:user",
		Content: map[string]interface{}{
			"query": "SELECT * FROM users",
			"data":  "db results",
		},
		Metadata: mail.MailMetadata{
			Taints: []string{},
		},
	}

	result, err := AttachToolTaints("innerDbQuery", resultMail, registry)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMail, ok := result.(*mail.Mail)
	if !ok {
		t.Fatalf("Expected *mail.Mail, got %T", result)
	}

	if resultMail.Metadata.Boundary != mail.InnerBoundary {
		t.Errorf("Expected boundary 'inner', got %v", resultMail.Metadata.Boundary)
	}

	if len(resultMail.Metadata.Taints) != 2 {
		t.Errorf("Expected 2 taints, got %d: %v", len(resultMail.Metadata.Taints), resultMail.Metadata.Taints)
	}

	expectedTaints := map[string]bool{"TOOL_OUTPUT": true, "INNER_ONLY": true}
	for _, taint := range resultMail.Metadata.Taints {
		if !expectedTaints[taint] {
			t.Errorf("Unexpected taint %q", taint)
		}
	}
}

func TestToolTaintOutput_NoAttach(t *testing.T) {
	registry := NewToolRegistry()
	registry.RegisterTool(&ToolConfig{
		Name:        "safeTool",
		Boundary:    mail.OuterBoundary,
		TaintOutput: []string{},
	})

	resultMail := &mail.Mail{
		ID:     "result-3",
		Type:   mail.MailTypeToolResult,
		Source: "sys:tools",
		Target: "agent:user",
		Content: map[string]interface{}{
			"data": "safe results",
		},
		Metadata: mail.MailMetadata{
			Taints: []string{},
		},
	}

	result, err := AttachToolTaints("safeTool", resultMail, registry)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMail, ok := result.(*mail.Mail)
	if !ok {
		t.Fatalf("Expected *mail.Mail, got %T", result)
	}

	if len(resultMail.Metadata.Taints) != 0 {
		t.Errorf("Expected no automatic taints, got %v", resultMail.Metadata.Taints)
	}
}

func TestReportTaints_GeneratesTaintMap(t *testing.T) {
	engineImpl := &taintEngineImpl{
		taints: TaintMap{
			"user":   []string{"PII"},
			"secret": []string{"SECRET"},
		},
	}

	service := NewBoundaryService(engineImpl)
	taintMap, err := service.ReportTaints("runtime-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(taintMap) != 2 {
		t.Errorf("Expected 2 entries in TaintMap, got %d", len(taintMap))
	}

	userTaints, ok := taintMap["user"]
	if !ok {
		t.Error("Expected 'user' key in TaintMap")
	} else if len(userTaints) != 1 || userTaints[0] != "PII" {
		t.Errorf("Expected user taints ['PII'], got %v", userTaints)
	}

	secretTaints, ok := taintMap["secret"]
	if !ok {
		t.Error("Expected 'secret' key in TaintMap")
	} else if len(secretTaints) != 1 || secretTaints[0] != "SECRET" {
		t.Errorf("Expected secret taints ['SECRET'], got %v", secretTaints)
	}
}

func TestReportTaints_RecursiveCollection(t *testing.T) {
	engineImpl := &taintEngineImpl{
		taints: TaintMap{
			"config.inner.key": []string{"SECRET", "INNER_ONLY"},
		},
	}

	service := NewBoundaryService(engineImpl)
	taintMap, err := service.ReportTaints("runtime-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedKey := "config.inner.key"
	taints, ok := taintMap[expectedKey]
	if !ok {
		t.Errorf("Expected key '%s' in TaintMap", expectedKey)
	} else if len(taints) != 2 {
		t.Errorf("Expected 2 taints for '%s', got %d: %v", expectedKey, len(taints), taints)
	} else {
		hasSecret := false
		hasInnerOnly := false
		for _, t := range taints {
			if t == "SECRET" {
				hasSecret = true
			}
			if t == "INNER_ONLY" {
				hasInnerOnly = true
			}
		}
		if !hasSecret {
			t.Error("Expected SECRET taint")
		}
		if !hasInnerOnly {
			t.Error("Expected INNER_ONLY taint")
		}
	}
}

func TestReportTaints_EmptyForUntainted(t *testing.T) {
	engineImpl := &taintEngineImpl{
		taints: make(TaintMap),
	}

	service := NewBoundaryService(engineImpl)
	taintMap, err := service.ReportTaints("runtime-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(taintMap) != 0 {
		t.Errorf("Expected empty TaintMap, got %d entries: %v", len(taintMap), taintMap)
	}
}
