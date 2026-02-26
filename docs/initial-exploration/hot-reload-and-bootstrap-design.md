# Hot-Reload

Discussion date: 2026-02-25
Status: Design decisions for implementation

---

## 1. Hot-Reload: Load-on-Next-Start Methodology

### Decision
Replace in-flight chart replacement with a clean "load on next start" approach using history mechanisms.

### Rationale
- Simplifies implementation significantly
- Avoids complex in-flight state migration
- Allows long-running charts to complete gracefully or stop elegantly
- Predictable behavior vs. uncertain runtime replacement

### Proposed Implementation

```yaml
migrationPolicy:
  onVersionChange: shallowHistory | deepHistory | cleanStart
  timeoutMs: 30000  # wait for quiescence before force-stop
  contextTransform: "optional_template_or_script"
```

### History Behavior
- **shallowHistory**: Restore to the parent state's default sub-state
- **deepHistory**: Restore to the specific sub-state (if it still exists)
- **Deleted state fallback**: If deep history targets a deleted state, fall back to shallow history

### Context Migration
Application context is schemaless by design. Migration handled via optional transform:

```yaml
contextTransform: |
  # Go template with access to old context
  {{if .oldContext.deprecatedKey}}
  newKey: {{.oldContext.deprecatedKey}}
  {{end}}
```

On transform failure → fall back to `cleanStart`.

### In-Flight Work Handling
Respect `stabilityPolicy`:
- If chart reaches quiescence before timeout → start new version with history
- If timeout expires with active work → force-stop, then clean start (no history)
