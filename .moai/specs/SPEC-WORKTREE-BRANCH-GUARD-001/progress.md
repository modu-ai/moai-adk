# SPEC-WORKTREE-BRANCH-GUARD-001 — Progress

> Plan-phase artifact set created 2026-07-28 v0.1.0; audit-fix iter-2 v0.1.1.
> Status: draft (manager-spec owns the `(none) → draft` transition).
> Next phase owner: manager-develop (run-phase, `draft → in-progress`).

## §A. Plan-Phase Artifact Status

| Artifact | Status | Path |
|----------|--------|------|
| spec.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/spec.md` |
| plan.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/plan.md` |
| acceptance.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/acceptance.md` |
| progress.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/progress.md` |

## §B. Plan-Phase Research Summary

- All 5 verified preconditions confirmed (settings defaultMode, PreToolUse
  infra, static deny entries, `.worktreeinclude` divergence, `workflow.worktree`
  config).
- 3 design questions resolved: manager-git exemption boundary (hybrid
  identity-based), discriminant portability (`--path-format=absolute` primary,
  `--absolute-git-dir` + cwd-normalize fallback), fail-open with advisory.
- 1 research contradiction surfaced: `HookInput.AgentType` is main-thread
  `--agent`-only (NOT populated for `Agent(manager-git, ...)` sub-agent spawns).
  Closed by REQ-WBG-011b — sentinel env var `MOAI_BRANCH_GUARD_EXEMPT=1`.
- Census P1-B constraint analyzed: the deny rides the fast regex path
  (`checkBashCommand`), NOT the 10.58s quality-gate path (which only fires for
  `quality.IsGitCommit`).

## §C. Plan-Phase Decisions (encoded in spec.md)

| Decision | Resolution | Rationale |
|----------|------------|-----------|
| manager-git exemption boundary | Identity-based, unconditional within agent identity | Phase-D scoping rejected (not observable at PreToolUse time) |
| Discriminant portability | `--path-format=absolute` (git 2.31+) + `--absolute-git-dir` fallback | macOS Apple Git-155 = 2.50.1 supports the flag |
| Fail-open vs fail-deny | Fail-open with advisory (stderr + audit log) | Bash Risk-Amplifier Doctrine + Recovery-Signal Carve-Out norm |
| Sentinel deny reason prefix | `BRANCH_GUARD_VIOLATION:` | Orchestrator pattern-match without parsing full reason |
| Sentinel env var | `MOAI_BRANCH_GUARD_EXEMPT=1` | Closes the `agent_type` main-thread-only gap |

## §D. NEEDS-CLARIFICATION (orchestrator to resolve before run-phase)

- **B-1 (minor, non-blocking)**: orchestrator-side spawn contract for setting
  `MOAI_BRANCH_GUARD_EXEMPT=1` when spawning `manager-git` for Phase D work.
  The hook works correctly without it (deny fires as designed); the env var
  merely provides the exemption path. Plan.md §B-1 documents the one-line
  orchestrator-side edit required.

## §E.1 Plan-phase Audit-Ready Signal

_Populated by manager-spec at plan-phase completion; updated v0.1.1 (iter-2)._

- All 4 plan-phase artifacts (spec.md + plan.md + acceptance.md + progress.md)
  exist with `status: draft`, `version: 0.1.1`.
- `moai spec lint --strict` PASS on all 3 main artifacts (0 findings).
- All 13 REQs written in GEARS notation (no residual `IF/THEN` modality —
  verified via `grep -cE '^\s*(IF|THEN)\s' spec.md` → 0).
- All 13 ACs cite mechanical verification commands with falsification arms.
- Out of Scope sections satisfy `OutOfScopeRule` in all 3 files (spec.md 6
  H3 sub-headings; plan.md §H, acceptance.md §F — each with `-` bullets).
- SPEC ID regex self-check PASS: `SPEC-WORKTREE-BRANCH-GUARD-001`.
- Frontmatter carries all 12 canonical fields.
- 1 research contradiction surfaced (agent_type main-thread-only) and closed
  via REQ-WBG-011b.
- Plan-audit iter-1 verdict: PASS-With-Debt 0.84 (Tier M threshold 0.80
  cleared). Iter-2 fixes applied: D1-D7 + N1-N2 (8 findings total).
- 1 NEEDS-CLARIFICATION remains (B-1/D7: orchestrator-side spawn contract —
  deferred to follow-up SPEC per orchestrator iteration-2 decision;
  `[NEEDS CLARIFICATION: ...]` canonical marker form in plan.md).
- Ready for Implementation Kickoff Approval (plan→run HUMAN gate).

## §E.2 Run-phase Evidence

- **Run-phase entry 2026-07-28** (orchestrator). Pre-Spawn Sync Check executed:
  `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD`
  returned `3 1` (diverged — main advanced 3 commits via #1184/#1185/#1186,
  branch 1 ahead). File-overlap check vs planned work files = **0 overlap**.
  Active-session query `moai session list --filter-spec=...` = `[]`.
  Resolution: `git merge origin/main` (merge commit `c6139581f`, clean — no
  rebase, preserves plan commit SHA `0e9e3ee26` per
  `feedback_rebase_orphans_recorded_shas`). Post-merge divergence = `0 2`
  (clean; main fully absorbed). Working-tree dirty file `llm.yaml` is
  runtime-managed (`team_mode: glm`, §22.3) — excluded from pathspec commits.
  Implementation Kickoff Approval: granted (prior session, semi-autonomous
  progression — checkpoint after M1-M2).
- **M1-M2 implemented 2026-07-28** (manager-develop). Commits:
  - `9065d14e9` M1 — `branch_guard.go` (discriminant `isPrimaryCheckout`, exemption
    `isExemptAgent`, `branchStatePatterns` regex set named+blankable via `@MX:ANCHOR`,
    `checkBranchState`, `execCommand` mock indirection) + `branch_guard_test.go`
  - `fe3a50558` M2 — `pre_tool.go` Handle integration (checkBranchState after
    checkBashCommand, before default-allow) + `pre_tool_test.go` (4 integration tests)
  - Frontmatter `status: draft → in-progress` on spec/plan/acceptance (M1 commit).
- **Orchestrator independent verification 2026-07-28** (trust-but-verify, all run in
  the L2 worktree — not the stale gopls workspace that emitted wrong-tree "undefined"
  diagnostics): `go build ./...` exit 0; `go vet ./internal/hook/...` exit 0; full
  `go test ./internal/hook/... -count=1` → ALL packages `ok` (no cascade); `golangci-lint
  run ./internal/hook/...` → 0 issues; subagent-boundary grep → 0; `branch_guard.go`
  coverage 87.8% per-function avg / 91.7% statement (≥85% AC §C gate met); commits
  pushed `0 0` clean, 🗯 MoAI trailers present (2).
- **AC status (M1-M2 scope)**: AC-WBG-001/002/004/005/011/012 PASS; AC-WBG-003 PASS
  at branch-state layer (see Gap-G1). AC-006..009/010/013 deferred to M3-M6.
- **Gap-G1 (manager-develop surfaced, orchestrator confirmed in-scope)**:
  `git\s+reset\s+--hard` is in the pre-existing `AskBashPatterns` (`pre_tool.go:250`),
  so `checkBashCommand` returns DecisionAsk BEFORE the branch-state check for that ONE
  command. The other 6 patterns are NOT in AskBashPatterns → branch-guard is the first
  gate for them (works as designed). `git reset --hard` pre-gating == the documented
  Phase-D interim break (acceptance.md F-1, spec.md §E — "static deny ALSO blocks
  Phase D, not addressed by this SPEC"). AC-WBG-003 end-to-end via `git reset --hard`
  is not satisfiable through Handle; the exemption is proven at the branch-state layer
  + via non-ask command `git switch main` (`Handle_AgentType_manager_git_nonAskCommand`).
  Sync-phase action: note this in AC-WBG-003 verification (M6).
- **Gap-G2 (confirmed, defensible)**: `branch_guard.go:82` uses `[^\s-]` not `\S`
  (plan table `\S` would false-positive on `git checkout -- <path>`, violating the
  plan's own "non-flag token" rule + SPEC §E E-5). Documented in-source (lines 62-68).
  Faithful to prose intent; plan table token is a typo.
- **M3-M6 implemented 2026-07-28** (manager-develop). Commits `7fd49d48f` (M3 rule
  v1.1.0+mirror), `7c24044d8` (M4 CLI advisory), `84f257f3e` (M5 .worktreeinclude),
  `14004045e` (M3 sanitized-pair fix), `df71f1f84` (M6 tests).
- **Orchestrator independent verification (M3-M6)**: `go build ./...` exit 0;
  `go vet ./internal/hook/...` exit 0; full `go test ./internal/hook/... -count=1` →
  ALL `ok` (27s, no cascade); `go test ./internal/template/... -count=1` → `ok` (§25
  guard + mirror + sanitized-pair all green); `golangci-lint run` hook+cli+template →
  0 issues. LSP "undefined"/"unused" diagnostics were again wrong-tree (gopls workspace
  ≠ worktree); the test package compiles (`go test -run xxx` → `ok [no tests to run]`
  proves compilation) and `emitWorktreeAdvisory` IS wired (init.go:683, update.go:312/
  1039, web.go:81). `moai init` advisory grep = 1 (AC-WBG-009 empirically PASS).
- **AC status (full)**: AC-WBG-001/002/004/005/006/008/009/010/011/012/013 PASS;
  AC-WBG-003 branch-state-layer PASS (Gap-G1 documented); **AC-WBG-007 PASS-WITH-DEBT**
  (sanitized-pair, not byte-parity — see §25 resolution below).
- **§25 resolution (orchestrator premise corrected)**: the M3 delegation premise
  ("~10 mirrored rules carry SPEC IDs, byte-mirrored") was WRONG. The §25 CI guards
  (`TestTemplateNoInternalContentLeak`, `TestRuleProvenanceAudit`) FLAG a SPEC-ID +
  REQ-token template mirror; existing byte-mirrored rules (spec-workflow.md,
  session-handflow.md, hooks-system.md, model-policy.md) carry 0 SPEC-IDs in BOTH
  trees. The branch-guard rule CANNOT be both byte-identical AND carry a SPEC-ID
  under §25. **Resolution applied** (matches runtime-recovery-doctrine / zone-registry
  precedent): source rule keeps SPEC-ID+REQ tokens (AC-WBG-006); template mirror is
  §25-sanitized (SPEC-ID→generic prose); rule enrolled in the pre-existing
  `sanitizedPairPaths` registry (`sanitized_pair_parity_test.go`), `TestSanitizedPairParity`
  enforces doctrine parity, `TestTemplateNoInternalContentLeak` passes.
  **SPEC-body drift**: REQ-WBG-007/AC-WBG-007 say "byte-identical + workflowOptMirroredPaths";
  reality is sanitized-pair. Needs a manager-spec amendment before/at sync (the D-NEW-1
  inline-fix pattern). The sanitized-pair implementation is correct; only the SPEC
  wording was based on the orchestrator's false premise.

## §E.3 Run-phase Audit-Ready Signal

_Populated by orchestrator 2026-07-28 (run-phase functionally complete)._

- M1-M6 all implemented, committed, pushed to `origin/feature/SPEC-WORKTREE-BRANCH-GUARD-001`
  (commits `9065d14e9`..`df71f1f84`; `origin` sync `0 0` clean; 🗯 MoAI trailers present).
- All ACs satisfied: 11 PASS + AC-WBG-003 branch-state-layer PASS (Gap-G1 documented) +
  AC-WBG-007 PASS-WITH-DEBT (sanitized-pair, §25 resolution — SPEC-body amendment pending).
- `go build ./...` + cross-platform (GOOS=windows) exit 0; `go vet` exit 0; full test
  suites (hook + template) green, no cascade/regression; `golangci-lint` 0 issues.
- `branch_guard.go` coverage 94.1% floor / 91.7% statement (≥85% AC §C gate).
- Subagent-boundary grep (C-HRA-008): 0 in new code.
- Two orchestrator premise corrections captured: (1) LSP diagnostics twice wrong-tree
  (resolved via authoritative `go build`/`go test`); (2) §25 mirrored-rule SPEC-ID
  premise wrong (resolved via sanitized-pair).
- **Audit-ready**: YES, with two documented debts → sync-phase actions:
  (a) manager-spec amends REQ-WBG-007/AC-WBG-007 (byte-parity → sanitized-pair);
  (b) manager-docs notes AC-WBG-003 Gap-G1 (branch-state-layer PASS, end-to-end
  `git reset --hard` unsatisfiable via pre-existing ask-gate) in AC verification.
- **Next phase**: sync (manager-docs) — CHANGELOG, frontmatter `in-progress →
  implemented → completed` (single sync commit), `sync_commit_sha`, PR (repo is
  PR-mandatory per repo-local-pr-policy.md). REQ-WBG-007 amendment routes to
  manager-spec within sync (manager-docs cannot edit spec/acceptance body).

## §E.4 Sync-phase Audit-Ready Signal

_Populated by manager-docs 2026-07-28 (sync-phase 3-phase close)._

- **sync_commit_sha**: `e89d01461`
  (PR #1192 squash-merge commit; backfilled in this follow-up commit per the
  self-referential-SHA pattern documented in
  `.claude/rules/moai/development/spec-frontmatter-schema.md` § D3 SHA
  placeholder backfill exemption).
- **sync_status**: completed (single sync commit carries the merged
  `in-progress → implemented → completed` transition on spec.md / plan.md /
  acceptance.md; `updated:` refreshed to 2026-07-28 on all three).
- **Run-phase complete**: YES. M1-M6 all implemented, committed, pushed
  (commits `9065d14e9`..`df71f1f84`, `4764e2579`; `origin` sync clean).
- **ACs satisfied (13 total)**: 11 PASS + 2 documented-debt:
  - PASS: AC-WBG-001/002/004/005/006/008/009/010/011/012/013
  - PASS (branch-state layer, Gap-G1 documented): AC-WBG-003 — end-to-end
    `git reset --hard` allow unsatisfiable via pre-existing `AskBashPatterns`
    pre-gate (returns DecisionAsk before branch-state check); exemption proven
    at branch-state layer + via non-ask `git switch main`. Documented Phase-D
    interim break (acceptance.md F-1, spec.md §E).
  - PASS-WITH-DEBT (sanitized-pair): AC-WBG-007 — REQ amended in `4764e2579`
    after orchestrator M3 premise correction (§25 forbids SPEC-ID in template
    mirror; byte-parity impossible without violating
    `TestTemplateNoInternalContentLeak`). Sanitized-pair implementation is
    correct; mirror enrolled in `sanitizedPairPaths` registry.
- **Build/test/lint/coverage green**: `go build ./...` exit 0; `GOOS=windows`
  cross-build exit 0; `go test ./internal/hook/... ./internal/template/...
  -count=1` ALL `ok` (no cascade); `golangci-lint run` 0 issues;
  `branch_guard.go` coverage 94.1% floor / 91.7% statement (≥85% AC §C gate).
- **§25 mirror cleanliness**: sanitized-pair — `TestTemplateNoInternalContentLeak`
  PASS (mirror carries 0 SPEC-ID/REQ tokens); `TestSanitizedPairParity` PASS
  (doctrine parity between source and sanitized mirror).
- **B12 self-test (CHANGELOG emission discipline)**:
  - `grep -c 'SPEC-WORKTREE-BRANCH-GUARD-001' CHANGELOG.md` = 1 (single entry,
    no duplicate from parallel BATCH-SYNC sessions).
  - AC count: acceptance.md `grep -cE '^### AC-WBG-[0-9]+'` = 13 — matches
    CHANGELOG entry (11 PASS + 2 documented-debt = 13).
  - File paths cited in CHANGELOG verified via `ls` before commit
    (`internal/hook/branch_guard.go`, `internal/hook/pre_tool.go`,
    `internal/cli/{init,update,web}.go`, both rule paths, both
    `.worktreeinclude` paths, both template test files).
- **Frontmatter status transitions**: `in-progress → implemented → completed`
  on all 3 artifacts (spec.md + plan.md + acceptance.md); `updated:` field
  refreshed to 2026-07-28 on all 3. No body sections touched (only `status:`
  + `updated:` per the spec-frontmatter-schema forbidden-crossings rule).
- **Follow-up SPEC (recorded, not blocking)**: orchestrator-side spawn
  contract for setting `MOAI_BRANCH_GUARD_EXEMPT=1` when spawning
  `Agent(manager-git, ...)` Phase D invocations (plan.md §B-1 / acceptance.md
  F-1). Until that follow-up lands, manager-git Phase D invocations via
  sub-agent spawn WILL BE DENIED; Phase D must run via main-thread
  `claude --agent manager-git` (where `HookInput.AgentType == "manager-git"`
  populates correctly).
- **Audit-ready**: YES.

## §F Phase 4 Mode Selection

Recorded by orchestrator before the first run-phase `Agent()` spawn
(`orchestration-mode-selection.md` §D logging contract).

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | M |
| scope (file count) | ~8-10 (branch_guard.go + test, pre_tool.go ext, pre_tool_test.go ext, integration test, rule + template mirror, rule_template_mirror_test.go, .worktreeinclude x2, init/update/web.go) |
| domain count | 4 (hook Go code, rule/template mirror, CLI advisory, worktree config) |
| file language mix | Go source + markdown rules + YAML — mixed |
| concurrency benefit | LOW (coding-heavy implementation: regex logic, exec.Command dispatch, hook wiring) |

### Mode evaluation

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | multi-file Go implementation, not a typo |
| 2 background | no | write-capable, must commit |
| 3 agent-team | no (RETIRED) | tombstone |
| 4 parallel | no | multi-domain yes, but coding-heavy → Anthropic coding-task parallelism caveat |
| 5 sub-agent | **YES** | sequential coding work; M1→M2 tightly coupled (M2 wires M1 functions) |
| 6 workflow | no | <30 files, semantic/new-code not uniform-mechanical |

### Decision

`Decision: sub-agent` (Mode 5, sequential)

### Justification

This is coding-heavy implementation (regex matching, `exec.Command` dispatch
indirection, PreToolUse handler wiring) where sequential sub-agent delegation
is the safe default per Anthropic's coding-task parallelism caveat. M1
(discriminant + exemption + branch-state regex in `branch_guard.go`) and M2
(wiring `checkBranchState` into `pre_tool.go` Handle) are tightly coupled —
M2 calls M1's functions — so a single `manager-develop` spawn covers both to
avoid context handoff. Progression mode: semi-autonomous (checkpoint with the
user after M1-M2, per the kickoff-approved resume). M3-M6 (rule/template
mirror, CLI advisory, `.worktreeinclude` reconciliation, tests + sync) are
deferred to after the M1-M2 checkpoint.
