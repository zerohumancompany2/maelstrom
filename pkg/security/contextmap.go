package security

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
	for _, info := range contextBlockRegistry {
		isForbidden := false
		for _, taint := range info.Taints {
			if taint == "INNER_ONLY" && (boundary == DMZBoundary || boundary == OuterBoundary) {
				isForbidden = true
				break
			}
			if taint == "SECRET" && (boundary == DMZBoundary || boundary == OuterBoundary) {
				isForbidden = true
				break
			}
		}
		if !isForbidden {
			filtered = append(filtered, info.Block)
		}
	}
	contextBlockRegistry = make(map[string]BlockTaintInfo)
	for _, block := range filtered {
		contextBlockRegistry[block.Name] = BlockTaintInfo{Block: block, Taints: []string{"TOOL_OUTPUT"}}
	}
	return nil
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
	if block.TaintPolicy.RedactMode != "" {
		return FilterContextBlock(block, boundary)
	}
	return block, nil
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
