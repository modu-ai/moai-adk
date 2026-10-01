# progress.md — SPEC-INSTRUCTIONS-BUDGET-001

Card t1318 · worktree `.moai/worktrees/t1318-budget` · branch `WT-instr-budget`.
Run-phase record (manager-develop owns §E.2/§E.3). Full verbatim evidence:
`.moai/reports/t1318/run2-20261001.md` (primary checkout).

## §E.2 Run-phase Evidence

### Entry and transition

- Plan-artifact amendment SHA cited per the kickoff decision record: `2e88fc8ca` (D1 applied before dispatch; run-phase base).
- Status transition draft → in-progress: commit `888fa46fc` — `feat(SPEC-INSTRUCTIONS-BUDGET-001): M1 run-phase entry (draft -> in-progress)`, body `Card: t1318`, trailer `Authored-By-Agent: manager-develop`.

### Pre-change baseline reproduction (tree `2e88fc8ca`, before any edit)

- Command: `unset MOAI_KANBAN_ID MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_AUTONOMY_TIER MOAI_LAUNCH_PROVIDER CLAUDE_PROJECT_DIR && go test ./internal/hook/ -count=1` (one compound invocation).
- Observed: exit 1, 362.766s, exactly the 3 known pre-existing failures (t1350 env class, NOT this card's): TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration, TestMaybeDeclareGLMContextWindow, TestMaybeSet1MAutoCompactWindow. (A `^--- FAIL` grep shows 2 of 3 — the third name is glued to a preceding output line by parallel-test interleaving; `grep 'FAIL: Test'` finds all 3.)

### TDD RED (tests written first, tree `888fa46fc`)

- RED-now greps (pre-implementation): `grep -c sessionCharBudget internal/hook/instructions_loaded.go` → `0` exit 1; `grep -c aggregateInstructionChars ...` → `0` exit 1; `grep -c TestInstructionsLoadedAggregateBudget internal/hook/instructions_loaded_test.go` → `0` exit 1.
- RED run: `go test ./internal/hook/ -run '^TestInstructionFileSetDerivation$|^TestInstructionFileSetCycleSafe$|^TestAggregateInstructionCharsMovement$|^TestInstructionsLoadedAggregateBudget$' -count=1` → build failed, exit 1 — `undefined: instructionFileSet` / `undefined: aggregateInstructionChars` / `undefined: sessionCharBudget` (9 deciding lines).

### Implementation (commit `d21e4079e`)

- `internal/hook/instructions_loaded.go`: `const sessionCharBudget = 210000` (provenance comment + @MX:NOTE, line 142); `instructionFileSet` (mechanical derivation: CLAUDE.md anchor + transitive `^@` closure, repo-relative, cycle-safe, skip-missing + outside-root refusal; `.claude/rules/moai/` walk, always-loaded = no frontmatter OR no top-level `paths:`; malformed frontmatter → always-loaded; sorted, deduped); `parseImports`, `resolveUnder`, `ruleFileAlwaysLoaded`, `aggregateInstructionChars` (Σ utf8.RuneCount, skip unreadable); Handle integration advisory-only after the unchanged per-file path, empty set skips comparison, roots at `resolveProjectRoot` (t1160 precedent).
- `internal/hook/instructions_loaded_test.go`: regression suite — TestInstructionFileSetDerivation (S1+S2 semantics, growth/scoped/malformed/indented-paths mutants), TestInstructionFileSetCycleSafe, TestAggregateInstructionCharsMovement (exact-N movement), TestInstructionsLoadedAggregateBudget (over → advisory SystemMessage naming aggregate/budget/count with NO block decision; under → empty; exactly-210000 → no trip, strictly greater-than; per-file 40,000 unchanged).

### GREEN and verification

- Scoped selector post-implementation: `ok github.com/modu-ai/moai-adk/internal/hook 0.582s` (all four new tests pass).
- Green-side probes: `grep -c sessionCharBudget` → 4, literal `const sessionCharBudget = 210000` at line 142; `grep -c aggregateInstructionChars` → 3; `grep -c TestInstructionsLoadedAggregateBudget` → 2.
- Full package, same scrub as baseline: exit 1, 355.100s — same 3 pre-existing failure names, NO new failures; all new tests green inside the full run.
- `go vet ./internal/hook/`: clean, exit 0. `gofmt -l` on both touched files: clean.

### Arrival aggregate (AC-INSTRBUDGET-4, regression-guard)

- **files=18, chars=214,155** — measured WITH the implemented code (`aggregateInstructionChars` over `instructionFileSet`) against the implementation tree `d21e4079e`, via a temporary in-package test deleted immediately after the run (method re-creatable; the figure is a moving-window regression-guard per acceptance.md).
- vs operator ruling 210,000: over by 4,155. Independently corroborates the plan-phase figure (214,155 on 18 files). The advisory fires from the first InstructionsLoaded event — intended behavior (spec.md §B); the surface diet is out of card scope (spec.md §F).

### Gaps

- golangci-lint not run locally (plan §E.3 lists it; the CI-pinned lint verdict is CI's — scoped verification per dispatch was vet + gofmt).
- The trio's failure mode was compared by name+subtest signature, not byte-identical output.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-01
run_commit_sha: d21e4079e
baseline_reproduction: yes (3/3 names reproduced pre-change; unchanged post-change)
new_test_status: 4/4 green (scoped 0.582s + full package run)
arrival_aggregate: 214,155 chars / 18 files (over 210,000 ruling by 4,155)
