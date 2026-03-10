package security

import (
	"strings"
	"testing"
)

func TestPrepareContextForBoundary_Filter(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	innerBlock := &ContextBlock{
		Name: "inner-block",
		TaintPolicy: TaintPolicy{
			RedactMode: "strict",
		},
	}
	toolBlock := &ContextBlock{
		Name: "tool-block",
		TaintPolicy: TaintPolicy{
			RedactMode: "strict",
		},
	}

	contextBlockRegistry["inner-block"] = BlockTaintInfo{
		Block:  innerBlock,
		Taints: []string{"INNER_ONLY"},
	}
	contextBlockRegistry["tool-block"] = BlockTaintInfo{
		Block:  toolBlock,
		Taints: []string{"TOOL_OUTPUT"},
	}

	err := PrepareContextForBoundary("runtime-1", DMZBoundary)

	if err != nil {
		t.Fatalf("PrepareContextForBoundary returned error: %v", err)
	}

	if _, exists := contextBlockRegistry["inner-block"]; exists {
		t.Errorf("Expected INNER_ONLY blocks to be filtered out, but inner-block still exists")
	}
	if _, exists := contextBlockRegistry["tool-block"]; !exists {
		t.Errorf("Expected TOOL_OUTPUT block to be preserved, but tool-block was removed")
	}
}

func TestFilterContextBlock_Redact(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "secret-block",
		Content: "This contains SECRET data",
		TaintPolicy: TaintPolicy{
			RedactMode: "redact",
			RedactRules: []RedactRule{
				{Taint: "SECRET", Replacement: "[REDACTED]"},
			},
		},
	}

	filtered, err := FilterContextBlock(block, OuterBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Content != "This contains [REDACTED] data" {
		t.Errorf("Expected SECRET to be redacted, got: %s", filtered.Content)
	}
	if filtered.Name != "secret-block" {
		t.Errorf("Expected block structure to be preserved, got name: %s", filtered.Name)
	}
}

func TestFilterContextBlock_DropBlock(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "pii-block",
		Content: "This contains PII data",
		TaintPolicy: TaintPolicy{
			RedactMode: "dropBlock",
		},
	}

	filtered, err := FilterContextBlock(block, OuterBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Name != "" {
		t.Errorf("Expected block to be dropped (empty name), got: %s", filtered.Name)
	}
	if filtered.Content != "" {
		t.Errorf("Expected block content to be empty after drop, got: %s", filtered.Content)
	}
}

func TestFilterContextBlock_PerBlockOverride(t *testing.T) {
	ClearAuditLog()
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "audit-block",
		Content: "This contains SECRET data",
		TaintPolicy: TaintPolicy{
			RedactMode: "audit",
		},
	}

	globalPolicy := TaintPolicyConfig{
		Enforcement: EnforcementStrict,
	}

	filtered, err := FilterContextBlockWithGlobalPolicy(block, DMZBoundary, globalPolicy)

	if err != nil {
		t.Fatalf("FilterContextBlockWithGlobalPolicy returned error (per-block audit should override global strict): %v", err)
	}

	if filtered.Name != "audit-block" {
		t.Errorf("Expected per-block audit mode to allow block through, got name: %s", filtered.Name)
	}
	if filtered.Content != "This contains SECRET data" {
		t.Errorf("Expected block content to be preserved with audit mode, got: %s", filtered.Content)
	}
}

func TestFilterContextBlock_AllowedForBoundary(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "inner-only-block",
		Content: "This contains SECRET data",
		TaintPolicy: TaintPolicy{
			AllowedForBoundary: []BoundaryType{InnerBoundary},
		},
	}

	filtered, err := FilterContextBlock(block, DMZBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Name != "" {
		t.Errorf("Expected block to be filtered out (dmz not in allowedForBoundary), got name: %s", filtered.Name)
	}
}

func TestFilterContextBlock_Audit(t *testing.T) {
	ClearAuditLog()
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "pii-block",
		Content: "This contains PII data",
		TaintPolicy: TaintPolicy{
			RedactMode: "audit",
		},
	}

	filtered, err := FilterContextBlock(block, OuterBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Name != "pii-block" {
		t.Errorf("Expected block to pass through unchanged with audit mode, got name: %s", filtered.Name)
	}
	if filtered.Content != "This contains PII data" {
		t.Errorf("Expected block content to be unchanged with audit mode, got: %s", filtered.Content)
	}

	lastLog := GetLastAuditLog()
	if lastLog == "" {
		t.Errorf("Expected violation to be logged to audit trail, but log is empty")
	}
	if !strings.Contains(lastLog, "VIOLATION") {
		t.Errorf("Expected audit log to contain VIOLATION, got: %s", lastLog)
	}
}

func TestContextMap_TokenBudgetEnforcement(t *testing.T) {
	block1 := &ContextBlock{
		Name:      "block1",
		Content:   strings.Repeat("a", 2000),
		MaxTokens: 2000,
		Priority:  1,
		Eviction:  "summarize",
	}
	block2 := &ContextBlock{
		Name:      "block2",
		Content:   strings.Repeat("b", 1500),
		MaxTokens: 1500,
		Priority:  5,
		Eviction:  "summarize",
	}
	block3 := &ContextBlock{
		Name:      "block3",
		Content:   strings.Repeat("c", 1500),
		MaxTokens: 1500,
		Priority:  10,
		Eviction:  "summarize",
	}

	blocks := []*ContextBlock{block1, block2, block3}
	totalTokens := 5000
	budget := 4000

	cm := NewContextMap(blocks, budget)
	result, err := cm.AssembleWithBudget()

	if err != nil {
		t.Fatalf("AssembleWithBudget returned error: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected some blocks to be returned")
	}

	total := 0
	for _, block := range result {
		total += block.MaxTokens
	}

	if total > budget {
		t.Errorf("Expected total tokens <= %d, got %d (total was %d before assembly)", budget, total, totalTokens)
	}
}

func TestContextMap_EvictionPriorityOrder(t *testing.T) {
	block1 := &ContextBlock{
		Name:      "block1",
		Content:   strings.Repeat("a", 1000),
		MaxTokens: 1000,
		Priority:  1,
	}
	block5 := &ContextBlock{
		Name:      "block5",
		Content:   strings.Repeat("b", 1000),
		MaxTokens: 1000,
		Priority:  5,
	}
	block10 := &ContextBlock{
		Name:      "block10",
		Content:   strings.Repeat("c", 1000),
		MaxTokens: 1000,
		Priority:  10,
	}
	block15 := &ContextBlock{
		Name:      "block15",
		Content:   strings.Repeat("d", 1000),
		MaxTokens: 1000,
		Priority:  15,
	}

	blocks := []*ContextBlock{block1, block5, block10, block15}
	budget := 2000

	cm := NewContextMap(blocks, budget)
	result, err := cm.AssembleWithBudget()

	if err != nil {
		t.Fatalf("AssembleWithBudget returned error: %v", err)
	}

	resultNames := make(map[string]bool)
	for _, block := range result {
		resultNames[block.Name] = true
	}

	if !resultNames["block1"] {
		t.Error("Expected block1 (priority 1) to be preserved")
	}
	if !resultNames["block5"] {
		t.Error("Expected block5 (priority 5) to be preserved")
	}
	if resultNames["block10"] {
		t.Error("Expected block10 (priority 10) to be evicted")
	}
	if resultNames["block15"] {
		t.Error("Expected block15 (priority 15, lowest) to be evicted first")
	}
}

func TestContextMap_SystemBlockPreservation(t *testing.T) {
	systemBlock := &ContextBlock{
		Name:      "system",
		Content:   strings.Repeat("s", 500),
		MaxTokens: 500,
		Priority:  0,
	}
	userBlock5 := &ContextBlock{
		Name:      "user5",
		Content:   strings.Repeat("a", 1000),
		MaxTokens: 1000,
		Priority:  5,
	}
	userBlock10 := &ContextBlock{
		Name:      "user10",
		Content:   strings.Repeat("b", 1000),
		MaxTokens: 1000,
		Priority:  10,
	}

	blocks := []*ContextBlock{systemBlock, userBlock5, userBlock10}
	budget := 600

	cm := NewContextMap(blocks, budget)
	result, err := cm.AssembleWithBudget()

	if err != nil {
		t.Fatalf("AssembleWithBudget returned error: %v", err)
	}

	resultNames := make(map[string]bool)
	for _, block := range result {
		resultNames[block.Name] = true
	}

	if !resultNames["system"] {
		t.Error("Expected system block (priority 0) to always be preserved")
	}
	if resultNames["user5"] {
		t.Error("Expected user block at priority 5 to be evicted")
	}
	if resultNames["user10"] {
		t.Error("Expected user block at priority 10 to be evicted")
	}

	total := 0
	for _, block := range result {
		total += block.MaxTokens
	}
	if total > budget {
		t.Errorf("Expected total tokens <= %d, got %d", budget, total)
	}
}

func TestContextMap_EvictionSummarize(t *testing.T) {
	block1 := &ContextBlock{
		Name:      "system",
		Content:   strings.Repeat("s", 500),
		MaxTokens: 500,
		Priority:  0,
		Eviction:  "summarize",
	}
	block5 := &ContextBlock{
		Name:      "user",
		Content:   strings.Repeat("u", 2000),
		MaxTokens: 2000,
		Priority:  5,
		Eviction:  "summarize",
	}

	blocks := []*ContextBlock{block1, block5}
	budget := 800

	cm := NewContextMap(blocks, budget)
	result, err := cm.AssembleWithBudget()

	if err != nil {
		t.Fatalf("AssembleWithBudget returned error: %v", err)
	}

	resultNames := make(map[string]bool)
	for _, block := range result {
		resultNames[block.Name] = true
	}

	if !resultNames["system"] {
		t.Error("Expected system block to be preserved")
	}

	total := 0
	for _, block := range result {
		total += block.MaxTokens
	}

	if total > budget {
		t.Errorf("Expected total tokens <= %d after compression/eviction, got %d", budget, total)
	}
}

func TestPrepareContextForBoundary_StripsForbidden(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	innerBlock := &ContextBlock{
		Name:    "inner-block",
		Content: "inner content",
	}
	secretBlock := &ContextBlock{
		Name:    "secret-block",
		Content: "secret content",
	}
	toolBlock := &ContextBlock{
		Name:    "tool-block",
		Content: "tool content",
	}

	contextBlockRegistry["inner-block"] = BlockTaintInfo{
		Block:  innerBlock,
		Taints: []string{"INNER_ONLY"},
	}
	contextBlockRegistry["secret-block"] = BlockTaintInfo{
		Block:  secretBlock,
		Taints: []string{"SECRET"},
	}
	contextBlockRegistry["tool-block"] = BlockTaintInfo{
		Block:  toolBlock,
		Taints: []string{"TOOL_OUTPUT"},
	}

	err := PrepareContextForBoundary("runtime-1", DMZBoundary)

	if err != nil {
		t.Fatalf("PrepareContextForBoundary returned error: %v", err)
	}

	if _, exists := contextBlockRegistry["inner-block"]; exists {
		t.Errorf("Expected INNER_ONLY blocks to be stripped, but inner-block still exists")
	}
	if _, exists := contextBlockRegistry["secret-block"]; exists {
		t.Errorf("Expected SECRET blocks to be stripped, but secret-block still exists")
	}
	if _, exists := contextBlockRegistry["tool-block"]; !exists {
		t.Errorf("Expected TOOL_OUTPUT block to be preserved, but tool-block was removed")
	}
}

func TestPrepareContextForBoundary_EnforcesPolicy(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	piiBlock := &ContextBlock{
		Name:    "pii-block",
		Content: "This contains PII data",
		TaintPolicy: TaintPolicy{
			RedactMode: "redact",
			RedactRules: []RedactRule{
				{Taint: "PII", Replacement: "[REDACTED]"},
			},
		},
	}

	contextBlockRegistry["pii-block"] = BlockTaintInfo{
		Block:  piiBlock,
		Taints: []string{"PII"},
	}

	err := PrepareContextForBoundary("runtime-1", OuterBoundary)

	if err != nil {
		t.Fatalf("PrepareContextForBoundary returned error: %v", err)
	}

	info, exists := contextBlockRegistry["pii-block"]
	if !exists {
		t.Errorf("Expected PII block to remain in context after redaction")
	}

	if info.Block.Content != "This contains [REDACTED] data" {
		t.Errorf("Expected PII-tainted values to be redacted, got: %s", info.Block.Content)
	}
}
