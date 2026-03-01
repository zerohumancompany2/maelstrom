# Security Layer Implementation Plan

**Date**: 2026-03-01  
**Status**: Planning Phase  
**Based on**: arch-v1.md Section 4 (Security & Boundaries), Section 11 (DataSources), Section 14.2 (API Contracts)

---

## Executive Summary

This plan implements **Layer 4: Security & Boundaries** including ContextMap assembly, Taint Engine, Boundary Enforcement, and DataSource system. This layer enforces the critical invariant: **all data entering the runtime is tainted at the border**.

**Current State**: Phase 1 complete with 121 tests passing (mail system + 4 core services).

**Target**: Security layer operational with 22 new tests, bringing total to **143 tests**.

---

## Architectural Dependencies

From arch-v1.md Section 4 and Section 11:

```
Security Layer Dependencies:
1. Layer 0: Statechart Engine ✓ COMPLETE
2. Layer 2: Core Services (sys:security, sys:communication, etc.) ✓ COMPLETE
3. Layer 3: Mail System ✓ COMPLETE

New Components:
- ContextMap: Assemble blocks from 4 sources (static, session, memory, tool)
- TaintEngine: Mark data on ingestion/read, propagate, detect violations
- BoundaryService: Enforce inner/DMZ/outer, ValidateAndSanitize(mail)
- DataSource: Pluggable storage interface (localDisk with xattr)

Key Invariant: "All data entering the runtime is tainted at the border"
- No untainted info inside runtime (compile-time guarantee via types)
- Tainting happens as soon as application ingests/touches/knows about info
- ContextMap filters blocks by boundary before LLM calls
```

---

## Phase 3a: ContextMap Assembly (5 tests)

**Goal**: Assemble ContextMap from 4 sources with priority ordering and boundary filtering

**Dependencies**: None (standalone assembly logic)

### Components

| Component        | File                         | Purpose                             | Lines |
| ---------------- | ---------------------------- | ----------------------------------- | ----- |
| ContextBlock     | `pkg/security/contextmap.go` | Block definition with source config | ~60   |
| ContextMap       | `pkg/security/contextmap.go` | Assembly logic, 4 sources           | ~150  |
| Source Factories | `pkg/security/contextmap.go` | static, session, memory, tool       | ~100  |

### ContextBlock Type

```go
type ContextBlock struct {
    Name          string        // identifier
    Source        string        // static|session|memory|tool
    Content       string        // for static source
    Strategy      string        // lastN|full|summarize|RAG
    N             int           // for lastN strategy
    MaxTokens     int           // hard cap
    Eviction      string        // summarize|truncate|error
    StreamCommit  bool          // false = partials never enter next prompt
    TaintPolicy   TaintPolicy   // per-block security policy
    BoundaryFilter BoundaryType  // only tools visible at this boundary
}

type TaintPolicy struct {
    RedactMode      string   // redact|dropBlock|audit
    RedactRules     []RedactRule
    AllowedForBoundary []BoundaryType
}

type RedactRule struct {
    Taint       string
    Replacement string
}
```

### Test Specifications

| Test                           | Semantics Covered                                   |
| ------------------------------ | --------------------------------------------------- |
| `TestContextMap_AssemblyOrder` | Blocks assembled by priority (lower number = first) |
| `TestContextMap_SourceStatic`  | Static content injected directly                    |
| `TestContextMap_SourceSession` | LastN/full session retrieval from session store     |
| `TestContextMap_SourceMemory`  | Vector DB query via memory service                  |
| `TestContextMap_SourceTool`    | Tool registry filtered by caller boundary           |

### Implementation Order

1. `TestContextMap_AssemblyOrder` → Define `ContextBlock`, `ContextMap`, `assemble()` sorts by priority
2. `TestContextMap_SourceStatic` → `staticSource()` returns content directly
3. `TestContextMap_SourceSession` → `sessionSource()` queries session store for lastN/full
4. `TestContextMap_SourceMemory` → `memorySource()` calls memory service with query
5. `TestContextMap_SourceTool` → `toolSource()` filters tools by BoundaryType

**Commit**: `feat: contextmap assembly with 4 sources`

---

## Phase 3b: Taint Engine (6 tests)

**Goal**: Mark data on ingestion/read, propagate taints, detect violations, apply redaction

**Dependencies**: ContextMap (for taint filtering), Mail types (for taint metadata)

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| TaintSet | `pkg/security/taint.go` | Map[string]bool for taint tracking | ~40 |
| TaintEngine | `pkg/security/taint.go` | Mark, Propagate, Check, Redact | ~150 |
| TaintPolicy | `pkg/security/taint.go` | Enforcement mode, allowed taints | ~60 |
| RedactRules | `pkg/security/rules.go` | Apply redaction to data | ~80 |

### TaintEngine API

```go
type TaintEngine interface {
    Mark(obj any, taints []string) (any, error)          // Mark on ingestion
    MarkRead(path string, taints []string) ([]string, error)  // Mark on read
    Propagate(obj any, newTaints []string) (any, error)   // Copy with taints
    CheckForbidden(taints []string, boundary BoundaryType) error
    Redact(obj any, rules []RedactRule) (any, error)
    ReportTaints(chartID string) (TaintMap, error)
}

type TaintSet map[string]bool

func (t TaintSet) Add(taint string)
func (t TaintSet) Has(taint string) bool
func (t TaintSet) Union(other TaintSet) TaintSet

type TaintPolicy struct {
    Enforcement     string   // strict|redact|audit
    AllowedOnExit   []string // taints allowed to leave boundary
    RedactRules     []RedactRule
}
```

### Test Specifications

| Test                               | Semantics Covered                                   |
| ---------------------------------- | --------------------------------------------------- |
| `TestTaintEngine_ProactiveMarking` | Mark data on ingestion (create with taints)         |
| `TestTaintEngine_ReactiveMarking`  | Mark data on read (read file, get taints)           |
| `TestTaintEngine_Propagation`      | Copy object with taints preserved                   |
| `TestTaintEngine_Violation`        | Detect forbidden taints for boundary                |
| `TestTaintEngine_Redaction`        | Apply redaction rules (replace PII with [REDACTED]) |
| `TestTaintEngine_ReportTaints`     | Report per-chart taint map                          |

### Implementation Order

1. `TestTaintEngine_ProactiveMarking` → `Mark(obj, taints)` creates copy with taints attached
2. `TestTaintEngine_ReactiveMarking` → `MarkRead(path, taints)` reads file, returns taints
3. `TestTaintEngine_Propagation` → `Propagate(obj, newTaints)` merges taint sets
4. `TestTaintEngine_Violation` → `CheckForbidden(taints, boundary)` returns error if forbidden
5. `TestTaintEngine_Redaction` → `Redact(obj, rules)` recursively replaces tainted values
6. `TestTaintEngine_ReportTaints` → `ReportTaints(chartID)` returns TaintMap for chart

**Commit**: `feat: taint engine with marking, propagation, redaction`

---

## Phase 3c: Boundary Enforcement (4 tests)

**Goal**: Enforce inner/DMZ/outer boundaries on mail and charts

**Dependencies**: TaintEngine (for violation checks), Mail types (for validation)

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| BoundaryService | `pkg/security/boundary.go` | ValidateAndSanitize, NamespaceIsolate | ~150 |
| BoundaryType | `pkg/security/types.go` | inner|dmz|outer enum | ~30 |
| IsolatedView | `pkg/security/types.go` | NamespaceIsolate result | ~40 |

### BoundaryService API

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

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestBoundaryService_Enforcement` | Inner/DMZ/outer boundary checks work |
| `TestBoundaryService_MailValidate` | ValidateAndSanitize(mail) filters taints |
| `TestBoundaryService_ChartCheck` | ChartRuntime boundary validation |
| `TestBoundaryService_Namespace` | NamespaceIsolate creates isolated view |

### Implementation Order

1. `TestBoundaryService_Enforcement` → Define `BoundaryType`, implement basic checks
2. `TestBoundaryService_MailValidate` → `ValidateAndSanitize(mail, src, tgt)` filters taints
3. `TestBoundaryService_ChartCheck` → `CheckChart(chartID, boundary)` validates chart boundary
4. `TestBoundaryService_Namespace` → `NamespaceIsolate(chartID, operation)` returns IsolatedView

**Commit**: `feat: boundary enforcement for mail and charts`

---

## Phase 3d: DataSource System (4 tests)

**Goal**: Pluggable storage interface with xattr tagging for localDisk

**Dependencies**: BoundaryType (for access validation)

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| DataSource interface | `pkg/datasource/datasource.go` | TagOnWrite, GetTaints, ValidateAccess | ~50 |
| Registry | `pkg/datasource/datasource.go` | Register(), Get() for plugin system | ~60 |
| localDisk | `pkg/datasource/local_disk.go` | xattr tagging + sidecar fallback | ~100 |
| Registration | `pkg/bootstrap/bootstrap.go` | Explicit registration in bootstrap | ~30 |

### DataSource Interface

```go
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary BoundaryType) error
}

// Registration (explicit in bootstrap)
func Register(name string, factory func(config map[string]any) (DataSource, error))
func Get(name string, config map[string]any) (DataSource, error)
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestDataSource_Register` | Plugin registration via Register() |
| `TestDataSource_LocalDisk` | xattr tagging on write |
| `TestDataSource_GetTaints` | Read taints from xattr |
| `TestDataSource_ValidateAccess` | Boundary checks on access |

### Implementation Order

1. `TestDataSource_Register` → Define `Register()`, `Get()`, registry map
2. `TestDataSource_LocalDisk` → `localDisk.TagOnWrite()` sets xattr
3. `TestDataSource_GetTaints` → `localDisk.GetTaints()` reads xattr
4. `TestDataSource_ValidateAccess` → `localDisk.ValidateAccess()` checks boundary

**Commit**: `feat: DataSource system with pluggable localDisk`

---

## Phase 3e: Integration (3 tests)

**Goal**: Wire Security into Kernel bootstrap and validate end-to-end

**Dependencies**: All previous phases

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestSecurity_KernelBootstrap` | Kernel registers MarkTaint, ValidateAndSanitize actions/guards |
| `TestSecurity_ContextMapFiltering` | ContextMap.Assemble() filters by boundary |
| `TestSecurity_MailValidation` | Mail validated before delivery to agents |

### Implementation Order

1. `TestSecurity_KernelBootstrap` → Kernel registers security actions/guards in init()
2. `TestSecurity_ContextMapFiltering` → ContextMap.Assemble() calls CheckForbidden
3. `TestSecurity_MailValidation` → Mail validated via ValidateAndSanitize before delivery

**Commit**: `feat: Security integration with Kernel bootstrap`

---

## File Structure

```
pkg/security/
├── contextmap.go      # ContextBlock, ContextMap, 4 source factories, assembly
├── taint.go           # TaintSet, TaintPolicy, Mark, Propagate, Check, Redact
├── boundary.go        # BoundaryService interface + implementation
├── types.go           # BoundaryType, TaintMap, IsolatedView, MailMetadata
├── rules.go           # RedactRule, validation logic
└── security_test.go   # 18 tests (5+6+4+3)

pkg/datasource/
├── datasource.go      # DataSource interface, registry, Register()
├── local_disk.go      # localDisk implementation (xattr + sidecar fallback)
└── datasource_test.go # 4 tests

pkg/bootstrap/
└── bootstrap.go       # Register data sources explicitly (append)

pkg/kernel/
└── kernel_test.go     # 3 integration tests (append to existing)
```

---

## Total Test Count

- **ContextMap**: 5 tests
- **Taint Engine**: 6 tests
- **Boundary Enforcement**: 4 tests
- **DataSource**: 4 tests
- **Integration**: 3 tests
- **Total**: 22 new tests

**Current Total**: 121 (existing) + 22 (new) = **143 tests**

---

## Implementation Order (TDD Workflow)

### Phase 3a: ContextMap (Day 1)
1. `TestContextMap_AssemblyOrder` → Define types, assemble()
2. `TestContextMap_SourceStatic` → staticSource()
3. `TestContextMap_SourceSession` → sessionSource()
4. `TestContextMap_SourceMemory` → memorySource()
5. `TestContextMap_SourceTool` → toolSource()
6. Commit: `feat: contextmap assembly with 4 sources`

### Phase 3b: Taint Engine (Day 2)
1. `TestTaintEngine_ProactiveMarking` → Mark()
2. `TestTaintEngine_ReactiveMarking` → MarkRead()
3. `TestTaintEngine_Propagation` → Propagate()
4. `TestTaintEngine_Violation` → CheckForbidden()
5. `TestTaintEngine_Redaction` → Redact()
6. `TestTaintEngine_ReportTaints` → ReportTaints()
7. Commit: `feat: taint engine with marking, propagation, redaction`

### Phase 3c: Boundary Enforcement (Day 3)
1. `TestBoundaryService_Enforcement` → BoundaryType, basic checks
2. `TestBoundaryService_MailValidate` → ValidateAndSanitize()
3. `TestBoundaryService_ChartCheck` → CheckChart()
4. `TestBoundaryService_Namespace` → NamespaceIsolate()
5. Commit: `feat: boundary enforcement for mail and charts`

### Phase 3d: DataSource (Day 4)
1. `TestDataSource_Register` → Register(), Get()
2. `TestDataSource_LocalDisk` → TagOnWrite()
3. `TestDataSource_GetTaints` → GetTaints()
4. `TestDataSource_ValidateAccess` → ValidateAccess()
5. Commit: `feat: DataSource system with pluggable localDisk`

### Phase 3e: Integration (Day 5)
1. `TestSecurity_KernelBootstrap` → Register actions/guards
2. `TestSecurity_ContextMapFiltering` → Filter by boundary
3. `TestSecurity_MailValidation` → Validate before delivery
4. Commit: `feat: Security integration with Kernel bootstrap`

---

## Key Design Decisions

### 1. DataSource Registration (Explicit in Bootstrap)

**Why Explicit?**
- Clear dependency graph: Kernel knows what data sources exist
- Testable: Easy to mock in unit tests
- Configurable: Bootstrap YAML can specify which sources to register

**Implementation**:
```go
// pkg/bootstrap/bootstrap.go
func init() {
    datasource.Register("localDisk", func(config map[string]any) (datasource.DataSource, error) {
        return &localDisk{
            path:    config["path"].(string),
            xattrNS: config["xattrNamespace"].(string),
        }, nil
    })
}
```

### 2. Taint Invariant (Compile-Time Guarantee)

**Why No Runtime Checks?**
- Compile-time type system enforces tainted vs untainted
- Runtime checks add overhead without safety benefit
- Invariant guaranteed by design: all data tainted at border

**Implementation**:
- `Mark()` attaches taints when data enters runtime
- `MarkRead()` attaches taints when data read from storage
- No data exists without taints inside runtime

### 3. ContextMap All 4 Sources

**Why All 4?**
- Static: System prompts, constants
- Session: LastN/full conversation history
- Memory: Vector DB queries for long-term memory
- Tool: Tool registry filtered by boundary

**Implementation**:
- Each source is a factory function
- Assemble() calls all sources in priority order
- Boundary filtering applied per-block via TaintPolicy

### 4. Test Count (22 Tests)

**Why 22 Tests?**
- 1:1 semantics-to-tests ratio maintained
- ContextMap (5) covers all 4 sources + assembly
- Taint Engine (6) covers marking, propagation, violation, redaction
- Boundary (4) covers enforcement, mail, chart, namespace
- DataSource (4) covers registration, localDisk, taints, access
- Integration (3) covers kernel, contextmap, mail

**Acceptable**: Yes, matches Phase 1's 27 tests for similar scope.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| ContextMap too complex | Low | Medium | Keep assembly simple, defer advanced features |
| Taint propagation bugs | Medium | High | Extensive unit tests for Mark, Propagate |
| Boundary enforcement gaps | Medium | High | Integration tests for mail + chart validation |
| DataSource plug-in issues | Low | Medium | Clean interface, explicit registration |
| Test coverage gaps | Medium | Medium | Track semantics vs tests (1:1 ratio target) |

---

## Success Criteria

1. ✅ ContextMap assembles from all 4 sources (5 tests pass)
2. ✅ Taint Engine marks, propagates, detects violations (6 tests pass)
3. ✅ Boundary Service enforces inner/DMZ/outer (4 tests pass)
4. ✅ DataSource pluggable with localDisk (4 tests pass)
5. ✅ Security integrated with Kernel bootstrap (3 tests pass)
6. ✅ 1:1 semantics-to-tests ratio maintained
7. ✅ All 22 new tests pass + 121 existing tests pass = 143 total

---

## Questions for Review

1. **ContextMap Sources**: All 4 sources (static, session, memory, tool) implemented - confirmed?
2. **DataSource Registration**: Explicit in bootstrap via Register() calls - confirmed?
3. **Taint Invariant**: Compile-time guarantee via types, no runtime checks - confirmed?
4. **Test Count**: 22 tests total - acceptable given semantics covered?
5. **Integration Tests**: Should they live in `pkg/kernel/` or `pkg/security/`?

---

## Deferred Features

### Phase 4: Remaining Platform Services (Layer 5)
- YAML-defined services loadable via ChartRegistry
- sys:gateway, sys:admin, sys:persistence, sys:heartbeat
- sys:memory, sys:tools, sys:datasources, sys:human-gateway

**Milestone**: Full platform operational.

### Phase 5: Execution (Layer 6)
- Tool registry
- Orchestrator sub-chart
- Execution policies (seq_failfast, seq_continue, par_continue)

**Milestone**: Can execute tools and sub-agents.

### Phase 6: Agents (Layer 7)
- Agent Charts (OODA loop)
- LLM bridge
- ContextMap assembly for LLM calls

**Milestone**: Agents can reason and use tools.

### Phase 7: External Interface (Layer 8)
- HTTP gateway
- SSE/WS streaming
- 2FA for admin

**Milestone**: External clients can interact with agents.