# SPEC-PREPUSH-MODE-WIRING-001 — Implementation Plan

> Tier S (minimal). Module: `internal/cli` + `internal/config`.
> cycle_type: **tdd** (RED-GREEN-REFACTOR — new behavior with clear contract).
> Estimated production change: ~50-80 LOC (two pure helpers — `resolvePrePushAction` +
> `decideExit` — one branch in `runPrePush` + one optional env const).

## A. Context

- **Location**: project root `/Users/goos/MoAI/moai-adk-go`.
- **Branch**: `feat/SPEC-PREPUSH-MODE-WIRING-001` (current). Run-phase commits stack here.
- **SPEC artifacts**: `.moai/specs/SPEC-PREPUSH-MODE-WIRING-001/{spec,plan,acceptance}.md`.
- **Predecessor**: SPEC-PREPUSH-WIRING-001 (completed, origin/main `6e648a0be`) wired the
  `enforce_on_push` master gate end-to-end. This SPEC adds the severity dial on top.
- **Existing infrastructure (PRESERVE)**:
  - `runPrePush` validation-and-print loop (`hook_pre_push.go:72-98`) — reused verbatim.
  - `isEnforceOnPushEnabled()` (`hook_pre_push.go:154`) — reused as the master gate, unchanged.
  - `resolveConventionName()` / `resolveAutoDetectOptions()` — unchanged.
  - `ActiveModeProfile()` (`types.go:165`) — read-only accessor, reused.
  - `checkStringField` (`validation.go:296`) — deliberately NOT extended (resolver-side normalize instead).
- **EXTEND target**: add the two pure helpers `resolvePrePushAction()` + `decideExit()`
  + the 3-way branch in `runPrePush`; optionally add `EnvPrePushMode` to `envkeys.go`.

### A.1 Testability seam (REQ-PMW-002a) — REQUIRED design

[HARD design constraint — read this BEFORE writing any test or code.] `runPrePush` reads
commit subjects from `/dev/stdin` (`readStdinLines` → `os.ReadFile("/dev/stdin")`) AND
terminates the test binary via `os.Exit(2)` on the violation path. **Both are
in-process-test barriers**:

- The existing `TestRunPrePush_WithViolations` (`internal/cli/coverage_improvement_test.go:3634`)
  is FALSE-NAMED — its own inline comment states "readStdinLines reads from /dev/stdin, but
  we can't easily mock that... This will fail at stdin read, but covers the convention
  loading path." It never reaches validation or `os.Exit(2)`.
- There is NO subprocess / exit-code harness anywhere in `internal/cli/*_test.go` (verified:
  `grep -rln 'GO_TEST_SUBPROCESS\|exec.Command(os.Args' internal/cli/*_test.go` → empty).
  So a naive "assert exit 2" test would terminate the whole test binary, not assert.

Resolution (the seam): split the DECISION from the EFFECT.

1. `resolvePrePushAction() prePushAction` — PURE. Reads gate + `ActiveModeProfile().Hooks.PrePush`
   (via injected `deps.Config`), normalizes, returns the action enum. No `os.Exit`, no stdin.
2. `decideExit(action prePushAction, violationCount int) int` — PURE. Maps
   (action, violationCount) → intended exit code: `skip`→0, `warn`→0 (even with violations),
   `enforce`+0 violations→0, `enforce`+≥1 violations→2. No `os.Exit`, no stdin.
3. `runPrePush` (thin cobra `RunE` boundary) — the ONLY site that calls
   `os.Exit(decideExit(...))` when the returned code is 2. It also owns gate short-circuit,
   convention load, stdin read, and the print loop.

Every severity-branch DECISION (AC-PMW-002/003/005/006/007) is then asserted against the two
pure functions at the unit level — no process exit, no stdin injection. Only the thin
`os.Exit` translation at the `runPrePush` boundary remains untested-by-unit (acceptable: it
is a one-line mechanical translation with no logic). This seam is two small pure functions,
NOT a refactor of the convention loader — it stays within Tier S scope.

## B. Known Issues (filtered to relevant Tier S categories)

- **B4 Frontmatter canonical schema**: spec.md uses `created:`/`updated:`/`tags:` (verified;
  + `era: V3R6` optional override). Snake_case aliases avoided.
- **B9 Git commit + push (Hybrid Trunk)**: manager-develop commits + pushes within this SPEC
  scope (main直진 per Tier S). Conventional Commits `feat(SPEC-PREPUSH-MODE-WIRING-001): M{N} ...`.
  `--no-verify` forbidden.
- **B10 Scope discipline**: touch ONLY `internal/cli/hook_pre_push.go`,
  `internal/cli/hook_pre_push_test.go`, and (if REQ-PMW-012) `internal/config/envkeys.go`.
  Do NOT touch `validation.go`, `types.go`, `defaults.go`, the shell hook, or templates.
- **B11 AskUserQuestion forbidden**: subagent returns blocker report, never prompts.
- **Coverage**: `internal/cli` is a critical package → target ≥ 90% per repo norm
  (CLAUDE.local.md §6 + go.md). New resolver + branch must be fully covered.

## C. Pre-flight Check List (before code change)

```bash
# 1. Branch + baseline
git branch --show-current        # expect feat/SPEC-PREPUSH-MODE-WIRING-001
git rev-parse HEAD

# 2. Cross-platform build sanity
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Dead-config proof (must be EMPTY before — establishes the "before" state)
grep -rn '\.Hooks\.PrePush\|\.PrePush' internal/ cmd/ --include='*.go' \
  | grep -v _test.go | grep -v validation.go | grep -v types.go | grep -v defaults.go

# 4. Current binary behavior anchor (existing test snapshot)
go test ./internal/cli/ -run 'PrePush' -v 2>&1 | tail -20

# 5. envkeys baseline (confirm no EnvPrePush yet)
grep -n 'EnvEnforceOnPush\|EnvPrePush' internal/config/envkeys.go
```

## D. Constraints (DO NOT VIOLATE)

- PRESERVE: the `runPrePush` validation-and-print loop, `isEnforceOnPushEnabled`,
  `resolveConventionName`, `resolveAutoDetectOptions`, `readStdinLines`, `ActiveModeProfile`,
  `checkStringField`. Reuse, do not rewrite.
- D3 precedence is FIXED: `env(MOAI_ENFORCE_ON_PUSH) > enforce_on_push (gate) > pre_push (severity)`.
  The gate-OFF path MUST short-circuit to `return nil` BEFORE any `pre_push` read
  (REQ-PMW-007) — `pre_push` is never consulted when the gate is off.
- Fail-safe defaults are FIXED: nil ModeProfile → `enforce` (REQ-PMW-010); unknown
  `pre_push` value → `enforce` (REQ-PMW-011).
- Normalization is resolver-side ONLY — do NOT add an enum gate to `validation.go`.
- Forbidden commands: `--no-verify`, `--amend`, force-push to main.
- Required: Conventional Commits + `🗿 MoAI` trailer.

## E. Self-Verification Deliverables (manager-develop reports these)

> **In-process-test barrier (read A.1 first).** `runPrePush` reads `/dev/stdin` and calls
> `os.Exit` — neither is in-process unit-testable. The REQUIRED resolution is the A.1 seam:
> assert the pure `resolvePrePushAction` + `decideExit` functions, NOT process exit or stdin
> injection. The existing `TestRunPrePush_WithViolations` is false-named (never reaches
> validation/`os.Exit`); do NOT model new tests on it. There is no subprocess exit-code
> harness in `internal/cli/*_test.go` and this SPEC does not add one.
>
> **Gate-OFF is the only matrix row reachable by the legacy empty-stdin harness.** When
> `enforce_on_push=false`, `runPrePush` short-circuits to `return nil` before any stdin read or
> `os.Exit` — so it IS directly exercisable in-process (the existing
> `TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately` at `target_coverage_test.go:575`
> proves the short-circuit). Every gate-ON row (skip / warn / enforce) depends on the A.1 seam
> for its decision assertion; only the gate-OFF row is reachable without it.

- **E1** AC binary PASS/FAIL matrix (AC-PMW-001 .. AC-PMW-011 + AC-PMW-013 mandatory; AC-PMW-012
  only if REQ-PMW-012 implemented) — each with verification command + actual output. The
  gate-ON decision ACs (002/003/005/006/007) assert the pure `resolvePrePushAction` /
  `decideExit` helpers per A.1, NOT process exit.
- **E2** Cross-platform build: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3** Coverage: `go test -cover ./internal/cli/...` — report `resolvePrePushAction` +
  `decideExit` + `runPrePush` coverage (target ≥ 90%; the two pure helpers should reach ~100%).
- **E4** Subagent boundary grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/hook_pre_push.go | grep -v '_test.go'` → 0 matches.
- **E5** Lint: `golangci-lint run --timeout=2m` — NEW issues reported separately from baseline.
- **E6** Dead-config-eliminated grep (the headline AC): a runtime reader of `.Hooks.PrePush` now
  EXISTS (`resolvePrePushAction` reads `ActiveModeProfile().Hooks.PrePush`). Tolerate `read -r`
  / shellcheck-clean idioms per the sibling lesson (no while-read in this AC, but the grep
  pattern must match a real symbol).
- **E7** Branch HEAD + push status.

## F. Milestones (priority-ordered, no time estimates)

### M1 — RED: pure-helper decision tests (priority: High)
Write failing tests in `internal/cli/hook_pre_push_test.go` that assert the two PURE helpers
(`resolvePrePushAction` + `decideExit`), NOT process exit or stdin injection (per A.1 seam):
- `resolvePrePushAction`: gate-OFF (returns a sentinel / is not consulted — see A.1), gate-ON ×
  {skip, warn, enforce}, nil-ModeProfile → enforce, unknown-value → enforce. Inject config via
  `deps.Config` consistent with existing pre-push tests.
- `decideExit`: (skip,*)→0, (warn,0)→0, (warn,≥1)→0, (enforce,0)→0, (enforce,≥1)→2.
- Gate-OFF return-path regression: `runPrePush` returns nil immediately when `enforce_on_push=false`
  and `resolvePrePushAction` is not consulted (this row IS reachable by the legacy empty-stdin
  harness — see §E note). Confirm RED (the pure helpers do not yet exist; the warn-exit-0 and
  skip-no-op decisions are not yet expressible).

### M2 — GREEN: pure helpers `resolvePrePushAction()` + `decideExit()` (priority: High)
Add the `prePushAction` typed enum (3 constants) + the two PURE helpers in `hook_pre_push.go`,
mirroring the existing resolver idiom (per A.1 seam — neither calls `os.Exit` nor reads stdin):
- `resolvePrePushAction()`: when gate ON, read `ActiveModeProfile().Hooks.PrePush`, normalize
  (nil-profile → enforce; unknown → enforce), optionally honor `MOAI_PRE_PUSH` (REQ-PMW-012)
  above the config value. (Gate-OFF short-circuit stays in `runPrePush`; the caller does not
  invoke the resolver when the gate is off.)
- `decideExit(action prePushAction, violationCount int) int`: (skip,*)→0, (warn,*)→0,
  (enforce,0)→0, (enforce,≥1)→2.

### M3 — GREEN: `runPrePush` severity branch + thin `os.Exit` boundary (priority: High)
Replace the terminal `os.Exit(2)` (`hook_pre_push.go:101`) with an action-dependent branch that
calls the pure helpers. Order: gate check (`isEnforceOnPushEnabled`) → if OFF `return nil`
(REQ-PMW-007, BEFORE any `pre_push` read) → resolve action → if `skip` `return nil` before
convention load → load convention → validate → `code := decideExit(action, violations)` →
`warn`/clean → print + `return nil` (exit 0); `enforce`+violations → print + `os.Exit(2)`.
`runPrePush` is the ONLY site that calls `os.Exit` — the decision lives in the pure helpers.

### M4 — (conditional) env override (priority: Medium — only if REQ-PMW-012 in scope)
Add `EnvPrePushMode = "MOAI_PRE_PUSH"` to `internal/config/envkeys.go` near `EnvEnforceOnPush`;
wire it into `resolvePrePushAction()` above the config value (gate-independent — never turns
the gate on). Add env-override test. If the Tier S budget is tight, DEFER M4 and omit
AC-PMW-012 (REQ-PMW-012 is SHOULD/optional).

### M5 — REFACTOR + verify (priority: Medium)
Tidy the resolver (table-driven normalization), confirm the validation-and-print loop is
unchanged, run full `go test ./...`, coverage, cross-platform build, lint. Produce E1-E7.

## G. Anti-Patterns to Avoid

- Reading `pre_push` before the gate check (breaks D3 precedence + REQ-PMW-007).
- Adding an enum gate to `validation.go` (out of scope — resolver-side normalize only).
- Defaulting unknown/nil to `warn` instead of `enforce` (fail-OPEN is wrong; fail-safe = `enforce`).
- Rewriting the validation-and-print loop (PRESERVE it; only the terminal exit branches change).
- Letting `MOAI_PRE_PUSH` turn the gate on (it is a severity dial below the gate, not a gate).
- Writing a test that asserts the process exit code by letting `os.Exit(2)` run inline — it
  terminates the whole test binary. Assert `decideExit` (pure) instead (A.1 seam).
- Modeling a new test on the false-named `TestRunPrePush_WithViolations` — it fails at the
  `/dev/stdin` read and never reaches validation or `os.Exit`.
- Putting `os.Exit` inside `resolvePrePushAction` or `decideExit` — both MUST stay pure.

## H. Cross-References

- spec.md §B (REQ-PMW-001..012 incl. REQ-PMW-002a testability seam), §Exclusions (validation.go enum gate deferred).
- acceptance.md (AC-PMW-001..013, full behavior matrix + grep proofs + gate-OFF regression AC-PMW-013).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability (Tier S minimal form).
- `.moai/specs/SPEC-PREPUSH-WIRING-001/` (sibling — `enforce_on_push` wiring precedent).
