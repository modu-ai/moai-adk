---
id: SPEC-CLIFIX-CONCURRENCY-001
title: "CLI Concurrency Remediation — locked settings seam, ~/.claude.json RMW guard, preference store hardening (P2)"
version: "0.1.0"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P2
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, audit-remediation, concurrency, atomic-write, p2"
era: V3R6
tier: M
depends_on: [SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001]
---

# SPEC-CLIFIX-CONCURRENCY-001 — CLI Concurrency Remediation (P2)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §3 cluster 2/5 + §4 row "비원자/무락 공유 상태 RMW" + §5 P2 roadmap row |

## §A Context

Multi-session operation on a shared checkout is this repository's real usage pattern (audit §5 P2: "다중 세션(이 저장소의 실사용 패턴) 안정화"). The audit found 8 unlocked/non-atomic read-modify-write sites over shared JSON state: five `settings.local.json` writers, the `~/.claude.json` RMW racing a live Claude Code process, and preference-store TOCTOU windows — plus four named atomic-writer helpers (and at least one inline tmp+rename) that fragment the write discipline. SPEC-CLIFIX-CRITICAL-001 stopped the closed-struct data loss at two sites; this SPEC completes the concurrency hardening.

Findings SSOT: audit §3 cluster 2 (Major rows 1, 2, 4) and cluster 5 (preference rows), §4 cross-cutting row 5. Re-verify all anchors against the live tree at run time.

## §B Requirements (GEARS)

- REQ-CONC-001-001: The CLI shall route all five `settings.local.json` write sites (glm.go:562,667,1015; launcher.go:296,705) through a single locked and atomic `mutateSettingsLocal` seam (read-modify-write under a file lock, temp-file+rename publication), so concurrent sessions cannot lose updates or truncate the file.
- REQ-CONC-001-002: When `removeGLMEnv` restores Claude-mode settings (launcher.go:241-301), it shall remove `CLAUDE_CODE_AUTO_COMPACT_WINDOW` together with the other GLM-injected env keys, so a 1M auto-compact window does not persist into subsequent `moai cc` sessions.
- REQ-CONC-001-003: When glm_tools performs a read-modify-write of `~/.claude.json` (glm_tools.go:496-615), the CLI shall guard the RMW with flock or an mtime compare-and-retry loop, so concurrent writes from a live Claude Code process are not lost.
- REQ-CONC-001-004: The CLI shall consolidate the duplicate atomic-writer helpers — the verified inventory: `writeFileAtomic` (settings.go:79), `atomicWriteJSON` (update_noise.go), `writeClaudeJSONAtomic` (glm_tools.go:437), plus the inline tmp+rename in harness_mute.go — into one shared helper, and all former callers shall use the consolidated helper.
- REQ-CONC-001-005: When preference `Upsert` transitions an entry between tiers (preference/filestore.go:78-113), it shall remove the stale copy from the previous tier file in the same operation, so the core-first read cascade never returns the outdated value.
- REQ-CONC-001-006: While concurrent scans and toggles operate on the preference store, the store shall remain crash-consistent and TOCTOU-hardened: `DecayScan` archival and recall writes shall be transactional or reconciled at load (preference/decay.go:142-189), and the `ScanDue`/`MarkScanned` and toggle read-then-flip windows shall be closed via locking or single-file atomic swap.
- REQ-CONC-001-007: The run-phase implementation shall verify every concurrency fix with a reproduction test that fails on the pre-fix code, and the affected packages shall pass `go test -race`.

## §C Scope

### In Scope

- The five settings.local.json writers, removeGLMEnv key set, ~/.claude.json RMW guard, atomic-writer consolidation, preference filestore/decay/scan/toggle hardening, race verification.

### Out of Scope — Critical closed-struct fixes

- The SettingsLocal struct data-loss fix (glm.go:98, launcher.go:241) belongs to SPEC-CLIFIX-CRITICAL-001 and is a prerequisite: this SPEC routes writers through the seam that CRITICAL-001 establishes/validates.

### Out of Scope — Update-command locking

- `moai update` lock wiring is SPEC-CLIFIX-CRITICAL-001 territory (already covered by REQ-CRIT-001-005); this SPEC does not touch update.go.

### Out of Scope — Env-name constant sweep

- Replacing inline GLM env-name literals with envkeys.go constants and a shared glmEnvVarSet() is SPEC-CLIFIX-HYGIENE-001 work; here the removeGLMEnv fix corrects the key SET membership only, at its current representation.

## §D Acceptance Criteria

- AC-CONC-001-001: Given two concurrent settings.local.json mutations, When both complete, Then the file contains both updates or a serialized last-writer state with no lost keys and no truncation (maps REQ-CONC-001-001)
- AC-CONC-001-002: Given a GLM session that injected CLAUDE_CODE_AUTO_COMPACT_WINDOW, When removeGLMEnv restores Claude mode, Then the key is absent from settings.local.json env (maps REQ-CONC-001-002)
- AC-CONC-001-003: Given a simulated concurrent writer mutating ~/.claude.json between read and write, When the guarded RMW runs, Then the concurrent change is not lost (retry or lock observed) (maps REQ-CONC-001-003)
- AC-CONC-001-004: Given the internal/cli tree, When counting atomic-writer helper implementations, Then exactly one shared implementation remains and all call sites use it (maps REQ-CONC-001-004)
- AC-CONC-001-005: Given a preference entry upserted from tier A to tier B, When the core-first cascade reads the key, Then it returns the tier-B value and the tier-A file no longer contains the entry (maps REQ-CONC-001-005)
- AC-CONC-001-006: Given an induced crash between DecayScan archival and recall writes, When the store is next loaded, Then no duplicate entry survives reconciliation; and concurrent ScanDue/toggle operations produce no lost update under the race detector (maps REQ-CONC-001-006)
- AC-CONC-001-007: Given the affected packages, When `go test -race` runs, Then all tests pass including the new reproduction tests (maps REQ-CONC-001-007)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001 (seam base at glm.go/launcher.go), SPEC-CLIFIX-CONTRACT-001 (P0→P1→P2 series order; flock test enablement from CONTRACT-001 REQ-CONT-001-007 provides the running lock-test harness this SPEC extends).
- Non-goal: cross-process advisory-lock standardization for every file in `.moai/state/` — only the audited sites are hardened.
- Non-goal: Windows file-locking parity beyond graceful fallback (mtime-retry path acceptable where flock is unavailable).
