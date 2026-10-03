# plan.md — SPEC-INSTRUCTIONS-BUDGET-001

---
id: SPEC-INSTRUCTIONS-BUDGET-001
title: "Implementation plan — aggregate character budget for InstructionsLoaded"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
---

## §A — Context

Card t1318 mechanizes what t1303 measured by hand on 09-29. The implementation site is one Go file plus its test file: `internal/hook/instructions_loaded.go` (handler + per-file `checkCharacterBudget`, `const charBudget = 40000` at line 103) and `internal/hook/instructions_loaded_test.go`. The hook already carries every primitive this SPEC needs: `resolveProjectRoot(input)` (write-side project root), the skip-on-error read pattern (line 94), the advisory `SystemMessage` return shape (line 70-73), and `utf8.RuneCount` as the character unit (line 100). Arrival-state fact (measured on tree `2868afc0b`, this plan run): the derived 18-file set aggregates 214,155 chars — already above the operator's 210,000 ruling — so the advisory will fire from the first event after implementation. That is intended behavior of the metric, not a defect of this plan; the diet it calls for is out of scope (spec.md §F).

## §B — Known Issues (filtered to relevant)

- No aggregate metric exists anywhere in the tree (spec.md §B) — the whole point of this SPEC.
- `go doc ./internal/hook <unexported>` cannot resolve unexported symbols (measured: it reports "no symbol" even for the existing `checkCharacterBudget`) — therefore all acceptance RED-now probes are grep-based, not go-doc-based.
- An unanchored `-run` pattern trips lint's `VacuousTestAssertion` warning (observed on SPEC-WORKTREE-SWEEP-001's acceptance.md, 14 warnings) — every `-run` form in this plan and acceptance.md is anchored `^Test...$`.
- A test selector naming a not-yet-existing test is vacuously green (an unanchored `-run` selector exits 0 with `[no tests to run]`, measured) — never used as a RED-now cell.

## §C — Pre-flight

1. Work in the card worktree `.moai/worktrees/t1318-budget` (branch `WT-instr-budget`, base `2868afc0b` on local `develop` lineage).
2. Cycle type: **tdd** — write the failing test (RED), implement (GREEN), keep both result classes per `tdd-result-contract.md`.
3. Verify the base: `git rev-parse --short HEAD` → `2868afc0b` before starting; re-read before commit.
4. No template mirror is involved: `internal/hook/*.go` is Go source, not template content (`internal/template/templates/` carries `.claude/` and `.moai/` surfaces only); `make build` is NOT required by this change.

## §D — Constraints (DO NOT VIOLATE) and Design Decisions

- **D1 — advisory, not blocking.** The aggregate breach returns a `SystemMessage` in the same non-blocking shape as the per-file check (line 70-73 precedent). It must never gate the event: the current tree is already over budget, so a blocking form would fail every session at arrival — the ruling's intent is observability, and blocking is a separate operator decision.
- **D2 — constant placement: file-local.** `sessionCharBudget` sits adjacent to the existing `charBudget` in `instructions_loaded.go`. `AGENTS.local.md` §14 prefers thresholds in `internal/config/defaults.go`, but the in-file precedent is the 40k const itself and this package treats its budgets as handler-local; consistency within the file outranks the general rule (scope discipline). The auditor may re-home it; that is a one-line move after the fact.
- **D3 — derivation, not enumeration.** The set function parses `^@` imports recursively and scans `.claude/rules/moai/` for files without a top-level `paths:` frontmatter key. A hardcoded `[]string{...}` of the 18 paths is the named mutant this design exists to kill (REQ-INSTRBUDGET-004).
- **D4 — metric unit: runes.** `utf8.RuneCount` everywhere, matching line 100. External shell verification approximates with `wc -m` (UTF-8 locale); the delta between rune count and `wc -m` on this corpus is negligible and neither substitutes for the other in claims — name which one produced a figure.
- Frontmatter paths: `^@` lines are repo-relative (`@AGENTS.md`, `@.moai/config/sections/user.yaml`, `@AGENTS.local.md`) — resolve under R, never as absolute paths. Ignore anything after the path token on the line. Do not follow imports outside R.
- Frontmatter scan: a file starting with `---` on line 1 is scanned to its closing `---` for `^paths:`; a file without frontmatter is always-loaded. (This is the same selector the plan run used to derive the 13-rule set — portable shell form preserved in the evidence report.)

## §E — Self-Verification (deliverables at run close)

1. `go vet ./internal/hook/` — clean output, exit 0.
2. `go test ./internal/hook/ -count=1` — `ok`, exit 0 (package-scoped; the full suite is CI's verdict).
3. `golangci-lint run internal/hook/` — clean (use the CI-pinned version per the lane lint lesson).
4. The three regression tests named in §F M3 exist, pass, and each fails when its target behavior is muted (RED observed during tdd RED phase — record both result classes per `tdd-result-contract.md`).
5. progress.md records the implemented metric's aggregate figure on the implementation tree (feeds acceptance.md AC-INSTRBUDGET-4).

## §F — Milestones (ordered by decision-reversibility: data-model first, mechanics last)

### M1 — The derived set (data model)

New unexported function `instructionFileSet(projectRoot string) []string` in `instructions_loaded.go`:

- anchor `CLAUDE.md`, transitive `^@` import closure (repo-relative resolution, cycle-safe via a visited set, skip missing members);
- walk `.claude/rules/moai/` for `*.md`, always-loaded selector = no frontmatter block OR frontmatter without top-level `paths:`;
- deterministic order (sorted), duplicates de-duplicated.

Tests (RED first): fixture project under `t.TempDir()` — CLAUDE.md with two `@`-imports (one present, one dangling), `rules/moai/` with one always-loaded rule, one `paths:`-scoped rule, one non-md file → set contains exactly the expected members; adding a new always-loaded rule to the fixture grows the set; adding a `paths:`-scoped rule does not. **Mutant check:** a hardcoded-list implementation fails the fixture-growth assertions.

### M2 — Constant + aggregate metric + Handle integration

- `const sessionCharBudget = 210000` adjacent to `charBudget` (D2), with a comment carrying the ruling provenance: "operator ruling 2026-10-01 (card t1318 AskUserQuestion)".
- `aggregateInstructionChars(projectRoot string) int` — Σ `utf8.RuneCount` over `instructionFileSet`, skipping unreadable members (REQ-INSTRBUDGET-003).
- In `Handle`, after the existing per-file path (unchanged), when `aggregateInstructionChars(R) > sessionCharBudget`, set `SystemMessage` to a message naming aggregate, budget, and file count. Advisory only (D1). Empty set → no comparison.

### M3 — Regression trio + green

- `TestInstructionFileSetDerivation` (from M1), `TestAggregateInstructionCharsMovement` (growing one fixture member by N runes moves the aggregate by exactly N — catches wrong-set and stale-count mutants), `TestInstructionsLoadedAggregateBudget` (fixture set grown past 210,000 → SystemMessage contains budget text and count; under-budget fixture → empty output — catches boundary mutants).
- Verification batch of §E; record RED observations from the tdd RED phase alongside the GREEN output.

## §G — Anti-Patterns

- Hardcoding the 18 paths (D3) — the named mutant.
- Using an unanchored `-run` selector as a RED-now citation for a not-yet-existing test — vacuously green (§B).
- Unanchored `-run` patterns — lint `VacuousTestAssertion`.
- Reading the aggregate with byte count (`wc -c`) and reporting it as characters — name the unit; runes in code, `wc -m` externally.
- Making the breach blocking or "temporarily" raising the constant above 210,000 to quiet the arrival advisory — the ruling is recorded; changing it is an operator act.

## §H — Cross-References

- spec.md §D (requirements this plan implements), acceptance.md §D (the AC matrix).
- `internal/hook/instructions_loaded.go` — single implementation file; `internal/hook/instructions_loaded_test.go` — single test file.
- `tdd-result-contract.md`, `verification-plan-contract.md` — result and verification-key contracts for the run phase.
