---
id: SPEC-V3R6-MOAI-CLEAN-HOME-001
title: "moai doctor disk check + moai clean --home (safe ~/.moai cleanup with carve-outs)"
version: "0.1.0"
status: in-progress
created: 2026-08-16
updated: 2026-08-17
author: MoAI orchestrator
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/config"
lifecycle: spec-anchored
tags: "doctor, disk, clean, home, retention, safety"
tier: M
era: V3R6
related_specs: [SPEC-V3R6-MOAI-HOME-PATHS-001, SPEC-V3R2-RT-004]
depends_on: [SPEC-V3R6-MOAI-HOME-PATHS-001]
---

# SPEC: `moai doctor disk` + `moai clean --home`

## 1. Problem and Context

`~/.moai` accumulates regenerable artifacts (old release binaries, profile `debug/` logs, aged backups) with no visibility and no safe cleanup command. The existing `moai clean` is project-scoped (`.moai/state/runs/` only); `moai doctor` has no disk check (23 checks as of this tree — System 5 + MoAI-ADK 9 + Workspace 9 — none disk). Measurements and the transcripts-are-kept design decision are in research.md.

## 2. Requirements (GEARS)

- **REQ-MCH-001** — WHEN `moai doctor` runs, THE SYSTEM SHALL include an advisory `Home Disk Usage` check reporting: `~/.moai` top-level breakdown, per-profile sizes under `claude-profiles/`, cross-profile duplicate clusters (same category name with byte-equal sizes — heuristic, report-only), `releases/` count vs current version, and a `~/.claude` summary line marked report-only.
- **REQ-MCH-002** — WHERE the estimated cleanable bytes exceed a configurable threshold, the check SHALL emit WARN (default `DefaultHomeDiskWarnBytes` in `internal/config/defaults.go`, seeded 500MB; no inline literals).
- **REQ-MCH-003** — `moai clean` SHALL gain a `--home` flag that extends the existing contract: dry-run by default (list what would be deleted with sizes), `--force` to delete; scope is **`~/.moai` only** — never `~/.claude`.
- **REQ-MCH-004** — `clean --home` SHALL delete only from an explicit allowlist: per-profile `debug/` entries older than retention (under `claude-profiles/<p>/debug/`), `releases/` binaries beyond current + `keep` (default 3), root `~/.moai/logs/` files older than retention, and `backups/removed-*` directories older than retention. Everything not enumerated is untouched.
- **REQ-MCH-005** — THE SYSTEM SHALL enforce a carve-out invariant with recursive depth semantics: `isCarvedOut(relPath)` matches carve-out names (`projects/`, `config/`, `state/`, `credentials*`, `launch.yaml`, `preferences.yaml`, `worktrees/`, `mcp/`, `bin/`, `search/`, `studio/`, `plugins/`) against **any** path segment, at every depth — root level and per-profile under `claude-profiles/<p>/` alike — and the carve-out **wins inside allowlisted containers**: a `credentials*`-named file (or any carved-out segment) inside an aged `backups/removed-*` directory is never deleted, even under `--force`. Enforced by a path-predicate guard covered by a dedicated test.
- **REQ-MCH-006** — `plugins/` directories SHALL be reported by the doctor duplicate-cluster detection but never touched by `clean --home` — protection is enumerated in the REQ-MCH-005 carve-out list as well, so the safety property has a single home (Claude Code per-profile isolation assumption; dedupe strategy is an explicit non-goal — research.md §3.2).
- **REQ-MCH-007** — Retention SHALL come from `state.home_retention_days` (new key), read from the **home tier** — `~/.moai/config/sections/state.yaml` resolved via `paths.UserConfigSectionsDir()` — never from a project-tier state.yaml (cwd-dependent reads would let two projects clean one home with different retentions). Defaults to `DefaultHomeCleanRetentionDays = 30` (defaults.go) when the home-tier key is absent; explicit `0` disables cleaning (mirrors the existing retention semantics).
- **REQ-MCH-008** — Tests SHALL use a hermetic fixture home: `t.Setenv("HOME", t.TempDir())` with `MOAI_HOME` scrubbed (`t.Setenv("MOAI_HOME", "")`) so ambient overrides cannot leak in (non-parallel), covering: dry-run mutates nothing, `--force` deletes only allowlisted paths, carve-out guard (incl. carve-out-wins-inside-removed-* case), current+N releases survival, retention cutoff from the home-tier source, MOAI_HOME redirect (an absolute `MOAI_HOME` points `clean --home` at the fixture — the dependency SPEC's core promise), and project-scope `clean` regression.

### 2.1 Out of Scope

- **`~/.claude` mutation**: doctor reports it; clean never touches it (Claude Code's directory).
- **`plugins/` dedupe**: report-only in v1 (isolation risk; separate follow-up if pursued).
- **`projects/` archival or compression**: transcripts are kept by standing design decision.
- **`MOAI_HOME` override resolution**: home resolution inherits this SPEC's dependency on SPEC-V3R6-MOAI-HOME-PATHS-001 (`paths.MoaiHome()`); no parallel home-resolution logic is introduced.

## 3. Design Sketch

- `internal/cli/doctor_disk.go` — `checkHomeDisk(verbose bool) DiagnosticCheck` registered in the check slices of `runGroupedChecksObserved` (doctor.go:160-215; `runDiagnosticChecks` :226 is the backward-compat flattener); `~/.claude` line labeled `(report-only)`.
- `internal/cli/clean.go` — `--home` branch: category scanners → dry-run report → `--force` guarded deletion; carve-out predicate `isCarvedOut(relPath) bool` (recursive any-segment match, carve-out wins inside allowlisted containers) shared by scanner and guard test.
- `internal/config` — `DefaultHomeDiskWarnBytes`, `DefaultHomeCleanRetentionDays`, `DefaultReleaseKeep` in defaults.go; `state.home_retention_days` read from the home-tier state.yaml via `paths.UserConfigSectionsDir()` (a dedicated home-tier read — the existing project-scope `clean` keeps its own `stateYAMLWrapper` path untouched).

## 4. Interaction with SPEC-V3R6-MOAI-HOME-PATHS-001

`depends_on` is declared deliberately: `clean --home` and `checkHomeDisk` resolve the home root via `paths.MoaiHome()` once that SPEC's run phase lands, so the `MOAI_HOME` override is honored uniformly. Run-phase ordering: HOME-PATHS first, this SPEC second.

## 5. Provenance Note

Authored orchestrator-direct 2026-08-16 per user directive ("제안대로 진행") continuing the session's approved pattern (manager-spec spawns remain dead pre-PR-#1574; see SPEC-V3R6-MOAI-HOME-PATHS-001 progress.md). Evidence: research.md measurements (this session's du/grep outputs).

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-08-16 | 0.1.0 | MoAI orchestrator | Initial authoring (plan-phase, orchestrator-direct per user approval). |
| 2026-08-17 | 0.1.0 | MoAI orchestrator | Audit iteration 1 remediation (D1–D7, user-approved): duplicate-cluster strictness resolved (byte-equal + entry-count, report-only); carve-out depth semantics specified (recursive any-segment, carve-out wins inside allowlisted containers, plugins/ enumerated); retention source fixed to home tier via UserConfigSectionsDir; doctor registration point corrected (runGroupedChecksObserved); check count 21→23; fixture hermeticity (MOAI_HOME scrub + redirect test). |
