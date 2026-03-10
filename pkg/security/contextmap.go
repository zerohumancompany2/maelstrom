package security

import "fmt"

type ContextMap struct {
	Blocks     []*ContextBlock
	TokenCount int
	Budget     int
}

func NewContextMap(blocks []*ContextBlock, budget int) *ContextMap {
	return &ContextMap{
		Blocks: blocks,
		Budget: budget,
	}
}

func (cm *ContextMap) AssembleWithBudget() ([]*ContextBlock, error) {
	result := make([]*ContextBlock, 0, len(cm.Blocks))
	totalTokens := 0

	for _, block := range cm.Blocks {
		totalTokens += block.MaxTokens
	}

	for totalTokens > cm.Budget && len(result) < len(cm.Blocks) {
		evicted, err := cm.EvictLowestPriority()
		if err != nil {
			return nil, err
		}
		if evicted != nil {
			totalTokens -= evicted.MaxTokens
		} else {
			break
		}
	}

	for _, block := range cm.Blocks {
		if !isEvicted(block, cm.Blocks) {
			result = append(result, block)
		}
	}

	return result, nil
}

func isEvicted(block *ContextBlock, allBlocks []*ContextBlock) bool {
	for _, b := range allBlocks {
		if b.Name == block.Name && b.MaxTokens == 0 {
			return true
		}
	}
	return false
}

func (cm *ContextMap) EvictLowestPriority() (*ContextBlock, error) {
	if len(cm.Blocks) == 0 {
		return nil, nil
	}

	lowestIdx := 0
	n := len(cm.Blocks)
	for i := 1; i < n; i++ {
		if cm.Blocks[i].Priority > cm.Blocks[lowestIdx].Priority {
			lowestIdx = i
		}
	}

	if cm.Blocks[lowestIdx].Priority == 0 {
		return nil, nil
	}

	evicted := cm.Blocks[lowestIdx]
	cm.Blocks = append(cm.Blocks[:lowestIdx], cm.Blocks[lowestIdx+1:]...)
	return evicted, nil
}

func (cm *ContextMap) CompressBlock(block *ContextBlock) error {
	panic("not implemented")
}

func (cm *ContextMap) IsSystemBlock(block *ContextBlock) bool {
	panic("not implemented")
}

type BlockTaintInfo struct {
	Block  *ContextBlock
	Taints []string
}

var contextBlockRegistry = make(map[string]BlockTaintInfo)

func PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error {
	filtered := make([]*ContextBlock, 0)
	for name, info := range contextBlockRegistry {
		isForbidden := false
		forbiddenTaint := ""
		for _, taint := range info.Taints {
			if taint == "INNER_ONLY" && (boundary == DMZBoundary || boundary == OuterBoundary) {
				isForbidden = true
				forbiddenTaint = taint
				break
			}
			if taint == "SECRET" && (boundary == DMZBoundary || boundary == OuterBoundary) {
				isForbidden = true
				forbiddenTaint = taint
				break
			}
		}
		if isForbidden {
			auditLog = append(auditLog, "DROPPED: forbidden taint "+forbiddenTaint+" for "+string(boundary)+" boundary - block "+name)
			continue
		}
		filtered = append(filtered, info.Block)
	}
	contextBlockRegistry = make(map[string]BlockTaintInfo)
	for _, block := range filtered {
		processedBlock := applyTaintPolicy(block, boundary)
		contextBlockRegistry[block.Name] = BlockTaintInfo{Block: processedBlock, Taints: []string{"TOOL_OUTPUT"}}
	}
	return nil
}

func applyTaintPolicy(block *ContextBlock, boundary BoundaryType) *ContextBlock {
	if block.TaintPolicy.RedactMode == "redact" && len(block.TaintPolicy.RedactRules) > 0 {
		result := *block
		content := result.Content
		for _, rule := range block.TaintPolicy.RedactRules {
			content = applyRedactionRule(content, rule)
		}
		result.Content = content
		auditLog = append(auditLog, "REDACTION: block "+block.Name+" redacted for boundary "+string(boundary))
		return &result
	}
	return block
}

func FilterContextBlock(block ContextBlock, boundary BoundaryType) (ContextBlock, error) {
	if isBoundaryAllowed(block.TaintPolicy.AllowedForBoundary, boundary) {
		if block.TaintPolicy.RedactMode == "dropBlock" {
			return ContextBlock{}, nil
		}
		if block.TaintPolicy.RedactMode == "redact" {
			result := block
			content := result.Content
			for _, rule := range block.TaintPolicy.RedactRules {
				content = replaceTaint(content, rule.Taint, rule.Replacement)
			}
			result.Content = content
			return result, nil
		}
		if block.TaintPolicy.RedactMode == "audit" {
			logAuditViolation(block, boundary)
			return block, nil
		}
		return block, nil
	}
	return ContextBlock{}, nil
}

func logAuditViolation(block ContextBlock, boundary BoundaryType) {
	auditLog = append(auditLog, "VIOLATION at "+string(boundary)+": block "+block.Name+" with sensitive data")
}

func isBoundaryAllowed(allowed []BoundaryType, boundary BoundaryType) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, b := range allowed {
		if b == boundary {
			return true
		}
	}
	return false
}

func FilterContextBlockWithGlobalPolicy(block ContextBlock, boundary BoundaryType, globalPolicy TaintPolicyConfig) (ContextBlock, error) {
	if block.TaintPolicy.RedactMode == "strict" || globalPolicy.Enforcement == EnforcementStrict {
		return filterWithStrictEnforcement(block, boundary, globalPolicy)
	}
	if block.TaintPolicy.RedactMode != "" {
		return FilterContextBlock(block, boundary)
	}
	return block, nil
}

func filterWithStrictEnforcement(block ContextBlock, boundary BoundaryType, globalPolicy TaintPolicyConfig) (ContextBlock, error) {
	allowedSet := make(map[string]bool)
	for _, t := range globalPolicy.AllowedOnExit {
		allowedSet[t] = true
	}

	forbiddenTaints := getBlockForbiddenTaints(block, allowedSet)
	if len(forbiddenTaints) == 0 {
		return block, nil
	}

	forbiddenTaint := forbiddenTaints[0]
	err := fmt.Errorf("strict enforcement blocked block with forbidden taint: %s", forbiddenTaint)

	ReportContextBlockViolation(block.Name, forbiddenTaints, boundary, "strict")

	return ContextBlock{}, err
}

func getBlockForbiddenTaints(block ContextBlock, allowedSet map[string]bool) []string {
	forbidden := make([]string, 0)
	for taint := range block.Taints {
		if !allowedSet[taint] {
			forbidden = append(forbidden, taint)
		}
	}
	return forbidden
}

func replaceTaint(content, taint, replacement string) string {
	result := ""
	for i := 0; i < len(content); i++ {
		if i+len(taint) <= len(content) && content[i:i+len(taint)] == taint {
			result += replacement
			i += len(taint) - 1
		} else {
			result += string(content[i])
		}
	}
	return result
}

func applyRedactionRule(content string, rule RedactRule) string {
	result := content
	if rule.Taint == "PII" {
		redacted := replacePII(result, rule.Replacement)
		if redacted != result {
			return redacted
		}
		result = replaceTaint(result, "PII", rule.Replacement)
	} else {
		result = replaceTaint(result, rule.Taint, rule.Replacement)
	}
	return result
}

func replacePII(content, replacement string) string {
	result := content
	for i := 0; i < len(result); i++ {
		if result[i] == '@' {
			start := i - 1
			for start >= 0 && (result[start] >= 'a' && result[start] <= 'z' || result[start] >= 'A' && result[start] <= 'Z' || result[start] >= '0' && result[start] <= '9' || result[start] == '.' || result[start] == '_' || result[start] == '-' || result[start] == '+') {
				start--
			}
			start++
			end := i + 1
			for end < len(result) && (result[end] >= 'a' && result[end] <= 'z' || result[end] >= 'A' && result[end] <= 'Z' || result[end] >= '0' && result[end] <= '9' || result[end] == '.' || result[end] == '-') {
				end++
			}
			result = result[:start] + replacement + result[end:]
			break
		}
	}
	return result
}
