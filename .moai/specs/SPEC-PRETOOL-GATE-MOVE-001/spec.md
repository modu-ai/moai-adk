---
id: SPEC-PRETOOL-GATE-MOVE-001
title: "Relocate commit-quality gate off PreToolUse 5s budget to native git pre-commit hook"
version: "0.1.0"
status: draft
created: 2026-07-28
updated: 2026-07-28
author: manager-spec
priority: P0
phase: "v3.0.x"
module: "internal/hook/quality, internal/cli/hook_install, internal/template/templates"
lifecycle: spec-anchored
tier: M
tags: "pretool, gate, precommit, hook, census-p1b, critical"
---

## HISTORY

| Version | Date       | Author        | Change                                                                                                                                                                                                                                                                                                                                              |
|---------|------------|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-07-28 | manager-spec  | Initial draft — plan-phase artifacts (Tier M). Census C-2 (CRITICAL) fix: relocate heavy vet/lint/test gate off the PreToolUse 5s hook budget to a native git pre-commit hook. User-approved direction (e). Extends the deployment-layer `PreCommitInstaller` shipped by SPEC-PRECOMMIT-001; PRESERVEs the ast-grep scanner tuned by SPEC-FALSE-ALLCLEAR-GUARD-001 (PR #1183, the worktree base). |

## A. Context (Why)

### A.1 The defect (census C-2, line 107)

`internal/hook/quality/gate.go:254` `QualityGate.Run(ctx)` runs vet/lint/test **synchronously** and is **deny-intended** — a step failure MUST block the commit. The configured step timeouts sum to **210s** (vet 30s + lint 60s + test 120s, per `.moai/config/sections/gate.yaml:17-19`). The gate is wired as a PreToolUse hook in `.claude/settings.json:39-50` with `"timeout": 5`.

After 5s, the Claude Code runtime stops waiting for the hook's stdout. The deny JSON the gate produces is **silently dropped** — the commit proceeds as if the gate had passed. The blocking intent is physically defeated, not by a logic bug, but by a wall-clock budget mismatch.

The census report (`.moai/reports/census-2026-07-27-handoff.md` line 107, section C-2) empirically reproduced this:

```
$ /usr/bin/time -p ... hook pre-tool < git-commit-payload
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny", ...}}
real 10.58

$ gtimeout 5 ... hook pre-tool < same payload
EXIT=124
$ wc -c < stdout
       0            ← deny disappeared
```

This repo alone takes >3s for `go vet ./...` cold — the 5s budget is unreachable for any non-trivial project.

### A.2 The `gate.yaml:11-12` comment already named the risk — and under-scoped it

The config comment states: *"the PreToolUse hook budget is 5s while the scan timeout is 30s"*. The comment scoped this risk to the **advisory ast-grep step only** ("in advisory mode a timeout only loses the advisory output"). Census found that **vet/lint/test are NOT advisory** — they are deny-intended — so the same 5s-vs-210s mismatch silently disarms the blocking gate. `SPEC-GATE-001` REQ-GATE-011 already encodes the no-bypass intent ("When the command is `git commit --amend` or contains `--no-verify`, the quality gate SHALL still execute (no bypass)"). The 5s budget physically defeats that intent every commit.

### A.3 Cross-evidence raises severity (memory `project_worktree_branch_guard_handoff`)

Under maintainer `defaultMode: bypassPermissions` (`.claude/settings.json:309`), the settings.json `allow` / `ask` permission layer is fully **no-op** — every tool call auto-approves. The PreToolUse hook deny is the **sole remaining blocking mechanism** between Claude Code and a bad commit. P1-B does not merely weaken a defense — it **breaks the sole defense**. This is why census option (a) (async/advisory) was rejected: converting the sole blocking mechanism to advisory would multiply the defect's blast radius rather than fix it.

### A.4 The 16-language detection surface

`gate.go:88-225` `toolchains` detects Go / Node.js (TypeScript+JavaScript) / Python / Rust / Java / Kotlin / C# / Ruby / PHP / Swift / Dart-Flutter / Elixir / Scala / Haskell / Zig via marker files (`go.mod`, `package.json`, `pyproject.toml`, ...). Dart/Flutter is dynamically resolved via pubspec content (issue #652). The relocation MUST preserve this detection — a Go-only fix would regress the other 15 languages.

### A.5 Distributed-template obligation

The fix ships via `internal/template/templates/` (CLAUDE.local.md §2 Template-First). `moai init` and `moai update` MUST install the relocated gate for **all downstream users**, not only the maintainer repo. The distributed template content is consumed by all 16 supported languages and MUST remain language-neutral (CLAUDE.local.md §15) and internal-content-clean (CLAUDE.local.md §25).

## B. Approach (user-approved direction (e) — move the heavy gate OFF PreToolUse)

### B.1 Census offered three options; (e) is the user-approved direction

Census C-2 (line 119) proposed three on-PreToolUse remedies:
- **(a)** convert gate to async/advisory — **REJECTED** (would weaken the sole defense under `bypassPermissions`).
- **(b)** raise `settings.json.tmpl` PreToolUse `timeout` above 210s — **NOT CHOSEN** (turns every Write/Edit/Bash call into a 210s-or-timeout stall; also depends on the unverified Claude Code timeout-enforcement gap, census line 411).
- **(c)** cap gate wall-clock to ~4s and emit `ask` on incomplete — **NOT CHOSEN** (every non-trivial repo hits the cap; `ask` becomes the default, recreating the friction the gate was meant to remove).

The user-approved direction is **(e)**: relocate the heavy gate to a surface NOT subject to the 5s PreToolUse budget. The leading surface is a **native git pre-commit hook**:
- git invokes the hook in the user's shell when `git commit` executes — git's own mechanism, outside Claude Code's hook budget entirely.
- Reject (exit non-zero) fails the commit itself — **STRONGER** than PreToolUse deny, because the commit object never materializes.
- PreToolUse retains only the fast security checks (ast-grep scanner, frozen-zone guard) that the just-merged SPEC-FALSE-ALLCLEAR-GUARD-001 (PR #1183) already tuned.

### B.2 Census Gap (line 411) is SIDESTEPPED by direction (e)

Census line 411 flagged an open question: *does Claude Code actually enforce the `settings.json` PreToolUse `timeout` field?* Direction (e) does NOT depend on the answer — git pre-commit runs outside Claude Code's hook budget entirely, so the timeout-enforcement question is moot for this SPEC. (It remains an open question for any future SPEC that relies on PreToolUse timeout semantics.)

### B.3 The leading reuse path — extend SPEC-PRECOMMIT-001's existing installer

A prior completed SPEC, **SPEC-PRECOMMIT-001** (2026-07-05, status: completed), already installed a **deployment-layer pre-commit git hook** via `PreCommitInstaller` in `internal/cli/hook_install.go`, wired into `moai init` / `moai update`. Its hook body runs `gofmt -l` + `go vet` on staged Go files — a deliberately **fast subset** (the "2-tier commit=fast / push=full-ci" design).

The cleanest relocation reuses this existing installer: **extend** its hook body to additionally invoke the heavy gate (the existing `QualityGate.Run` path, or a thin `moai gate` CLI wrapper), preserving the fast-subset checks as the first line and adding the heavy vet/lint/test as the commit-tier enforcement. This avoids:
- Duplicating the installer infrastructure (`PreCommitInstaller`, byte-identity CI test, install wiring).
- A parallel hook file that would compete for `.git/hooks/pre-commit` (only one hook per git event).
- Regressing SPEC-PRECOMMIT-001's REQ-PC-015 byte-identity invariant.

The exact extension shape (single combined script vs layered call) is an M2 design decision grounded in M1's empirical findings.

### B.4 PRESERVE list (do NOT touch)

- `internal/hook/quality/astgrep_gate.go` and the ast-grep PreToolUse path — just tuned by SPEC-FALSE-ALLCLEAR-GUARD-001 PR #1183.
- The frozen-zone guard PreToolUse path.
- `internal/hook/quality/gate.go` `IsGitCommit` regex (still used by PreToolUse for the fast-check path).
- `QualityGate.detectToolchain()` and the 16-language toolchain table.
- SPEC-PRECOMMIT-001's fast-subset (gofmt + vet) behavior inside the existing hook body — **extended, not replaced**.
- The pre-push hook installer (`PrePushInstaller`, `prePushHookContent`, `TestPrePushTemplateMatchesConstant`, `SKIP_MOAI_PREPUSH`) — separate tier, untouched.

## C. Requirements (GEARS)

### C.1 Denial semantics + defect elimination

- **REQ-PGM-001** (Ubiquitous): The relocated commit-quality gate SHALL preserve deny-intended semantics — a vet/lint/test step failure MUST prevent the commit, not merely warn.

- **REQ-PGM-002** (Event-detected): **When** the gate's wall-clock duration exceeds the PreToolUse 5s hook budget, the gate's deny verdict SHALL still reach the caller via the relocation surface (native git pre-commit hook), eliminating the silent-drop defect documented in census C-2.

### C.2 Relocation target + reuse

- **REQ-PGM-003** (Ubiquitous): The heavy commit-quality gate (vet + lint + test) SHALL be invoked from within a native git pre-commit hook executed by git in the user's shell, NOT from the PreToolUse hook handler subject to Claude Code's 5s hook budget.

- **REQ-PGM-004** (Capability gate): **Where** SPEC-PRECOMMIT-001's `PreCommitInstaller` already installs a deployment-layer pre-commit hook, the relocation SHALL extend that installer's hook body to additionally invoke the heavy gate, rather than introduce a parallel installer.

### C.3 PreToolUse fast-check preservation

- **REQ-PGM-005** (Ubiquitous): The PreToolUse surface SHALL retain the fast security checks (ast-grep scanner tuned by SPEC-FALSE-ALLCLEAR-GUARD-001 PR #1183, and the frozen-zone guard); only the heavy vet/lint/test steps are relocated off PreToolUse.

### C.4 Bypass defense

- **REQ-PGM-006** (Event-detected): **When** a Bash command contains `git commit` together with `--no-verify`, the orchestrator (via PreToolUse doctrine or `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier destructive-primitive extension) SHALL deny or require explicit user confirmation, as defense-in-depth — git pre-commit is mechanically bypassed by `--no-verify` and cannot enforce this alone.

### C.5 Language neutrality

- **REQ-PGM-007** (Ubiquitous): The relocated gate SHALL preserve the existing 16-language toolchain detection (Go, Python, Node.js/TypeScript, Rust, Java, Kotlin, C#, Ruby, PHP, Swift, Dart/Flutter, Elixir, Scala, Haskell, Zig) via `QualityGate.detectToolchain()` — no language is dropped or deprioritized.

### C.6 Template distribution + neutrality

- **REQ-PGM-008** (Ubiquitous): The relocation SHALL ship via `internal/template/templates/.git_hooks/pre-commit` (and the byte-identical Go constant in `internal/cli/hook_install.go`) so that `moai init` and `moai update` install the relocated gate for all downstream users (Template-First cycle: edit template → `make build`).

- **REQ-PGM-009** (Unwanted): The distributed template content SHALL NOT carry internal SPEC IDs (e.g. `SPEC-PRETOOL-GATE-MOVE-001`), REQ tokens (e.g. `REQ-PGM-NNN`), audit citations, internal dates, commit SHAs, macOS-bias paths, or CLAUDE.local references (per `.moai/docs/template-internal-isolation-doctrine.md` §25).

### C.7 Bypass env var parity

- **REQ-PGM-010** (State-driven): **While** the `SKIP_MOAI_PRECOMMIT=1` environment variable is set (or the project-config equivalent), the relocated gate SHALL print a bypass notice to stderr and exit 0, preserving SPEC-PRECOMMIT-001 REQ-PC-012 bypass semantics for symmetry with the pre-push `SKIP_MOAI_PREPUSH` pattern.

### C.8 Error surfacing

- **REQ-PGM-011** (Event-detected): **When** the relocated git pre-commit hook rejects a commit (exit non-zero), git's stderr — carrying the gate's failure output — SHALL be visible to Claude Code via the Bash tool result so the agent can relay the rejection to the user.

### C.9 Fallback path

- **REQ-PGM-012** (State-driven): **While** M1 verification determines that git pre-commit does NOT fire under Claude Code's `git commit` Bash invocation, the plan SHALL fall back to either census option (d) split-gate (5s sync pre-check + async full check) or option (e-prime) standalone `moai gate` CLI invoked explicitly by the user/CI.

## D. Constraints (HARD)

- **16-language neutrality (CLAUDE.local.md §15)**: the distributed pre-commit template is consumed by all 16 supported languages. It SHALL contain no Go-specific assumptions beyond `command -v go` / `command -v gofmt` guards. Non-Go projects pass silently.
- **Template-First cycle (CLAUDE.local.md §2)**: edit `internal/template/templates/.git_hooks/pre-commit` first, then `make build` (recompiles the `//go:embed all:templates` binary), then sync to local `.git_hooks/`. The byte-identity invariant (REQ-PGM-008, mirroring SPEC-PRECOMMIT-001 REQ-PC-015) SHALL hold after the build.
- **Internal-content isolation (CLAUDE.local.md §25)**: the distributed template carries no internal SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs, macOS-bias paths, or CLAUDE.local references.
- **Cross-platform build (SPEC-PRECOMMIT-001 REQ-PC-019 parity)**: any new Go installer code SHALL compile cleanly under `GOOS=windows GOARCH=amd64 go build ./...` using only portable primitives (`os`, `bufio`, `filepath`, `strings`, `errors`, `io`); no OS-specific syscalls.
- **Scope discipline (PRESERVE)**: do NOT touch PRESERVE-listed surfaces in §B.4 — ast-grep scanner, frozen-zone guard, `IsGitCommit` regex, `detectToolchain()`, pre-push installer, SPEC-PRECOMMIT-001 fast-subset behavior.
- **Census Gap acknowledgement (sidestepped)**: census line 411 — *does Claude Code enforce `settings.json` PreToolUse `timeout`?* — is unverified. Direction (e) does NOT depend on the answer (git pre-commit runs outside Claude Code's hook budget entirely). If direction (e) is abandoned in favor of option (b) timeout-raise, this gap becomes MUST-FIX-blocking and must be resolved first.

## E. Cross-References

- Census report: `.moai/reports/census-2026-07-27-handoff.md` §C-2 (line 107), §7 Priority 1 P1-B (line 619).
- Defect-source SPEC: `SPEC-GATE-001` (status: implemented) — original deterministic quality gate; REQ-GATE-011 ("`--no-verify` and `--amend` do not bypass the gate") is the intent this SPEC restores mechanically.
- Sibling fast-check SPEC: `SPEC-FALSE-ALLCLEAR-GUARD-001` (PR #1183, the worktree base `3e6c92ef7`) — just tuned the ast-grep PreToolUse scanner to report unavailable rather than false all-clear.
- Deployment-layer installer SPEC: `SPEC-PRECOMMIT-001` (status: completed) — `PreCommitInstaller`, `preCommitHookContent`, `installPreCommitHookOptional`, REQ-PC-015 byte-identity test pattern.
- Codebase anchors:
  - `internal/hook/quality/gate.go:254` — `QualityGate.Run(ctx)` (synchronous vet/lint/test, deny-intended).
  - `internal/hook/quality/gate.go:88-225` — `toolchains` (16-language detection table).
  - `internal/hook/quality/gate.go:621-627` — `IsGitCommit` regex + `--no-verify`/`--amend` no-bypass intent.
  - `.claude/settings.json:39-50` — PreToolUse hook registration (`"timeout": 5`, matcher `Write|Edit|Bash`).
  - `.claude/settings.json:309` — `defaultMode: bypassPermissions` (severity-escalator context).
  - `.moai/config/sections/gate.yaml:11-19` — step timeouts (vet 30s, lint 60s, test 120s) + the 5s-vs-30s comment that under-scoped the risk.
- Doctrinal anchors:
  - `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier Doctrine — destructive-primitive set already includes `git push --no-verify`; this SPEC extends it to `git commit --no-verify`.
  - CLAUDE.local.md §2 (Template-First), §15 (16-language neutrality), §23 (PR-mandatory policy), §25 (internal-content isolation).

## F. Out of Scope

### Out of Scope — PreToolUse timeout-enforcement verification (census line 411)

- Whether Claude Code actually enforces `settings.json` PreToolUse `timeout` is unverified. Direction (e) makes this moot. If a future SPEC pursues option (b) timeout-raise, that SPEC owns the verification.

### Out of Scope — pre-push hook regression

- `PrePushInstaller`, `prePushHookContent`, `TestPrePushTemplateMatchesConstant`, and the `SKIP_MOAI_PREPUSH` bypass pattern are unchanged. The push tier continues to run `make ci-local` (lint + vet + test + cross-compile). Touching the push tier would violate SPEC-PRECOMMIT-001's 2-tier commit=fast / push=full-ci boundary.

### Out of Scope — `pre_commit` runtime severity dial

- SPEC-PRECOMMIT-001 REQ-PC-017 explicitly deferred a runtime severity dial (`skip` / `warn` / `enforce`) for the `pre_commit` config field. This SPEC does not wire it; the relocated heavy gate runs unconditionally (subject only to `SKIP_MOAI_PRECOMMIT=1`).

### Out of Scope — non-Go per-language fast checks at the commit tier

- SPEC-PRECOMMIT-001's fast subset targets staged Go files only (`gofmt -l` + `go vet`). Wiring per-language formatters (prettier, ruff, rustfmt, ...) for non-Go staged file types at the fast-subset tier remains out of scope. The heavy gate's `detectToolchain()` already handles per-language lint/test — this is preserved by relocation, not extended.

### Out of Scope — commit-msg and other git hooks

- Only the pre-commit hook is affected. The commit-msg convention validator, the pre-push hook, and any other git hook are untouched.

### Out of Scope — re-architecting PreCommitInstaller shared helpers

- The relocation reuses SPEC-PRECOMMIT-001's installer primitives (`fileHasMoaiMarker`, `ErrUserHookExists`, `0755` mode, non-fatal optional wrapper). It MUST NOT re-architect the shared helpers.

### Out of Scope — CI/CD pipeline changes

- The relocation is local-only (git pre-commit runs in the user's shell). CI pipeline integration is unchanged; CI remains the push-tier (or later) enforcement surface.

### Out of Scope — rewriting SPEC-GATE-001 or SPEC-PRECOMMIT-001 bodies

- Both prior SPECs are cross-referenced, not modified. SPEC-GATE-001's REQ-GATE-011 no-bypass intent is mechanically operationalized by this SPEC; SPEC-PRECOMMIT-001's installer is extended, not rewritten.
