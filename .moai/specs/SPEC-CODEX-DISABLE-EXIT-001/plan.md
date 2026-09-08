---
id: SPEC-CODEX-DISABLE-EXIT-001
title: "plan — moai skills disable --codex exit-code contract"
version: "0.1.0"
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
---

# plan.md — SPEC-CODEX-DISABLE-EXIT-001

## §A. Context

- Card t548 (Tier S, Class C), worktree `.claude/worktrees/t548`, branch `WT-codex-disable-exit`, base `a4855f0b2` (t577 merge into develop). Lane-9, Factory mode.
- Subject: `runCodexSkillDisable` (`internal/cli/codex_skills_disable.go:394-455`) exits 0 on five skip/no-op paths and non-zero only on name resolution (B2) plus I/O failures (B8/B9). The card's first work item is the **boundary-intent adjudication**, not a repair.
- Prerequisite verified: SPEC-CODEX-SKILL-PATH-SLASH-001 (t540) is landed in this tree — `toConfigPath` conversion precedes the guard at `:247`, and its CHANGELOG close note is present. The card's "after t540 lands" precondition HOLDS.
- Census (M1) is **DONE** — performed at plan phase per the card's mandatory-first investigation; results in §B.3 of spec.md and progress.md §E.1. Zero production in-repo callers; the only in-repo invoker is the dev-only t502 E2E evidence script.

## §B. Known Issues (relevant subset)

- **B4-class (frontmatter)**: canonical 12 fields only; `created:`/`updated:`/`tags:` — no snake_case aliases.
- **B5-class (CI 3-tier)**: `internal/cli` test changes can trip golangci-lint independently; distinguish pre-existing baseline from NEW defects at self-verification.
- **B8-class (working-tree hygiene)**: stage by explicit pathspec; the worktree carries parallel-lane residue risk like any lane tree.
- **B10-class (PRESERVE)**: do not touch `codex_skills_prune.go`, `doctor_codex.go`, or the t502 evidence script — all PRESERVE targets (spec.md Out of Scope).
- **Decision-gate hazard**: changing exit codes without the M2 adjudication silently breaks the verb's script contract; conversely, deferring the change forever keeps card rule 2 violated. M2 exists to force the explicit choice.

## §C. Pre-flight

```bash
git rev-parse --show-toplevel && git branch --show-current   # expect .claude/worktrees/t548 / WT-codex-disable-exit
git rev-parse --short HEAD                                    # baseline a4855f0b2 (re-read before any commit)
go test ./internal/cli/ -run 'TestRunCodexSkillDisable|TestUpsertCodexSkillDisable' -count=1   # green baseline before any change
```

## §D. Constraints

- PRESERVE: `internal/cli/codex_skills_prune.go` (byte-unchanged), `internal/cli/doctor_codex.go`, `internal/cli/skills.go` verb-tree shape (one verb), the documented fail-open doctrine (B1/B3/B4) and name-resolution policy (B2) under EITHER adjudication outcome.
- No `--no-verify`, no `--amend`, no force-push. Conventional Commits with card id t548 in the body.
- No Go code change at plan phase. Implementation begins only after M2's outcome is recorded.
- The verb's stdout marker strings (`Unchanged:`, `Skipped:`, `Refusing to write:`) are byte-tested by the existing suite — any change to them must be intentional and test-carried.

## §E. Self-Verification (run-phase deliverables)

Per the attribution discipline: each item names command + verbatim output + baseline SHA.

- **E1** AC binary PASS/FAIL matrix (acceptance.md AC-CDE-001..007).
- **E2** `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` — both exit 0.
- **E3** `go test -cover ./internal/cli/` — package coverage target 85% maintained.
- **E4** No AskUserQuestion in `internal/cli` (static guard convention; existing `TestNew_NoAskUserQuestion` covers the package).
- **E5** `golangci-lint run` — no NEW issues vs the pre-change baseline.
- **E6** New commit SHAs + push state (lane reports local develop merge SHA; push is the lead's batch act).
- **E8** (Outcome B only) verbatim RED failing-test output for the Skipped exit-contract test, captured on the pre-change tree before GREEN.

## §F. Milestones (ordered by decision-reversibility — the adjudication leads)

| # | Milestone | Content | Status |
|---|-----------|---------|--------|
| M1 | Caller census | Card-mandated first investigation: every caller of the verb + sibling verbs across scripts, hooks, templates, CI, tests, docs, docs-site, README. **DONE at plan phase** — census verdict: zero production in-repo callers; commands + verbatim outputs recorded in progress.md §E.1 | ✅ done |
| M2 | Boundary-intent adjudication | The decision gate. Pick Outcome A (document-only, with explicit rule-2 waiver) or Outcome B (Skipped → non-zero, one branch). RECOMMENDED default: B. Record the decision + rationale in progress.md §E.1 BEFORE any implementation commit. Operator confirmation via the orchestrator's kickoff gate | ⬜ pending |
| M3 | Implementation per adjudication | **B**: `codexSkillDisableSkipped` branch at `:422-424` returns a distinct `fmt.Errorf` carrying the guard reason (B5 Unchanged and B1/B3/B4 untouched). **A**: no Go change; documentation-only path | ⬜ pending |
| M4 | Per-branch exit-contract tests | Table-driven tests asserting each branch's exit contract: Skipped→non-zero (RED on `a4855f0b2` under B), Unchanged→0, absent-inputs→0, name-resolution→non-zero, dry-run→0, success→0. Under A: existing suite green + help-text doc test | ⬜ pending |
| M5 | Contract documentation | `--help` Long text carries the adjudicated exit-code table (REQ-CDE-008); CHANGELOG entry (under B: the behavior change; under A: the documented contract) | ⬜ pending |

## §G. Anti-Patterns

- Changing exit codes without the M2 record — the exact defect the card's rule 1 names.
- Folding Unchanged and Skipped into one non-zero bucket "for simplicity" — rule 2 requires them DISTINCT; they are different events (desired-state-already-holds vs request-refused).
- Dragging `moai clean --codex-skills` into the exit-code change to "keep the family consistent" — the sweep's per-entry `Kept:` is reporting, not a verb verdict (spec.md §D).
- Editing the t502 evidence script to satisfy a new exit code — it is a historical artifact, PRESERVE.

## §H. Cross-References

- spec.md §B.1 branch table, §B.3 census, §E adjudication outcomes
- progress.md §E.1 — census commands + verbatim outputs, code reading, adjudication analysis
- SPEC-CODEX-SKILL-DISABLE-001 (the verb's own card; documented intent surfaces), SPEC-CODEX-SKILL-PATH-SLASH-001 (t540, landed prerequisite), SPEC-CODEX-GHOST-SKILLS-PRUNE-001 (sibling surface)
- `.claude/rules/moai/development/verification-completeness.md` §2 — two-cell adoption discipline for the exit-contract ACs
