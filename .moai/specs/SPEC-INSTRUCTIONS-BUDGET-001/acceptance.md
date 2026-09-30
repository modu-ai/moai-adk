# acceptance.md — SPEC-INSTRUCTIONS-BUDGET-001

---
id: SPEC-INSTRUCTIONS-BUDGET-001
title: "Acceptance criteria — aggregate character budget for InstructionsLoaded"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
---

## §D — AC Matrix

All RED-now cells below were measured in this plan run on tree `2868afc0b` (worktree `.moai/worktrees/t1318-budget`, branch `WT-instr-budget`), before any implementation exists. Every command is a single read-only invocation — no pipes, no chaining — so the recorded exit code is the command's own (the t1360 pipe-masking hazard does not apply). Probes target unexported identifiers via grep because `go doc` cannot resolve unexported symbols in this package (measured: it reports "no symbol" for the existing `checkCharacterBudget`), and a not-yet-existing test selector is vacuously green (an unanchored `-run` selector exits 0 with `[no tests to run]`, measured).

| AC | Requirement | Criterion (release-blocking) | RED-now command | RED-now stdout (verbatim) | RED exit | Green path |
|----|-------------|------------------------------|-----------------|---------------------------|----------|------------|
| AC-INSTRBUDGET-1 | REQ-INSTRBUDGET-002 | `sessionCharBudget` defined as 210000 with the 2026-10-01 operator-ruling provenance comment | `grep -c sessionCharBudget internal/hook/instructions_loaded.go` | `0` | 1 | M2: grep count ≥ 1 and the literal `210000` present on the const line; `go doc`-free check via `go vet` clean |
| AC-INSTRBUDGET-2 | REQ-INSTRBUDGET-001, -003, -004 | `aggregateInstructionChars` exists, derives the set mechanically (D3: `^@` import closure + always-loaded rules without top-level `paths:`), and `Handle` emits the advisory SystemMessage on breach | `grep -c aggregateInstructionChars internal/hook/instructions_loaded.go` | `0` | 1 | M2: grep count ≥ 1; M2 integration verified by AC-INSTRBUDGET-3's tests |
| AC-INSTRBUDGET-3 | REQ-INSTRBUDGET-005 | Regression trio exists in `internal/hook/instructions_loaded_test.go` and the package suite is green | `grep -c TestInstructionsLoadedAggregateBudget internal/hook/instructions_loaded_test.go` | `0` | 1 | M3: `go test ./internal/hook/ -run '^TestInstructionsLoadedAggregateBudget$' -count=1` → `ok`; then `go test ./internal/hook/ -count=1` → `ok` |
| AC-INSTRBUDGET-4 | REQ-INSTRBUDGET-002 (**regression-guard**, not release-blocking) | The implemented metric's aggregate figure on the implementation tree is measured and recorded in the SPEC's progress.md (the arrival-state number feeds the diet-card decision) | `grep -c aggregate .moai/specs/SPEC-INSTRUCTIONS-BUDGET-001/progress.md` | (empty; stderr: `ugrep: warning: .moai/specs/SPEC-INSTRUCTIONS-BUDGET-001/progress.md: No such file or directory`) | 2 | plan.md §E.5: progress.md carries the figure measured with the implemented code; classified regression-guard because a future diet card legitimately moves the figure — this criterion can only be re-executed in a moving window |

**Why each RED cell is red for the right stated reason** (verification-completeness §2): all three release-blocking cells probe identifiers that no correct pre-implementation tree can contain — the symbols are the deliverable — so the red is neither vacuous (each command exits non-zero, no `[no tests to run]` shape) nor impossible (implementing M1-M3 flips each one) nor wrong-reason (the probed files exist on the tree; only the identifiers are absent). AC-INSTRBUDGET-4 is red because progress.md does not exist at plan close — plan phase ends before run phase writes it.

## §D.1 — Scenario detail (Given-When-Then)

**AC-INSTRBUDGET-3 scenario set (the mutant-catching trio):**

- **S1 derivation semantics.** Given a fixture project (`t.TempDir()`) with CLAUDE.md importing `AGENTS.md` (present) and `@missing.md` (dangling), a `rules/moai/` holding one always-loaded rule and one `paths:`-scoped rule — when `instructionFileSet` runs — the set contains CLAUDE.md, AGENTS.md, and the always-loaded rule; excludes the dangling import, the `paths:`-scoped rule, and non-md files; and is deterministically ordered.
- **S2 mechanical-derivation mutant check.** Given the same fixture, adding a new always-loaded rule file grows the set by one and the aggregate by that file's rune count; adding another rule WITH a top-level `paths:` key changes neither. A hardcoded-list mutant fails both assertions.
- **S3 aggregate movement.** Given any fixture set, appending N runes to one member increases `aggregateInstructionChars` by exactly N (catches wrong-unit and stale-count mutants).
- **S4 budget boundary.** Given a fixture whose aggregate exceeds 210,000, `Handle` returns `SystemMessage` naming the aggregate, the budget, and the file count; given a fixture under it, `Handle` returns an empty `HookOutput`. Per-file behavior on the same fixtures is unchanged (additive check — spec.md §E).

## §D.2 — Edge cases

- Missing/unreadable member mid-set: skipped silently-and-continue; aggregate = sum of readable members (REQ-INSTRBUDGET-003). No member → no comparison (empty derivation).
- Import cycles in `@`-chains: visited-set prevents infinite recursion.
- A rule file whose frontmatter is malformed (opening `---` with no closer): treated as always-loaded (no top-level `paths:` observable) — the same resolution the plan-phase derivation used.
- CWD-is-a-subdirectory sessions: the aggregate roots at `resolveProjectRoot(input)` (write-side), never `input.CWD` — the t1160 precedent.

## §D.3 — Quality gates

- `go vet ./internal/hook/` — clean.
- `go test ./internal/hook/ -count=1` — `ok` (package-scoped; full-suite verdict is CI's).
- `golangci-lint run internal/hook/` — clean, CI-pinned version.
- Every `-run` selector in this SPEC is anchored `^Test...$` (lint `VacuousTestAssertion` prevention).

## §D.4 — Definition of Done

All three release-blocking ACs green on the implementation tree with the RED-now cells above as their starting observations; AC-INSTRBUDGET-4's figure recorded in progress.md; tdd RED-phase observations for the trio recorded per `tdd-result-contract.md`; sync-phase closes per the 3-phase contract.
