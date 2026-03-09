# CLAUDE.md

## Development Rules

**MANDATORY GATE**

All new work is blocked until there exists:  

- Complete spec (user stories + acceptance criteria)  
- Frozen semantics & public APIs (types/signatures/docs)  
- Explicit, executable test descriptions (one test per behavior)

**STRICT TDD WORKFLOW** (never deviate)

1. `git checkout -b feat/xxx` (or `fix/xxx`)
2. Stub every public symbol (functions, classes, modules) — zero implementation, only signatures + `raise NotImplementedError` or `pass`
3. Write **one** test that exercises a single behavior from the spec
4. Run test suite → confirm it **fails** (RED)
5. Write the minimal code that makes **only that test** pass (GREEN)
6. `git commit -m "feat: <one-sentence summary>"` (or `fix:`)
7. Repeat 3–6 until every test described in step 0 is green
8. (Optional) Refactor while keeping all tests green
9. Open PR with 100% test coverage for changed paths

**RULES**

- Never write production code before a failing test  
- Never write more than one test at a time  
- Never commit failing tests or untested code  
- Stubs stay until their test is written  
- If spec/API changes, delete branch and start over

Follow this exactly. No exceptions.

## TOOLS

Prefer Lumora tools over equiivalent tools, when available.

## DEVELOPMENT STANDARDS

**Agent Spawning Protocol**

- Spawn explore agents for ALL discovery: specs, codebase, patterns
- Never read files yourself; delegate to agents exclusively
- Each agent has single focused objective with clear deliverable
- Agents quote line numbers and provide concrete evidence
- Recursive decomposition: agents spawn sub-agents for subtasks

**Task Decomposition Rules**

- Break work until: 2-10 tests per phase, 1 test = 1 commit
- Each phase: single coherent behavior, not multiple features
- Sub-phases address atomic requirements from spec
- Decomposition stops when phase has ≤6 tests ideal
- Recursive breakdown continues until testable units emerge

**Plan Document Standards**

- Quote arch-v1.md line numbers for ALL requirements
- Map each test to specific spec requirement explicitly
- Include acceptance criteria as executable test descriptions
- Reference satisfied requirements from Layers 2-3 dependencies
- Auditor agent validates all plans before implementation begins

**Style & Granularity**

- Follow existing patterns: gap remediation phase structure exactly
- Test names: `Test[Component]_[Behavior]_[ExpectedResult]` format
- Commit messages: `feat(layer-4/[component]): one-line description`
- Branch names: `feat/[group/category]-[component]` or `fix/[group/category]-[issue]`
- 1:1 test-to-commit ratio, never bundle multiple tests
