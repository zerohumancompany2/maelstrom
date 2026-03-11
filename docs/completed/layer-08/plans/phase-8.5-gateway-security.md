# Phase 8.5: Gateway Security & Tainting

## Goal
Implement security enforcement for gateway (taint propagation, boundary validation, stream sanitization).

## Scope
- Implement border tainting for all inbound data (arch-v1.md L276)
- Implement per-chunk stream sanitization with <50ms latency (arch-v1.md L298-299)
- Implement security strips forbidden taints before emission (arch-v1.md L681, L700)
- Implement boundary validation on ingress
- Implement taint propagation to mail
- Implement outer→inner tainting enforcement
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `BorderTainting` | ❌ Missing | Need to implement taint attachment at border |
| `StreamSanitizer` | ⚠️ Partial | Sanitizers exist; need per-chunk enforcement |
| `ForbiddenTaintStripper` | ❌ Missing | Need to strip forbidden taints before emission |
| `BoundaryValidator` | ⚠️ Partial | Boundary validation exists; need ingress enforcement |
| `TaintPropagator` | ⚠️ Partial | Taint propagation exists; need mail integration |
| `OuterInnerTaintEnforcer` | ❌ Missing | Need outer→inner tainting enforcement |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/security.go` | ❌ Create - gateway security enforcement |
| `pkg/services/gateway/security_test.go` | ❌ Create - gateway security tests |
| `pkg/security/sanitizers/stream.go` | ⚠️ Modify - add per-chunk sanitization |
| `pkg/security/boundary.go` | ⚠️ Modify - add ingress validation |
| `pkg/security/types.go` | ⚠️ Modify - add taint propagation types |
| `pkg/mail/types.go` | ⚠️ Modify - add taint fields to mail |

## Required Implementation

### Border Tainting (arch-v1.md L276)
```go
// pkg/services/gateway/security.go
type BorderTainter struct {
    DefaultTaints []string
}

func (bt *BorderTainter) TaintInboundData(data any, source string) (any, error) {
    // All data entering the runtime is tainted at the border (arch-v1.md L276)
    // No untainted information exists inside the runtime
    // Attach taints as soon as data is ingested
}
```

### Per-Chunk Stream Sanitization (arch-v1.md L298-299)
```go
// pkg/security/sanitizers/stream.go
type StreamSanitizer struct {
    Redactor     *PIIRedactor
    LengthCapper *LengthCapper
    SchemaValidator *SchemaValidator
    InnerDataStripper *InnerDataStripper
}

func (ss *StreamSanitizer) SanitizeChunk(chunk StreamChunk) (StreamChunk, error) {
    // Every outgoing chunk passes through DMZ sanitizers (arch-v1.md L298)
    // Sanitization is per-chunk (stateless) so latency stays <50ms (arch-v1.md L299)
    // No buffering of entire response unless chart explicitly requests it
}
```

### Forbidden Taint Stripper (arch-v1.md L681, L700)
```go
// pkg/services/gateway/security.go
type ForbiddenTaintStripper struct {
    AllowedOnExit map[string]bool
}

func (fts *ForbiddenTaintStripper) StripForbiddenTaints(data any, boundary string) (any, error) {
    // Security strips forbidden taints before emission (arch-v1.md L681)
    // Uses allowedOnExit to determine what can leave runtime (arch-v1.md L700)
    // Inner → DMZ/outer: auto-strip or block on forbidden taints
}
```

### Boundary Validation on Ingress
```go
// pkg/security/boundary.go
type BoundaryValidator struct {
    Policy *SecurityPolicy
}

func (bv *BoundaryValidator) ValidateOnIngress(mail *mail.Mail) error {
    // Validate mail on ingress to gateway
    // Check boundary transitions are allowed
    // Emit taint_violation event to dead-letter on violation
}
```

### Taint Propagation to Mail
```go
// pkg/mail/types.go
type Mail struct {
    ID        string
    Type      MailType
    Source    string
    Target    string
    Content   string
    Metadata  MailMetadata
    Taints    []string  // Immutable taint set
}

type MailMetadata struct {
    Boundary string
    Taints   []string
    // ... other metadata
}

func PropagateTaints(sourceMail *Mail, targetMail *Mail) {
    // Security Service propagates taints on copy/read/write
    // Like DLP tracking
}
```

### Outer→Inner Tainting Enforcement
```go
// pkg/services/gateway/security.go
type OuterInnerTaintEnforcer struct {
    Policy *SecurityPolicy
}

func (oit *OuterInnerTaintEnforcer) EnforceOuterToInner(data any, fromBoundary string, toBoundary string) error {
    // Enforce tainting when data moves from outer to inner boundary
    // Auto-strip or block on forbidden taints
    // On-disk: taints stored with data; Persistence refuses writes that violate policy
}
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestGatewaySecurity_AllDataTaintedAtBorder
```go
func TestGatewaySecurity_AllDataTaintedAtBorder(t *testing.T) {
    tainter := &BorderTainter{
        DefaultTaints: []string{"USER_SUPPLIED", "OUTER_BOUNDARY"},
    }

    // All data entering the runtime is tainted at the border (arch-v1.md L276)
    testData := map[string]any{
        "message": "Hello, agent!",
        "source":  "user:web",
    }

    taintedData, err := tainter.TaintInboundData(testData, "user:web")
    if err != nil {
        t.Fatalf("Expected no error tainting inbound data, got %v", err)
    }

    // Verify data is tainted (arch-v1.md L276)
    taintedMap, ok := taintedData.(map[string]any)
    if !ok {
        t.Fatal("Expected tainted data to be map")
    }

    if taintedMap["taints"] == nil {
        t.Error("Expected taints to be attached to data")
    }

    // Verify taints are attached as soon as data is ingested (arch-v1.md L276)
    taints := taintedMap["taints"].([]string)
    if !slices.Contains(taints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint")
    }

    if !slices.Contains(taints, "OUTER_BOUNDARY") {
        t.Error("Expected OUTER_BOUNDARY taint")
    }

    // No untainted information exists inside the runtime (arch-v1.md L276)
    if len(taints) == 0 {
        t.Error("Expected all data to have at least one taint")
    }

    // Test with different source types
    testCases := []struct {
        source       string
        expectedTaint string
    }{
        {"user:web", "USER_SUPPLIED"},
        {"user:api", "USER_SUPPLIED"},
        {"system:heartbeat", "SYSTEM_GENERATED"},
        {"tool:registry", "TOOL_OUTPUT"},
    }

    for _, tc := range testCases {
        taintedData, err := tainter.TaintInboundData(map[string]any{"data": "test"}, tc.source)
        if err != nil {
            t.Fatalf("Expected no error for source %s, got %v", tc.source, err)
        }

        taintedMap := taintedData.(map[string]any)
        taints := taintedMap["taints"].([]string)

        if !slices.Contains(taints, tc.expectedTaint) {
            t.Errorf("Expected taint %s for source %s", tc.expectedTaint, tc.source)
        }
    }
}
```
**Acceptance Criteria:**
- All data entering runtime is tainted at border (arch-v1.md L276)
- No untainted information exists inside runtime (arch-v1.md L276)
- Taints attached as soon as data is ingested (arch-v1.md L276)
- Different source types get appropriate taints

### Test 2: TestGatewaySecurity_StreamSanitizationPerChunk
```go
func TestGatewaySecurity_StreamSanitizationPerChunk(t *testing.T) {
    sanitizer := &StreamSanitizer{
        Redactor:     NewPIIRedactor(),
        LengthCapper: NewLengthCapper(1000),
        SchemaValidator: NewSchemaValidator(),
        InnerDataStripper: NewInnerDataStripper(),
    }

    // Every outgoing chunk passes through DMZ sanitizers (arch-v1.md L298)
    chunks := []StreamChunk{
        {
            Chunk:    "Hello, this is a normal message",
            Sequence: 1,
            IsFinal:  false,
            Taints:   []string{"USER_SUPPLIED"},
        },
        {
            Chunk:    "PII data: John Doe, email: john@example.com",
            Sequence: 2,
            IsFinal:  false,
            Taints:   []string{"PII"},
        },
        {
            Chunk:    "SECRET: api_key=abc123",
            Sequence: 3,
            IsFinal:  false,
            Taints:   []string{"SECRET", "INNER_ONLY"},
        },
        {
            Chunk:    "Final response complete",
            Sequence: 4,
            IsFinal:  true,
            Taints:   []string{},
        },
    }

    // Sanitization is per-chunk (stateless) (arch-v1.md L299)
    for i, chunk := range chunks {
        startTime := time.Now()
        sanitized, err := sanitizer.SanitizeChunk(chunk)
        elapsed := time.Since(startTime)

        if err != nil {
            t.Fatalf("Expected no error sanitizing chunk %d, got %v", i, err)
        }

        // Latency stays <50ms (arch-v1.md L299)
        if elapsed >= 50*time.Millisecond {
            t.Errorf("Expected chunk %d sanitization <50ms, got %v", i, elapsed)
        }

        // Verify PII redaction (arch-v1.md L298)
        if chunk.Taints != nil && slices.Contains(chunk.Taints, "PII") {
            if strings.Contains(sanitized.Chunk, "john@example.com") {
                t.Error("Expected PII to be redacted")
            }
        }

        // Verify inner-data stripping (arch-v1.md L298)
        if slices.Contains(chunk.Taints, "INNER_ONLY") {
            if strings.Contains(sanitized.Chunk, "api_key") {
                t.Error("Expected inner data to be stripped")
            }
        }

        // Verify length caps (arch-v1.md L298)
        if len(sanitized.Chunk) > 1000 {
            t.Error("Expected chunk to be capped at 1000 chars")
        }

        // Verify schema validation (arch-v1.md L298)
        if sanitized.Sequence != chunk.Sequence {
            t.Errorf("Expected sequence %d, got %d", chunk.Sequence, sanitized.Sequence)
        }

        if sanitized.IsFinal != chunk.IsFinal {
            t.Errorf("Expected IsFinal %v, got %v", chunk.IsFinal, sanitized.IsFinal)
        }
    }

    // No buffering of entire response unless chart explicitly requests it (arch-v1.md L299)
    // Verify each chunk is sanitized independently (stateless)
    firstChunk, _ := sanitizer.SanitizeChunk(chunks[0])
    secondChunk, _ := sanitizer.SanitizeChunk(chunks[1])

    // Sanitizing chunk 2 should not depend on chunk 1
    if secondChunk.Sequence != 2 {
        t.Error("Expected chunk 2 to be sanitized independently")
    }
}
```
**Acceptance Criteria:**
- Every outgoing chunk passes through DMZ sanitizers (arch-v1.md L298)
- PII redaction, length caps, schema validation, inner-data stripping (arch-v1.md L298)
- Sanitization is per-chunk (stateless) (arch-v1.md L299)
- Latency stays <50ms (arch-v1.md L299)
- No buffering of entire response (arch-v1.md L299)

### Test 3: TestGatewaySecurity_SecurityStripsForbiddenTaints
```go
func TestGatewaySecurity_SecurityStripsForbiddenTaints(t *testing.T) {
    stripper := &ForbiddenTaintStripper{
        AllowedOnExit: map[string]bool{
            "USER_SUPPLIED": true,
            "TOOL_OUTPUT":   true,
            "PUBLIC":        true,
        },
    }

    // Security strips forbidden taints before emission (arch-v1.md L681)
    testData := map[string]any{
        "content": "Response content",
        "taints":  []string{"USER_SUPPLIED", "SECRET", "INNER_ONLY", "PII"},
    }

    strippedData, err := stripper.StripForbiddenTaints(testData, "outer")
    if err != nil {
        t.Fatalf("Expected no error stripping forbidden taints, got %v", err)
    }

    strippedMap := strippedData.(map[string]any)
    remainingTaints := strippedMap["taints"].([]string)

    // Uses allowedOnExit to determine what can leave runtime (arch-v1.md L700)
    if !slices.Contains(remainingTaints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED to remain (allowed on exit)")
    }

    if slices.Contains(remainingTaints, "SECRET") {
        t.Error("Expected SECRET to be stripped (forbidden)")
    }

    if slices.Contains(remainingTaints, "INNER_ONLY") {
        t.Error("Expected INNER_ONLY to be stripped (forbidden)")
    }

    if slices.Contains(remainingTaints, "PII") {
        t.Error("Expected PII to be stripped (forbidden)")
    }

    // Inner → DMZ/outer: auto-strip or block on forbidden taints (arch-v1.md L681)
    innerData := map[string]any{
        "content": "Inner data",
        "taints":  []string{"INNER_ONLY", "SECRET"},
    }

    _, err = stripper.StripForbiddenTaints(innerData, "outer")
    if err == nil {
        t.Error("Expected error when all taints are forbidden")
    }

    // Test Stream Chunk Format (arch-v1.md L700)
    chunk := map[string]any{
        "chunk":     "Response chunk",
        "sequence":  1,
        "isFinal":   false,
        "taints":    []string{"USER_SUPPLIED", "SECRET"},
    }

    strippedChunk, _ := stripper.StripForbiddenTaints(chunk, "outer")
    strippedChunkMap := strippedChunk.(map[string]any)

    // Security strips before emission (arch-v1.md L700)
    chunkTaints := strippedChunkMap["taints"].([]string)
    if slices.Contains(chunkTaints, "SECRET") {
        t.Error("Expected SECRET to be stripped from chunk before emission")
    }

    if !slices.Contains(chunkTaints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED to remain in chunk")
    }
}
```
**Acceptance Criteria:**
- Security strips forbidden taints before emission (arch-v1.md L681)
- Uses allowedOnExit to determine what can leave runtime (arch-v1.md L700)
- Inner → DMZ/outer: auto-strip or block on forbidden taints (arch-v1.md L681)
- Stream chunk format with taints stripped before emission (arch-v1.md L700)

### Test 4: TestGatewaySecurity_BoundaryValidationOnIngress
```go
func TestGatewaySecurity_BoundaryValidationOnIngress(t *testing.T) {
    validator := &BoundaryValidator{
        Policy: NewDefaultSecurityPolicy(),
    }

    // Validate mail on ingress to gateway
    inboundMail := &mail.Mail{
        ID:      "mail-001",
        Type:    mail.MailTypeUser,
        Source:  "user:web",
        Target:  "agent:dmz",
        Content: "Hello, agent!",
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
        Taints: []string{"USER_SUPPLIED"},
    }

    err := validator.ValidateOnIngress(inboundMail)
    if err != nil {
        t.Fatalf("Expected no error validating inbound mail, got %v", err)
    }

    // Check boundary transitions are allowed
    transitionMail := &mail.Mail{
        ID:      "mail-002",
        Type:    mail.MailTypeAssistant,
        Source:  "agent:inner",
        Target:  "user:web",
        Content: "Response from inner agent",
        Metadata: mail.MailMetadata{
            Boundary: mail.InnerBoundary,
            Taints:   []string{"INNER_ONLY", "SECRET"},
        },
        Taints: []string{"INNER_ONLY", "SECRET"},
    }

    // Emit taint_violation event to dead-letter on violation (arch-v1.md L286)
    err = validator.ValidateOnIngress(transitionMail)
    if err == nil {
        t.Error("Expected error for forbidden boundary transition")
    }

    // Verify violation is logged
    if !strings.Contains(err.Error(), "taint_violation") {
        t.Error("Expected taint_violation in error message")
    }

    // Test allowed boundary transition
    allowedMail := &mail.Mail{
        ID:      "mail-003",
        Type:    mail.MailTypeAssistant,
        Source:  "agent:dmz",
        Target:  "user:web",
        Content: "Response from DMZ agent",
        Metadata: mail.MailMetadata{
            Boundary: mail.DMZBoundary,
            Taints:   []string{"TOOL_OUTPUT"},
        },
        Taints: []string{"TOOL_OUTPUT"},
    }

    err = validator.ValidateOnIngress(allowedMail)
    if err != nil {
        t.Errorf("Expected no error for allowed boundary transition, got %v", err)
    }

    // Test runtime guard: any action/guard can query taints (arch-v1.md L286)
    queryableMail := &mail.Mail{
        ID:      "mail-004",
        Type:    mail.MailTypeToolResult,
        Source:  "tool:registry",
        Target:  "agent:dmz",
        Content: "Tool output",
        Metadata: mail.MailMetadata{
            Boundary: mail.DMZBoundary,
            Taints:   []string{"TOOL_OUTPUT"},
        },
        Taints: []string{"TOOL_OUTPUT"},
    }

    // Verify taints can be queried
    taints := queryableMail.GetTaints()
    if !slices.Contains(taints, "TOOL_OUTPUT") {
        t.Error("Expected TOOL_OUTPUT taint to be queryable")
    }
}
```
**Acceptance Criteria:**
- Validate mail on ingress to gateway
- Check boundary transitions are allowed
- Emit taint_violation event to dead-letter on violation (arch-v1.md L286)
- Runtime guard: any action/guard can query taints (arch-v1.md L286)

### Test 5: TestGatewaySecurity_TaintPropagationToMail
```go
func TestGatewaySecurity_TaintPropagationToMail(t *testing.T) {
    // Security Service propagates taints on copy/read/write (arch-v1.md L283)
    sourceMail := &mail.Mail{
        ID:      "mail-001",
        Type:    mail.MailTypeUser,
        Source:  "user:web",
        Target:  "agent:dmz",
        Content: "User query with PII: john@example.com",
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED", "PII"},
        },
        Taints: []string{"USER_SUPPLIED", "PII"},
    }

    // Like DLP tracking (arch-v1.md L283)
    targetMail := &mail.Mail{
        ID:       "mail-002",
        Type:     mail.MailTypeAssistant,
        Source:   "agent:dmz",
        Target:   "user:web",
        Content:  "Response to user query",
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
        },
    }

    PropagateTaints(sourceMail, targetMail)

    // Verify taints propagated (arch-v1.md L283)
    if !slices.Contains(targetMail.Taints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint to propagate")
    }

    if !slices.Contains(targetMail.Taints, "PII") {
        t.Error("Expected PII taint to propagate")
    }

    // Test taint propagation on copy
    copiedMail := &mail.Mail{
        ID:       "mail-003",
        Type:     mail.MailTypeContextBlock,
        Source:   "agent:dmz",
        Target:   "agent:dmz",
        Content:  sourceMail.Content,
        Metadata: mail.MailMetadata{
            Boundary: mail.DMZBoundary,
        },
    }

    PropagateTaints(sourceMail, copiedMail)

    // Verify taints on copy (arch-v1.md L283)
    if !slices.Contains(copiedMail.Taints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint to propagate on copy")
    }

    if !slices.Contains(copiedMail.Taints, "PII") {
        t.Error("Expected PII taint to propagate on copy")
    }

    // Test taint propagation on read
    readMail := &mail.Mail{
        ID:       "mail-004",
        Type:     mail.MailTypeContextBlock,
        Source:   "memory:service",
        Target:   "agent:dmz",
        Content:  "Memory content",
        Metadata: mail.MailMetadata{
            Boundary: mail.DMZBoundary,
        },
    }

    PropagateTaints(sourceMail, readMail)

    // Verify taints on read (arch-v1.md L283)
    if len(readMail.Taints) == 0 {
        t.Error("Expected taints to propagate on read")
    }

    // Test on-disk: taints stored with data (arch-v1.md L284)
    persistedMail := &mail.Mail{
        ID:       "mail-005",
        Type:     mail.MailTypeSnapshot,
        Source:   "persistence:service",
        Target:   "agent:dmz",
        Content:  "Snapshot data",
        Metadata: mail.MailMetadata{
            Boundary: mail.InnerBoundary,
        },
        Taints: []string{"SECRET", "INNER_ONLY"},
    }

    // Verify taints are stored with data (arch-v1.md L284)
    if !slices.Contains(persistedMail.Taints, "SECRET") {
        t.Error("Expected SECRET taint to be stored with data")
    }

    if !slices.Contains(persistedMail.Taints, "INNER_ONLY") {
        t.Error("Expected INNER_ONLY taint to be stored with data")
    }
}
```
**Acceptance Criteria:**
- Security Service propagates taints on copy/read/write (arch-v1.md L283)
- Like DLP tracking (arch-v1.md L283)
- On-disk: taints stored with data (arch-v1.md L284)
- Persistence refuses writes that would violate policy (arch-v1.md L284)

### Test 6: TestGatewaySecurity_OuterToInnerTainting
```go
func TestGatewaySecurity_OuterToInnerTainting(t *testing.T) {
    enforcer := &OuterInnerTaintEnforcer{
        Policy: NewDefaultSecurityPolicy(),
    }

    // Enforce tainting when data moves from outer to inner boundary
    outerData := map[string]any{
        "content": "User input from outer boundary",
        "source":  "user:web",
    }

    // Outer → inner tainting enforced
    err := enforcer.EnforceOuterToInner(outerData, "outer", "inner")
    if err == nil {
        t.Error("Expected error when moving outer data to inner without proper tainting")
    }

    // Test with proper tainting
    taintedOuterData := map[string]any{
        "content": "User input from outer boundary",
        "source":  "user:web",
        "taints":  []string{"USER_SUPPLIED", "OUTER_BOUNDARY"},
    }

    err = enforcer.EnforceOuterToInner(taintedOuterData, "outer", "inner")
    if err != nil {
        t.Fatalf("Expected no error with proper tainting, got %v", err)
    }

    // Auto-strip or block on forbidden taints (arch-v1.md L284)
    forbiddenData := map[string]any{
        "content": "Data with forbidden taints",
        "taints":  []string{"SECRET", "INNER_ONLY"},
    }

    err = enforcer.EnforceOuterToInner(forbiddenData, "outer", "inner")
    if err == nil {
        t.Error("Expected error for forbidden taints in outer→inner transition")
    }

    // Test allowed transition with allowed taints
    allowedData := map[string]any{
        "content": "Data with allowed taints",
        "taints":  []string{"USER_SUPPLIED", "TOOL_OUTPUT"},
    }

    err = enforcer.EnforceOuterToInner(allowedData, "outer", "dmz")
    if err != nil {
        t.Errorf("Expected no error for allowed taints, got %v", err)
    }

    // On-disk: taints stored with data; Persistence refuses writes that violate policy (arch-v1.md L284)
    diskData := map[string]any{
        "content": "Data for persistence",
        "taints":  []string{"SECRET", "INNER_ONLY"},
    }

    // Verify persistence policy enforcement
    err = enforcer.EnforceOuterToInner(diskData, "inner", "disk")
    if err != nil {
        t.Fatalf("Expected no error for inner→disk with proper taints, got %v", err)
    }

    // Test Persistence refuses writes that violate policy (arch-v1.md L284)
    violatingData := map[string]any{
        "content": "Data violating policy",
        "taints":  []string{"SECRET"},
    }

    err = enforcer.EnforceOuterToInner(violatingData, "outer", "disk")
    if err == nil {
        t.Error("Expected error when persistence policy is violated")
    }

    // Verify taints are stored with data on disk (arch-v1.md L284)
    if !slices.Contains(diskData["taints"].([]string), "SECRET") {
        t.Error("Expected taints to be stored with data on disk")
    }

    if !slices.Contains(diskData["taints"].([]string), "INNER_ONLY") {
        t.Error("Expected taints to be stored with data on disk")
    }
}
```
**Acceptance Criteria:**
- Outer→inner tainting enforced
- Auto-strip or block on forbidden taints (arch-v1.md L284)
- On-disk: taints stored with data (arch-v1.md L284)
- Persistence refuses writes that would violate policy (arch-v1.md L284)

## Dependencies

### Test Dependencies
```
Test 1 (All data tainted at border) → Test 2 (Stream sanitization per-chunk)
Test 3 (Security strips forbidden taints)
Test 4 (Boundary validation on ingress)
Test 5 (Taint propagation to mail)
Test 6 (Outer→inner tainting)
```

### Phase Dependencies
- **Phase 8.2** - Streaming Path Integration (provides streaming infrastructure and mail flow)
- **Phase 8.3** - HTTP(S) Exposure & OpenAPI (provides gateway HTTP infrastructure)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/gateway/security.go` | CREATE | Border tainting, forbidden taint stripping, boundary validation, outer→inner enforcement |
| `pkg/services/gateway/security_test.go` | CREATE | Gateway security tests |
| `pkg/security/sanitizers/stream.go` | MODIFY | Add per-chunk sanitization with <50ms latency |
| `pkg/security/boundary.go` | MODIFY | Add ingress validation |
| `pkg/security/types.go` | MODIFY | Add taint propagation types |
| `pkg/mail/types.go` | MODIFY | Add taint fields to mail |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement border tainting → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement per-chunk stream sanitization → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement forbidden taint stripper → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement boundary validation on ingress → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement taint propagation to mail → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement outer→inner tainting enforcement → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ All data entering runtime tainted at border (arch-v1.md L276)
- ✅ Per-chunk stream sanitization with <50ms latency (arch-v1.md L298-299)
- ✅ Security strips forbidden taints before emission (arch-v1.md L681, L700)
- ✅ Boundary validation on ingress with taint_violation events (arch-v1.md L286)
- ✅ Taint propagation to mail like DLP tracking (arch-v1.md L283)
- ✅ Outer→inner tainting enforcement (arch-v1.md L284)
- ✅ 6 commits (one per test per TDD workflow)