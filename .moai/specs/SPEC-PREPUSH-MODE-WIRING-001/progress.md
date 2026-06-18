# SPEC-PREPUSH-MODE-WIRING-001 — Progress

## Phase status

| Phase | Status | Commit SHA | Notes |
|-------|--------|-----------|-------|
| Plan  | complete | ec011109b | spec.md + plan.md + acceptance.md + progress.md authored. status: draft. |
| Run   | complete | 093e14922 | cycle_type=tdd, M1-M5 complete (M4 REQ-PMW-012 INCLUDED — env override fit Tier S cleanly). status: in-progress. |
| Sync  | pending  | — | — |
| Mx    | pending  | — | — |

## Plan-phase summary

- **Tier**: S (minimal — ~50-80 LOC production change: two pure helpers `resolvePrePushAction` +
  `decideExit` + one branch in `runPrePush` + optional env const).
- **REQ count**: 13 (REQ-PMW-001 .. REQ-PMW-012 + REQ-PMW-002a testability seam; REQ-PMW-012 SHOULD/optional).
- **AC count**: 13 (AC-PMW-001 .. AC-PMW-013; AC-PMW-012 conditional on REQ-PMW-012; AC-PMW-013 = gate-OFF predecessor-preservation regression pin).
- **Module**: `internal/cli` (+ `internal/config` only if REQ-PMW-012 env const).
- **Predecessor**: SPEC-PREPUSH-WIRING-001 (completed) — 1st dead-config follow-up (`enforce_on_push`).
  This is the 2nd dead-config follow-up (`git_strategy.<mode>.hooks.pre_push` severity dial).

## Precedence model (as encoded)

```
env(MOAI_ENFORCE_ON_PUSH)  >  enforce_on_push (MASTER GATE)  >  pre_push (SEVERITY dial)
                                       |                              |
                              gate OFF (default) ⇒ no-op,      gate ON ⇒ skip / warn / enforce
                              pre_push NEVER consulted          via ActiveModeProfile().Hooks.PrePush
```

- Fail-safe defaults: nil ModeProfile → `enforce`; unknown pre_push value → `enforce`.
- Optional `MOAI_PRE_PUSH` severity override sits BELOW the gate (never opens the gate).

## §E.1 Plan-phase audit-ready signal

- plan_complete_at: 2026-06-08
- plan_status: audit-ready
- SPEC ID self-check: `decomposition: SPEC ✓ | PREPUSH ✓ | MODE ✓ | WIRING ✓ | 001 ✓ → PASS`
- plan-auditor verdict: PASS-WITH-DEBT 0.84 (Tier S threshold 0.80); 4 defects, all orchestrator-re-verified against live source, all 4 patched:
  - D1 (SHOULD-FIX): drifted template citation — `pre_push` default at git-strategy.yaml.tmpl:34/66/104; `enforce|warn|skip` vocabulary on sibling `pre_commit` line :33/65/103 (NOT on pre_push). Fixed §A.1 + Cross-References.
  - D4 (MINOR): off-by-two — `HooksConfig.PrePush` field at types.go:92 (line 90 is the struct decl). Fixed §A.1 + Cross-References.
  - D2 (SHOULD-FIX, borderline BLOCKING): exit-2 path not in-process testable; `TestRunPrePush_WithViolations` false-named (fails at /dev/stdin, never reaches os.Exit); no subprocess harness in internal/cli/*_test.go. Added REQ-PMW-002a testability seam (pure `decideExit` + `resolvePrePushAction`); rewrote AC-PMW-002/003/005/006/007 to assert pure helpers; flagged the barrier in plan.md §A.1 + §E.
  - D3 (MINOR): added AC-PMW-013 gate-OFF predecessor-preservation regression pin (existing `TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately` + new `TestRunPrePush_GateOff_PrePushNotConsulted`); noted gate-OFF is the only legacy-harness-reachable row in plan.md §E.

## §E.2 Run-phase Evidence

cycle_type=tdd. M1 RED → M2/M3/M4 GREEN → M5 REFACTOR+verify. M4 (REQ-PMW-012
MOAI_PRE_PUSH env override) INCLUDED — it fit the Tier S budget cleanly (one env
const + one resolver branch). Run commit: `093e14922`.

> **Testability seam note (REQ-PMW-002a).** The gate-ON decision ACs
> (002/003/005/006/007) assert the PURE helpers `resolvePrePushAction()` (action
> enum) and `decideExit(action, violations) int` (exit code), NOT process exit or
> stdin injection. The gate-OFF row (001/013) is reachable in-process. No test lets
> `os.Exit(2)` run inline; no subprocess exit-code harness was added.

### Ground-truth finding (surfaced to orchestrator — does NOT block AC closure)

During M1 the implementer discovered that the `git_strategy` config section is
**never loaded from the user's `git-strategy.yaml` file** at runtime: `Loader.Load()`
(`internal/config/loader.go:31-91`) has no `loadGitStrategySection`, and the only
production assignment to `cfg.GitStrategy` (besides compiled defaults) is via
`ConfigManager.SetSection("git_strategy", ...)`, which nothing in production calls.
Consequence: `resolvePrePushAction()` reads `ActiveModeProfile().Hooks.PrePush`
correctly per the SPEC contract, but at runtime that accessor returns the compiled
default (`Mode: team`, `PrePush: warn`), NOT the user's on-disk value. The wiring is
correct per the SPEC's literal AC contract (resolver reads via `deps.Config`), and
all ACs are satisfiable/satisfied via the supported `SetSection` injection path, but
the feature is NOT end-to-end functional for a user editing the YAML file until a
**3rd dead-config follow-up** wires `loadGitStrategySection` into the loader chain.
This is a pre-existing separate dead-config (the whole `git_strategy` section), out
of this SPEC's scope (§B10 + §Exclusions). Recommended follow-up:
`SPEC-PREPUSH-LOADER-WIRING-001` (wire `git_strategy` into `loader.go`).

### AC PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-PMW-001 | PASS | `go test ./internal/cli/ -run 'TestRunPrePush.*GateOff'` | PASS (gate-OFF no-op, pre_push not consulted) |
| AC-PMW-002 | PASS | `go test ./internal/cli/ -run 'TestResolvePrePushAction_GateOnEnforce' && ...'TestDecideExit_EnforceViolation'` | PASS (resolver→enforce; decideExit(enforce,≥1)→2) |
| AC-PMW-003 | PASS | `go test ./internal/cli/ -run 'TestResolvePrePushAction_GateOnWarn' && ...'TestDecideExit_WarnViolation'` + `TestRunPrePush_WarnBranch_GateOnEmptyStdin` (named print-loop seam) | PASS (resolver→warn; decideExit(warn,≥1)→0) |
| AC-PMW-004 | PASS | `go test ./internal/cli/ -run 'TestResolvePrePushAction_GateOnSkip\|TestRunPrePush_Skip_GateOnNoOp'` | PASS (skip→nil no-op, no convention load) |
| AC-PMW-005 | PASS | `go test ./internal/cli/ -run 'TestDecideExit_CleanCommits'` | PASS (decideExit(enforce,0)→0 AND decideExit(warn,0)→0) |
| AC-PMW-006 | PASS | `go test ./internal/cli/ -run 'TestResolvePrePushAction_NilProfile'` | PASS (nil ModeProfile → enforce fail-safe) |
| AC-PMW-007 | PASS | `go test ./internal/cli/ -run 'TestResolvePrePushAction_UnknownValue'` | PASS (`garbage` → enforce normalization) |
| AC-PMW-008 | PASS | `grep -n '\.Hooks\.PrePush' internal/cli/hook_pre_push.go` | line 71: `parsePrePushAction(profile.Hooks.PrePush)` (≥1 match — dead config eliminated) |
| AC-PMW-009 | PASS | `grep -n 'func resolvePrePushAction\|type prePushAction' internal/cli/hook_pre_push.go` | line 59 resolver + line 36 enum (3 constants) |
| AC-PMW-010 | PASS | `grep -n 'isEnforceOnPushEnabled\|resolvePrePushAction' internal/cli/hook_pre_push.go` | within runPrePush: gate (line 120) precedes resolver (line 125) |
| AC-PMW-011 | PASS | `grep -c 'hooks.pre_push' internal/config/validation.go` == 3; enum-gate grep == 0 | 3 (unchanged) + 0 (no enum gate leaked) |
| AC-PMW-012 | PASS | `grep -n 'EnvPrePushMode\|"MOAI_PRE_PUSH"' internal/config/envkeys.go` + `TestResolvePrePushAction_EnvOverride` + `TestRunPrePush_EnvSeverity_GateOff` | env override implemented (M4 NOT deferred); env wins over config; env does NOT open gate |
| AC-PMW-013 | PASS | `go test ./internal/cli/ -run 'TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately\|TestRunPrePush_GateOff_PrePushNotConsulted'` | PASS (predecessor short-circuit byte-preserved; pre_push value zero effect when gate OFF) |

### Coverage (E3)

`go tool cover -func` on `internal/cli/hook_pre_push.go` (run-phase symbols):
- `resolvePrePushAction`: **100.0%**
- `parsePrePushAction`: **100.0%**
- `decideExit`: **100.0%**
- `runPrePush`: **52.4%** — the uncovered portion is exactly the `/dev/stdin` read +
  convention-validate + violation-print + `os.Exit(2)` path, documented as
  NOT in-process unit-testable per the REQ-PMW-002a seam (the `os.Exit` boundary +
  stdin barrier). The reachable branches (gate-OFF return, skip return, gate-ON warn
  empty-stdin) are covered. The two pure helpers reach the ~100% E3 target.

### Cross-platform + lint + vet

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./...` → exit 0
- `golangci-lint run ./internal/cli/... ./internal/config/... --timeout=2m` → 0 issues
- `go test ./...` (FULL suite) → exit 0, no FAIL lines (cascade check clean)
- Subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/hook_pre_push.go | grep -v _test.go` → 0 matches
- Scope discipline: `git diff --name-only` → hook_pre_push.go, hook_pre_push_test.go,
  envkeys.go, spec.md (tier:S + status) — NOT validation.go/types.go/defaults.go/templates/shell hook

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-08
run_commit_sha: 093e14922
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0   # validation.go/types.go/defaults.go/templates/shell hook untouched
l44_pre_commit_fetch: n/a (isolated worktree, single-session run)
l44_post_push_fetch: n/a (push deferred to orchestrator per Hybrid Trunk close)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host_go_build: exit 0
  windows_go_build: exit 0
total_run_phase_files: 4   # hook_pre_push.go + hook_pre_push_test.go + envkeys.go + spec.md (tier+status)
m1_to_mN_commit_strategy: single cohesive run commit 093e14922 (Tier S; M1 RED + M2/M3/M4 GREEN + M5 refactor folded — pure helpers 100% covered, runPrePush os.Exit boundary documented untestable-by-unit)
m4_status: INCLUDED (REQ-PMW-012 MOAI_PRE_PUSH env override fit Tier S; AC-PMW-012 PASS, not N/A-deferred)
ground_truth_finding: git_strategy section not loaded by loader.go — wiring correct per SPEC literal AC contract; end-to-end functional only after a 3rd dead-config follow-up wires loadGitStrategySection (out of this SPEC scope)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-08
sync_commit_sha: 4e14233c2
sync_status: audit-ready
changelog_entry: added (### Added, immediately after SPEC-PREPUSH-WIRING-001)
spec_status_transition: in-progress -> implemented
version_bump: 0.1.0 -> 0.2.0
readme_change: none (internal CLI behavior; git_strategy YAML not yet loader-wired -> advertising would mislead)
sync_method: orchestrator-direct (Tier S bounded + active-race avoidance; in-progress->implemented Authored-By-Agent trailer omitted -> OwnershipTransitionRule silent SKIP)
ground_truth_carryover: git_strategy section loader-unwired (3rd dead-config) -> recommended follow-up SPEC-PREPUSH-LOADER-WIRING-001
```

### (Migrated from §E.5)

```yaml
mx_complete_at: 2026-06-08
mx_commit_sha: 0700c559d
mx_status: audit-ready
spec_status_transition: implemented -> completed
four_phase_close: "plan ec011109b -> run 093e14922 (evidence 60b0676db) -> sync 4e14233c2 -> Mx 0700c559d"
ac_final: 13/13 PASS
plan_audit: "iter-1 0.84 PASS-WITH-DEBT -> iter-2 0.89 PASS (Tier S threshold 0.80, monotonic +0.05)"
mx_tags: none added (Tier S bounded; the 2-helper seam has no high-fan-in / danger-zone target)
followup_spec: SPEC-PREPUSH-LOADER-WIRING-001 (wire git_strategy section into loader.go - 3rd dead-config)
```
