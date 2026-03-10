package security

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestContextBlock_Strict_BLOCKS_VIOLATIONS(t *testing.T) {
	// Given: ContextBlock marked with taints=["SECRET"] and taintPolicy.enforcement=strict
	// Boundary does not include SECRET in allowedOnExit list
	block := ContextBlock{
		Name:    "secret-block",
		Source:  "session",
		Content: "Secret API key: sk-12345",
		Taints:  TaintSet{"SECRET": true},
		TaintPolicy: TaintPolicy{
			RedactMode: "strict",
		},
	}

	globalPolicy := TaintPolicyConfig{
		Enforcement:   EnforcementStrict,
		AllowedOnExit: []string{"TOOL_OUTPUT", "USER_SUPPLIED"},
	}

	// When: prepareContextForBoundary is called (FilterContextBlock simulates this)
	filtered, err := FilterContextBlockWithGlobalPolicy(block, OuterBoundary, globalPolicy)

	// Then: operation returns error
	if err == nil {
		t.Fatal("Expected error when strict enforcement blocks forbidden taint, but got nil")
	}

	// Then: Error message indicates strict enforcement blocked block with forbidden taint: SECRET
	expectedErrorMsg := "strict enforcement blocked block with forbidden taint: SECRET"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got: %s", expectedErrorMsg, err.Error())
	}

	// Then: Block is excluded from LLM prompt assembly (empty block returned)
	if filtered.Name != "" {
		t.Errorf("Expected block to be excluded from prompt assembly (empty name), got: %s", filtered.Name)
	}
	if filtered.Content != "" {
		t.Errorf("Expected block content to be excluded (empty), got: %s", filtered.Content)
	}
}

func TestContextBlock_Strict_EMITS_VIOLATION_EVENT(t *testing.T) {
	// Given: ContextBlock with taintPolicy.enforcement=strict violates boundary policy
	block := ContextBlock{
		Name:    "blocked-block",
		Source:  "memory",
		Content: "Blocked secret data",
		Taints:  TaintSet{"SECRET": true},
		TaintPolicy: TaintPolicy{
			RedactMode: "strict",
		},
	}

	globalPolicy := TaintPolicyConfig{
		Enforcement:   EnforcementStrict,
		AllowedOnExit: []string{"TOOL_OUTPUT"},
	}

	// Setup mail router for violation event routing
	router := mail.NewMailRouter()
	observabilityInbox := &mail.ServiceInbox{ID: "observability"}
	router.SubscribeService("observability", observabilityInbox)
	SetViolationRouter(router)

	// When: prepareContextForBoundary blocks the violation (FilterContextBlock simulates this)
	_, err := FilterContextBlockWithGlobalPolicy(block, OuterBoundary, globalPolicy)

	// Then: prepareContextForBoundary blocks the violation (returns error)
	if err == nil {
		t.Fatal("Expected error when strict enforcement blocks violation, but got nil")
	}

	// Then: TaintViolation event emitted with: blockID, taints, boundary, policy, timestamp
	// Verify violation was routed to dead-letter queue (observability inbox)
	receivedMail, err := observabilityInbox.Pop()
	if err != nil {
		t.Fatalf("Expected to receive TaintViolation event in observability inbox, got error: %v", err)
	}

	// Verify event type
	if receivedMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected Mail type to be MailTypeTaintViolation, got %v", receivedMail.Type)
	}

	// Verify event source
	if receivedMail.Source != "sys:security" {
		t.Errorf("Expected source to be 'sys:security', got %v", receivedMail.Source)
	}

	// Verify event target (dead-letter queue)
	if receivedMail.Target != "sys:observability" {
		t.Errorf("Expected target to be 'sys:observability', got %v", receivedMail.Target)
	}

	// Verify event content contains required fields
	content, ok := receivedMail.Content.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected content to be map[string]interface{}, got %T", receivedMail.Content)
	}

	// Verify blockID is present
	if content["blockID"] != "blocked-block" {
		t.Errorf("Expected blockID to be 'blocked-block', got %v", content["blockID"])
	}

	// Verify taints are present
	forbiddenTaints, ok := content["taints"].([]interface{})
	if !ok {
		t.Fatalf("Expected taints to be []interface{}, got %T", content["taints"])
	}
	if len(forbiddenTaints) != 1 || forbiddenTaints[0] != "SECRET" {
		t.Errorf("Expected taints to contain 'SECRET', got %v", forbiddenTaints)
	}

	// Verify boundary is present
	if content["boundary"] != "outer" {
		t.Errorf("Expected boundary to be 'outer', got %v", content["boundary"])
	}

	// Verify policy is present
	if content["policy"] != "strict" {
		t.Errorf("Expected policy to be 'strict', got %v", content["policy"])
	}

	// Verify timestamp is present
	if _, ok := content["timestamp"]; !ok {
		t.Error("Expected timestamp to be present in content")
	}

	// Then: Event routed to dead-letter queue for later analysis
	// (Verified by: mail received in observability inbox)

	// Then: Chart execution may continue but without the blocked block in context
	// (Verified by: error returned but no panic, block excluded from context)
	_ = time.Now() // Ensure time package is used
}
