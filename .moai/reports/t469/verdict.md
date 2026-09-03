# Card t469 — Sync-Phase Completion Verdict

Card: t469 — AC-HWD-015 strip-aware mirror amendment (SPEC-ACHWD-STRIP-EXEMPT-001 wrapper over SPEC-HOOK-WIRING-DRIFT-001 v0.4.0)
Branch: `WT-achwd-strip-exempt` · Sync commit: `<sync-commit-sha>` (backfilled below)
Date: 2026-09-03

## Claim

1. SPEC-ACHWD-STRIP-EXEMPT-001 is closed (`in-progress → completed`) on a single
   sync commit carrying the 3-phase close.
2. SPEC-HOOK-WIRING-DRIFT-001 is re-closed to `completed` (v0.4.0 amendment
   re-close, the path its `## Amendments` `re_close_path` row declared). No body
   content and no other frontmatter field was touched — the record was already
   written consistently by the amendment commit `820db6cf9`.
3. CHANGELOG carries one `[Unreleased]` Added row for the wrapper SPEC, placed
   directly above the t216 row it amends; no second row for the re-close.
4. The §I.4/AC-HWD-015 strip-aware mirror check exits 0 on the final tree.
5. Both spec lints report 0 errors on the final tree.

## Evidence

All commands run from worktree `t469`, recorded verbatim in
`.moai/reports/t469/sync-evidence/` (exported, tracked path):

| Check | Command | Observed | File |
|---|---|---|---|
| wrapper lint | `go run ./cmd/moai spec lint .moai/specs/SPEC-ACHWD-STRIP-EXEMPT-001/spec.md` | `0 error(s), 4 warning(s)` (1 `ModalityMalformed` on REQ-ASE-001 + 3 `CoverageIncomplete`, exit 0) | `sync-evidence/lint-wrapper.txt` |
| HWD lint | `go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md` | `0 error(s), 18 warning(s)` (4 `ModalityMalformed` + 14 `CoverageIncomplete`, exit 0) — identical composition to the run-phase A5b measurement, both classes pre-existing | `sync-evidence/lint-hwd.txt` |
| mirror check | AC-HWD-015 perl command (program extracted byte-for-byte to `sync-evidence/t469-mirror-check.pl`; the inline one-liner form was refused by the worktree session guard as unverifiable) | no output, `exit=0` | `sync-evidence/mirror-check.txt` |

Wrapper warning profile note (not smoothed): the wrapper SPEC has its own
warning profile — 1 `ModalityMalformed` (REQ-ASE-001's lead clause breaks the
linter's subject matcher) + 3 `CoverageIncomplete` (the known cross-file
spec.md↔acceptance.md resolution limitation; this Tier S SPEC carries its ACs
inline in spec.md §3). Pre-existing from plan phase, unchanged by sync.

Sync-phase commit contents (`git show --stat`): wrapper `spec.md` (frontmatter
status only), wrapper `progress.md` (§E.4 populated), HWD `spec.md` (frontmatter
status only), `CHANGELOG.md` (one row), plus this report's `sync-evidence/`
artifacts.

## Baseline-attribution

- All three commands measured in this run, on this tree, at the sync-commit
  HEAD. Verbatim outputs at the `sync-evidence/` paths above.
- Run-phase A5 measurements (§E.1 of wrapper progress.md, HEAD `820db6cf9`)
  were re-measured, not carried: lint composition and mirror exit are identical
  across the two trees, as expected — sync touched frontmatter `status:` and
  prose in CHANGELOG/progress.md only, none of it lint- or mirror-relevant.

## Gaps

- **sync-auditor review not yet performed.** The lane dispatches sync-auditor
  after this commit; the auditor's score row belongs here:
  | Auditor | Score | Verdict path |
  |---|---|---|
  | sync-auditor | _(pending — dispatched separately by the lane)_ | _(to be recorded by the auditor)_ |
- `sync_commit_sha` in wrapper progress.md §E.4 carries the
  `pending-backfill-sync` placeholder and is backfilled in a follow-up commit
  (D3 exemption — a commit cannot cite its own SHA).
- No Go packages, tests, `go vet`, or build were in scope: this run's file
  deltas are `.moai/specs/**`, `CHANGELOG.md`, and `.moai/reports/**` only
  (scope statement in wrapper progress.md §E.1).

## Residual-risk

- **Carried debt from SPEC-HOOK-WIRING-DRIFT-001 plan.md §I.5** (recorded in
  wrapper progress.md "Carried Debt"): ① normalization over-absorption at
  parenthetical granularity; ② the internal-date class is NOT normalized — a
  future date-mandated template strip will report a false-FAIL; ③ the
  (ii-bare) absorption gap — a bare forbidden token inserted into the template
  copy passes the mirror check and is closed only by AC-HWD-016's
  template-side neutrality scan.
- **Plan-audit was PASS-WITH-DEBT 0.80** (Tier M boundary value) — the audit
  itself flagged D1–D5, all patched at `8eb5b9102`; the residual is that the
  margin was exactly at threshold, so any future re-audit of §I could tip
  either way without new facts.
- The re-close of SPEC-HOOK-WIRING-DRIFT-001 relies on its own frontmatter
  `status:` flip only; the drift detector will read the re-close commit as the
  wrapper's close subject — the commit subject names the wrapper ID, and the
  amendment's `re_close_path` row is the record tying the two closes together.
