# Layer 4: Security & Boundaries + DataSources - Phase Breakdown

**Generated**: 2026-03-09  
**Source Spec**: `docs/arch-v1.md` v1.3 (Section 4: L261-302, Section 11: L739-798, Section 14.2: L1303-1322)  
**Status**: Planning document for TDD implementation

---

## Executive Summary

Layer 4 implements **Security & Boundaries + DataSources** - the core security enforcement layer that ensures data crossing boundaries (inner/DMZ/outer) is properly sanitized and that file/object storage operations carry appropriate taints.

**Key Guarantee** (arch-v1.md L20): "Never allow outer/DMZ ContextMap to contain untagged inner data"

### Phase Overview

| Phase | Name | Tests | Files | Dependencies | Spec Reference |
|-------|------|-------|-------|--------------|----------------|
| P1.1 | AttachTaint functionality | 4 | 2 | None | L277-281, L1312 |
| P1.2 | PropagateTaint recursive propagation | 5 | 2 | P1.1 | L283, L1305 |
| P1.3 | StripTaint removal | 4 | 2 | P1.1 | L284, L298 |
| P1.4 | TaintSet operations | 3 | 1 | None | L273, spec-extraction L36-58 |
| P1.6 | Tool TaintOutput | 3 | 2 | P1.1, P1.2 | L562, L1070, L1075 |
| P2.1 | LocalDisk xattr implementation | 5 | 2 | P1 | L757-761, L787-793 |
| P2.2 | S3 DataSource | 4 | 2 | P1, P2.1 | L796 |
| P2.3 | InMemory DataSource | 3 | 2 | P1 | L796 |
| P2.4 | DataSource Registry | 3 | 1 | P2.1-P2.3 | L787-793 |
| P2.5 | Persistence Taint Policy | 3 | 2 | P1.1, P2.1, P3.1 | L285-286 |
| P2.6 | AlwaysTaintAs Mode | 2 | 2 | P2.1, P1 | L751, L1092, L764-767 |
| P3.1 | Boundary validation | 5 | 2 | P1, P2 | L261-270, L290 |
| P3.2 | Boundary transitions | 6 | 2 | P3.1 | L284, L290-291 |
| P3.3 | Violation reporting | 4 | 2 | P3.1, P3.2 | L286, L170 |
| P3.4 | Boundary Immutability | 3 | 3 | P3.1, P3.2, P3.3 | L126-127, L893, L1781-1783 |
| P3.5 | Sub-Agent Boundary Inheritance | 4 | 2 | P3.1, P3.2, P3.3, P3.4 | L1780, L265-270 |
| P3.6 | Boundary-aware Tool Resolution | 4 | 2 | P3.1, P3.2, P3.3 | L489, L552, L578 |
| P4.1 | TaintPolicy enforcement modes | 5 | 2 | P1, P3 | L209, L916 |
| P4.2 | ContextBlock filtering | 6 | 2 | P4.1 | L292-296, L216-223 |
| P4.3 | Isolation hooks | 3 | 2 | P4.1, P4.2 | L770-771 |
| P4.4 | ContextMap Token Budgeting | 4 | 2 | P4.1, P4.2 | L432, L1544, L200-223 |
| P4.5 | ContextBlock Sources | 5 | 2 | P4.2, P4.4 | L995-1030, L2028-2032 |
| P4.6 | applyBlockTaints | 2 | 2 | P1.1, P4.1 | L279, L283-288, L292-294 |
| P4.7 | Orchestrator Isolated Contexts | 2 | 2 | P4.1, P4.6 | L679, L1558, L1565, L1765 |
| P4.8 | Isolation Hooks API | 2 | 2 | P4.1, P4.2, P1 | L400, L516-570, L660 |
| P4.9 | ContextBlock TaintPolicy | 2 | 2 | P4.1, P4.2 | L292-296, L216-223 |
| P4.10 | ContextBlock Redaction | 2 | 2 | P4.2, P4.9 | L294, L218-221 |
| P4.11 | ContextBlock Audit | 2 | 2 | P4.2, P4.9 | L295, L210-213 |
| P4.12 | ContextBlock Strict | 2 | 2 | P4.2, P4.9 | L294, L210-213 |
| P5.1 | NamespaceIsolation | 3 | 2 | P1.1, P3.1, P4.1 | L781-783, L1303-1309 |
| P5.2 | TaintReporting | 3 | 2 | P1.1, P1.4 | L286, L1303-1309 |
| P5.3 | ContextPreparation | 3 | 2 | P1.3, P4.1, P4.2 | L292-296, L1308 |
| P5.4 | Stream Length Caps | 2 | 2 | P5.1, P5.2, P5.3 | L268, L298 |
| P5.5 | Stream Schema Validation | 2 | 2 | P5.1, P5.3, P5.4 | L268, L298 |
| P5.6 | sys:human-gateway Sanitization | 2 | 2 | P5.1, P5.3, P5.4, P5.5 | L731, L2348, L268 |
| P6.1 | SecurityService integration | 5 | 2 | All | L1303-1310 |
| P6.2 | Mail router hooks | 4 | 2 | P6.1 | L290, L1317 |
| P6.3 | ApplicationContext wiring | 4 | 2 | P6.1 | L383-387 |
| P6.4 | Runtime Taint Query | 2 | 2 | P1.1, P5.2, P6.1 | L286, L1308 |
| P6.5 | allowedOnExit Snapshots | 2 | 3 | P1.1, P5.2, P6.1, P6.4 | L129, L280, L916, L1317, L1363, L1518, L2402 |
| P6.6 | allowedOnExit Sub-Agent Returns | 2 | 3 | P1, P6.1, P6.5 | L280, L1317, L1518 |
| P7.1 | Data Tainting E2E | 2 | 2 | P1-P6 | L277-287, L276, L1312-1315 |
| P7.2 | Context Assembly E2E | 2 | 2 | P1-P6 | L288-302, L261-270, L1320-1322 |
| P7.3 | Boundary Enforcement E2E | 2 | 2 | P1-P6 | L298-301, L644-648, L1518 |
| P7.4 | Stream Sanitization E2E | 2 | 2 | P1-P6 | L298-301, L644-648, L1518 |
| P7.5 | Persistence Enforcement E2E | 2 | 2 | P1-P6 | L285-286, L1317 |
| P7.6 | Violation Reporting E2E | 2 | 2 | P1-P6 | L286, L170 |

**Total: 158 tests, ~80 files modified/created, 46 sub-phases**

### Current State (from implementation-assessment.md)

**Completed**:
- ✅ Core type definitions (BoundaryType, TaintPolicy, ContextBlock)
- ✅ DataSource interface and registry pattern
- ✅ Security service skeleton with `sys:security` ID
- ✅ TaintEngine interface with stub implementations
- ✅ TaintSet and TaintMap types with basic operations

**Missing**:
- ❌ No actual boundary enforcement logic
- ❌ No taint tracking/propagation implementation
- ❌ No xattr-based taint storage for local disk
- ❌ No ContextMap filtering/redaction implementation
- ❌ No `allowedOnExit` policy enforcement

---

## Phase P1: Taint Engine Core (Foundation)

**Goal**: Implement the foundational taint tracking mechanisms that all other security features depend on.

### P1.1: AttachTaint functionality

**Parent**: Phase P1 (Taint Engine Core)  
**Spec References**: arch-v1.md L277-281 (Taint Attachment Points), L1312 (SHALL attach taints)  
**Status**: ❌ PENDING

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: None

#### Requirements

From `arch-v1.md L277-281` - Taint Attachment Points:
1. Attach taints at Message creation (Gateway → user, Orchestrator → tool_result)
2. Attach taints at ContextBlock assembly
3. Attach taints on data leaving runtime (uses allowedOnExit)
4. Attach taints at file read operations

#### Implementation Details

**Files to create/modify**:
- `pkg/security/taint.go` - Create AttachTaint function
- `pkg/security/taint_test.go` - Add tests

**Functions to implement**:
```go
func AttachTaint(obj any, taints []string) (any, error)
```

**Test scenarios**:
1. Attach taints to map[string]interface{}
2. Attach taints to Mail struct
3. Attach taints to nested structures
4. Attach taints to slice of maps

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestAttachTaint_Mail**
```go
func TestAttachTaint_Mail(t *testing.T)
```
**Acceptance Criteria**:
- Mail.Metadata.Taints is updated with new taints
- Existing taints are preserved (union operation)

**Test 2: TestAttachTaint_Map**
```go
func TestAttachTaint_Map(t *testing.T)
```
**Acceptance Criteria**:
- Map receives `_taints` key with taint array
- Map values are not modified

**Test 3: TestAttachTaint_Nested**
```go
func TestAttachTaint_Nested(t *testing.T)
```
**Acceptance Criteria**:
- Nested maps each receive taints
- Taints attached at root and nested levels

**Test 4: TestAttachTaint_Empty**
```go
func TestAttachTaint_Empty(t *testing.T)
```
**Acceptance Criteria**:
- Empty taint list does not modify object
- Nil object returns error

---

### P1.2: PropagateTaint recursive propagation

**Parent**: Phase P1 (Taint Engine Core)  
**Spec References**: arch-v1.md L283 (Taint Propagation), L1305 (TaintPropagate interface)  
**Status**: ❌ PENDING

#### Overview
- Tests: 5
- Commits: 5 (1:1 ratio)
- Dependencies: P1.1

#### Requirements

From `arch-v1.md L283` - Taint Propagation:
1. Security Service propagates taints on copy/read/write
2. Merge taints from source to destination
3. Handle nested map structures
4. Handle slice structures
5. Deduplicate taints during propagation

#### Implementation Details

**Files to create/modify**:
- `pkg/security/taint.go` - Create PropagateTaint function
- `pkg/security/taint_test.go` - Add tests

**Functions to implement**:
```go
func PropagateTaint(obj any, newTaints []string) (any, error)
```

#### Tests to Write (5 tests, 5 commits)

**Test 1: TestPropagateTaint_MapMerge**
```go
func TestPropagateTaint_MapMerge(t *testing.T)
```
**Acceptance Criteria**:
- Existing taints merged with new taints
- No duplicate taints in result

**Test 2: TestPropagateTaint_NestedMap**
```go
func TestPropagateTaint_NestedMap(t *testing.T)
```
**Acceptance Criteria**:
- Taints propagated to all nested map levels
- Each level has complete taint set

**Test 3: TestPropagateTaint_Slice**
```go
func TestPropagateTaint_Slice(t *testing.T)
```
**Acceptance Criteria**:
- All map elements in slice receive taints
- Non-map elements in slice unchanged

**Test 4: TestPropagateTaint_Deduplication**
```go
func TestPropagateTaint_Deduplication(t *testing.T)
```
**Acceptance Criteria**:
- Duplicate taints removed
- Taint order preserved for unique values

**Test 5: TestPropagateTaint_EmptySource**
```go
func TestPropagateTaint_EmptySource(t *testing.T)
```
**Acceptance Criteria**:
- Object without existing taints receives new taints
- Empty newTaints list preserves existing taints

---

### P1.3: StripTaint removal

**Parent**: Phase P1 (Taint Engine Core)  
**Spec References**: arch-v1.md L284 (Inner → DMZ/outer auto-strip), L298 (Stream sanitization)  
**Status**: ❌ PENDING

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P1.1

#### Requirements

From `arch-v1.md L284` - Taint Stripping:
1. Strip forbidden taints on inner → DMZ/outer transitions
2. Strip based on allowedOnExit policy
3. Handle nested structure stripping
4. Return stripped taint list for audit

#### Implementation Details

**Files to create/modify**:
- `pkg/security/taint.go` - Create StripTaint function
- `pkg/security/taint_test.go` - Add tests

**Functions to implement**:
```go
func StripTaint(obj any, forbiddenTaints []string) (any, []string, error)
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestStripTaint_Basic**
```go
func TestStripTaint_Basic(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden taints removed from object
- Allowed taints preserved

**Test 2: TestStripTaint_ReturnsStripped**
```go
func TestStripTaint_ReturnsStripped(t *testing.T)
```
**Acceptance Criteria**:
- Function returns list of stripped taints
- Stripped list matches forbidden taints that existed

**Test 3: TestStripTaint_Nested**
```go
func TestStripTaint_Nested(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden taints stripped from all nested levels
- Stripped taints aggregated from all levels

**Test 4: TestStripTaint_NoForbidden**
```go
func TestStripTaint_NoForbidden(t *testing.T)
```
**Acceptance Criteria**:
- Object unchanged when no forbidden taints present
- Empty stripped list returned

---

### P1.4: TaintSet operations

**Parent**: Phase P1 (Taint Engine Core)  
**Spec References**: arch-v1.md L273 (Taint sets), spec-extraction.md L36-58 (TaintSet type)  
**Status**: ⚠️ PARTIAL (types exist, operations need testing)

#### Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: None

#### Requirements

From `arch-v1.md L273` - Taint Sets:
1. Add taint to set
2. Check if taint exists in set
3. Union two taint sets

#### Implementation Details

**Files to create/modify**:
- `pkg/security/types.go` - Verify TaintSet operations
- `pkg/security/security_test.go` - Add tests

**Functions to verify**:
```go
func (ts TaintSet) Add(taint string)
func (ts TaintSet) Has(taint string) bool
func (ts TaintSet) Union(other TaintSet) TaintSet
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestTaintSet_AddHas**
```go
func TestTaintSet_AddHas(t *testing.T)
```
**Acceptance Criteria**:
- Add adds taint to set
- Has returns true for added taint
- Has returns false for non-existent taint

**Test 2: TestTaintSet_Union**
```go
func TestTaintSet_Union(t *testing.T)
```
**Acceptance Criteria**:
- Union combines both sets
- No duplicates in union result

**Test 3: TestTaintSet_EmptyOperations**
```go
func TestTaintSet_EmptyOperations(t *testing.T)
```
**Acceptance Criteria**:
- Operations on empty set work correctly
- Union with empty set returns original

---

## Phase P2: DataSource (Depends on P1)

**Goal**: Implement pluggable DataSource abstraction for filesystem/object storage with automatic tainting.

### P2.1: LocalDisk xattr implementation

**Parent**: Phase P2 (DataSource)  
**Spec References**: arch-v1.md L757-761 (On Write workflow), L787-793 (DataSource interface)  
**Status**: ⚠️ PARTIAL (interface exists, xattr not implemented)

#### Overview
- Tests: 5
- Commits: 5 (1:1 ratio)
- Dependencies: P1

#### Requirements

From `arch-v1.md L757-761` - LocalDisk On Write:
1. Set xattr `user.maelstrom.taints` on file write
2. JSON-encode taint list for xattr value
3. Create sidecar fallback for non-xattr filesystems
4. Read xattr on file read
5. Return taints from GetTaints

#### Implementation Details

**Files to create/modify**:
- `pkg/datasource/local_disk.go` - Implement xattr operations
- `pkg/datasource/local_disk_test.go` - Add tests

**Functions to implement**:
```go
func (ld *localDisk) TagOnWrite(path string, taints []string) error
func (ld *localDisk) GetTaints(path string) ([]string, error)
```

#### Tests to Write (5 tests, 5 commits)

**Test 1: TestLocalDisk_TagOnWrite**
```go
func TestLocalDisk_TagOnWrite(t *testing.T)
```
**Acceptance Criteria**:
- xattr set on file with correct namespace
- Taints JSON-encoded in xattr value

**Test 2: TestLocalDisk_GetTaints**
```go
func TestLocalDisk_GetTaints(t *testing.T)
```
**Acceptance Criteria**:
- xattr read and decoded correctly
- Returns original taint list

**Test 3: TestLocalDisk_SidecarFallback**
```go
func TestLocalDisk_SidecarFallback(t *testing.T)
```
**Acceptance Criteria**:
- Sidecar file created when xattr fails
- Sidecar contains JSON taint data

**Test 4: TestLocalDisk_ValidateAccess_Allowed**
```go
func TestLocalDisk_ValidateAccess_Allowed(t *testing.T)
```
**Acceptance Criteria**:
- Boundary access allowed when in allowedForBoundary list
- No error returned

**Test 5: TestLocalDisk_ValidateAccess_Denied**
```go
func TestLocalDisk_ValidateAccess_Denied(t *testing.T)
```
**Acceptance Criteria**:
- Boundary access denied when not in allowedForBoundary
- Error returned with boundary mismatch

---

### P2.2: S3 DataSource

**Parent**: Phase P2 (DataSource)  
**Spec References**: arch-v1.md L796 (Built-in implementations), L758 (Object storage tags)  
**Status**: ❌ MISSING

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P1, P2.1

#### Requirements

From `arch-v1.md L796` - S3 DataSource:
1. Set object tags on write
2. Read object tags on read
3. Validate boundary access
4. Handle S3-specific errors

#### Implementation Details

**Files to create/modify**:
- `pkg/datasource/s3.go` - Create S3 implementation
- `pkg/datasource/s3_test.go` - Add tests

**Functions to implement**:
```go
func NewS3DataSource(config map[string]any) (DataSource, error)
func (s *s3DataSource) TagOnWrite(key string, taints []string) error
func (s *s3DataSource) GetTaints(key string) ([]string, error)
func (s *s3DataSource) ValidateAccess(boundary string) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestS3DataSource_TagOnWrite**
```go
func TestS3DataSource_TagOnWrite(t *testing.T)
```
**Acceptance Criteria**:
- Object tags set on S3 object
- Taints stored as tag values

**Test 2: TestS3DataSource_GetTaints**
```go
func TestS3DataSource_GetTaints(t *testing.T)
```
**Acceptance Criteria**:
- Object tags read and decoded
- Returns taint list from tags

**Test 3: TestS3DataSource_ValidateAccess**
```go
func TestS3DataSource_ValidateAccess(t *testing.T)
```
**Acceptance Criteria**:
- Boundary validation works
- Error on unauthorized boundary

**Test 4: TestS3DataSource_Config**
```go
func TestS3DataSource_Config(t *testing.T)
```
**Acceptance Criteria**:
- Bucket name from config
- Region from config

---

### P2.3: InMemory DataSource

**Parent**: Phase P2 (DataSource)  
**Spec References**: arch-v1.md L796 (inMemoryWorkspace)  
**Status**: ❌ MISSING

#### Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P1

#### Requirements

From `arch-v1.md L796` - InMemory Workspace:
1. Store taints in memory map
2. Fast lookup for agent workspace
3. No persistence (ephemeral)

#### Implementation Details

**Files to create/modify**:
- `pkg/datasource/in_memory.go` - Create in-memory implementation
- `pkg/datasource/in_memory_test.go` - Add tests

**Functions to implement**:
```go
func NewInMemoryDataSource() DataSource
func (im *inMemoryDataSource) TagOnWrite(path string, taints []string) error
func (im *inMemoryDataSource) GetTaints(path string) ([]string, error)
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestInMemory_TagOnWrite**
```go
func TestInMemory_TagOnWrite(t *testing.T)
```
**Acceptance Criteria**:
- Taints stored in memory
- Path used as key

**Test 2: TestInMemory_GetTaints**
```go
func TestInMemory_GetTaints(t *testing.T)
```
**Acceptance Criteria**:
- Taints retrieved from memory
- Empty list for unknown path

**Test 3: TestInMemory_Concurrent**
```go
func TestInMemory_Concurrent(t *testing.T)
```
**Acceptance Criteria**:
- Thread-safe operations
- No race conditions

---

### P2.4: DataSource Registry

**Parent**: Phase P2 (DataSource)  
**Spec References**: arch-v1.md L787-793 (DataSource interface), implementation-assessment.md L88-93  
**Status**: ✅ EXISTS (needs verification tests)

#### Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P2.1, P2.2, P2.3

#### Requirements

From `implementation-assessment.md L88-93`:
1. Register DataSource implementations by name
2. Get DataSource by name
3. List registered DataSources

#### Implementation Details

**Files to create/modify**:
- `pkg/datasource/datasource.go` - Verify registry
- `pkg/datasource/datasource_test.go` - Add tests

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestRegistry_Register**
```go
func TestRegistry_Register(t *testing.T)
```
**Acceptance Criteria**:
- DataSource registered by name
- No duplicate registration

**Test 2: TestRegistry_Get**
```go
func TestRegistry_Get(t *testing.T)
```
**Acceptance Criteria**:
- Returns registered DataSource
- Error for unknown name

**Test 3: TestRegistry_List**
```go
func TestRegistry_List(t *testing.T)
```
**Acceptance Criteria**:
- Returns all registered names
- Includes built-in implementations

---

## Phase P3: Boundary Enforcement (Depends on P1, P2)

**Goal**: Implement boundary validation and enforcement for inner/DMZ/outer transitions.

### P3.1: Boundary validation

**Parent**: Phase P3 (Boundary Enforcement)  
**Spec References**: arch-v1.md L261-270 (Boundary Model), L290 (Mail-only transitions)  
**Status**: ⚠️ PARTIAL (types exist, enforcement missing)

#### Overview
- Tests: 5
- Commits: 5 (1:1 ratio)
- Dependencies: P1, P2

#### Requirements

From `arch-v1.md L261-270` - Boundary Model:
1. Inner: Full read/write to sensitive data
2. DMZ: Mediated access with sanitization
3. Outer: Untrusted ingress only
4. Validate boundary types
5. Validate boundary capabilities

#### Implementation Details

**Files to create/modify**:
- `pkg/security/boundary.go` - Create boundary validation
- `pkg/security/boundary_test.go` - Add tests

**Functions to implement**:
```go
func ValidateBoundary(boundary BoundaryType) error
func GetBoundaryCapabilities(boundary BoundaryType) Capabilities
```

#### Tests to Write (5 tests, 5 commits)

**Test 1: TestValidateBoundary_Valid**
```go
func TestValidateBoundary_Valid(t *testing.T)
```
**Acceptance Criteria**:
- inner, dmz, outer all valid
- No error for valid boundaries

**Test 2: TestValidateBoundary_Invalid**
```go
func TestValidateBoundary_Invalid(t *testing.T)
```
**Acceptance Criteria**:
- Unknown boundary returns error
- Empty boundary returns error

**Test 3: TestInnerCapabilities**
```go
func TestInnerCapabilities(t *testing.T)
```
**Acceptance Criteria**:
- Inner has full read/write
- Inner can access secrets

**Test 4: TestDMZCapabilities**
```go
func TestDMZCapabilities(t *testing.T)
```
**Acceptance Criteria**:
- DMZ has mediated access
- DMZ requires sanitization

**Test 5: TestOuterCapabilities**
```go
func TestOuterCapabilities(t *testing.T)
```
**Acceptance Criteria**:
- Outer is ingress only
- Outer cannot access inner data

---

### P3.2: Boundary transitions

**Parent**: Phase P3 (Boundary Enforcement)  
**Spec References**: arch-v1.md L284 (auto-strip/block), L290-291 (Mail-only transitions)  
**Status**: ⚠️ PARTIAL (partial implementation exists)

#### Overview
- Tests: 6
- Commits: 6 (1:1 ratio)
- Dependencies: P3.1

#### Requirements

From `arch-v1.md L284, L290-291`:
1. Inner → DMZ: auto-strip forbidden taints
2. Inner → outer: auto-strip forbidden taints
3. Outer → inner: block on forbidden taints
4. DMZ → inner: block on forbidden taints
5. Same boundary: pass through
6. All transitions via Mail only

#### Implementation Details

**Files to create/modify**:
- `pkg/security/boundary.go` - Create transition enforcement
- `pkg/security/boundary_test.go` - Add tests

**Functions to implement**:
```go
func EnforceTransition(source, target BoundaryType, taints []string) ([]string, error)
```

#### Tests to Write (6 tests, 6 commits)

**Test 1: TestTransition_InnerToDMZ**
```go
func TestTransition_InnerToDMZ(t *testing.T)
```
**Acceptance Criteria**:
- INNER_ONLY taints stripped
- PII taints stripped
- Allowed taints preserved

**Test 2: TestTransition_InnerToOuter**
```go
func TestTransition_InnerToOuter(t *testing.T)
```
**Acceptance Criteria**:
- All sensitive taints stripped
- Only EXTERNAL allowed

**Test 3: TestTransition_OuterToInner**
```go
func TestTransition_OuterToInner(t *testing.T)
```
**Acceptance Criteria**:
- PII blocked with error
- USER_SUPPLIED allowed with EXTERNAL taint added

**Test 4: TestTransition_DMZToInner**
```go
func TestTransition_DMZToInner(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden taints blocked
- Valid DMZ data allowed

**Test 5: TestTransition_SameBoundary**
```go
func TestTransition_SameBoundary(t *testing.T)
```
**Acceptance Criteria**:
- No modification
- No error

**Test 6: TestTransition_ForbiddenTaints**
```go
func TestTransition_ForbiddenTaints(t *testing.T)
```
**Acceptance Criteria**:
- SECRET taint blocked on outer→inner
- Error includes forbidden taint name

---

### P3.3: Violation reporting

**Parent**: Phase P3 (Boundary Enforcement)  
**Spec References**: arch-v1.md L286 (taint_violation event), L170 (Mail types)  
**Status**: ❌ MISSING

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P3.1, P3.2

#### Requirements

From `arch-v1.md L286`:
1. Emit taint_violation event on policy violation
2. Send to dead-letter queue
3. Include violation details
4. Track violation count

#### Implementation Details

**Files to create/modify**:
- `pkg/security/violation.go` - Create violation reporting
- `pkg/security/violation_test.go` - Add tests

**Functions to implement**:
```go
func ReportViolation(runtimeId string, violation TaintViolation) error
type TaintViolation struct {
    RuntimeID    string
    Source       BoundaryType
    Target       BoundaryType
    ForbiddenTaints []string
    Timestamp    time.Time
}
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestReportViolation_CreateEvent**
```go
func TestReportViolation_CreateEvent(t *testing.T)
```
**Acceptance Criteria**:
- taint_violation Mail created
- Mail type is MailTypeTaintViolation

**Test 2: TestReportViolation_DeadLetter**
```go
func TestReportViolation_DeadLetter(t *testing.T)
```
**Acceptance Criteria**:
- Violation sent to dead-letter queue
- sys:observability receives event

**Test 3: TestReportViolation_Details**
```go
func TestReportViolation_Details(t *testing.T)
```
**Acceptance Criteria**:
- All violation details included
- Timestamp recorded

**Test 4: TestReportViolation_Count**
```go
func TestReportViolation_Count(t *testing.T)
```
**Acceptance Criteria**:
- Violation count incremented
- Count retrievable via ReportTaints

---

## Phase P4: ContextMap Filtering (Depends on P1, P3)

**Goal**: Implement ContextMap filtering and redaction based on taint policies.

### P4.1: TaintPolicy enforcement modes

**Parent**: Phase P4 (ContextMap Filtering)  
**Spec References**: arch-v1.md L209, L916 (enforcement modes)  
**Status**: ⚠️ PARTIAL (struct exists, enforcement missing)

#### Overview
- Tests: 5
- Commits: 5 (1:1 ratio)
- Dependencies: P1, P3

#### Requirements

From `arch-v1.md L209, L916` - Enforcement Modes:
1. strict: Block forbidden taints entirely
2. redact: Replace forbidden values with replacement text
3. audit: Allow but log violations
4. allowedOnExit: Define allowed taints for outbound
5. Redact rules application

#### Implementation Details

**Files to create/modify**:
- `pkg/security/policy.go` - Create policy enforcement
- `pkg/security/policy_test.go` - Add tests

**Functions to implement**:
```go
func EnforcePolicy(data any, policy TaintPolicy, boundary BoundaryType) (any, error)
```

#### Tests to Write (5 tests, 5 commits)

**Test 1: TestPolicy_Strict**
```go
func TestPolicy_Strict(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden taints cause error
- No data passes through

**Test 2: TestPolicy_Redact**
```go
func TestPolicy_Redact(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden values replaced
- Replacement text from redactRules

**Test 3: TestPolicy_Audit**
```go
func TestPolicy_Audit(t *testing.T)
```
**Acceptance Criteria**:
- Data passes through
- Violation logged

**Test 4: TestPolicy_AllowedOnExit**
```go
func TestPolicy_AllowedOnExit(t *testing.T)
```
**Acceptance Criteria**:
- Only allowed taints pass
- Others stripped or blocked

**Test 5: TestPolicy_RedactRules**
```go
func TestPolicy_RedactRules(t *testing.T)
```
**Acceptance Criteria**:
- Multiple redact rules applied
- Each taint has correct replacement

---

### P4.2: ContextBlock filtering

**Parent**: Phase P4 (ContextMap Filtering)  
**Spec References**: arch-v1.md L292-296 (ContextMap enforcement), L216-223 (per-block policy)  
**Status**: ❌ MISSING

#### Overview
- Tests: 6
- Commits: 6 (1:1 ratio)
- Dependencies: P4.1

#### Requirements

From `arch-v1.md L292-296`:
1. Call prepareContextForBoundary in assembleContextMap
2. Redact blocks with forbidden taints
3. Drop blocks with forbidden taints (dropBlock mode)
4. Per-block taintPolicy overrides global
5. allowedForBoundary enforcement
6. Audit mode logging

#### Implementation Details

**Files to create/modify**:
- `pkg/security/contextmap.go` - Create ContextMap filtering
- `pkg/security/contextmap_test.go` - Add tests

**Functions to implement**:
```go
func PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error
func FilterContextBlock(block ContextBlock, boundary BoundaryType) (ContextBlock, error)
```

#### Tests to Write (6 tests, 6 commits)

**Test 1: TestPrepareContextForBoundary_Filter**
```go
func TestPrepareContextForBoundary_Filter(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden blocks filtered out
- Allowed blocks preserved

**Test 2: TestFilterContextBlock_Redact**
```go
func TestFilterContextBlock_Redact(t *testing.T)
```
**Acceptance Criteria**:
- Forbidden taint values redacted
- Block structure preserved

**Test 3: TestFilterContextBlock_DropBlock**
```go
func TestFilterContextBlock_DropBlock(t *testing.T)
```
**Acceptance Criteria**:
- Entire block dropped
- Block not in output

**Test 4: TestFilterContextBlock_PerBlockOverride**
```go
func TestFilterContextBlock_PerBlockOverride(t *testing.T)
```
**Acceptance Criteria**:
- Per-block policy overrides global
- Block-specific rules applied

**Test 5: TestFilterContextBlock_AllowedForBoundary**
```go
func TestFilterContextBlock_AllowedForBoundary(t *testing.T)
```
**Acceptance Criteria**:
- Block allowed only for specified boundaries
- Other boundaries filter block

**Test 6: TestFilterContextBlock_Audit**
```go
func TestFilterContextBlock_Audit(t *testing.T)
```
**Acceptance Criteria**:
- Block passes through
- Violation audited

---

### P4.3: Isolation hooks

**Parent**: Phase P4 (ContextMap Filtering)  
**Spec References**: arch-v1.md L770-771 (Isolation Hook)  
**Status**: ❌ MISSING

#### Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P4.1, P4.2

#### Requirements

From `arch-v1.md L770-771`:
1. Mount workspace via isolated executor
2. No direct open() syscall allowed
3. All FS ops through mediated tool layer

#### Implementation Details

**Files to create/modify**:
- `pkg/security/isolation.go` - Create isolation hooks
- `pkg/security/isolation_test.go` - Add tests

**Functions to implement**:
```go
func NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestNamespaceIsolate_CreateView**
```go
func TestNamespaceIsolate_CreateView(t *testing.T)
```
**Acceptance Criteria**:
- IsolatedView created for runtime
- View scoped to agent namespace

**Test 2: TestNamespaceIsolate_FilterData**
```go
func TestNamespaceIsolate_FilterData(t *testing.T)
```
**Acceptance Criteria**:
- Data filtered by boundary
- Forbidden data excluded

**Test 3: TestNamespaceIsolate_Operation**
```go
func TestNamespaceIsolate_Operation(t *testing.T)
```
**Acceptance Criteria**:
- Operation type tracked
- Different operations have different views

---

## Phase P5: Stream Integration (Depends on P1, P3)

**Goal**: Implement stream sanitization and boundary enforcement for streaming data.

### P5.1: Stream taint stripping

**Parent**: Phase P5 (Stream Integration)  
**Spec References**: arch-v1.md L298-301 (Stream Sanitization)  
**Status**: ⚠️ PARTIAL (StripForbiddenTaints exists, not integrated)

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P1, P3

#### Requirements

From `arch-v1.md L298-301`:
1. Strip PII from stream chunks
2. Apply length caps
3. Schema validation
4. Inner-data stripping

#### Implementation Details

**Files to create/modify**:
- `pkg/mail/stream.go` - Enhance stream sanitization
- `pkg/mail/stream_test.go` - Add tests

**Functions to implement**:
```go
func SanitizeStreamChunk(chunk StreamChunk, policy TaintPolicy) (StreamChunk, error)
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestSanitizeStreamChunk_PII**
```go
func TestSanitizeStreamChunk_PII(t *testing.T)
```
**Acceptance Criteria**:
- PII taints stripped from chunk
- Chunk content redacted

**Test 2: TestSanitizeStreamChunk_LengthCap**
```go
func TestSanitizeStreamChunk_LengthCap(t *testing.T)
```
**Acceptance Criteria**:
- Chunk truncated to max length
- isFinal set on truncated chunks

**Test 3: TestSanitizeStreamChunk_Schema**
```go
func TestSanitizeStreamChunk_Schema(t *testing.T)
```
**Acceptance Criteria**:
- Schema validation applied
- Invalid chunks rejected

**Test 4: TestSanitizeStreamChunk_InnerData**
```go
func TestSanitizeStreamChunk_InnerData(t *testing.T)
```
**Acceptance Criteria**:
- INNER_ONLY data stripped
- Only allowed taints in output

---

### P5.2: Stream boundary enforcement

**Parent**: Phase P5 (Stream Integration)  
**Spec References**: arch-v1.md L298, L1317 (allowedOnExit enforcement)  
**Status**: ❌ MISSING

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P5.1

#### Requirements

From `arch-v1.md L298, L1317`:
1. Per-chunk stateless sanitization
2. Latency <50ms
3. No buffering unless requested
4. allowedOnExit enforcement on streams

#### Implementation Details

**Files to create/modify**:
- `pkg/mail/stream.go` - Add boundary enforcement
- `pkg/mail/stream_test.go` - Add tests

**Functions to implement**:
```go
func EnforceStreamBoundary(chunk StreamChunk, boundary BoundaryType) (StreamChunk, error)
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestEnforceStreamBoundary_PerChunk**
```go
func TestEnforceStreamBoundary_PerChunk(t *testing.T)
```
**Acceptance Criteria**:
- Each chunk processed independently
- No state between chunks

**Test 2: TestEnforceStreamBoundary_Latency**
```go
func TestEnforceStreamBoundary_Latency(t *testing.T)
```
**Acceptance Criteria**:
- Processing under 50ms
- Measured with benchmark

**Test 3: TestEnforceStreamBoundary_NoBuffer**
```go
func TestEnforceStreamBoundary_NoBuffer(t *testing.T)
```
**Acceptance Criteria**:
- Chunks emitted immediately
- No accumulation

**Test 4: TestEnforceStreamBoundary_AllowedOnExit**
```go
func TestEnforceStreamBoundary_AllowedOnExit(t *testing.T)
```
**Acceptance Criteria**:
- Only allowed taints in stream
- Forbidden taints stripped

---

## Phase P6: Service Wiring (Depends on All)

**Goal**: Wire all components into SecurityService and integrate with Mail router and ApplicationContext.

### P6.1: SecurityService integration

**Parent**: Phase P6 (Service Wiring)  
**Spec References**: arch-v1.md L1303-1310 (Security interface)  
**Status**: ⚠️ PARTIAL (skeleton exists, methods stubbed)

#### Overview
- Tests: 5
- Commits: 5 (1:1 ratio)
- Dependencies: All previous phases

#### Requirements

From `arch-v1.md L1303-1310`:
1. ValidateAndSanitize implementation
2. TaintPropagate implementation
3. NamespaceIsolate implementation
4. CheckTaintPolicy implementation
5. ReportTaints implementation

#### Implementation Details

**Files to create/modify**:
- `pkg/services/security/service.go` - Implement all methods
- `pkg/services/security/service_test.go` - Add integration tests

#### Tests to Write (5 tests, 5 commits)

**Test 1: TestSecurityService_ValidateAndSanitize_Full**
```go
func TestSecurityService_ValidateAndSanitize_Full(t *testing.T)
```
**Acceptance Criteria**:
- Full boundary validation
- Taint stripping applied
- Mail modified correctly

**Test 2: TestSecurityService_TaintPropagate_Full**
```go
func TestSecurityService_TaintPropagate_Full(t *testing.T)
```
**Acceptance Criteria**:
- Recursive propagation
- Nested structures handled
- Taints merged correctly

**Test 3: TestSecurityService_NamespaceIsolate_Full**
```go
func TestSecurityService_NamespaceIsolate_Full(t *testing.T)
```
**Acceptance Criteria**:
- Isolated view created
- Data filtered by namespace
- Operation tracked

**Test 4: TestSecurityService_CheckTaintPolicy_Full**
```go
func TestSecurityService_CheckTaintPolicy_Full(t *testing.T)
```
**Acceptance Criteria**:
- Policy enforced
- Returns correct allow/deny
- Boundary considered

**Test 5: TestSecurityService_ReportTaints_Full**
```go
func TestSecurityService_ReportTaints_Full(t *testing.T)
```
**Acceptance Criteria**:
- TaintMap returned
- All taints tracked
- RuntimeID keyed

---

### P6.2: Mail router hooks

**Parent**: Phase P6 (Service Wiring)  
**Spec References**: arch-v1.md L290 (Mail-only transitions), L1317 (allowedOnExit)  
**Status**: ❌ MISSING

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P6.1

#### Requirements

From `arch-v1.md L290, L1317`:
1. Mail routing through security
2. Boundary validation on route
3. allowedOnExit enforcement
4. taint_violation routing

#### Implementation Details

**Files to create/modify**:
- `pkg/mail/router.go` - Add security hooks
- `pkg/mail/router_test.go` - Add tests

**Functions to implement**:
```go
func (r *MailRouter) RouteWithSecurity(mail Mail) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestMailRouter_RouteWithSecurity**
```go
func TestMailRouter_RouteWithSecurity(t *testing.T)
```
**Acceptance Criteria**:
- Mail validated before routing
- Security service called

**Test 2: TestMailRouter_BoundaryValidation**
```go
func TestMailRouter_BoundaryValidation(t *testing.T)
```
**Acceptance Criteria**:
- Boundary checked on route
- Invalid transitions blocked

**Test 3: TestMailRouter_AllowedOnExit**
```go
func TestMailRouter_AllowedOnExit(t *testing.T)
```
**Acceptance Criteria**:
- allowedOnExit enforced
- Forbidden taints stripped

**Test 4: TestMailRouter_ViolationRouting**
```go
func TestMailRouter_ViolationRouting(t *testing.T)
```
**Acceptance Criteria**:
- taint_violation routed to dead-letter
- sys:observability receives

---

### P6.3: ApplicationContext wiring

**Parent**: Phase P6 (Service Wiring)  
**Spec References**: arch-v1.md L383-387 (ApplicationContext interface)  
**Status**: ⚠️ PARTIAL (interface exists, taints not stored)

#### Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P6.1

#### Requirements

From `arch-v1.md L383-387`:
1. Get returns taints
2. Set stores taints
3. Boundary-filtered Get
4. Namespace isolation

#### Implementation Details

**Files to create/modify**:
- `pkg/kernel/kernel.go` - Implement taint-aware Get/Set
- `pkg/kernel/kernel_test.go` - Add tests

**Functions to implement**:
```go
func (k *kernelApplicationContext) Get(key string, callerBoundary BoundaryType) (any, []string, error)
func (k *kernelApplicationContext) Set(key string, value any, taints []string, callerBoundary BoundaryType) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestApplicationContext_GetTaints**
```go
func TestApplicationContext_GetTaints(t *testing.T)
```
**Acceptance Criteria**:
- Taints returned with value
- Taints match stored values

**Test 2: TestApplicationContext_SetTaints**
```go
func TestApplicationContext_SetTaints(t *testing.T)
```
**Acceptance Criteria**:
- Taints stored with value
- Retrieved correctly

**Test 3: TestApplicationContext_BoundaryFilter**
```go
func TestApplicationContext_BoundaryFilter(t *testing.T)
```
**Acceptance Criteria**:
- Get filtered by caller boundary
- Forbidden taints hidden

**Test 4: TestApplicationContext_Namespace**
```go
func TestApplicationContext_Namespace(t *testing.T)
```
**Acceptance Criteria**:
- Namespace isolation works
- Different namespaces separated

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Dependencies | Status |
|-------|-------|-------|--------------|--------|
| P1.1 | 4 | 2 | None | ❌ |
| P1.2 | 5 | 2 | P1.1 | ❌ |
| P1.3 | 4 | 2 | P1.1 | ❌ |
| P1.4 | 3 | 1 | None | ⚠️ |
| P1.6 | 3 | 2 | P1.1, P1.2 | ❌ |
| P2.1 | 5 | 2 | P1 | ⚠️ |
| P2.2 | 4 | 2 | P1, P2.1 | ❌ |
| P2.3 | 3 | 2 | P1 | ❌ |
| P2.4 | 3 | 1 | P2.1-P2.3 | ✅ |
| P2.5 | 3 | 2 | P1.1, P2.1, P3.1 | ❌ |
| P2.6 | 2 | 2 | P2.1, P1 | ❌ |
| P3.1 | 5 | 2 | P1, P2 | ⚠️ |
| P3.2 | 6 | 2 | P3.1 | ⚠️ |
| P3.3 | 4 | 2 | P3.1, P3.2 | ❌ |
| P3.4 | 3 | 3 | P3.1, P3.2, P3.3 | ❌ |
| P3.5 | 4 | 2 | P3.1-P3.4 | ❌ |
| P3.6 | 4 | 2 | P3.1, P3.2, P3.3 | ❌ |
| P4.1 | 5 | 2 | P1, P3 | ⚠️ |
| P4.2 | 6 | 2 | P4.1 | ❌ |
| P4.3 | 3 | 2 | P4.1, P4.2 | ❌ |
| P4.4 | 4 | 2 | P4.1, P4.2 | ❌ |
| P4.5 | 5 | 2 | P4.2, P4.4 | ❌ |
| P4.6 | 2 | 2 | P1.1, P4.1 | ❌ |
| P4.7 | 2 | 2 | P4.1, P4.6 | ❌ |
| P4.8 | 2 | 2 | P4.1, P4.2, P1 | ❌ |
| P4.9 | 2 | 2 | P4.1, P4.2 | ❌ |
| P4.10 | 2 | 2 | P4.2, P4.9 | ❌ |
| P4.11 | 2 | 2 | P4.2, P4.9 | ❌ |
| P4.12 | 2 | 2 | P4.2, P4.9 | ❌ |
| P5.1 | 3 | 2 | P1.1, P3.1, P4.1 | ❌ |
| P5.2 | 3 | 2 | P1.1, P1.4 | ❌ |
| P5.3 | 3 | 2 | P1.3, P4.1, P4.2 | ❌ |
| P5.4 | 2 | 2 | P5.1, P5.2, P5.3 | ❌ |
| P5.5 | 2 | 2 | P5.1, P5.3, P5.4 | ❌ |
| P5.6 | 2 | 2 | P5.1, P5.3-P5.5 | ❌ |
| P6.1 | 5 | 2 | All | ⚠️ |
| P6.2 | 4 | 2 | P6.1 | ❌ |
| P6.3 | 4 | 2 | P6.1 | ❌ |
| P6.4 | 2 | 2 | P1.1, P5.2, P6.1 | ❌ |
| P6.5 | 2 | 3 | P1.1, P5.2, P6.1, P6.4 | ❌ |
| P6.6 | 2 | 3 | P1, P6.1, P6.5 | ❌ |
| P7.1 | 2 | 2 | P1-P6 | ❌ |
| P7.2 | 2 | 2 | P1-P6 | ❌ |
| P7.3 | 2 | 2 | P1-P6 | ❌ |
| P7.4 | 2 | 2 | P1-P6 | ❌ |
| P7.5 | 2 | 2 | P1-P6 | ❌ |
| P7.6 | 2 | 2 | P1-P6 | ❌ |
| **Total** | **158** | **~80** | | |

### Execution Order

```
P1 (Taint Engine Core - Foundation)
├── P1.1: AttachTaint (4 tests)
├── P1.2: PropagateTaint (5 tests)
├── P1.3: StripTaint (4 tests)
├── P1.4: TaintSet (3 tests)
└── P1.6: Tool TaintOutput (3 tests)
    ↓
P2 (DataSource)
├── P2.1: LocalDisk (5 tests)
├── P2.2: S3 (4 tests)
├── P2.3: InMemory (3 tests)
├── P2.4: Registry (3 tests)
├── P2.5: Persistence Taint Policy (3 tests)
└── P2.6: AlwaysTaintAs Mode (2 tests)
    ↓
P3 (Boundary Enforcement)
├── P3.1: Validation (5 tests)
├── P3.2: Transitions (6 tests)
├── P3.3: Violations (4 tests)
├── P3.4: Boundary Immutability (3 tests)
├── P3.5: Sub-Agent Boundary Inheritance (4 tests)
└── P3.6: Boundary-aware Tool Resolution (4 tests)
    ↓
P4 (ContextMap Filtering & Isolation)
├── P4.1: Policy Modes (5 tests)
├── P4.2: Block Filtering (6 tests)
├── P4.3: Isolation Hooks (3 tests)
├── P4.4: ContextMap Token Budgeting (4 tests)
├── P4.5: ContextBlock Sources (5 tests)
├── P4.6: applyBlockTaints (2 tests)
├── P4.7: Orchestrator Isolated Contexts (2 tests)
├── P4.8: Isolation Hooks API (2 tests)
├── P4.9: ContextBlock TaintPolicy (2 tests)
├── P4.10: ContextBlock Redaction (2 tests)
├── P4.11: ContextBlock Audit (2 tests)
└── P4.12: ContextBlock Strict (2 tests)
    ↓
P5 (Stream Integration)
├── P5.1: NamespaceIsolation (3 tests)
├── P5.2: TaintReporting (3 tests)
├── P5.3: ContextPreparation (3 tests)
├── P5.4: Stream Length Caps (2 tests)
├── P5.5: Stream Schema Validation (2 tests)
└── P5.6: sys:human-gateway Sanitization (2 tests)
    ↓
P6 (Service Wiring)
├── P6.1: SecurityService (5 tests)
├── P6.2: Mail Hooks (4 tests)
├── P6.3: ApplicationContext (4 tests)
├── P6.4: Runtime Taint Query (2 tests)
├── P6.5: allowedOnExit Snapshots (2 tests)
└── P6.6: allowedOnExit Sub-Agent Returns (2 tests)
    ↓
P7 (E2E Workflows)
├── P7.1: Data Tainting E2E (2 tests)
├── P7.2: Context Assembly E2E (2 tests)
├── P7.3: Boundary Enforcement E2E (2 tests)
├── P7.4: Stream Sanitization E2E (2 tests)
├── P7.5: Persistence Enforcement E2E (2 tests)
└── P7.6: Violation Reporting E2E (2 tests)
```

### Next Steps

1. **Start Phase P1.1**: Create branch `feat/layer4-taint-attach`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after final phase to verify integration

### Key Files to Create

```
pkg/security/
├── taint.go          (P1.1-P1.3)
├── taint_test.go     (P1.1-P1.3)
├── boundary.go       (P3.1-P3.2)
├── boundary_test.go  (P3.1-P3.2)
├── violation.go      (P3.3)
├── violation_test.go (P3.3)
├── policy.go         (P4.1)
├── policy_test.go    (P4.1)
├── contextmap.go     (P4.2)
├── contextmap_test.go (P4.2)
├── isolation.go      (P4.3)
└── isolation_test.go (P4.3)

pkg/datasource/
├── local_disk.go     (P2.1 - enhance)
├── local_disk_test.go (P2.1 - enhance)
├── s3.go             (P2.2)
├── s3_test.go        (P2.2)
├── in_memory.go      (P2.3)
└── in_memory_test.go (P2.3)
```

---

*Document generated: 2026-03-09*  
*Based on: arch-v1.md v1.3, spec-extraction.md, implementation-assessment.md, implementation-patterns.md*
