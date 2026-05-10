# SPEC-V3R2-ORC-004: Worktree MUST Rule for write-heavy role profiles

> Plan PR for SPEC-V3R2-ORC-004 (Step 1 plan-in-main per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline). Run + sync execute on a fresh SPEC worktree after this PR merges.

## Summary

Promotes `isolation: worktree` from **SHOULD** to **MUST** for the 5 v3r2 write-heavy agents (`manager-cycle`, `expert-backend`, `expert-frontend`, `expert-refactoring`, `researcher`) and reaffirms MUST for the 3 team-mode role profiles (`implementer`, `tester`, `designer`). Adds `ORC_WORKTREE_*` sentinel keys to lint violation messages and introduces a new `moai workflow lint` CLI for static `workflow.yaml` enforcement.

Closes the worktree-isolation discipline gap surfaced by R5 audit problems P-A11 (six cross-file-write agents missing the flag) and P-A22 (researcher self-contradiction between frontmatter and body).

## Why this matters

R5 audit identified a silent failure mode: when MoAI orchestrator spawns multiple write-heavy agents in parallel via `Agent()` calls (e.g., `/moai run` against two SPECs simultaneously, or team-mode parallel teammates), file-write conflicts corrupt both outputs. The conflict surfaces only at git merge time, with no runtime warning.

The previous SHOULD-style rule in `worktree-integration.md` line 135 was insufficient: contributors regularly omitted `isolation: worktree` from new agent definitions, propagating the failure mode. The promotion to MUST + CI lint enforcement closes the loophole.

## Scope (this Plan PR)

Plan-phase artifacts only — no run-phase code changes in this PR.

| File | Status | Bytes (approx) |
|------|--------|----------------|
| `.moai/specs/SPEC-V3R2-ORC-004/spec.md` | EXISTING | 21,200 |
| `.moai/specs/SPEC-V3R2-ORC-004/research.md` | NEW | ~14,000 |
| `.moai/specs/SPEC-V3R2-ORC-004/plan.md` | NEW | ~16,500 |
| `.moai/specs/SPEC-V3R2-ORC-004/tasks.md` | NEW | ~6,500 |
| `.moai/specs/SPEC-V3R2-ORC-004/acceptance.md` | NEW | ~13,000 |
| `.moai/specs/SPEC-V3R2-ORC-004/progress.md` | NEW | ~3,500 |
| `.moai/specs/SPEC-V3R2-ORC-004/issue-body.md` | NEW (this file) | ~5,500 |

Total: 7 files in `.moai/specs/SPEC-V3R2-ORC-004/`.

## Run-phase scope (next PR)

Run phase opens a fresh worktree (`feat/SPEC-V3R2-ORC-004`, `--base origin/main`) and executes:

1. Add `isolation: worktree` to **4** existing agents (template-first):
   - `expert-backend.md`, `expert-frontend.md`, `expert-refactoring.md`, `researcher.md`
2. Add `isolation: worktree` to **manager-cycle.md** (conditional on SPEC-V3R2-ORC-001 merge).
3. Update `.claude/rules/moai/workflow/worktree-integration.md` line 135 SHOULD → MUST + sentinel key glossary section.
4. Update `researcher.md` body line P-A22 reconciliation ("when possible" → definitive).
5. Append sentinel keys to `agent_lint.go` LR-05 / LR-09 violation messages (`ORC_WORKTREE_MISSING`, `ORC_WORKTREE_ON_READONLY`).
6. New `internal/cli/workflow_lint.go` (~150 LOC) with `moai workflow lint` command + `ORC_WORKTREE_REQUIRED` sentinel.
7. Add `internal/cli/workflow_lint_test.go` (~120 LOC, 4 tests).
8. Extend `agent_lint_test.go` (~80 LOC, 2 sentinel tests for AC-06 / AC-07).
9. `make build && make install && moai update` to mirror templates.
10. CHANGELOG entry + 5 MX tag annotations.

Estimated run-phase PR: ~12-15 files modified, ~400 LOC added (mostly tests + new CLI), ~10 LOC modified.

## Pre-completion discovery

`research.md` §3 documents critical pre-completion findings at HEAD `3356aa9a9`:

- **LR-05 lint rule** is already implemented in `agent_lint.go:444-491` with the exact 5 v3r2 agent names hardcoded and severity set to `SeverityError`. No additional Go code change required for the classifier.
- **LR-09 lint rule** is already implemented in `agent_lint.go:520-535` with `permissionMode == "plan" && isolation == "worktree"` rejection. No additional Go code change required.
- **`workflow.yaml` role_profiles** is already 100% aligned with REQ-004 (3 worktree, 4 none).

This reduces run-phase scope from "full lint engine implementation" to "frontmatter additions + sentinel message text + new workflow lint CLI".

## Migration impact

Backwards compatibility: `breaking: false`. The change is enforcement-layer-only:

- Existing agent invocations work unchanged at runtime.
- New CI lint failures appear only for PRs that introduce or modify the 5 v3r2 write-heavy agents without declaring `isolation: worktree`. Fix is 1-line frontmatter addition.
- External forks customizing these agents see LR-05 errors on next upstream sync; migration cost <5 minutes per affected fork.

No `bc_id` declared because this is a quality-gate enforcement over a previously SHOULD-style rule, not an API or schema change.

## REQ → AC traceability summary

15 SPEC requirements (REQ-V3R2-ORC-004-001 .. -015) map to 10 acceptance criteria (AC-V3R2-ORC-004-01 .. -10). Full matrix in `plan.md` §1.4.

| REQ count | EARS pattern |
|-----------|--------------|
| 5 | Ubiquitous |
| 3 | Event-Driven |
| 2 | State-Driven |
| 2 | Optional |
| 3 | Unwanted Behavior |

## Risk register

12 risks documented in `plan.md` §8. Top 3 by impact:

1. **Worktree merge-back conflicts on overlapping file writes** (HIGH impact / LOW likelihood) — mitigation: CC runtime handles merge; team-mode file ownership prevents pre-spawn overlap.
2. **ORC-001 not merged when ORC-004 run starts** (HIGH / LOW) — mitigation: M2 conditional task with structured blocker report path.
3. **Agent Teams workflow regression from worktree overhead** (MEDIUM / LOW) — mitigation: only write-heavy agents flagged; single-file writers stay at `none`.

## Test plan

RED → GREEN → REFACTOR per `quality.yaml` `development_mode: tdd`.

- **RED**: 6 new test fixtures (2 in `agent_lint_test.go` + 4 in `workflow_lint_test.go`).
- **GREEN**: 11 implementation deltas across 4 frontmatter files, 1 rule file, 1 lint.go message text, 1 new CLI.
- **REFACTOR**: extract sentinel constants into shared location.

Integration: `make test ./...`, `moai agent lint` exit 0, `moai workflow lint` exit 0, manual injection scenarios for AC-06/07/09.

## Rollout plan

1. **This PR (plan)**: Squash-merge to main. Plan artifacts persist in main for cross-SPEC reference.
2. **Run PR**: Open after this merges. `moai worktree new SPEC-V3R2-ORC-004 --base origin/main` then `/moai run SPEC-V3R2-ORC-004`. Phase 0.5 (Plan Audit Gate) auto-runs.
3. **Sync PR**: After run merges. Continue in same worktree per `spec-workflow.md` Step 3.
4. **Cleanup**: `moai worktree done SPEC-V3R2-ORC-004` after BOTH run AND sync PRs merged.

## Cross-references

- `spec.md` (in this directory)
- `.claude/rules/moai/workflow/worktree-integration.md` §HARD Rules
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline
- `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-021..024
- SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE codification, dependency)
- SPEC-V3R2-ORC-001 (manager-cycle creator, dependency)
- SPEC-V3R2-ORC-002 (LR-05 origin, dependency)
- SPEC-V3R2-RT-006 plan structure (reference pattern for partial-pre-completion)
- R5 audit `r5-agent-audit.md` §Worktree correctness table
- `problem-catalog.md` P-A11, P-A22

## Test plan checklist (post-merge)

- [ ] `cd ~/.moai/worktrees/moai-adk-go/SPEC-V3R2-ORC-004` resolves
- [ ] `moai worktree new SPEC-V3R2-ORC-004 --base origin/main` succeeds
- [ ] Plan-Auditor verdict on this PR: PASS ≥ 0.85 first iteration
- [ ] `gh pr view <PR>` MERGED state confirmed before run-phase entry
- [ ] Auto-merge SQUASH enabled on this PR
- [ ] Plan-phase preconditions log entry appended to `progress.md`

🗿 MoAI <namgoos@gmail.com>
