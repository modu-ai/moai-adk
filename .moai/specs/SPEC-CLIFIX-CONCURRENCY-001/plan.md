# SPEC-CLIFIX-CONCURRENCY-001 — Implementation Plan

## §A Context

P2 row of the CLI audit roadmap: multi-session stability. The dominant defect family is non-atomic, unlocked RMW over shared JSON state (audit cluster 2 headline). CRITICAL-001 shipped the locked+atomic `mutateSettingsLocal` seam (settings.go:79, flock at :93, `writeFileAtomic` publication at :110) AND the Upsert `removeFromRecall` cross-tier fix (filestore.go:96). This SPEC generalizes the seam to the remaining writers, guards the `~/.claude.json` RMW, consolidates atomic-writer helpers, and hardens the preference store's DecayScan/ScanDue/toggle paths.

**Re-grounded against post-CRITICAL-001/CONTRACT-001 HEAD (iter-1 plan-audit FAIL 0.77 → iter-2).** All anchors in this plan were re-verified by grep against HEAD on 2026-07-11.

## §B Known Issues (findings inventory — anchors re-verified 2026-07-11)

| # | File anchor (re-verify before edit) | Defect | Fix direction |
|---|---|---|---|
| 1 | `glm.go:548` (ensureSettingsLocalJSON), `glm.go:638` (injectGLMEnvForTeam), `glm.go:971` (injectGLMEnv), `launcher.go:288` (removeGLMEnv), `launcher.go:685` (syncPermissionModeToSettingsLocal) | 5 unlocked, non-atomic settings.local.json writers — all call `writeSettingsMap` (plain `os.WriteFile`, no lock, no temp+rename); lost updates/truncation across parallel sessions. Note: `launcher.go:209` already routes through `mutateSettingsLocal` and is NOT a defect site. | re-route all 5 through existing locked+atomic `mutateSettingsLocal` (settings.go:79, already hardened by CRITICAL-001 with lockFile :93 + writeFileAtomic :110). Do NOT re-harden the seam — that is a no-op. |
| 2 | `launcher.go:244-289` (removeGLMEnv delete-set at `:268-281`) | removeGLMEnv leaves `CLAUDE_CODE_AUTO_COMPACT_WINDOW` (`config.EnvClaudeCodeAutoCompactWindow`) → 1M window persists in cc mode. The 6 inject sites (`glm.go:202,425,503,623,919,963`) with partial self-cleanup at `glm.go:625`/`:965` are a separate consolidation concern. | add key to the removeGLMEnv delete set ONLY. The 6-site inject consolidation + envkeys.go constant sweep is deferred to SPEC-CLIFIX-HYGIENE-001 (scope kept tight — this SPEC corrects key membership at its current representation). |
| 3 | `glm_tools.go:503-548` (runEnableMCPServerForTool), `glm_tools.go:562-616` (enableMCPServerIdempotentForTool) | `~/.claude.json` RMW (`readClaudeJSON` → mutate → `writeClaudeJSONAtomic` at glm_tools.go:437) races a live Claude Code process; `writeClaudeJSONAtomic` does temp+rename but has NO flock. | flock or mtime compare-and-retry around the RMW. |
| 4 | 6 sites (verified inventory, corrected 0.2.1): 4 named — `atomicWriteJSON` (update_noise.go:74), `writeFileAtomic` (settings.go:121), `writeClaudeJSONAtomic` (glm_tools.go:437), `atomicWrite` (preference/filestore.go:473); 2 inline — harness_mute.go:250-254, glm.go:689-704 (saveLLMSection/llm.yaml) | 6 duplicate atomic-writer sites with slightly varying semantics (perm, fsync). | consolidate the 5 sites in `internal/cli` (3 named + 2 inline) to one shared helper; `atomicWrite` (preference/filestore.go:473) is option (b) — package-internal exception in `internal/cli/preference` (import direction cli→preference is one-way per root.go:14; leaf-package extraction out of scope for this Tier M SPEC), EXCLUDED from M3 consolidation but observed by AC-004 grep. Characterize each before deletion (DDD PRESERVE — do not silently drop an fsync a caller relied on). |
| 5 | `preference/decay.go:142-189` (DecayScan: writeArchivalEntry `:169` + writeRecall `:185` as two separate writes); `preference/decay.go:317` (ScanDue) / `:340` (MarkScanned, plain `os.WriteFile` defect at `:346`); `preference/toggle.go:172` (newToggleCmd) / `:204` (runToggle) / `:106` (sentinel os.WriteFile) | DecayScan archival-then-recall not crash-consistent (crash between the two writes → duplicate or lost entry); ScanDue/MarkScanned TOCTOU window; toggle read-then-flip TOCTOU window. | DecayScan: transactionalize (single atomic multi-file commit) or reconcile-at-load (dedupe on next load). ScanDue/MarkScanned + toggle sentinel: lock or atomic single-file swap. |

## §C Pre-flight

1. Confirm CRITICAL-001 + CONTRACT-001 merged (the `mutateSettingsLocal` seam at settings.go:79-115 is present with `lockFile` at :93 + `writeFileAtomic` at :110; `removeFromRecall` call is present at filestore.go:96; flock test harness is running).
2. Locate the 6 atomic-writer sites (verified inventory, corrected 0.2.1: update_noise.go:74 atomicWriteJSON, settings.go:121 writeFileAtomic, glm_tools.go:437 writeClaudeJSONAtomic, preference/filestore.go:473 atomicWrite, harness_mute.go:250-254 inline, glm.go:689-704 inline saveLLMSection) and diff their semantics (fsync? perms preservation? rename vs write-in-place?) before consolidation — the union of guarantees wins. M3 consolidates the 5 sites in `internal/cli`; `atomicWrite` (preference/filestore.go:473) is option (b) — package-internal exception, EXCLUDED from M3 but observed by AC-004 grep.
3. Decide lock primitive: reuse the flock helper family proven by CRITICAL-001's team_spawn lock (`lockFile`/`unlockFile` already declared with `team_spawn_lock_{unix,windows}.go` build tags); define Windows fallback (mtime-retry).
4. Baseline `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1`.

## §D Constraints

- File format and key layout of settings.local.json / ~/.claude.json unchanged — only write discipline changes.
- The consolidated atomic writer must preserve each caller's file-mode expectations (0600 for credential-bearing files).
- ~/.claude.json guard must never block a live Claude Code process for a human-noticeable duration (bounded retry, fail-open with warning after ceiling).
- No dependency additions; use stdlib + existing helpers.
- Do NOT re-harden `mutateSettingsLocal` — that work is complete (CRITICAL-001). M1 re-routes callers through the existing seam only.

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md.
- E2: `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` verbatim.
- E3: coverage of preference package not below baseline.
- E4: `grep -rn 'func writeFileAtomic\|func atomicWriteJSON\|func writeClaudeJSONAtomic\|func atomicWrite' internal/cli --include='*.go' | grep -v _test.go` — exactly 2 matches remain: 1 consolidated helper in `internal/cli` + 1 acknowledged package-internal exception `func atomicWrite` at `internal/cli/preference/filestore.go:473` (option b).
- E5: `golangci-lint run` no new findings.

## §F Milestones (priority order)

- M1 — Re-route writers + removeGLMEnv key-set: re-route the 5 `writeSettingsMap` callers (glm.go:548/638/971, launcher.go:288/685) through the EXISTING `mutateSettingsLocal` seam (settings.go:79); add `CLAUDE_CODE_AUTO_COMPACT_WINDOW` to the removeGLMEnv delete set (launcher.go:268-281). DROP "harden mutateSettingsLocal" — the seam is already locked+atomic per CRITICAL-001. Repro tests first: parallel-writer test (N goroutines, no lost keys, no truncation), leftover-key test.
- M2 — `~/.claude.json` guard: flock/mtime-retry with bounded backoff around the RMW at glm_tools.go:503-548 and :562-616 (writeClaudeJSONAtomic at :437 does temp+rename but lacks flock); concurrent-writer repro test.
- M3 — Atomic-writer consolidation: single helper, migrate all callers from the 5-site inventory in `internal/cli` (update_noise.go:74, settings.go:121, glm_tools.go:437, harness_mute.go:250-254 inline, glm.go:689-704 inline saveLLMSection), delete duplicates. `atomicWrite` (preference/filestore.go:473) is option (b) — EXCLUDED from M3 (package-internal exception in `internal/cli/preference`; cannot import `internal/cli` without cycle), but remains listed in the inventory and observed by AC-004 grep.
- M4 — Preference store: DecayScan crash-consistency (decay.go:142-189, two-write non-transactional), ScanDue/MarkScanned TOCTOU (decay.go:317/340, defect at :346), toggle TOCTOU (preference/toggle.go:172/204/106); `-race` full pass; §E self-verification.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4; this SPEC is third. Shared-file overlap: glm.go/launcher.go (CRITICAL-001 shipped; HYGIENE-001 env-constant sweep follows later and must rebase on the seam), glm_tools.go (HYGIENE-001 env constants).
- Anti-pattern: re-hardening `mutateSettingsLocal` — the seam is already locked+atomic (CRITICAL-001). M1 re-routes callers only; adding a second lock layer would double-lock and deadlock.
- Anti-pattern: adding a second lock file convention — reuse one lock helper family for both settings.local.json and ~/.claude.json guards.
- Anti-pattern: sleeping in tests to "win" races — use synchronization points/injected hooks so race tests are deterministic.
- Risk: flock on network filesystems is unreliable — document the mtime-retry fallback as the portable path.
- Risk: consolidating atomic writers can silently drop an fsync a caller relied on — characterize each of the 4 helpers before deletion (DDD PRESERVE).

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 2/5, §4 row 5, §5 P2.
- Depends on: SPEC-CLIFIX-CRITICAL-001 (shipped seam + Upsert fix), SPEC-CLIFIX-CONTRACT-001. Followed by: SPEC-CLIFIX-LINTER-STALE-001, SPEC-CLIFIX-HYGIENE-001 (6-site inject consolidation + envkeys.go sweep).
- CLAUDE.local.md §2 settings.local.json separation; §22 dev settings intent (keys this seam must never drop).
- agent-common-protocol.md § Pre-Spawn Sync Check (multi-session race context motivating P2).
