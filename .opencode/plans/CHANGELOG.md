# Changelog Notes

## v1.4 - 2026-03-01

### Security Invariant Added

Added explicit security invariant to Section 4.2 (Data Tainting):

> **Security Invariant**: All data entering the runtime is tainted at the border. No untainted information exists inside the runtime. This is guaranteed by compile-time type checking: taints are attached as soon as data is ingested, touched, or known about by the application.

This invariant becomes the foundation for Phase 3 (Security Layer) implementation, where:
- TaintEngine marks data on ingestion (`Mark()`) and read (`MarkRead()`)
- DataSource system tags files with taints at write time
- ContextMap assembly filters blocks by boundary before LLM calls
- BoundaryService enforces inner/DMZ/outer separation

The invariant is enforced via compile-time type checking rather than runtime assertions, ensuring no untainted data can exist inside the runtime by design.

---

## v1.3 - 2025-02-28

### ChartRegistry API Finalized

- Added directory-partitioned sources (charts/, agents/, services/)
- Defined Source abstraction for event streaming
- Hot-reload protocol with quiescence detection
- Migration policies (shallowHistory, deepHistory, cleanStart)

---

## v1.2 - 2025-02-28

### Registry Infrastructure

- Added ChartRegistry for YAML loading and hydration
- Introduced Source interface for event streaming
- File watching with debounced change detection
- Versioned storage with hot-reload capability

---

## v1.1 - 2025-02-25

### Security Appendix

- Added boundary model (inner/DMZ/outer)
- Defined data tainting system
- ContextMap assembly with taint filtering
- Stream sanitization per-chunk

---

## v1.0 - 2025-02-24

### Initial Creation

- Statechart-native agentic runtime architecture
- Chart abstraction (state machines in YAML)
- Node-based graph (atomic/compound/parallel)
- Action/Guard registry system