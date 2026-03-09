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
	return block, nil
}
