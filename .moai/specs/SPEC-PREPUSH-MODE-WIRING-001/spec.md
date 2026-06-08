---
id: SPEC-PREPUSH-MODE-WIRING-001
title: "Wire dormant git_strategy.<mode>.hooks.pre_push severity into the pre-push runtime"
version: "0.2.0"
status: implemented
created: 2026-06-08
updated: 2026-06-08
author: GOOS
priority: P2
phase: "v0.2.0"
module: "internal/cli, internal/config"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "pre-push, git-strategy, mode-profile, severity, dead-config, wiring"
---

# SPEC-PREPUSH-MODE-WIRING-001 — Wire dormant pre-push mode severity dial

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-08 | GOOS | Initial draft — plan-phase artifacts. Wire `git_strategy.{manual,personal,team}.hooks.pre_push` (warn / enforce / skip) into `runPrePush` as a severity dial, gated by the `enforce_on_push` master gate. 2nd dead-config follow-up to SPEC-PREPUSH-WIRING-001. |

## A. Context (Why)

This SPEC is the **2nd dead-config follow-up** to the just-closed
**SPEC-PREPUSH-WIRING-001**, which wired `git_convention.validation.enforce_on_push`
end-to-end (the shell hook now pipes commit messages into `moai hook pre-push`).
That sibling SPEC explicitly recorded — in its own `Exclusions` section — that the
nested `git_strategy.{manual,personal,team}.hooks.pre_push` field remained a separate
dead config to be addressed by a follow-up. This SPEC is that follow-up.

### A.1 The dead config (ground-truth verified)

`git_strategy.{manual,personal,team}.hooks.pre_push` is a `string` field
(`HooksConfig.PrePush`, `internal/config/types.go:92`). Its template default is `warn`
(`internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:34/66/104`);
the `enforce | warn | skip` value vocabulary is documented on the sibling `pre_commit`
comment line (`git-strategy.yaml.tmpl:33/65/103` — `pre_commit: enforce  # enforce, warn, skip`),
NOT on the `pre_push` line itself. The compiled-in default is `warn` for all three modes
(`internal/config/defaults.go:249,261,276`).

The field is **validated only for dynamic-token leakage** by `checkStringField`
(`internal/config/validation.go:264,272,280` → `checkStringField` body at
`validation.go:296`). There is **NO enum validation** — an arbitrary string such as
`"garbage"` passes validation untouched.

Crucially the field has **ZERO runtime consumers**. Verified:

```bash
grep -rn '\.Hooks\.PrePush\|\.PrePush' internal/ cmd/ --include='*.go' \
  | grep -v _test.go | grep -v validation.go | grep -v types.go | grep -v defaults.go
# (empty — no runtime reader)
```

So the per-mode pre-push policy is a **dead config**: a user who sets
`git_strategy.team.hooks.pre_push: skip` observes no behavioral effect whatsoever.

### A.2 The current pre-push runtime (the wiring target)

`runPrePush` (`internal/cli/hook_pre_push.go:33-103`) is the `moai hook pre-push`
subcommand body. Today its behavior is **binary**:

- `isEnforceOnPushEnabled()` (`hook_pre_push.go:154-167`) reads
  `MOAI_ENFORCE_ON_PUSH` env > `cfg.GitConvention.Validation.EnforceOnPush` (bool) >
  default `false`. When the result is `false` → `return nil` (no-op).
- When the gate is enabled, `runPrePush` loads the convention (honoring
  `resolveConventionName()` and `resolveAutoDetectOptions()`), reads commit subjects
  from stdin, validates each, and on **any** violation does `os.Exit(2)` (deny)
  (`hook_pre_push.go:101`). On clean → prints OK and returns nil.

The result: gate ON ⇒ unconditionally blocking (exit 2 on violation); gate OFF ⇒ no-op.
There is no middle ground (warn-but-allow).

### A.3 The accessor and idiom that this SPEC reuses

- `cfg.GitStrategy.ActiveModeProfile() (*ModeProfile, bool)`
  (`internal/config/types.go:165-176`) switches on `cfg.GitStrategy.Mode` ∈
  {`manual`, `personal`, `team`}, returns `(nil, false)` when Mode is empty/invalid.
  `ModeProfile.Hooks.PrePush` is the target value.
- The per-concern resolver-function idiom already established in `hook_pre_push.go`:
  `resolveConventionName()`, `isEnforceOnPushEnabled()`, `resolveAutoDetectOptions()` —
  each a small pure-ish function. This SPEC adds a fourth one in the same shape.
- Env-override idiom: env constants live in `internal/config/envkeys.go`
  (`EnvEnforceOnPush = "MOAI_ENFORCE_ON_PUSH"` at line 50,
  `EnvGitConvention = "MOAI_GIT_CONVENTION"` at line 47). There is currently **no**
  `EnvPrePush` / `MOAI_PRE_PUSH` constant.

### A.4 Why no live regression in this project

The local project ships `git_strategy.mode: team` with
`team.hooks.pre_push: enforce`, and `git_convention.validation.enforce_on_push: false`
(the master gate is OFF). Because this SPEC keeps `enforce_on_push` as the master gate
(see §B.3 D3), the OFF gate means `pre_push` is never consulted in this project — so
the change is a no-op for the local project and there is **zero live regression**. The
behavior change is scoped entirely to the gate-ON population (see REQ-PMW-009 / the
documented on-case transition).

## B. Requirements (GEARS)

The canonical 3-state action enum is **`skip` / `warn` / `enforce`** (D2). Throughout
this SPEC, `<resolved action>` refers to the output of the new resolver
`resolvePrePushAction()` (D1).

### B.1 Resolver — new dedicated helper (D1)

- **REQ-PMW-001** (Ubiquitous): The pre-push runtime shall introduce a new dedicated
  resolver helper `resolvePrePushAction()` in `internal/cli/hook_pre_push.go`, mirroring
  the existing `resolveConventionName()` / `isEnforceOnPushEnabled()` /
  `resolveAutoDetectOptions()` resolver idiom. It shall return a small typed enum value
  (named `prePushAction`) with three constants representing `skip` / `warn` / `enforce`.

- **REQ-PMW-002** (Ubiquitous): `runPrePush` shall branch on the resolved
  `prePushAction` value to choose between the three behaviors of D2, rather than the
  current binary exit-2-on-violation behavior.

- **REQ-PMW-002a** (Ubiquitous, testability seam): The exit-code DECISION shall be
  extracted into a pure helper `decideExit(action prePushAction, violationCount int) int`
  that returns the intended process exit code (`0` for skip / warn / clean-enforce; `2`
  for enforce-with-violations) WITHOUT calling `os.Exit`. The thin cobra `RunE` boundary
  (`runPrePush`) shall be the ONLY site that translates a non-zero `decideExit` result
  into `os.Exit(2)`. Rationale: `runPrePush` reads stdin from `/dev/stdin` and terminates
  the test binary via `os.Exit`, so the branch decision is NOT in-process unit-testable
  inline. Keeping `resolvePrePushAction()` (pure) and `decideExit()` (pure) free of
  `os.Exit` and stdin makes every severity-branch decision unit-testable. This is a small
  seam extraction (two pure functions), NOT a refactor of the convention loader, and stays
  within Tier S scope.

### B.2 Three-state severity semantics (D2)

- **REQ-PMW-003** (State-driven): **While** the resolved action is `skip`, the
  pre-push runtime shall perform an immediate no-op — `return nil`, no convention load,
  no validation, no output.

- **REQ-PMW-004** (State-driven): **While** the resolved action is `warn`, the
  pre-push runtime shall run the full convention validation and print any violations,
  but shall complete with **exit 0** (the push is allowed). This is a new non-blocking
  middle state.

- **REQ-PMW-005** (State-driven): **While** the resolved action is `enforce`, the
  pre-push runtime shall run the full convention validation, print any violations, and
  on **any** violation exit with code 2 (the push is blocked) — identical to the
  current `runPrePush` blocking behavior.

- **REQ-PMW-006** (Event-driven): **When** the resolved action is `warn` or `enforce`
  and all commit subjects are clean (no violations), the pre-push runtime shall print
  the existing OK message and complete with exit 0.

### B.3 Precedence — master gate then severity dial (D3)

- **REQ-PMW-007** (Capability gate): **Where** `enforce_on_push` is disabled (the
  shipped default, resolved by the existing `isEnforceOnPushEnabled()` with its
  `MOAI_ENFORCE_ON_PUSH` env override), the pre-push runtime shall remain a no-op and
  shall **never consult** `pre_push`. This preserves SPEC-PREPUSH-WIRING-001 OFF
  behavior 100% (zero behavior change in the gate-OFF case).

- **REQ-PMW-008** (Capability gate): **Where** `enforce_on_push` is enabled, the
  pre-push runtime shall consult the `pre_push` of the active mode profile
  (`ActiveModeProfile().Hooks.PrePush`) to select the severity (`skip` / `warn` /
  `enforce`) per D2. The resolution order is therefore:
  `env(MOAI_ENFORCE_ON_PUSH) > enforce_on_push (gate) > pre_push (severity)`.

- **REQ-PMW-009** (Ubiquitous): The on-case severity transition shall be documented as
  an intentional, accepted behavior change: for any config with `enforce_on_push=true`
  AND active-mode `pre_push=warn` (the template default), the pre-push behavior is now
  **non-blocking (exit 0)** — a downgrade from the pre-change unconditional exit-2. The
  local project (`enforce_on_push=false`) is not affected, so there is no live
  regression; the change applies only to the gate-ON population and is accepted per D3
  (the user accepted "on일 때 pre_push가 warn vs enforce 선택").

### B.4 Edge-case normalization (fail-safe)

- **REQ-PMW-010** (Event-driven): **When** the gate is ON but `ActiveModeProfile()`
  returns `(nil, false)` (Mode empty or invalid — no per-mode severity readable), the
  resolver shall default the severity to `enforce` (fail-safe toward the historical
  blocking behavior, since the gate was explicitly turned on).

- **REQ-PMW-011** (Event-driven): **When** the gate is ON and the active-mode
  `pre_push` value is unrecognized (NOT one of `skip` / `warn` / `enforce` — recall
  `checkStringField` does NOT enum-validate it, so `"garbage"` reaches the resolver),
  the resolver shall normalize the value to `enforce` (fail-safe). This normalization
  is performed **resolver-side**; this SPEC does NOT add an enum gate to
  `validation.go` (see Exclusions).

### B.5 Optional severity env override (SHOULD — only if Tier S budget permits)

- **REQ-PMW-012** (Capability gate, SHOULD): **Where** the new env var
  `MOAI_PRE_PUSH` (constant `EnvPrePushMode`, added near `EnvEnforceOnPush` in
  `internal/config/envkeys.go`) is set to one of `skip` / `warn` / `enforce`, the
  resolver shall use that value as the severity in preference to the config
  `pre_push`, when the gate is ON. The env severity override sits **below** the gate:
  `MOAI_PRE_PUSH` does NOT turn the gate on; the gate remains governed solely by
  `enforce_on_push` / `MOAI_ENFORCE_ON_PUSH`. An unrecognized `MOAI_PRE_PUSH` value is
  ignored (falls through to the config `pre_push`, then REQ-PMW-011 normalization).

  This requirement is OPTIONAL: include it only if it fits the Tier S budget cleanly.
  If omitted, REQ-PMW-007/008 still fully define the precedence model
  (env-gate > gate > config-severity) without an env severity layer.

## C. Acceptance Criteria

Enumerated in `acceptance.md`. AC count: 13 (AC-PMW-001 .. AC-PMW-013), of which
AC-PMW-012 is conditional on the OPTIONAL REQ-PMW-012 (MOAI_PRE_PUSH env override).
AC-PMW-002/003/005/006/007 assert the pure-helper decision (`resolvePrePushAction` +
`decideExit`) per the REQ-PMW-002a testability seam, NOT process exit or stdin injection.
AC-PMW-013 pins the gate-OFF return-path regression (predecessor byte-preserved behavior).

## Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly OUT OF SCOPE for this SPEC and are recorded as
known-deferred or deliberately-rejected items:

- **`enum gate in validation.go`** — This SPEC does NOT add an enum-validation rule to
  `checkStringField` (or any new `validation.go` rule) for the `pre_push` field. Adding
  an enum gate touches the shared validation path and risks cascade. The resolver-side
  normalization of REQ-PMW-011 (unknown → `enforce`) is sufficient for Tier S and keeps
  the change bounded to `internal/cli`. A `validation.go` enum gate is recorded here as
  a SHOULD-level optional follow-up SPEC, NOT addressed here.

- **`pre_commit` / `commit_msg` hook severity wiring** — Only `pre_push` is in scope.
  The sibling `HooksConfig.PreCommit` and `HooksConfig.CommitMsg` fields remain dead
  config and are NOT wired by this SPEC.

- **Flipping the `enforce_on_push` template default to `true`** — The master gate stays
  `false` (opt-in) in the template
  (`internal/template/templates/.moai/config/sections/git-convention.yaml`). This SPEC
  changes ONLY the severity behavior when the gate is already ON; it does not change the
  shipped default.

- **Changing the convention engine, stdin protocol, or shell hook** — The convention
  engine (`convention.Manager`), `resolveConventionName`, `resolveAutoDetectOptions`,
  `readStdinLines`, and the shell-side `moai hook pre-push` invocation wired by
  SPEC-PREPUSH-WIRING-001 are complete and correct. This SPEC adds only the severity
  branch inside `runPrePush` plus the new resolver; it does NOT touch the shell hook,
  the stdin protocol, or the convention loader.

- **Broad refactor of `runPrePush`** — The validation-and-print loop
  (`hook_pre_push.go:72-98`) is reused as-is. The structural changes are bounded to:
  (a) replacing the single terminal `os.Exit(2)` with an action-dependent branch
  (skip-return / warn-exit-0 / enforce-exit-2), and (b) extracting the two pure helpers
  `resolvePrePushAction()` + `decideExit()` (REQ-PMW-002a testability seam). This is NOT a
  broad refactor of the convention loader, stdin reader, or any other `runPrePush` logic —
  no new config sections, no convention-loader refactor.

- **Web console surface for `pre_push`** — Exposing the `git_strategy.<mode>.hooks.pre_push`
  field in the `moai web` console (the web-console-009 cohort touched `git_convention`,
  not `git_strategy.hooks`) is NOT in scope. This SPEC is a backend runtime-reader wiring
  only.

## Cross-References

- `internal/cli/hook_pre_push.go` — `runPrePush` (wiring target + thin `os.Exit` boundary,
  line 33), existing resolver idiom (`resolveConventionName` :107, `isEnforceOnPushEnabled`
  :154, `resolveAutoDetectOptions` :127), terminal `os.Exit(2)` (:101); NEW pure helpers
  `resolvePrePushAction()` + `decideExit()` land here (REQ-PMW-002a seam)
- `internal/config/types.go` — `HooksConfig` struct (:90), `HooksConfig.PrePush` field (:92), `ActiveModeProfile()` (:165)
- `internal/config/defaults.go` — per-mode `pre_push: warn` defaults (:249/:261/:276)
- `internal/config/validation.go` — `checkStringField` dynamic-token-only validation
  (:264/:272/:280 call sites, :296 body) — NO enum gate (deliberately preserved)
- `internal/config/envkeys.go` — `EnvEnforceOnPush` (:50), `EnvGitConvention` (:47);
  new `EnvPrePushMode` lands here if REQ-PMW-012 is included
- `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` — `pre_push`
  default `warn` (:34/:66/:104); the `enforce | warn | skip` value vocabulary lives on the
  sibling `pre_commit` comment line (:33/:65/:103), NOT on the `pre_push` line
- `.moai/specs/SPEC-PREPUSH-WIRING-001/` — sibling/predecessor SPEC (1st dead-config
  follow-up: `enforce_on_push` wiring); its Exclusions section records this SPEC's scope
  as a deferred follow-up
- CLAUDE.local.md §14 (Hardcoding prevention — env const idiom), §6 (Test isolation)
