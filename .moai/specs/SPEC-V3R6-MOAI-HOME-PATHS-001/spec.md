---
id: SPEC-V3R6-MOAI-HOME-PATHS-001
title: "internal/paths single source of truth for ~/.moai resolution (MoaiHome() + MOAI_HOME)"
version: "0.1.0"
status: in-progress
created: 2026-08-16
updated: 2026-08-17
author: MoAI orchestrator
priority: P1
phase: "v3.2.0 target"
module: "internal/paths, internal/statusline, internal/config, internal/cli, internal/update, internal/kanban, internal/hook, internal/glmcred"
lifecycle: spec-anchored
tags: "paths, home-resolution, moai-home, env-override, ssot, refactor"
tier: M
era: V3R6
related_specs: [SPEC-GLM-KEY-INPUT-001, SPEC-V3R6-SESSION-HANDOFF-AUTO-001, SPEC-DEPRECATEDPATHS-RECONCILE-001, SPEC-KANBAN-BOOTSTRAP-001]
---

# SPEC: internal/paths single source of truth for `~/.moai` resolution

## 1. Problem and Context

`~/.moai` path construction is scattered across seven packages with three incompatible home-resolution styles and no override mechanism:

- **Style 1 — inline `os.UserHomeDir()`**: 13 join sites (statusline, config resolver, cli launcher/deps, update, kanban, hook).
- **Style 2 — HOME-first wrapper** (`internal/cli/homedir.go`, duplicated verbatim in `internal/cli/preference/cmd.go:163-168`): checks `os.Getenv("HOME")` first so `t.Setenv("HOME", ...)` works on Windows where `os.UserHomeDir()` ignores `HOME`.
- **Style 3 — injected seams** (`glmcred.HomeDirFn`, `NewUsageCollector(homeDir)`, `ReadModelCache(homeDir)`).

Exact inventory: **17** non-test `filepath.Join(<home-var>, ".moai", ...)` sites (16 matched by the original predicate + 1 hidden by receiver-prefix form `Join(r.homeDir, ".moai", ...)` at `internal/cli/migrate_agency.go:184`). Full table in plan.md §2.

No `MOAI_HOME` environment variable exists anywhere in the repo (verified: zero grep hits, all file types). The env-name registration discipline (`internal/config/envkeys.go`) applies in full to a new variable.

Evidence base: `research.md` in this directory (4-lens synthesis with contradictions and open gaps; the membership contradiction was resolved by direct line reads — see plan.md §2 notes).

## 2. Requirements (GEARS)

- **REQ-MHP-001** — WHEN any Go code needs the `~/.moai` root directory, THE SYSTEM SHALL resolve it exclusively through `internal/paths.MoaiHome()`, WHICH SHALL honor `MOAI_HOME` when set to a non-empty absolute path, treating an empty value as unset (fall back to home), and disregarding relative values (XDG semantics).
- **REQ-MHP-002** — WHEN `MOAI_HOME` is unset/empty/relative and the fallback runs, `MoaiHome()` SHALL resolve the user's home HOME-first: `os.Getenv("HOME")` when non-empty, else `os.UserHomeDir()` (the `internal/cli/homedir.go` contract, characterized by `home_isolation_test.go`).
- **REQ-MHP-003** — `internal/paths` SHALL be stdlib-only (standard library imports only), so `internal/glmcred` (stdlib-only by design) can adopt it without import cycles.
- **REQ-MHP-004** — The `MOAI_HOME` name SHALL be registered as `EnvHome` in `internal/config/envkeys.go`; `internal/paths` SHALL reference the name via the glmcred-style local alias pattern (glmcred.go:22-27 precedent), because importing `internal/config` from `internal/paths` would create a cycle. No call site in any direction (read or write) may inline the env name as a string literal.
- **REQ-MHP-005** — `MoaiHome()` and sub-path accessors SHALL resolve per call; caching (`sync.Once` or package-level memoization) is prohibited (breaks the 107 `t.Setenv("HOME", ...)` tests that re-resolve later).
- **REQ-MHP-006** — `MoaiHome()` SHALL return `(string, error)`; callers preserve their existing degradation behavior on error. The `internal/profile.GetBaseDir()` `"."` relative fallback MUST NOT be replicated.
- **REQ-MHP-007** — `internal/paths` SHALL expose typed sub-path accessors for the migrated targets — `.env.glm`, `cache/`, `state/`, `releases/`, `worktrees/`, `claude-profiles/`, user-tier `settings.json`, `config/sections` — consuming segment constants from `internal/defs/dirs.go` (`MoAIDir` and documented sub-segments); no accessor re-literals `".moai"`.
- **REQ-MHP-008** — The 17 migration sites listed in plan.md §2 SHALL route through `internal/paths` accessors; after migration the predicate in acceptance.md AC-1 SHALL return zero matches.
- **REQ-MHP-009** — WHERE a consumer compares against or whitelists the `~/.moai` path for correctness — `internal/hook/pre_tool.go:277` (allowed-external-paths security whitelist) and `internal/core/project/root.go:35-50` (home special-case preventing `~/.moai` as project root) — THE SYSTEM SHALL resolve through the same accessor in lockstep, so an overridden `MOAI_HOME` is consistently trusted.
- **REQ-MHP-010** — The helper landscape SHALL collapse to one implementation: `internal/cli/homedir.go` becomes a thin delegate to `internal/paths`; the verbatim duplicate in `internal/cli/preference/cmd.go:163-168` is deleted; `glmcred`'s `.moai` join routes through the accessor. Existing parallel-safe test seams (`HomeDirFn`, injected `homeDir` parameters) are preserved.
- **REQ-MHP-011** — Tests SHALL cover: MOAI_HOME honored / empty==unset / relative-disregarded / HOME-first precedence (non-parallel `t.Setenv` discipline per CLAUDE.local.md §6/§13), with coverage ≥ 85% in `internal/paths`.
- **REQ-MHP-012** — The SPEC supersedes the recorded position in SPEC-V3R6-SESSION-HANDOFF-AUTO-001 spec.md:30 ("new code MUST reuse `os.UserHomeDir()` — no abstraction layer exists") for home resolution: the abstraction layer now exists and is mandatory for `~/.moai` paths.

### 2.1 Out of Scope — non-goals

- **Directory relocation**: moving or re-layouting `~/.moai` itself is a rejected direction (prior design review: older binaries reading old paths outweigh tidiness).
- **Shell-side honoring**: ~42 shell files (`handle-*.sh`, `status_line.sh`) keep `$HOME/.moai`. Known hazard documented in §4; follow-up card recommended, not delivered here.
- **Non-`.moai` home sites**: the other ~25 `os.UserHomeDir` call sites (including `.claude` joins such as `doctor.go:551` and the `statusline.go:39` chdir fallback) stay as-is.
- **`internal/profile.GetBaseDir()`**: deferred — PR #1569 (e63551b7d) just landed `os.SameFile`/ledger matching on that path; changing its base concurrently with the profile-key semantics is risk without need. Its accessor migration is a follow-up.
- **Caching**: explicitly out (see REQ-MHP-005).

## 3. Design Sketch

```go
package paths // stdlib-only

// Home returns the raw user home, HOME-first (homedir.go contract).
func Home() (string, error)
// MoaiHome returns the ~/.moai root, honoring MOAI_HOME (REQ-MHP-001/002).
func MoaiHome() (string, error)
// Sub-path accessors (consume internal/defs/dirs.go segments).
func StateDir() (string, error)      // ~/.moai/state
func CacheDir() (string, error)      // ~/.moai/cache
func ReleasesDir() (string, error)   // ~/.moai/releases
func WorktreesDir() (string, error)  // ~/.moai/worktrees
func ProfilesDir() (string, error)   // ~/.moai/claude-profiles
func GlmEnvFile() (string, error)    // ~/.moai/.env.glm
func UserSettingsFile() (string, error)        // ~/.moai/settings.json
func UserConfigSectionsDir() (string, error)   // ~/.moai/config/sections
```

Call sites with injected `homeDir` seams (statusline collectors) keep their signatures; the seam callers resolve via `paths.Home()` / accessors.

## 4. Known Hazard (documented, not fixed here)

`status_line.sh:33-35` sources `$HOME/.moai/.env.glm` BEFORE the Go binary runs, and Go reads the same file via glmcred. Under a non-default `MOAI_HOME`, Go would read the overridden location while shell keeps the default — one credential file split across two locations. Recommended follow-up backlog card: shell-side `MOAI_HOME` support in `status_line.sh` + `handle-*.sh` (42 files), or documenting `MOAI_HOME` as Go-process-only.

## 5. Supersession

This SPEC supersedes SPEC-V3R6-SESSION-HANDOFF-AUTO-001's home-resolution position (spec.md:30) as recorded in REQ-MHP-012. No other SPEC addresses `~/.moai` home resolution (verified: `.moai/specs/` 628 dirs, zero hits for internal/paths / MoaiHome / MOAI_HOME).

## 6. Provenance Note

Authored orchestrator-direct on 2026-08-16 with user approval: two `manager-spec` delegations failed on subagent context-window limits (13:04Z, 13:05Z; zero artifacts produced). Evidence chain: research.md (4-lens fan-out, run wf_ebb53184-4f0) + direct line reads for the disputed sites (plan.md §2 notes).

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-08-16 | 0.1.0 | MoAI orchestrator | Initial authoring (plan-phase, orchestrator-direct per user approval after subagent context-limit failures). |
| 2026-08-16 | 0.1.0 | MoAI orchestrator | Audit iteration 1 remediation (D1–D6, user-approved): accessor granularity resolved within the sketched 8-accessor surface; stale WIP premises corrected (#1571 landed, 504797021); AC-MHP-016 added (agency-adapter override reach, AC count 16/16); line citations re-verified against HEAD; HISTORY section added. |
