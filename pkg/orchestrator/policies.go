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

// PolicySeqFailFast executes tools sequentially, stopping on first failure.
var PolicySeqFailFast = ExecutionPolicy{
	Mode:       "seq_failfast",
	MaxRetries: 2,
	Isolation:  "strict",
}

// PolicyParContinue executes tools in parallel, continuing on failure.
var PolicyParContinue = ExecutionPolicy{
	Mode:        "par_continue",
	MaxRetries:  1,
	Isolation:   "strict",
	MaxParallel: 8,
}
