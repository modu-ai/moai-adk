---
id: SPEC-HOOK-DEADCODE-001
title: "internal/hook package dead-code cleanup (3 corroborated scopes)"
version: "0.1.0"
status: draft
created: 2026-07-03
updated: 2026-07-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tier: M
tags: "cleanup, dead-code, hook, refactor, go, internal-hook"
---

# SPEC-HOOK-DEADCODE-001 — Implementation Plan

## §A. Context

This is a pure dead-code removal SPEC across 3 independently-landable milestones. No new production logic is introduced anywhere in this plan; every milestone is a subtraction (deletion) plus, where a subtraction leaves a dangling reference, a minimal correction (doc text, a stale comment, or a now-broken test). See `spec.md` §A for full evidence and the §A.4 material correction to the original Scope 3 framing.

## §B. Known Issues / Risks Entering Run-Phase

1. **Shared working tree has unrelated in-flight edits.** `internal/hook/{pre_tool.go,registry.go,types.go,coverage_boost_test.go}` and several `internal/cli/*.go` files carry uncommitted edits from a separate, unrelated in-progress task (task #5, "PreToolUse 기본값 모드 인식 분기 구현") at SPEC-authoring time. Run-phase MUST re-run `git status` immediately before starting M1 and confirm those files are either (a) still unrelated to this SPEC's 3 scopes, or (b) already committed/resolved — do not silently assume the baseline is unchanged from plan-phase.
2. **M2's true scope is larger than the initial handoff implied.** Deleting `dual_parse.go` requires touching `response_test.go` (3 test functions unrelated in *file* to `dual_parse.go` but dependent in *symbol* on its functions) — this is a cross-file edit, not a clean single-file deletion. See `spec.md` §A.3.
3. **M3's design-option framing changed.** The original evidence suggested deleting the wrapper script and retiring the doc; independent verification found the wrapper is live (4 agents depend on it). Run-phase MUST implement the corrected, narrower M3 scope in `spec.md` §A.4 — NOT the original "delete wrapper" framing that may still circulate in earlier chat context.
4. **Doc mirror-parity.** `agent-hooks.md` and `hooks-system.md` each exist in 2 copies (live `.claude/rules/moai/core/` + template mirror `internal/template/templates/.claude/rules/moai/core/`). Both copies MUST be edited identically in the same commit or `internal/template/rule_template_mirror_test.go` will fail.

## §C. Pre-flight Checklist (run-phase, before M1)

- [ ] `git status --short` — confirm no new unrelated changes have landed in `internal/hook/` or `internal/cli/hook.go` since plan-phase.
- [ ] `go build ./...` — confirm green baseline (was green at plan-phase: exit 0).
- [ ] `go test ./...` — confirm green baseline before any deletion (plan-phase did not run the full suite; run-phase MUST establish this baseline before M1).
- [ ] Re-run the 3 corroboration commands from `spec.md` §A.2 table to confirm the evidence has not gone stale between plan-phase and run-phase.

## §D. Non-Functional Constraints

See `spec.md` §C — zero behavior change, per-milestone green build/test, Template-First mirror-parity, Template Internal-Content Isolation.

## §E. Self-Verification Deliverables (for the run-phase agent's completion report)

Per `verification-claim-integrity.md`, every milestone's completion claim MUST be backed by verbatim command output, not a summary:

- **E1**: Per-milestone AC PASS/FAIL matrix (one row per acceptance criterion in `acceptance.md`).
- **E2**: `go build ./...` verbatim output, captured immediately after each milestone's deletion (3 captures total, one per milestone).
- **E3**: `go test ./...` verbatim output, captured immediately after each milestone's deletion (3 captures total).
- **E4**: `go list -deps ./cmd/moai | grep -E 'internal/hook/(agents|lifecycle)'` verbatim output after M1 (expected: empty, exit 1).
- **E5**: `deadcode ./cmd/moai | grep -E 'ParseHookOutput|synthesizeFromExitCode|ValidateHookResponse|ToHookOutput|ToHookResponse'` verbatim output after M2 (expected: empty).
- **E6**: `grep -rn '\.Data\b' --include='*.go' internal/ cmd/ pkg/ | grep -v _test.go` verbatim output after M3 (expected: empty — zero writers, zero readers).
- **E7**: `internal/template/rule_template_mirror_test.go` test result after any doc edit (M1's doc-hygiene fix, M3's doc corrections).

## §F. Milestones

### M1 — Delete `internal/hook/agents/` + `internal/hook/lifecycle/`

**Scope**: Delete all 18 files (13 production + 5 test, 2871 LOC combined — see `spec.md` §A.3 for the exact breakdown) under `internal/hook/agents/` and `internal/hook/lifecycle/`. Additionally, perform a minimal doc-reference hygiene fix: `agent-hooks.md` (live + template mirror) names `internal/hook/agents/factory.go` by exact path in its "Handler Architecture" section — after this milestone, that path no longer exists. Update that paragraph to describe the actual (generic `EventType`-based) dispatch mechanism instead of the now-deleted per-agent factory, without waiting for M3 (M3 owns the *rest* of the doc correction — the Actions-table staleness and the `hooks-system.md` factual error — but the dangling-path reference created by *this* milestone's own deletion must not be left dangling by this milestone).

**Files deleted** (13 production + 5 test):
- `internal/hook/agents/{backend_handler,cycle_handler,default_handler,devops_handler,docs_handler,factory,frontend_handler,quality_handler,retired_handler,spec_handler}.go` (10 files, 531 LOC)
- `internal/hook/agents/{base_handler_test,factory_test}.go` (2 files, 607 LOC)
- `internal/hook/lifecycle/{cleanup,persistence,types}.go` (3 files, 471 LOC)
- `internal/hook/lifecycle/{cleanup_coverage_test,cleanup_test,persistence_test}.go` (3 files, 1262 LOC)

**Files edited** (doc hygiene, 2 files — live + template mirror of the same content):
- `.claude/rules/moai/core/agent-hooks.md` — "Handler Architecture" section
- `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` — same edit, byte-identical

**Priority**: High (largest LOC removal, cleanest independent scope, no cross-file test dependencies).

### M2 — Delete `dual_parse.go`'s 5 unreachable functions + dependent tests + stale doc comment

**Scope**: Delete `internal/hook/dual_parse.go` (172 LOC — confirm via Read that the file contains ONLY the 5 functions + package/import boilerplate before whole-file deletion; if any other live symbol shares the file, extract it first). Delete `internal/hook/dual_parse_test.go` (375 LOC) entirely. Edit `internal/hook/response_test.go` to remove the 3 test functions that exclusively exercise the deleted functions: `TestPermissionDecisionValues`, `TestHookResponseContinue`, `TestHookResponseAdditionalContextTruncation`. Update the dangling doc-comment reference to `ParseHookOutput` at `internal/hook/response.go:9`.

**MUST-PRESERVE (do not touch)**: The `HookResponse` type definition (`internal/hook/response.go:11`) and every field on it; `internal/permission/resolver.go`'s 6 references to `ctx.HookResponse`.

**Residual-risk note for run-phase**: Deleting `TestHookResponseAdditionalContextTruncation` removes the only test coverage for 64 KiB `AdditionalContext` truncation and `TestPermissionDecisionValues` removes the only coverage for `PermissionDecision` enum validation — both behaviors exist solely inside the deleted `ValidateHookResponse` function and, per independent verification, are not applied anywhere in the live production hook-response pipeline today. This SPEC does NOT propose relocating that validation logic to a live call site (that would be new production behavior, out of scope) — it only removes dead code and its dead-code-exclusive tests. If a future need arises for `AdditionalContext` size-limiting or `PermissionDecision` validation on a live path, that is a separate feature SPEC.

**Priority**: Medium (depends on M1 having established the pattern; cross-file test edit needs care).

### M3 — Remove `HookInput.Data` dead injection + correct 2 stale doc references (corrected scope, see spec.md §A.4)

**Scope** (narrowed from the original evidence handoff — §A.4 correction applies):
1. Remove the `input.Data = actionJSON` injection (and its preceding `json.Marshal` block) in `internal/cli/hook.go`'s `runAgentHook` (~5 lines) — the field has zero readers repo-wide.
2. Correct `agent-hooks.md`'s "Agent Hook Actions" table (live + template mirror) to add the missing `sync-auditor` → `evaluator-completion` row (pre-existing staleness, unrelated to M1/M2's deletions, fixed here for doc-accuracy completeness in the same pass).
3. Correct `hooks-system.md:322` (live + template mirror) — remove the false claim that `handle-agent-hook.sh` handles `TeammateIdle`/`TaskCompleted`; those events are served by `handle-teammate-idle.sh`/`handle-task-completed.sh` per `settings.json.tmpl`.

**MUST-PRESERVE (do not touch)**: `moai hook agent <action>` subcommand registration and behavior; `.claude/hooks/moai/handle-agent-hook.sh(.tmpl)`; all 4 agent-frontmatter `hooks:` blocks (`manager-develop.md`, `manager-docs.md`, `manager-spec.md`, `sync-auditor.md`).

**Priority**: Medium (small code diff, but doc corrections require careful mirror-parity + Template Internal-Content Isolation review).

## §G. Anti-Patterns (explicitly rejected for this SPEC)

- Deleting all ~120 `deadcode`-reported functions in `internal/hook/*` in one pass without the second corroboration signal — rejected per `spec.md` § Out of Scope.
- Deleting `.claude/hooks/moai/handle-agent-hook.sh(.tmpl)` or the `moai hook agent` subcommand — rejected per §A.4 correction; these are live.
- "While I'm in `dual_parse.go` anyway" scope creep into unrelated `internal/hook/response.go` behavior changes — only the dangling comment at line 9 is touched, nothing else in that file.
- Completing the `internal/hook/agents/factory.go` wiring instead of deleting it — that is feature work, explicitly out of scope (see `spec.md` § Out of Scope).

## §H. Cross-References

- `spec.md` §A (full evidence + §A.4 correction), §B (REQ-HDC-001..010), § Exclusions
- `acceptance.md` (Given-When-Then scenarios, Definition of Done)
- `.claude/rules/moai/core/verification-claim-integrity.md` (the doctrine motivating the 2-signal corroboration methodology in §A.2)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (this agent sets only `(none) → draft`; `draft → in-progress` is manager-develop's at run-phase M1 commit start)
