---
id: SPEC-CLIFIX-CRITICAL-001
title: "CLI Critical Remediation — data loss, ledger corruption, cross-harness deletion (P0)"
version: "0.1.0"
status: in-progress
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P0
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, audit-remediation, data-loss, critical, p0"
era: V3R6
tier: M
depends_on: []
---

# SPEC-CLIFIX-CRITICAL-001 — CLI Critical Remediation (P0)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §1 (Critical 8/131) + §5 P0 roadmap row |

## §A Context

The 2026-07-10 CLI audit (`.moai/reports/cli-improvement-audit-20260710.html`, 131 verified findings across `internal/cli/**` + `cmd/moai`) identified 8 Critical defects in which user data loss or file corruption is reproducible in ordinary usage flows. This SPEC remediates all 8, each with a reproduction test written before the fix (Reproduction-First rule). Every fix is local (1 file each per the audit's P0 assessment).

Findings SSOT: audit §1 Critical table. All file:line coordinates below were verified by 5 parallel read-only static audits; the run phase MUST re-verify each anchor against the live tree before editing (line numbers drift).

## §B Requirements (GEARS)

- REQ-CRIT-001-001: When `moai cc` or `moai glm` performs a read-modify-write of `.claude/settings.local.json` (glm.go:98, launcher.go:241), the CLI shall preserve every top-level key not being mutated, by representing the file as `map[string]any` on every write-back path — including the `mutateSettingsLocal` seam itself (settings.go:28), whose current `mutate func(*SettingsLocal)` signature round-trips the 6-key closed struct and therefore cannot satisfy this requirement unchanged. The seam's internal representation shall be opened to `map[string]any` (the closed struct may remain for read-only convenience).
- REQ-CRIT-001-002: When `ClaimTask` writes a claim line to the team tasklist ledger (team_spawn.go:316-384), the CLI shall open the ledger with `O_APPEND|O_WRONLY` so the write lands at the ledger tail and the append-only ledger head is never overwritten.
- REQ-CRIT-001-003: When `saveWorkflowMuteConfig` persists a mute change (harness_mute.go:198-228), the CLI shall mutate only the target keys of `workflow.yaml` via the yaml.v3 Node API (reusing the harness.go:363 pattern) so all sibling keys (agentic_loop, team, ...) survive the round-trip.
- REQ-CRIT-001-004: When `RemoveHarness` or `EditHarness` resolves harness artifacts by name (harness/v4lifecycle.go:257,285), the matcher shall use a `prefix+"-"` boundary (or exact-name equality) so operating on harness `release` never touches artifacts of harness `release-update`.
- REQ-CRIT-001-005: When `runUpdate` reaches its destructive steps (clean/deploy/restore), the CLI shall first acquire the update lock via `acquireUpdateLock` (update_cleanup.go:55) and release it on completion; while another update holds the lock, a second invocation shall fail fast with a diagnostic instead of interleaving destructive steps.
- REQ-CRIT-001-006: When the migrate_agency rollback runs after a phase failure (migrate_agency.go:87-95,398,467), the rollback shall restore pre-existing paths from a snapshot taken before mutation and shall delete only paths newly created by the migration; it shall not delete user paths that existed before the migration started.
- REQ-CRIT-001-007: When migrate_agency archives directory entries (migrate_agency.go:446-452), the symlink guard shall detect symlinks via `os.Lstat` (or the existing `isSymlinkEntry` helper) so symlinked entries are skipped rather than dereferenced and copied from outside the tree into `.agency.archived/`.
- REQ-CRIT-001-008: When the Stop-hook auto-classify pipeline evaluates tier promotions (hook.go:773-792,1167-1206), the CLI shall track a per-pattern high-water mark and append a promotion record only when the tier actually changes, so `tier-promotions.jsonl` does not accumulate duplicate records on every session end.
- REQ-CRIT-001-009: The run-phase implementation shall add a failing reproduction test for each of the eight defects before applying the corresponding fix, and each reproduction test shall pass after the fix while the full `internal/cli` suite stays green.

## §C Scope

### In Scope

- The 8 Critical fix sites listed in §B (glm.go, launcher.go, settings.go — the `mutateSettingsLocal` seam representation per REQ-CRIT-001-001, team_spawn.go, harness_mute.go, harness/v4lifecycle.go, update.go + update_cleanup.go lock wiring, migrate_agency.go, hook.go).
- One reproduction test per defect plus regression coverage for the fixed behavior.

### Out of Scope — Broader concurrency hardening

- Routing all 5 `settings.local.json` writers through the locked seam, `~/.claude.json` flock, atomic-writer consolidation, and preference TOCTOU work are deferred to SPEC-CLIFIX-CONCURRENCY-001. This SPEC only stops the closed-struct data loss at the two Critical RMW sites.

### Out of Scope — Contract, linter, and hygiene remediation

- os.Exit removal / exit-code contracts (SPEC-CLIFIX-CONTRACT-001), agentlint/doctor staleness (SPEC-CLIFIX-LINTER-STALE-001), and update.go decomposition, dead-code removal, hardcoding sweep (SPEC-CLIFIX-HYGIENE-001) are excluded here even where they touch the same files.

### Out of Scope — Behavior redesign

- No new features, no CLI surface changes, no config schema changes. Fixes restore intended documented behavior only.

## §D Acceptance Criteria

- AC-CRIT-001-001: Given a `settings.local.json` fixture carrying top-level keys beyond the closed struct (hooks, outputStyle, permissions), When the GLM/CC env mutation round-trips the file, Then every pre-existing top-level key survives the write (maps REQ-CRIT-001-001)
- AC-CRIT-001-002: Given a tasklist ledger with a header and existing task lines, When `ClaimTask` claims a pending task, Then the claim line is appended at the tail and the ledger head bytes are unchanged (maps REQ-CRIT-001-002)
- AC-CRIT-001-003: Given a `workflow.yaml` containing agentic_loop and team sections, When `moai harness mute <domain>` persists the mute, Then all sibling keys survive byte-equivalent except the mute target (maps REQ-CRIT-001-003)
- AC-CRIT-001-004: Given both `release` and `release-update` harness artifacts installed, When `RemoveHarness("release")` runs, Then no `release-update` artifact is deleted (maps REQ-CRIT-001-004)
- AC-CRIT-001-005: Given an update lock held by a first invocation, When a second `runUpdate` starts, Then it fails fast with a lock diagnostic and performs no destructive step (maps REQ-CRIT-001-005)
- AC-CRIT-001-006: Given a pre-existing user directory recorded in the migration plan, When a phase failure triggers rollback, Then the pre-existing directory is restored to its prior content and only migration-created paths are removed (maps REQ-CRIT-001-006)
- AC-CRIT-001-007: Given a symlink entry pointing outside the tree, When migrate_agency archives the directory, Then the symlink is skipped and no out-of-tree content is copied (maps REQ-CRIT-001-007)
- AC-CRIT-001-008: Given a tier-promotions ledger already containing a promotion for a pattern, When the Stop hook classifies the same pattern at the same tier again, Then no duplicate record is appended (maps REQ-CRIT-001-008)
- AC-CRIT-001-009: Given the run-phase commit history, When each fix is inspected, Then a failing reproduction test precedes (or accompanies with RED evidence) the fix and `go test ./internal/cli/... -count=1` passes at completion (maps REQ-CRIT-001-009)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: none. This SPEC is the P0 entry of the CLIFIX series; the four sibling SPECs (CONTRACT, CONCURRENCY, LINTER-STALE, HYGIENE) sequence strictly after it because they share files with it (see plan.md §G overlap matrix).
- Non-goal: performance optimization of the touched paths.
- Non-goal: retroactive cleanup of already-corrupted user files (a one-shot repair tool is not in scope; the fixes stop new corruption).
