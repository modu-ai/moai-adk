# SPEC-PREPUSH-MODE-WIRING-001 — Acceptance Criteria

> 13 ACs (AC-PMW-001 .. AC-PMW-013). AC-PMW-012 is conditional on the OPTIONAL
> REQ-PMW-012 (`MOAI_PRE_PUSH` env override) — if M4 is deferred, AC-PMW-012 is marked
> N/A-deferred and the remaining 12 ACs (001-011 + 013) are the closure gate.
>
> **Testability seam (REQ-PMW-002a / plan.md §A.1).** `runPrePush` reads `/dev/stdin` and
> calls `os.Exit` — neither is in-process unit-testable, and there is NO subprocess exit-code
> harness in `internal/cli/*_test.go`. The DECISION ACs (AC-PMW-002/003/005/006/007) therefore
> assert the two PURE helpers `resolvePrePushAction()` (action enum) and
> `decideExit(action, violationCount) int` (intended exit code) at the unit level — NOT process
> exit, NOT stdin injection. The gate-OFF row (AC-PMW-001 / AC-PMW-013) IS reachable in-process
> (short-circuit before stdin/`os.Exit`). The existing `TestRunPrePush_WithViolations` is
> false-named (fails at the `/dev/stdin` read, never reaches validation/`os.Exit`) and is NOT a
> verification mechanism for any AC here.
>
> All grep-style ACs are byte-precise and reference symbols/paths/test-names that exist
> post-implementation. While-read idiom tolerance (lesson L_ac_grep_read_r_flag): no AC below
> uses a `while read` loop, but every grep AC matches a real Go symbol/path, avoiding
> sibling-enum pollution, nonexistent paths, and vacuous infix patterns (lesson
> L_plan_audit_grep_idiom_recurrence).

## D. Behavior Matrix (Given-When-Then)

### AC-PMW-001 — Gate OFF ⇒ no-op, pre_push not consulted (preserves sibling OFF behavior)
- **Given** `enforce_on_push=false` (the shipped default) and active-mode `pre_push` set to
  any value (`skip` / `warn` / `enforce` / `garbage`)
- **When** `moai hook pre-push` runs with a violating commit subject on stdin
- **Then** the command is a no-op: `runPrePush` returns `nil` (exit 0), performs no convention
  load, no validation, no output — **identical to SPEC-PREPUSH-WIRING-001 OFF behavior**, and
  `resolvePrePushAction()` is NOT invoked (the gate-OFF short-circuit fires first).
- **Verify**: `go test ./internal/cli/ -run 'TestRunPrePush.*GateOff' -v` → PASS.

### AC-PMW-002 — Gate ON + pre_push=enforce + violating ⇒ decided exit code 2
- **Given** `enforce_on_push=true`, active mode profile with `hooks.pre_push: enforce`
- **When** the resolved action is `enforce` and at least one commit subject violates the convention
- **Then** `resolvePrePushAction()` returns the `enforce` constant AND
  `decideExit(enforce, violationCount≥1)` returns `2` (the decided block code) — asserted at the
  pure-function level, NOT by letting `os.Exit(2)` terminate the test binary (per the seam note).
  The `runPrePush` boundary is the one site that translates that `2` into `os.Exit(2)`.
- **Verify** (real post-impl test names; assert the pure decision, not process exit):
  ```bash
  go test ./internal/cli/ -run 'TestResolvePrePushAction_GateOnEnforce' -v   # PASS (returns enforce)
  go test ./internal/cli/ -run 'TestDecideExit_EnforceViolation' -v          # PASS (returns 2)
  ```

### AC-PMW-003 — Gate ON + pre_push=warn + violating ⇒ decided exit code 0 (NEW non-blocking)
- **Given** `enforce_on_push=true`, active mode profile with `hooks.pre_push: warn`
- **When** the resolved action is `warn` and at least one commit subject violates the convention
- **Then** `resolvePrePushAction()` returns the `warn` constant AND
  `decideExit(warn, violationCount≥1)` returns `0` (push allowed despite violations) — the new
  non-blocking middle state, asserted at the pure-function level. (The print-loop side effect is
  exercised separately where stdin is reachable; the DECISION asserted here is exit 0.)
- **Verify** (real post-impl test names):
  ```bash
  go test ./internal/cli/ -run 'TestResolvePrePushAction_GateOnWarn' -v   # PASS (returns warn)
  go test ./internal/cli/ -run 'TestDecideExit_WarnViolation' -v          # PASS (returns 0)
  ```

### AC-PMW-004 — Gate ON + pre_push=skip ⇒ no-op (return nil, no validation)
- **Given** `enforce_on_push=true`, active mode profile with `hooks.pre_push: skip`
- **When** `moai hook pre-push` runs with a violating commit subject on stdin
- **Then** the runtime returns `nil` (exit 0) without loading the convention or validating
  (skip is a per-mode opt-out even when the gate is on).
- **Verify**: `go test ./internal/cli/ -run 'TestRunPrePush.*Skip' -v` → PASS.

### AC-PMW-005 — Gate ON + clean commits + pre_push∈{enforce,warn} ⇒ decided exit code 0
- **Given** `enforce_on_push=true`, resolved action = `enforce` (and a second subcase = `warn`)
- **When** there are zero convention violations (`violationCount == 0`)
- **Then** `decideExit(enforce, 0)` returns `0` AND `decideExit(warn, 0)` returns `0` — the clean
  path is non-blocking for both severities, asserted at the pure-function level.
- **Verify** (real post-impl test name; table-driven over both severities at violationCount 0):
  ```bash
  go test ./internal/cli/ -run 'TestDecideExit_CleanCommits' -v   # PASS (enforce→0 AND warn→0)
  ```

### AC-PMW-006 — Gate ON + Mode empty/invalid (ActiveModeProfile false) ⇒ default enforce
- **Given** `enforce_on_push=true` and `git_strategy.mode` empty or an invalid value, so
  `ActiveModeProfile()` returns `(nil, false)`
- **When** `resolvePrePushAction()` is evaluated
- **Then** it defaults the severity to the `enforce` constant (fail-safe). Combined with
  `decideExit(enforce, ≥1)==2`, a violation would block — but the DECISION asserted here is the
  resolver returning `enforce`, at the pure-function level (no process exit).
- **Verify**: `go test ./internal/cli/ -run 'TestResolvePrePushAction_NilProfile' -v` → PASS
  (asserts `resolvePrePushAction` returns the enforce constant when ActiveModeProfile is `(nil,false)`).

### AC-PMW-007 — Gate ON + unknown pre_push value ⇒ normalized to enforce
- **Given** `enforce_on_push=true`, active mode profile with `hooks.pre_push: garbage`
  (an unrecognized value — recall `checkStringField` does NOT enum-validate it)
- **When** `resolvePrePushAction()` is evaluated
- **Then** it normalizes `garbage` to the `enforce` constant (fail-safe), at the pure-function
  level (no process exit). The downstream `decideExit(enforce, ≥1)==2` would then block, but the
  DECISION asserted here is the resolver returning `enforce`.
- **Verify**: `go test ./internal/cli/ -run 'TestResolvePrePushAction_UnknownValue' -v` → PASS.

### AC-PMW-008 — Dead config eliminated: a runtime reader of .Hooks.PrePush now exists
- **Given** the implementation is complete
- **When** the dead-config grep (from spec.md §A.1 / plan.md §C.3, which returned EMPTY
  before this SPEC) is re-run
- **Then** it now returns at least one match in `internal/cli/hook_pre_push.go` — the new
  `resolvePrePushAction()` reads `ActiveModeProfile().Hooks.PrePush`.
- **Verify** (byte-precise; matches the real post-impl symbol):
  ```bash
  grep -n '\.Hooks\.PrePush' internal/cli/hook_pre_push.go
  ```
  Expected: ≥ 1 match (the `ActiveModeProfile().Hooks.PrePush` read inside `resolvePrePushAction`).

### AC-PMW-009 — `resolvePrePushAction` resolver exists and mirrors the idiom (D1)
- **Given** the implementation is complete
- **When** the resolver helper is grepped for
- **Then** a function named `resolvePrePushAction` exists in `hook_pre_push.go`, alongside a
  `prePushAction` typed enum with three constants for skip / warn / enforce.
- **Verify**:
  ```bash
  grep -n 'func resolvePrePushAction' internal/cli/hook_pre_push.go      # ≥ 1
  grep -n 'type prePushAction' internal/cli/hook_pre_push.go             # ≥ 1
  ```
  Both expected ≥ 1 match.

### AC-PMW-010 — Precedence: gate short-circuit before pre_push read (D3 REQ-PMW-007)
- **Given** the implementation is complete
- **When** the source of `runPrePush` is inspected (ordering invariant)
- **Then** `runPrePush` evaluates `isEnforceOnPushEnabled()` and returns `nil` on false BEFORE
  any call to `resolvePrePushAction()` / any `ActiveModeProfile().Hooks.PrePush` read — proving
  `pre_push` is never consulted when the gate is off. The behavioral proof (and the dedicated
  predecessor-preservation regression) is pinned by AC-PMW-013.
- **Verify** (source ordering — the gate read precedes the resolver call in `runPrePush`):
  ```bash
  # isEnforceOnPushEnabled() appears in runPrePush BEFORE resolvePrePushAction()
  grep -n 'isEnforceOnPushEnabled\|resolvePrePushAction' internal/cli/hook_pre_push.go
  ```
  Expected: within `runPrePush`, the `isEnforceOnPushEnabled()` call line precedes the
  `resolvePrePushAction()` call line (gate-first ordering).

### AC-PMW-011 — No enum gate added to validation.go (scope discipline)
- **Given** the implementation is complete
- **When** `internal/config/validation.go` is inspected
- **Then** no new enum-validation rule for `pre_push` was added; the three `pre_push`
  `checkStringField` call sites remain dynamic-token-only (unchanged from baseline).
- **Verify** (must equal the baseline count of 3 — manual/personal/team, all `checkStringField`):
  ```bash
  grep -c 'hooks.pre_push' internal/config/validation.go    # == 3 (unchanged)
  grep -n 'prePushAction\|pre_push.*enum\|validatePrePushEnum' internal/config/validation.go  # 0 matches
  ```
  First expected `3`; second expected 0 matches (no enum gate leaked into validation.go).

### AC-PMW-012 — (CONDITIONAL) MOAI_PRE_PUSH env severity override (REQ-PMW-012, OPTIONAL)
- **Status**: applies ONLY if M4 (REQ-PMW-012) is implemented. If M4 is deferred, this AC is
  recorded as **N/A-deferred** and the closure gate is AC-PMW-001 .. AC-PMW-011.
- **Given** `enforce_on_push=true`, active mode profile with `hooks.pre_push: enforce`, AND
  env `MOAI_PRE_PUSH=warn`
- **When** `moai hook pre-push` runs with a violating commit subject
- **Then** the env severity wins over the config `pre_push`: the runtime prints violations and
  exits 0 (warn behavior). Conversely, `MOAI_PRE_PUSH` set while `enforce_on_push=false` does
  NOT turn the gate on (still a no-op — env severity sits below the gate).
- **Verify** (only when M4 present):
  ```bash
  grep -n 'EnvPrePushMode\|"MOAI_PRE_PUSH"' internal/config/envkeys.go   # ≥ 1
  go test ./internal/cli/ -run 'TestResolvePrePushAction.*EnvOverride' -v # PASS
  go test ./internal/cli/ -run 'TestRunPrePush.*EnvSeverity.*GateOff' -v  # PASS (env does NOT open gate)
  ```

### AC-PMW-013 — Gate-OFF return path byte-preserved vs predecessor (D3 regression pin)
- **Given** `enforce_on_push=false` (the shipped default AND the local project's current value)
- **When** `moai hook pre-push` runs (this is the ONE matrix row reachable in-process by the
  legacy empty-stdin harness — `runPrePush` short-circuits before any stdin read or `os.Exit`)
- **Then** `runPrePush` returns `nil` immediately, `resolvePrePushAction()` is NOT consulted, and
  the observable behavior is **identical to SPEC-PREPUSH-WIRING-001 (predecessor)** — zero
  behavior change in the gate-OFF population. The existing
  `TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately`
  (`internal/cli/target_coverage_test.go:575`) proves the unchanged short-circuit; a NEW test
  additionally asserts that a `skip`/`garbage` `pre_push` value has zero effect while the gate
  is OFF (the resolver is never reached).
- **Verify** (real existing test + new gate-OFF-isolation test):
  ```bash
  go test ./internal/cli/ -run 'TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately' -v  # PASS (predecessor short-circuit unchanged)
  go test ./internal/cli/ -run 'TestRunPrePush_GateOff_PrePushNotConsulted' -v                # PASS (pre_push value has zero effect when gate OFF)
  ```

## D.1 Edge Cases Pinned

| Edge case | Resolved behavior | AC |
|-----------|-------------------|-----|
| Gate OFF, any pre_push (predecessor-preserved) | no-op, pre_push never read, byte-identical to predecessor | AC-PMW-001, AC-PMW-010, AC-PMW-013 |
| Gate ON, ActiveModeProfile()==(nil,false) | resolver → `enforce`; decideExit(enforce,≥1)→2 | AC-PMW-006 |
| Gate ON, unknown pre_push value | resolver normalizes → `enforce` | AC-PMW-007 |
| Gate ON, pre_push=warn, violation | decideExit(warn,≥1)→0 (non-blocking, accepted downgrade) | AC-PMW-003 |
| Gate ON, pre_push=skip | no-op return nil | AC-PMW-004 |
| Gate ON, clean commits (enforce or warn) | decideExit(*,0)→0 | AC-PMW-005 |
| MOAI_PRE_PUSH while gate OFF (if M4) | still no-op (env ≠ gate) | AC-PMW-012 |

## D.2 Quality Gate / Definition of Done

- [ ] AC-PMW-001 .. AC-PMW-011 + AC-PMW-013 all PASS (AC-PMW-012 PASS or N/A-deferred).
- [ ] DECISION ACs (002/003/005/006/007) assert the pure `resolvePrePushAction` / `decideExit`
  helpers — NOT process exit, NOT stdin injection (REQ-PMW-002a seam). No test lets `os.Exit(2)`
  run inline; no new subprocess exit-code harness is added.
- [ ] `go test ./...` exit 0 (full suite, not just pre-push tests — cascade check).
- [ ] Coverage `go test -cover ./internal/cli/...` — `resolvePrePushAction` + `decideExit` +
  `runPrePush` ≥ 90% (the two pure helpers should reach ~100%).
- [ ] `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- [ ] `golangci-lint run --timeout=2m` — no NEW issues vs baseline.
- [ ] Subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/hook_pre_push.go | grep -v _test.go` → 0 matches.
- [ ] Scope discipline: `git diff --name-only` touches only `internal/cli/hook_pre_push.go`,
  `internal/cli/hook_pre_push_test.go`, and (if M4) `internal/config/envkeys.go` — NOT
  `validation.go`, `types.go`, `defaults.go`, templates, or the shell hook.
- [ ] Conventional Commits + `🗿 MoAI` trailer on every commit; no `--no-verify`.
