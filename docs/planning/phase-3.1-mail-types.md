# Phase 3.1: Mail Core Types

## Goal
Define all Mail types, addressing formats, and metadata structures required by the communication layer.

## Scope
- Create `pkg/mail/types.go` with Mail, MailType, MailMetadata, Ack
- Create `pkg/mail/address.go` with address parsing and validation
- Define all 11 Mail types
- Implement address format validation (agent:<id>, topic:<name>, sys:<service>)
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Type | Status | Notes |
|------|--------|-------|
| `Mail` | ❌ Missing | Core message structure |
| `MailType` | ❌ Missing | 11 mail types enum |
| `MailMetadata` | ❌ Missing | Token, cost, taint metadata |
| `Ack` | ❌ Missing | Delivery acknowledgment |
| `StreamChunk` | ⚠️ Exists in Layer 2 | Will be reused |

### Files Status
| File | Status |
|------|-------|
| `pkg/mail/types.go` | ❌ MISSING - create |
| `pkg/mail/address.go` | ❌ MISSING - create |

## Required Implementation

### Mail Type
```go
// pkg/mail/types.go
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
```

### MailType Enum
```go
type MailType string

const (
    MailTypeUser            MailType = "user"
    MailTypeAssistant       MailType = "assistant"
    MailTypeToolResult      MailType = "tool_result"
    MailTypeToolCall        MailType = "tool_call"
    MailTypeMailReceived    MailType = "mail_received"
    MailTypeHeartbeat       MailType = "heartbeat"
    MailTypeError           MailType = "error"
    MailTypeHumanFeedback   MailType = "human_feedback"
    MailTypePartialAssistant MailType = "partial_assistant"
    MailTypeSubagentDone    MailType = "subagent_done"
    MailTypeTaintViolation  MailType = "taint_violation"
)
```

### MailMetadata
```go
type MailMetadata struct {
    Tokens   int
    Model    string
    Cost     float64
    Boundary BoundaryType
    Taints   []string
    Stream   bool
    IsFinal  bool
}

type BoundaryType string

const (
    InnerBoundary BoundaryType = "inner"
    DmzBoundary   BoundaryType = "dmz"
    OuterBoundary BoundaryType = "outer"
)
```

### Ack
```go
type Ack struct {
    CorrelationID string
    DeliveredAt   time.Time
}
```

### Address Validation
```go
// pkg/mail/address.go
func IsValidAgentAddress(addr string) bool
func IsValidTopicAddress(addr string) bool
func IsValidSysAddress(addr string) bool
func ParseAddress(addr string) (addrType AddressType, id string, err error)
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestMail_AddressFormats
```go
func TestMail_AddressFormats(t *testing.T) {
    // Test agent:<id> format
    if !IsValidAgentAddress("agent:recommendation-agent") {
        t.Error("Expected agent:recommendation-agent to be valid")
    }
    if IsValidAgentAddress("topic:market-data") {
        t.Error("Expected topic:market-data to be invalid for agent address")
    }
    
    // Test topic:<name> format
    if !IsValidTopicAddress("topic:market-data") {
        t.Error("Expected topic:market-data to be valid")
    }
    
    // Test sys:<service> format
    if !IsValidSysAddress("sys:heartbeat") {
        t.Error("Expected sys:heartbeat to be valid")
    }
    if !IsValidSysAddress("sys:persistence") {
        t.Error("Expected sys:persistence to be valid")
    }
    
    // Invalid formats
    if IsValidAgentAddress("invalid-format") {
        t.Error("Expected invalid-format to be rejected")
    }
}
```
**Acceptance Criteria:**
- `agent:<id>` format validated correctly
- `topic:<name>` format validated correctly
- `sys:<service>` format validated correctly
- Invalid formats rejected

### Test 2: TestMail_Types
```go
func TestMail_Types(t *testing.T) {
    types := []MailType{
        MailTypeUser,
        MailTypeAssistant,
        MailTypeToolResult,
        MailTypeToolCall,
        MailTypeMailReceived,
        MailTypeHeartbeat,
        MailTypeError,
        MailTypeHumanFeedback,
        MailTypePartialAssistant,
        MailTypeSubagentDone,
        MailTypeTaintViolation,
    }
    
    if len(types) != 11 {
        t.Errorf("Expected 11 mail types, got %d", len(types))
    }
    
    // Verify unique values
    seen := make(map[MailType]bool)
    for _, t := range types {
        if seen[t] {
            t.Errorf("Duplicate mail type: %s", t)
        }
        seen[t] = true
    }
    
    // Verify specific values
    if MailTypeUser != "user" {
        t.Errorf("Expected MailTypeUser to be 'user', got '%s'", MailTypeUser)
    }
    if MailTypeAssistant != "assistant" {
        t.Errorf("Expected MailTypeAssistant to be 'assistant', got '%s'", MailTypeAssistant)
    }
}
```
**Acceptance Criteria:**
- All 11 mail types defined
- Each type has unique string value
- Types can be compared and matched

### Test 3: TestMail_Metadata
```go
func TestMail_Metadata(t *testing.T) {
    meta := MailMetadata{
        Tokens:   150,
        Model:    "gpt-4",
        Cost:     0.03,
        Boundary: InnerBoundary,
        Taints:   []string{"USER_SUPPLIED", "TOOL_OUTPUT"},
        Stream:   false,
        IsFinal:  true,
    }
    
    if meta.Tokens != 150 {
        t.Errorf("Expected Tokens 150, got %d", meta.Tokens)
    }
    if meta.Model != "gpt-4" {
        t.Errorf("Expected Model 'gpt-4', got '%s'", meta.Model)
    }
    if meta.Boundary != InnerBoundary {
        t.Errorf("Expected Boundary InnerBoundary, got %s", meta.Boundary)
    }
    
    // Test boundary types
    if InnerBoundary != "inner" {
        t.Errorf("Expected InnerBoundary to be 'inner', got '%s'", InnerBoundary)
    }
    if DmzBoundary != "dmz" {
        t.Errorf("Expected DmzBoundary to be 'dmz', got '%s'", DmzBoundary)
    }
    if OuterBoundary != "outer" {
        t.Errorf("Expected OuterBoundary to be 'outer', got '%s'", OuterBoundary)
    }
    
    // Test empty taints
    emptyMeta := MailMetadata{Taints: []string{}}
    if len(emptyMeta.Taints) != 0 {
        t.Error("Expected empty Taints slice")
    }
}
```
**Acceptance Criteria:**
- MailMetadata has all required fields
- BoundaryType enum has inner, dmz, outer values
- Taints is a slice that can be empty or populated

### Test 4: TestMail_Structure
```go
func TestMail_Structure(t *testing.T) {
    mail := Mail{
        ID:            "msg-001",
        CorrelationID: "corr-001",
        Type:          MailTypeUser,
        CreatedAt:     time.Now(),
        Source:        "agent:user-agent",
        Target:        "agent:recommendation-agent",
        Content:       map[string]any{"text": "hello"},
        Metadata: MailMetadata{
            Tokens:   10,
            Boundary: OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }
    
    if mail.ID != "msg-001" {
        t.Errorf("Expected ID 'msg-001', got '%s'", mail.ID)
    }
    if mail.Type != MailTypeUser {
        t.Errorf("Expected Type MailTypeUser, got %s", mail.Type)
    }
    if mail.Source != "agent:user-agent" {
        t.Errorf("Expected Source 'agent:user-agent', got '%s'", mail.Source)
    }
    if mail.Target != "agent:recommendation-agent" {
        t.Errorf("Expected Target 'agent:recommendation-agent', got '%s'", mail.Target)
    }
    
    // Test Content accepts any type
    mail.Content = "string content"
    if mail.Content != "string content" {
        t.Error("Expected Content to accept string")
    }
    mail.Content = 42
    if mail.Content != 42 {
        t.Error("Expected Content to accept int")
    }
}
```
**Acceptance Criteria:**
- Mail struct has all required fields
- Can instantiate Mail with all fields populated
- Content field accepts any type

### Test 5: TestAck_Structure
```go
func TestAck_Structure(t *testing.T) {
    now := time.Now()
    ack := Ack{
        CorrelationID: "corr-001",
        DeliveredAt:   now,
    }
    
    if ack.CorrelationID != "corr-001" {
        t.Errorf("Expected CorrelationID 'corr-001', got '%s'", ack.CorrelationID)
    }
    if !ack.DeliveredAt.Equal(now) {
        t.Errorf("Expected DeliveredAt %v, got %v", now, ack.DeliveredAt)
    }
    
    // Test zero value
    zeroAck := Ack{}
    if zeroAck.CorrelationID != "" {
        t.Error("Expected zero value CorrelationID to be empty string")
    }
    if !zeroAck.DeliveredAt.IsZero() {
        t.Error("Expected zero value DeliveredAt to be zero time")
    }
}
```
**Acceptance Criteria:**
- Ack has CorrelationID and DeliveredAt fields
- Can instantiate Ack with values
- DeliveredAt is time.Time type

### Test 6: TestAddress_ParseAddress
```go
func TestAddress_ParseAddress(t *testing.T) {
    // Test agent address parsing
    addrType, id, err := ParseAddress("agent:recommendation-agent")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if addrType != AddressTypeAgent {
        t.Errorf("Expected AddressTypeAgent, got %v", addrType)
    }
    if id != "recommendation-agent" {
        t.Errorf("Expected id 'recommendation-agent', got '%s'", id)
    }
    
    // Test topic address parsing
    addrType, id, err = ParseAddress("topic:market-data")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if addrType != AddressTypeTopic {
        t.Errorf("Expected AddressTypeTopic, got %v", addrType)
    }
    if id != "market-data" {
        t.Errorf("Expected id 'market-data', got '%s'", id)
    }
    
    // Test sys address parsing
    addrType, id, err = ParseAddress("sys:heartbeat")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if addrType != AddressTypeSys {
        t.Errorf("Expected AddressTypeSys, got %v", addrType)
    }
    if id != "heartbeat" {
        t.Errorf("Expected id 'heartbeat', got '%s'", id)
    }
    
    // Test invalid format
    _, _, err = ParseAddress("invalid-format")
    if err == nil {
        t.Error("Expected error for invalid format")
    }
}
```
**Acceptance Criteria:**
- ParseAddress extracts type and id correctly
- Returns error for invalid formats
- Handles all three address formats

## Dependencies

### Test Dependencies
```
Test 1 → Test 6 (Address parsing)
Test 2 → Independent
Test 3 → Independent
Test 4 → Test 2, Test 3 (Mail structure)
Test 5 → Independent
```

### Phase Dependencies
- **None** - This is the first phase of Layer 3
- **Phases 3.2-3.8** depend on this phase completing first

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/types.go` | CREATE | Mail, MailType, MailMetadata, Ack, BoundaryType |
| `pkg/mail/address.go` | CREATE | Address validation and parsing |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement address validation functions → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Define all 11 MailType constants → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement MailMetadata and BoundaryType → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement Mail struct → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement Ack struct → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement ParseAddress function → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ Mail, MailType, MailMetadata, Ack types in `pkg/mail/types.go`
- ✅ Address validation in `pkg/mail/address.go`
- ✅ All 11 mail types defined
- ✅ 6 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 6 tests is within acceptable range (2-5 recommended, 6 is close)
- Types are tightly coupled (Mail depends on MailType, MailMetadata)
- Single coherent feature: Mail type definitions
- Splitting would create unnecessary fragmentation

**Alternative (if split needed):**
- 3.1a: Core types (Mail, MailType, MailMetadata, Ack) - 4 tests
- 3.1b: Address validation (ParseAddress, IsValid* functions) - 2 tests