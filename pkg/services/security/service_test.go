package security

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

func TestSecurityService_NewSecurityServiceReturnsNonNil(t *testing.T) {
	svc := NewSecurityService()

	if svc == nil {
		t.Error("Expected NewSecurityService to return non-nil")
	}
}

func TestSecurityService_IDReturnsCorrectString(t *testing.T) {
	svc := NewSecurityService()

	id := svc.ID()

	if id != "sys:security" {
		t.Errorf("Expected ID sys:security, got %s", id)
	}
}

func TestSecurityService_HandleMailReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	mail := mail.Mail{}
	err := svc.HandleMail(&mail)

	if err != nil {
		t.Errorf("Expected HandleMail to return nil, got %v", err)
	}
}

func TestSecurityService_ValidateBoundaryReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.ValidateBoundary(mail.InnerBoundary, mail.DMZBoundary)

	if err != nil {
		t.Errorf("Expected ValidateBoundary to return nil, got %v", err)
	}
}

func TestSecurityService_StartReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.Start()

	if err != nil {
		t.Errorf("Expected Start to return nil, got %v", err)
	}
}

func TestSecurityService_StopReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.Stop()

	if err != nil {
		t.Errorf("Expected Stop to return nil, got %v", err)
	}
}

func TestSecurityService_BootstrapChart(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:security" {
		t.Errorf("Expected ID sys:security, got %s", chart.ID)
	}

	if chart.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", chart.Version)
	}
}

func TestSecurityService_HandleMail(t *testing.T) {
	svc := NewSecurityService()
	m := mail.Mail{
		ID:     "test-mail-1",
		Source: "test-source",
		Target: "sys:security",
	}
	err := svc.HandleMail(&m)
	if err != nil {
		t.Errorf("HandleMail should return nil for Phase 1 pass-through, got: %v", err)
	}
}

func TestSecurityService_ID(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:security" {
		t.Errorf("Expected ID sys:security, got %s", chart.ID)
	}
}

func TestSecurityService_ValidateAndSanitizePassThrough(t *testing.T) {
	svc := NewSecurityService()

	inputMail := mail.Mail{
		ID:     "test-mail-1",
		Source: "agent:test",
		Target: "sys:security",
	}

	result, err := svc.ValidateAndSanitize(inputMail, mail.InnerBoundary, mail.DMZBoundary)

	if err != nil {
		t.Errorf("Expected ValidateAndSanitize to return nil error, got %v", err)
	}

	if result.ID != inputMail.ID {
		t.Errorf("Expected mail ID to be unchanged, got %s", result.ID)
	}

	if result.Source != inputMail.Source {
		t.Errorf("Expected mail Source to be unchanged, got %s", result.Source)
	}
}

func TestSecurityService_TaintPropagateReturnsObject(t *testing.T) {
	svc := NewSecurityService()

	inputObj := "test-string"

	result, err := svc.TaintPropagate(inputObj, []string{"PII", "SECRET"})

	if err != nil {
		t.Errorf("Expected TaintPropagate to return nil error, got %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Error("Expected result to be string type")
	}

	if resultStr != inputObj {
		t.Errorf("Expected object to be unchanged, got %s", resultStr)
	}
}

func TestSecurityService_ReportTaintsReturnsEmptyMap(t *testing.T) {
	svc := NewSecurityService()

	result, err := svc.ReportTaints("runtime-123")

	if err != nil {
		t.Errorf("Expected ReportTaints to return nil error, got %v", err)
	}

	if result == nil {
		t.Error("Expected ReportTaints to return non-nil TaintMap")
	}

	if len(result) != 0 {
		t.Errorf("Expected empty TaintMap, got %d entries", len(result))
	}
}

func TestSecurityService_PrepareContextForBoundaryNoOp(t *testing.T) {
	svc := NewSecurityService()

	err := svc.PrepareContextForBoundary("runtime-123", mail.InnerBoundary)

	if err != nil {
		t.Errorf("Expected PrepareContextForBoundary to return nil for InnerBoundary, got %v", err)
	}

	err = svc.PrepareContextForBoundary("runtime-123", mail.OuterBoundary)

	if err != nil {
		t.Errorf("Expected PrepareContextForBoundary to return nil for OuterBoundary, got %v", err)
	}
}

func TestSecurityService_ValidateAndSanitize_innerToOuter(t *testing.T) {
	svc := NewSecurityService()

	inputMail := mail.Mail{
		ID:     "test-mail-inner-outer",
		Source: "agent:test",
		Target: "sys:security",
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Taints:   []string{"INTERNAL"},
		},
	}

	result, err := svc.ValidateAndSanitize(inputMail, mail.InnerBoundary, mail.OuterBoundary)

	if err != nil {
		t.Errorf("Expected inner→outer transition to be allowed, got error: %v", err)
	}

	if result.Metadata.Boundary != mail.OuterBoundary {
		t.Errorf("Expected boundary to be updated to OuterBoundary, got: %s", result.Metadata.Boundary)
	}
}

func TestSecurityService_ValidateAndSanitize_outerToInner(t *testing.T) {
	svc := NewSecurityService()

	inputMail := mail.Mail{
		ID:     "test-mail-outer-inner",
		Source: "agent:external",
		Target: "sys:security",
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{},
		},
	}

	result, err := svc.ValidateAndSanitize(inputMail, mail.OuterBoundary, mail.InnerBoundary)

	if err != nil {
		t.Errorf("Expected outer→inner transition to be allowed, got error: %v", err)
	}

	hasExternalTaint := false
	for _, taint := range result.Metadata.Taints {
		if taint == "EXTERNAL" {
			hasExternalTaint = true
			break
		}
	}

	if !hasExternalTaint {
		t.Errorf("Expected EXTERNAL taint to be added for outer→inner transition, got: %v", result.Metadata.Taints)
	}
}

func TestSecurityService_NamespaceIsolate(t *testing.T) {
	svc := NewSecurityService()

	_, err := svc.NamespaceIsolate("agent-1", "read")
	if err == nil {
		t.Error("Expected NamespaceIsolate to return error for stub")
	}
}

func TestSecurityService_NamespaceIsolate_multipleAgents(t *testing.T) {
	svc := NewSecurityService()

	_, err := svc.NamespaceIsolate("agent-alpha", "read")
	if err == nil {
		t.Error("Expected NamespaceIsolate to return error for stub")
	}
}

func TestSecurityService_CheckTaintPolicy_allowed(t *testing.T) {
	svc := NewSecurityService()

	_, err := svc.CheckTaintPolicy(nil, mail.OuterBoundary, security.TaintPolicy{})
	if err == nil {
		t.Error("Expected CheckTaintPolicy to return error for stub")
	}
}

func TestSecurityService_CheckTaintPolicy_denied(t *testing.T) {
	svc := NewSecurityService()

	_, err := svc.CheckTaintPolicy(nil, mail.OuterBoundary, security.TaintPolicy{})
	if err == nil {
		t.Error("Expected CheckTaintPolicy to return error for stub")
	}
}

func TestSecurityService_TaintPropagate_addTaints(t *testing.T) {
	svc := NewSecurityService()

	inputObj := map[string]interface{}{
		"name": "test",
		"data": "value",
	}

	result, err := svc.TaintPropagate(inputObj, []string{"PII"})

	if err != nil {
		t.Errorf("Expected TaintPropagate to return nil error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be map[string]interface{}")
	}

	if _, hasTaints := resultMap["_taints"]; !hasTaints {
		t.Error("Expected _taints key to be added to object")
	}

	taints, _ := resultMap["_taints"].([]string)
	if len(taints) != 1 || taints[0] != "PII" {
		t.Errorf("Expected taints to be [PII], got %v", taints)
	}

	if resultMap["name"] != "test" {
		t.Error("Expected original data to be preserved")
	}
}

func TestSecurityService_TaintPropagate_mergeTaints(t *testing.T) {
	svc := NewSecurityService()

	inputObj := map[string]interface{}{
		"name":    "test",
		"data":    "value",
		"_taints": []string{"EXTERNAL"},
	}

	result, err := svc.TaintPropagate(inputObj, []string{"PII", "SECRET"})

	if err != nil {
		t.Errorf("Expected TaintPropagate to return nil error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be map[string]interface{}")
	}

	taints, _ := resultMap["_taints"].([]string)
	if len(taints) != 3 {
		t.Errorf("Expected 3 taints after merge, got %d", len(taints))
	}

	hasExternal := false
	hasPII := false
	hasSecret := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
		}
		if t == "PII" {
			hasPII = true
		}
		if t == "SECRET" {
			hasSecret = true
		}
	}

	if !hasExternal || !hasPII || !hasSecret {
		t.Errorf("Expected all taints to be merged, got %v", taints)
	}
}

func TestSecurityService_TaintPropagate_nestedObjects(t *testing.T) {
	svc := NewSecurityService()

	inputObj := map[string]interface{}{
		"name": "test",
		"nested": map[string]interface{}{
			"value": "nested-value",
			"deep": map[string]interface{}{
				"data": "deep-data",
			},
		},
	}

	result, err := svc.TaintPropagate(inputObj, []string{"PII"})

	if err != nil {
		t.Errorf("Expected TaintPropagate to return nil error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be map[string]interface{}")
	}

	if _, hasTaints := resultMap["_taints"]; !hasTaints {
		t.Error("Expected _taints key to be added to root object")
	}

	nested, ok := resultMap["nested"].(map[string]interface{})
	if !ok {
		t.Error("Expected nested to be map[string]interface{}")
	}

	if _, hasNestedTaints := nested["_taints"]; !hasNestedTaints {
		t.Error("Expected _taints key to be added to nested object")
	}

	deep, ok := nested["deep"].(map[string]interface{})
	if !ok {
		t.Error("Expected deep to be map[string]interface{}")
	}

	if _, hasDeepTaints := deep["_taints"]; !hasDeepTaints {
		t.Error("Expected _taints key to be added to deep object")
	}
}

func TestSecurityService_HandleMail_BoundaryTransition(t *testing.T) {
	svc := NewSecurityService()

	inputMail := mail.Mail{
		ID:     "test-mail-boundary",
		Source: "agent:external",
		Target: "sys:security",
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}

	err := svc.HandleMail(&inputMail)

	if err != nil {
		t.Errorf("Expected HandleMail to return nil error, got %v", err)
	}

	hasUserSupplied := false
	hasExternal := false
	for _, taint := range inputMail.Metadata.Taints {
		if taint == "USER_SUPPLIED" {
			hasUserSupplied = true
		}
		if taint == "EXTERNAL" {
			hasExternal = true
		}
	}

	if !hasUserSupplied {
		t.Error("Expected USER_SUPPLIED taint to be preserved")
	}

	if !hasExternal {
		t.Error("Expected EXTERNAL taint to be added for outer→inner transition")
	}

	if inputMail.Metadata.Boundary != mail.InnerBoundary {
		t.Errorf("Expected boundary to be updated to inner, got %s", inputMail.Metadata.Boundary)
	}
}

func TestSecurityService_TaintPropagate_sliceObjects(t *testing.T) {
	svc := NewSecurityService()

	inputObj := map[string]interface{}{
		"name": "test",
		"items": []interface{}{
			map[string]interface{}{"id": 1, "value": "first"},
			map[string]interface{}{"id": 2, "value": "second"},
		},
	}

	result, err := svc.TaintPropagate(inputObj, []string{"SECRET"})

	if err != nil {
		t.Errorf("Expected TaintPropagate to return nil error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be map[string]interface{}")
	}

	if _, hasTaints := resultMap["_taints"]; !hasTaints {
		t.Error("Expected _taints key to be added to root object")
	}

	items, ok := resultMap["items"].([]interface{})
	if !ok {
		t.Error("Expected items to be []interface{}")
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	firstItem, ok := items[0].(map[string]interface{})
	if !ok {
		t.Error("Expected first item to be map[string]interface{}")
	}

	if _, hasFirstTaints := firstItem["_taints"]; !hasFirstTaints {
		t.Error("Expected _taints key to be added to first item")
	}

	secondItem, ok := items[1].(map[string]interface{})
	if !ok {
		t.Error("Expected second item to be map[string]interface{}")
	}

	if _, hasSecondTaints := secondItem["_taints"]; !hasSecondTaints {
		t.Error("Expected _taints key to be added to second item")
	}
}
