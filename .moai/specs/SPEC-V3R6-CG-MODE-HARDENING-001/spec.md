---
id: SPEC-V3R6-CG-MODE-HARDENING-001
title: "moai cg (Claude+GLM) launch-mode hardening and detector SSOT"
version: "0.1.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: High
phase: "v3.0.0"
module: "internal/cli, internal/tmux, internal/config"
lifecycle: spec-anchored
tags: "cg-mode, glm, launcher, tmux, settings-atomicity, security"
related_specs: [SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001]
tier: M
---

## HISTORY

- 2026-06-22 — v0.1.0 — Initial draft. Authored by manager-spec from a 4-dimension read-only audit of `moai cg`. All defect claims independently re-verified against the cited source during plan-phase (per `.claude/rules/moai/core/verification-claim-integrity.md` — a defect claim is a hypothesis until the code confirms it). The AGENT-REPORTED security finding (REQ-CGH-007) and the EXCLUDED process-env-pollution hypothesis were both checked at the code level before inclusion / exclusion.

## §A. Context and Motivation

`moai cg` launches a hybrid session: a **Claude leader** pane (whose env is kept CLEAN of GLM credentials) plus **GLM teammate** panes (spawned into new tmux windows that inherit GLM credentials from the tmux SESSION env). The cost-optimization premise is that the expensive leader runs on Claude while implementation teammates run on the cheaper GLM backend.

The launch path (`applyCGMode` in `internal/cli/launcher.go`, the GLM env helpers in `internal/cli/glm.go`, and the CG-mode detector in `internal/tmux/cg_detect.go`) carries a cluster of correctness and robustness defects that were surfaced by a read-only audit and confirmed against the code:

- The shared launch path uses an unguarded `syscall.Exec` that returns `EWINDOWS` on Windows.
- The CG settings mutation is a multi-step, non-atomic sequence with a failure window that can leave stale GLM credentials in the leader config or drop `teammateMode` entirely.
- The CG-mode detector keys off the PROCESS environment — a signal the CG design deliberately keeps clean on the leader — instead of the deterministic `llm.yaml team_mode` source of truth, so teammate spawning can silently route to Claude instead of GLM and defeat the cost optimization.
- A `teammateMode` value documented in `CLAUDE.local.md §22.3` does not match what the code writes.
- A `GLM base_url` from a tracked config file is injected into the process/tmux environment with no validation.
- The credential-routing functions are under-tested at the CLI layer because they early-return in the test environment.

The **headline fix** is the detector SSOT realignment (REQ-CGH-006): making `IsCGMode` authoritative on a deterministic, design-clean signal rather than the process env. The launch-path/atomicity fixes (REQ-CGH-001, REQ-CGH-002, REQ-CGH-003, REQ-CGH-005) are a robustness cluster; the doc fix (REQ-CGH-004), the tmux-availability precondition (REQ-CGH-008), the URL-validation security fix (REQ-CGH-007), and the test-coverage requirement (REQ-CGH-009) round out the SPEC.

### §A.1 Verified facts (file:line, re-confirmed during plan-phase)

| Fact | Location | Verified |
|------|----------|----------|
| `syscall.Exec(claudeBin, buildArgs(false), launchEnv)` unguarded; `runtime` NOT imported | `internal/cli/launcher.go:541` (+ grep: no `"runtime"` import) | YES |
| Correct Windows spawn-child-and-exit pattern to mirror | `internal/cli/update.go:461` (`reexecNewBinary`) | YES |
| Build-tag split precedent | `internal/cli/worktree/team_launch_posix.go` + `team_launch_windows.go` | YES |
| `applyCGMode` order: injectTmux(174) → persistTeamMode(189) → removeGLMEnv(193) → ensureSettingsLocalJSON(197); returns at 175 on tmux-inject failure BEFORE removeGLMEnv | `internal/cli/launcher.go:143-204` | YES |
| `removeGLMEnv` sets `settings.TeammateMode = ""` | `internal/cli/launcher.go:229` | YES |
| `ensureSettingsLocalJSON` then sets `settings.TeammateMode = "tmux"` | `internal/cli/glm.go:509` | YES |
| All settings.local.json mutators use unlocked `ReadFile→Unmarshal→MarshalIndent→os.WriteFile` (non-atomic) | `launcher.go:209-268, 624-669`; `glm.go:492-529` | YES |
| Correct atomic temp-file + `os.Rename` pattern that exists in-repo | `internal/cli/glm.go:680-698` (`saveLLMSection`) | YES |
| `IsCGMode` requires `hasGLMEnv()` reading process env (`ANTHROPIC_AUTH_TOKEN` / `ANTHROPIC_BASE_URL` containing `z.ai`) | `internal/tmux/cg_detect.go:41-87` | YES |
| Existing `REQ-WTL-009` drift warning to reconcile (not delete) | `internal/tmux/cg_detect.go:3, 32-34, 63-72` | YES |
| `persistTeamMode(root, "cg")` writes `llm.yaml team_mode` (deterministic SSOT, NOT consumed by `IsCGMode`) | `internal/cli/launcher.go:189`; `internal/cli/glm.go:472-487` | YES |
| `Detector.IsAvailable()` + `Version()` exist; `applyCGMode` only calls `InTmuxSession()` | `internal/tmux/detector.go:60-84`; `launcher.go:160` | YES |
| `loadGLMConfig` returns `cfg.LLM.GLM.BaseURL` verbatim; no `https://`/`api.z.ai` allowlist | `internal/cli/glm.go:742-778` | YES |
| `validateLLM` checks only model-name strings for dynamic-token patterns; no URL validation | `internal/config/validation.go:349-352` | YES |
| Canonical default `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"` | `internal/config/defaults.go:41` | YES |
| Live coverage: `injectTmuxSessionEnv` 10.5%, `clearTmuxSessionEnv` 25.0%, `ensureSettingsLocalJSON` 83.3%, `applyCGMode` 60.7%, `removeGLMEnv` 90.9%, `loadGLMConfig` 100% | `go tool cover -func` on `internal/cli` (baseline 71.8%) | YES |

### §A.2 Disproven hypothesis (explicitly excluded — see §H Exclusions)

The audit's "leader inherits stale GLM env from the process env" hypothesis was DISPROVEN at the code level: `applyCGMode` does NOT call `setGLMEnv()`, so the `moai cg` CLI process env is never polluted with GLM credentials. This is NOT a requirement of this SPEC.

## §B. Requirements (GEARS notation)

### §B.1 Cross-platform launch safety

- **REQ-CGH-001** — **Where** the host platform is Windows, the launcher shall NOT call `syscall.Exec` on the shared `launchClaudeDefault` path; it shall instead spawn the `claude` child process, wait for it, and propagate its exit code (mirroring the `reexecNewBinary` pattern at `update.go:461`), so that `moai cc` / `moai glm` launch correctly on Windows. **Where** the host platform is POSIX, the launcher shall continue to use `syscall.Exec` to replace the current process.

### §B.2 CG settings mutation atomicity and ordering (robustness cluster)

- **REQ-CGH-002** — **When** `applyCGMode` runs, the leader-config cleanup that strips stale GLM credentials shall execute BEFORE the failure-prone tmux session-env injection, so that a tmux-injection failure cannot leave stale GLM credentials in the leader's `settings.local.json` `env` block.
- **REQ-CGH-003** — The CG settings mutation shall set `teammateMode = "tmux"` AND strip the GLM credential keys from `settings.local.json` in a SINGLE read-modify-write, so that no intermediate file state exists in which `teammateMode` is absent. The launcher shall not perform two separate writes (one clearing `teammateMode`, a later one re-setting it) for a single `moai cg` invocation.
- **REQ-CGH-005** — The `settings.local.json` mutations on the launch path shall be performed through a single helper that (a) serializes concurrent writers (file lock) AND (b) writes atomically via a temp-file + `os.Rename` (mirroring the `saveLLMSection` pattern at `glm.go:680-698`), so that concurrent `moai cc`/`glm`/`cg` launches cannot produce a truncated or last-writer-wins-clobbered settings file. The helper shall preserve user-only keys (`defaultMode`, `env.PATH`, and any unrecognized keys) per `internal/cli/CLAUDE.md` (settings mutation through helpers) and `CLAUDE.local.md §22`.

### §B.3 Detector source-of-truth realignment (HEADLINE)

- **REQ-CGH-006** — The CG-mode detector (`IsCGMode`) shall determine CG mode from a DETERMINISTIC, design-clean source of truth — the `llm.yaml team_mode` value (written by `persistTeamMode`) and/or the tmux SESSION environment (`tmux show-environment`) — and shall NOT rely on the leader PROCESS environment (`os.Getenv("ANTHROPIC_AUTH_TOKEN")` / `ANTHROPIC_BASE_URL`), so that a teammate spawn issued from a clean CG leader pane is correctly classified as CG mode and routes teammates to GLM. The existing `REQ-WTL-009` drift warning shall be RECONCILED with the new source of truth (preserved as a meaningful signal, not deleted).

### §B.4 Documentation correctness

- **REQ-CGH-004** — `CLAUDE.local.md §22.3` shall document the `teammateMode` local value as `"tmux"` (CG/GLM modes) or `""` (cc mode) to match what the code writes, and shall disambiguate the `.claude/settings.local.json` `teammateMode` field (`"tmux"`/`""`) from the `llm.yaml team_mode` field (`cg`/`glm`/`""`). **Where** the SPEC also proposes editing the user-facing CG Mode description in the template-managed root `CLAUDE.md §15`, the change shall ALSO be applied to `internal/template/templates/CLAUDE.md` followed by `make build` (per `CLAUDE.local.md §2` Template-First). `CLAUDE.local.md` itself is a LOCAL dev file and requires no template sync.

### §B.5 tmux availability precondition

- **REQ-CGH-008** — **When** `applyCGMode` detects an active tmux session (`InTmuxSession()` true) but the tmux binary is unavailable (`Detector.IsAvailable()` false), the launcher shall fail with a clear "tmux not installed / not executable" message (including install guidance) rather than the misleading "restart your tmux session" message.

### §B.6 Security: GLM base_url validation

- **REQ-CGH-007** — **When** a GLM `base_url` is loaded from `llm.yaml` and injected into the process/tmux environment, the configuration layer shall validate it against an allowlist constraint (well-formed `https://` URL; host matching the canonical `api.z.ai` family or an explicitly user-configured override), and **When** the value fails validation the system shall reject it with a clear error rather than silently routing `ANTHROPIC_AUTH_TOKEN` to the unvalidated endpoint. The canonical default `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"` (`defaults.go:41`) shall always pass.

### §B.7 Test coverage of credential-routing invariants

- **REQ-CGH-009** — The test suite shall exercise the production `applyCGMode` credential-routing path (NOT only the `isTestEnvironment()` early-return short-circuit), asserting two invariants: (1) the leader's `settings.local.json` env block is STRIPPED of GLM credentials, and (2) the teammate-facing injection sets the GLM credential set. The test suite shall additionally assert a list-parity invariant: every key set by `injectTmuxSessionEnv` (except the intentionally-excluded `ANTHROPIC_AUTH_TOKEN`, which `clearTmuxSessionEnv` documents as deliberately retained) appears in `clearTmuxSessionEnv`'s removal list.

### §B.8 Regression-safety (ubiquitous)

- **REQ-CGH-010** — The launcher shall preserve all existing `moai cc` / `moai glm` / `moai cg` behavior not named above: the full `internal/cli` and `internal/tmux` test suites shall pass, and `golangci-lint` shall report zero new findings.

## §C. Constraints

- **C-1 — Subagent boundary (C-HRA-008 / REQ-PGN-012)**: CLI code MUST NOT call `AskUserQuestion` / `mcp__askuser__*`. All new error paths use positional-arg + `--flag` + structured stderr errors. The static guard `TestNew_NoAskUserQuestion`-style grep test applies.
- **C-2 — Env var access**: New env-var references MUST use constants from `internal/config/envkeys.go` — never inline `os.Getenv("ANTHROPIC_*")` strings (per `internal/cli/CLAUDE.md` + `CLAUDE.local.md §14`).
- **C-3 — Settings mutation through helpers**: All `settings.local.json` writes go through `internal/cli/settings.go` helpers, never raw `json.Marshal` + `os.WriteFile`, and MUST preserve user-only keys (`defaultMode`, `teammateMode`, `env.PATH`).
- **C-4 — Cross-platform**: All changes MUST build under `GOOS=windows GOARCH=amd64 go build ./...`. `syscall.Exec` is POSIX-only; gate it behind a build-tag split or a `runtime.GOOS` guard.
- **C-5 — No GLM integration tests in the dev project** (`CLAUDE.local.md §13`): `moai cc`/`glm`/`cg` command flows mutate real settings files. Tests MUST use `t.TempDir()` fixtures; never `t.Setenv("HOME", ...)` in parallel tests; auth token via `loadGLMKey()` with `t.Skip()` fallback.
- **C-6 — Template neutrality** (`CLAUDE.local.md §15`, §25): If `internal/template/templates/CLAUDE.md` is edited (REQ-CGH-004 template-sync branch), the edit MUST remain language-neutral and free of internal SPEC IDs / dates / commit SHAs per the template internal-content isolation doctrine.
- **C-7 — Do NOT regress `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`**: `cg_detect.go` and `REQ-WTL-009` are owned by the sibling SPEC. REQ-CGH-006 RECONCILES with them; it MUST NOT delete the drift-warning behavior, and its tests MUST keep the sibling SPEC's `IsCGMode` tests green.

## §D. Acceptance Criteria Summary

Full Given-When-Then scenarios + the AC matrix live in `acceptance.md`. Ten requirements map to ten AC groups (AC-CGH-001 .. AC-CGH-010), each mechanically verifiable via `go test`, `grep`, `GOOS=windows go build`, or a file/coverage assertion.

## §E. Self-Verification (plan-phase)

See `progress.md` §E.1 for the plan-phase audit-ready signal. Run-phase / sync-phase evidence sections are placeholder headings only at plan-phase.

## §H. Exclusions

The following are explicitly out of scope for this SPEC. (This section satisfies the `OutOfScopeRule` lint via the `### Out of Scope — <topic>` H3 sub-headings below.)

### Out of Scope — disproven process-env-pollution claim
- The "leader inherits stale GLM env from the CLI process env" hypothesis is DISPROVEN (§A.2): `applyCGMode` never calls `setGLMEnv()`. No requirement addresses it. Do NOT add leader-process-env scrubbing.

### Out of Scope — sibling-SPEC detector ownership
- This SPEC does NOT redefine the four `--team` launch patterns (P1-P4), the swarm registry, or the `moai worktree new --team` dispatch — those are owned by `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`. REQ-CGH-006 only changes the SIGNAL `IsCGMode` reads; it does not change the pattern-selection contract.

### Out of Scope — new swarm/session commands
- `moai swarm status/done/kill-all` and any new session-management subcommands are NOT introduced here.

### Out of Scope — GLM model selection / cost routing
- Choosing which GLM model maps to High/Medium/Low slots, the `[1m]` 1M-context activation logic, and statusline context-window resolution are unchanged. This SPEC only hardens credential ROUTING and detection, not model SELECTION.

### Out of Scope — broad settings.local.json schema redesign
- REQ-CGH-005 introduces a single locked+atomic write helper for the EXISTING mutation set. It does NOT redesign the `SettingsLocal` struct, add new settings keys, or migrate the file format.
