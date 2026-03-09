# Layer 1 Minor Gaps

**Created:** 2026-03-08  
**Source:** Layer 1 Audit Report  
**Priority:** Low (Non-blocking for Layer 2)

---

## Summary

The audit identified **3 minor gaps** that do not block Layer 2 development. These are documented here for future reference and implementation.

| Gap | Priority | Effort | Blocks Layer 2? |
|-----|----------|--------|-----------------|
| Error path tests | Low | 2 hours | No |
| ChartRegistry | Medium | 8 hours | No (Layer 2 component) |
| File watching | Medium | 6 hours | No (Layer 2+ component) |

---

## Gap 1: Error Path Tests Incomplete

### Description

Bootstrap chart has a `failed` state and error transitions defined, but no tests verify that error conditions actually trigger the failed state.

### Spec Reference

- `layer-01-kernel-bootstrap.md:L159-161` - Failed state specification
- `bootstrap/chart.go:L46-58` - Error transition implementation

### Current Implementation

```yaml
# bootstrap/chart.go - Error transitions exist
transitions:
  - from: security
    to: failed
    on: "*"  # Any error event
  - from: communication
    to: failed
    on: "*"
  - from: observability
    to: failed
    on: "*"
  - from: lifecycle
    to: failed
    on: "*"
```

### Missing Tests

Need 2-3 tests to verify:

1. **TestBootstrap_ErrorTransitionToFailed**
   - Simulate service spawn failure in `loadSecurityService`
   - Verify transition to `failed` state
   - Verify no further state transitions occur

2. **TestBootstrap_RecoveryFromFailed**
   - Verify `failed` state is terminal (no recovery in Phase 1)
   - Or document recovery mechanism for future phases

3. **TestKernel_BootstrapFailureHandling**
   - Verify kernel handles bootstrap failure gracefully
   - Verify error is returned to caller

### Implementation Plan

```go
// pkg/bootstrap/actions_test.go

func TestBootstrap_ErrorTransitionToFailed(t *testing.T) {
    engine := statechart.NewEngine()
    RegisterBootstrapActions(engine)
    
    // Mock engine to fail on spawn
    var failOnSpawn bool
    originalSpawn := engine.Spawn
    engine.Spawn = func(def, ctx) (RuntimeID, error) {
        if failOnSpawn {
            return "", errors.New("spawn failed")
        }
        return originalSpawn(def, ctx)
    }
    
    // Spawn bootstrap and trigger failure
    rtID, _ := engine.Spawn(bootstrapChart, mockCtx)
    engine.Control(rtID, CmdStart)
    
    // Verify failed state reached
    states := getStatesEntered(rtID)
    assertContains(t, states, "failed")
}
```

### Effort Estimate

- **Implementation:** 2 hours
- **Testing:** 1 hour
- **Total:** 3 hours

### Recommendation

**Defer to next sprint** after Layer 2 begins. Not critical for bootstrap to work, but important for production robustness.

---

## Gap 2: ChartRegistry Not Implemented

### Description

ChartRegistry component specified in Layer 1 but not implemented. This is intentionally deferred to Layer 2 as Layer 1 only needs the hard-coded bootstrap chart.

### Spec Reference

- `layer-01-kernel-bootstrap.md:L72-105` - ChartRegistry specification

### Required Functionality

```go
type ChartRegistry interface {
    // Store chart definition with version
    Store(chartID string, version string, def ChartDefinition) error
    
    // Retrieve latest or specific version
    Get(chartID string, version *string) (ChartDefinition, error)
    
    // List available versions
    ListVersions(chartID string) ([]string, error)
    
    // Watch for chart updates (hot-reload)
    Watch(chartID string) (<-chan ChartDefinition, error)
}
```

### Current Workaround

Bootstrap chart is hard-coded in `pkg/bootstrap/chart.go`:

```go
func LoadBootstrapChart() (ChartDefinition, error) {
    // Returns embedded YAML
    return parseBootstrapYAML()
}
```

### Implementation Plan (Layer 2)

1. **Create ChartRegistry interface** (`pkg/registry/registry.go`)
2. **Implement in-memory storage** (`pkg/registry/memory.go`)
3. **Implement file-based storage** (`pkg/registry/filesystem.go`)
4. **Add versioning support** (semver or timestamp-based)
5. **Add watch functionality** (fsnotify for hot-reload)

### Effort Estimate

- **Interface design:** 2 hours
- **In-memory implementation:** 3 hours
- **File-based implementation:** 4 hours
- **Versioning:** 2 hours
- **Watching:** 3 hours
- **Total:** 14 hours

### Recommendation

**Implement in Layer 2** as part of core services. This is a foundational component for chart management.

---

## Gap 3: File Watching Not Implemented

### Description

FileSystemSource and file watching with fsnotify not implemented. This enables hot-reload of chart definitions.

### Spec Reference

- `layer-01-kernel-bootstrap.md:L107-125` - File watching specification

### Required Functionality

```go
type FileSystemSource struct {
    watchDir string
    watcher  *fsnotify.Watcher
}

func (s *FileSystemSource) Load(chartID string) (ChartDefinition, error) {
    // Load from file: charts/{chartID}.yaml
    path := filepath.Join(s.watchDir, chartID+".yaml")
    return loadYAML(path)
}

func (s *FileSystemSource) Watch(chartID string) (<-chan ChartDefinition, error) {
    // Watch file for changes
    // Return channel that receives updated definitions
}
```

### Current Workaround

Charts are loaded from:
- Hard-coded YAML (bootstrap)
- Service bootstrap.go files (core services)

### Implementation Plan (Layer 2+)

1. **Create FileSystemSource** (`pkg/registry/filesystem.go`)
2. **Integrate fsnotify** for file watching
3. **Add debouncing** to prevent rapid reloads
4. **Add validation** before applying updates
5. **Integrate with ChartRegistry** for unified interface

### Effort Estimate

- **FileSystemSource:** 3 hours
- **fsnotify integration:** 2 hours
- **Debouncing:** 1 hour
- **Validation:** 2 hours
- **Integration:** 2 hours
- **Total:** 10 hours

### Recommendation

**Implement in Layer 3** as part of hot-reload feature. Not needed for initial bootstrap or Layer 2.

---

## Priority Matrix

| Gap | Impact | Effort | Priority | When to Implement |
|-----|--------|--------|----------|-------------------|
| Error path tests | Medium (robustness) | Low (3h) | **Low** | After Layer 2 starts |
| ChartRegistry | High (foundation) | Medium (14h) | **Medium** | Layer 2 |
| File watching | Medium (UX) | Medium (10h) | **Low** | Layer 3 |

---

## Implementation Timeline

### Sprint 1 (Current): Layer 1 ✅
- All critical bootstrap functionality complete
- 93 tests passing
- Ready for Layer 2

### Sprint 2: Layer 2 (Core Services)
- Implement full service functionality
- **Implement ChartRegistry** (Gap 2)
- Add error path tests (Gap 1) - optional

### Sprint 3: Layer 3 (Memory & Persistence)
- Implement memory service
- Implement persistence service
- **Implement file watching** (Gap 3) - optional

---

## Notes

### Why These Are Non-Blocking

1. **Error path tests**: Bootstrap works correctly in happy path. Error handling is defensive programming, not core functionality.

2. **ChartRegistry**: Layer 1 only needs bootstrap chart (hard-coded). Layer 2 services can also be hard-coded initially.

3. **File watching**: Hot-reload is a quality-of-life feature, not required for system to function.

### Risk Assessment

| Gap | Risk if Deferred | Mitigation |
|-----|-----------------|------------|
| Error path tests | Bugs in error handling undiscovered | Manual testing, production monitoring |
| ChartRegistry | Cannot load charts dynamically | Hard-code charts initially |
| File watching | No hot-reload capability | Restart kernel for updates |

---

## Action Items

### Immediate (Optional)

- [ ] Add error path tests (3 hours) - **Recommend: Defer**

### Layer 2 (Required)

- [ ] Implement ChartRegistry interface (2 hours)
- [ ] Implement in-memory ChartRegistry (3 hours)
- [ ] Implement file-based ChartRegistry (4 hours)
- [ ] Add versioning support (2 hours)

### Layer 3 (Optional)

- [ ] Implement fsnotify integration (2 hours)
- [ ] Add debouncing logic (1 hour)
- [ ] Integrate with hot-reload system (3 hours)

---

## Conclusion

All 3 gaps are **non-blocking** for Layer 2 development:

- Error path tests improve robustness but don't affect core functionality
- ChartRegistry is a Layer 2 component that can be implemented when needed
- File watching is a Layer 3+ feature for hot-reload

**Recommendation:** Proceed with Layer 2 development. Address gaps according to the timeline above.

---

**Document Status:** Ready for reference  
**Next Review:** Before Layer 2 planning