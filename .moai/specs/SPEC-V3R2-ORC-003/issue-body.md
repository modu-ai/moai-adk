# SPEC-V3R2-ORC-003 — Effort-Level Calibration Matrix for 17 agents (Plan PR)

> Step 1 plan-in-main per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline.
> Branch: `plan/SPEC-V3R2-ORC-003` · Base: `origin/main` HEAD `3356aa9a9` · Merge strategy: squash.

## Summary

Publish the canonical 17-agent effort-level calibration matrix in `.claude/rules/moai/development/agent-authoring.md`, populate every v3r2 agent's `effort:` frontmatter to match, harden lint posture (LR-03 verification + new LR-12/13/14), and convert the constitution's inline effort guidance into a cross-reference (FROZEN clause text preserved).

This PR contains **plan artifacts only** (research / plan / acceptance / tasks / progress / issue-body). No code or agent file changes ship in this PR — those land in the run PR (`feat/SPEC-V3R2-ORC-003`) after this plan PR squash-merges into `main`.

## Why this matters

R5 audit (problem-catalog P-A02 HIGH + P-A03 HIGH) found 32% of agents have effort calibration drift or missing values:

- **19 of 22 agents** (per spec.md §1.1) omit `effort:` entirely; session default `medium` applies silently — under-invoking Opus 4.7 Adaptive Thinking for reasoning-intensive work.
- **3 agents** are explicitly drifted in spec.md (expert-security, evaluator-active, plan-auditor declared `high` while constitution names them for `xhigh`); plan-phase research found a **4th drift** (expert-refactoring, also declared `high` while matrix targets `xhigh`). Plan binds 4 corrections.
- LR-03 already lives at Error severity per ORC-002 implementation (`internal/cli/agent_lint.go:386`); this SPEC's REQ-006 is **operationally idempotent** (regression test + comment anchor only).

## Plan-phase deliverables (this PR)

- `spec.md` (already on main; no change)
- `research.md` — 30 evidence anchors, 9 OQ resolutions, cross-SPEC boundary survey for ORC-001/-002/-004/-005, HRN-001, MIG-001, V3-AGT-001, CON-001/-002.
- `plan.md` — 5 milestones (M1..M5), 27 tasks, 14 REQ → 10 AC → 27 Task traceability matrix; 4 @MX tags planned; risk mitigation cross-referenced to spec §8.
- `acceptance.md` — 10 ACs in Given/When/Then form with verification commands, test fixtures (Go), and Definition of Done.
- `tasks.md` — 27 tasks with REQ/AC traceback, dependency graph, ORC-001 fallback path documented.
- `progress.md` — milestone tracker shell + AC status + risk watch.
- `issue-body.md` — this file.

## Run-phase scope (next PR after this plan PR merges)

After plan PR squash-merge:

1. `moai worktree new SPEC-V3R2-ORC-003 --base origin/main` (Step 2 spec-workflow).
2. `/moai run SPEC-V3R2-ORC-003` invokes Phase 0.5 plan-audit gate.
3. M1 RED test scaffolding (6 new test fixtures FAIL initially).
4. M2 lint rule + matrix table:
   - Implement LR-12 (matrix drift), LR-13 (invalid effort enum), LR-14 (fixed budget_tokens prohibition) in `internal/cli/agent_lint.go`.
   - Insert canonical matrix in `agent-authoring.md` (17 rows).
   - Constitution cross-reference (no FROZEN clause modification).
5. M3 17-agent frontmatter population:
   - 4 drift corrections (expert-security / evaluator-active / plan-auditor / expert-refactoring all `high → xhigh`) — BC-V3R2-002.
   - 13 missing-`effort:` populations on remaining v3r2 roster.
   - 2 already-correct (manager-spec, manager-strategy) untouched.
6. M4 CI integration + JSON drift (`effort_drift` field in `--format=json` output).
7. M5 verification + audit (test suite + linter + diff -r parity + CHANGELOG + @MX tags).

Estimated artifacts: 0 new files + 1 doc table + 1 doc cross-ref + 17 agent frontmatter edits × 2 trees + 3 new lint rules + 6 new test fixtures + CHANGELOG = **~430 LOC delta**.

## Canonical 17-agent effort matrix (from plan §1.4)

| Agent | Effort |
|---|---|
| manager-spec | xhigh |
| manager-strategy | xhigh |
| manager-cycle | high |
| manager-quality | high |
| manager-docs | medium |
| manager-git | medium |
| manager-project | medium |
| expert-backend | high |
| expert-frontend | high |
| expert-security | xhigh |
| expert-devops | medium |
| expert-performance | high |
| expert-refactoring | xhigh |
| builder-platform | medium |
| evaluator-active | xhigh |
| plan-auditor | xhigh |
| researcher | xhigh |

## Acknowledged discrepancies (from plan §1.2.1)

1. **Drift count: spec says 3, plan binds 4.** expert-refactoring is currently declared `effort: high` (drift from xhigh). plan §1.2.1 acknowledges; sync-phase HISTORY entry will reconcile spec.md.
2. **LR-03 promotion is operationally idempotent.** Already at Error severity per ORC-002 implementation. REQ-006 verification = regression test only; no severity flip needed.
3. **LR-12 / LR-13 / LR-14 are NEW lint rules.** ORC-002 stops at LR-10; LR-11 reserved for ORC-004; this SPEC claims 12-14.
4. **AGT_INVALID_FRONTMATTER ownership.** SPEC-V3-AGT-001 owns the schema validator. LR-13 in this SPEC is the lint-side defense-in-depth surface; both gates fire with consistent error code.
5. **expert-debug, expert-testing, expert-mobile, manager-brain, claude-code-guide are out-of-roster.** This SPEC scopes effort population to 17 v3r2 active agents per ORC-001 §5.1.

## Breaking changes

**BC-V3R2-002**: 4 agents upgrade from `effort: high` to `effort: xhigh`:
- `expert-security`
- `evaluator-active`
- `plan-auditor`
- `expert-refactoring`

Reasoning depth increases on Opus 4.7 (Adaptive Thinking allocates more tokens). Latency may increase 10-30% on these agents. Mitigation: HRN-001 harness routing override (out-of-scope this SPEC; documented in matrix section).

Migration path: SPEC-V3R2-MIG-001 migrator rewrites legacy v2 SPEC references that may have hardcoded `effort: high` for these agents (REQ-007 advisory).

## Plan-auditor target

- **PASS verdict** ≥ 0.85 first iteration.
- 14 REQs traced to ≥1 AC and ≥1 task per plan §1.5 traceability matrix.
- 10 ACs each with Given/When/Then + verification command + REQ traceback.
- 30 evidence anchors in research.md (counted [E-01]..[E-30]).
- 9 OQ resolutions in research §6 (with rationale).
- §1.2.1 explicitly addresses 3 known plan-vs-spec discrepancies (drift count, LR-03 idempotency, LR-13 ownership).
- 4 @MX tags planned (1 ANCHOR + 1 NOTE + 1 WARN + 1 ANCHOR markdown) covering all 3 of {ANCHOR, WARN, NOTE} types.
- Worktree-base alignment per Step 2 (`moai worktree new --base origin/main`).
- Parallel SPEC isolation: only `.claude/agents/moai/`, `.claude/rules/moai/{core,development}/`, `internal/cli/agent_lint{,_test}.go`, `internal/template/templates/.claude/`, `CHANGELOG.md`.

## File-level scope (run-phase)

| Layer | Files |
|---|---|
| Documentation | `.claude/rules/moai/development/agent-authoring.md` (matrix), `.claude/rules/moai/core/moai-constitution.md` (cross-ref) |
| Agent frontmatter (17 × 2 trees) | `.claude/agents/moai/*.md` (13 populate + 4 drift correct) + template parity |
| Lint rules | `internal/cli/agent_lint.go` (+90 LOC) + `internal/cli/agent_lint_test.go` (+180 LOC) |
| Build | `make build` (regenerates `internal/template/embedded.go`) |
| Trackability | `CHANGELOG.md` Unreleased entry, 4 @MX tags |

## Test plan

- [ ] M1 RED tests added (6 new fixtures FAIL initially)
- [ ] M2 GREEN tests pass after lint rule + matrix table applied
- [ ] M3 frontmatter edits land per matrix
- [ ] M4 JSON drift field tested
- [ ] M5 final verification: `go test -race -count=1 ./...` PASS, `golangci-lint run` clean, `make build` exit 0, `moai agent lint` reports 0 LR-03/12/13/14 errors on 17 v3r2 roster, `diff -r` byte-identical between local and template trees

## Dependencies

| SPEC | Status | Role |
|---|---|---|
| SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE) | merged (assumed) | Blocks (consumed) |
| SPEC-V3R2-ORC-001 (17-agent roster) | TBD at run-time | Blocks (consumed); fallback path documented |
| SPEC-V3R2-ORC-002 (LR-01..LR-10 lint) | merged | Blocks (consumed); LR-03 already Error |
| SPEC-V3R2-ORC-004 (worktree MUST) | in-flight | Reserves LR-11 |
| SPEC-V3R2-HRN-001 (harness routing) | in-flight | Downstream consumer (effort_mapping) |
| SPEC-V3R2-MIG-001 (migrator) | in-flight | Downstream consumer (drift rewrite) |
| SPEC-V3-AGT-001 (frontmatter schema) | merged | Co-defense surface (AGT_INVALID_FRONTMATTER) |

## References

- `docs/design/major-v3-master.md:L961` (BC-V3R2-002 — effort field), `:L1054` (§11.4 ORC-003 definition)
- `.moai/design/v3-redesign/research/r5-agent-audit.md` § Effort-level calibration matrix
- `.moai/design/v3-redesign/synthesis/problem-catalog.md` P-A02, P-A03
- `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy
- Anthropic guidance (Sep 2025): https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7

---

🗿 MoAI <email@mo.ai.kr>
