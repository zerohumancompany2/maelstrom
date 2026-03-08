# Layer 4: Security & Boundaries + DataSources

**Reference**: `docs/arch-v1.md` Section 4, 11  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Overview

Layer 4 implements security enforcement, boundary controls, tainting mechanisms, and pluggable data sources. This layer ensures that data crossing boundaries (inner/DMZ/outer) is properly sanitized and that file/object storage operations carry appropriate taints.

**Key Guarantee**: "Never allow outer/DMZ ContextMap to contain untagged inner data"

---

## Architectural Dependencies

```
Layer 4 depends on:
├── Layer 3 (Communication & Mail System)
│   └── Mail system for violation reports and boundary enforcement
├── Layer 2 (Core Platform Services)
│   └── sys:security must be online to enforce boundaries
└── Layer 0 (Statechart Engine) ✓ COMPLETE
    └── ApplicationContext with taint-aware Get/Set
```

**Why Layer 3 First**: Security (sys:security) USES the Mail system to report violations. Mail is the transport; Security is the policy enforcement.

---

## Boundary Model (arch-v1.md 4.1)

Three boundary types:

| Boundary | Capabilities | Security Model |
|----------|-------------|----------------|
| **inner** | Full read/write to sensitive application-context slices (secrets, internal DBs, production APIs) | Prompt injection surface is zero because no untrusted Messages ever enter an inner Chart |
| **DMZ** | Mediated access. Tools/sub-agents are wrapped; outputs are sanitized (PII redaction, length limits, schema validation) | ContextMap blocks marked with forbidden taints are stripped before any LLM call |
| **outer** | Untrusted ingress only (user chat, webhooks, mail from external) | Can only talk to DMZ agents via mail; cannot directly read/write inner data |

### Boundary Transitions

- **Explicit**: Transitions across boundaries are explicit
- **Mail-only**: Always go through the Mail system (never direct event passing)
- **Enforced by**: `sys:security` System Service Chart at runtime

---

## Data Tainting (arch-v1.md 4.2)

### Taint Categories (Examples)

- `PII` - Personally Identifiable Information
- `SECRET` - API keys, credentials
- `INNER_ONLY` - Internal-only data
- `USER_SUPPLIED` - User-provided input
- `TOOL_OUTPUT` - Tool-generated output
- `EXTERNAL` - External data sources
- `WORKSPACE` - Agent workspace data

### Taint Attachment Points

1. **At Message creation** (Gateway → user, Orchestrator → tool_result, MemoryService → injected block, Heartbeat, etc.)
2. **At ContextBlock assembly** (Security calls `applyBlockTaints`)
3. **On any data leaving the runtime** (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`

### Taint Propagation Rules

- **Security Service propagates taints** on copy/read/write (like DLP tracking)
- **Inner → DMZ/outer**: auto-strip or block on forbidden taints
- **On-disk** (snapshots, long-term memory): taints stored with data; Persistence refuses writes that would violate policy
- **Runtime guard**: any action/guard can query taints; violation emits `taint_violation` event to dead-letter

---

## Boundary Enforcement (arch-v1.md 4.3)

### Transitions Across Boundaries

- Explicit and always go through the **Mail system** (never direct event passing)
- Enforced by the `sys:security` System Service Chart at runtime

### ContextMap Assembly Enforcement

- Security service calls `prepareContextForBoundary(runtimeId, boundary)` inside every `assembleContextMap`
- Any block containing forbidden taint for current boundary is **redacted or dropped** before LLM call
- Per-block `taintPolicy` overrides global settings

### Stream Sanitization

- Every outgoing chunk passes through DMZ sanitizers (PII redaction, length caps, schema validation, inner-data stripping)
- **Sanitization is per-chunk (stateless)** so latency stays **<50ms**
- No buffering of entire response unless chart explicitly requests it

---

## DataSource Abstraction (arch-v1.md 11)

### Purpose

Pluggable abstraction for filesystem/object storage with automatic tainting. Owned by `sys:datasources`.

### DataSource Types (Built-ins)

1. **localDisk** - xattr + fallback sidecar
2. **s3** - object storage
3. **gcs** - object storage
4. **inMemoryWorkspace** - in-memory workspace

### YAML Declaration Example

```yaml
dataSources:
  - name: agentWorkspace
    type: localDisk
    path: "/agents/{{agentId}}/workspace"
    mountMode: readWrite
    taintMode: inheritFromXattr   # or alwaysTaintAs=INNER_ONLY, or none
    defaultTaints: ["WORKSPACE"]   # baseline taint for anything written here
    xattrNamespace: user.maelstrom   # where we store taints on files

  - name: sharedInnerRepo
    type: localDisk
    path: "/inner/repo"
    mountMode: readOnly
    taintMode: enforceInnerOnly
    allowedForBoundary: ["inner"]

  - name: s3Bucket
    type: objectStorage
    bucket: my-agent-bucket
    taintMode: inheritFromObjectTags
    region: us-east-1
```

### DataSource Interface

```go
type DataSource interface {
    // TagOnWrite sets taints when writing data
    TagOnWrite(path string, taints []string) error
    
    // GetTaints retrieves taints for a path
    GetTaints(path string) ([]string, error)
    
    // ValidateAccess checks if boundary can access this data source
    ValidateAccess(boundary string) error
}
```

### DataSource Behavior

**On Write** (any tool that writes a file):
1. Write action calls `sys:datasources.tagOnWrite(path, incomingTaints)`
2. For localDisk: sets xattr `user.maelstrom.taints=PII,TOOL_OUTPUT` (or JSON-encoded)
3. For object stores: sets object tags/metadata
4. For workspaces: updates sidecar manifest for fast lookup

**On Read** (any tool that reads):
1. Read tool calls `sys:datasources.getTaints(path)`
2. Security service immediately attaches those taints to resulting `tool_result` or injected ContextBlock Message
3. Existing `prepareContextForBoundary` + per-block `taintPolicy` redacts/drops/audits as configured

**Isolation Hook**: Every agent's persistent workspace is mounted via isolated executor (container/namespace) with exact DataSource policy. No direct `open()` syscall allowed — all FS ops go through mediated tool layer.

---

## ContextBlock TaintPolicy (arch-v1.md 3.5)

```yaml
taintPolicy:              # per-block security
  redactMode: redact | dropBlock | audit
  redactRules:
    - taint: string
      replacement: string
  allowedForBoundary: [inner | dmz | outer]
```

### TaintPolicy Fields

- **redactMode**: `redact` | `dropBlock` | `audit`
- **redactRules**: Array of `{taint: string, replacement: string}`
- **allowedForBoundary**: List of boundaries where this block is allowed (`[inner | dmz | outer]`)

### Enforcement Modes

| Mode | Behavior |
|------|----------|
| `strict` | Block forbidden taints entirely |
| `redact` | Replace forbidden values with replacement text |
| `audit` | Allow but log violations |

---

## ChartDefinition TaintPolicy (arch-v1.md 13.1)

```yaml
taintPolicy:
  enforcement: strict | redact | audit
  allowedOnExit: ["PII", "SECRET", "TOOL_OUTPUT"]  # taints allowed to leave
  redactRules:
    - taint: "SECRET"
      replacement: "[REDACTED]"
```

### allowedOnExit Usage

- Defines which taint categories are allowed to leave the runtime
- Enforced on all outbound data paths: Mail, stream chunks, snapshots, sub-agent returns
- Taints not in this list are blocked/redacted at boundaries

### redactRules Structure

```yaml
redactRules:
  - taint: string           # taint category to match
    replacement: string     # replacement text when redacted
```
- Applied when `enforcement` is `redact`
- Can be ChartDefinition-level (global) or ContextBlock-level (per-block)

---

## Security Interface (arch-v1.md 14.2)

```go
type Security interface {
    // ValidateAndSanitize sanitizes mail crossing boundaries
    ValidateAndSanitize(mail Mail, sourceBoundary, targetBoundary BoundaryType) (Mail, error)
    
    // TaintPropagate tracks taints through copy/read/write operations
    TaintPropagate(obj any, newTaints []string) (any, error)
    
    // NamespaceIsolate provides isolated views for agents
    NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)
    
    // CheckTaintPolicy validates data against taint policy
    CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)
    
    // ReportTaints returns TaintMap (runtimeId → object→taints mapping)
    ReportTaints(runtimeId string) (TaintMap, error)
    
    // PrepareContextForBoundary called inside every assembleContextMap
    PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error
}
```

### Security SHALL Statements

- SHALL attach taints at Message creation and file read
- SHALL call `prepareContextForBoundary` inside every `assembleContextMap`
- SHALL redact/drop/audit per-block `taintPolicy` before LLM prompt
- SHALL enforce `allowedOnExit` on Mail, streams, snapshots, sub-agent returns
- SHALL use DataSource interface for all FS/object access

### Security SHALL NOT Statements

- SHALL NOT allow inner data to reach outer/DMZ LLM prompts without redaction
- SHALL NOT permit boundary crossing via Events (only Mail)
- SHALL NOT bypass taint checks on file read operations

---

## Service Implementation Details

### SecurityService

```go
type SecurityService struct {
    id          string  // "sys:security"
    boundaries  BoundaryConfig
    taintEngine TaintEngine
    datasources map[string]DataSource
    mu          sync.RWMutex
}

type BoundaryConfig struct {
    InnerOnly   []string  // paths only inner can access
    DMZAllowed  []string  // paths DMZ can access (sanitized)
    OuterAllowed []string  // paths outer can access
}

type TaintEngine struct {
    BloomFilters map[string]*bloom.Filter  // per-path filters
    Rules        []TaintRule
}

type TaintRule struct {
    SourceTaint   string
    TargetTaint   string
    Action        string  // redact, drop, audit
    Replacement   string
}

func NewSecurityService() *SecurityService

func (s *SecurityService) ID() string

func (s *SecurityService) HandleMail(mail Mail) error

func (s *SecurityService) ValidateAndSanitize(mail Mail, sourceBoundary, targetBoundary BoundaryType) (Mail, error)

func (s *SecurityService) TaintPropagate(obj any, newTaints []string) (any, error)

func (s *SecurityService) NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)

func (s *SecurityService) CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)

func (s *SecurityService) ReportTaints(runtimeId string) (TaintMap, error)

func (s *SecurityService) PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error

func (s *SecurityService) RegisterDataSource(name string, ds DataSource) error
```

### DataSourceRegistry

```go
type DataSourceRegistry struct {
    datasources map[string]DataSource
    mu          sync.RWMutex
}

func NewDataSourceRegistry() *DataSourceRegistry

func (r *DataSourceRegistry) Register(name string, ds DataSource) error

func (r *DataSourceRegistry) Get(name string) (DataSource, error)

func (r *DataSourceRegistry) TagOnWrite(path string, taints []string) error

func (r *DataSourceRegistry) GetTaints(path string) ([]string, error)
```

---

## File Structure

```
maelstrom/
├── pkg/security/
│   ├── boundary.go        # Boundary enforcement (~150 lines)
│   ├── taint.go           # Taint engine (~200 lines)
│   ├── contextmap.go      # ContextMap filtering (~100 lines)
│   ├── rules.go           # Taint rules (~50 lines)
│   ├── service.go         # SecurityService (~200 lines)
│   └── security_test.go   # Tests (~300 lines)
│
├── pkg/datasource/
│   ├── datasource.go      # DataSource interface + impl (~150 lines)
│   ├── localDisk.go       # localDisk implementation (~100 lines)
│   ├── objectStorage.go   # s3/gcs implementations (~150 lines)
│   ├── inMemory.go        # inMemoryWorkspace (~50 lines)
│   ├── registry.go        # DataSourceRegistry (~100 lines)
│   └── datasource_test.go # Tests (~200 lines)
│
└── pkg/services/
    └── security/
        └── service.go     # SecurityService integration (~100 lines)
```

---

## TDD Implementation Plan

### Phase 4.1: Boundary Types

**Test**: `TestBoundaryTypes`
- inner, dmz, outer defined correctly

**Test**: `TestBoundaryTransitions`
- Transitions validated correctly

**Implementation**:
- ~100 lines boundary.go
- Define BoundaryType, BoundaryConfig

### Phase 4.2: Taint Engine

**Test**: `TestTaintEngine_Attach`
- Taints attached to objects

**Test**: `TestTaintEngine_Propagate`
- Taints propagated on copy/read/write

**Test**: `TestTaintEngine_CheckPolicy`
- Policy enforcement works

**Implementation**:
- ~200 lines taint.go
- Bloom filter optimization, rule engine

### Phase 4.3: ContextMap Filtering

**Test**: `TestContextMap_FilterTaints`
- Forbidden taints filtered correctly

**Test**: `TestContextMap_Redact`
- Redaction works

**Test**: `TestContextMap_DropBlock`
- Block dropping works

**Implementation**:
- ~100 lines contextmap.go
- Per-block taintPolicy enforcement

### Phase 4.4: DataSource Interface

**Test**: `TestDataSource_TagOnWrite`
- Taints set on write

**Test**: `TestDataSource_GetTaints`
- Taints retrieved correctly

**Test**: `TestDataSource_ValidateAccess`
- Boundary validation works

**Implementation**:
- ~150 lines datasource.go
- Interface definition

### Phase 4.5: DataSource Implementations

**Test**: `TestLocalDiskDataSource`
- localDisk with xattr works

**Test**: `TestObjectStorageDataSource`
- s3/gcs implementations work

**Test**: `TestInMemoryDataSource`
- inMemoryWorkspace works

**Implementation**:
- ~300 lines localDisk.go + objectStorage.go + inMemory.go

### Phase 4.6: SecurityService

**Test**: `TestSecurityService_HandleMail`
- Mail sanitization works

**Test**: `TestSecurityService_TaintPropagate`
- Taint propagation works

**Test**: `TestSecurityService_PrepareContextForBoundary`
- Context preparation works

**Implementation**:
- ~200 lines service.go
- Integration with Communication layer

### Phase 4.7: Integration

**Test**: `TestFullSecurityFlow`
- End-to-end security enforcement

**Test**: `TestDataSourceIntegration`
- DataSources integrated with Security

**Implementation**:
- Wire all components together

---

## Dependencies & External Libraries

- `sync` - Mutex for thread-safe state
- `github.com/bits-and-blooms/bloom` - Bloom filters for taint optimization (optional)
- `github.com/maelstrom/v3/pkg/statechart` - ApplicationContext compatibility
- `github.com/maelstrom/v3/pkg/mail` - Mail types

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Taint propagation bugs | Medium | High | Comprehensive tests for copy/read/write |
| DataSource integration issues | Medium | Medium | Interface-based design, mocks |
| Bloom filter false positives | Low | Low | False positives are safe (just extra work) |
| Stream sanitization latency | Low | Medium | Per-chunk stateless processing, <50ms target |

---

## Open Questions

1. **Bloom filters**: Use bloom filters for taint optimization, or simple sets?
2. **DataSource persistence**: Store taints in xattr, sidecar, or object tags?
3. **Taint inheritance**: How to handle nested data structures?
4. **Audit logging**: Where to log taint violations (sys:observability)?

---

## References

- `docs/arch-v1.md` Section 4 - Security & Boundaries
- `docs/arch-v1.md` Section 11 - DataSources & Tainting
- `docs/arch-v1.md` Section 3.5 - ContextBlock TaintPolicy
- `docs/arch-v1.md` Section 13.1 - ChartDefinition TaintPolicy
- `docs/arch-v1.md` Section 14.2 - Security SHALL Statements
- `docs/planning/layer-03-communication-mail.md` - Layer 3 (dependency)
- `docs/planning/implementation-scope.md` - Parent scope document

---
