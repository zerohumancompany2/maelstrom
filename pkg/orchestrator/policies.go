package orchestrator

// ExecutionPolicy defines how tool calls are executed.
type ExecutionPolicy struct {
	Mode        string
	MaxRetries  int
	Isolation   string
	MaxParallel int
	TimeoutMs   int
}

// PolicySeqContinue executes tools sequentially, continuing on failure.
var PolicySeqContinue = ExecutionPolicy{
	Mode:       "seq_continue",
	MaxRetries: 1,
	Isolation:  "process",
}
