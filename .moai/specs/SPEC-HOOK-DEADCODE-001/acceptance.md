---
id: SPEC-HOOK-DEADCODE-001
title: "internal/hook package dead-code cleanup (3 corroborated scopes)"
version: "0.1.0"
status: in-progress
created: 2026-07-03
updated: 2026-07-12
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tier: M
tags: "cleanup, dead-code, hook, refactor, go, internal-hook"
---

# SPEC-HOOK-DEADCODE-001 — Acceptance Criteria

## §A. Given-When-Then Scenarios

### AC-HDC-001 — M1: agents/ + lifecycle/ packages absent from binary dependency graph

- **Given** `internal/hook/agents/` and `internal/hook/lifecycle/` have been deleted (18 files),
- **When** `go list -deps ./cmd/moai | grep -E 'internal/hook/(agents|lifecycle)'` is run,
- **Then** the output is empty and the command exits 1 (grep no-match), confirming both packages are absent from the binary's dependency graph.

### AC-HDC-002 — M1: build and full test suite green after deletion

- **Given** M1's 18-file deletion has been applied,
- **When** `go build ./...` and `go test ./...` are run,
- **Then** both exit 0 with no compile errors and no test failures anywhere in the repository (not just `internal/hook`) — confirming no hidden production or test dependency on the deleted packages was missed.

### AC-HDC-003 — M1: no dangling doc reference to the deleted factory.go path

- **Given** M1's deletion + doc-hygiene edit to `agent-hooks.md` (live + template mirror) has been applied,
- **When** `grep -rn 'internal/hook/agents/factory.go' .claude/ internal/template/templates/.claude/` is run (excluding `.claude/worktrees/`),
- **Then** zero matches are found — the "Handler Architecture" paragraph no longer names the deleted path.

### AC-HDC-004 — M2: the 5 dual_parse.go functions are unreachable-and-gone (not just unreachable)

- **Given** `internal/hook/dual_parse.go` has been deleted,
- **When** `deadcode ./cmd/moai | grep -E 'ParseHookOutput|synthesizeFromExitCode|ValidateHookResponse|ToHookOutput|ToHookResponse'` is run,
- **Then** the output is empty (the functions no longer exist to be reported as unreachable — distinct from merely staying unreachable-but-present).

### AC-HDC-005 — M2: HookResponse type and resolver.go consumer are untouched

- **Given** M2 has been applied,
- **When** `grep -n 'type HookResponse struct' internal/hook/response.go` and `grep -c 'HookResponse' internal/permission/resolver.go` are run,
- **Then** `type HookResponse struct` is still present at its original definition site in `response.go`, and `grep -c 'HookResponse' internal/permission/resolver.go` returns the same count as the measured pre-M2 baseline of **9** matched lines (per `spec.md` §A.3). `resolver.go` is MUST-PRESERVE and untouched by M2, so any deviation from 9 signals accidental collateral modification and fails this AC.

### AC-HDC-006 — M2: build and full test suite green, zero undefined-symbol errors

- **Given** M2 has deleted `dual_parse.go`, `dual_parse_test.go`, and the 3 dependent test functions in `response_test.go` (`TestPermissionDecisionValues`, `TestHookResponseContinue`, `TestHookResponseAdditionalContextTruncation`),
- **When** `go build ./...` and `go test ./...` are run,
- **Then** both exit 0 — specifically confirming no `undefined: ValidateHookResponse` / `undefined: ToHookOutput` / `undefined: ParseHookOutput` / `undefined: synthesizeFromExitCode` / `undefined: ToHookResponse` compile errors anywhere (this is the specific failure mode M2's expanded scope, spec.md §A.3, exists to prevent).

### AC-HDC-007 — M2: no dangling doc-comment reference to ParseHookOutput

- **Given** M2 has updated the doc comment at `internal/hook/response.go:9`,
- **When** `grep -n 'ParseHookOutput' internal/hook/response.go` is run,
- **Then** zero matches are found.

### AC-HDC-008 — M3: HookInput.Data has zero writers and zero readers repo-wide

- **Given** M3 has removed the `input.Data = actionJSON` injection in `internal/cli/hook.go`'s `runAgentHook`,
- **When** `grep -rn '\.Data\b' --include='*.go' internal/ cmd/ pkg/ | grep -v _test.go` is run,
- **Then** zero matches are found anywhere in production code (previously exactly 1 match: the now-removed `HookInput.Data` write site at `internal/cli/hook.go:340`). This AC targets **`HookInput.Data`** specifically. The broad `\.Data\b` pattern is safe here because the only `.Data` *field access* in production is that single `HookInput.Data` write — the same-named live `HookOutput.Data` field (`internal/hook/types.go:380`) is written ONLY via the `Data:` struct literal inside `NewAllowOutputWithData` (which `\.Data\b` does not match) and is never read via `.Data`, so it does not appear in this grep. Do NOT conflate the two fields, and do NOT touch `HookOutput.Data` / `NewAllowOutputWithData`.

### AC-HDC-009 — M3: moai hook agent subcommand and wrapper remain fully functional (negative-deletion check)

- **Given** M3 has been applied,
- **When** (a) `grep -n 'runAgentHook' internal/cli/hook.go` is run, (b) `ls .claude/hooks/moai/handle-agent-hook.sh` and `ls internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` are run, and (c) `grep -l '^hooks:' .claude/agents/moai/{manager-develop,manager-docs,manager-spec,sync-auditor}.md` is run,
- **Then** (a) `runAgentHook` is still defined and still registered as the `agent` subcommand's `RunE`, (b) both wrapper files still exist, and (c) all 4 agent files still declare a `hooks:` frontmatter block — confirming the §A.4-corrected scope (delete only the dead `Data` field, preserve everything else) was honored.

### AC-HDC-010 — M3: existing hook subcommand tests still pass

- **Given** M3 has been applied,
- **When** `go test ./internal/cli/... -run 'Hook|Agent'` is run (covering `hook_test.go`, `hook_e2e_test.go`, `hook_pre_push_test.go`, `misc_coverage_test.go`, `coverage_fixes_test.go`, `mx_query_test.go` — the 6 files confirmed at plan-phase to exercise the `agent` subcommand),
- **Then** all matched tests pass — confirming the `Data`-field removal did not alter `runAgentHook`'s observable behavior (event inference, registry dispatch, exit-code handling).

### AC-HDC-011 — M3: doc corrections are factually accurate and mirror-identical

- **Given** M3 has corrected `agent-hooks.md`'s Actions table (adding the `sync-auditor` → `evaluator-completion` row) in both live and template-mirror copies,
- **When** (a) `diff .claude/rules/moai/core/agent-hooks.md internal/template/templates/.claude/rules/moai/core/agent-hooks.md` is run, (b) `grep -n 'evaluator-completion' .claude/rules/moai/core/agent-hooks.md` is run, (c) `grep -n 'handle-teammate-idle.sh.*TeammateIdle\|handle-task-completed.sh.*TaskCompleted' .claude/rules/moai/core/hooks-system.md` is run (positive verification that `hooks-system.md` correctly attributes TeammateIdle/TaskCompleted to their dedicated wrapper scripts), and (d) `diff .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md` is run,
- **Then** (a) the `agent-hooks.md` diff is empty (byte-identical mirror pair after the evaluator-completion row addition), (b) the evaluator-completion grep finds ≥1 match (the newly-added row), (c) the `hooks-system.md` positive-attribution grep finds ≥1 match (confirming the doc correctly attributes TeammateIdle/TaskCompleted to `handle-teammate-idle.sh`/`handle-task-completed.sh` — this is a NON-VACUOUS positive verification: the earlier plan-phase finding claimed a false `handle-agent-hook.sh` attribution existed at line ~322, but 2026-07-12 orchestrator cross-verification found the doc was ALREADY CORRECT; this conjunct verifies that correct state is preserved through M3), and (d) the `hooks-system.md` diff is empty (byte-identical mirror pair — no M3 edit needed, the doc is already correct in both copies).
- **Note on scope narrowing (D3 remediation 2026-07-12)**: The earlier plan-phase version of this AC asserted `grep -n 'handle-agent-hook.sh.*TeammateIdle' .claude/rules/moai/core/hooks-system.md → zero` (a negative check that the false attribution was corrected by M3). That check passed VACUOUSLY because the false attribution never existed in the live doc — per `verification-claim-integrity.md` §1.1 surface 3, the "false attribution exists" claim was an unverified defect-claim (a hypothesis the domain re-check did not confirm). M3 scope is narrowed accordingly (see `plan.md` §F M3): `hooks-system.md` is NOT edited by M3; the positive-attribution grep (c) above replaces the vacuous negative check.
- **Preservation constraint (unchanged)**: The bare string `TeammateIdle, TaskCompleted` legitimately occurs at TWO lines in `hooks-system.md` — line ~74 (the general "Agent and Task Events" list: `SubagentStart, SubagentStop, TeammateIdle, TaskCompleted, TaskCreated`) and line ~350 (the registered-events timeout table) — and BOTH MUST be preserved unchanged. The positive-attribution grep (c) targets the dedicated-wrapper attribution lines (~324-325), not these event-list lines.

### AC-HDC-012 — Cross-milestone: no scope-creep into the 120 false-positive deadcode findings

- **Given** all 3 milestones have been applied,
- **When** `git diff --stat HEAD~N..HEAD -- internal/hook/` is inspected against the file list in `plan.md` §F (M1/M2/M3 "Files deleted"/"Files edited" lists),
- **Then** every changed/deleted file in the diff appears in one of the 3 milestones' explicit file lists — no additional `internal/hook/*.go` file outside the 3 corroborated scopes is touched.

## §B. Edge Cases

- **Empty-file edge case (M2)**: If, upon Read, `dual_parse.go` is found to contain any symbol beyond the 5 named functions (e.g., a shared constant or helper also used elsewhere), the run-phase agent MUST extract that symbol to a preserved location before whole-file deletion — do not blindly `rm` the file without this check (flagged explicitly in `plan.md` §F M2 scope description).
- **Test-coverage regression disclosure (M2)**: The removal of `TestHookResponseAdditionalContextTruncation` and `TestPermissionDecisionValues` is a deliberate, disclosed test-coverage reduction for dead-code-exclusive behavior (see `plan.md` §F M2 residual-risk note) — this is NOT a silent coverage drop; it must be named in the M2 completion report.
- **Concurrent unrelated edits (M1 pre-flight)**: If `git status` at run-phase start shows the `internal/hook/{pre_tool.go,registry.go,types.go}` edits from the unrelated in-flight task (plan.md §B item 1) have since been committed or reverted, re-verify the M1/M2/M3 evidence commands are still accurate against the new baseline before proceeding — do not assume plan-phase evidence is still current without re-running it (verification-claim-integrity §2 baseline-attribution).

## §C. Quality Gate Criteria

- `go vet ./...` exits 0 after each milestone.
- `golangci-lint run --timeout=2m` shows no NEW findings introduced by this SPEC's edits (pre-existing findings elsewhere in the repo are out of scope).
- `go build ./...` and `go test ./...` exit 0 after EACH of the 3 milestones individually (not only at the end) — per REQ-HDC-009, each milestone must independently land green.
- Net LOC delta is negative (this SPEC only removes code; the only additions are corrective doc-text edits and the M1/M2 comment fixes, which are net-negative-to-neutral in LOC).

## §D. Definition of Done

- [ ] All 12 acceptance criteria (AC-HDC-001 through AC-HDC-012) PASS with captured verbatim command output (per `verification-claim-integrity.md` §3 Evidence — not a summary).
- [ ] 3 milestones each landed as independently green (buildable + testable) commits.
- [ ] `plan.md` §E Self-Verification Deliverables (E1-E7) captured in the run-phase completion report.
- [ ] Zero scope-creep beyond the file lists in `plan.md` §F (AC-HDC-012).
- [ ] `progress.md` §E.2/§E.3 populated by manager-develop at run-phase completion (NOT by this agent — see Status Responsibility Matrix).
