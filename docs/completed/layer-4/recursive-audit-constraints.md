# Constraint Coverage Audit - Layer 4

**Audit Date**: 2026-03-09  
**Scope**: All constraints from `arch-v1.md` and `spec-extraction.md` Section 5  
**Status**: Audit complete - gaps identified

---

## Executive Summary

This audit verifies that ALL constraints from the specification have corresponding validation tests across the 19 phase plans. Six primary constraints were identified and traced through the test plans.

**Coverage Summary**:
| Constraint | Coverage Status | Gaps |
|------------|-----------------|------|
| Security Invariant | ✅ 100% | None |
| Performance <50ms | ⚠️ Partial | No explicit benchmark test |
| Boundary Immutability | ❌ Missing | No tests in any phase |
| TaintPolicy Modes | ✅ 100% | None |
| allowedOnExit Policy | ✅ 100% | None |
| Per-chunk Stateless | ✅ 100% | None |

---

## Constraint Coverage Table

| Constraint | Phase | Test Type | Validation Method | Status | Spec Reference |
|------------|-------|-----------|-------------------|--------|----------------|
| **Security Invariant**: All data entering runtime is tainted at border | P1.1-AttachTaint | Unit | TestTaintEngine_AttachTaint_Empty verifies error on nil/untainted input | ✅ Covered | arch-v1.md L276 |
| | P6.3-ApplicationContextWiring | Integration | TestApplicationContext_SetTaints_StoresWithKey ensures all data stored has taints | ✅ Covered | arch-v1.md L276 |
| **Performance**: Stream sanitization <50ms | P5.1-StreamTaintStripping | Benchmark | TestSanitizeStreamChunk_LengthCap includes latency measurement | ⚠️ Partial | arch-v1.md L300 |
| | P5.2-StreamBoundaryEnforcement | Benchmark | TestEnforceStreamBoundary_Latency measures <50ms completion | ⚠️ Partial | arch-v1.md L300 |
| **Boundary Immutability**: Never migratable | None | N/A | No tests exist for boundary immutability | ❌ Missing | arch-v1.md L127, L1783 |
| **TaintPolicy Enforcement Modes**: strict/redact/audit | P4.1-TaintPolicyModes | Unit | TestPolicyEnforcement_Strict validates strict mode blocking | ✅ Covered | arch-v1.md L209, L916 |
| | P4.1-TaintPolicyModes | Unit | TestPolicyEnforcement_Redact validates redact mode replacement | ✅ Covered | arch-v1.md L209, L916 |
| | P4.1-TaintPolicyModes | Unit | TestPolicyEnforcement_Audit validates audit mode logging | ✅ Covered | arch-v1.md L209, L916 |
| **allowedOnExit Policy**: Taints allowed to leave runtime | P4.1-TaintPolicyModes | Unit | TestPolicyEnforcement_AllowedOnExit validates whitelist filtering | ✅ Covered | arch-v1.md L210-213, L916 |
| | P5.2-StreamBoundaryEnforcement | Integration | TestEnforceStreamBoundary_AllowedOnExit validates stream taint filtering | ✅ Covered | arch-v1.md L1317 |
| | P6.2-MailRouterHooks | Integration | TestMailRouter_AllowedOnExit_Stripping validates mail taint filtering | ✅ Covered | arch-v1.md L1317 |
| **Per-chunk Stateless Processing**: No buffering, independent chunks | P5.1-StreamTaintStripping | Unit | TestSanitizeStreamChunk_PII validates per-chunk processing | ✅ Covered | arch-v1.md L299 |
| | P5.2-StreamBoundaryEnforcement | Unit | TestEnforceStreamBoundary_PerChunk validates stateless processing | ✅ Covered | arch-v1.md L299 |
| | P5.2-StreamBoundaryEnforcement | Integration | TestEnforceStreamBoundary_NoBuffer validates no buffering | ✅ Covered | arch-v1.md L300 |

---

## Detailed Analysis by Constraint

### 1. Security Invariant: "All data entering the runtime is tainted at the border"

**Spec Reference**: arch-v1.md L276, spec-extraction.md L316-320

**Test Coverage**: ✅ 100%

| Phase | Test | Validation Method |
|-------|------|-------------------|
| P1.1-AttachTaint | TestTaintEngine_AttachTaint_Empty | Verifies nil object returns error, enforcing taint attachment |
| P1.1-AttachTaint | TestTaintEngine_AttachTaint_Mail | Verifies taints attached at Message creation |
| P6.3-ApplicationContextWiring | TestApplicationContext_SetTaints_StoresWithKey | Verifies all Set operations store taints with data |

**Edge Cases Covered**:
- Nil/empty input handling (P1.1 Test 4)
- Empty taint list handling (P1.1 Test 4)
- Nested structure taint attachment (P1.1 Test 3)

**Validation Method**: Unit tests verify taint attachment at all entry points; integration tests verify taints persist through ApplicationContext operations.

---

### 2. Performance: Stream sanitization <50ms

**Spec Reference**: arch-v1.md L299-300, spec-extraction.md L324-328

**Test Coverage**: ⚠️ PARTIAL

| Phase | Test | Validation Method |
|-------|------|-------------------|
| P5.1-StreamTaintStripping | TestSanitizeStreamChunk_LengthCap | Includes latency measurement in benchmark |
| P5.2-StreamBoundaryEnforcement | TestEnforceStreamBoundary_Latency | Measures processing time <50ms |

**Edge Cases Covered**:
- Typical size chunks (1KB)
- Length cap enforcement

**GAP IDENTIFIED**: 
- No explicit benchmark test with `go test -bench` annotation
- No load testing under concurrent conditions
- No worst-case scenario testing (max size chunks, complex taint sets)

**Recommendation**: Add explicit benchmark test:
```go
func BenchmarkSanitizeStreamChunk(b *testing.B) {
    // Verify <50ms under various load conditions
}
```

---

### 3. Boundary Immutability: "Never migratable"

**Spec Reference**: arch-v1.md L127, L1783, spec-extraction.md L332-335

**Test Coverage**: ❌ MISSING

**CRITICAL GAP**: No tests exist in any phase for boundary immutability.

**Required Tests**:
1. TestBoundaryImmutability_HotReload - Verify boundary unchanged after hot-reload
2. TestBoundaryImmutability_VersionUpgrade - Verify boundary unchanged across version changes
3. TestBoundaryImmutability_MigrationPolicy - Verify migrationPolicy does not affect boundary

**Recommendation**: Add new phase P7.1-BoundaryImmutability with tests for:
- Hot-reload preserves boundary
- Version upgrades preserve boundary
- Migration policies cannot change boundary

---

### 4. TaintPolicy Enforcement Modes: strict/redact/audit

**Spec Reference**: arch-v1.md L209, L916, spec-extraction.md L342-347

**Test Coverage**: ✅ 100%

| Phase | Test | Validation Method |
|-------|------|-------------------|
| P4.1-TaintPolicyModes | TestPolicyEnforcement_Strict | Verifies forbidden taints blocked with error |
| P4.1-TaintPolicyModes | TestPolicyEnforcement_Redact | Verifies forbidden values replaced with replacement text |
| P4.1-TaintPolicyModes | TestPolicyEnforcement_Audit | Verifies data passes through with violation logged |
| P4.1-TaintPolicyModes | TestPolicyEnforcement_MultipleRedactRules | Verifies multiple redact rules applied correctly |

**Edge Cases Covered**:
- Multiple redact rules per taint category (P4.1 Test 5)
- Empty taint lists (P4.1 Test 4)
- Policy mode transitions (implicit in tests)

**Validation Method**: Unit tests verify each enforcement mode produces correct behavior; integration tests verify modes work in context of full security service.

---

### 5. allowedOnExit Policy: Taints allowed to leave runtime

**Spec Reference**: arch-v1.md L210-213, L916, spec-extraction.md L348-350

**Test Coverage**: ✅ 100%

| Phase | Test | Validation Method |
|-------|------|-------------------|
| P4.1-TaintPolicyModes | TestPolicyEnforcement_AllowedOnExit | Verifies whitelist filtering at policy level |
| P5.2-StreamBoundaryEnforcement | TestEnforceStreamBoundary_AllowedOnExit | Verifies stream taint filtering |
| P6.2-MailRouterHooks | TestMailRouter_AllowedOnExit_Stripping | Verifies mail taint filtering |

**Edge Cases Covered**:
- Multiple taints with partial allowedOnExit (P4.1 Test 4)
- Stream chunks with forbidden taints (P5.2 Test 4)
- Mail routing with forbidden taints (P6.2 Test 3)

**Validation Method**: Unit tests verify policy-level filtering; integration tests verify filtering at stream and mail boundaries.

---

### 6. Per-chunk Stateless Processing: No buffering, independent chunks

**Spec Reference**: arch-v1.md L299, spec-extraction.md L324-328

**Test Coverage**: ✅ 100%

| Phase | Test | Validation Method |
|-------|------|-------------------|
| P5.1-StreamTaintStripping | TestSanitizeStreamChunk_PII | Verifies per-chunk processing |
| P5.2-StreamBoundaryEnforcement | TestEnforceStreamBoundary_PerChunk | Verifies stateless processing |
| P5.2-StreamBoundaryEnforcement | TestEnforceStreamBoundary_NoBuffer | Verifies no buffering |

**Edge Cases Covered**:
- Multiple sequential chunks with different taints (P5.2 Test 1)
- Real-time chunk emission (P5.2 Test 3)
- No state carried between chunks (P5.2 Test 1)

**Validation Method**: Unit tests verify per-chunk independence; integration tests verify real-time emission without buffering.

---

## Missing Constraints Summary

### Critical Gap: Boundary Immutability

**Constraint**: `metadata.boundary` is NEVER migratable (immutable identity)

**Spec References**:
- arch-v1.md L127: "Note: boundary is NEVER migratable — immutable identity"
- arch-v1.md L1783: "metadata.boundary — NEVER migratable (immutable identity)"
- spec-extraction.md L332-335: Boundary Immutability section

**Impact**: HIGH - Without tests, boundary could be accidentally changed during hot-reload or version migration, breaking security guarantees.

**Required Phase**: P7.1-BoundaryImmutability

**Required Tests**:
1. `TestBoundaryImmutability_HotReload` - Verify boundary unchanged after hot-reload
2. `TestBoundaryImmutability_VersionUpgrade` - Verify boundary unchanged across version changes  
3. `TestBoundaryImmutability_MigrationPolicy` - Verify migrationPolicy does not affect boundary
4. `TestBoundaryImmutability_ChartDefinition` - Verify ChartDefinition rejects boundary changes

### Medium Gap: Performance Benchmark

**Constraint**: Stream sanitization <50ms

**Spec References**:
- arch-v1.md L299: "Sanitization is per-chunk (stateless) so latency stays <50ms"
- spec-extraction.md L324-328: Performance Constraints section

**Impact**: MEDIUM - Tests exist but lack explicit benchmark annotations for CI verification.

**Required Enhancement**: Add benchmark tests to P5.1 and P5.2:
```go
func BenchmarkSanitizeStreamChunk(b *testing.B)
func BenchmarkEnforceStreamBoundary(b *testing.B)
```

---

## Recommendations

### Immediate Actions

1. **Create Phase P7.1-BoundaryImmutability** to test boundary immutability constraint
2. **Add benchmark tests** to P5.1 and P5.2 for explicit <50ms verification
3. **Add load testing** for concurrent stream chunk processing

### Future Considerations

1. Consider adding tests for Bloom filter optimization (arch-v1.md L1798, spec-extraction.md L328)
2. Consider adding tests for taint inheritance in nested structures (implementation-assessment.md L280-285)
3. Consider adding tests for DataSource taint persistence across restarts

---

## Appendix: Constraint-to-Phase Mapping

| Constraint ID | Constraint Name | Primary Phase | Secondary Phases |
|---------------|-----------------|---------------|------------------|
| C1 | Security Invariant | P1.1-AttachTaint | P6.3-ApplicationContextWiring |
| C2 | Performance <50ms | P5.2-StreamBoundaryEnforcement | P5.1-StreamTaintStripping |
| C3 | Boundary Immutability | MISSING | - |
| C4 | TaintPolicy Modes | P4.1-TaintPolicyModes | P4.2-ContextBlockFiltering |
| C5 | allowedOnExit | P6.2-MailRouterHooks | P4.1-TaintPolicyModes, P5.2-StreamBoundaryEnforcement |
| C6 | Per-chunk Stateless | P5.2-StreamBoundaryEnforcement | P5.1-StreamTaintStripping |

---

*Document generated: 2026-03-09*  
*Based on: arch-v1.md v1.3, spec-extraction.md, all 19 phase plans in docs/layer-4/plans/*
