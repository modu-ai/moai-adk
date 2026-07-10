# SPEC-CLIFIX-CONCURRENCY-001 — Implementation Plan

## §A Context

P2 row of the CLI audit roadmap: multi-session stability. The dominant defect family is non-atomic, unlocked RMW over shared JSON state (audit cluster 2 headline). CRITICAL-001 established the open-representation seam; this SPEC generalizes it to every writer and hardens the preference store.

## §B Known Issues (findings inventory)

| # | File anchor (re-verify before edit) | Defect | Fix direction |
|---|---|---|---|
| 1 | glm.go:562,667,1015; launcher.go:296,705 | 5 unlocked, non-atomic settings.local.json writers — lost updates/truncation across parallel sessions | route all through locked+atomic mutateSettingsLocal |
| 2 | launcher.go:241-301 | removeGLMEnv leaves CLAUDE_CODE_AUTO_COMPACT_WINDOW → 1M window persists in cc mode | add key to removal set (single SSOT set with inject side) |
| 3 | glm_tools.go:496-615 | ~/.claude.json RMW races live Claude Code process; plaintext key also lands in backups (related minor) | flock or mtime compare-and-retry |
| 4 | 3 sites (audit §4 row 5) | 3 duplicate atomic-writer helpers | consolidate to one |
| 5 | preference/filestore.go:78-113 | Upsert tier transition leaves stale copy in old tier → core-first cascade returns old value forever | remove old-tier entry in same op |
| 6 | preference/decay.go:142-189 | DecayScan archival-then-recall not crash-consistent → duplicates | transactionalize or reconcile-at-load |
| 7 | preference ScanDue/MarkScanned + toggle | TOCTOU windows; --memory-dir consumes shared 24h gate (related minor) | lock or atomic single-file swap |

## §C Pre-flight

1. Confirm CRITICAL-001 + CONTRACT-001 merged (glm.go/launcher.go bases moved; flock test running).
2. Locate the 3 atomic-writer helpers (`grep -rn 'rename\|\.tmp' internal/cli --include='*.go'` refine) and diff their semantics (fsync? perms preservation?) before consolidation — the union of guarantees wins.
3. Decide lock primitive: reuse the flock helper family proven by team_spawn lock; define Windows fallback (mtime-retry).
4. Baseline `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1`.

## §D Constraints

- File format and key layout of settings.local.json / ~/.claude.json unchanged — only write discipline changes.
- The consolidated atomic writer must preserve each caller's file-mode expectations (0600 for credential-bearing files).
- ~/.claude.json guard must never block a live Claude Code process for a human-noticeable duration (bounded retry, fail-open with warning after ceiling).
- No dependency additions; use stdlib + existing helpers.

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md.
- E2: `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` verbatim.
- E3: coverage of preference package not below baseline.
- E4: `grep` audit for atomic-writer duplicates = 1 implementation.
- E5: `golangci-lint run` no new findings.

## §F Milestones (priority order)

- M1 — Seam + writers: harden mutateSettingsLocal (lock + temp/rename), route the 5 writers, removeGLMEnv key-set fix (repro tests first: parallel-writer test, leftover-key test).
- M2 — ~/.claude.json guard: flock/mtime-retry with bounded backoff; concurrent-writer repro test.
- M3 — Atomic-writer consolidation: single helper, migrate all callers, delete duplicates.
- M4 — Preference store: Upsert tier-transition fix, DecayScan crash-consistency, ScanDue/toggle TOCTOU; `-race` full pass; §E self-verification.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4; this SPEC is third. Shared-file overlap: glm.go/launcher.go (CRITICAL-001 a; HYGIENE-001 env-constant sweep follows later and must rebase on the seam), glm_tools.go (HYGIENE-001 env constants).
- Anti-pattern: adding a second lock file convention — reuse one lock helper for both settings.local.json and ~/.claude.json guards.
- Anti-pattern: sleeping in tests to "win" races — use synchronization points/injected hooks so race tests are deterministic.
- Risk: flock on network filesystems is unreliable — document the mtime-retry fallback as the portable path.
- Risk: consolidating atomic writers can silently drop an fsync a caller relied on — characterize each helper before deletion (DDD PRESERVE).

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 2/5, §4 row 5, §5 P2.
- Depends on: SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001. Followed by: SPEC-CLIFIX-LINTER-STALE-001.
- CLAUDE.local.md §2 settings.local.json separation; §22 dev settings intent (keys this seam must never drop).
- agent-common-protocol.md § Pre-Spawn Sync Check (multi-session race context motivating P2).
