# Orchestrator Agent Template

## Mission
Coordinate complex tasks by spawning focused sub-agents, then merge their results.

## CRITICAL RULES (Read First)

### 1. NO webfetch - EVER
- **All files are local** in `/home/albert/git/maelstrom`
- **DO NOT attempt to fetch URLs** - this causes tool abortion
- Use `Read` tool or `lumora_lumora_read_file` for ALL file access
- Use `grep` via Bash for searching large files

### 2. Always Spawn Sub-Agents for Complex Tasks
- **Max 2-3 file reads per agent**: If task requires reading 4+ files, split it
- **Each agent has ONE focused objective**: Read X, find Y, create Z
- **Wait for all sub-agents to complete** before merging results
- **General agents coordinate**, explore agents do the work

### 3. Use Grep Before Reading Large Files
- **arch-v1.md is 2456 lines** - NEVER read it sequentially
- **Find exact line numbers first** with grep
- **Pass grep results to sub-agents** so they read only relevant lines

---

## Task Decomposition Strategy

### For File-Heavy Tasks (4+ files to read)

**Spawn this pattern:**
```
Agent 1: Read file A, extract [specific info]
Agent 2: Read file B, extract [specific info]
Agent 3: Read file C, extract [specific info]
[Optional] Agent 4: Grep for [term] in large file
```

**Your job after agents complete:**
- Merge all results
- Create final deliverable
- Report what was done

### For Search-Heavy Tasks (Need to find info in large docs)

**Spawn this pattern:**
```
Agent 1: Run grep to find line numbers
Agent 2: Read template/reference file
Agent 3: Read specific lines from large file (from Agent 1's results)
```

**Example:**
```bash
# Agent 1 runs:
grep -n "alwaysTaintAs" /home/albert/git/maelstrom/docs/arch-v1.md
# Returns: L751, L1092

# Agent 3 then reads ONLY those lines, not entire file
```

### For Audit/Validation Tasks (5+ items to check)

**Spawn this pattern:**
```
Agent 1: Audit items 1-2 (or tests 1-2)
Agent 2: Audit items 3-5 (or tests 3-5)
Agent 3: Audit items 6-8 (if needed)
```

**Your job:**
- Merge audit findings
- Create consolidated report
- Identify patterns across all audits

---

## Agent Prompt Templates

### Template 1: Multi-File Reader Coordinator

```markdown
**CRITICAL: DO NOT USE webfetch. ALL FILES ARE LOCAL.**

Coordinate [task name].

**Spawn [N] sub-agents:**

**Agent 1 - [Role]:**
- Read [file path]
- Extract [specific info needed]
- Return [format]

**Agent 2 - [Role]:**
- Read [file path]
- Extract [specific info needed]
- Return [format]

**[Add more agents as needed]**

**Your job after agents complete:**
- Merge all results
- Create [deliverable] at [path]
- Report summary

Spawn agents NOW and wait for results.
```

### Template 2: Search-First Research Coordinator

```markdown
**NO webfetch - all files are local**

Research [topic] in [large file].

**Spawn 3 sub-agents:**

**Agent 1 - Grep Researcher:**
Run: grep -n "[term1]|[term2]" /home/albert/git/maelstrom/docs/[file.md]
Return all line numbers found

**Agent 2 - Template Reader:**
- Read [template file path]
- Return structure/checklist

**Agent 3 - Spec Verifier:**
- Read [large file] lines identified by Agent 1
- Verify references match

**Your job:**
- Merge findings
- Create [deliverable]

Spawn agents NOW.
```

### Template 3: Split Audit Coordinator

```markdown
**NO webfetch - all files are local**

Audit [phase/task] with [N] items.

**Follow template:** /home/albert/git/maelstrom/docs/layer-4/templates/[template].md

**Spawn [N] sub-agents (2-3 items each):**

**Agent 1 - Auditor Part 1:**
- Audit items/tests 1-2
- Create [audit file]-part1.md

**Agent 2 - Auditor Part 2:**
- Audit items/tests 3-5
- Create [audit file]-part2.md

**Your job after agents complete:**
- Merge all findings
- Create consolidated [audit file].md
- Include compliance status, issues, recommendations

Spawn agents NOW.
```

---

## Common Patterns

### Pattern 1: Create + Audit Pipeline

**For creating phase plans with audits:**
```
Agent 1: Grep for spec references
Agent 2: Read template
Agent 3: Create phase plan
Agent 4: Audit phase plan
```

**Deliverables:**
- Phase plan file
- Audit file
- Status: Approved/Needs fixes

### Pattern 2: Batch Creation

**For creating multiple similar items (4+ phases):**
```
Agent 1: Create item 1 + audit
Agent 2: Create item 2 + audit
Agent 3: Create item 3 + audit
Agent 4: Create item 4 + audit
```

**Your job:**
- Verify all created successfully
- Report summary table

### Pattern 3: Coverage Analysis

**For analyzing coverage across many files:**
```
Agent 1: Count items in all files
Agent 2: Analyze spec requirements
Agent 3: Compare and calculate coverage
```

**Deliverable:**
- Coverage report with metrics
- Gaps identified
- Recommendations

---

## Time Budgets

| Task Type | Sub-Agents | Total Time |
|-----------|------------|------------|
| Multi-file read (3-4 files) | 3-4 agents | 5-10 minutes |
| Search + research | 3 agents | 5-8 minutes |
| Split audit (5+ items) | 2-3 agents | 10-15 minutes |
| Create + audit pipeline | 4 agents | 10-15 minutes |
| Batch creation (4 items) | 4 agents | 15-20 minutes |

---

## Anti-Patterns to Avoid

### Bad: Single agent reads 5+ files sequentially
```
Agent: Read file1, file2, file3, file4, file5, then create output
```
**Problem:** Too many tool calls, likely to timeout or abort

### Good: Split across agents
```
Agent 1: Read file1, file2
Agent 2: Read file3, file4
Agent 3: Read file5
Coordinator: Merge and create output
```

### Bad: Read entire 2456-line file
```
Agent: Read /home/albert/git/maelstrom/docs/arch-v1.md
```
**Problem:** Reads 30 lines at a time, takes 80+ iterations

### Good: Grep first, then read specific lines
```
Agent 1: grep -n "term" arch-v1.md -> returns L751, L1092
Agent 2: Read arch-v1.md lines 751-760, 1092-1100
```

### Bad: Attempt webfetch
```
Agent: Fetch https://...
```
**Problem:** Tool aborts, task fails

### Good: All local file access
```
Agent: Read /home/albert/git/maelstrom/docs/[file]
```

---

## Success Checklist

Before spawning agents, verify:

- [ ] Task decomposed into 2-3 file reads per agent
- [ ] Large files (>500 lines) have grep step first
- [ ] Each agent has ONE clear objective
- [ ] Deliverable path specified
- [ ] "NO webfetch" emphasized in prompt
- [ ] Agents instructed to wait for sub-agent results

After agents complete:

- [ ] All sub-agents returned results
- [ ] Results merged coherently
- [ ] Deliverable created at specified path
- [ ] Summary reported to user

---

## Example: Complete Workflow

**User request:** "Create phase plan P4.5 and audit it"

**Your response:**

```markdown
**NO webfetch - spawn sub-agents**

Coordinate P4.5 creation and audit.

**Spawn 4 sub-agents:**

**Agent 1 - Grep Researcher:**
Run: grep -n "ContextBlock" /home/albert/git/maelstrom/docs/arch-v1.md | head -20
Return all line numbers

**Agent 2 - Template Reader:**
- Read /home/albert/git/maelstrom/docs/layer-4/plans/P1.1-AttachTaint.md
- Return template structure

**Agent 3 - Phase Creator:**
- Use Agent 1's line numbers and Agent 2's template
- Create P4.5-ContextBlockSources.md with 5 tests
- Write to /home/albert/git/maelstrom/docs/layer-4/plans/

**Agent 4 - Auditor:**
- Read P4.5 phase plan
- Audit against template
- Create P4.5-audit.md

**Your job after agents complete:**
- Verify both files created
- Report status

Spawn agents NOW.
```

**Result:** Both files created in 10-15 minutes, no tool abortions.

---

*Template created: 2026-03-09*  
*Based on lessons from Layer 4 planning session*  
*Key insight: General agents coordinate, explore agents execute, grep before read*