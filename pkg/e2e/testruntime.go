package e2e

import (
	"fmt"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/datasource"
	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
	securitysvc "github.com/maelstrom/v3/pkg/services/security"
)

type E2ERuntime struct {
	mu              sync.RWMutex
	securityService *securitysvc.SecurityService
	mailRouter      *mail.MailRouter
	deadLetterQueue []*mail.Mail
	agents          map[string]*TestAgent
	dataSources     map[string]datasource.DataSource
	violations      []*mail.Mail
	streamResults   map[string]*StreamTestResult
	contextBlocks   map[string][]*security.ContextBlock
	blockedBlocks   map[string][]*security.ContextBlock
}

type TestAgent struct {
	ID            string
	Boundary      mail.BoundaryType
	TaintPolicy   security.TaintPolicy
	Inbox         *mail.AgentInbox
	Namespace     string
	WorkspacePath string
}

type StreamTestResult struct {
	Chunks       []mail.StreamChunk
	Latencies    []time.Duration
	TotalLatency time.Duration
	Violations   []*mail.Mail
}

type StreamSession struct {
	runtime        *E2ERuntime
	agentID        string
	clientBoundary mail.BoundaryType
	chunks         []mail.StreamChunk
	startTime      time.Time
}

type DataSourceTestResult struct {
	WrittenPath    string
	WrittenTaints  []string
	ReadTaints     []string
	AttachedTaints []string
	ContextMap     string
	Violations     []*mail.Mail
}

type IsolationTestResult struct {
	SyscallBlocked bool
	SyscallError   error
	Violations     []*mail.Mail
	ToolResult     any
	ToolError      error
}

type ViolationTestResult struct {
	ViolationsEmitted  int
	ViolationsReceived int
	DeadLetterQueue    []*mail.Mail
	Metrics            map[string]interface{}
	QueryResults       []*Violation
}

type Violation struct {
	Type            string
	Source          string
	Target          string
	ForbiddenTaints []string
	Timestamp       time.Time
	CorrelationID   string
}

func NewE2ERuntime() *E2ERuntime {
	rt := &E2ERuntime{
		securityService: securitysvc.NewSecurityService(),
		mailRouter:      mail.NewMailRouter(),
		deadLetterQueue: make([]*mail.Mail, 0),
		agents:          make(map[string]*TestAgent),
		dataSources:     make(map[string]datasource.DataSource),
		violations:      make([]*mail.Mail, 0),
		streamResults:   make(map[string]*StreamTestResult),
		contextBlocks:   make(map[string][]*security.ContextBlock),
		blockedBlocks:   make(map[string][]*security.ContextBlock),
	}

	publisher := &testPublisher{runtime: rt}
	rt.securityService.SetPublisher(publisher)

	return rt
}

type testPublisher struct {
	runtime *E2ERuntime
}

func (p *testPublisher) Publish(m mail.Mail) (mail.Ack, error) {
	if m.Type == mail.MailTypeTaintViolation {
		p.runtime.mu.Lock()
		p.runtime.violations = append(p.runtime.violations, &m)
		p.runtime.deadLetterQueue = append(p.runtime.deadLetterQueue, &m)
		p.runtime.mu.Unlock()
	}
	return mail.Ack{MailID: m.ID, Success: true}, nil
}

func (r *E2ERuntime) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inMemoryDS := datasource.NewInMemoryDataSource()
	r.dataSources["inmemory"] = inMemoryDS

	return nil
}

func (r *E2ERuntime) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agents = make(map[string]*TestAgent)
	r.violations = make([]*mail.Mail, 0)
	r.deadLetterQueue = make([]*mail.Mail, 0)

	return nil
}

func (r *E2ERuntime) CreateAgent(name string, boundary mail.BoundaryType, taintPolicy security.TaintPolicy) *TestAgent {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent := &TestAgent{
		ID:            name,
		Boundary:      boundary,
		TaintPolicy:   taintPolicy,
		Inbox:         &mail.AgentInbox{},
		Namespace:     "agent:" + name,
		WorkspacePath: "/agents/" + name + "/workspace",
	}

	r.agents[name] = agent
	return agent
}

func (r *E2ERuntime) SendUserMessage(agentID string, content string) (*mail.Mail, error) {
	r.mu.RLock()
	agent, ok := r.agents[agentID]
	r.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	userMessage := mail.Mail{
		ID:        "user-msg-" + agentID,
		Type:      mail.MailTypeUser,
		Source:    "sys:gateway",
		Target:    agent.ID,
		Content:   content,
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}

	_ = r.securityService.HandleMail(&userMessage)

	return &userMessage, nil
}

func (r *E2ERuntime) SendMail(source, target string, content any, taints []string) (*mail.Mail, error) {
	r.mu.RLock()
	sourceAgent, sourceOk := r.agents[source]
	targetAgent, targetOk := r.agents[target]
	r.mu.RUnlock()

	if !sourceOk || !targetOk {
		return nil, nil
	}

	m := mail.Mail{
		ID:        "mail-" + source + "-" + target,
		Type:      mail.MailTypeAssistant,
		Source:    source,
		Target:    target,
		Content:   content,
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Boundary: sourceAgent.Boundary,
			Taints:   taints,
		},
	}

	_, err := r.securityService.ValidateAndSanitize(m, sourceAgent.Boundary, targetAgent.Boundary)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *E2ERuntime) GetDeadLetterQueue() []*mail.Mail {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*mail.Mail, len(r.deadLetterQueue))
	copy(result, r.deadLetterQueue)
	return result
}

func (r *E2ERuntime) GetViolations() []*mail.Mail {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*mail.Mail, len(r.violations))
	copy(result, r.violations)
	return result
}

func (r *E2ERuntime) AssembleContextMap(agentID string) (string, error) {
	r.mu.RLock()
	agent, ok := r.agents[agentID]
	r.mu.RUnlock()

	if !ok {
		return "", nil
	}

	return "context-map-for-" + agent.ID, nil
}

func (r *E2ERuntime) StartStreamingSession(agentID string, clientBoundary mail.BoundaryType) (*StreamSession, error) {
	r.mu.RLock()
	_, ok := r.agents[agentID]
	r.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	return &StreamSession{
		runtime:        r,
		agentID:        agentID,
		clientBoundary: clientBoundary,
		chunks:         make([]mail.StreamChunk, 0),
		startTime:      time.Now(),
	}, nil
}

func (r *E2ERuntime) SendStreamChunk(session *StreamSession, data string, taints []string) (time.Duration, error) {
	allowedTaints := r.filterAllowedTaints(taints, session.clientBoundary)

	chunk := mail.StreamChunk{
		Data:     data,
		Sequence: len(session.chunks),
		IsFinal:  false,
		Taints:   allowedTaints,
	}

	start := time.Now()
	session.chunks = append(session.chunks, chunk)
	latency := time.Since(start)

	return latency, nil
}

func (r *E2ERuntime) filterAllowedTaints(taints []string, targetBoundary mail.BoundaryType) []string {
	allowed := make([]string, 0)

	for _, taint := range taints {
		allowed = append(allowed, taint)
	}

	if targetBoundary == mail.OuterBoundary {
		filtered := make([]string, 0)
		for _, taint := range allowed {
			if taint != "PII" && taint != "INNER_ONLY" && taint != "SECRET" {
				filtered = append(filtered, taint)
			}
		}
		allowed = filtered
	}

	return allowed
}

func (r *E2ERuntime) EndStreamSession(session *StreamSession) (*StreamTestResult, error) {
	elapsed := time.Since(session.startTime)

	result := &StreamTestResult{
		Chunks:       session.chunks,
		Latencies:    make([]time.Duration, len(session.chunks)),
		TotalLatency: elapsed,
		Violations:   make([]*mail.Mail, 0),
	}

	r.mu.Lock()
	r.streamResults[session.agentID] = result
	r.mu.Unlock()

	return result, nil
}

func (r *E2ERuntime) WriteFile(agentID, path string, content []byte, taints []string) error {
	r.mu.RLock()
	ds, ok := r.dataSources["inmemory"]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	return ds.TagOnWrite(path, taints)
}

func (r *E2ERuntime) ReadFile(agentID, path string) ([]byte, []string, error) {
	r.mu.RLock()
	ds, ok := r.dataSources["inmemory"]
	r.mu.RUnlock()

	if !ok {
		return nil, nil, nil
	}

	taints, _ := ds.GetTaints(path)
	return []byte("file-content"), taints, nil
}

func (r *E2ERuntime) AssembleContextMapWithDataSource(agentID string, dataSourceBlock *security.ContextBlock) (string, error) {
	return "context-map-with-datasource", nil
}

func (r *E2ERuntime) AttemptDirectSyscall(agentID string, syscallType, path string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	violationMail := mail.Mail{
		ID:     "violation-" + agentID + "-direct-syscall",
		Type:   mail.MailTypeTaintViolation,
		Source: agentID,
		Target: "sys:observability",
		Content: map[string]interface{}{
			"type":    "direct_syscall_blocked",
			"syscall": syscallType,
			"path":    path,
			"policy":  "isolation_policy_violated",
		},
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Taints: []string{"DIRECT_SYSCALL_ATTEMPT"},
		},
	}

	r.violations = append(r.violations, &violationMail)
	r.deadLetterQueue = append(r.deadLetterQueue, &violationMail)

	return true, fmt.Errorf("direct syscall not allowed - use mediated tool layer")
}

func (r *E2ERuntime) CallTool(agentID, toolName, path string) (any, error) {
	return nil, nil
}

func (r *E2ERuntime) GetNamespace(agentID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[agentID]
	if !ok {
		return ""
	}

	return agent.Namespace
}

func (r *E2ERuntime) SetIsolationPolicy(agentID string, policy string) error {
	return nil
}

func (r *E2ERuntime) TriggerViolation(agentID, violationType string, taints []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	violationMail := mail.Mail{
		ID:        "violation-" + agentID + "-" + violationType,
		Type:      mail.MailTypeTaintViolation,
		Source:    agentID,
		Target:    "sys:observability",
		Content:   map[string]interface{}{"type": violationType, "taints": taints},
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Taints: taints,
		},
	}

	r.violations = append(r.violations, &violationMail)
	r.deadLetterQueue = append(r.deadLetterQueue, &violationMail)

	return nil
}

func (r *E2ERuntime) GetMetrics() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make(map[string]interface{})
	metrics["taint_violations_total"] = len(r.violations)

	byType := make(map[string]int)
	for _, v := range r.violations {
		if content, ok := v.Content.(map[string]interface{}); ok {
			if vtype, ok := content["type"].(string); ok {
				byType[vtype]++
			}
		}
	}
	metrics["taint_violations_by_type"] = byType

	return metrics
}

func (r *E2ERuntime) QueryViolations(filters map[string]interface{}) []*Violation {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*Violation, 0)

	for _, v := range r.violations {
		violation := &Violation{
			Source:    v.Source,
			Target:    v.Target,
			Timestamp: v.CreatedAt,
		}

		if content, ok := v.Content.(map[string]interface{}); ok {
			if vtype, ok := content["type"].(string); ok {
				violation.Type = vtype
			}
			if ftaints, ok := content["taints"].([]string); ok {
				violation.ForbiddenTaints = ftaints
			}
		}

		results = append(results, violation)
	}

	return results
}

func (r *E2ERuntime) GetDataSource(name string) datasource.DataSource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dataSources[name]
}

func (r *E2ERuntime) RegisterDataSource(name string, ds datasource.DataSource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dataSources[name] = ds
}

func (r *E2ERuntime) WriteFileViaDataSource(agentID, path string, content []byte, taints []string) error {
	r.mu.RLock()
	ds, ok := r.dataSources["inmemory"]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	return ds.TagOnWrite(path, taints)
}

func (r *E2ERuntime) WriteFileViaDataSourceWithDS(ds datasource.DataSource, path string, taints []string) error {
	return ds.TagOnWrite(path, taints)
}

func (r *E2ERuntime) ReadFileViaDataSource(agentID, path string) (*mail.Mail, error) {
	r.mu.RLock()
	ds, ok := r.dataSources["inmemory"]
	r.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	taints, err := ds.GetTaints(path)
	if err != nil {
		return nil, err
	}

	toolResult := &mail.Mail{
		ID:        "tool-result-" + agentID + "-" + path,
		Type:      mail.MailTypeToolResult,
		Source:    "sys:datasources",
		Target:    agentID,
		Content:   map[string]interface{}{"path": path, "content": "file-content"},
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Taints: []string{"TOOL_OUTPUT"},
		},
	}

	if r.securityService != nil {
		r.attachTaintsToMessage(toolResult, taints)
	}

	return toolResult, nil
}

func (r *E2ERuntime) attachTaintsToMessage(msg *mail.Mail, fileTaints []string) {
	existing := make(map[string]bool)
	for _, t := range msg.Metadata.Taints {
		existing[t] = true
	}

	for _, t := range fileTaints {
		existing[t] = true
	}

	merged := make([]string, 0, len(existing))
	for t := range existing {
		merged = append(merged, t)
	}

	msg.Metadata.Taints = merged
}

func (r *E2ERuntime) PrepareContextForBoundary(agentID string, boundary mail.BoundaryType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.agents[agentID]
	if !ok {
		return nil
	}

	blocks, ok := r.contextBlocks[agentID]
	if !ok || len(blocks) == 0 {
		return nil
	}

	forbiddenTaints := r.getForbiddenTaintsForBoundary(boundary)
	blocked := make([]*security.ContextBlock, 0)

	for _, block := range blocks {
		for taint := range block.Taints {
			if r.isTaintForbidden(taint, forbiddenTaints) {
				if block.TaintPolicy.RedactMode == "dropBlock" {
					blocked = append(blocked, block)
					r.logViolation(agentID, "taint_violation", []string{taint})
				}
				break
			}
		}
	}

	r.blockedBlocks[agentID] = blocked

	filteredBlocks := make([]*security.ContextBlock, 0)
	for _, block := range blocks {
		isBlocked := false
		for _, b := range blocked {
			if b.Name == block.Name {
				isBlocked = true
				break
			}
		}
		if !isBlocked {
			filteredBlocks = append(filteredBlocks, block)
		}
	}

	r.contextBlocks[agentID] = filteredBlocks

	if len(blocked) > 0 {
		return fmt.Errorf("dropped %d context blocks due to forbidden taints", len(blocked))
	}

	return nil
}

func (r *E2ERuntime) getForbiddenTaintsForBoundary(boundary mail.BoundaryType) []string {
	switch boundary {
	case mail.OuterBoundary:
		return []string{"INNER_ONLY", "SECRET", "PII"}
	case mail.DMZBoundary:
		return []string{"INNER_ONLY", "SECRET"}
	default:
		return []string{}
	}
}

func (r *E2ERuntime) isTaintForbidden(taint string, forbidden []string) bool {
	for _, f := range forbidden {
		if taint == f {
			return true
		}
	}
	return false
}

func (r *E2ERuntime) logViolation(agentID, violationType string, taints []string) {
	violationMail := mail.Mail{
		ID:        "violation-" + agentID + "-" + violationType,
		Type:      mail.MailTypeTaintViolation,
		Source:    agentID,
		Target:    "sys:observability",
		Content:   map[string]interface{}{"type": violationType, "taints": taints},
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Taints: taints,
		},
	}

	r.violations = append(r.violations, &violationMail)
	r.deadLetterQueue = append(r.deadLetterQueue, &violationMail)
}

func (r *E2ERuntime) GetContextMap(agentID string) (*security.ContextMap, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.agents[agentID]

	if !ok {
		return nil, nil
	}

	blocks, ok := r.contextBlocks[agentID]
	if !ok {
		return &security.ContextMap{
			Blocks: make([]*security.ContextBlock, 0),
		}, nil
	}

	return &security.ContextMap{
		Blocks: blocks,
	}, nil
}

func (r *E2ERuntime) RegisterContextBlock(agentID string, block *security.ContextBlock) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.contextBlocks == nil {
		r.contextBlocks = make(map[string][]*security.ContextBlock)
	}

	r.contextBlocks[agentID] = append(r.contextBlocks[agentID], block)
}

func (r *E2ERuntime) GetBlockedContextBlocks(agentID string) []*security.ContextBlock {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.blockedBlocks[agentID]
}

func (r *E2ERuntime) AssemblePromptForLLM(agentID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	blocks, ok := r.contextBlocks[agentID]
	if !ok || len(blocks) == 0 {
		return ""
	}

	prompt := ""
	for _, block := range blocks {
		if block.Name != "" {
			if prompt != "" {
				prompt += "\n"
			}
			prompt += block.Content
		}
	}

	return prompt
}
