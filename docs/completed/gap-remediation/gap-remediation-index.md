# Gap Remediation Index

**Generated**: 2026-03-09  
**Parent**: `docs/planning/gap-remediation-plan.md`

---

## Phase Tree

```
G1 (Critical Fixes) - P0
├── G1.1: ParseAddress Implementation (L3-H4) - 4 tests
├── G1.2: StreamSession Send/Close (L3-C1) - 4 tests
├── G1.3: Security Boundary Enforcement (L2-C1) - 2 tests
├── G1.4: NamespaceIsolate Method (L2-C2) - 2 tests
├── G1.5: CheckTaintPolicy Method (L2-C3) - 2 tests
└── G1.6: Human Gateway Chat Endpoint (L3-C3) - 2 tests
    └── Total: 16 tests, 16 commits

G2 (Core Functionality) - P1
├── G2.1: Address Validation Helpers (L3-H1) - 3 tests
├── G2.2: Taint Propagation (L2-H1) - 3 tests
├── G2.3: At-Least-Once Delivery (L2-H2) - 4 tests
├── G2.4: Request-Reply Pattern (L3-H2) - 4 tests
└── G2.5: Gateway Servers (L3-H3) - 4 tests
    └── Total: 18 tests, 18 commits

G3 (Integration & Wiring) - P1
├── G3.1: Service Registry State Tracking (L2-H5) - 3 tests
├── G3.2: Service Bootstrap (L2-M5) - 4 tests
├── G3.3: Dead-Letter Integration (L3-M3) - 4 tests
└── G3.4: Stream Taint Integration (L3-M4) - 3 tests
    └── Total: 14 tests, 14 commits

G4 (Gateway & External APIs) - P2
├── G4.1: TopicSubscriber Interface Fix (L3-M2) - 2 tests
├── G4.2: OpenAPI Generation (L3-C2) - 4 tests
├── G4.3: Hot-Reloadable Services (L2-M4) - 10 tests
└── G4.4: HTTP Endpoint Exposure (L3-C2) - 4 tests
    └── Total: 20 tests, 20 commits

G5 (Observability & Metrics) - P2
├── G5.1: Mail Metadata Type Fix (L3-M1) - 3 tests
├── G5.2: Metrics Collection (L2-H3) - 4 tests
├── G5.3: Dead-Letter Query Optimization (L2-M2) - 3 tests
└── G5.4: Runtime Tracking (L2-M3) - 3 tests
    └── Total: 13 tests, 13 commits

G6 (Hot-Reload & Advanced) - P3
├── G6.1: Deduplication (L2-M1) - 4 tests
└── G6.2: Hot-Reload (L2-H4) - 4 tests
    └── Total: 8 tests, 8 commits

Grand Total: 89 tests, 89 commits
```

---

## Execution Order

### P0 - Start Immediately (Phase G1)

| Order | Sub-Phase | Tests | Dependencies | Gap |
|-------|-----------|-------|--------------|-----|
| 1 | G1.1: ParseAddress | 4 | None | L3-H4 |
| 2 | G1.2: StreamSession | 4 | G1.1 | L3-C1 |
| 3 | G1.3: Security Boundary | 2 | None | L2-C1 |
| 4 | G1.4: NamespaceIsolate | 2 | G1.3 | L2-C2 |
| 5 | G1.5: CheckTaintPolicy | 2 | G1.3 | L2-C3 |
| 6 | G1.6: Human Gateway Chat | 2 | G1.1 | L3-C3 |

### P1 - After G1 (Phases G2, G3)

| Order | Sub-Phase | Tests | Dependencies | Gap |
|-------|-----------|-------|--------------|-----|
| 7 | G2.1: Address Validation | 3 | G1.1 | L3-H1 |
| 8 | G2.2: Taint Propagation | 3 | G1.3 | L2-H1 |
| 9 | G2.3: At-Least-Once | 4 | G1.1 | L2-H2 |
| 10 | G2.4: Request-Reply | 4 | G2.3 | L3-H2 |
| 11 | G2.5: Gateway Servers | 4 | G1.1 | L3-H3 |
| 12 | G3.1: Registry State | 3 | G1, G2 | L2-H5 |
| 13 | G3.2: Service Bootstrap | 4 | G3.1 | L2-M5 |
| 14 | G3.3: Dead-Letter | 4 | G2.3 | L3-M3 |
| 15 | G3.4: Stream Taints | 3 | G1.2, G1.3 | L3-M4 |

### P2 - After G1, G2 (Phases G4, G5)

| Order | Sub-Phase | Tests | Dependencies | Gap |
|-------|-----------|-------|--------------|-----|
| 16 | G5.1: Mail Metadata | 3 | None | L3-M1 |
| 17 | G5.2: Metrics Collection | 4 | G5.1 | L2-H3 |
| 18 | G5.3: Dead-Letter Opt | 3 | G3.3 | L2-M2 |
| 19 | G5.4: Runtime Tracking | 3 | G3.1 | L2-M3 |
| 20 | G4.1: TopicSubscriber | 2 | G1.1 | L3-M2 |
| 21 | G4.2: OpenAPI | 4 | G2.5 | L3-C2 |
| 22 | G4.3: Hot-Reloadable | 10 | G3.2 | L2-M4 |
| 23 | G4.4: HTTP Endpoints | 4 | G4.2, G2.5 | L3-C2 |

### P3 - After All Above (Phase G6)

| Order | Sub-Phase | Tests | Dependencies | Gap |
|-------|-----------|-------|--------------|-----|
| 24 | G6.1: Deduplication | 4 | G2.3 | L2-M1 |
| 25 | G6.2: Hot-Reload | 4 | G3.2, G5.4 | L2-H4 |

---

## Gap References

| Gap | Phase | Sub-Phase | Priority |
|-----|-------|-----------|----------|
| L2-C1 | G1 | G1.3 | P0 |
| L2-C2 | G1 | G1.4 | P0 |
| L2-C3 | G1 | G1.5 | P0 |
| L2-H1 | G2 | G2.2 | P1 |
| L2-H2 | G2 | G2.3 | P1 |
| L2-H3 | G5 | G5.2 | P2 |
| L2-H4 | G6 | G6.2 | P3 |
| L2-H5 | G3 | G3.1 | P1 |
| L2-M1 | G6 | G6.1 | P3 |
| L2-M2 | G5 | G5.3 | P2 |
| L2-M3 | G5 | G5.4 | P2 |
| L2-M4 | G4 | G4.3 | P2 |
| L2-M5 | G3 | G3.2 | P1 |
| L3-C1 | G1 | G1.2 | P0 |
| L3-C2 | G4 | G4.2, G4.4 | P2 |
| L3-C3 | G1 | G1.6 | P0 |
| L3-H1 | G2 | G2.1 | P1 |
| L3-H2 | G2 | G2.4 | P1 |
| L3-H3 | G2 | G2.5 | P1 |
| L3-H4 | G1 | G1.1 | P0 |
| L3-M1 | G5 | G5.1 | P2 |
| L3-M2 | G4 | G4.1 | P2 |
| L3-M3 | G3 | G3.3 | P1 |
| L3-M4 | G3 | G3.4 | P1 |

---

## Dependencies Graph

```
G1.1 (ParseAddress)
  └── G1.2, G1.6, G2.1, G2.3, G2.5, G4.1

G1.2 (StreamSession)
  └── G3.4

G1.3 (Security Boundary)
  └── G1.4, G1.5, G2.2, G3.4

G1.4 (NamespaceIsolate)
  └── (none)

G1.5 (CheckTaintPolicy)
  └── (none)

G1.6 (Human Gateway)
  └── (none)

G2.1 (Address Validation)
  └── (none)

G2.2 (Taint Propagation)
  └── (none)

G2.3 (At-Least-Once)
  └── G2.4, G3.3, G6.1

G2.4 (Request-Reply)
  └── (none)

G2.5 (Gateway Servers)
  └── G4.2, G4.4

G3.1 (Registry State)
  └── G3.2, G5.4

G3.2 (Service Bootstrap)
  └── G4.3, G6.2

G3.3 (Dead-Letter)
  └── G5.3

G3.4 (Stream Taints)
  └── (none)

G4.1 (TopicSubscriber)
  └── (none)

G4.2 (OpenAPI)
  └── G4.4

G4.3 (Hot-Reloadable)
  └── (none)

G4.4 (HTTP Endpoints)
  └── (none)

G5.1 (Mail Metadata)
  └── G5.2

G5.2 (Metrics)
  └── (none)

G5.3 (Dead-Letter Opt)
  └── (none)

G5.4 (Runtime Tracking)
  └── G6.2

G6.1 (Deduplication)
  └── (none)

G6.2 (Hot-Reload)
  └── (none)
```

---

## Summary

| Metric | Value |
|--------|-------|
| Total Sub-Phases | 25 |
| Total Tests | 89 |
| Total Commits | 89 |
| P0 Sub-Phases | 6 (G1.*) |
| P1 Sub-Phases | 9 (G2.*, G3.*) |
| P2 Sub-Phases | 8 (G4.*, G5.*) |
| P3 Sub-Phases | 2 (G6.*) |

---

**Index End**