---
id: SPEC-PRECOMMIT-001
title: "Deployment-layer fast-subset pre-commit git hook installer"
version: "0.1.0"
status: in-progress
created: 2026-07-05
updated: 2026-07-05
author: GOOS행님
priority: P2
phase: "v0.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "pre-commit, git-hook, installer, fast-subset, gofmt, go-vet, template-first"
tier: S
---

# SPEC-PRECOMMIT-001 — Deployment-layer fast-subset pre-commit git hook installer

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-05 | GOOS행님 | Initial draft — plan-phase artifacts. Symmetric extension of the pre-push installer: add a fast-subset (`gofmt -l` + `go vet`) pre-commit hook installer, wired into `moai init` / `moai update`, with template ↔ const byte-identity parity. |

## A. Context (Why)

The existing pre-push git hook installer (`internal/cli/hook_install.go`) installs a
`#!/bin/sh` hook into `.git/hooks/pre-push` that runs the full local CI mirror
(`make ci-local` → `scripts/ci-mirror/run.sh`: lint + vet + test + cross-compile).
The installer is marker-based (overwrites its own hook, preserves foreign hooks via
`ErrUserHookExists`), non-fatal on failure, and is wired into both `moai init`
(`init.go:473`) and `moai update` (`update.go:886`) via `installPrePushHookOptional`.
The hook body is a Go constant (`prePushHookContent`, `hook_install.go:27-98`) kept
**byte-identical** with the distributed template
`internal/template/templates/.git_hooks/pre-push`, enforced by
`TestPrePushTemplateMatchesConstant` (`internal/cli/wave1_sync_test.go:19`, REQ-CIAUT-001).

There is currently **no commit-tier gate**. The `git_strategy.{manual,personal,team}.hooks.pre_commit`
config field ships `enforce` in all three mode blocks
(`git-strategy.yaml:22/47/79`, `defaults.go:277/290/306`, `types.go:92`) but is
**DEAD**: `validation.go` only value-checks the string; **no installer consumes it**.
The declared "enforce pre-commit" intent has no mechanism behind it.

This SPEC closes that gap by adding a **symmetric** pre-commit installer that mirrors
the pre-push installer's exact structure, and a fast-subset pre-commit hook. The
design is deliberately **2-tier**:

- **commit tier (this SPEC — fast):** on staged Go files, run `gofmt -l` (format
  check) + `go vet` (affected packages). Sub-second, no heavy suite.
- **push tier (existing — full):** `make ci-local` (lint + vet + test + cross-compile).

Heavy lint+test is intentionally EXCLUDED from the commit tier — it is already
covered by the pre-push hook. This gives users fast feedback at `git commit` time
(catch un-`gofmt`'d code and obvious `go vet` errors before they accumulate) without
paying the full-CI cost on every commit.

Scope: this is a **deployment-layer** feature — it ships to ALL moai users via the
template tree (Template-First per CLAUDE.local.md §2), NOT a local-dev-only hook.

## B. Requirements (GEARS)

### B.1 Installer parity (mirror `PrePushInstaller`)

- **REQ-PC-001** (Ubiquitous): moai shall install a pre-commit git hook into
  `.git/hooks/pre-commit` with mode `0755` during `moai init` and `moai update`,
  mirroring the pre-push installer's marker-based semantics via a new
  `PreCommitInstaller` type and an `installPreCommitHookOptional` helper.

- **REQ-PC-002** (Event-driven): **When** no `.git/hooks/pre-commit` file exists,
  the installer shall create it with the canonical MoAI pre-commit content.

- **REQ-PC-003** (Event-driven): **When** a pre-commit hook exists AND carries the
  `# MoAI-ADK pre-commit hook` marker within its first 3 lines, the installer shall
  overwrite it safely (mirror of `fileHasMoaiMarker`).

- **REQ-PC-004** (Unwanted behavior): The installer shall not overwrite a
  pre-existing pre-commit hook that lacks the MoAI marker in its first 3 lines;
  it shall preserve the file unchanged and return `ErrUserHookExists` (the shared
  sentinel is reused), so a user's own pre-commit hook is never clobbered.

- **REQ-PC-005** (State-driven): **While** hook installation fails for any reason
  other than a foreign-hook preserve, `moai init` / `moai update` shall not abort —
  the failure shall be surfaced as a non-fatal warning to the command's output
  writer (mirror of `installPrePushHookOptional`).

- **REQ-PC-006** (Capability gate): **Where** the `--no-hooks` flag is supplied to
  `moai init` / `moai update` (`getBoolFlag(cmd, "no-hooks")`), the pre-commit
  installer shall be a no-op (`skip=true`), mirroring the pre-push wiring.

### B.2 Hook content — fast subset (2-tier design)

- **REQ-PC-007** (Ubiquitous): The pre-commit hook shall be a `#!/bin/sh` POSIX
  script carrying the `# MoAI-ADK pre-commit hook` marker in its first lines, so it
  runs cross-platform (git-bash / wsl) and is safely re-installable.

- **REQ-PC-008** (Event-driven): **When** git invokes the pre-commit hook, the hook
  shall collect the staged Go files via
  `git diff --cached --name-only --diff-filter=ACM` filtered to the `.go` suffix.

- **REQ-PC-009** (State-driven): **While** the staged set contains no `.go` files,
  the hook shall exit 0 without running any check (fast no-op for non-Go or
  docs-only commits).

- **REQ-PC-010a** (Event-driven): **When** the staged Go set is non-empty, the hook
  shall run `gofmt -l` on the staged Go files.

- **REQ-PC-010b** (Event-driven): **When** `gofmt -l` reports one or more staged Go
  files needing formatting, the hook shall print a remediation hint and exit 1
  (commit blocked).

- **REQ-PC-011a** (Event-driven): **When** the staged Go set is non-empty and the
  `gofmt -l` check reported nothing, the hook shall run `go vet` on the affected
  packages.

- **REQ-PC-011b** (Event-driven): **When** `go vet` exits non-zero on the affected
  packages, the hook shall print a remediation hint and exit 1 (commit blocked).

- **REQ-PC-012** (Capability gate): **Where** the environment variable
  `SKIP_MOAI_PRECOMMIT=1` is set, the hook shall print a bypass notice to stderr and
  exit 0 without running any check (mirror of the pre-push `SKIP_MOAI_PREPUSH`
  pattern).

- **REQ-PC-013** (Capability gate): **Where** `gofmt` or `go` is not resolvable on
  `PATH`, the hook shall skip the corresponding check and complete normally, so that
  projects without a Go toolchain are unaffected (16-language template neutrality).

- **REQ-PC-014** (Unwanted behavior): The pre-commit hook shall not invoke
  `make ci-local`, `golangci-lint`, or any full test suite — the commit tier is the
  fast subset only; heavy lint+test remains the pre-push hook's responsibility. This
  is the 2-tier commit=fast / push=full-ci boundary.

### B.3 Byte-identity / Template-First

- **REQ-PC-015** (Ubiquitous): The distributed template
  `internal/template/templates/.git_hooks/pre-commit` shall be byte-identical with a
  `preCommitHookContent` constant in `internal/cli/hook_install.go`, enforced by a
  new CI byte-identity test that mirrors `TestPrePushTemplateMatchesConstant`
  (walk-to-`go.mod`-root, read template, compare with the constant, graceful skip
  when the template is absent).

- **REQ-PC-016** (Ubiquitous): The template source shall be edited first and the
  embedded FS regenerated via `make build`; the byte-identity invariant of REQ-PC-015
  shall hold after the build (Template-First Rule, CLAUDE.local.md §2).

### B.4 Config reconciliation (chosen approach — explicit)

- **REQ-PC-017** (Ubiquitous): The pre-commit installer shall ALWAYS install the hook
  regardless of the `git_strategy.{manual,personal,team}.hooks.pre_commit` config
  value — it shall be **config-agnostic**, mirroring the pre-push *installer*
  (`hook_install.go` `InstallPrePushHook`, which is verifiably config-agnostic: it
  never reads a config field). Materializing the currently-dead `pre_commit` config's
  declared intent is achieved by making a real hook EXIST and run the fast subset at
  commit time. A future pre-commit **runtime severity dial** (applying `pre_commit`'s
  `skip` / `warn` / `enforce` at hook-run time, install-always) is a separate follow-up
  (see Exclusions); this SPEC does not implement it.

  Rationale: (a) symmetry with the proven pre-push installer keeps the change mechanical
  and low-risk; (b) **only `pre_commit` is a genuinely dead field** — its sole consumers
  are the `validation.go` `checkStringField` value-checks
  (`internal/config/validation.go:363/371/379`), with no runtime reader (verified by
  grep this session). By contrast the sibling `pre_push` field is **NOT dead**:
  `resolvePrePushAction()` (`internal/cli/hook_pre_push.go:60`) reads
  `ActiveModeProfile().Hooks.PrePush` (`hook_pre_push.go:71-72`) via `parsePrePushAction`
  (`hook_pre_push.go:88`) as a live **runtime severity dial** (`skip` / `warn` /
  `enforce`, fail-safe to `enforce`) applied at hook-run time — a mechanism DISTINCT
  from install-time gating. The config-agnostic INSTALL decision therefore rests on the
  installer-vs-runtime-action distinction (the pre-push *installer* is config-agnostic;
  the pre-push *runtime action* is config-driven), NOT on any claim that `pre_push` is
  unused.

### B.5 Install assertion

- **REQ-PC-018** (Event-driven): **When** the pre-commit hook is installed into a
  fresh repository, the installed hook content shall contain all four tokens: the
  `# MoAI-ADK pre-commit hook` marker, the `gofmt -l` invocation token, the `go vet`
  invocation token, and the `SKIP_MOAI_PRECOMMIT` bypass token (mirror of the
  pre-push fresh-repo install assertion).

### B.6 Cross-platform build

- **REQ-PC-019** (Ubiquitous): The new installer code shall compile cleanly under
  `GOOS=windows GOARCH=amd64 go build ./...`. It shall use only the same portable
  primitives as `PrePushInstaller` (`os`, `bufio`, `filepath`, `strings`, `errors`,
  `io`) with no OS-specific syscalls, so no build-tag file split is required.

## C. Acceptance Criteria

Enumerated in `acceptance.md`. AC count: 15 (AC-PC-001 .. AC-PC-015).

## D. Constraints (HARD)

- **Template neutrality (CLAUDE.local.md §15 / §25):** the distributed pre-commit
  template is consumed by all 16 supported languages. It shall contain NO internal
  SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs, macOS-bias
  paths, or CLAUDE.local references — generic POSIX shell only. The `command -v go` /
  `command -v gofmt` guards (REQ-PC-013) keep the hook neutral for non-Go projects.
- **Structural mirror:** the new installer MUST mirror the existing `PrePushInstaller`
  structure (marker-in-first-3-lines, `ErrUserHookExists` reuse, `0755`, non-fatal
  optional wrapper). It MUST NOT re-architect the shared helpers.
- **No pre-push regression:** the pre-push installer, its constant, its template, and
  `TestPrePushTemplateMatchesConstant` MUST remain unchanged and passing.

## Exclusions (What NOT to Build)

The following are explicitly out of scope for this SPEC and are recorded as
known-deferred follow-up candidates.

### Out of Scope — pre_commit runtime severity dial (hook-run-time, NOT an install gate)

- A future pre-commit **runtime severity dial** that reads
  `ActiveModeProfile().Hooks.PreCommit` (`skip` / `warn` / `enforce`) and applies it at
  **hook-run time** — mirroring the live `resolvePrePushAction()`
  (`internal/cli/hook_pre_push.go:60`) that already does this for `pre_push`. This SPEC
  installs the hook config-agnostically at install time (REQ-PC-017) and runs the fast
  subset unconditionally; wiring `pre_commit` into a runtime severity dial (so `skip`
  silences the check, `warn` reports without blocking, `enforce` blocks) is deferred.
  Note: the sibling `pre_push` field already HAS such a runtime dial, so this follow-up
  brings pre-commit to parity — it is NOT a co-wiring of a shared dead field (only
  `pre_commit` is dead; `pre_push` is live).

### Out of Scope — heavy lint / test in the pre-commit tier

- `make ci-local`, `golangci-lint`, and the full `go test ./...` suite stay in the
  pre-push hook. The commit tier runs only the fast subset (`gofmt -l` + `go vet` on
  staged Go files). Adding heavy checks to pre-commit would defeat the 2-tier design.

### Out of Scope — commit-msg and other hooks

- Only the pre-commit hook is added. The commit-msg convention validator, the pre-push
  hook, and any other git hook are untouched.

### Out of Scope — flipping any config default

- No change to the `pre_commit: enforce` value in `git-strategy.yaml` / `defaults.go`,
  and no change to any other config default. The dead field's literal value is left as-is.

### Out of Scope — non-Go language format / vet checks

- The fast subset targets staged Go files only (`gofmt -l` + `go vet`). Wiring
  per-language formatters/linters (prettier, ruff, rustfmt, etc.) for other staged
  file types is not included; non-Go staged files are ignored by the hook.

### Out of Scope — dev-repo root .git_hooks/pre-commit dogfood copy

- The pre-push line maintains a hand-edited dev-repo root copy (`.git_hooks/pre-push`,
  mirror leg 3). This SPEC's mandatory scope is the template (leg 1) + the Go constant
  (leg 2) + the CI byte-identity gate. Adding a self-dogfooding root `.git_hooks/pre-commit`
  copy for this maintainer repo is optional maintainer convenience, not a required leg,
  and is left to the run phase's discretion (not an AC).

## Cross-References

- `internal/cli/hook_install.go` — `PrePushInstaller`, `prePushHookContent`,
  `ErrUserHookExists`, `installPrePushHookOptional`, `fileHasMoaiMarker` (mirror source)
- `internal/template/templates/.git_hooks/pre-push` — distributed pre-push template (mirror pattern)
- `internal/cli/wave1_sync_test.go` — `TestPrePushTemplateMatchesConstant` (byte-identity gate to mirror)
- `internal/cli/hook_install_test.go` — pre-push installer tests (mirror source for installer tests)
- `internal/cli/init.go:473`, `internal/cli/update.go:886` — `installPrePushHookOptional` call sites (wiring pattern)
- `internal/config/{defaults,types,validation}.go` + `.moai/config/sections/git-strategy.yaml` — dead `pre_commit` config (context; NOT consumed by this SPEC per REQ-PC-017)
- `Makefile:83` (`ci-local`) + `scripts/ci-mirror/run.sh` — the full-CI push tier the commit tier deliberately does NOT duplicate
- CLAUDE.local.md §2 (Template-First Rule), §15 / §25 (Template neutrality / internal-content isolation)
