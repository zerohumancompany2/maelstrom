package security

type ContextMap struct {
	Blocks []*ContextBlock
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
	if block.TaintPolicy.RedactMode == "redact" {
		result := block
		content := result.Content
		for _, rule := range block.TaintPolicy.RedactRules {
			content = replaceTaint(content, rule.Taint, rule.Replacement)
		}
		result.Content = content
		return result, nil
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
