# Auditor Agent Template

## Mission
Audit phase plans for template compliance and spec accuracy.

## ⚠️ CRITICAL RULES (Read First)

### NO webfetch - EVER
- **All files are local** in `/home/albert/git/maelstrom`
- **DO NOT attempt to fetch URLs** - this causes tool abortion
- Use `Read` tool or `lumora_lumora_read_file` for ALL file access
- Use `grep` via Bash for searching large files

### Split Large Tasks
- **Max 3-5 tests per audit**: If phase has 5+ tests, split into multiple audit agents
- **Example**: P4.5 has 5 tests → spawn 2 agents (2 tests + 3 tests)
- **Each agent creates separate audit file**: `P4.5-audit-part1.md`, `P4.5-audit-part2.md`
- **Orchestrator merges** into final `P4.5-audit.md`

## Pre-Flight: Find Spec References FIRST

Before reading any files, run grep to locate exact line numbers:

```bash
# Find all references to key terms in arch-v1.md
grep -n "TERM1\|TERM2\|TERM3" /home/albert/git/maelstrom/docs/arch-v1.md

# Example for P2.6:
grep -n "alwaysTaintAs" /home/albert/git/maelstrom/docs/arch-v1.md
# Result: L751, L1092
```

**Why this matters:**
- Avoids reading 2456-line docs 30 lines at a time
- Gives exact line numbers for verification
- Provides context for what to look for

## Files to Read (After Grep)

1. **Phase plan to audit**: `/home/albert/git/maelstrom/docs/layer-4/plans/P[X.Y]-*.md`
2. **Template reference**: `/home/albert/git/maelstrom/docs/layer-4/plans/P1.1-AttachTaint.md`
3. **Spec verification**: `/home/albert/git/maelstrom/docs/arch-v1.md` (only lines identified by grep)

## Audit Checklist

| Criteria | What to Check |
|----------|---------------|
| **Template structure** | All sections from P1.1 present: Phase ID, Title, Parent, Status, Parent Requirements, Dependencies, Satisfied Lower-Layer Requirements, Acceptance Criteria, Test Descriptions, Implementation Plan, Commit Plan, Deliverables |
| **Line numbers** | Verify EACH line number from grep matches actual content in arch-v1.md |
| **Test naming** | Format: `Test[Component]_[Behavior]_[ExpectedResult]` |
| **Acceptance criteria** | Each test has Given/When/Then/Expected Result/Spec Reference |
| **Dependencies** | Listed with correct phase IDs, no circular deps |
| **Security implications** | Taint propagation, boundary enforcement, violation handling addressed |

## Deliverable Format

Create `/home/albert/git/maelstrom/docs/layer-4/audits/P[X.Y]-audit.md`:

```markdown
# Audit Report: P[X.Y]-[Component]

## Compliance Status: [PASS/FAIL/ISSUES]

### Checklist Results

| Criteria | Status | Details |
|----------|--------|---------|
| Template structure | ✅ PASS | All sections present |
| Line numbers verified | ✅ PASS | L751, L1092 confirmed |
| Test naming | ✅ PASS | Both tests follow format |
| Acceptance criteria | ✅ PASS | Complete Given/When/Then |
| Dependencies | ✅ PASS | P2.1, P1 identified |
| Security implications | ✅ PASS | Override behavior documented |

### Issues Found

[None / List with specific fixes]

### Test-to-Spec Mapping

| Test | Maps To | Status |
|------|---------|--------|
| Test1 | arch-v1.md L751 | ✅ |
| Test2 | arch-v1.md L1092 | ✅ |

### Recommendation

**[PROCEED TO IMPLEMENTATION / FIX FIRST]**

[Specific actions if fixes needed]
```

## Common Issues to Catch

1. **Missing setup requirements**: Runtime tests need mocks/stubs documented
2. **Circular dependencies**: P6.4 → P5.2 → P6.4 is invalid
3. **Wrong line numbers**: L764-767 vs L776-779 for same concept
4. **Unclear test scope**: "Guard detects" vs "Event emission mechanism"
5. **Missing test for acceptance criterion**: e.g., `taintMode=none` not tested

## Time Budget

- **Grep phase**: 1-2 minutes
- **Read phase**: 2-3 minutes (3 files, ~500 lines total)
- **Audit phase**: 3-5 minutes
- **Total**: <10 minutes per audit

## Agent Spawning Strategy

### For Large Docs (arch-v1.md, spec-extraction.md)
1. **Spawn search agent first**: "Find all references to X in arch-v1.md using grep"
2. **Pass results to audit agent**: Include grep output in prompt
3. **Audit agent reads only**: Phase plan + template + specific lines from grep

This avoids 30-line iterative reads and completes in 1-2 tool calls.

### For Large Phase Plans (5+ tests)
1. **Split by test count**: 
   - Agent 1: Tests 1-2 (or 1-3)
   - Agent 2: Tests 3-5 (or 4-5)
2. **Each agent audits subset**: Check template compliance for assigned tests only
3. **Orchestrator merges**: Combine findings into single audit report

**Example for P4.5 (5 tests):**
```
Agent 1: Audit P4.5 tests 1-2 (Message, tool_result sources)
Agent 2: Audit P4.5 tests 3-5 (static, template, dynamic sources)
Orchestrator: Merge P4.5-audit-part1.md + P4.5-audit-part2.md → P4.5-audit.md
```

### For Complex Research Tasks (Spawn Sub-Agents)
**If you need to read multiple files or do complex research:**
1. **Spawn sub-agents for each task**:
   - Sub-agent 1: Read template file
   - Sub-agent 2: Read spec sections
   - Sub-agent 3: Read related phase plans
2. **Wait for all sub-agents to complete**
3. **Merge results** and perform audit

**Why this works:**
- Each sub-agent has simple, focused task (1-2 file reads)
- Avoids tool abortion from too many sequential reads
- General agents can coordinate multiple explore agents

**Example prompt for coordinator:**
```
Spawn 3 sub-agents:
Agent 1: Read /path/to/template.md, return structure
Agent 2: Read /path/to/spec.md lines X-Y, return key points
Agent 3: Read /path/to/related.md, return patterns
After all complete: Merge and create audit report
```

## Time Budget

- **Grep phase**: 1-2 minutes
- **Read phase**: 2-3 minutes (3 files, ~500 lines total)
- **Audit phase**: 3-5 minutes per 2-3 tests
- **Total**: <10 minutes per 2-3 tests audited