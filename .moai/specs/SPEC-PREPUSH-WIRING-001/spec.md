---
id: SPEC-PREPUSH-WIRING-001
title: "Wire dormant pre-push convention engine into the distributed git hook"
version: "0.1.0"
status: draft
created: 2026-06-08
updated: 2026-06-08
author: GOOS
priority: P2
phase: "v0.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "pre-push, git-hook, convention, enforce-on-push, wiring"
---

# SPEC-PREPUSH-WIRING-001 — Wire dormant pre-push convention engine

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-08 | GOOS | Initial draft — plan-phase artifacts. Wire `moai hook pre-push` into the distributed pre-push shell hook, gated by `enforce_on_push`. |

## A. Context (Why)

The distributed pre-push git hook (`internal/template/templates/.git_hooks/pre-push`,
mirrored byte-identically into the Go constant `prePushHookContent` at
`internal/cli/hook_install.go:27-68` and into the dev-repo copy at the project root
`.git_hooks/pre-push`) runs **only** `make ci-local`. It never invokes
`moai hook pre-push`.

Meanwhile a fully-functional Go convention engine exists and is reachable through the
`moai hook pre-push` cobra subcommand (`internal/cli/hook_pre_push.go`). The
`runPrePush` function (line 33) self-gates on `isEnforceOnPushEnabled()` (line 154),
loads the convention engine honoring the auto-detection knobs and `max_length`, reads
commit messages from stdin, validates each, and exits code 2 on violation (line 101).

The engine is therefore **dormant**: even when a user enables `enforce_on_push` (via the
web console or config), no shell path ever pipes commit messages into the subcommand, so
commit-message convention validation never runs on push. The deployed hook and the Go
engine are wired to nothing on the shell side.

This SPEC closes the gap end-to-end by appending a gated `moai hook pre-push` invocation
to the existing pre-push shell hook. The block runs only after `make ci-local` passes and
is a no-op by default (because the engine self-gates on `enforce_on_push`, which ships
`false`). The change is backward-compatible: existing default-config users observe no
behavior change.

This SPEC follows from the web-console-008/009 git_convention redesign cohort, which
exposed `enforce_on_push` to the web console but left the underlying engine unreachable
from the shipped shell hook (the "dormant engine" follow-up).

## B. Requirements (GEARS)

### B.1 Wiring — gated convention invocation

- **REQ-PPW-001** (Ubiquitous): The pre-push hook shall append a commit-message
  convention validation block that invokes `moai hook pre-push`, in addition to the
  existing `make ci-local` step which shall be retained unchanged.

- **REQ-PPW-002** (State-driven): **While** `make ci-local` has failed (the existing
  fail branch is taken and the hook exits before reaching the convention block), the
  hook shall not invoke `moai hook pre-push`. The convention block shall be placed after
  the `ci-local` fail-check exit and before the final `exit 0`.

- **REQ-PPW-003** (Capability gate): **Where** the `moai` executable is not resolvable
  on `PATH`, the hook shall skip the convention block and complete normally, so that
  projects without `moai` installed are unaffected (16-language template neutrality).

- **REQ-PPW-004** (Capability gate): **Where** `enforce_on_push` is disabled (the shipped
  default), the `moai hook pre-push` invocation shall be a no-op (the engine self-gates via
  `isEnforceOnPushEnabled()` and returns without validating), so existing default-config
  users observe no behavior change.

### B.2 git stdin protocol translation

- **REQ-PPW-005** (Event-driven): **When** git invokes the pre-push hook, git passes
  ref-update lines on stdin in the form `<local ref> <local sha> <remote ref> <remote sha>`,
  NOT commit messages. The hook shall translate these ref-update lines into commit messages
  by computing the commit range per ref and piping `git log --format=%s <range>` output to
  `moai hook pre-push`.

- **REQ-PPW-006** (Event-driven): **When** a ref-update line reports a remote SHA of all
  zeros (new branch / new remote ref with no remote ancestor), the hook shall use a
  range that enumerates the locally-new commits being pushed rather than attempting a
  two-dot diff against the zero SHA.

- **REQ-PPW-007** (Event-driven): **When** a ref-update line reports a local SHA of all
  zeros (branch deletion — nothing pushed for that ref), the hook shall skip that ref and
  not run convention validation for it.

- **REQ-PPW-008** (Ubiquitous): The hook shall preserve git's ref-update stdin so it is
  available to the convention block. Because `make ci-local` precedes the convention block,
  the hook shall ensure stdin is captured into a variable (or otherwise made available)
  before any step that could consume it.

### B.3 Mirror parity

- **REQ-PPW-009** (Ubiquitous): The three pre-push hook copies — the distributed template
  (`internal/template/templates/.git_hooks/pre-push`), the Go constant `prePushHookContent`
  (`internal/cli/hook_install.go`), and the dev-repo root copy (`.git_hooks/pre-push`) —
  shall remain byte-identical after the change. The template ↔ constant parity is enforced
  by `TestPrePushTemplateMatchesConstant`; the root copy is hand-maintained and shall be
  updated in the same change.

- **REQ-PPW-010** (Ubiquitous): The `embedded.go` generated artifact shall be regenerated
  via `make build` after the template edit (it is generated, never hand-edited).

### B.4 Install assertion

- **REQ-PPW-011** (Event-driven): **When** the pre-push hook is installed into a fresh
  repository, the installed hook content shall contain the `moai hook pre-push` invocation
  token in addition to the existing `ci-local` and MoAI-marker tokens.

## C. Template Neutrality Constraints (HARD)

The appended block goes into a distributed template consumed by all 16 supported
languages. Per CLAUDE.local.md §15 / §25 it shall contain NO internal SPEC IDs, REQ
tokens, audit citations, internal dates, commit SHAs, macOS-bias paths, or CLAUDE.local
references — generic POSIX shell only. The `command -v moai` guard (REQ-PPW-003) keeps
the block neutral for non-moai projects.

## D. Acceptance Criteria

Enumerated in `acceptance.md`. AC count: 9 (AC-PPW-001 .. AC-PPW-009).

## Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly OUT OF SCOPE for this SPEC and are recorded as known-deferred
follow-up candidates:

- **`git_strategy.hooks.pre_push` runtime reader** — The nested git-strategy config field
   `git_strategy.{manual,personal,team}.hooks.pre_push` (values `warn` / `enforce` / `skip`)
   is validation-only (`internal/config/validation.go:264/272/280`) with default `"warn"`
   (`internal/config/defaults.go:249/261/276`) and **zero runtime consumers**. Wiring a
   runtime reader for `ActiveModeProfile().Hooks.PrePush` is a separate follow-up SPEC and
   is NOT addressed here. This SPEC wires ONLY `enforce_on_push` → `moai hook pre-push`.

- **Flipping the `enforce_on_push` template default to `true`** — The template config
   (`internal/template/templates/.moai/config/sections/git-convention.yaml`,
   `enforce_on_push: false`) stays `false` (opt-in). This SPEC does NOT change the default;
   the wired hook is a no-op for all projects until the user enables enforcement.

- **Any change to `runPrePush` Go logic** — The convention engine, `isEnforceOnPushEnabled`,
   `readStdinLines`, `resolveConventionName`, `resolveAutoDetectOptions`, and the
   `moai hook pre-push` subcommand registration are complete and correct. This SPEC adds
   only the shell-side invocation; it does NOT modify the Go engine.

- **New convention validators or convention definitions** — No new convention parsing,
   detection, or validation logic. The existing engine behavior is reused verbatim.

- **Pre-commit hook wiring** — Only the pre-push hook is in scope. Pre-commit / commit-msg
   hooks are untouched.

## Cross-References

- `internal/cli/hook_pre_push.go` — complete Go convention engine (DO NOT MODIFY)
- `internal/cli/hook_install.go` — `prePushHookContent` constant (mirror leg 2)
- `internal/template/templates/.git_hooks/pre-push` — distributed template (mirror leg 1)
- `.git_hooks/pre-push` — dev-repo root copy (mirror leg 3, hand-maintained)
- `internal/cli/wave1_sync_test.go` — `TestPrePushTemplateMatchesConstant` byte-parity gate
- `internal/cli/hook_install_test.go` — `TestInstallPrePushHook_FreshRepo` install assertion
- `internal/template/templates/.moai/config/sections/git-convention.yaml` — `enforce_on_push` default
- CLAUDE.local.md §2 (Template-First Rule), §15 / §25 (Template neutrality / internal-content isolation)
