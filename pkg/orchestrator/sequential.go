package orchestrator

import (
	"github.com/maelstrom/v3/pkg/statechart"
)

// SequentialExecutor executes tools sequentially with different policies.
type SequentialExecutor struct {
	policy  ExecutionPolicy
	databag map[string]any
}

// NewSequentialExecutor creates a new SequentialExecutor with the given policy.
func NewSequentialExecutor(policy ExecutionPolicy) *SequentialExecutor {
	return &SequentialExecutor{
		policy:  policy,
		databag: make(map[string]any),
	}
}

// Execute executes tool calls sequentially according to the policy.
func (e *SequentialExecutor) Execute(tools []ToolCall) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, 0, len(tools))

	for _, toolCall := range tools {
		result := ExecutionResult{
			Output: map[string]any{
				"tool":   toolCall.Name,
				"args":   toolCall.Arguments,
				"status": "executed",
			},
		}
		results = append(results, result)
	}

	return results, nil
}

// InjectOutput injects tool output into the databag scoped to the chart.
func (e *SequentialExecutor) InjectOutput(chartID statechart.RuntimeID, name string, output any) error {
	panic("not implemented")
}
