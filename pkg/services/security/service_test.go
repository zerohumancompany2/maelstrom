package security

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
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

	err := svc.HandleMail(mail.Mail{})

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
	mail := mail.Mail{
		ID:     "test-mail-1",
		Source: "test-source",
		Target: "sys:security",
	}
	err := svc.HandleMail(mail)
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

	data := []interface{}{
		map[string]interface{}{"agentID": "agent-1", "value": "data-1"},
		map[string]interface{}{"agentID": "agent-2", "value": "data-2"},
		map[string]interface{}{"agentID": "agent-1", "value": "data-3"},
	}

	result := svc.NamespaceIsolate(data, "agent-1")

	resultSlice, ok := result.([]interface{})
	if !ok {
		t.Error("Expected result to be []interface{}")
	}

	if len(resultSlice) != 2 {
		t.Errorf("Expected 2 items for agent-1, got %d", len(resultSlice))
	}

	if resultSlice[0].(map[string]interface{})["value"] != "data-1" {
		t.Error("Expected first item to be data-1")
	}

	if resultSlice[1].(map[string]interface{})["value"] != "data-3" {
		t.Error("Expected second item to be data-3")
	}
}

func TestSecurityService_NamespaceIsolate_multipleAgents(t *testing.T) {
	svc := NewSecurityService()

	data := []interface{}{
		map[string]interface{}{"agentID": "agent-alpha", "value": "alpha-1"},
		map[string]interface{}{"agentID": "agent-beta", "value": "beta-1"},
		map[string]interface{}{"agentID": "agent-gamma", "value": "gamma-1"},
		map[string]interface{}{"agentID": "agent-alpha", "value": "alpha-2"},
		map[string]interface{}{"agentID": "agent-beta", "value": "beta-2"},
	}

	resultAlpha := svc.NamespaceIsolate(data, "agent-alpha")
	resultBeta := svc.NamespaceIsolate(data, "agent-beta")
	resultGamma := svc.NamespaceIsolate(data, "agent-gamma")
	resultUnknown := svc.NamespaceIsolate(data, "agent-unknown")

	alphaSlice := resultAlpha.([]interface{})
	betaSlice := resultBeta.([]interface{})
	gammaSlice := resultGamma.([]interface{})
	unknownSlice := resultUnknown.([]interface{})

	if len(alphaSlice) != 2 {
		t.Errorf("Expected 2 items for agent-alpha, got %d", len(alphaSlice))
	}

	if len(betaSlice) != 2 {
		t.Errorf("Expected 2 items for agent-beta, got %d", len(betaSlice))
	}

	if len(gammaSlice) != 1 {
		t.Errorf("Expected 1 item for agent-gamma, got %d", len(gammaSlice))
	}

	if len(unknownSlice) != 0 {
		t.Errorf("Expected 0 items for unknown agent, got %d", len(unknownSlice))
	}

	if alphaSlice[0].(map[string]interface{})["value"] != "alpha-1" {
		t.Error("Expected alpha-1 for first alpha item")
	}

	if betaSlice[1].(map[string]interface{})["value"] != "beta-2" {
		t.Error("Expected beta-2 for second beta item")
	}
}

func TestSecurityService_CheckTaintPolicy_allowed(t *testing.T) {
	svc := NewSecurityService()

	result := svc.CheckTaintPolicy([]string{"INTERNAL"}, "read")

	if !result {
		t.Error("Expected CheckTaintPolicy to return true for allowed transition")
	}
}

func TestSecurityService_CheckTaintPolicy_denied(t *testing.T) {
	svc := NewSecurityService()

	result := svc.CheckTaintPolicy([]string{"UNTRUSTED"}, "write")

	if result {
		t.Error("Expected CheckTaintPolicy to return false for denied transition")
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
