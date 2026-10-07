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

*(populated by the implementer — the four-element RED observations for
`RED-HZB-001`/`RED-HZB-002`/`RED-HZB-003` land here and in acceptance.md's evidence
ledger, plus the affected-package pre-change baseline.)*

## M2 — Minimal Repair (run phase)

*(populated by the implementer.)*

## M3 — Family Re-check Sweep (run phase)

*(populated by the implementer — the 3-surface table with per-surface state, owner,
evidence; the repo-wide grep with its swept count and per-hit classification; the
package re-measurement verdict against the M1 baseline.)*

## §E — Self-Verification Evidence (run phase)

*(populated by the implementer — E1-E7 of plan.md §E.)*
