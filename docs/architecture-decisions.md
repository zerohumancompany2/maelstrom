# Architecture Decision Records

This document records key architectural decisions made during the development of Maelstrom Phase 1.

---

## ADR-001: Registry-per-Type Pattern

### Context

The system needed a way to store and retrieve typed objects (Charts, Releases, etc.) while working within Go's type system. We needed to:
- Store heterogeneous types in a unified way
- Provide type-safe access at API boundaries
- Avoid excessive complexity or performance overhead

Three approaches were considered:
1. **Generics**: `Registry[T any]` with type parameters
2. **interface{} dispatch**: Store `interface{}`, cast everywhere
3. **Registry-per-Type**: Separate registry instances per type, type assertion at API boundary

### Decision

Use the **Registry-per-Type Pattern**: each type has its own registry instance storing `interface{}`, with type assertions performed at the API boundary via `GetID()` methods.

```go
// Registry stores interface{} internally
type Registry struct {
    items map[string]interface{}
    mu    sync.RWMutex
}

// Per-type registries provide type-safe access
var ChartRegistry = NewRegistry()
var ReleaseRegistry = NewRegistry()

// Type assertion at boundary via GetID()
func (c *Chart) GetID() string { return c.Name }
```

### Consequences

**Positive:**
- Runtime type safety through API boundary assertions
- No generics complexity (works with Go 1.18+ but simpler mental model)
- Clear separation: one registry per domain type
- Easy to test and mock

**Negative:**
- Runtime panics possible if types are mixed up (mitigated by tests)
- Slightly more boilerplate for each new type
- No compile-time type checking across registry boundaries

**Neutral:**
- Code duplication minimal due to shared Registry implementation

---

## ADR-002: Source Decoupling

### Context

The system needs to support multiple event sources:
- `fsnotify` for filesystem watching
- HTTP webhooks for external triggers
- Test sources for deterministic testing

The challenge was designing a boundary that allows swappable implementations without leaking implementation details.

### Decision

Use a `Source` interface with a receive-only channel pattern:

```go
type Source interface {
    Events() <-chan SourceEvent
    Start() error
    Stop() error
}

type SourceEvent struct {
    Type      EventType
    Path      string
    Timestamp time.Time
}
```

### Consequences

**Positive:**
- Clean producer/consumer boundary - consumers can only read, not write
- Testable: test sources can feed deterministic events
- Swappable backends without changing consumer code
- Natural backpressure via channel buffering
- No exposed internals (file descriptors, HTTP handlers, etc.)

**Negative:**
- Channel semantics must be well-documented (closed vs nil channels)
- Error handling must be explicit (separate from event stream)
- Potential for goroutine leaks if Stop() not called

**Neutral:**
- Requires understanding of Go channel patterns

---

## ADR-003: Clone-Under-Lock Pattern

### Context

Registries need to provide consistent snapshots for iteration without:
- Holding locks during long-running operations
- Risking deadlocks from callbacks under lock
- Exposing internal data structures

### Decision

Use `CloneUnderLock` pattern: acquire lock, copy data to new map, release lock, then operate on copy:

```go
func (r *Registry) CloneUnderLock() map[string]interface{} {
    r.mu.RLock()
    defer r.mu.RUnlock()

    clone := make(map[string]interface{}, len(r.items))
    for k, v := range r.items {
        clone[k] = v
    }
    return clone
}

// Usage: iterate without holding lock
items := registry.CloneUnderLock()
for id, obj := range items {
    process(obj) // May take arbitrary time
}
```

### Consequences

**Positive:**
- Minimal lock contention - lock held only for map copy
- Consistent iteration snapshot - no concurrent modification issues
- No deadlocks from callbacks trying to acquire same lock
- Predictable performance characteristics

**Negative:**
- Memory overhead for large registries (temporary copy)
- Stale data possible (snapshot may not reflect latest state)
- Copy cost O(n) regardless of iteration needs

**Neutral:**
- Trade-off favors read-heavy workloads with occasional large scans

---

## ADR-004: Sequential vs Parallel Bootstrap

### Context

Core services have implicit dependencies:
- Security must initialize before Communication (TLS setup)
- Storage must initialize before Controllers (data access)
- Random source needed for cryptographic operations

The question was whether to use parallel regions (statecharts) or sequential compound states.

### Decision

Use **Sequential Compound States** for bootstrap, not parallel regions:

```yaml
# Bootstrap chart - sequential phases
bootstrap:
  type: compound
  states:
    - security      # Phase 1: TLS, certs, keys
    - storage       # Phase 2: DB, caches
    - communication # Phase 3: HTTP server, listeners
    - controllers   # Phase 4: Business logic
```

### Consequences

**Positive:**
- Deterministic startup order - no race conditions
- Explicit dependency chain visible in chart structure
- Simpler state machine (no parallel region synchronization)
- Easier to debug startup failures

**Negative:**
- Slower startup (phases execute sequentially)
- Cannot exploit parallel initialization of independent services
- Adding new services requires editing sequence

**Neutral:**
- Startup time acceptable for long-running daemon
- Future optimization possible via sub-state parallel regions within phases

---

## ADR-005: Hard-Coded Bootstrap Chart

### Context

The system needs a deterministic startup sequence. Options included:
- Loading bootstrap chart from YAML file on disk
- Generating bootstrap chart programmatically
- Compiling bootstrap chart as constant

### Decision

Use a **hard-coded BootstrapChartYAML constant** compiled into the binary:

```go
const BootstrapChartYAML = `
apiVersion: maelstrom.v1
kind: BootstrapChart
spec:
  phases:
    - name: security
      states:
        - RandomSource
        - CertificateManager
    - name: storage
      states:
        - Database
        - Cache
    # ... etc
`
```

### Consequences

**Positive:**
- Zero external dependencies for boot (no file I/O during early startup)
- Tamper-proof core - cannot modify bootstrap via file edits
- Deterministic behavior across all deployments
- Faster startup (no YAML parsing from disk)
- Self-contained binary - single artifact deployment

**Negative:**
- Recompilation required to change bootstrap sequence
- Cannot customize bootstrap for different environments
- Binary size slightly larger (minimal, YAML is small)

**Neutral:**
- Extension point: post-bootstrap charts can load from files
- Override possible via build tags for specialized builds

---

## Status

| ADR | Status | Date |
|-----|--------|------|
| ADR-001 | Accepted | 2026-02-28 |
| ADR-002 | Accepted | 2026-02-28 |
| ADR-003 | Accepted | 2026-02-28 |
| ADR-004 | Accepted | 2026-02-28 |
| ADR-005 | Accepted | 2026-02-28 |

---

## References

- `/home/albert/git/maelstrom-v4/docs/arch-v1.md` - Architecture specification
- `/home/albert/git/maelstrom-v4/CLAUDE.md` - Development workflow
