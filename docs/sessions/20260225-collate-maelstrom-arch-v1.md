# Session: Collate Maelstrom Architecture v1

**Date:** 2026-02-25

## Summary

Reviewed conversation transcripts from `docs/initial-exploration/` and created a unified architectural specification at `docs/arch-v1.md`.

## Source Material

- `maelstrom-arch.md` - Original sketch
- `ARCH_CONVO.md` - Extended design refinement (primary source)
- `lit-review.md` - External vetting sources
- `self-improvement.md` - Meta-agent/self-improvement concepts

## Outputs

- `docs/arch-v1.md` (~2100 lines) - Complete technical specification including:
  - Core abstractions (Chart, Node, Events, Actions/Guards)
  - Data models (ChartDefinition, Message, Session, ContextBlock, AgentSpec)
  - Security model (inner/DMZ/outer boundaries, tainting)
  - Statechart Library vs Maelstrom App seam
  - Agent Charts and canonical LLM Inference Loop
  - Platform Services (sys:*)
  - Tool/Orchestrator patterns
  - Mail-based inter-agent communication
  - Streaming (SSE/WebSocket) via Gateway
  - DataSources with xattr tainting
  - Bootstrap sequence (hard-coded kernel → Bootstrap Chart → YAML services)
  - Complete YAML schemas
  - API contracts with SHALL/SHALL NOT test specs
  - Full OODA agent example

## Key Decisions Collated

1. **Unified Node primitive** - State type derived from child count (0=atomic, 1=compound, ≥2=parallel)
2. **Locked Library/App seam** - Pure statechart library knows nothing of Mail/LLM/taints; Maelstrom provides domain actions
3. **Proactive tainting** - Taints at Message creation, ContextMap assembly, and file read; enforcement at boundaries
4. **Hard-coded core** - Bootstrap kernel only; everything else YAML-defined and hot-reloadable
5. **Orchestrator as sub-chart** - Tool/sub-agent calls through configurable sub-chart with isolation
6. **Mail vs Events** - Mail for cross-boundary (async), Events for internal (sync where guaranteed)

## Open Items

- Implementation of pure statechart library interface
- ChartRegistry implementation (file watch, hydration, versioning)
- Security service implementation (taint propagation, boundary enforcement)
- Go interface definitions for all functional units
- Test suite YAML format

## Next Steps

1. Draft Go interfaces for locked APIs
2. Implement minimal statechart library
3. Build bootstrap kernel
4. Implement core Platform Services
