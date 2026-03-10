package devops

import (
	"errors"
)

type IsolationLevel string

const (
	IsolationStrict    IsolationLevel = "strict"
	IsolationProcess   IsolationLevel = "process"
	IsolationContainer IsolationLevel = "container"
)

type ToolDefinition struct {
	Name           string
	Signature      string
	Isolation      IsolationLevel
	Implementation any
}

type IsolationHooks struct {
	replaceDefinition func(oldDef, newDef *ToolDefinition) error
	onToolExecution   func(input any) (any, error)
}

func NewIsolationHooks() *IsolationHooks {
	return &IsolationHooks{}
}

func (h *IsolationHooks) ReplaceDefinition(oldDef, newDef *ToolDefinition) error {
	if oldDef == nil || newDef == nil {
		return errors.New("tool definitions cannot be nil")
	}
	if oldDef.Signature != newDef.Signature {
		return errors.New("signature mismatch: hot-reload requires same signature")
	}
	if oldDef.Isolation != newDef.Isolation {
		return errors.New("isolation level mismatch: cannot change isolation during hot-reload")
	}
	return nil
}

func (h *IsolationHooks) OnToolExecution(input any) (any, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	switch v := input.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		existingTaints := make(map[string]bool)

		if taints, ok := v["_taints"].([]string); ok {
			for _, t := range taints {
				existingTaints[t] = true
			}
		}

		existingTaints["TOOL_OUTPUT"] = true
		mergedTaints := make([]string, 0, len(existingTaints))
		for t := range existingTaints {
			mergedTaints = append(mergedTaints, t)
		}
		result["_taints"] = mergedTaints

		for k, val := range v {
			if k == "_taints" {
				continue
			}
			result[k] = val
		}

		return result, nil
	default:
		return input, nil
	}
}
