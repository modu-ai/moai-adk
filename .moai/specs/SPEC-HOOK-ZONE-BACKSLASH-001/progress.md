# SPEC-HOOK-ZONE-BACKSLASH-001 — Progress

## Plan-phase Seed (2026-10-07)

- **Card**: t1566 (security P1, measured/demonstrated). Factory lane card; plan →
  plan-audit → run → sync.
- **Base**: `f97edcc55` (branch `WT-protected-zone-backslash`, HEAD == origin/main
  tip at tree entry; working tree clean).
- **SPEC**: SPEC-HOOK-ZONE-BACKSLASH-001 — protected-zone target resolution preserves
  POSIX component identity (`zoneSlash`, `internal/hook/protected_zone_path.go:35`-`:37`).
- **Card mandate**: RED reproduction FIRST (M1), then minimal repair of the named
  instance ONLY (M2), then a 3-surface family re-check with read + test evidence
  (M3). Repairs of the sibling surfaces belong to their own cards — this card adds
  no repair to `pre_tool.go` (t1556) or `internal/cli/worktree/landing_predicate.go`
  (t1561).
- **Mechanism hypothesis** (spec.md §B; the M1 RED test pins the truth):
  `zoneSlash` rewrites `\`→`/` unconditionally before any filesystem step, so a
  backslash-named symlink component loses its identity; the walk resolves a
  fictional spelling, both arms classify outside `zone_dir/`, the guard allows —
  while the OS follows the literal symlink and lands the write inside the protected
  zone.
- **Motivating evidence (measured elsewhere, re-measured here at M1)**: round 8 of
  the card-t1556 codex review gate (2026-10-07, delta commit `9d78421a4`) — an
  actual harness-learner Write call through a backslash-named link received
  `decision="allow"` and the protected file recorded `"changed"`. Primary copy:
  `.moai/reports/t1556/codex-review-gate-1.md` § 라운드 8 (t1556 card worktree —
  disposable tree; durable landing is the t1556 PR).
- **Plan-phase timestamp**: 2026-10-07 (SPEC authored by manager-spec in the card
  worktree; artifacts: spec.md / plan.md / acceptance.md / progress.md).

## §E.1 Plan-phase Audit-Ready Signal

- **plan_status: audit-ready**
- **plan_complete_at: 2026-10-07**
- **Verdict**: PASS — 0.91 / 1.00 (Tier M threshold 0.80), blocking findings 0.
- **Verdict file**: `.moai/reports/t1566/plan-audit-1.md` (plan-auditor, independent,
  tree `f97edcc55`, branch `WT-protected-zone-backslash`, 2026-10-07).
- **Audited artifact set**: spec.md · plan.md · acceptance.md · progress.md (as of
  2026-10-07, tree `f97edcc55`).
- **Auditor instructions carried into the run phase** (non-blocking findings of the
  verdict, binding on M1/M3 execution):
  1. **Positive-control row at M1** (verdict non-blocking #2): when authoring the
     reproduction test, add one positive-control row — a file inside a literal
     `lnk\dir` ORDINARY directory (non-symlink) outside the zone remains ALLOWED.
     This closes the character-blacklist mutant gap (a mutant that repairs the
     resolver then blanket-denies all backslash-bearing POSIX paths would otherwise
     pass the whole AC set; existing tests carry no POSIX literal-backslash
     filename-component case).
  2. **Reinforcement evidence rows from gate-turnend-1/-2** (verdict non-blocking
     #3): at M1 execution, add `.moai/reports/t1566/gate-turnend-1.md` and
     `.moai/reports/t1566/gate-turnend-2.md` as reinforcement corroboration rows in
     the acceptance.md evidence ledger — these are same-tree (`f97edcc55`)
     independent reproductions of the defect class, stronger than the round-8
     cross-tree record.
  - Also noted (advisory): M3's surface-2 row records OBSERVATION, not narrative —
    if the sweep finds no backslash-rewrite pattern at
    `landing_predicate.go:161` (t1561's ledgered defect is the patch-id whitespace
    class), record exactly that; AC-HZB-005 requires observed state + owner +
    evidence, never class agreement.

## M1 — RED Reproduction (run phase)

- **Executed**: 2026-10-07, tree `da2d74eef` (branch `WT-protected-zone-backslash`;
  code-identical to the pinned base `f97edcc55` — `git show --stat da2d74eef`
  verifies the commit carries only the four SPEC artifact files, parent
  `f97edcc55`).
- **Test authored**: `internal/hook/protected_zone_backslash_repro_test.go` —
  `TestCheckProtectedZonePosixBackslashLinkBypass` (guard level, table-driven:
  absolute + relative raw path shapes; non-deny branch performs the ACTUAL write
  through the literal `lnk\dir` path and reads it back from `zone_dir/`) and
  `TestResolveZoneTargetPosixBackslashLinkDivergence` (resolver level). Both
  POSIX-gated (`runtime.GOOS` check) with graceful symlink-unavailable skips;
  `t.TempDir()`-based roots throughout.
- **Auditor non-blocking #2 honored**: the positive-control row is in — a literal
  `lnk\dir` ORDINARY directory (non-symlink) outside the zone stays ALLOWED, at
  both guard level and resolver level. It passed in the same RED run, proving the
  RED is the bypass mechanism and closing the character-blacklist mutant gap.
- **Auditor non-blocking #3 honored**: reinforcement rows `RED-HZB-X1`
  (gate-turnend-1) and `RED-HZB-X2` (gate-turnend-2) appended to acceptance.md's
  evidence ledger — same-tree (`f97edcc55`) independent codex-gate reproductions
  of the defect class, cited as corroboration under the binding RED cells.
- **Four-element RED observations** (full raw bytes in acceptance.md § Evidence
  Ledger):
  - `RED-HZB-001` — `go test -count=1 ./internal/hook/ -run
    'TestCheckProtectedZonePosixBackslashLinkBypass'` → exit 1; both shapes:
    `BYPASS — decision="allow" ... the write through the literal backslash link
    landed INSIDE the protected zone (zone_dir/secret.md="bypass")`.
  - `RED-HZB-002` — same run; the verbatim failure line IS the in-zone landing
    observation (allow + protected file changed — the measured round-8 shape).
  - `RED-HZB-003` — `go test -count=1 ./internal/hook/ -run
    'TestResolveZoneTargetPosixBackslashLinkDivergence'` → exit 1; resolver
    returns ONLY `lnk/dir/secret.md` (zero in-zone forms).
- **Measured note (RED pins the truth over the §B narrative)**: the §B hypothesis
  guessed the fictional rejoin form would be `dir/secret.md`; the measured rejoin
  starts at the FIRST missing component (`lnk`), yielding the single fictional
  form `lnk/dir/secret.md`. The core mechanism (rewritten spelling misses the
  real link; both arms classify outside the zone; allow; the OS write lands
  inside the zone) is confirmed exactly.
- **Affected-package pre-change baseline (M1-BASELINE)**: `go test -count=1
  -timeout 30m ./internal/hook/` on `da2d74eef` → exit 1 (804.0s). Failing set:
  `TestAstgrepCorpusRunDoesNotSkip` (120.01s), `TestStaleRunNoticeLegacyLeaderSpelling`,
  `TestStaleRunNoticeLegacySessionRecord`, `TestStaleRunNoticeFactoryLegacyLabel`
  — all pre-existing, unrelated to the zone path (environmental corpus + legacy
  stale-run-notice spelling) — PLUS the two new RED tests above (expected RED).
  The AC-HZB-004 post-fix verdict is the countable delta against this set: the
  only allowed post-M2 failures are exactly these four pre-existing tests.
- **Pre-change baseline raw tail (verbatim)**:

```
--- FAIL: TestAstgrepCorpusRunDoesNotSkip (120.01s)
--- FAIL: TestCheckProtectedZonePosixBackslashLinkBypass (0.23s)
--- FAIL: TestResolveZoneTargetPosixBackslashLinkDivergence (0.02s)
--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.06s)
--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.08s)
--- FAIL: TestStaleRunNoticeFactoryLegacyLabel (0.59s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	804.025s
FAIL
```

## M2 — Minimal Repair (run phase)

*(populated by the implementer.)*

## M3 — Family Re-check Sweep (run phase)

*(populated by the implementer — the 3-surface table with per-surface state, owner,
evidence; the repo-wide grep with its swept count and per-hit classification; the
package re-measurement verdict against the M1 baseline.)*

## §E — Self-Verification Evidence (run phase)

*(populated by the implementer — E1-E7 of plan.md §E.)*
