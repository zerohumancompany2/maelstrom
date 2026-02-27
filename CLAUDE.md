# CLAUDE.md

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
