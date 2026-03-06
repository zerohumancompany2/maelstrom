# Maelstrom Security Architecture

**Version**: 1.0  
**Date**: 2026-03-02  
**Based on**: arch-v1.md Section 4 (Security & Boundaries)

---

## Philosophy

### Core Principle: Security by Design

Maelstrom enforces a fundamental invariant: **all data entering the runtime is tainted at the border**. This is not a runtime check—it's a compile-time guarantee enforced by the type system. No untainted information exists inside the runtime.

### Why This Approach?

1. **Zero Human Intervention**: Agents operate autonomously without human oversight. Security must be automatic and foolproof.
2. **Prompt Injection Resistance**: Untrusted data (user input, external webhooks, mail) can never directly access sensitive inner data.
3. **Data Loss Prevention**: Taint tracking follows data through the entire system, preventing accidental exfiltration of secrets or PII.
4. **Boundary Enforcement**: Inner/DMZ/outer boundaries are not suggestions—they're hard constraints enforced by the runtime.

---

## Architecture Overview

### Three-Tier Boundary Model

```
┌─────────────────────────────────────────────────────────┐
│                      Outer Boundary                       │
│  (Untrusted ingress: user chat, webhooks, external mail) │
├─────────────────────────────────────────────────────────┤
│                      DMZ Boundary                         │
│  (Mediated access: tools wrapped, outputs sanitized)      │
├─────────────────────────────────────────────────────────┤
│                      Inner Boundary                       │
│  (Sensitive data: secrets, production APIs, internal DBs)│
└─────────────────────────────────────────────────────────┘
```

**Boundary Rules**:
- **Inner**: Full read/write to sensitive data. Zero prompt injection surface.
- **DMZ**: Mediated access. All outputs sanitized (PII redaction, length limits, schema validation).
- **Outer**: Untrusted ingress only. Can only talk to DMZ agents via mail.

### Data Tainting System

Every piece of data carries an immutable taint set:
- `USER_SUPPLIED` - User input from chat/webhooks
- `EXTERNAL` - Data from external sources (web searches, APIs)
- `PII` - Personally identifiable information
- `SECRET` - API keys, credentials, secrets
- `INNER_ONLY` - Data that must never leave inner boundary
- `WORKSPACE` - Files written to agent workspace
- `TOOL_OUTPUT` - Results from tool execution

**Taint Attachment Points**:
1. Message creation (Gateway → user, Orchestrator → tool_result)
2. ContextBlock assembly (Security calls `applyBlockTaints`)
3. File read (DataSource returns taints from xattr)
4. Any data leaving runtime (uses `allowedOnExit` policy)

**Taint Propagation**:
- Security Service propagates taints on copy/read/write (like DLP tracking)
- Inner → DMZ/outer: auto-strip or block on forbidden taints
- On-disk (snapshots, memory): taints stored with data
- Runtime guard: any action/guard can query taints; violation emits `taint_violation` event

---

## Components

### ContextMap

Assembles context blocks from 4 sources for LLM calls:

| Source | Purpose | Example |
|--------|---------|---------|
| `static` | System prompts, constants | "You are a secure agent..." |
| `session` | LastN/full conversation history | Last 30 messages |
| `memory` | Vector DB queries (RAG) | "current goal" retrieval |
| `tool` | Tool registry filtered by boundary | `webSearch`, `fileRead` (not `innerDbQuery`) |

**Assembly Rules**:
- Blocks ordered by priority (lower number = first)
- Boundary filtering applied per-block via `TaintPolicy`
- Tools visible based on caller boundary (Inner tools invisible to Outer)

**Per-Block Taint Policy**:
```yaml
taintPolicy:
  redactMode: redact | dropBlock | audit
  redactRules:
    - taint: "PII"
      replacement: "[PII REDACTED]"
  allowedForBoundary: [dmz, outer]
```

### TaintEngine

Tracks and enforces taints throughout the runtime:

**API**:
```go
type TaintEngine interface {
    Mark(obj any, taints []string) (any, error)           // Mark on ingestion
    MarkRead(path string, taints []string) ([]string, error)  // Mark on read
    Propagate(obj any, newTaints []string) (any, error)   // Copy with taints
    CheckForbidden(taints []string, boundary BoundaryType) error
    Redact(obj any, rules []RedactRule) (any, error)
    ReportTaints(chartID string) (TaintMap, error)
}
```

**Behavior**:
- `Mark()`: Attaches taints when data enters runtime (proactive)
- `MarkRead()`: Attaches taints when data read from storage (reactive)
- `Propagate()`: Copies object with taints preserved
- `CheckForbidden()`: Returns error if taints forbidden for boundary
- `Redact()`: Recursively replaces tainted values per rules
- `ReportTaints()`: Returns per-chart taint map for observability

### BoundaryService

Enforces boundaries on mail and charts:

**API**:
```go
type BoundaryService interface {
    ValidateAndSanitize(mail Mail, src, tgt BoundaryType) (Mail, error)
    MarkTaint(obj any, taints []string) (any, error)
    CheckForbidden(taints []string, boundary BoundaryType) error
    NamespaceIsolate(chartID, operation string) (IsolatedView, error)
    ReportTaints(chartID string) (TaintMap, error)
    PrepareContextForBoundary(chartID string, boundary BoundaryType) error
}
```

**Behavior**:
- `ValidateAndSanitize()`: Filters taints from mail based on src/tgt boundaries
- `MarkTaint()`: Attaches taints to objects
- `CheckForbidden()`: Validates taints allowed for boundary
- `NamespaceIsolate()`: Creates isolated view for chart operation
- `PrepareContextForBoundary()`: Called before every `assembleContextMap`

### DataSource

Pluggable storage interface with automatic tainting:

**API**:
```go
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary BoundaryType) error
}
```

**Built-ins**:
- `localDisk` - xattr tagging + sidecar fallback
- `s3` - object storage with tags
- `gcs` - Google Cloud Storage
- `inMemoryWorkspace` - ephemeral storage

**Registration** (explicit in bootstrap):
```go
func Register(name string, factory func(config map[string]any) (DataSource, error))
func Get(name string, config map[string]any) (DataSource, error)
```

**Behavior**:
- On write: sets xattr `user.maelstrom.taints=PII,TOOL_OUTPUT`
- On read: returns taints from xattr (Security attaches to result)
- On access: validates boundary permissions

---

## Implementation Details

### ContextMap Assembly Flow

```
1. Agent requests context for LLM call
2. Security calls PrepareContextForBoundary(chartID, boundary)
3. ContextMap.Assemble() called:
   - staticSource(): returns static content directly
   - sessionSource(): queries session store for lastN/full
   - memorySource(): calls memory service with query
   - toolSource(): filters tools by caller boundary
4. Each block filtered by TaintPolicy:
   - forbidden taints redacted or dropped
   - per-block rules override global settings
5. Assembled prompt sent to LLM
```

### Taint Propagation Flow

```
1. Data enters runtime (Message, file read, tool result)
2. TaintEngine.Mark() or MarkRead() attaches taints
3. Data copied/modified: TaintEngine.Propagate() preserves taints
4. Data leaves boundary: CheckForbidden() validates allowed taints
5. Violation: emits taint_violation event to dead-letter queue
6. Sanitization: Redact() applies per-block rules
```

### Boundary Enforcement Flow

```
1. Mail sent from agent A (boundary A) to agent B (boundary B)
2. BoundaryService.ValidateAndSanitize(mail, A, B) called:
   - If A=Inner, B=Outer: strip forbidden taints (PII, SECRET, INNER_ONLY)
   - If A=DMZ, B=Outer: pass through (DMZ already sanitized)
   - If A=Outer, B=Inner: reject (Outer cannot reach Inner directly)
3. Mail delivered to target inbox
4. Agent processes mail with sanitized content
```

### DataSource Integration Flow

```
1. Tool calls sys:datasources.TagOnWrite(path, taints)
2. DataSource implementation (e.g., localDisk):
   - Writes file to disk
   - Sets xattr with taints
   - Updates sidecar manifest
3. Later, tool calls sys:datasources.GetTaints(path)
4. Security Service attaches taints to tool_result
5. ContextMap assembly applies per-block redaction rules
```

---

## Security Guarantees

### Compile-Time Guarantees

1. **No Untainted Data**: All data tainted at border (type system enforcement)
2. **Boundary Enforcement**: Inner tools invisible to Outer callers
3. **Taint Propagation**: Taints follow data through copies/reads/writes
4. **Audit Trail**: All taint violations logged to dead-letter queue

### Runtime Guarantees

1. **Mail Validation**: All mail validated before delivery
2. **Context Filtering**: ContextMap blocks filtered by boundary
3. **Stream Sanitization**: Every chunk passes through DMZ sanitizers
4. **Namespace Isolation**: Each chart operation in isolated namespace

### What's NOT Enforced (Deferred)

1. **Quiescence Detection**: Conservative `IsQuiescent()` returns false for parallel states
2. **Context Transform Validation**: Templates validated at load time, not runtime
3. **Advanced RAG**: Vector DB integration is stubbed
4. **Cross-Chart Taint Tracking**: Per-chart taints, not global tracking

---

## Test Coverage

### Security Layer Tests (22 total)

| Phase | Tests | Coverage |
|-------|-------|----------|
| 3a: ContextMap | 5 | Assembly order, all 4 sources |
| 3b: Taint Engine | 6 | Mark, Propagate, CheckForbidden, Redact, Report |
| 3c: Boundary Service | 4 | Enforcement, MailValidate, ChartCheck, Namespace |
| 3d: DataSource | 4 | Register, LocalDisk, GetTaints, ValidateAccess |
| 3e: Integration | 3 | KernelBootstrap, ContextMapFiltering, MailValidation |

**Total**: 161 tests (121 existing + 40 new including security)

### Test Philosophy

- **1:1 semantics-to-tests ratio**: Each test covers one behavior
- **TDD workflow**: Stub → test → implement → commit per behavior
- **Boundary-focused**: Tests verify taints don't cross forbidden boundaries
- **Integration coverage**: End-to-end validation of security chain

---

## Design Decisions

### 1. Explicit DataSource Registration

**Why**: Clear dependency graph, testable, configurable.

**Implementation**:
```go
func init() {
    datasource.Register("localDisk", func(config map[string]any) (datasource.DataSource, error) {
        return &localDisk{
            path:    config["path"].(string),
            xattrNS: config["xattrNamespace"].(string),
        }, nil
    })
}
```

### 2. Taint Invariant (No Runtime Checks)

**Why**: Compile-time guarantee is stronger and cheaper than runtime checks.

**Implementation**:
- `Mark()` attaches taints when data enters runtime
- `MarkRead()` attaches taints when data read from storage
- No data exists without taints inside runtime

### 3. ContextMap All 4 Sources

**Why**: Each source serves different use cases.

**Implementation**:
- Static: System prompts, constants
- Session: LastN/full conversation history
- Memory: Vector DB queries (RAG)
- Tool: Tool registry filtered by boundary

### 4. Boundary Filtering at Assembly Time

**Why**: Prevents forbidden taints from ever entering LLM prompt.

**Implementation**:
- `PrepareContextForBoundary()` called before every `assembleContextMap`
- Blocks with forbidden taints redacted or dropped
- Per-block `taintPolicy` overrides global settings

### 5. TaintMap Per-Chart Tracking

**Why**: Each chart has independent taint state.

**Implementation**:
- `TaintMap map[string][]string` per runtime ID
- `ReportTaints(chartID)` returns chart-specific taints
- `Merge()` combines taints from multiple sources

---

## Future Enhancements

### Phase 4: Platform Services (Layer 5)

- `sys:gateway` - Channel adapters (HTTP/SSE/WS/Email/Slack)
- `sys:admin` - k9s-style terminal/web console (outer-only, 2FA-gated)
- `sys:persistence` - Snapshots, event sourcing, version migration
- `sys:heartbeat` - Scheduled agent wake-ups with HEARTBEAT.md injection
- `sys:memory` - Long-term memory (vector/graph stores)
- `sys:human-gateway` - Chat interface for human-in-the-loop
- `sys:tools` - Tool registry and resolution
- `sys:datasources` - Pluggable FS/object storage with tainting

### Phase 5: Execution (Layer 6)

- Tool registry with boundary-aware schemas
- Orchestrator sub-chart (seq_failfast, seq_continue, par_continue)
- Execution policies with isolation (container/process/sandbox/strict)

### Phase 6: Agents (Layer 7)

- Agent Charts (OODA loop: Observe, Orient, Decide, Act, Reflect)
- LLM bridge with ContextMap assembly
- Sub-agents (attached/detached) with lifecycle binding

### Phase 7: External Interface (Layer 8)

- HTTP gateway with OpenAPI 3.1 generation
- SSE/WS streaming with chunk-level taint stripping
- 2FA for admin operations
- Human chat sessions with read-only ContextMap access

---

## Questions for Review

1. **Taint Granularity**: Should taints be more fine-grained (e.g., per-field vs per-object)?
2. **Redaction Strategy**: Is `redact` the right default, or should `dropBlock` be preferred?
3. **DataSource Abstraction**: Should we support more than just xattr (e.g., database metadata)?
4. **Boundary Visibility**: Is the current tool visibility model (Inner→DMZ→Outer) correct?
5. **Test Coverage**: Are we missing any critical security scenarios?

---

## References

- [arch-v1.md Section 4](../arch-v1.md#4-security--boundaries) - Boundary model and tainting
- [arch-v1.md Section 11](../arch-v1.md#11-datasources--tainting) - DataSource abstraction
- [arch-v1.md Section 12](../arch-v1.md#12-bootstrap-sequence) - Kernel bootstrap security
- [security-layer-implementation-plan.md](./initial-exploration/security-layer-implementation-plan.md) - Implementation plan
- [20260302-security-layer-implementation.md](./sessions/20260302-security-layer-implementation.md) - Session documentation