# Layer 7: Agent Charts & LLM Integration

**Parent Scope:** `implementation-scope.md`  
**Dependencies:** Layer 6 (Tools & Orchestration)  
**Status:** Planning Phase

---

## Overview

Layer 7 provides the intelligence layer of Maelstrom, enabling autonomous agent execution through statechart-based LLM integration. This layer implements the core agent runtime that orchestrates multi-agent workflows, LLM interactions, and autonomous decision-making.

### Core Capabilities

- **Agent Runtime**: Autonomous agent execution engine
- **LLM Integration**: Multi-model LLM support with abstraction layer
- **Multi-Agent Orchestration**: Coordination of multiple agents working together
- **Statechart-Driven Agents**: Agents controlled by statechart definitions
- **Memory Management**: Short-term and long-term memory systems
- **Learning & Adaptation**: Agent learning from interactions

---

## Dependencies

### Required from Layer 6

- `pkg/orchestration/tool_registry.go` - Tool discovery and registration
- `pkg/orchestration/tool_executor.go` - Tool execution capabilities
- `pkg/orchestration/context_manager.go` - Context management
- `pkg/orchestration/workflow_engine.go` - Workflow orchestration primitives
- `pkg/communication/mailbox.go` - Agent communication channels
- `pkg/communication/message.go` - Message structures

### Required from Layer 5

- `pkg/platform/database.go` - Persistent storage
- `pkg/platform/cache.go` - Caching layer
- `pkg/platform/eventbus.go` - Event bus for agent events
- `pkg/platform/scheduler.go` - Task scheduling

### Required from Layer 4

- `pkg/security/datasource.go` - Secure data access
- `pkg/security/auth.go` - Authentication

---

## Detailed Specifications

### 7.1 Agent Runtime Core

#### 7.1.1 Agent Interface

```go
// pkg/agent/agent.go

package agent

type Agent interface {
    ID() string
    Name() string
    State() StatechartState
    Execute(context.Context) error
    Stop()
    GetMemory() Memory
    GetTools() []Tool
    UpdateState(StatechartState) error
}

type AgentConfig struct {
    ID          string
    Name        string
    Model       string
    Temperature float64
    MaxTokens   int
    MemorySize  int
}
```

#### 7.1.2 Agent Lifecycle

- **Created**: Agent instantiated with configuration
- **Initialized**: Statechart loaded, memory allocated, tools registered
- **Running**: Actively processing messages and executing actions
- **Paused**: Temporarily suspended (can be resumed)
- **Stopped**: Gracefully terminated
- **Destroyed**: Resources cleaned up

#### 7.1.3 Agent State Machine

Each agent runs a statechart with these states:
- `idle`: Waiting for work
- `processing`: Executing a task
- `learning`: Updating internal models
- `error`: Error state requiring intervention

---

### 7.2 LLM Integration Layer

#### 7.2.1 Model Abstraction

```go
// pkg/llm/interface.go

package llm

type Model interface {
    Generate(ctx context.Context, messages []Message) (*Response, error)
    GenerateStream(ctx context.Context, messages []Message) (<-chan *Response, error)
    Embed(ctx context.Context, text string) ([]float32, error)
    Name() string
    Config() ModelConfig
}

type ModelConfig struct {
    APIKey       string
    Endpoint     string
    ModelName    string
    Temperature  float64
    MaxTokens    int
    Timeout      time.Duration
}
```

#### 7.2.2 Supported Models

- **OpenAI**: GPT-4, GPT-3.5
- **Anthropic**: Claude 2, Claude Instant
- **Google**: PaLM 2, Gemini
- **Open Source**: Llama 2, Mistral (via local deployment)
- **Custom**: Any model with compatible API

#### 7.2.3 Prompt Management

- **Prompt Templates**: Versioned prompt definitions
- **Prompt Cache**: Cached responses for repeated prompts
- **Prompt Testing**: Built-in prompt evaluation tools
- **Prompt Versioning**: Git-like version control for prompts

```go
// pkg/llm/prompt.go

type Prompt struct {
    ID          string
    Name        string
    Version     string
    Content     string
    Variables   []string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (p *Prompt) Render(vars map[string]string) (string, error)
func (p *Prompt) Validate() error
```

---

### 7.3 Multi-Agent Orchestration

#### 7.3.1 Agent Teams

```go
// pkg/agent/team.go

type Team struct {
    ID        string
    Name      string
    Agents    []Agent
    Leader    Agent
    Protocol  TeamProtocol
}

type TeamProtocol interface {
    DistributeTask(task Task) error
    AggregateResults(results []Result) Result
    ResolveConflict(conflict Conflict) error
}
```

#### 7.3.2 Communication Patterns

- **Broadcast**: Send to all agents
- **Unicast**: Send to specific agent
- **Request-Response**: Synchronous interaction
- **Pub/Sub**: Event-based communication
- **Blackboard**: Shared memory space

#### 7.3.3 Coordination Strategies

1. **Centralized**: Leader agent coordinates all others
2. **Distributed**: Peer-to-peer coordination
3. **Hybrid**: Combination of both
4. **Market-based**: Agents bid for tasks

---

### 7.4 Memory Systems

#### 7.4.1 Short-Term Memory

- **Working Memory**: Current task context
- **Conversation History**: Recent interactions
- **Token-Limited**: Configurable size limit

#### 7.4.2 Long-Term Memory

```go
// pkg/agent/memory.go

type Memory interface {
    Store(ctx context.Context, item MemoryItem) error
    Retrieve(ctx context.Context, query string) ([]MemoryItem, error)
    Search(ctx context.Context, embedding []float32, topK int) ([]MemoryItem, error)
    Delete(ctx context.Context, id string) error
    Clear() error
}

type MemoryItem struct {
    ID        string
    Content   string
    Embedding []float32
    Metadata  map[string]interface{}
    CreatedAt time.Time
}
```

#### 7.4.3 Memory Management

- **Vector Storage**: Semantic search capability
- **TLV (Time, Location, Value)**: Rich metadata
- **Automatic Pruning**: Remove old/irrelevant memories
- **Memory Consolidation**: Merge related memories

---

### 7.5 Learning & Adaptation

#### 7.5.1 Learning Types

1. **One-Shot**: Learn from single example
2. **Few-Shot**: Learn from multiple examples
3. **Continuous**: Incremental learning over time
4. **Reinforcement**: Learn from rewards/punishments

#### 7.5.2 Feedback Integration

```go
// pkg/agent/learning.go

type LearningEngine interface {
    ProcessFeedback(feedback Feedback) error
    UpdateModel(ctx context.Context) error
    GetConfidence() float64
    RequestHumanHelp() error
}

type Feedback struct {
    Type        FeedbackType
    Content     string
    Rating      int // 1-5
    Context     map[string]interface{}
    Timestamp   time.Time
}

type FeedbackType string

const (
    FeedbackCorrect   FeedbackType = "correct"
    FeedbackIncorrect FeedbackType = "incorrect"
    FeedbackHelpful  FeedbackType = "helpful"
    FeedbackUnhelpful FeedbackType = "unhelpful"
)
```

---

## File Structure

```
pkg/agent/
├── agent.go           # Agent interface and implementation
├── agent_factory.go   # Agent creation and lifecycle
├── team.go            # Multi-agent team coordination
├── memory.go          # Memory management
├── learning.go        # Learning and adaptation
├── statechart.go      # Agent statechart integration
└── types.go           # Common types and enums

pkg/llm/
├── interface.go       # LLM model interface
├── factory.go         # Model factory and registration
├── openai.go          # OpenAI integration
├── anthropic.go       # Anthropic integration
├── google.go          # Google integration
├── local.go           # Local model support
├── prompt.go          # Prompt management
├── cache.go           # Prompt/response caching
└── types.go           # LLM types

pkg/orchestration/
├── coordinator.go     # Agent team coordinator
├── communicator.go    # Inter-agent communication
├── task_distributor.go # Task distribution logic
└── protocol.go        # Communication protocols
```

---

## TDD Implementation Plan

### Phase 1: Core Agent Runtime (20 tests)

#### Test Suite 1: Agent Lifecycle
1. TestAgent_CreatedWithValidConfig
2. TestAgent_InitialStateIsIdle
3. TestAgent_TransitionsToInitialized
4. TestAgent_TransitionsToRunning
5. TestAgent_TransitionsToPaused
6. TestAgent_TransitionsToStopped
7. TestAgent_CanBeRestarted
8. TestAgent_ReleasesResourcesOnDestroy

#### Test Suite 2: Agent Execution
9. TestAgent_ExecuteProcessesMessage
10. TestAgent_ExecuteCallsAppropriateAction
11. TestAgent_ExecuteHandlesErrorGracefully
12. TestAgent_UpdateStateChangesAgentState
13. TestAgent_GetMemoryReturnsMemoryInstance
14. TestAgent_GetToolsReturnsRegisteredTools

#### Test Suite 3: Agent Statechart Integration
15. TestAgent_LoadsStatechartFromDefinition
16. TestAgent_TransitionsFollowStatechartRules
17. TestAgent_HandlesInvalidTransitions
18. TestAgent_SavesStateToPersistence
19. TestAgent_LoadsStateFromPersistence
20. TestAgent_ResetsToInitialState

### Phase 2: LLM Integration (25 tests)

#### Test Suite 4: Model Abstraction
21. TestModelFactory_RegistersModels
22. TestModelFactory_CreatesCorrectModel
23. TestModel_GenerateReturnsValidResponse
24. TestModel_GenerateStreamReturnsChannel
25. TestModel_UsesCorrectConfiguration
26. TestModel_HandlesAPIErrors
27. TestModel_TimesOutAfterConfiguredDuration

#### Test Suite 5: Model Implementations
28. TestOpenAIModel_GeneratesText
29. TestOpenAIModel_HandlesRateLimits
30. TestAnthropicModel_GeneratesText
31. TestGoogleModel_GeneratesText
32. TestLocalModel_GeneratesTextOffline

#### Test Suite 6: Prompt Management
33. TestPrompt_RenderReplacesVariables
34. TestPrompt_ValidateRejectsInvalid
35. TestPrompt_CacheStoresResponse
36. TestPrompt_CacheReturnsCachedResponse
37. TestPrompt_VersioningCreatesNewVersion
38. TestPrompt_RollbackRevertsVersion

#### Test Suite 7: Memory Systems
39. TestMemory_StoreStoresItem
40. TestMemory_RetrieveReturnsMatchingItems
41. TestMemory_SearchReturnsSimilarItems
42. TestMemory_DeleteRemovesItem
43. TestMemory_ClearRemovesAll
44. TestMemory_EmbeddingGeneration
45. TestMemory_AutomaticPruning
46. TestMemory_ConsolidationMergesItems

#### Test Suite 8: Learning
47. TestLearningEngine_ProcessFeedbackStoresFeedback
48. TestLearningEngine_UpdateModelAdjustsWeights
49. TestLearningEngine_GetConfidenceReturnsScore
50. TestLearningEngine_RequestHumanHelpNotifiesUser

### Phase 3: Multi-Agent Orchestration (25 tests)

#### Test Suite 9: Team Formation
51. TestTeam_CreatedWithAgents
52. TestTeam_SetsLeaderAgent
53. TestTeam_AddsNewAgent
54. TestTeam_RemovesAgent
55. TestTeam_DissolvesWhenEmpty

#### Test Suite 10: Communication
56. TestCommunicator_BroadcastSendsToAll
57. TestCommunicator_UnicastSendsToOne
58. TestCommunicator_RequestResponseExchanges
59. TestCommunicator_PubSubDeliversEvents
60. TestCommunicator_BlackboardSharedMemory

#### Test Suite 11: Coordination
62. TestCoordinator_CentralizedDistributesTasks
63. TestCoordinator_DistributedCoordinatesPeers
64. TestCoordinator_HybridApproachWorks
65. TestCoordinator_MarketBasedAuctions
66. TestCoordinator_ResolveConflictResolves

#### Test Suite 12: Agent Interaction
67. TestAgent_InteractsWithOtherAgent
68. TestAgent_CollaboratesOnTask
69. TestAgent_HandlesConflictingGoals
70. TestAgent_AggregatesTeamResults

### Phase 4: Integration & Edge Cases (5 tests)

71. TestAgent_LifecycleWithPersistence
72. TestTeam_CommunicationUnderLoad
73. TestMemory_PerformanceWithLargeDataset
74. TestLearning_ConsecutiveFeedbackLoops
75. TestOrchestration_FailoverScenarios

---

## Dependencies

### External Dependencies

- **LLM APIs**: OpenAI, Anthropic, Google (API keys required)
- **Vector Database**: Weaviate, Pinecone, or pgvector
- **Embedding Models**: Sentence Transformers, OpenAI embeddings
- **Cache**: Redis or in-memory for prompt caching

### Internal Dependencies

- Layer 6: Tool registry, workflow engine
- Layer 5: Database, cache, event bus
- Layer 4: Authentication for model APIs
- Layer 0: Statechart engine for agent state machines

---

## Risk Assessment

### Technical Risks

1. **LLM API Reliability**
   - **Mitigation**: Multi-provider fallback, local model support
   - **Impact**: High - Core functionality depends on LLMs

2. **Memory Scalability**
   - **Mitigation**: Efficient vector indexing, pagination
   - **Impact**: Medium - Affects performance with large memories

3. **Multi-Agent Coordination Complexity**
   - **Mitigation**: Simple protocols first, incremental complexity
   - **Impact**: High - Core to Layer 7 purpose

4. **Learning Convergence**
   - **Mitigation**: Conservative updates, human oversight
   - **Impact**: Medium - Affects agent improvement

### Operational Risks

1. **Cost Control**
   - **Mitigation**: Token limits, caching, usage monitoring
   - **Impact**: Medium - API costs can escalate

2. **Performance**
   - **Mitigation**: Async processing, batching, caching
   - **Impact**: Medium - LLM calls are slow

---

## Open Questions

1. **Which vector database to use?**
   - Options: pgvector, Weaviate, Pinecone, Milvus
   - Decision criteria: Cost, scalability, existing infrastructure

2. **What embedding models to support?**
   - Options: OpenAI, Sentence Transformers, custom
   - Decision criteria: Quality, cost, latency

3. **Learning update frequency?**
   - Options: Real-time, batch, on-demand
   - Decision criteria: Use case requirements

4. **Team size limits?**
   - Options: No limit, configurable, fixed
   - Decision criteria: Performance, complexity

5. **Memory persistence strategy?**
   - Options: Every change, batch, on-demand
   - Decision criteria: Performance vs. durability

---

## Testing Strategy

### Unit Tests
- Agent lifecycle states
- LLM model implementations
- Memory operations
- Learning engine logic
- Communication protocols

### Integration Tests
- Agent with LLM integration
- Team coordination scenarios
- Memory persistence
- Multi-agent workflows

### End-to-End Tests
- Complete agent workflows
- Multi-agent collaboration
- Learning over time
- Failure recovery

### Performance Tests
- Memory with 10K+ items
- Team with 100+ agents
- Concurrent agent execution
- LLM API rate limiting

---

## Success Criteria

- [ ] Agents can execute autonomously with LLM guidance
- [ ] Multi-agent teams coordinate effectively
- [ ] Memory systems store and retrieve efficiently
- [ ] Learning engine improves agent performance
- [ ] All 75 tests pass
- [ ] Performance meets requirements (latency, throughput)
- [ ] Fails gracefully under various failure scenarios

---
