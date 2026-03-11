package security

import (
	"testing"
	"time"

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
	outcome := svc.HandleMail(&mail)

	if outcome.Status != "success" {
		t.Errorf("Expected HandleMail to return success, got %s", outcome.Status)
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
	outcome := svc.HandleMail(&m)
	if outcome.Status != "success" {
		t.Errorf("HandleMail should return success for Phase 1 pass-through, got: %s", outcome.Status)
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

	hasOuterBoundaryTaint := false
	for _, taint := range result.Metadata.Taints {
		if taint == "OUTER_BOUNDARY" {
			hasOuterBoundaryTaint = true
			break
		}
	}

	if !hasOuterBoundaryTaint {
		t.Errorf("Expected OUTER_BOUNDARY taint to be added for outer→inner transition, got: %v", result.Metadata.Taints)
	}
}

func TestSecurityService_NamespaceIsolate(t *testing.T) {
	svc := NewSecurityService()

	view, err := svc.NamespaceIsolate("agent-1", "read")
	if err != nil {
		t.Errorf("Expected NamespaceIsolate to return nil error, got %v", err)
	}

	if view.RuntimeID != "agent-1" {
		t.Errorf("Expected RuntimeID to be agent-1, got %s", view.RuntimeID)
	}

	if view.Operation != "read" {
		t.Errorf("Expected Operation to be read, got %s", view.Operation)
	}
}

func TestSecurityService_NamespaceIsolate_multipleAgents(t *testing.T) {
	svc := NewSecurityService()

	view1, err1 := svc.NamespaceIsolate("agent-alpha", "read")
	view2, err2 := svc.NamespaceIsolate("agent-beta", "write")

	if err1 != nil {
		t.Errorf("Expected NamespaceIsolate to return nil error, got %v", err1)
	}

	if err2 != nil {
		t.Errorf("Expected NamespaceIsolate to return nil error, got %v", err2)
	}

	if view1.RuntimeID != "agent-alpha" {
		t.Errorf("Expected RuntimeID to be agent-alpha, got %s", view1.RuntimeID)
	}

	if view2.RuntimeID != "agent-beta" {
		t.Errorf("Expected RuntimeID to be agent-beta, got %s", view2.RuntimeID)
	}
}

func TestSecurityService_CheckTaintPolicy_allowed(t *testing.T) {
	svc := NewSecurityService()

	data := map[string]interface{}{
		"_taints": []string{"INTERNAL"},
	}

	policy := security.TaintPolicy{
		RedactMode: "audit",
		AllowedForBoundary: []security.BoundaryType{
			security.InnerBoundary,
			security.OuterBoundary,
		},
	}

	allowed, err := svc.CheckTaintPolicy(data, mail.OuterBoundary, policy)
	if err != nil {
		t.Errorf("Expected CheckTaintPolicy to return nil error, got %v", err)
	}

	if !allowed {
		t.Error("Expected CheckTaintPolicy to return true for allowed transition")
	}
}

func TestSecurityService_CheckTaintPolicy_denied(t *testing.T) {
	svc := NewSecurityService()

	data := map[string]interface{}{
		"_taints": []string{"INNER_ONLY"},
	}

	policy := security.TaintPolicy{
		RedactMode: "strict",
		AllowedForBoundary: []security.BoundaryType{
			security.InnerBoundary,
		},
	}

	allowed, err := svc.CheckTaintPolicy(data, mail.OuterBoundary, policy)
	if err != nil {
		t.Errorf("Expected CheckTaintPolicy to return nil error, got %v", err)
	}

	if allowed {
		t.Error("Expected CheckTaintPolicy to return false for denied transition")
	}
}

func TestSecurityService_NamespaceIsolate_FilteredView(t *testing.T) {
	svc := NewSecurityService()

	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"agentID": "agent-123", "value": "data-1"},
			map[string]interface{}{"agentID": "agent-456", "value": "data-2"},
			map[string]interface{}{"agentID": "agent-123", "value": "data-3"},
		},
	}

	view, err := svc.NamespaceIsolateWithFilter("agent-123", "read", data)

	if err != nil {
		t.Errorf("Expected NamespaceIsolate to return nil error, got %v", err)
	}

	if view.RuntimeID != "agent-123" {
		t.Errorf("Expected RuntimeID to be agent-123, got %s", view.RuntimeID)
	}

	if view.Operation != "read" {
		t.Errorf("Expected Operation to be read, got %s", view.Operation)
	}

	contextData, ok := view.ContextData["items"].([]interface{})
	if !ok {
		t.Error("Expected ContextData to contain items")
	}

	if len(contextData) != 2 {
		t.Errorf("Expected 2 items for agent-123, got %d", len(contextData))
	}

	firstItem, ok := contextData[0].(map[string]interface{})
	if !ok {
		t.Error("Expected first item to be map[string]interface{}")
	}

	if firstItem["value"] != "data-1" {
		t.Error("Expected first item to be data-1")
	}

	secondItem, ok := contextData[1].(map[string]interface{})
	if !ok {
		t.Error("Expected second item to be map[string]interface{}")
	}

	if secondItem["value"] != "data-3" {
		t.Error("Expected second item to be data-3")
	}
}

func TestSecurityService_CheckTaintPolicy_EnforcementModes(t *testing.T) {
	svc := NewSecurityService()

	dataWithTaints := map[string]interface{}{
		"name":    "test",
		"_taints": []string{"PII", "INNER_ONLY"},
	}

	strictPolicy := security.TaintPolicy{
		RedactMode: "strict",
		AllowedForBoundary: []security.BoundaryType{
			security.InnerBoundary,
		},
	}

	allowed, err := svc.CheckTaintPolicy(dataWithTaints, mail.OuterBoundary, strictPolicy)

	if err != nil {
		t.Errorf("Expected CheckTaintPolicy to return nil error, got %v", err)
	}

	if allowed {
		t.Error("Expected CheckTaintPolicy to return false for forbidden taints in strict mode")
	}

	auditPolicy := security.TaintPolicy{
		RedactMode: "audit",
		AllowedForBoundary: []security.BoundaryType{
			security.InnerBoundary,
			security.OuterBoundary,
		},
	}

	allowed2, err := svc.CheckTaintPolicy(dataWithTaints, mail.OuterBoundary, auditPolicy)

	if err != nil {
		t.Errorf("Expected CheckTaintPolicy to return nil error, got %v", err)
	}

	if !allowed2 {
		t.Error("Expected CheckTaintPolicy to return true for allowed taints in audit mode")
	}

	allowedData := map[string]interface{}{
		"name":    "test",
		"_taints": []string{"INTERNAL"},
	}

	outerAllowedPolicy := security.TaintPolicy{
		RedactMode: "strict",
		AllowedForBoundary: []security.BoundaryType{
			security.InnerBoundary,
			security.OuterBoundary,
		},
	}

	allowed3, err := svc.CheckTaintPolicy(allowedData, mail.OuterBoundary, outerAllowedPolicy)

	if err != nil {
		t.Errorf("Expected CheckTaintPolicy to return nil error, got %v", err)
	}

	if !allowed3 {
		t.Error("Expected CheckTaintPolicy to return true for allowed taints")
	}
}

func TestSecurityService_ReportTaints_CompleteMap(t *testing.T) {
	svc := NewSecurityService()

	_ = svc.TrackTaint("agent-456", "object-1", "PII")
	_ = svc.TrackTaint("agent-456", "object-1", "SECRET")
	_ = svc.TrackTaint("agent-456", "object-2", "EXTERNAL")
	_ = svc.TrackTaint("agent-789", "object-3", "INTERNAL")

	taintMap, err := svc.ReportTaints("agent-456")

	if err != nil {
		t.Errorf("Expected ReportTaints to return nil error, got %v", err)
	}

	if taintMap == nil {
		t.Error("Expected ReportTaints to return non-nil TaintMap")
	}

	object1Taints, ok := taintMap["object-1"]
	if !ok {
		t.Error("Expected object-1 to be in TaintMap")
	}

	hasPII := false
	hasSecret := false
	for _, t := range object1Taints {
		if t == "PII" {
			hasPII = true
		}
		if t == "SECRET" {
			hasSecret = true
		}
	}

	if !hasPII || !hasSecret {
		t.Error("Expected object-1 to have PII and SECRET taints")
	}

	object2Taints, ok := taintMap["object-2"]
	if !ok {
		t.Error("Expected object-2 to be in TaintMap")
	}

	if len(object2Taints) != 1 || object2Taints[0] != "EXTERNAL" {
		t.Error("Expected object-2 to have EXTERNAL taint")
	}

	unknownTaintMap, err := svc.ReportTaints("unknown-agent")

	if err != nil {
		t.Errorf("Expected ReportTaints to return nil error for unknown agent, got %v", err)
	}

	if len(unknownTaintMap) != 0 {
		t.Errorf("Expected empty TaintMap for unknown agent, got %d entries", len(unknownTaintMap))
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

	outcome := svc.HandleMail(&inputMail)

	if outcome.Status != "success" {
		t.Errorf("Expected HandleMail to return success, got %s", outcome.Status)
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

func TestSecurityService_TaintPropagate_NestedStructures(t *testing.T) {
	svc := NewSecurityService()

	inputObj := map[string]interface{}{
		"name": "test",
		"nested": map[string]interface{}{
			"value": "nested-value",
			"deep": map[string]interface{}{
				"data": "deep-data",
			},
		},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "value": "first"},
			map[string]interface{}{"id": 2, "value": "second"},
		},
	}

	result, err := svc.TaintPropagate(inputObj, []string{"PII", "SECRET"})

	if err != nil {
		t.Errorf("Expected TaintPropagate to return nil error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be map[string]interface{}")
	}

	rootTaints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Error("Expected _taints key to be added to root object")
	}

	hasPII := false
	hasSecret := false
	for _, t := range rootTaints {
		if t == "PII" {
			hasPII = true
		}
		if t == "SECRET" {
			hasSecret = true
		}
	}

	if !hasPII || !hasSecret {
		t.Error("Expected root to have PII and SECRET taints")
	}

	nested, ok := resultMap["nested"].(map[string]interface{})
	if !ok {
		t.Error("Expected nested to be map[string]interface{}")
	}

	nestedTaints, ok := nested["_taints"].([]string)
	if !ok {
		t.Error("Expected _taints key to be added to nested object")
	}

	nestedHasPII := false
	nestedHasSecret := false
	for _, t := range nestedTaints {
		if t == "PII" {
			nestedHasPII = true
		}
		if t == "SECRET" {
			nestedHasSecret = true
		}
	}

	if !nestedHasPII || !nestedHasSecret {
		t.Error("Expected nested to have PII and SECRET taints")
	}

	deep, ok := nested["deep"].(map[string]interface{})
	if !ok {
		t.Error("Expected deep to be map[string]interface{}")
	}

	deepTaints, ok := deep["_taints"].([]string)
	if !ok {
		t.Error("Expected _taints key to be added to deep object")
	}

	deepHasPII := false
	deepHasSecret := false
	for _, t := range deepTaints {
		if t == "PII" {
			deepHasPII = true
		}
		if t == "SECRET" {
			deepHasSecret = true
		}
	}

	if !deepHasPII || !deepHasSecret {
		t.Error("Expected deep to have PII and SECRET taints")
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

	firstTaints, ok := firstItem["_taints"].([]string)
	if !ok {
		t.Error("Expected _taints key to be added to first item")
	}

	firstHasPII := false
	firstHasSecret := false
	for _, t := range firstTaints {
		if t == "PII" {
			firstHasPII = true
		}
		if t == "SECRET" {
			firstHasSecret = true
		}
	}

	if !firstHasPII || !firstHasSecret {
		t.Error("Expected first item to have PII and SECRET taints")
	}

	secondItem, ok := items[1].(map[string]interface{})
	if !ok {
		t.Error("Expected second item to be map[string]interface{}")
	}

	secondTaints, ok := secondItem["_taints"].([]string)
	if !ok {
		t.Error("Expected _taints key to be added to second item")
	}

	secondHasPII := false
	secondHasSecret := false
	for _, t := range secondTaints {
		if t == "PII" {
			secondHasPII = true
		}
		if t == "SECRET" {
			secondHasSecret = true
		}
	}

	if !secondHasPII || !secondHasSecret {
		t.Error("Expected second item to have PII and SECRET taints")
	}

	if len(rootTaints) != 2 {
		t.Errorf("Expected no duplicates in root taints, got %d taints: %v", len(rootTaints), rootTaints)
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

func TestRuntimeTaintQuery_ReturnsActiveTaints(t *testing.T) {
	svc := NewSecurityService()

	_ = svc.TrackTaint("agent-789", "user-profile", "PII")
	_ = svc.TrackTaint("agent-789", "user-profile", "SECRET")
	_ = svc.TrackTaint("agent-789", "internal-config", "INNER_ONLY")
	_ = svc.TrackTaint("agent-789", "api-key", "SECRET")

	taintMap, err := svc.ReportTaints("agent-789")

	if err != nil {
		t.Errorf("Expected ReportTaints to return nil error, got %v", err)
	}

	if taintMap == nil {
		t.Error("Expected ReportTaints to return non-nil TaintMap")
	}

	userProfileTaints, ok := taintMap["user-profile"]
	if !ok {
		t.Error("Expected user-profile to be in TaintMap")
	}

	hasPII := false
	hasSecret := false
	for _, t := range userProfileTaints {
		if t == "PII" {
			hasPII = true
		}
		if t == "SECRET" {
			hasSecret = true
		}
	}

	if !hasPII {
		t.Error("Expected user-profile to have PII taint")
	}

	if !hasSecret {
		t.Error("Expected user-profile to have SECRET taint")
	}

	internalConfigTaints, ok := taintMap["internal-config"]
	if !ok {
		t.Error("Expected internal-config to be in TaintMap")
	}

	hasInnerOnly := false
	for _, t := range internalConfigTaints {
		if t == "INNER_ONLY" {
			hasInnerOnly = true
		}
	}

	if !hasInnerOnly {
		t.Error("Expected internal-config to have INNER_ONLY taint")
	}

	apiKeyTaints, ok := taintMap["api-key"]
	if !ok {
		t.Error("Expected api-key to be in TaintMap")
	}

	hasApiKeySecret := false
	for _, t := range apiKeyTaints {
		if t == "SECRET" {
			hasApiKeySecret = true
		}
	}

	if !hasApiKeySecret {
		t.Error("Expected api-key to have SECRET taint")
	}

	if len(taintMap) != 3 {
		t.Errorf("Expected 3 objects in TaintMap, got %d", len(taintMap))
	}
}

type mockPublisher struct {
	publishedMails []mail.Mail
}

func (m *mockPublisher) Publish(msg mail.Mail) (mail.Ack, error) {
	m.publishedMails = append(m.publishedMails, msg)
	return mail.Ack{DeliveredAt: time.Now()}, nil
}

func TestRuntimeTaintQuery_EnablesViolationDetection(t *testing.T) {
	svc := NewSecurityService()

	mockPub := &mockPublisher{publishedMails: []mail.Mail{}}
	svc.SetPublisher(mockPub)

	err := svc.EmitTaintViolation("agent-789", "user-profile", []string{"INNER_ONLY"}, "forbidden taint crossing boundary")

	if err != nil {
		t.Errorf("Expected EmitTaintViolation to return nil error, got %v", err)
	}

	if len(mockPub.publishedMails) != 1 {
		t.Errorf("Expected 1 mail to be published, got %d", len(mockPub.publishedMails))
	}

	publishedMail := mockPub.publishedMails[0]

	if publishedMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected mail type to be taint_violation, got %s", publishedMail.Type)
	}

	if publishedMail.Target != "sys:observability" {
		t.Errorf("Expected mail target to be sys:observability, got %s", publishedMail.Target)
	}

	if publishedMail.Source != "sys:security" {
		t.Errorf("Expected mail source to be sys:security, got %s", publishedMail.Source)
	}

	event, ok := publishedMail.Content.(TaintViolationEvent)
	if !ok {
		t.Error("Expected mail content to be TaintViolationEvent")
	}

	if event.RuntimeID != "agent-789" {
		t.Errorf("Expected RuntimeID to be agent-789, got %s", event.RuntimeID)
	}

	if event.ObjectID != "user-profile" {
		t.Errorf("Expected ObjectID to be user-profile, got %s", event.ObjectID)
	}

	hasInnerOnly := false
	for _, t := range event.Taints {
		if t == "INNER_ONLY" {
			hasInnerOnly = true
		}
	}

	if !hasInnerOnly {
		t.Error("Expected Taints to contain INNER_ONLY")
	}

	if event.Reason != "forbidden taint crossing boundary" {
		t.Errorf("Expected Reason to be 'forbidden taint crossing boundary', got %s", event.Reason)
	}
}

func TestHardcodedServices_SecurityBoundaryEnforcement(t *testing.T) {
	svc := NewSecurityService()

	outerMail := mail.Mail{
		ID:       "test-001",
		Source:   "agent:outer",
		Target:   "agent:inner",
		Content:  map[string]any{"data": "sensitive"},
		Metadata: mail.MailMetadata{Boundary: mail.OuterBoundary},
	}

	taintedMail, err := svc.ValidateAndSanitize(outerMail, mail.OuterBoundary, mail.InnerBoundary)
	if err != nil {
		t.Errorf("Expected no error for outer→inner, got %v", err)
	}
	if !containsTaint(taintedMail.Metadata.Taints, "OUTER_BOUNDARY") {
		t.Error("Expected OUTER_BOUNDARY taint added for outer→inner transition")
	}

	innerMail := mail.Mail{
		ID:      "test-002",
		Source:  "agent:inner",
		Target:  "agent:outer",
		Content: map[string]any{"secret": "password123"},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Taints:   []string{"SECRET", "PII"},
		},
	}

	_, err = svc.ValidateAndSanitize(innerMail, mail.InnerBoundary, mail.OuterBoundary)
	if err == nil {
		t.Error("Expected error for inner→outer with SECRET/PII taints")
	}
}

func containsTaint(taints []string, target string) bool {
	for _, t := range taints {
		if t == target {
			return true
		}
	}
	return false
}

func TestHardcodedServices_SecurityTaintPropagation(t *testing.T) {
	svc := NewSecurityService()

	nestedMap := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"value": "test",
			},
		},
		"direct": "value",
	}

	propagated, err := svc.TaintPropagate(nestedMap, []string{"USER_SUPPLIED"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	propagatedMap, ok := propagated.(map[string]any)
	if !ok {
		t.Fatal("Expected propagated to be map[string]any")
	}

	if propagatedMap["_taints"] == nil {
		t.Error("Expected _taints field at top level")
	}

	level1 := propagatedMap["level1"].(map[string]any)
	if level1["_taints"] == nil {
		t.Error("Expected _taints field at level1")
	}

	level2 := level1["level2"].(map[string]any)
	if level2["_taints"] == nil {
		t.Error("Expected _taints field at level2")
	}

	existingTainted := map[string]any{
		"_taints": []string{"EXISTING"},
		"data":    "value",
	}
	merged, err := svc.TaintPropagate(existingTainted, []string{"NEW"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	mergedMap, ok := merged.(map[string]any)
	if !ok {
		t.Fatal("Expected merged to be map[string]any")
	}
	taints := mergedMap["_taints"].([]string)
	if !containsTaint(taints, "EXISTING") || !containsTaint(taints, "NEW") {
		t.Error("Expected both EXISTING and NEW taints in merged result")
	}
}

func TestTaintPropagation_MailChainResponse(t *testing.T) {
	sourceMail := &mail.Mail{
		ID:     "mail-001",
		Source: "agent:user",
		Target: "agent:assistant",
		Content: map[string]any{
			"message": "What is the weather?",
		},
		Metadata: mail.MailMetadata{
			Taints:   []string{"USER_SUPPLIED"},
			Boundary: mail.DMZBoundary,
		},
	}

	responseMail := &mail.Mail{
		ID:     "mail-002",
		Source: "agent:assistant",
		Target: "agent:user",
		Content: map[string]any{
			"message": "The weather is sunny.",
		},
		Metadata: mail.MailMetadata{
			Taints:   []string{},
			Boundary: mail.DMZBoundary,
		},
	}

	mail.PropagateTaints(sourceMail, responseMail)

	if !containsTaint(responseMail.GetTaints(), "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint to propagate from source to response mail")
	}
}

func TestTaintPropagation_ToolCallResult(t *testing.T) {
	toolCallMail := &mail.Mail{
		ID:     "tool-call-001",
		Source: "agent:assistant",
		Target: "sys:tools",
		Type:   mail.MailTypeToolCall,
		Content: map[string]any{
			"tool": "get_weather",
			"args": map[string]any{"location": "NYC"},
		},
		Metadata: mail.MailMetadata{
			Taints:   []string{"USER_SUPPLIED"},
			Boundary: mail.DMZBoundary,
		},
	}

	toolResultMail := &mail.Mail{
		ID:     "tool-result-001",
		Source: "sys:tools",
		Target: "agent:assistant",
		Type:   mail.MailTypeToolResult,
		Content: map[string]any{
			"result": "Sunny, 72F",
		},
		Metadata: mail.MailMetadata{
			Taints:   []string{},
			Boundary: mail.DMZBoundary,
		},
	}

	mail.PropagateTaints(toolCallMail, toolResultMail)

	if !containsTaint(toolResultMail.GetTaints(), "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint to propagate through tool call to result")
	}
}

func TestTaintPropagation_BoundaryStrip(t *testing.T) {
	svc := NewSecurityService()

	innerMail := mail.Mail{
		ID:     "inner-001",
		Source: "agent:inner",
		Target: "agent:outer",
		Content: map[string]any{
			"data": "sensitive",
		},
		Metadata: mail.MailMetadata{
			Taints:   []string{"SECRET", "PII", "INTERNAL"},
			Boundary: mail.InnerBoundary,
		},
	}

	_, err := svc.ValidateAndSanitize(innerMail, mail.InnerBoundary, mail.OuterBoundary)

	if err == nil {
		t.Error("Expected error when SECRET/PII taints cross to outer boundary")
	}
}

func TestTaintPropagation_ExternalAdd(t *testing.T) {
	svc := NewSecurityService()

	outerMail := mail.Mail{
		ID:     "outer-001",
		Source: "agent:external",
		Target: "agent:inner",
		Content: map[string]any{
			"data": "external",
		},
		Metadata: mail.MailMetadata{
			Taints:   []string{"USER_SUPPLIED"},
			Boundary: mail.OuterBoundary,
		},
	}

	sanitized, err := svc.ValidateAndSanitize(outerMail, mail.OuterBoundary, mail.InnerBoundary)

	if err != nil {
		t.Errorf("Expected no error for outer to inner transition, got %v", err)
	}

	hasExternal := false
	hasOuterBoundary := false
	for _, taint := range sanitized.Metadata.Taints {
		if taint == "EXTERNAL" {
			hasExternal = true
		}
		if taint == "OUTER_BOUNDARY" {
			hasOuterBoundary = true
		}
	}

	if !hasExternal {
		t.Error("Expected EXTERNAL taint to be added for outbound to external systems")
	}
	if !hasOuterBoundary {
		t.Error("Expected OUTER_BOUNDARY taint to be added for outer to inner transition")
	}
}

func TestTaintPropagation_Deduplication(t *testing.T) {
	sourceMail := &mail.Mail{
		ID:     "mail-001",
		Source: "agent:source",
		Target: "agent:target",
		Content: map[string]any{
			"data": "source",
		},
		Taints: []string{"USER_SUPPLIED", "EXTERNAL"},
	}

	targetMail := &mail.Mail{
		ID:     "mail-002",
		Source: "agent:target",
		Target: "agent:response",
		Content: map[string]any{
			"data": "target",
		},
		Taints: []string{"EXTERNAL", "INTERNAL"},
	}

	mail.PropagateTaints(sourceMail, targetMail)

	taints := targetMail.GetTaints()

	hasUserSupplied := false
	hasExternal := false
	hasInternal := false
	externalCount := 0
	for _, t := range taints {
		if t == "USER_SUPPLIED" {
			hasUserSupplied = true
		}
		if t == "EXTERNAL" {
			hasExternal = true
			externalCount++
		}
		if t == "INTERNAL" {
			hasInternal = true
		}
	}

	if !hasUserSupplied {
		t.Error("Expected USER_SUPPLIED taint from source")
	}
	if !hasExternal {
		t.Error("Expected EXTERNAL taint")
	}
	if !hasInternal {
		t.Error("Expected INTERNAL taint from target")
	}
	if externalCount > 1 {
		t.Errorf("Expected EXTERNAL taint to be deduplicated, found %d times", externalCount)
	}
}
