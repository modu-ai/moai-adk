# progress.md — SPEC-HIERARCHICAL-TEAM-001

> Tier M (3-artifact set: spec + plan + acceptance). Era V3R6 (created 2026-08-07 ≥ `modernEraThreshold` 2026-04-01 + `phase: "v3.x target"` carries modern release prefix). The §E section skeleton below is the canonical plan-phase emission per manager-spec §E.1..§E.4 discipline; §E.2-§E.4 are placeholder headings ONLY at plan-phase (populated by manager-develop at run-phase / manager-docs at sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- `plan_status: audit-ready`
- `plan_complete_at: 2026-08-07T00:00:00Z` (ISO-8601 placeholder — manager-spec emission; the actual timestamp is the plan-phase commit's author date)
- `plan_artifacts: spec.md, plan.md, acceptance.md, progress.md` (4 files — Tier M 3-artifact set + the always-emitted progress.md)
- `tier: M`
- `req_count: 13` (REQ-LEAD-001..REQ-CLOSE-001)
- `ac_count: 14` (AC-LEAD-001..AC-REGRESS-001)
- `oq_count: 4` (OQ-1 G-CTX-FOLD, OQ-2 G-PEER-SCOPE, OQ-3 G-LEAD-TRIGGER, OQ-4 G-DEPTH-VERIFICATION — all DEFERRED to Implementation Kickoff per the AUTONOMY-TIERS precedent)
- `frontmatter_phase_valid: true` (`phase: "v3.x target"` — release-target label, NOT a prohibited lifecycle-stage token)
- `spec_lint_strict: pending` (manager-spec emits 0-error target; plan-auditor verifies at gate)

## §E.2 Run-phase Evidence

_(Run-phase evidence rows per REQ-FOLD-001 fold-row format. The literal `§E.2` heading above is the run-evidence START marker `internal/spec/era.go` `hasAnyProgressMarker` detects.)_

M1: AC-LEAD-001=PASS, AC-LEAD-002=PASS, AC-LEAD-003=PASS, AC-DEPTH-001=PASS | evidence: builder-harness commit cbeaa25c5 (agent file + depth-2 CI guard) + M1-close commit 250d88ca9 (CLAUDE.md §4 catalog amendments) | fold-at: 2026-08-07T13:00:00Z

M2: AC-WORKTREE-001=PASS, AC-WORKTREE-002=PASS | evidence: commit 62f13051b (worktree-integration.md decision tree + HARD rule re-key; agent-common-protocol.md § Background Agent concurrency-safeguard retain + binding sentence); grep residuals verified (team-mode-implementation = 0 in re-keyed sections) | fold-at: 2026-08-07T13:25:00Z

M3: AC-FOLD-001=PASS, AC-FOLD-002=PASS, AC-FOLD-003=PASS | evidence: manager-lead.md § Context-Folding Procedure (3-step: persist evidence → append fold row → /compact with retain instructions); fold-row prefix M<n>: verified non-colliding with era.go §E.* matchers; OQ-1 decision wired as workflow.yaml context_folding: opt-in (default), auto at Tier L, opt-in at Tier M, off available | fold-at: 2026-08-07T13:30:00Z

M4: AC-PEER-001=PASS, AC-PEER-002=PASS | evidence: manager-lead.md § Peer Cross-Validation (read-only Agent(general-purpose) spawn at Tier M/L; tools: omitting Write/Edit/NotebookEdit; blocker-report path on FAIL/PARTIAL; grep -c AskUserQuestion manager-lead.md = 0); OQ-2 decision applied (all ACs at Tier M/L, Tier S skips) | fold-at: 2026-08-07T13:30:00Z

M5: AC-FANOUT-001=PASS, AC-FANOUT-002=PASS | evidence: manager-lead.md § Schema-Driven Fan-Out Reduce (consume plan-research-fanout fixed-heading markdown schema verbatim; mechanical merge; contradictions annotated as named section; concurrency ceiling ≤5 concurrent leaf-worker spawns, sequencing in batches when >5) | fold-at: 2026-08-07T13:30:00Z

M6: AC-CLOSE-001=PASS, AC-REGRESS-001=PASS | evidence: orchestration-mode-selection.md §G.2 added (manager-lead as Mode-5-shaped delegation target, NOT Mode 7; Mode 1-6 catalog unchanged; --mode dispatch-axis values unchanged; MODE_TEAM_UNAVAILABLE unchanged); moai spec lint --strict on spec.md = 0 errors / 0 warnings | fold-at: 2026-08-07T13:35:00Z

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — manager-develop populates this section when all M1-M6 MUST ACs PASS.)_

## §E.4 Sync-phase Audit-Ready Signal

- `sync_status: completed`
- `sync_complete_at: 2026-08-07` (ISO-8601 placeholder — manager-docs emission; actual timestamp is the sync commit's author date)
- `sync_commit_sha: a83934d554fe79e2d7d0cdfad66c91ecdb7bef48` (self-referential-hazard workaround per `spec-frontmatter-schema.md` D3; backfilled in a follow-up commit after the sync PR merges — same pattern as SPEC-NAVIGATOR-SYNC-003 / SPEC-MOAI-MCP-SERVER-001)
- `changelog_entry_position: CHANGELOG.md `[Unreleased] → Added` (top of section, above SPEC-MOAI-MCP-SERVER-001)`
- `frontmatter_status_transitions`:
  - spec.md: `draft → in-progress → completed` (3-phase close merged on this sync commit; `updated:` refreshed to 2026-08-07)
  - plan.md / acceptance.md / progress.md: no YAML frontmatter (plain markdown) — no frontmatter transition applicable
- `canary_compliance_check`: 15/15 AC PASS (orchestrator trust-but-verify `go test ./internal/template/...` ok 37.859s; `moai spec lint --strict` clean; manager-lead Agent-carrier + AskUserQuestion=0 verified; depth-2 seal CI guard + catalog-count guards green)
- `b12_self_test`:
  - a (pre-emission grep): `grep -c 'SPEC-HIERARCHICAL-TEAM-001' CHANGELOG.md` returned 0 before emission (no duplicate)
  - b (AC count match): 15 distinct AC IDs in acceptance.md → 15/15 AC referenced in run-phase evidence (§E.2)
  - c (file path verification): every claimed file path verified via `ls` / `git ls-files` before commit

`sync_commit_sha: a83934d554fe79e2d7d0cdfad66c91ecdb7bef48`

## Retrofit record (t390)

**What was revised (in place, card t390 — lane-10)**: the citation requirement of REQ-FOLD-002 / AC-FOLD-002 was split into a two-part form across the SPEC set — (1) the local-scratch recording step is KEPT (verification commands still redirect verbatim output to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}`), and (2) what is citable changed: any cited evidence path MUST name a tracked file under `.moai/reports/<card-id>/` (export before citing). Touched surfaces: spec.md §C.3 mechanism step 1, spec.md REQ-FOLD-002 body, spec.md §H AC-FOLD-002 matrix row, acceptance.md §A evidence-persistence bullet, acceptance.md AC-FOLD-002 Given-When-Then + Verify, plan.md M3 self-verify line.

**False premise replaced** — one-line quote of the pre-revision wording (acceptance.md §A bullet): "Cited paths MUST resolve at audit time." attached to citations of `.moai/state/verify/<session>/`. The premise was false in one sentence: `.moai/state/` is gitignored, so a path under it survives to no clone, no CI runner, and no other machine — a criterion demanding audit-time resolvability cannot point citations there.

**Disproof basis**: card t375 measured the accepted criterion against reality — 231 citations of that path family across SPEC artifacts, 0 resolvable from a fresh worktree. The corrected principle is now encoded in the rule tree (landed after this SPEC closed): `.claude/rules/moai/core/agent-common-protocol.md` § Evidence export obligation — `.moai/state/verify/<session>/` is machine-local scratch; the citation target is a tracked path, in this repository `.moai/reports/<card-id>/`.

**Census (measured by card t390's lead run; recorded verbatim as the baseline for follow-up cards)**:
- 132 files under `.moai/specs/` cite `state/verify`.
- 37 files match the requirement-form (resolve/audit co-occurrence).
- The requirement-form appears in AC/REQ bodies of at least 4 SPECs — SPEC-HIERARCHICAL-TEAM-001 (this card), SPEC-TOKEN-BUDGET-STOP-001 (spec+acceptance), SPEC-GITFLOW-DOCTRINE-ALIGN-001 (acceptance), plus progress-file-level hits to be triaged as follow-ups.

Scope note: only THIS SPEC was revised. The sibling SPECs named in the census are follow-up cards, not this card's scope.
