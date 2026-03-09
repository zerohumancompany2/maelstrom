# Phase G4.3: Hot-Reloadable Services

**Parent**: Phase G4 (Gateway & External APIs)  
**Gap References**: L2-M4  
**Status**: ❌ PENDING

## Overview
- Tests: 10
- Commits: 10 (1:1 ratio)
- Dependencies: G3.2 (Service Bootstrap)

## SPEC

### Requirements
From `arch-v1.md 7.2` - Hot-reloadable services:
- sys:admin - Execute admin commands with 2FA
- sys:persistence - Snapshot/restore
- sys:heartbeat - Schedule wake-ups
- sys:memory - Store/query
- sys:tools - Register/resolve tools

### Implementation Details

**Files to create/modify**:
- `pkg/services/admin/service.go` - Implement admin service
- `pkg/services/persistence/service.go` - Implement persistence service
- `pkg/services/heartbeat/service.go` - Implement heartbeat service
- `pkg/services/memory/service.go` - Implement memory service
- `pkg/services/tools/service.go` - Implement tools service
- `pkg/services/*_test.go` - Add tests

**Functions to implement**:
```go
// Admin
func (s *AdminService) ExecuteCommand(cmd string, token string) error

// Persistence
func (s *PersistenceService) Snapshot() error
func (s *PersistenceService) Restore(snapshotID string) error

// Heartbeat
func (s *HeartbeatService) Schedule(cron string) error
func (s *HeartbeatService) Trigger() error

// Memory
func (s *MemoryService) Store(key string, value interface{}) error
func (s *MemoryService) Query(key string) (interface{}, error)

// Tools
func (s *ToolsService) Register(name string, tool Tool) error
func (s *ToolsService) Resolve(name string) (Tool, error)
```

**Test scenarios**:
1. Execute admin command
2. 2FA required for commands
3. Take snapshot
4. Restore from snapshot
5. Schedule wake-ups
6. Trigger heartbeat
7. Store in memory
8. Query memory
9. Register tool
10. Resolve tool

## TDD Workflow

### Iteration 1: TestAdminService_AdminCommand
1. Write test: `TestAdminService_AdminCommand` in `pkg/services/admin/service_test.go`
2. Run → RED
3. Implement: Add ExecuteCommand method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement admin command execution"`

### Iteration 2: TestAdminService_2FAGate
1. Write test: `TestAdminService_2FAGate`
2. Run → RED
3. Implement: Add 2FA requirement
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement admin 2FA gate"`

### Iteration 3: TestPersistenceService_Snapshot
1. Write test: `TestPersistenceService_Snapshot` in `pkg/services/persistence/service_test.go`
2. Run → RED
3. Implement: Add Snapshot method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement persistence snapshot"`

### Iteration 4: TestPersistenceService_Restore
1. Write test: `TestPersistenceService_Restore`
2. Run → RED
3. Implement: Add Restore method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement persistence restore"`

### Iteration 5: TestHeartbeatService_Schedule
1. Write test: `TestHeartbeatService_Schedule` in `pkg/services/heartbeat/service_test.go`
2. Run → RED
3. Implement: Add Schedule method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement heartbeat scheduling"`

### Iteration 6: TestHeartbeatService_Trigger
1. Write test: `TestHeartbeatService_Trigger`
2. Run → RED
3. Implement: Add Trigger method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement heartbeat trigger"`

### Iteration 7: TestMemoryService_Store
1. Write test: `TestMemoryService_Store` in `pkg/services/memory/service_test.go`
2. Run → RED
3. Implement: Add Store method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement memory store"`

### Iteration 8: TestMemoryService_Query
1. Write test: `TestMemoryService_Query`
2. Run → RED
3. Implement: Add Query method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement memory query"`

### Iteration 9: TestToolsService_Register
1. Write test: `TestToolsService_Register` in `pkg/services/tools/service_test.go`
2. Run → RED
3. Implement: Add Register method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement tools registration"`

### Iteration 10: TestToolsService_Resolve
1. Write test: `TestToolsService_Resolve`
2. Run → RED
3. Implement: Add Resolve method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M4): implement tools resolution"`

## Deliverables
- 10 commits
- All tests passing
- Files modified: `pkg/services/admin/service.go`, `pkg/services/persistence/service.go`, `pkg/services/heartbeat/service.go`, `pkg/services/memory/service.go`, `pkg/services/tools/service.go`