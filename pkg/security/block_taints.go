package security

func ApplyBlockTaints(blocks []ContextBlock, boundary BoundaryType) ([]ContextBlock, error) {
	if len(blocks) == 0 {
		return []ContextBlock{}, nil
	}

	mergedTaints := make(TaintSet)
	for _, block := range blocks {
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
