# Implementation Patterns from Completed Layers

**Generated**: 2026-03-09  
**Purpose**: Document proven patterns from Layer 1, 2, 3, and Gap Remediation for Layer 4 planning

---

## 1. Documentation Template Structure

### 1.1 Phase Breakdown Document Template

**File naming**: `layer-XX-phase-breakdown.md` or `phase-XX.YY-description.md`

```markdown
# Phase [X.Y]: [Descriptive Name]

**Parent**: Phase [X] ([Category])  
**Gap References**: [L2-H1, L3-C2, etc.]  
**Status**: ❌ PENDING | ⏳ IN PROGRESS | ✅ COMPLETE

## Overview
- Tests: [N]
- Commits: [N] (1:1 ratio)
- Dependencies: [Phase X.Y, Phase X.Z]

## SPEC

### Requirements
From `arch-v1.md [Section]` - [Feature description]:
- [Requirement 1]
- [Requirement 2]
- [Requirement 3]

### Implementation Details

**Files to create/modify**:
- `[path/to/file.go]` - [Description of changes]
- `[path/to/file_test.go]` - Add tests

**Functions to implement**:
```go
func (s *ServiceName) MethodName(params) (returnType, error)
```

**Test scenarios**:
1. [Test scenario 1]
2. [Test scenario 2]
3. [Test scenario 3]

## TDD Workflow

### Iteration 1: [TestName]
1. Write test: `[TestName]` in `[file_test.go]`
2. Run → RED
3. Implement: [Minimal implementation]
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/[ID]): [one-line description]"`

### Iteration 2: [TestName]
...

## Deliverables
- [N] commits
- All tests passing
- Files modified: [list of files]
```

### 1.2 Example: Phase G1.1 ParseAddress (Actual)

From `docs/completed/gap-remediation/phase-g1.1-parse-address.md`:

```markdown
# Phase G1.1: ParseAddress Implementation

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L3-H4  
**Status**: ❌ PRIORITY

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: None

## SPEC

### Requirements
From `arch-v1.md 9.2` - Mail addressing system:
- Parse agent:id format addresses
- Parse topic:name format addresses  
- Parse sys:service format addresses
- Return error for invalid formats

### Implementation Details

**Files to create/modify**:
- `pkg/mail/router.go` - Add ParseAddress function
- `pkg/mail/router_test.go` - Add tests

**Functions to implement**:
```go
func ParseAddress(address string) (AddressType, string, error)
```

**AddressType enum**:
```go
type AddressType int

const (
    AgentAddress AddressType = iota
    TopicAddress
    SysAddress
)
```

**Test scenarios**:
1. Parse agent:id format
2. Parse topic:name format
3. Parse sys:service format
4. Parse invalid format returns error

## TDD Workflow

### Iteration 1: TestParseAddress_agent
1. Write test: `TestParseAddress_agent` in `pkg/mail/router_test.go`
2. Run → RED
3. Implement: Add AddressType constants and ParseAddress function with agent: support
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H4): add ParseAddress for agent: format"`
```

---

## 2. Granularity Guidelines

### 2.1 Phase Size Rules

| Metric | Recommended | Acceptable | Maximum |
|--------|-------------|------------|---------|
| Tests per phase | 2-5 | 6-8 | 10 |
| Files per phase | 1-3 | 4-5 | 6 |
| Commits per phase | = tests | = tests | = tests |
| Implementation time | 1-2 hours | 2-4 hours | 4-6 hours |

### 2.2 When to Split a Phase

**Split when**:
1. **Multiple independent components** (e.g., Phase 3.6 had 6 adapters → split into 3.6a, 3.6b, 3.6c)
2. **Tests exceed 8** (e.g., Phase 2.5 had 10 tests, acceptable but borderline)
3. **Different files with loose coupling** (e.g., Phase 3.3 split into AgentInbox, Topic, ServiceInbox)
4. **Different dependencies** (some tests depend on A, others on B)

**Keep together when**:
1. **Tightly coupled types** (e.g., Mail, MailType, MailMetadata in Phase 3.1)
2. **Single coherent feature** (e.g., ParseAddress in G1.1)
3. **Tests < 6** (no need to split)

### 2.3 Actual Phase Sizes from Completed Work

| Phase | Tests | Files | Category |
|-------|-------|-------|----------|
| G1.1: ParseAddress | 4 | 2 | Ideal |
| G1.2: StreamSession | 4 | 2 | Ideal |
| G1.3: Security Boundary | 2 | 1 | Ideal |
| G2.3: At-Least-Once | 4 | 2 | Ideal |
| G4.3: Hot-Reloadable Services | 10 | 5 | Large (5 services) |
| Layer 2.1: Type Definitions | 6 | 3 | Acceptable |
| Layer 2.5: Security Service | 10 | 1 | Large (4 methods) |

### 2.4 Phase Decomposition Strategy

**Example: Layer 3 Phase Breakdown**

```
Layer 3 (38 tests, 16 files, 8 phases)
├── 3.1: Mail Core Types (6 tests) - foundation
├── 3.2: Mail Router (5 tests) - depends on 3.1
├── 3.3: Inboxes & Topics (6 tests) - depends on 3.1
├── 3.4: Publisher/Subscriber (4 tests) - depends on 3.1, 3.3
├── 3.5: Streaming Support (5 tests) - depends on 3.1
├── 3.6: Gateway Adapters (6 tests) - depends on 3.1, 3.5
├── 3.7: Human Gateway Service (4 tests) - depends on 3.1, 3.6
└── 3.8: Integration (2 tests) - depends on all above
```

**Key insight**: Foundation types first (3.1), then build up, integration last (3.8).

---

## 3. Test-to-Requirement Mapping

### 3.1 Test Naming Convention

**Pattern**: `Test[Component]_[Behavior]_[ExpectedResult]`

**Examples from completed work**:

```go
// Type structure tests
func TestMail_AddressFormats(t *testing.T)
func TestMail_Types(t *testing.T)
func TestMail_Metadata(t *testing.T)
func TestMail_Structure(t *testing.T)

// Method behavior tests
func TestMailRouter_RouteToAgent(t *testing.T)
func TestMailRouter_RouteToTopic(t *testing.T)
func TestMailRouter_RouteToService(t *testing.T)
func TestMailRouter_RouteToUnknownAddress(t *testing.T)

// Edge case tests
func TestSecurityService_ValidateAndSanitizePassThrough(t *testing.T)
func TestSecurityService_ValidateAndSanitize_innerToOuter(t *testing.T)
func TestSecurityService_ValidateAndSanitize_outerToInner(t *testing.T)

// Integration tests
func TestFullMailFlow(t *testing.T)
func TestCommunicationService_Integration(t *testing.T)
```

### 3.2 Test Structure Template

```go
package [package_name]

import (
    "testing"
    
    "[internal_dependencies]"
)

func Test[Component]_[Behavior](t *testing.T) {
    // 1. Arrange: Create test fixtures
    svc := New[Component]()
    input := [InputType]{...}
    
    // 2. Act: Call the method
    result, err := svc.Method(input)
    
    // 3. Assert: Verify expectations
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if result.Field != expected {
        t.Errorf("Expected %v, got %v", expected, result.Field)
    }
}
```

### 3.3 Actual Test Example

From `pkg/services/security/service_test.go`:

```go
func TestSecurityService_ValidateAndSanitize_outerToInner(t *testing.T) {
    svc := NewSecurityService()

    inputMail := mail.Mail{
        ID:     "test-mail-outer-inner",
        Source: "agent:external",
        Target: "sys:security",
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{},
        },
    }

    result, err := svc.ValidateAndSanitize(inputMail, mail.OuterBoundary, mail.InnerBoundary)

    if err != nil {
        t.Errorf("Expected outer→inner transition to be allowed, got error: %v", err)
    }

    hasExternalTaint := false
    for _, taint := range result.Metadata.Taints {
        if taint == "EXTERNAL" {
            hasExternalTaint = true
            break
        }
    }

    if !hasExternalTaint {
        t.Errorf("Expected EXTERNAL taint to be added for outer→inner transition, got: %v", result.Metadata.Taints)
    }
}
```

### 3.4 Test-to-Requirement Matrix

| Requirement | Test Function | File |
|-------------|---------------|------|
| Parse agent:id format | TestParseAddress_agent | router_test.go |
| Parse topic:name format | TestParseAddress_topic | router_test.go |
| Parse sys:service format | TestParseAddress_sys | router_test.go |
| Invalid format returns error | TestParseAddress_invalid | router_test.go |

---

## 4. Commit Strategy

### 4.1 Commit Message Format

**Pattern**: `[type](scope): [one-line description]`

**Types**:
- `feat`: New feature (rare in gap remediation)
- `fix`: Bug fix or gap remediation (most common)
- `docs`: Documentation only
- `refactor`: Code restructuring without behavior change

**Scopes**:
- `gap/[ID]`: Gap remediation (e.g., `gap/L3-H4`)
- `layer/[N]`: Layer implementation (e.g., `layer/2`)
- `[component]`: Specific component (e.g., `mail`, `security`)

### 4.2 Actual Commit Messages from Gap Remediation

```
fix(gap/L3-H4): add ParseAddress for agent: format
fix(gap/L3-H4): add ParseAddress for topic: format
fix(gap/L3-H4): add ParseAddress for sys: format
fix(gap/L3-H4): add ParseAddress error handling

fix(gap/L3-C1): implement StreamSession.Send
fix(gap/L3-C1): implement StreamSession.Send for multiple chunks
fix(gap/L3-C1): implement StreamSession.Close
fix(gap/L3-C1): implement StreamSession.Close after sends

fix(gap/L2-H2): implement retry on delivery failure
fix(gap/L2-H2): implement exponential backoff
fix(gap/L2-H2): implement max retries limit
fix(gap/L2-H2): implement delivery attempt tracking

fix(gap/L2-M4): implement admin command execution
fix(gap/L2-M4): implement admin 2FA gate
fix(gap/L2-M4): implement persistence snapshot
fix(gap/L2-M4): implement persistence restore
```

### 4.3 Commit Granularity

**Rule**: One test = One commit

**Rationale**:
1. Each commit is a complete, testable unit
2. Easy to revert individual behaviors
3. Clear audit trail of what was implemented when
4. Matches TDD workflow exactly

**Example commit sequence for Phase G1.1**:
```
Commit 1: fix(gap/L3-H4): add ParseAddress for agent: format
  - Added: TestParseAddress_agent
  - Modified: router.go (AddressType enum, agent: parsing)

Commit 2: fix(gap/L3-H4): add ParseAddress for topic: format
  - Added: TestParseAddress_topic
  - Modified: router.go (topic: parsing)

Commit 3: fix(gap/L3-H4): add ParseAddress for sys: format
  - Added: TestParseAddress_sys
  - Modified: router.go (sys: parsing)

Commit 4: fix(gap/L3-H4): add ParseAddress error handling
  - Added: TestParseAddress_invalid
  - Modified: router.go (error handling)
```

---

## 5. Code Style

### 5.1 Import Patterns

**Order**:
1. Standard library
2. External dependencies (if any)
3. Internal packages (maelstrom)

**Example from `pkg/services/security/service.go`**:
```go
package security

import (
    "sync"  // standard library
    
    "github.com/maelstrom/v3/pkg/mail"      // internal
    "github.com/maelstrom/v3/pkg/security"  // internal
)
```

### 5.2 Type Definitions

**Pattern**: Types grouped by related functionality, enums as typed constants

**Example from `pkg/mail/types.go`**:
```go
package mail

import "time"

// Core message structure
type Mail struct {
    ID            string
    CorrelationID string
    Type          MailType
    CreatedAt     time.Time
    Source        string
    Target        string
    Content       any
    Metadata      MailMetadata
}

// Mail type enum
type MailType string

const (
    MailTypeUser             MailType = "user"
    MailTypeAssistant        MailType = "assistant"
    MailTypeToolResult       MailType = "tool_result"
    // ... more types
    
    // Aliases for backward compatibility
    User             = MailTypeUser
    Assistant        = MailTypeAssistant
    // ... more aliases
)

// Metadata structure
type MailMetadata struct {
    Tokens      int
    Model       string
    Cost        float64
    Boundary    BoundaryType
    Taints      []string
    Stream      bool
    StreamChunk *StreamChunk
    IsFinal     bool
}

// Boundary type enum
type BoundaryType string

const (
    InnerBoundary BoundaryType = "inner"
    DMZBoundary   BoundaryType = "dmz"
    OuterBoundary BoundaryType = "outer"
)
```

### 5.3 Service Interface Pattern

**Pattern**: Services implement common interface with ID(), HandleMail(), Start(), Stop()

**Example from `pkg/services/security/service.go`**:
```go
type SecurityService struct {
    mu sync.Mutex
}

func NewSecurityService() *SecurityService {
    return &SecurityService{}
}

// Service interface methods
func (s *SecurityService) ID() string {
    return "sys:security"
}

func (s *SecurityService) HandleMail(mail mail.Mail) error {
    return nil
}

func (s *SecurityService) Start() error {
    return nil
}

func (s *SecurityService) Stop() error {
    return nil
}

// Domain-specific methods
func (s *SecurityService) ValidateAndSanitize(m mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
    // Implementation
}

func (s *SecurityService) TaintPropagate(obj any, newTaints []string) (any, error) {
    // Implementation
}
```

### 5.4 Error Handling

**Pattern**: Return errors, don't panic; use `errors.New()` for simple errors

**Example from `pkg/mail/router.go`**:
```go
func (r *MailRouter) Route(mail Mail) error {
    addrType, id, err := ParseAddress(mail.Target)
    if err != nil {
        return err  // propagate parsing error
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    switch addrType {
    case AddressTypeAgent:
        inbox, exists := r.agents[id]
        if !exists {
            return errors.New("agent not found: " + id)
        }
        return inbox.Push(mail)
    case AddressTypeTopic:
        topic, exists := r.topics[id]
        if !exists {
            return errors.New("topic not found: " + id)
        }
        return topic.Publish(mail)
    default:
        return errors.New("unknown address type")
    }
}
```

### 5.5 Concurrency Patterns

**Pattern**: Use `sync.RWMutex` for read-heavy, `sync.Mutex` for write-heavy

**Example from `pkg/mail/router.go`**:
```go
type MailRouter struct {
    agents   map[string]*AgentInbox
    topics   map[string]*Topic
    services map[string]*ServiceInbox
    mu       sync.RWMutex  // RWMutex for read-heavy access
}

func (r *MailRouter) SubscribeAgent(id string, inbox *AgentInbox) error {
    r.mu.Lock()      // Write lock for mutation
    defer r.mu.Unlock()
    r.agents[id] = inbox
    return nil
}

func (t *Topic) Publish(mail Mail) error {
    t.mu.RLock()     // Read lock for iteration
    defer t.mu.RUnlock()

    for _, sub := range t.Subscribers {
        ch := sub.Receive()
        select {
        case ch <- mail:
        default:
        }
    }
    return nil
}
```

---

## 6. Phase Structure Template

### 6.1 Complete Phase Document Template

```markdown
# Phase [X.Y]: [Descriptive Name]

**Parent**: Phase [X] ([Category])  
**Gap References**: [L2-H1, L3-C2, etc.]  
**Status**: ❌ PENDING | ⏳ IN PROGRESS | ✅ COMPLETE

## Overview
- Tests: [N]
- Commits: [N] (1:1 ratio)
- Dependencies: [Phase X.Y, Phase X.Z]

## SPEC

### Requirements
From `arch-v1.md [Section]` - [Feature description]:
- [Requirement 1]
- [Requirement 2]
- [Requirement 3]

### Implementation Details

**Files to create/modify**:
- `[path/to/file.go]` - [Description]
- `[path/to/file_test.go]` - Add tests

**Functions to implement**:
```go
func (s *ServiceName) MethodName(params) (returnType, error)
```

**Test scenarios**:
1. [Scenario 1]
2. [Scenario 2]
3. [Scenario 3]

## TDD Workflow

### Iteration 1: [TestName]
1. Write test: `[TestName]` in `[file_test.go]`
2. Run → RED
3. Implement: [Minimal implementation]
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/[ID]): [description]"`

### Iteration 2: [TestName]
1. Write test: `[TestName]`
2. Run → RED
3. Implement: [Add functionality]
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/[ID]): [description]"`

## Deliverables
- [N] commits
- All tests passing
- Files modified: [list]
```

### 6.2 Phase Breakdown Document Template

```markdown
# Layer [N]: [Layer Name] - Phase Breakdown

## Executive Summary

Layer [N] implements [description]. Based on analysis of [previous layers], this document breaks down Layer [N] into **[X phases]** with **~[Y tests]** across **~[Y commits]**.

### Current State
- ✅ [Completed items]
- ⏳ [In progress items]
- ❌ [Missing items]

### Layer [N] Goal
Complete [objective]:
1. [Goal 1]
2. [Goal 2]
3. [Goal 3]

---

## Phase Overview

| Phase | Name | Tests | Files | Branch | Dependencies |
|-------|------|-------|-------|--------|--------------|
| [N].1 | [Name] | [N] | [N] | `feat/layer[N]-[name]` | None |
| [N].2 | [Name] | [N] | [N] | `feat/layer[N]-[name]` | [N].1 |

**Total: [Y] tests, [Z] files, [X] phases**

---

## Phase [N].1: [Name]

### Goal
[One sentence goal]

### Scope
- [Scope item 1]
- [Scope item 2]

### Required Implementation

```go
// Type or function definition
```

### Tests to Write ([N] tests, [N] commits)

#### Test 1: [TestName]
```go
func Test[Component]_[Behavior](t *testing.T)
```
**Acceptance Criteria:**
- [Criterion 1]
- [Criterion 2]

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement → verify GREEN → commit

**Total: [N] tests, [N] commits**

### Deliverables
- ✅ [Deliverable 1]
- ✅ [Deliverable 2]
- ✅ [N] commits (one per test per TDD workflow)

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Branch | Dependencies |
|-------|-------|-------|--------|--------------|
| [N].1 | [N] | [N] | `feat/layer[N]-[name]` | None |

### Execution Order

```
Phase [N].1
    ↓
Phase [N].2
    ↓
Phase [N].3
```

### Next Steps

1. **Start Phase [N].1**: Create branch `feat/layer[N]-[name]`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after final phase to verify integration
```

---

## 7. Branch Naming Convention

### 7.1 Branch Patterns

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feat/[layer]-[component]` | `feat/layer3-mail-types` |
| Bug Fix | `fix/gap/[ID]` | `fix/gap-g1` |
| Gap Phase | `fix/gap/[Phase]` | `fix/gap-g1.1` |

### 7.2 Actual Branch Names from Completed Work

```
feat/layer2-type-definitions
feat/layer2-communication
feat/layer2-observability
feat/layer2-lifecycle
feat/layer2-security
feat/layer2-registry

feat/layer3-mail-types
feat/layer3-mail-router
feat/layer3-inboxes
feat/layer3-pubsub
feat/layer3-streaming
feat/layer3-gateway
feat/layer3-human-gateway
feat/layer3-integration

fix/gap-g1
fix/gap-g2
fix/gap-g3
fix/gap-g4
fix/gap-g5
fix/gap-g6
```

---

## 8. Before/After: Work Decomposition

### 8.1 Before: High-Level Gap List

From `gap-remediation-plan.md`:

```
Phase G1: Critical Fixes (P0)
- L3-H4: ParseAddress function missing
- L3-C1: StreamSession.Send() and Close() panic
- L2-C1: sys:security boundary enforcement
- L2-C2: NamespaceIsolate method missing
- L2-C3: CheckTaintPolicy method missing
- L3-C3: sys:human-gateway chat endpoint missing
```

### 8.2 After: Detailed Phase Breakdown

From `gap-remediation-index.md`:

```
G1 (Critical Fixes) - P0
├── G1.1: ParseAddress Implementation (L3-H4) - 4 tests
├── G1.2: StreamSession Send/Close (L3-C1) - 4 tests
├── G1.3: Security Boundary Enforcement (L2-C1) - 2 tests
├── G1.4: NamespaceIsolate Method (L2-C2) - 2 tests
├── G1.5: CheckTaintPolicy Method (L2-C3) - 2 tests
└── G1.6: Human Gateway Chat Endpoint (L3-C3) - 2 tests
    └── Total: 16 tests, 16 commits
```

### 8.3 Decomposition Principles

1. **One gap per sub-phase** (G1.1 = L3-H4)
2. **Count tests per gap** (L3-H4 = 4 tests)
3. **1:1 commit ratio** (4 tests = 4 commits)
4. **Explicit dependencies** (G1.2 depends on G1.1)

---

## 9. Verification Checklist

### 9.1 Phase Document Checklist

- [ ] Parent phase and gap references documented
- [ ] Test count specified
- [ ] Dependencies listed
- [ ] SPEC section with requirements from arch-v1.md
- [ ] Files to create/modify listed
- [ ] Functions to implement with signatures
- [ ] Test scenarios enumerated
- [ ] TDD workflow with iterations
- [ ] Commit messages specified
- [ ] Deliverables listed

### 9.2 Test Checklist

- [ ] Test name follows `Test[Component]_[Behavior]` pattern
- [ ] Arrange-Act-Assert structure
- [ ] Clear error messages in t.Errorf
- [ ] Acceptance criteria documented

### 9.3 Commit Checklist

- [ ] One test per commit
- [ ] Commit message follows `[type](scope): description` pattern
- [ ] Gap ID referenced in scope
- [ ] All tests pass before commit

---

## Appendix A: Quick Reference

### A.1 File Locations

| Document Type | Location |
|---------------|----------|
| Phase breakdown | `docs/completed/layer-XX-phase-breakdown.md` |
| Phase detail | `docs/completed/phase-XX.YY-description.md` |
| Gap remediation | `docs/completed/gap-remediation/phase-gX.Y-description.md` |

### A.2 Key Metrics

| Metric | Value |
|--------|-------|
| Tests per phase | 2-10 |
| Commits per phase | = tests |
| Files per phase | 1-5 |
| Branch per phase | Yes |
| Merge after phase | Yes |

### A.3 Common Patterns

```go
// Service constructor
func NewServiceName() *ServiceName {
    return &ServiceName{}
}

// Service ID
func (s *ServiceName) ID() string {
    return "sys:service-name"
}

// Service interface
func (s *ServiceName) HandleMail(mail mail.Mail) error {
    // Implementation
}

// Thread-safe access
func (s *ServiceName) Method(params) (result, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // Implementation
}
```

---

**Document End**
