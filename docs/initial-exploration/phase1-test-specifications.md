# Phase 1 Test Specifications

**Date**: 2026-02-28
**Scope**: Source, Registry, Service, Hydration, Integration
**Total Tests**: 24

---

## Source Tests (pkg/source/)

### TestFileSystemSource_EmitsCreated
**Behavior**: When a YAML file is created in the watched directory, FileSystemSource emits a Created event.
**Setup**: Create FileSystemSource with temp directory, start watching.
**Action**: Write new file `test.yaml` to directory.
**Assert**: Receive SourceEvent{Type: Created, Key: "test.yaml"} within 500ms.

### TestFileSystemSource_EmitsUpdated
**Behavior**: When a YAML file is modified, FileSystemSource emits an Updated event.
**Setup**: Create source, start watching, create initial file.
**Action**: Modify file content.
**Assert**: Receive SourceEvent{Type: Updated, Key: "test.yaml"} within 500ms.

### TestFileSystemSource_EmitsDeleted
**Behavior**: When a YAML file is removed, FileSystemSource emits a Deleted event.
**Setup**: Create source, start watching, create file.
**Action**: Delete file.
**Assert**: Receive SourceEvent{Type: Deleted, Key: "test.yaml"} within 500ms.

### TestFileSystemSource_Debounces
**Behavior**: Rapid changes to same file coalesce into single event.
**Setup**: Create source with 100ms debounce.
**Action**: Write file 5 times in 50ms.
**Assert**: Receive exactly 1 SourceEvent (not 5) within 200ms.

### TestFileSystemSource_GracefulShutdown
**Behavior**: Stop() closes channel and returns Err() after Run() exits.
**Setup**: Create source, start Run() in goroutine.
**Action**: Call Stop().
**Assert**: Events channel closed, Err() returns nil (clean shutdown).

### TestFileSystemSource_ErrAfterClose
**Behavior**: Err() returns error if Run() exits abnormally.
**Setup**: Create source, delete watched directory during Run().
**Action**: Wait for Run() to exit.
**Assert**: Err() returns non-nil error.

### TestFileSystemSource_FiltersYAML
**Behavior**: Only .yaml files emit events (not .txt, .json).
**Setup**: Create source, start watching.
**Action**: Create `test.txt` and `test.yaml`.
**Assert**: Only receive event for `test.yaml`.

---

## Registry Tests (pkg/registry/)

### TestRegistry_SetGet
**Behavior**: Set stores value, Get retrieves latest.
**Setup**: Create Registry with mock hydrator.
**Action**: registry.Set("chart", []byte("yaml")).
**Assert**: registry.Get("chart") returns hydrated value, no error.

### TestRegistry_GetNotFound
**Behavior**: Get returns error for unknown key.
**Setup**: Create empty Registry.
**Action**: registry.Get("missing").
**Assert**: Returns error (KeyNotFound).

### TestRegistry_VersionTracking
**Behavior**: Multiple Set calls store versions, GetVersion retrieves specific.
**Setup**: Create Registry.
**Action**: Set same key 3 times with different content.
**Assert**: ListVersions returns 3 items, GetVersion returns correct version.

### TestRegistry_GetVersionNotFound
**Behavior**: GetVersion returns error for unknown version.
**Setup**: Create Registry, Set once.
**Action**: GetVersion("chart", "nonexistent-version").
**Assert**: Returns error (VersionNotFound).

### TestRegistry_CloneUnderLock
**Behavior**: Hooks execute without holding mutex (concurrent safe).
**Setup**: Registry with slow pre-hook (100ms sleep).
**Action**: Call Set from 2 goroutines concurrently.
**Assert**: No deadlock, both complete within 150ms each.

### TestRegistry_PreLoadHooks
**Behavior**: Pre-hooks transform value before storage.
**Setup**: Registry with pre-hook that appends "-transformed" to string value.
**Action**: Set("chart", []byte("test")).
**Assert**: Get returns "test-transformed".

### TestRegistry_PreLoadHookError
**Behavior**: Pre-hook error aborts Set, nothing stored.
**Setup**: Registry with pre-hook returning error.
**Action**: Set("chart", []byte("test")).
**Assert**: Set returns error, Get returns KeyNotFound.

### TestRegistry_PostLoadHooks
**Behavior**: Post-hooks observe stored value.
**Setup**: Registry with post-hook capturing key/value.
**Action**: Set("chart", []byte("test")).
**Assert**: Post-hook received "chart" and hydrated "test".

### TestRegistry_PostLoadHookMultiple
**Behavior**: Multiple post-hooks all called.
**Setup**: Registry with 3 post-hooks, each appending to slice.
**Action**: Set("chart", []byte("test")).
**Assert**: All 3 hooks executed (slice length 3).

---

## Service Tests (pkg/registry/)

### TestService_ProcessesEvents
**Behavior**: Service reads Source events and calls Registry.Set.
**Setup**: MockSource with buffered channel, Registry with mock.
**Action**: Send SourceEvent{Created, "chart", []byte("yaml")} to source.
**Assert**: Registry.Set called with "chart" and []byte("yaml").

### TestService_ObserverNotifications
**Behavior**: Service notifies observers via OnChange.
**Setup**: Service with Registry, add observer capturing events.
**Action**: Source emits Created event.
**Assert**: Observer called with key and hydrated value.

### TestService_ContextCancellation
**Behavior**: Run exits when context cancelled.
**Setup**: Service running with context.
**Action**: Cancel context.
**Assert**: Run returns ctx.Err() within 100ms.

### TestService_SourceErrorHandling
**Behavior**: Source error stops Service, returns error.
**Setup**: MockSource that closes channel with error.
**Action**: Start Run(), wait for exit.
**Assert**: Run returns error from source.Err().

### TestService_MultipleEvents
**Behavior**: Sequential events all processed.
**Setup**: Service running.
**Action**: Emit 10 events rapidly.
**Assert**: All 10 processed, Registry has 10 entries.

---

## Hydration Tests (pkg/chart/)

### TestHydrateChart_SimpleYAML
**Behavior**: Valid YAML unmarshals to ChartDefinition.
**Setup**: Valid chart YAML with id, version, states.
**Action**: HydrateChart(yaml).
**Assert**: Returns ChartDefinition with correct fields, no error.

### TestHydrateChart_EnvSubstitution
**Behavior**: ${ENV_VAR} replaced with environment value.
**Setup**: os.Setenv("TEST_VAR", "test-value"), YAML with id: ${TEST_VAR}.
**Action**: HydrateChart(yaml).
**Assert**: ChartDefinition.ID == "test-value".

### TestHydrateChart_MissingEnvVar
**Behavior**: Missing env var returns error.
**Setup**: YAML with ${MISSING_VAR}, env var not set.
**Action**: HydrateChart(yaml).
**Assert**: Returns error mentioning MISSING_VAR.

### TestHydrateChart_TemplateExecution
**Behavior**: {{ .AppVars.key }} replaced with value.
**Setup**: Hydrator with AppVars{"key": "value"}, YAML using template.
**Action**: HydrateChart(yaml).
**Assert**: Template substituted correctly.

### TestHydrateChart_TemplateSyntaxError
**Behavior**: Invalid template syntax returns error.
**Setup**: YAML with malformed {{ .BadSyntax }.
**Action**: HydrateChart(yaml).
**Assert**: Returns template parse error.

### TestHydrateChart_InvalidYAML
**Behavior**: Invalid YAML returns error.
**Setup**: Malformed YAML (unclosed brace).
**Action**: HydrateChart(yaml).
**Assert**: Returns YAML unmarshal error.

### TestHydrateChart_ValidationError
**Behavior**: Missing required fields returns error.
**Setup**: YAML without required "id" field.
**Action**: HydrateChart(yaml).
**Assert**: Returns validation error.

---

## Integration Tests (pkg/chart/)

### TestChartRegistry_LoadsFromDirectory
**Behavior**: NewChartRegistry loads all YAML files from directory.
**Setup**: Temp dir with 3 chart YAML files.
**Action**: NewChartRegistry(dir), Start(ctx).
**Assert**: All 3 charts loadable via Get().

### TestChartRegistry_HotReload
**Behavior**: File change triggers reload via OnChange.
**Setup**: ChartRegistry running, OnChange registered.
**Action**: Modify file in watched directory.
**Assert**: OnChange callback invoked with updated definition.

### TestChartRegistry_TypeAssertion
**Behavior**: Get returns ChartDefinition type.
**Setup**: ChartRegistry with valid chart.
**Action**: reg.Get("chart").
**Assert**: No panic on type assertion, returns ChartDefinition.

### TestChartRegistry_StartStop
**Behavior**: Start blocks, Stop unblocks.
**Setup**: ChartRegistry created.
**Action**: Start in goroutine, call Stop.
**Assert**: Start returns after Stop, no goroutine leak.

### TestChartRegistry_VersionHistory
**Behavior**: Multiple modifications tracked.
**Setup**: ChartRegistry, create file, modify twice.
**Action**: ListVersions.
**Assert**: 3 versions present, GetVersion retrieves each.

---

## Test Execution Order

TDD workflow follows dependency order:

1. **Source Tests** (7) - Foundation, no dependencies
2. **Registry Tests** (9) - Uses hydrator mock
3. **Service Tests** (5) - Uses Source + Registry
4. **Hydration Tests** (7) - Standalone, uses real YAML
5. **Integration Tests** (5) - Full stack

**Total**: 24 tests

Each test:
- Written first (RED)
- Minimal implementation (GREEN)
- Commit
- Next test

No test written before previous passes.
