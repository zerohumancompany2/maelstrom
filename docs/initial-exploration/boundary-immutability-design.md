# Boundary Immutability Design

Discussion date: 2026-02-25
Status: Core security invariant

---

## The Invariant

**Boundary is permanent identity.** Charts are born into a boundary and die there.

```yaml
metadata:
  name: ceo-agent
  version: "1.2.3"
  boundary: inner  # Permanent. Immutable. Forever.
```

---

## Security Property

This invariant prevents a class of attacks by construction.

### Attack Scenario Blocked

```
Attacker compromises ChartRegistry or git repo
├── Pushes ceo-agent-v2.yaml with boundary: outer
├── Attempts to impersonate trusted inner chart
└── BLOCKED: outer chart cannot access inner data, regardless of name
```

### Why Boundary == Identity

| Aspect | Implication |
|--------|-------------|
| **Name** | Human-readable; user-defined; not security-relevant |
| **Version** | Semver; indicates compatibility; mutable via hot-reload |
| **Boundary** | Security domain; machine-enforced; **never changes** |

Two charts with the same name but different boundaries are **different security identities** with **no data relationship**.

---

## Enforcement Mechanisms

### 1. Load-Time Rejection

ChartRegistry rejects any YAML where `(name, boundary)` tuple differs from existing registry entry:

```go
if existing, ok := registry.Get(name); ok {
    if existing.Boundary != newDef.Boundary {
        return fmt.Errorf("boundary mismatch: %s is %s, cannot load as %s",
            name, existing.Boundary, newDef.Boundary)
    }
}
```

### 2. No Runtime Migration

Even with `load-on-next-start` hot-reload, boundary is excluded from migratable properties:

```yaml
migrationPolicy:
  onVersionChange: shallowHistory | deepHistory | cleanStart
  # boundary is NOT in this list; it is identity, not configuration
```

### 3. Sub-Agent Inheritance

Child charts inherit parent's boundary **or stricter**:

| Parent | Allowed Children | Forbidden |
|--------|------------------|-----------|
| `inner` | `inner`, `dmz` | `outer` |
| `dmz` | `dmz` | `inner`, `outer` |
| `outer` | `outer` | `inner`, `dmz` |

```go
func CanSpawnChild(parentBoundary, childBoundary BoundaryType) bool {
    return childBoundary >= parentBoundary  // inner (0) < dmz (1) < outer (2)
}
```

---

## Operational Pattern

When a chart needs "boundary migration," create a new identity:

```
ceo-agent (inner) reaches end-of-life
├── Option A: Create ceo-agent-v2 (inner)
│   └── Fresh identity, same security domain, clean slate
├── Option B: Create ceo-assistant (dmz)
│   └── Different purpose, relaxed access, no secret access
└── Old chart: deprecate in registry, drain, stop
```

No identity continuity across boundaries. No migration ceremony. Clean semantics.

---

## Why Reject Alternatives

| Alternative | Why Rejected |
|-------------|--------------|
| **Boundary as configuration** | Attacker with Registry access can degrade security |
| **Explicit escalation only** | Adds complexity; still allows mutation; operational hazard |
| **Bootstrap-only change** | Creates "maintenance window" risk; violates "permanent" semantics |

The AppSec principle: **What cannot change cannot be attacked.**

---

## Security Guarantees

### For Data Exfiltration

```
outer:malicious-ceo-agent  --cannot access-->  inner:secrets
outer:malicious-ceo-agent  --cannot access-->  dmz:filtered-data
```

Outer charts operate blind to secure information, regardless of naming.

### For Prompt Injection

```
outer:user-input  --passes through-->  dmz:sanitizer  --before-->  inner:llm-prompt
```

User inputs never reach inner-boundary agents directly.

### The Remaining Attack Surface

Social engineering: tricking a human into trusting `outer:ceo-agent` as legitimate. This is **outside the system's threat model**—defended via out-of-band channels (documentation, URLs, certificate pinning, etc.).

---

## Exceptions = Impossible

There is no code path, no admin override, no emergency procedure for boundary mutation. Even with:

- Stolen admin credentials
- Compromised Kernel
- Physical server access

Boundary is identity; identity is permanent.

---

## Implementation Notes

### In ChartRegistry

```go
type ChartIdentity struct {
    Name      string       // "ceo-agent"
    Boundary  BoundaryType // inner | dmz | outer
    // version is NOT part of identity; mutable
}

func (r *Registry) ValidateReplacement(old, new ChartDefinition) error {
    if old.Identity() != new.Identity() {
        return ErrIdentityMismatch
    }
    return nil
}
```

### In Security Service

Boundary check happens at the outermost edge:

```go
func (s *Security) ValidateAccess(caller, target BoundaryType) error {
    // This check is redundant with Registry enforcement
    // but defense in depth demands it
    if caller < target {
        return ErrBoundaryViolation
    }
    return nil
}
```

---

## Related Documents

- `arch-v1.md` Section 4.1 (Boundary Model)
- `hot-reload-and-bootstrap-design.md` (Migration policy exclusions)
