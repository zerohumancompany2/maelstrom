# Layer 4 Implementation Assessment

**Assessment Date**: 2026-03-09  
**Scope**: Security Service, Taint Engine, DataSource, Boundary Enforcement, ContextMap/TaintPolicy  
**Reference Spec**: `docs/layer-4/spec-extraction.md` (derived from `docs/arch-v1.md` v1.3)

---

## Executive Summary

Layer 4 implementation is **PARTIALLY COMPLETE** with foundational types and interfaces defined, but most enforcement logic remains stubbed or unimplemented. The codebase has:

- ✅ Core type definitions (BoundaryType, TaintPolicy, ContextBlock)
- ✅ DataSource interface and registry pattern
- ✅ Security service skeleton with `sys:security` ID
- ✅ Taint engine interface with stub implementations
- ❌ No actual boundary enforcement logic
- ❌ No taint tracking/propagation implementation
- ❌ No xattr-based taint storage for local disk
- ❌ No ContextMap filtering/redaction implementation
- ❌ No `allowedOnExit` policy enforcement

---

## 1. What Exists

### 1.1 Security Service Types

**File**: `/home/albert/git/maelstrom/pkg/security/types.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `BoundaryType` enum (inner/dmz/outer) | L3-9 | ✅ Complete |
| `ContextBlock` struct with TaintPolicy | L11-23 | ✅ Complete |
| `TaintPolicy` struct | L25-29 | ✅ Complete |
| `RedactRule` struct | L31-34 | ✅ Complete |
| `TaintSet` type with Add/Has/Union | L36-58 | ✅ Complete |
| `TaintMap` type with Add/Merge | L60-70 | ✅ Complete |
| `TaintEngine` interface | L152-159 | ✅ Defined |
| `taintEngineImpl` struct | L161-163 | ✅ Defined |
| `NewTaintEngine()` constructor | L165-169 | ✅ Implemented |
| `BoundaryService` interface | L195-202 | ✅ Defined |
| `IsolatedView` struct | L204-209 | ✅ Defined |
| `boundaryServiceImpl` struct | L211-213 | ✅ Defined |
| `NewBoundaryService()` constructor | L215-218 | ✅ Implemented |

**ContextMap Helper Functions** (pkg/security/types.go):
- `assemble()` - Priority-based sorting (L72-85) ✅
- `staticSource()` - Static content retrieval (L87-90) ✅
- `sessionSource()` - Session message retrieval (L91-100) ✅
- `memorySource()` - Memory query stub (L102-105) ✅
- `toolSource()` - Tool filtering with boundary visibility (L107-120) ✅
- `getToolBoundary()` - Tool boundary determination (L122-127) ✅
- `isBoundaryVisible()` - Boundary visibility check (L128-140) ✅

### 1.2 Security Service Implementation

**File**: `/home/albert/git/maelstrom/pkg/services/security/service.go`

| Method | Lines | Implementation Status |
|--------|-------|----------------------|
| `NewSecurityService()` | L14-16 | ✅ Returns initialized struct |
| `ID()` | L18-20 | ✅ Returns "sys:security" |
| `HandleMail()` | L22-24 | ⚠️ Stub (returns nil) |
| `ValidateBoundary()` | L26-28 | ⚠️ Stub (returns nil) |
| `ValidateAndSanitize()` | L30-47 | ⚠️ Partial - updates boundary, adds EXTERNAL taint on outer→inner |
| `TaintPropagate()` | L49-96 | ✅ Partial - propagates taints to nested maps/slices via `_taints` key |
| `ReportTaints()` | L98-100 | ⚠️ Stub (returns empty map) |
| `PrepareContextForBoundary()` | L102-104 | ⚠️ Stub (returns nil) |
| `CheckTaintPolicy()` | L106-113 | ⚠️ Partial - only blocks "UNTRUSTED" taint |
| `NamespaceIsolate()` | L115-134 | ✅ Filters data by agentID |
| `Start()` | L136-138 | ⚠️ Stub |
| `Stop()` | L140-142 | ⚠️ Stub |

**File**: `/home/albert/git/maelstrom/pkg/services/security/bootstrap.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `BootstrapChart()` | L5-10 | ✅ Returns minimal ChartDefinition with ID "sys:security" |

### 1.3 DataSource Interface

**File**: `/home/albert/git/maelstrom/pkg/datasource/datasource.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `DataSource` interface | L8-12 | ✅ Complete (TagOnWrite, GetTaints, ValidateAccess) |
| `Registry` struct | L14-16 | ✅ Complete |
| `NewRegistry()` | L19-23 | ✅ Complete |
| `Registry.Register()` | L25-29 | ✅ Complete |
| `Registry.Get()` | L31-41 | ✅ Complete |
| `Registry.List()` | L43-51 | ✅ Complete |
| `globalRegistry` singleton | L53-69 | ✅ Complete with init() |

### 1.4 DataSource Implementations

**File**: `/home/albert/git/maelstrom/pkg/datasource/local_disk.go`

| Method | Lines | Implementation Status |
|--------|-------|----------------------|
| `localDisk` struct | L9-11 | ✅ Defined with path and xattrNamespace |
| `NewLocalDisk()` | L14-25 | ✅ Parses config, returns initialized struct |
| `TagOnWrite()` | L27-38 | ⚠️ Creates file/dir but does NOT set xattr taints |
| `GetTaints()` | L40-43 | ❌ Returns empty array (no xattr read) |
| `ValidateAccess()` | L44-46 | ❌ Returns nil (no boundary validation) |
| `init()` registration | L48-50 | ✅ Registers "localDisk" factory |

**File**: `/home/albert/git/maelstrom/pkg/services/datasources/service.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `DatasourceService` interface | L8-15 | ✅ Complete |
| `datasourceService` struct | L17-19 | ✅ Complete |
| `NewDatasourceService()` | L21-25 | ✅ Complete |
| `Get()` | L27-33 | ✅ Complete |
| `List()` | L35-41 | ✅ Complete |
| `TagOnWrite()` | L43-45 | ❌ Stub (returns nil) |
| `GetTaints()` | L47-49 | ❌ Stub (returns empty array) |
| `ValidateAccess()` | L51-53 | ❌ Stub (returns nil) |
| `Register()` | L55-58 | ✅ Complete |
| `LocalDiskDatasource` struct | L60 | ✅ Defined |
| `S3Datasource` struct | L74 | ✅ Defined (all methods stubbed) |

### 1.5 Mail System Integration

**File**: `/home/albert/git/maelstrom/pkg/mail/types.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `MailTypeTaintViolation` | L29, L42 | ✅ Defined |
| `MailMetadata.Boundary` | L49 | ✅ Field exists |
| `MailMetadata.Taints` | L50 | ✅ Field exists |
| `BoundaryType` constants | L56-62 | ✅ Complete (inner/dmz/outer) |
| `StreamChunk.Taints` | L68 | ✅ Field exists |

**File**: `/home/albert/git/maelstrom/pkg/mail/stream.go`

| Method | Lines | Status |
|--------|-------|--------|
| `StripForbiddenTaints()` | L59-79 | ✅ Filters taints against allowed list |
| `stripTaints()` | L81-97 | ✅ Internal method, hardcodes USER_SUPPLIED/TOOL_OUTPUT |
| `propagateTaints()` | L99-114 | ✅ Merges new taints into chunk |
| `checkSecurityBoundary()` | L116-122 | ⚠️ Partial - only blocks DMZ→Inner transition |

### 1.6 Kernel Integration

**File**: `/home/albert/git/maelstrom/pkg/kernel/kernel.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `kernelApplicationContext.Get()` | L55-66 | ✅ Returns value, taints (always nil), error |
| `kernelApplicationContext.Set()` | L68-73 | ✅ Accepts taints parameter but does NOT store them |
| `sys:security` in service order | L459 | ✅ Listed first in bootstrap |
| Service ready tracking | L88-93, L388-394 | ✅ Tracks sys:security readiness |

### 1.7 Bootstrap Actions

**File**: `/home/albert/git/maelstrom/pkg/bootstrap/actions.go`

| Action | Lines | Status |
|--------|-------|--------|
| `LoadSecurityService()` | L146-180 | ✅ Spawns security chart, stores RTID |
| Security service integration | L11 | ✅ Imports security package |

### 1.8 Human Gateway ContextMap

**File**: `/home/albert/git/maelstrom/pkg/services/humangateway/chat.go`

| Component | Lines | Status |
|-----------|-------|--------|
| `ContextMapSnapshot` type | L14 | ✅ Defined as map[string]any |
| `ChatSession.ContextMap` | L11 | ✅ Field exists |
| `SanitizeContextForBoundary()` | L45-51 | ❌ Stub - copies all keys without filtering |

### 1.9 ApplicationContext Interface

**File**: `/home/albert/git/maelstrom/pkg/statechart/interfaces.go`

| Method | Lines | Status |
|--------|-------|--------|
| `Get(key, callerBoundary)` | L8 | ✅ Signature includes taints return |
| `Set(key, value, taints, callerBoundary)` | L12 | ✅ Signature includes taints parameter |

---

## 2. What's Missing

### 2.1 Security Service Missing Methods

Per spec-extraction.md L38-45, the `Security` interface should have:

| Required Method | Spec Lines | Current Status |
|-----------------|------------|----------------|
| `ValidateAndSanitize(mail Mail, sourceBoundary, targetBoundary BoundaryType) (Mail, error)` | L39 | ⚠️ Exists but uses `mail.Mail` not generic `any` |
| `TaintPropagate(obj any, newTaints []string) (any, error)` | L40 | ✅ Exists |
| `NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)` | L41 | ⚠️ Exists but signature differs (takes `data interface{}, agentID string`) |
| `CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)` | L42 | ❌ Wrong signature - current: `(taints []string, action string) bool` |
| `ReportTaints(runtimeId string) (TaintMap, error)` | L43 | ✅ Exists |
| `PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error` | L44 | ✅ Exists (stubbed) |

### 2.2 Taint Engine Missing Implementation

Per spec-extraction.md L152-159, `TaintEngine` interface methods are stubbed:

| Method | File:Line | Missing Implementation |
|--------|-----------|------------------------|
| `Mark(obj, taints)` | types.go:L171-173 | Does not attach taints to object |
| `MarkRead(path, taints)` | types.go:L175-177 | Returns taints but does not persist |
| `Propagate(obj, newTaints)` | types.go:L179-181 | Does not merge taints |
| `CheckForbidden(taints, boundary)` | types.go:L183-185 | Always returns nil (no enforcement) |
| `Redact(obj, rules)` | types.go:L187-189 | Does not apply redaction rules |
| `ReportTaints(chartID)` | types.go:L191-193 | Returns empty map (no tracking) |

### 2.3 DataSource Missing Implementation

Per spec-extraction.md L787-793, `DataSource` implementations are incomplete:

| Feature | Spec Reference | Current Status |
|---------|----------------|----------------|
| xattr-based taint storage | L757-759, L796 | ❌ Not implemented in local_disk.go |
| xattr namespace config | L754 | ⚠️ Field exists but not used |
| Sidecar fallback | L759 | ❌ Not implemented |
| Object storage (S3/GCS) tags | L758 | ❌ S3Datasource stubbed |
| inMemoryWorkspace | L796 | ❌ Not implemented |
| Taint inheritance on read | L764-767 | ❌ GetTaints returns empty |

### 2.4 Boundary Enforcement Missing

Per spec-extraction.md L261-270, L290-296:

| Feature | Spec Reference | Current Status |
|---------|----------------|----------------|
| Inner→DMZ/outer auto-strip | L284 | ❌ Not implemented |
| Block on forbidden taints | L284 | ❌ Not implemented |
| Mail-only boundary transitions | L290 | ⚠️ Partial check in stream.go:L116-122 |
| `prepareContextForBoundary` enforcement | L292-293 | ❌ Stubbed, no filtering |
| Per-block taintPolicy enforcement | L295-296 | ❌ Not implemented |
| `taint_violation` event emission | L286, L170 | ⚠️ Type exists, no emission logic |

### 2.5 ContextMap Filtering Missing

Per spec-extraction.md L292-296, L205-223:

| Feature | Spec Reference | Current Status |
|---------|----------------|----------------|
| Redact forbidden taints before LLM | L295 | ❌ SanitizeContextForBoundary is no-op |
| Drop blocks with forbidden taints | L295 | ❌ Not implemented |
| Per-block taintPolicy overrides | L296 | ❌ Not implemented |
| `redactMode: redact\|dropBlock\|audit` | L216-223 | ⚠️ Field exists, not enforced |
| `allowedForBoundary` enforcement | L223 | ⚠️ Field exists, not enforced |

### 2.6 Stream Sanitization Missing

Per spec-extraction.md L298-301:

| Feature | Spec Reference | Current Status |
|---------|----------------|----------------|
| PII redaction per-chunk | L298 | ❌ Not implemented |
| Length caps | L298 | ❌ Not implemented |
| Schema validation | L298 | ❌ Not implemented |
| Inner-data stripping | L298 | ❌ Not implemented |
| <50ms latency guarantee | L300 | ⚠️ Not measured |

### 2.7 TaintPolicy Enforcement Missing

Per spec-extraction.md L209-213, L914-919:

| Feature | Spec Reference | Current Status |
|---------|----------------|----------------|
| `enforcement: strict\|redact\|audit` | L209, L916 | ⚠️ Field not in TaintPolicy struct |
| `allowedOnExit` policy | L210-213, L917 | ❌ Not implemented anywhere |
| Redact rules application | L918-920 | ⚠️ RedactRule exists, not applied |

---

## 3. Partially Implemented Features

### 3.1 Taint Propagation

**File**: `/home/albert/git/maelstrom/pkg/services/security/service.go:L49-96`

**What Works**:
- Recursively propagates taints to nested `map[string]interface{}` structures
- Handles slices containing maps
- Uses `_taints` key to store taint array

**What's Missing**:
- Does not handle non-map objects (strings, structs, etc.)
- Does not deduplicate taints
- Does not enforce any policy during propagation
- Taints are stored in `_taints` key but spec does not define this format

**Spec Reference**: spec-extraction.md L283 (taint propagation on copy/read/write)

### 3.2 Boundary Validation

**File**: `/home/albert/git/maelstrom/pkg/services/security/service.go:L30-47`

**What Works**:
- Updates `mail.Metadata.Boundary` to target boundary
- Adds "EXTERNAL" taint on outer→inner transitions

**What's Missing**:
- No blocking of forbidden transitions (e.g., outer→inner with PII)
- No stripping of forbidden taints on inner→outer
- No validation of `allowedOnExit` policy
- DMZ boundary transitions not handled

**Spec Reference**: spec-extraction.md L284 (auto-strip or block on forbidden taints)

### 3.3 Boundary Visibility Check

**File**: `/home/albert/git/maelstrom/pkg/security/types.go:L128-140`

**What Works**:
- Inner can only see Inner tools
- DMZ can only see DMZ tools
- Outer can only see DMZ tools

**What's Missing**:
- Only used in `toolSource()` helper, not enforced in actual tool invocation
- No integration with sys:tools service

**Spec Reference**: spec-extraction.md L267-269 (boundary capabilities)

### 3.4 Stream Taint Stripping

**File**: `/home/albert/git/maelstrom/pkg/mail/stream.go:L59-97`

**What Works**:
- `StripForbiddenTaints()` filters taints against allowed list
- `propagateTaints()` merges new taints

**What's Missing**:
- `stripTaints()` hardcodes allowed list instead of using policy
- No integration with `allowedOnExit` from ChartDefinition
- Not called in actual stream send path

**Spec Reference**: spec-extraction.md L298-301 (stream sanitization)

### 3.5 TaintPolicy Structure

**File**: `/home/albert/git/maelstrom/pkg/security/types.go:L25-29`

```go
type TaintPolicy struct {
    RedactMode         string
    RedactRules        []RedactRule
    AllowedForBoundary []BoundaryType
}
```

**What's Missing vs Spec** (spec-extraction.md L914-919):
- No `enforcement` field (strict/redact/audit)
- No `allowedOnExit` field
- `AllowedForBoundary` is per-block, should also exist at ChartDefinition level

---

## 4. Gap Analysis: Spec Requirements vs Implementation

### 4.1 Security SHALL Statements (spec-extraction.md L1312-1317)

| SHALL Statement | Spec Line | Implementation Status | Evidence |
|-----------------|-----------|----------------------|----------|
| Attach taints at Message creation and file read | L1312 | ❌ Missing | No taint attachment in mail creation or file read |
| Call `prepareContextForBoundary` inside every `assembleContextMap` | L1313 | ❌ Missing | Function exists but is stubbed; not called anywhere |
| Redact/drop/audit per-block `taintPolicy` before LLM prompt | L1314 | ❌ Missing | SanitizeContextForBoundary is no-op |
| Enforce `allowedOnExit` on Mail, streams, snapshots, sub-agent returns | L1315 | ❌ Missing | allowedOnExit not defined or enforced |
| Use DataSource interface for all FS/object access | L1316 | ⚠️ Partial | Interface exists but localDisk impl doesn't use xattr |

### 4.2 Security SHALL NOT Statements (spec-extraction.md L1320-1322)

| SHALL NOT Statement | Spec Line | Implementation Status | Evidence |
|---------------------|-----------|----------------------|----------|
| Allow inner data to reach outer/DMZ LLM prompts without redaction | L1320 | ⚠️ Not Enforced | No blocking exists; data can flow freely |
| Permit boundary crossing via Events (only Mail) | L1321 | ⚠️ Partial | stream.go:L116-122 blocks DMZ→Inner, but not comprehensive |
| Bypass taint checks on file read operations | L1322 | ❌ Missing | No taint checks exist on file read |

### 4.3 Taint Categories (spec-extraction.md L58-67, L563, L566)

| Category | Spec Reference | Used in Code |
|----------|----------------|--------------|
| PII | L277 | ✅ Referenced in tests |
| SECRET | L277 | ✅ Referenced in tests |
| INNER_ONLY | L277 | ✅ Referenced in tests |
| USER_SUPPLIED | L277 | ✅ Hardcoded in stream.go:L83 |
| TOOL_OUTPUT | L58, L563 | ✅ Hardcoded in stream.go:L83 |
| EXTERNAL | L563 | ✅ Added in service.go:L42 |
| WORKSPACE | L58, L753 | ✅ Referenced in tests |
| UNTRUSTED | - | ✅ Used in CheckTaintPolicy (not in spec) |

### 4.4 DataSource Implementations (spec-extraction.md L796)

| Implementation | Spec Reference | File | Status |
|----------------|----------------|------|--------|
| localDisk | L796 | pkg/datasource/local_disk.go | ⚠️ Partial (no xattr) |
| s3 | L796 | pkg/services/datasources/service.go | ❌ Stub |
| gcs | L796 | - | ❌ Missing |
| inMemoryWorkspace | L796 | - | ❌ Missing |

---

## 5. Dependencies from Layers 1-3

### 5.1 Layer 0 (Statechart Engine) - Available ✅

| Dependency | Spec Reference | File | Status |
|------------|----------------|------|--------|
| ApplicationContext.Get with taints return | L383-387 | pkg/statechart/interfaces.go:L8 | ✅ Interface defined |
| ApplicationContext.Set with taints param | L383-387 | pkg/statechart/interfaces.go:L12 | ✅ Interface defined |
| Library.Spawn | - | pkg/statechart/engine.go | ✅ Available |
| Library.Dispatch | - | pkg/statechart/engine.go | ✅ Available |

### 5.2 Layer 2 (Core Platform Services) - Partially Available ⚠️

| Service | Spec Reference | File | Status |
|---------|----------------|------|--------|
| sys:security | L457 | pkg/services/security/ | ⚠️ Partially implemented |
| sys:communication | L458 | pkg/services/communication/ | ✅ Available (assumed) |
| sys:observability | L459 | pkg/services/observability/ | ✅ Available (assumed) |
| sys:lifecycle | L460 | pkg/services/lifecycle/ | ✅ Available (assumed) |

### 5.3 Layer 3 (Mail System) - Available ✅

| Component | Spec Reference | File | Status |
|-----------|----------------|------|--------|
| Mail with Metadata.Taints | L170 | pkg/mail/types.go:L50 | ✅ Available |
| Mail with Metadata.Boundary | L170 | pkg/mail/types.go:L49 | ✅ Available |
| taint_violation MailType | L170 | pkg/mail/types.go:L29 | ✅ Available |
| StreamChunk with Taints | - | pkg/mail/types.go:L68 | ✅ Available |

---

## 6. Test Coverage Analysis

### 6.1 Existing Tests

| File | Tests | Coverage |
|------|-------|----------|
| pkg/security/security_test.go | 22 tests | Types, ContextMap helpers, stub validations |
| pkg/services/security/service_test.go | 20 tests | Service methods, boundary transitions, taint propagation |
| pkg/datasource/datasource_test.go | 4 tests | Registry, localDisk stubs |
| pkg/services/datasources/service_test.go | 5 tests | Service methods (all pass due to stubs) |
| pkg/mail/stream_test.go | Multiple | Taint stripping, propagation |

### 6.2 Missing Tests (per spec-extraction.md L505-536)

| Test Category | Spec Reference | Status |
|---------------|----------------|--------|
| TestBoundaryTypes | Phase 4.1 | ❌ Missing |
| TestBoundaryTransitions | Phase 4.1 | ⚠️ Partial (service_test.go) |
| TestTaintEngine_Attach | Phase 4.2 | ❌ Missing (engine is stub) |
| TestTaintEngine_Propagate | Phase 4.2 | ⚠️ Partial |
| TestTaintEngine_CheckPolicy | Phase 4.2 | ❌ Missing |
| TestContextMap_FilterTaints | Phase 4.3 | ❌ Missing |
| TestContextMap_Redact | Phase 4.3 | ❌ Missing |
| TestContextMap_DropBlock | Phase 4.3 | ❌ Missing |
| TestDataSource_TagOnWrite (with xattr) | Phase 4.4 | ❌ Missing |
| TestDataSource_GetTaints (with xattr) | Phase 4.4 | ❌ Missing |
| TestLocalDiskDataSource (full) | Phase 4.5 | ⚠️ Partial |
| TestObjectStorageDataSource | Phase 4.5 | ❌ Missing |
| TestInMemoryDataSource | Phase 4.5 | ❌ Missing |

---

## 7. Recommendations

### 7.1 Immediate Priorities

1. **Implement xattr-based taint storage** in `pkg/datasource/local_disk.go`
   - Use `golang.org/x/sys/unix` for GETXATTR/SETXATTR
   - Store taints as JSON in `user.maelstrom.taints` xattr
   - Add sidecar JSON fallback for non-xattr filesystems

2. **Implement boundary enforcement logic** in `pkg/services/security/service.go`
   - Add taint stripping on inner→outer transitions
   - Block transitions with forbidden taint combinations
   - Emit `taint_violation` mail on policy violations

3. **Implement ContextMap filtering** in `pkg/services/humangateway/chat.go`
   - Apply `RedactMode` from TaintPolicy
   - Drop blocks with `allowedForBoundary` not matching target
   - Apply `RedactRule` replacements

4. **Add `allowedOnExit` enforcement**
   - Add field to `TaintPolicy` struct
   - Enforce on all outbound paths (Mail, streams, snapshots)

### 7.2 Medium Priorities

5. **Complete TaintEngine implementation**
   - Implement `Mark()` to attach taints to objects
   - Implement `Redact()` to apply redaction rules
   - Implement `CheckForbidden()` with actual policy enforcement

6. **Implement S3/GCS DataSource backends**
   - Use object metadata/tags for taint storage
   - Follow same interface as localDisk

7. **Implement inMemoryWorkspace DataSource**
   - For agent workspace isolation
   - Fast in-memory taint lookup

### 7.3 Lower Priorities

8. **Bloom filter optimization** (spec-extraction.md L1798)
   - Optional performance optimization
   - Implement after core functionality is complete

9. **Stream sanitization features**
   - PII redaction patterns
   - Length caps
   - Schema validation

---

## 8. File Reference Summary

### 8.1 Existing Layer 4 Files

| File | Purpose | Lines |
|------|---------|-------|
| pkg/security/types.go | Core types, TaintEngine/BoundaryService interfaces | 245 |
| pkg/security/security_test.go | Tests for types and helpers | 290 |
| pkg/services/security/service.go | SecurityService implementation | 142 |
| pkg/services/security/service_test.go | Tests for SecurityService | 507 |
| pkg/services/security/bootstrap.go | ChartDefinition for sys:security | 10 |
| pkg/datasource/datasource.go | DataSource interface and Registry | 70 |
| pkg/datasource/local_disk.go | localDisk implementation (partial) | 50 |
| pkg/datasource/datasource_test.go | Tests for DataSource | 94 |
| pkg/services/datasources/service.go | DatasourceService, stub implementations | 86 |
| pkg/services/datasources/service_test.go | Tests for DatasourceService | 82 |
| pkg/mail/types.go | Mail, BoundaryType, StreamChunk types | 90 |
| pkg/mail/stream.go | StreamSession, taint stripping | 122 |
| pkg/services/humangateway/chat.go | ContextMapSnapshot, SanitizeContextForBoundary | 51 |
| pkg/kernel/kernel.go | ApplicationContext implementation | 547 |
| pkg/bootstrap/actions.go | LoadSecurityService action | 345 |
| pkg/statechart/interfaces.go | ApplicationContext interface | 64 |

### 8.2 Missing Files (per spec-extraction.md L420-443)

| Expected File | Purpose | Status |
|---------------|---------|--------|
| pkg/security/boundary.go | Boundary enforcement (~150 lines) | ❌ Missing |
| pkg/security/taint.go | Taint engine (~200 lines) | ❌ Missing (logic in types.go is stubbed) |
| pkg/security/contextmap.go | ContextMap filtering (~100 lines) | ❌ Missing |
| pkg/security/rules.go | Taint rules (~50 lines) | ❌ Missing |
| pkg/datasource/objectStorage.go | S3/GCS implementations (~150 lines) | ❌ Missing |
| pkg/datasource/inMemory.go | inMemoryWorkspace (~50 lines) | ❌ Missing |
| pkg/datasource/registry.go | DataSourceRegistry (~100 lines) | ⚠️ Partial (in datasource.go) |

---

## 9. Conclusion

Layer 4 has **foundational scaffolding** in place with correct type definitions and interface contracts, but **no substantive security enforcement** is implemented. The codebase is ready for implementation work but requires significant development to meet spec requirements.

**Estimated Implementation Effort**:
- Critical path (boundary enforcement, xattr storage, ContextMap filtering): 2-3 weeks
- Full implementation (all DataSource backends, stream sanitization, taint engine): 4-6 weeks

**Risk**: Current implementation provides **zero security guarantees** - all enforcement methods are stubs that return success without validation.

---

*Assessment completed: 2026-03-09*
