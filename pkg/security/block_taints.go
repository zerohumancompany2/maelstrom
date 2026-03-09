package security

func ApplyBlockTaints(blocks []ContextBlock, boundary BoundaryType) ([]ContextBlock, error) {
	if len(blocks) == 0 {
		return []ContextBlock{}, nil
	}

	engine := NewTaintEngine()

	var filteredBlocks []ContextBlock
	for _, block := range blocks {
		taintList := make([]string, 0, len(block.Taints))
		for taint := range block.Taints {
			taintList = append(taintList, taint)
		}

		err := engine.CheckForbidden(taintList, boundary)
		if err != nil {
			continue
		}

		filteredBlocks = append(filteredBlocks, block)
	}

	if len(filteredBlocks) == 0 {
		return []ContextBlock{}, nil
	}

	if len(filteredBlocks) == 1 {
		return filteredBlocks, nil
	}

	mergedTaints := make(TaintSet)
	for _, block := range filteredBlocks {
		for taint := range block.Taints {
			mergedTaints[taint] = true
		}
	}

	result := ContextBlock{
		Name:    "merged",
		Source:  "merged",
		Content: "",
		Taints:  mergedTaints,
	}

	return []ContextBlock{result}, nil
}
