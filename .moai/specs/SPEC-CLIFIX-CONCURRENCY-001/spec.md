---
id: SPEC-CLIFIX-CONCURRENCY-001
title: "CLI Concurrency Remediation — re-route writers through locked seam, ~/.claude.json RMW guard, preference store hardening (P2)"
version: "1.0.0"
status: completed
created: 2026-07-10
updated: 2026-07-11
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
| 0.2.0 | 2026-07-11 | manager-spec | Re-grounded against post-CRITICAL-001/CONTRACT-001 HEAD (iter-1 plan-audit FAIL 0.77). D1: fixed 5 stale writer anchors (glm.go:548/638/971; launcher.go:288/685). D2: corrected writeFileAtomic to settings.go:121. D3: acknowledged seam already hardened by CRITICAL-001 — M1 now re-routes callers only. D4: dropped REQ-001-005 (Upsert tier-transition already fixed at filestore.go:96 — no uncovered cross-tier path exists; Get cascade + Query dedupe provide read-side correctness). D5: narrowed REQ-001-002 to removeGLMEnv delete-set only (6-site inject consolidation deferred to HYGIENE-001). D6: completed 4-helper atomic inventory. D7: re-anchored ScanDue/MarkScanned (decay.go:317/340) + toggle TOCTOU (preference/toggle.go). REQ/AC reduced 7/7 → 6/6. |
| 0.2.1 | 2026-07-11 | manager-spec | D1 micro-fix (iter-2 plan-audit PASS 0.89 with this single debt). Corrected atomic-writer inventory under-count (D6 had claimed "completed 4-helper" but missed 2 sites). Full inventory is 6 sites (4 named + 2 inline): atomicWriteJSON (update_noise.go:74), writeFileAtomic (settings.go:121), writeClaudeJSONAtomic (glm_tools.go:437), atomicWrite (preference/filestore.go:473), inline harness_mute.go:250-254, inline glm.go:689-704 (saveLLMSection/llm.yaml). REQ-001-004 + AC-001-004 updated. Scope decision: atomicWrite is option (b) — package-internal exception in internal/cli/preference (separate package; import direction cli→preference is one-way per internal/cli/root.go:14, and preference does NOT import cli back, so a consolidated helper in internal/cli cannot be imported by preference without a cycle; leaf-package extraction is scope expansion beyond this Tier M micro-fix). glm.go:689-704 inline is IN M3 scope (same package internal/cli). AC-004 grep widened to include `func atomicWrite` so it cannot pass vacuously while that helper persists. No other REQ/AC touched. |

## §A Context

Multi-session operation on a shared checkout is this repository's real usage pattern (audit §5 P2: "다중 세션(이젝 이 저장소의 실사용 패턴) 안정화"). The audit found unlocked/non-atomic read-modify-write sites over shared JSON state: five `settings.local.json` writers calling `writeSettingsMap` (plain `os.WriteFile`), the `~/.claude.json` RMW racing a live Claude Code process, and preference-store TOCTOU windows — plus four named atomic-writer helpers that fragment the write discipline.

**Re-grounding against post-prerequisite HEAD.** Both prerequisite SPECs merged AFTER the initial draft, moving every cited line and shipping two of the fixes the 0.1.0 draft claimed to add:

- SPEC-CLIFIX-CRITICAL-001 already hardened `mutateSettingsLocal` (`internal/cli/settings.go:79`) with a sibling-file flock (`lockFile` at settings.go:93) and temp-file+rename publication (`writeFileAtomic` at settings.go:110). This SPEC's remaining work for M1 is to **re-route the 5 `writeSettingsMap` callers** through that EXISTING seam — hardening the seam itself is a no-op.
- SPEC-CLIFIX-CRITICAL-001 also already shipped the Upsert cross-tier fix: `Upsert` at `filestore.go:96` calls `removeFromRecall` BEFORE routing. The 0.1.0 draft's REQ-001-005 targeted that defect; it is now dropped (D4 — see §C Out of Scope for rationale).

Findings SSOT: audit §3 cluster 2 (Major rows 1, 2, 4) and cluster 5 (preference rows), §4 cross-cutting row 5. Re-verify all anchors against the live tree at run time.

## §B Requirements (GEARS)

- REQ-CONC-001-001: The CLI shall route all five non-atomic `settings.local.json` write sites — `internal/cli/glm.go:548` (inside `ensureSettingsLocalJSON`), `glm.go:638` (`injectGLMEnvForTeam`), `glm.go:971` (`injectGLMEnv`), `internal/cli/launcher.go:288` (inside `removeGLMEnv`), `launcher.go:685` (`syncPermissionModeToSettingsLocal`) — which currently call `writeSettingsMap` (plain `os.WriteFile`, no lock, no temp+rename), through the existing locked+atomic `mutateSettingsLocal` seam at `internal/cli/settings.go:79` (already hardened by SPEC-CLIFIX-CRITICAL-001 with `lockFile` at settings.go:93 and `writeFileAtomic` publication at settings.go:110), so concurrent sessions cannot lose updates or truncate the file. The sixth caller at `launcher.go:209` already routes through `mutateSettingsLocal` and is NOT a defect site.
- REQ-CONC-001-002: When `removeGLMEnv` (`internal/cli/launcher.go:244-289`) restores Claude-mode settings, the CLI shall add `CLAUDE_CODE_AUTO_COMPACT_WINDOW` (via `config.EnvClaudeCodeAutoCompactWindow`) to the delete set currently at `launcher.go:268-281`, so a 1M auto-compact window does not persist into subsequent `moai cc` sessions.
- REQ-CONC-001-003: When glm_tools performs a read-modify-write of `~/.claude.json` — `runEnableMCPServerForTool` at `internal/cli/glm_tools.go:503-548` and `enableMCPServerIdempotentForTool` at `glm_tools.go:562-616` — the CLI shall guard the RMW with flock or an mtime compare-and-retry loop around the existing `writeClaudeJSONAtomic` (glm_tools.go:437, which does temp+rename but has NO flock), so concurrent writes from a live Claude Code process are not lost.
- REQ-CONC-001-004: The CLI shall consolidate the six duplicate atomic-writer sites — four named helpers (`writeFileAtomic` at `internal/cli/settings.go:121`, `atomicWriteJSON` at `internal/cli/update_noise.go:74`, `writeClaudeJSONAtomic` at `internal/cli/glm_tools.go:437`, `atomicWrite` at `internal/cli/preference/filestore.go:473`) plus two inline tmp+rename blocks (`internal/cli/harness_mute.go:250-254` and `internal/cli/glm.go:689-704` inside `saveLLMSection` writing `llm.yaml`) — into one shared helper, and all former callers within `internal/cli` shall use the consolidated helper. Scope decision for `atomicWrite` (option b): `atomicWrite` lives in package `internal/cli/preference`, which is a SEPARATE package from `internal/cli`; the import direction is one-way (`internal/cli` imports `internal/cli/preference` at `internal/cli/root.go:14`, and `internal/cli/preference` does NOT import `internal/cli` back), so a consolidated helper placed in `internal/cli` cannot be imported by `internal/cli/preference` without creating an import cycle. Extracting a new leaf utility package that both can import is feasible but is scope expansion beyond this Tier M SPEC; therefore `atomicWrite` is acknowledged as a package-internal exception EXCLUDED from M3 consolidation, but REMAINS listed in the inventory so AC-CONC-001-004's grep still observes it. The `glm.go:689-704` inline is IN M3 scope (same package `internal/cli`). The consolidated helper MUST preserve the union of each former caller's guarantees (file-mode / fsync / rename semantics — DDD PRESERVE, do not silently drop an fsync a caller relied on).
- REQ-CONC-001-005: While concurrent scans and toggles operate on the preference store, the store shall remain crash-consistent and TOCTOU-hardened: `DecayScan` (`internal/cli/preference/decay.go:142-189`) archival (`writeArchivalEntry` at decay.go:169) and recall writes (`writeRecall` at decay.go:185) shall be transactional or reconciled at load, and the `ScanDue` (decay.go:317) / `MarkScanned` (decay.go:340, plain `os.WriteFile` defect at decay.go:346) and toggle read-then-flip windows (`newToggleCmd` at `internal/cli/preference/toggle.go:172`, `runToggle` at toggle.go:204, sentinel write at toggle.go:106) shall be closed via locking or single-file atomic swap.
- REQ-CONC-001-006: The run-phase implementation shall verify every concurrency fix with a reproduction test that fails on the pre-fix code, and the affected packages (`internal/cli/...`, `internal/cli/preference/...`) shall pass `go test -race`.

## §C Scope

### In Scope

- The five `settings.local.json` writers (re-route through existing seam), `removeGLMEnv` delete-set key addition, `~/.claude.json` RMW guard, atomic-writer consolidation, preference store DecayScan crash-consistency + ScanDue/MarkScanned/toggle TOCTOU hardening, race verification.

### Out of Scope — Critical closed-struct fixes

- The SettingsLocal struct data-loss fix (glm.go/launcher.go closed-struct field set) belongs to SPEC-CLIFIX-CRITICAL-001 and is a prerequisite: this SPEC routes writers through the seam that CRITICAL-001 established and validated.

### Out of Scope — Update-command locking

- `moai update` lock wiring is SPEC-CLIFIX-CRITICAL-001 territory (already covered by REQ-CRIT-001-005); this SPEC does not touch update.go.

### Out of Scope — Env-name constant sweep

- Replacing inline GLM env-name literals with `envkeys.go` constants, consolidating the 6 scattered `CLAUDE_CODE_AUTO_COMPACT_WINDOW` inject sites (`glm.go:202,425,503,623,919,963`) with their partial self-cleanup at `glm.go:625`/`:965` into a single SSOT set, and a shared `glmEnvVarSet()` helper, are SPEC-CLIFIX-HYGIENE-001 work. This SPEC's REQ-CONC-001-002 corrects the removeGLMEnv delete-set key membership ONLY, at its current representation — scope is kept tight so the two SPECs do not overlap.

### Out of Scope — Upsert tier-transition (already fixed by CRITICAL-001)

- The 0.1.0 draft's REQ-001-005 targeted a cross-tier stale-copy defect in `Upsert`. Re-investigation against HEAD confirmed `filestore.go:96` already calls `removeFromRecall` BEFORE routing (write-side fix, godoc at filestore.go:93-95 states the anti-stale-copy intent verbatim). The read side is also covered: `Get` (filestore.go:117-137) implements a core-first cascade that never touches recall on a core hit, and `Query` (filestore.go:143-183) carries a `seen` dedupe map ("dedupe by decision_key across tiers") at filestore.go:147. No genuinely-uncovered cross-tier path exists, so REQ-001-005 is dropped (not re-scoped) to avoid a vacuous AC. REQ/AC reduced 7/7 → 6/6 with REQ-001-006/007 renumbered to 005/006.

## §D Acceptance Criteria

- AC-CONC-001-001: Given two concurrent settings.local.json mutations, When both complete, Then the file contains both updates or a serialized last-writer state with no lost keys and no truncation (maps REQ-CONC-001-001)
- AC-CONC-001-002: Given a GLM session that injected CLAUDE_CODE_AUTO_COMPACT_WINDOW, When removeGLMEnv restores Claude mode, Then the key is absent from settings.local.json env (maps REQ-CONC-001-002)
- AC-CONC-001-003: Given a simulated concurrent writer mutating ~/.claude.json between read and write, When the guarded RMW runs, Then the concurrent change is not lost (retry or lock observed) (maps REQ-CONC-001-003)
- AC-CONC-001-004: Given the internal/cli tree, When counting atomic-writer helper implementations, Then exactly one shared implementation remains and all call sites use it (maps REQ-CONC-001-004)
- AC-CONC-001-005: Given an induced crash between DecayScan archival and recall writes, When the store is next loaded, Then no duplicate entry survives reconciliation; and concurrent ScanDue/toggle operations produce no lost update under the race detector (maps REQ-CONC-001-005)
- AC-CONC-001-006: Given the affected packages, When `go test -race` runs, Then all tests pass including the new reproduction tests (maps REQ-CONC-001-006)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001 (shipped the `mutateSettingsLocal` locked+atomic seam at settings.go:79-115 AND the Upsert `removeFromRecall` cross-tier fix at filestore.go:96), SPEC-CLIFIX-CONTRACT-001 (P0→P1→P2 series order; flock test enablement from CONTRACT-001 provides the running lock-test harness this SPEC extends).
- Non-goal: cross-process advisory-lock standardization for every file in `.moai/state/` — only the audited sites are hardened.
- Non-goal: Windows file-locking parity beyond graceful fallback (mtime-retry path acceptable where flock is unavailable).
- Non-goal: re-hardening `mutateSettingsLocal` — that work is complete (CRITICAL-001). This SPEC re-routes callers through the existing seam only.
