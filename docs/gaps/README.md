# Documentation Gaps

This directory contains documented gaps between specifications and implementations.

---

## Current Gaps

### Layer 1 Minor Gaps

**[layer-01-minor-gaps.md](layer-01-minor-gaps.md)** - Non-blocking gaps from Layer 1 audit

| Gap | Priority | Effort | Status |
|-----|----------|--------|--------|
| Error path tests | Low | 3 hours | Documented |
| ChartRegistry | Medium | 14 hours | Defer to Layer 2 |
| File watching | Medium | 10 hours | Defer to Layer 3 |

**Impact:** None - Layer 2 can proceed without addressing these gaps.

---

## Gap Categories

### Critical Gaps
- Block progress to next layer
- Must be addressed before continuing
- **Current: 0**

### Minor Gaps
- Do not block progress
- Can be addressed in future sprints
- **Current: 3**

### Deferred Features
- Intentionally not implemented
- Scheduled for future layers
- **Current: 2** (ChartRegistry, File watching)

---

## Gap Lifecycle

```
Identified → Documented → Prioritized → Scheduled → Implemented → Closed
```

### Status Definitions

| Status | Meaning |
|--------|---------|
| Documented | Gap identified and documented |
| Scheduled | Assigned to sprint/layer |
| In Progress | Currently being implemented |
| Closed | Gap resolved |

---

## How to Use

### For Developers

1. **Before starting a layer:** Review gaps to understand known issues
2. **During implementation:** If you find a new gap, document it here
3. **After implementation:** Update gap status or close if resolved

### For Planning

1. **Sprint planning:** Review minor gaps for potential inclusion
2. **Layer planning:** Check if any gaps block the next layer
3. **Risk assessment:** Use gaps to identify potential risks

---

## Related Documents

- **Completed Work:** `../completed/INDEX.md`
- **Layer 1 Audit:** `../completed/layer-01-audit-report.md`
- **Architecture:** `../arch-v1.md`

---

## Gap Template

When documenting a new gap, use this template:

```markdown
## Gap X: [Title]

### Description
[What is missing or incomplete]

### Spec Reference
[Link to specification]

### Current Implementation
[What exists now]

### Missing
[What needs to be added]

### Effort Estimate
- Implementation: X hours
- Testing: Y hours
- Total: Z hours

### Recommendation
[When to implement, priority]
```

---

**Last Updated:** 2026-03-08  
**Total Gaps:** 3 (all minor)