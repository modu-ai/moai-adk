# SPEC-HOOK-ZONE-SHELL-ESCAPE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: authored (plan-audit pending — the card's plan-audit gate promotes this to audit-ready after PASS)
plan_complete_at: 2026-10-09
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M set)

### RED evidence (the reproduction the card was gated on)

- Canonical record: `evidence-red-repro.md` — TRACKED in this SPEC directory
  (gate round 2: evidence must be readable from a fresh checkout). It carries
  the 2026-10-08 lane-19 measurement, the 2026-10-09 byte corrections
  (`e2 8a 87`), and the full ground-truth re-measurement on this host's bash
  3.2.57 (all seven rows, `od`). Additional committed carriers: the
  instrument header comment in the baseline-first commit `9dbe40c0a`
  (test(t1585), branch `WT-zone-gate-defects`) and the verbatim RED strings
  quoted in the committed acceptance.md cells. The gitignored
  `.moai/reports/t1585/red-repro.md` (verified absent from `9dbe40c0a` via
  `git cat-file -e` exit 128) is the demoted card-scoped duplicate. Ordering
  attribution satisfied: the RED baseline precedes every repair commit
  (verification-claim-integrity §2.3).

### Reproduced defect claims (each with its evidence line)

1. **① P1 NUL-truncation protected-path bypass** — the guard judges
   `zone_dir\x00/sub` while bash truncates the ANSI-C part at the NUL and
   removes `zone_dir` itself. Evidence (evidence-red-repro.md §Defect ①, exit 1):
   `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected directory is GONE (rm output: "")`
2. **② P2 `\x` raw-byte vs code-point confusion** — a raw-byte spelling of a
   protected path decodes to re-encoded text no zone entry matches. Evidence
   (red-repro.md §Defect ②, exit 1):
   `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected marker is GONE (rm output: "")`
3. **③ P2 no-digit hex panic + mangling** — `\x`/`\u`/`\U` with no digit
   panics (slice bounds) at end-of-string and mangles text after a non-digit.
   Evidence (evidence-red-repro.md §Defect ③, exit 1): `"\x": PANICKED — slice bounds
   out of range on the no-digit escape` (×3 prefixes) + `"\xZ": decoded
   "xZZ", want the bash literal "\xZ"` (×2 shapes).

### Decision record

decided_by: lane-19 — RED-first gate satisfied: the three defects were
UNREPRODUCED at dispatch and the RED reproduction was measured 2026-10-08 in
the card worktree (bash ground truth first, then the instrument; exit 1 on
all three defect rows), evidence canonicalized as the tracked
`evidence-red-repro.md` with the instrument pinned baseline-first
as commit `9dbe40c0a`. Run phase may not repair beyond what the RED rows
measure.

### Repair record — review-gate round 1 (2026-10-09)

Verdict fail, 3 P2 findings, 2 on the SPEC artifacts (finding 3, the
instrument, repaired by the lane at `7be9f41b5` — controls split into
independent tests, pre-repair green MEASURED for both): (1) the `\u` pin
bytes `e2 a8 87` corrected to `e2 8a 87` — re-measured with `od` on this
host's bash 3.2.57 (`⊇` → `e2 8a 87`; `\U00002287` → literal, no `\U`
support) and the pin input restated as the escape texts `⊇` /
`\U00002287`; (2) the red-repro.md commit attribution corrected — the report
is gitignored machine-local, the committed carrier is the instrument header
at `9dbe40c0a`. decided_by=lane-19 re-delegation (msg 97ea7afa).

Gate round 2 escalated the custody fix: the canonical record now lives
TRACKED at `evidence-red-repro.md` in this SPEC directory (fresh-checkout
readable; carries the byte corrections + the full 2026-10-09 re-measurement
on bash 3.2.57 — all seven ground-truth rows, `od`). Plan-audit iteration-1
(FAIL 0.79, 6 blocking) folded into the same repair commit: D1 AC-HZS-006
§2.1 conditional-demotion sentence; D4 new AC-HZS-009 (part-level
`$'a\x00b'X` → `aX`) + AC-HZS-010 (octal-origin command), instrument rows
owed at M1; D5 new AC-HZS-011 (non-ASCII outside-zone allow controls,
literal + raw-byte) closing the §E.2 mutant-claim gap; D6 AC-HZS-008 cell 1
re-scoped to the recorded 3-test run + M1's own package baseline; O1 excerpt
wording; O2 selector-form alignment noted in the acceptance header.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
