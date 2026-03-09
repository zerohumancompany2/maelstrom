# Implementer Agent Template

## Mission
Execute strict TDD workflow for Layer 4 implementation: 1 test → RED → GREEN → commit.

## ⚠️ CRITICAL RULES (Read First)

### 1. STRICT TDD WORKFLOW (Never Deviate)
- **ONE test at a time**: Write test → verify RED → implement → verify GREEN → commit
- **1:1 test-to-commit ratio**: Each test gets its own commit
- **NO production code before failing test**: Stubs with `raise NotImplementedError` or `pass` only
- **NO bundling multiple tests**: Never write 2+ tests in one iteration
- **NO committing failing tests or untested code**

### 2. NO webfetch - EVER
- **All files are local** in `/home/albert/git/maelstrom`
- **DO NOT attempt to fetch URLs** - this causes tool abortion
- Use `Read` tool or `lumora_lumora_read_file` for ALL file access

### 3. Branch & Commit Standards
- **Branch format**: `feat/layer4-[component]` or `fix/layer4-[issue]`
- **Commit format**: `feat(layer-4/[component]): one-line description`
- **Test naming**: `Test[Component]_[Behavior]_[ExpectedResult]`

---

## Pre-Flight Checklist

Before starting implementation:

1. **Identify phase to implement**: Read phase plan from `/home/albert/git/maelstrom/docs/layer-4/plans/`
2. **Verify audit status**: Check `/home/albert/git/maelstrom/docs/layer-4/audits/` for approval
3. **Confirm branch**: `feat/layer4-implementation` (or component-specific branch)
4. **Review dependencies**: Ensure prerequisite phases are implemented

---

## TDD Workflow (Repeat for Each Test)

### Step 1: Select Next Test
- Read phase plan
- Identify first unimplemented test (check what's already committed)
- **Deliverable**: Test number and name (e.g., "Test 1: TestTaintEngine_AttachTaint_AddsTaintToMessage")

### Step 2: Write Test Stub (RED Phase)
- Create test file in appropriate test directory
- Write ONE test that exercises the behavior from spec
- **DO NOT implement production code yet**
- **Deliverable**: Test file with failing test

### Step 3: Verify Test Fails (RED)
- Run test suite for that specific test
- **Expected**: Test FAILS (red)
- **If test passes**: You've made a mistake - remove test and restart
- **Deliverable**: Test failure output

### Step 4: Implement Minimal Code (GREEN Phase)
- Write ONLY the code needed to make THIS test pass
- Follow existing patterns from `/home/albert/git/maelstrom/docs/layer-4/implementation-patterns.md`
- Keep implementation minimal - no over-engineering
- **Deliverable**: Production code that makes test pass

### Step 5: Verify Test Passes (GREEN)
- Run test again
- **Expected**: Test PASSES (green)
- **If test fails**: Fix implementation, repeat until green
- **Deliverable**: Test success output

### Step 6: Commit (1:1 Ratio)
- Add test file and production code
- Commit with message: `feat(layer-4/[component]): [test behavior]`
- **Example**: `feat(layer-4/taint-engine): Attach taint to message on creation`
- **Deliverable**: Commit hash

### Step 7: Repeat
- Return to Step 1 for next test
- Continue until all tests in phase are green and committed

---

## Implementation Patterns

### Stubbing Public Symbols

**Before any test is written, stub all public symbols:**

```python
# taint_engine.py
class TaintEngine:
    def attach_taint(self, message, source, policy):
        raise NotImplementedError
    
    def propagate_taint(self, context):
        raise NotImplementedError
    
    def check_violation(self, boundary, data):
        raise NotImplementedError
```

**Or for TypeScript:**

```typescript
export class TaintEngine {
  attachTaint(message: Message, source: string, policy: TaintPolicy): void {
    throw new Error('Not implemented');
  }
  
  propagateTaint(context: Context): void {
    throw new Error('Not implemented');
  }
  
  checkViolation(boundary: Boundary, data: Data): ViolationResult {
    throw new Error('Not implemented');
  }
}
```

### Test Structure Template

**Python (pytest):**

```python
def TestTaintEngine_AttachTaint_AddsTaintToMessage():
    # Given
    engine = TaintEngine()
    message = Message(content="test")
    source = "user-input"
    policy = TaintPolicy.ALWAYS
    
    # When
    engine.attach_taint(message, source, policy)
    
    # Then
    assert message.taint is not None
    assert message.taint.source == source
    assert message.taint.policy == policy
```

**TypeScript (Jest):**

```typescript
describe('TaintEngine', () => {
  it('TestTaintEngine_AttachTaint_AddsTaintToMessage', () => {
    // Given
    const engine = new TaintEngine();
    const message = new Message('test');
    const source = 'user-input';
    const policy = TaintPolicy.ALWAYS;
    
    // When
    engine.attachTaint(message, source, policy);
    
    // Then
    expect(message.taint).toBeDefined();
    expect(message.taint?.source).toBe(source);
    expect(message.taint?.policy).toBe(policy);
  });
});
```

---

## Common Scenarios

### Scenario 1: Multiple Files Need Changes

**Pattern:**
```
Test requires:
- New class in file A
- Interface update in file B  
- Import in file C
```

**Approach:**
1. Write test (it will fail to compile/import)
2. Add stub to file A
3. Update interface in file B
4. Add import in file C
5. Run test - should now fail with "NotImplementedError" (good!)
6. Implement minimal code in file A
7. Run test - should pass
8. Commit all 3 files together

**Key**: All files needed for ONE test go in ONE commit.

### Scenario 2: Test Needs Mocks/Fixtures

**Pattern:**
```
Test requires:
- Mock boundary object
- Mock context with taint
- Mock event emitter
```

**Approach:**
1. Create mock/fixture files FIRST (as part of test setup)
2. Write test using mocks
3. Verify test fails (RED) - should fail on implementation, not setup
4. Implement production code
5. Verify test passes (GREEN)
6. Commit test + mocks + production code together

**Key**: Mocks/fixtures are test infrastructure, committed with test.

### Scenario 3: Refactoring Needed

**Pattern:**
```
After implementing test 3, you realize tests 1-2 could be cleaner
```

**Approach:**
1. **DO NOT refactor yet** - finish all tests first
2. After all tests green, create refactoring commit
3. Run full test suite to ensure nothing broke
4. Commit with message: `refactor(layer-4/[component]): improve [aspect]`

**Key**: Refactoring happens AFTER all tests green, in separate commit.

---

## File Location Patterns

### Test Files
```
/home/albert/git/maelstrom/
├── tests/
│   ├── layer-4/
│   │   ├── taint_engine/
│   │   │   ├── test_attach_taint.py
│   │   │   ├── test_propagate_taint.py
│   │   │   └── test_check_violation.py
│   │   ├── boundary_enforcement/
│   │   │   ├── test_guard.py
│   │   │   └── test_enforcer.py
│   │   └── ...
```

### Production Files
```
/home/albert/git/maelstrom/
├── src/
│   ├── maelstrom/
│   │   ├── security/
│   │   │   ├── taint_engine.py
│   │   │   ├── boundary.py
│   │   │   ├── guard.py
│   │   │   └── enforcer.py
```

**Note**: Adjust based on actual project structure. Read existing tests to match patterns.

---

## Running Tests

### Discover Test Command

**Check existing test structure:**

```bash
# Look for test configuration
ls /home/albert/git/maelstrom/ | grep -E "pytest|jest|test|spec"

# Check package.json or pyproject.toml
cat /home/albert/git/maelstrom/package.json | grep -A 5 "scripts"
# or
cat /home/albert/git/maelstrom/pyproject.toml | grep -A 10 "\[tool.pytest"
```

### Run Single Test

**Python:**
```bash
pytest tests/layer-4/taint_engine/test_attach_taint.py::TestTaintEngine_AttachTaint_AddsTaintToMessage -v
```

**TypeScript:**
```bash
npm test -- testAttachTaint --testNamePattern="TestTaintEngine_AttachTaint_AddsTaintToMessage"
```

### Run All Tests in File

**Python:**
```bash
pytest tests/layer-4/taint_engine/test_attach_taint.py -v
```

**TypeScript:**
```bash
npm test -- testAttachTaint
```

---

## Commit Message Examples

### Good Examples
```
feat(layer-4/taint-engine): Attach taint to message on creation
feat(layer-4/taint-engine): Propagate taint through context chain
feat(layer-4/boundary): Guard detects taint violation on boundary crossing
feat(layer-4/boundary): Enforcer blocks message with policy violation
feat(layer-4/integration): Wire taint engine to message bus
```

### Bad Examples
```
feat(layer-4): add tests  # Too vague
feat(layer-4/taint-engine): implement attach and propagate  # Multiple tests
tests for taint engine  # Wrong format
feat: add taint stuff  # Missing component
```

---

## Time Budgets

| Task | Time |
|------|------|
| Write single test (RED) | 2-5 minutes |
| Implement minimal code (GREEN) | 3-10 minutes |
| Commit and verify | 1-2 minutes |
| **Total per test** | **6-17 minutes** |

**For a phase with 3 tests**: ~30-60 minutes
**For a phase with 5 tests**: ~50-90 minutes

---

## Quality Gates

**Before committing each test:**

- [ ] Test follows naming convention: `Test[Component]_[Behavior]_[ExpectedResult]`
- [ ] Test has Given/When/Then structure
- [ ] Test was RED before implementation
- [ ] Test is GREEN after implementation
- [ ] Implementation follows existing code patterns
- [ ] Commit message matches format: `feat(layer-4/[component]): [behavior]`
- [ ] Only files for THIS test are committed (1:1 ratio)

**Before completing phase:**

- [ ] All tests from phase plan are implemented
- [ ] All tests pass (full test suite, not just individual)
- [ ] Code coverage is 100% for changed paths
- [ ] No linting/type checking errors
- [ ] Phase can be marked complete in tracking docs

---

## Anti-Patterns to Avoid

### Bad: Write All Tests First
```
1. Write all 5 tests for phase
2. All tests fail (expected)
3. Now implement all code
4. Commit everything together
```
**Problem**: Violates 1:1 ratio, hard to isolate issues, commits untested code

### Good: One Test at a Time
```
1. Write test 1 → RED
2. Implement code for test 1 → GREEN
3. Commit test 1 + code
4. Write test 2 → RED
5. Implement code for test 2 → GREEN
6. Commit test 2 + code
... repeat
```

### Bad: Over-Engineering
```
Test: Attach taint to message
Implementation: Full taint propagation system with caching, events, persistence
```
**Problem**: Implementing features not tested yet

### Good: Minimal Implementation
```
Test: Attach taint to message
Implementation: message.taint = Taint(source, policy)
```

### Bad: Skipping RED Phase
```
1. Write stub
2. Implement code immediately
3. Write test
4. Test passes
5. Commit
```
**Problem**: No verification that test actually tests anything

### Good: Verify RED
```
1. Write stub
2. Write test
3. Run test → FAILS (RED) ✓
4. Implement code
5. Run test → PASSES (GREEN) ✓
6. Commit
```

---

## When to Spawn Sub-Agents

### For Complex Research (Spawn explore agent)
```
Task: "What's the existing test pattern in this codebase?"
→ Spawn explore agent to find and read existing tests
```

### For Multi-File Changes (You can handle)
```
Test requires changes to 3-4 files
→ You can do this directly, no agent needed
```

### For Parallel Independent Tests (DON'T split)
```
Phase has 5 independent tests
→ DO NOT spawn 5 agents for 5 tests
→ You implement sequentially, 1 test at a time
```

**Key Insight**: TDD is inherently sequential. Don't parallelize the RED→GREEN→COMMIT cycle.

---

## Example: Complete Workflow

**Phase**: P1.1 - AttachTaint (3 tests)

### Iteration 1: Test 1

**You:**
```
Starting P1.1 Test 1: TestTaintEngine_AttachTaint_AddsTaintToMessage

Step 1: Writing test...
[Creates test file]

Step 2: Running test (expect RED)...
[pytest output: FAILED]

Step 3: Implementing minimal code...
[Updates taint_engine.py]

Step 4: Running test (expect GREEN)...
[pytest output: PASSED]

Step 5: Committing...
git add tests/layer-4/taint_engine/test_attach_taint.py src/maelstrom/security/taint_engine.py
git commit -m "feat(layer-4/taint-engine): Attach taint to message on creation"

Test 1 complete. Moving to Test 2.
```

### Iteration 2: Test 2
[Repeat process...]

### Iteration 3: Test 3
[Repeat process...]

**Final:**
```
P1.1 complete: 3 tests, 3 commits, 100% coverage
Ready for next phase.
```

---

## Success Metrics

**Per Test:**
- ✅ Test was RED before implementation
- ✅ Test is GREEN after implementation  
- ✅ Single commit with test + code
- ✅ Commit message follows format

**Per Phase:**
- ✅ All tests from plan implemented
- ✅ All tests pass in full suite
- ✅ 100% coverage on changed paths
- ✅ No lint/type errors
- ✅ Phase marked complete

**Per Sprint (Multiple Phases):**
- ✅ All commits follow 1:1 ratio
- ✅ No untested code committed
- ✅ Branch ready for PR with full test coverage

---

*Template created: 2026-03-09*  
*Based on CLAUDE.md Layer 4 development rules*  
*Key principle: 1 test → RED → GREEN → commit, repeat*