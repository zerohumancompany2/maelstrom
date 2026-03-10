package orchestrator

import (
	"sync"
)

// ParallelExecutor executes tools concurrently with par_continue pattern.
type ParallelExecutor struct {
	policy      ExecutionPolicy
	maxParallel int
	databag     map[string]any
}

// NewParallelExecutor creates a new ParallelExecutor with the given policy.
func NewParallelExecutor(policy ExecutionPolicy) *ParallelExecutor {
	return &ParallelExecutor{
		policy:      policy,
		maxParallel: policy.MaxParallel,
		databag:     make(map[string]any),
	}
}

// Execute executes tool calls concurrently according to the par_continue policy.
func (e *ParallelExecutor) Execute(tools []ToolCall) (<-chan ExecutionResult, error) {
	resultChan := make(chan ExecutionResult, len(tools))
	var wg sync.WaitGroup

	for _, toolCall := range tools {
		wg.Add(1)
		go func(tc ToolCall) {
			defer wg.Done()
			result := ExecutionResult{
				Output: map[string]any{
					"tool": tc.Name,
					"args": tc.Arguments,
				},
			}
			resultChan <- result
		}(toolCall)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan, nil
}

// ExecutionResultWithTool contains the result of a tool execution with tool name.
type ExecutionResultWithTool struct {
	ToolName string
	ExecutionResult
}

// ResultAggregator collects and orders results from parallel execution.
type ResultAggregator struct {
	results []ExecutionResultWithTool
	mu      sync.Mutex
}

// NewResultAggregator creates a new ResultAggregator.
func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{
		results: make([]ExecutionResultWithTool, 0),
	}
}

// AddResult adds a result to the aggregator.
func (a *ResultAggregator) AddResult(result ExecutionResultWithTool) {
	panic("not implemented")
}

// GetResults returns all results ordered by tool name.
func (a *ResultAggregator) GetResults() []ExecutionResultWithTool {
	panic("not implemented")
}
