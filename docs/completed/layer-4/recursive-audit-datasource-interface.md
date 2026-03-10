# DataSource Interface Coverage Audit

**Audit Date**: 2026-03-09  
**Scope**: DataSource interface methods × implementations  
**Source**: arch-v1.md L787-791

---

## DataSource Interface Definition

From `arch-v1.md L787-791`:

```go
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary string) error
}
```

---

## Coverage Matrix

| Method | Implementation | Phase | Tests | Count | Status |
|--------|---------------|-------|-------|-------|--------|
| TagOnWrite | LocalDisk | P2.1 | TestLocalDisk_TagOnWrite | 1 | ✅ PLANNED |
| TagOnWrite | S3 | P2.2 | TestS3DataSource_TagOnWrite | 1 | ✅ PLANNED |
| TagOnWrite | InMemory | P2.3 | TestInMemory_TagOnWrite | 1 | ✅ PLANNED |
| TagOnWrite | Registry | P2.4 | N/A (registry pattern) | 0 | ⚠️ N/A |
| GetTaints | LocalDisk | P2.1 | TestLocalDisk_GetTaints | 1 | ✅ PLANNED |
| GetTaints | S3 | P2.2 | TestS3DataSource_GetTaints | 1 | ✅ PLANNED |
| GetTaints | InMemory | P2.3 | TestInMemory_GetTaints | 1 | ✅ PLANNED |
| GetTaints | Registry | P2.4 | N/A (registry pattern) | 0 | ⚠️ N/A |
| ValidateAccess | LocalDisk | P2.1 | TestLocalDisk_ValidateAccess_Allowed, TestLocalDisk_ValidateAccess_Denied | 2 | ✅ PLANNED |
| ValidateAccess | S3 | P2.2 | TestS3DataSource_ValidateAccess | 1 | ✅ PLANNED |
| ValidateAccess | InMemory | P2.3 | *None* | 0 | ❌ GAP |
| ValidateAccess | Registry | P2.4 | N/A (registry pattern) | 0 | ⚠️ N/A |

---

## Edge Cases Coverage

| Implementation | Edge Cases | Tests | Status |
|---------------|------------|-------|--------|
| LocalDisk | Sidecar fallback for non-xattr filesystems | TestLocalDisk_SidecarFallback | ✅ PLANNED |
| S3 | Config parsing (bucket, region, endpoint) | TestS3DataSource_Config | ✅ PLANNED |
| InMemory | Concurrent access / thread safety | TestInMemory_Concurrent | ✅ PLANNED |
| Registry | N/A | N/A | ⚠️ N/A |

---

## Gaps Found

### Gap 1: InMemory ValidateAccess Not Tested

**Location**: P2.3-InMemory.md  
**Issue**: The `inMemoryDataSource` struct includes `allowedForBoundary []security.BoundaryType` field (line 76) and declares `ValidateAccess` method (line 82), but **no test is planned** for this method.

**Impact**: ValidateAccess boundary enforcement for InMemory DataSource is untested.

**Recommended Fix**: Add test to P2.3-InMemory.md:
```
### Test 4: TestInMemory_ValidateAccess

**Given**: An InMemory DataSource with allowedForBoundary=["inner"]
**When**: ValidateAccess is called with boundary "inner", then "outer"
**Then**: Returns nil for "inner", returns error for "outer"
**Expected Result**: Boundary validation works correctly
**Spec Reference**: arch-v1.md L789 (ValidateAccess interface), L760 (allowedForBoundary)
```

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| Total Interface Methods | 3 |
| Total Implementations | 4 |
| Total Method × Implementation Combinations | 12 |
| Combinations with Tests Planned | 9 |
| Combinations with Gaps | 1 |
| N/A (Registry pattern) | 2 |
| Test Coverage Rate | 90% (9/10 applicable) |

---

## Phase Status Overview

| Phase | Title | Status | Tests Planned |
|-------|-------|--------|---------------|
| P2.1 | LocalDisk xattr Implementation | ❌ PENDING | 5 |
| P2.2 | S3 DataSource Implementation | ❌ PENDING | 4 |
| P2.3 | InMemory DataSource Implementation | ❌ PENDING | 3 |
| P2.4 | DataSource Registry Implementation | ⚠️ PARTIAL | 3 |

---

## Recommendations

1. **Add ValidateAccess test to P2.3-InMemory.md** - Missing test for boundary validation
2. **All phases remain PENDING** - No implementation or tests exist yet
3. **Registry tests are verification-only** - P2.4 tests verify existing registry, not DataSource methods

---

*Audit completed: 2026-03-09*  
*Based on: arch-v1.md v1.3, P2.1-LocalDisk.md, P2.2-S3.md, P2.3-InMemory.md, P2.4-Registry.md*
