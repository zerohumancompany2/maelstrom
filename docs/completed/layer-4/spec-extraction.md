# Layer 4 Specification Extraction

**Source**: `docs/arch-v1.md` (v1.3)  
**Extracted**: 2026-03-09  
**Status**: Specification only - NO implementation

---

## Executive Summary

Layer 4 implements **Security & Boundaries + DataSources** - the core security enforcement layer that ensures data crossing boundaries (inner/DMZ/outer) is properly sanitized and that file/object storage operations carry appropriate taints.

**Key Guarantee** (arch-v1.md L20): "Never allow outer/DMZ ContextMap to contain untagged inner data"

**Primary Sections in arch-v1.md**:
- Section 4: Security & Boundaries (L261-302)
- Section 11: DataSources & Tainting (L739-798)

---

## 1. Components & Responsibilities

### 1.1 Security Service (sys:security)

**Type**: Core Platform Service (Hard-Coded)  
**Reference**: arch-v1.md L457, L757-758

**Responsibilities**:
- Boundary enforcement (inner/DMZ/outer)
- Taint validation and propagation
- ContextMap sanitization
- Namespace isolation
- Stream chunk sanitization

**Interface** (arch-v1.md L1303-1309):

```go
type Security interface {
    ValidateAndSanitize(mail Mail, sourceBoundary, targetBoundary BoundaryType) (Mail, error)
    TaintPropagate(obj any, newTaints []string) (any, error)
    NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)
    CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)
    ReportTaints(runtimeId string) (TaintMap, error)
    PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error
}
```

### 1.2 Taint Engine

**Type**: Sub-component of Security Service  
**Reference**: arch-v1.md L277-287

**Responsibilities**:
- Attach taints at Message creation (L278)
- Propagate taints on copy/read/write (L283)
- Apply bloom filter optimization for fast-path checking (L1798)
- Enforce taint policies per-boundary

**Taint Categories** (arch-v1.md L277, L58, L563, L566):
- `PII` - Personally Identifiable Information
- `SECRET` - API keys, credentials
- `INNER_ONLY` - Internal-only data
- `USER_SUPPLIED` - User-provided input
- `TOOL_OUTPUT` - Tool-generated output
- `EXTERNAL` - External data sources
- `WORKSPACE` - Agent workspace data

### 1.3 DataSource Abstraction

**Type**: Pluggable Infrastructure Component  
**Reference**: arch-v1.md L739-798, L787-793

**Responsibilities**:
- Filesystem/object storage with automatic tainting
- Tag data on write with taints
- Retrieve taints on read
- Validate boundary access

**Interface** (arch-v1.md L787-793):

```go
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary string) error
}
```

**Built-in Implementations** (arch-v1.md L796):
- `localDisk` - xattr + fallback sidecar
- `s3` - object storage
- `gcs` - object storage
- `inMemoryWorkspace` - in-memory workspace

### 1.4 Boundary Enforcement

**Type**: Sub-component of Security Service  
**Reference**: arch-v1.md L261-270

**Responsibilities**:
- Enforce three boundary types (inner/DMZ/outer)
- Validate boundary transitions via Mail only
- Strip forbidden taints from streams

**Boundary Types** (arch-v1.md L261-270):

| Boundary | Capabilities | Security Model |
|----------|-------------|----------------|
| **inner** | Full read/write to sensitive application-context slices (secrets, internal DBs, production APIs) | Prompt injection surface is zero because no untrusted Messages ever enter an inner Chart |
| **DMZ** | Mediated access. Tools/sub-agents are wrapped; outputs are sanitized (PII redaction, length limits, schema validation) | ContextMap blocks marked with forbidden taints are stripped before any LLM call |
| **outer** | Untrusted ingress only (user chat, webhooks, mail from external) | Can only talk to DMZ agents via mail; cannot directly read/write inner data |

### 1.5 ContextMap Filtering

**Type**: Sub-component of Security Service  
**Reference**: arch-v1.md L292-296, L205-223

**Responsibilities**:
- Call `prepareContextForBoundary` inside every `assembleContextMap` (L293)
- Redact/drop/audit per-block taintPolicy before LLM call (L295)
- Enforce per-block taintPolicy overrides global settings (L296)

---

## 2. Interfaces & Contracts

### 2.1 Security Interface Contract

**Full Interface** (arch-v1.md L1303-1309):

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

### 2.2 Security SHALL Statements

**SHALL** (arch-v1.md L1312-1317):
- Attach taints at Message creation and file read (L1312)
- Call `prepareContextForBoundary` inside every `assembleContextMap` (L1313)
- Redact/drop/audit per-block `taintPolicy` before LLM prompt (L1314)
- Enforce `allowedOnExit` on Mail, streams, snapshots, sub-agent returns (L1315)
- Use DataSource interface for all FS/object access (L1316)

**SHALL NOT** (arch-v1.md L1320-1322):
- Allow inner data to reach outer/DMZ LLM prompts without redaction (L1320)
- Permit boundary crossing via Events (only Mail) (L1321)
- Bypass taint checks on file read operations (L1322)

### 2.3 DataSource Interface Contract

**Full Interface** (arch-v1.md L787-793):

```go
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary string) error
}
```

### 2.4 Platform Service Contract

**Base Contract** (arch-v1.md L479-481):

```yaml
handleMail(mail: Mail) → outcomeEvent
```

**sys:datasources specific** (arch-v1.md L491):
- `tagOnWrite(path, taints)`
- `getTaints(path) → []string`

---

## 3. Dependencies on Lower Layers

### 3.1 Dependency Map

```
Layer 4 depends on:
├── Layer 3 (Communication & Mail System)
│   └── Mail system for violation reports and boundary enforcement
├── Layer 2 (Core Platform Services)
│   └── sys:security must be online to enforce boundaries
└── Layer 0 (Statechart Engine)
    └── ApplicationContext with taint-aware Get/Set
```

### 3.2 Layer 2 Integration Points

**sys:security** (arch-v1.md L457):
- Core service spawned by Bootstrap Chart
- Well-known ID: `sys:security`
- Handles Mail for boundary enforcement

**Integration with Layer 2 Services** (arch-v1.md L457-460):
- `sys:security` - Boundary enforcement, taint validation, sanitization
- `sys:communication` - Mail pub/sub backbone (used for violation reports)
- `sys:observability` - Traces, metrics, dead-letter queue (for taint violations)

### 3.3 Layer 3 Integration Points

**Mail System Usage** (arch-v1.md L290, L286):
- Transitions across boundaries always go through Mail system (L290)
- Enforced by `sys:security` System Service Chart at runtime (L291)
- `taint_violation` event emitted to dead-letter queue (L286)

**Mail Types Used by Layer 4** (arch-v1.md L170):
- `taint_violation` - Taint policy violations

### 3.4 Layer 0 Integration Points

**ApplicationContext** (arch-v1.md L383-387):

```go
type ApplicationContext interface {
    Get(key string, callerBoundary BoundaryType) (any, []string, error)  // value, taints, error
    Set(key string, value any, taints []string, callerBoundary BoundaryType) error
    Namespace() string
}
```

**RuntimeContext** (arch-v1.md L374-381):
- Provides read-only context including `ActiveStates`, `Metrics`

---

## 4. Key Workflows & Behaviors

### 4.1 Data Tainting Workflow

**Taint Attachment Points** (arch-v1.md L277-281):

1. **At Message creation** (L278):
   - Gateway → user
   - Orchestrator → tool_result
   - MemoryService → injected block
   - Heartbeat, etc.

2. **At ContextBlock assembly** (L279):
   - Security calls `applyBlockTaints`

3. **On any data leaving the runtime** (L280):
   - Mail, stream chunks, snapshot, sub-agent result
   - Uses `allowedOnExit` policy

**Taint Propagation Rules** (arch-v1.md L283-286):
- Security Service propagates taints on copy/read/write (L283)
- Inner → DMZ/outer: auto-strip or block on forbidden taints (L284)
- On-disk (snapshots, long-term memory): taints stored with data (L285)
- Runtime guard: violation emits `taint_violation` event to dead-letter (L286)

### 4.2 Boundary Enforcement Workflow

**Transitions Across Boundaries** (arch-v1.md L290-291):
- Explicit and always go through Mail system (never direct event passing)
- Enforced by `sys:security` System Service Chart at runtime

**ContextMap Assembly Enforcement** (arch-v1.md L292-296):
- Security service calls `prepareContextForBoundary(runtimeId, boundary)` inside every `assembleContextMap`
- Any block containing forbidden taint for current boundary is redacted or dropped before LLM call
- Per-block `taintPolicy` overrides global settings

### 4.3 Stream Sanitization Workflow

**Stream Sanitization** (arch-v1.md L298-301):
- Every outgoing chunk passes through DMZ sanitizers (PII redaction, length caps, schema validation, inner-data stripping)
- Sanitization is per-chunk (stateless) so latency stays <50ms
- No buffering of entire response unless chart explicitly requests it

### 4.4 DataSource Write Workflow

**On Write** (arch-v1.md L757-761):
1. Write action calls `sys:datasources.tagOnWrite(path, incomingTaints)`
2. For localDisk: sets xattr `user.maelstrom.taints=PII,TOOL_OUTPUT` (or JSON-encoded)
3. For object stores: sets object tags/metadata
4. For workspaces: updates sidecar manifest for fast lookup

### 4.5 DataSource Read Workflow

**On Read** (arch-v1.md L764-767):
1. Read tool calls `sys:datasources.getTaints(path)`
2. Security service immediately attaches those taints to resulting `tool_result` or injected ContextBlock Message
3. Existing `prepareContextForBoundary` + per-block `taintPolicy` redacts/drops/audits as configured

### 4.6 Isolation Hook

**Isolation Hook** (arch-v1.md L770-771):
- Every agent's persistent workspace is mounted via isolated executor (container/namespace) with exact DataSource policy
- No direct `open()` syscall allowed — all FS ops go through mediated tool layer

---

## 5. Requirements & Constraints

### 5.1 Security Invariant

**Security Invariant** (arch-v1.md L276):
"All data entering the runtime is tainted at the border. No untainted information exists inside the runtime."

This is guaranteed by compile-time type checking: taints are attached as soon as data is ingested, touched, or known about by the application.

### 5.2 Performance Constraints

**Stream Sanitization Latency** (arch-v1.md L300):
- Sanitization must be per-chunk (stateless)
- Latency must stay <50ms

**Bloom Filter Optimization** (arch-v1.md L1798):
- Use Bloom filters for fast-path taint checking

### 5.3 Boundary Constraints

**Boundary Immutability** (arch-v1.md L127, L1783):
- `metadata.boundary` is NEVER migratable (immutable identity)
- Boundary is a permanent identity invariant

**Boundary Transition Rules** (arch-v1.md L290):
- Transitions across boundaries are explicit
- Always go through Mail system (never direct event passing)

### 5.4 TaintPolicy Constraints

**Enforcement Modes** (arch-v1.md L209, L916):
- `strict` - Block forbidden taints entirely
- `redact` - Replace forbidden values with replacement text
- `audit` - Allow but log violations

**allowedOnExit** (arch-v1.md L916, L210-213):
- Defines which taint categories are allowed to leave the runtime
- Enforced on all outbound data paths: Mail, stream chunks, snapshots, sub-agent returns
- Taints not in this list are blocked/redacted at boundaries

---

## 6. YAML Configuration Schemas

### 6.1 ChartDefinition TaintPolicy

**Schema** (arch-v1.md L914-919):

```yaml
taintPolicy:
  enforcement: strict | redact | audit
  allowedOnExit: ["PII", "SECRET", "TOOL_OUTPUT"]  # taints allowed to leave
  redactRules:
    - taint: "SECRET"
      replacement: "[REDACTED]"
```

### 6.2 ContextBlock TaintPolicy

**Schema** (arch-v1.md L216-223):

```yaml
taintPolicy:              # per-block security
  redactMode: redact | dropBlock | audit
  redactRules:
    - taint: string
      replacement: string
  allowedForBoundary: [inner | dmz | outer]
```

### 6.3 DataSource YAML Declaration

**Schema** (arch-v1.md L745-760):

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
```

### 6.4 TaintPolicy Fields

**Fields** (arch-v1.md L173-176, L216-223):
- **redactMode**: `redact` | `dropBlock` | `audit`
- **redactRules**: Array of `{taint: string, replacement: string}`
- **allowedForBoundary**: List of boundaries where this block is allowed (`[inner | dmz | outer]`)

---

## 7. File Structure (Target)

**Reference**: layer-04-security-datasources.md L342-359

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

## 8. Ambiguities & Gaps

### 8.1 Open Questions (from layer-04-security-datasources.md L489-493)

1. **Bloom filters**: Use bloom filters for taint optimization, or simple sets?
2. **DataSource persistence**: Store taints in xattr, sidecar, or object tags?
3. **Taint inheritance**: How to handle nested data structures?
4. **Audit logging**: Where to log taint violations (sys:observability)?

### 8.2 Missing Type Definitions (from implementation-scope.md)

- `IsolatedView` - Used in Security interface but never defined
- `TaintMap` - Referenced but structure not fully specified

### 8.3 Potential Gaps

| Gap | Description | Severity |
|-----|-------------|----------|
| IsolatedView type | Used in Security interface but structure undefined | HIGH |
| Error handling | No spec for error handling in DataSource operations | MEDIUM |
| TaintMap structure | Referenced but detailed structure not specified | MEDIUM |
| Bloom filter implementation | Optional optimization mentioned but not detailed | LOW |

---

## 9. Cross-Reference Summary

### 9.1 arch-v1.md Section Mapping

| Topic | Section | Line Range |
|-------|---------|------------|
| Boundary Model | 4.1 | L261-270 |
| Data Tainting | 4.2 | L271-287 |
| Boundary Enforcement | 4.3 | L288-302 |
| Core Services | 7.1 | L453-461 |
| Platform Service Contract | 7.3 | L477-491 |
| DataSource Abstraction | 11 | L739-798 |
| Security Interface | 14.2 | L1303-1322 |
| Security SHALL Statements | 16.1 | L1513-1526 |

### 9.2 Integration with Other Layers

| Layer | Integration Point |
|-------|-------------------|
| Layer 0 | ApplicationContext with taint-aware Get/Set |
| Layer 2 | sys:security core service |
| Layer 3 | Mail system for violation reports |
| Layer 5 | sys:datasources YAML service |
| Layer 6 | Tool boundary filtering |
| Layer 7 | ContextMap assembly with taint filtering |

---

## 10. Test Specification Summary

### 10.1 Unit Tests (from layer-04-security-datasources.md)

**Phase 4.1: Boundary Types**
- TestBoundaryTypes
- TestBoundaryTransitions

**Phase 4.2: Taint Engine**
- TestTaintEngine_Attach
- TestTaintEngine_Propagate
- TestTaintEngine_CheckPolicy

**Phase 4.3: ContextMap Filtering**
- TestContextMap_FilterTaints
- TestContextMap_Redact
- TestContextMap_DropBlock

**Phase 4.4: DataSource Interface**
- TestDataSource_TagOnWrite
- TestDataSource_GetTaints
- TestDataSource_ValidateAccess

**Phase 4.5: DataSource Implementations**
- TestLocalDiskDataSource
- TestObjectStorageDataSource
- TestInMemoryDataSource

**Phase 4.6: SecurityService**
- TestSecurityService_HandleMail
- TestSecurityService_TaintPropagate
- TestSecurityService_PrepareContextForBoundary

**Phase 4.7: Integration**
- TestFullSecurityFlow
- TestDataSourceIntegration

---

## Appendix A: Complete Security Interface

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

## Appendix B: Complete DataSource Interface

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

## Appendix C: Taint Categories

| Category | Description | Source |
|----------|-------------|--------|
| PII | Personally Identifiable Information | arch-v1.md L277 |
| SECRET | API keys, credentials | arch-v1.md L277 |
| INNER_ONLY | Internal-only data | arch-v1.md L277 |
| USER_SUPPLIED | User-provided input | arch-v1.md L277 |
| TOOL_OUTPUT | Tool-generated output | arch-v1.md L58, L563 |
| EXTERNAL | External data sources | arch-v1.md L563 |
| WORKSPACE | Agent workspace data | arch-v1.md L58, L753 |

---

*End of Layer 4 Specification Extraction*
