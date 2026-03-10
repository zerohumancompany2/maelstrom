package security

func NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error) {
	return IsolatedView{
		RuntimeID:   runtimeId,
		Operation:   operation,
		Boundary:    DMZBoundary,
		ContextData: make(map[string]any),
	}, nil
}

func (iv *IsolatedView) FilterData(data any) any {
	switch v := data.(type) {
	case map[string]interface{}:
		forbidden := getForbiddenTaintsForBoundary(iv.Boundary)
		forbiddenSet := make(map[string]bool)
		for _, f := range forbidden {
			forbiddenSet[f] = true
		}

		result := make(map[string]interface{})
		taints, ok := v["_taints"].([]string)
		if ok {
			filteredTaints := make([]string, 0)
			for _, t := range taints {
				if !forbiddenSet[t] {
					filteredTaints = append(filteredTaints, t)
				}
			}
			if len(filteredTaints) > 0 {
				result["_taints"] = filteredTaints
			}
		}

		for k, val := range v {
			if k == "_taints" {
				continue
			}
			result[k] = val
		}

		return result
	default:
		return data
	}
}

func getForbiddenTaintsForBoundary(boundary BoundaryType) []string {
	switch boundary {
	case InnerBoundary:
		return nil
	case DMZBoundary:
		return []string{"INNER_ONLY", "SECRET"}
	case OuterBoundary:
		return []string{"INNER_ONLY", "SECRET", "PII", "TOOL_OUTPUT"}
	default:
		return nil
	}
}

func (iv *IsolatedView) GetOperation() string {
	return iv.Operation
}

func (iv *IsolatedView) IsReadOperation() bool {
	return iv.Operation == "read" || iv.Operation == "list"
}

func (iv *IsolatedView) IsWriteOperation() bool {
	return iv.Operation == "write"
}
