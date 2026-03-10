package orchestrator

import (
	"sort"
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
func (e *ParallelExecutor) Execute(tools []ToolCall) (<-chan ExecutionResultWithTool, error) {
	resultChan := make(chan ExecutionResultWithTool, len(tools))
	var wg sync.WaitGroup
	results := make([]ExecutionResultWithTool, 0, len(tools))
	var mu sync.Mutex

	for _, toolCall := range tools {
		wg.Add(1)
		go func(tc ToolCall) {
			defer wg.Done()
			result := ExecutionResultWithTool{
				ToolName: tc.Name,
				ExecutionResult: ExecutionResult{
					Output: map[string]any{
						"tool": tc.Name,
						"args": tc.Arguments,
					},
				},
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(toolCall)
	}

	go func() {
		wg.Wait()
		sort.Slice(results, func(i, j int) bool {
			return results[i].ToolName < results[j].ToolName
		})
		for _, result := range results {
			resultChan <- result
		}
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
