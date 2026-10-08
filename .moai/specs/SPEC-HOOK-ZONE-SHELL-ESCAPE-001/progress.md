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
   (evidence-red-repro.md §Defect ②, exit 1):
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
host's bash 3.2.57 (byte-0x5C+`u2287` probe → `e2 8a 87` `\U00002287` → literal, no `\U`
support) and the pin input restated as the escape texts byte-0x5C+`u2287` /
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

Gate round 3: the prior deny-safe claim about `\U` on old bash DELETED as
false — the reviewer measured the bypass on this host (guard allow + the
protected file deleted through a literal-named entry: the guard judges the
decoded path while the `\U`-less shell acts on the literal-named symlink);
replaced with a support-boundary + residual statement (version-variance
family, alongside zsh; follow-up-card material; candidate closures noted
without decision). Host split pinned: `\u` renders / `\U` literal on
`/bin/bash` 3.2.57 (three pinned re-measurements). Pin input forms
re-verified already-compliant in f4024fbd3 (escape texts byte-0x5C+`u2287` /
`\U00002287`, both `e2 8a 87`).

Gate round 4 (verbatim gate-shaped texts applied; the lead's message
rendered the byte-0x5C+u2287 escape text as the literal glyph — restored to the escape text per the
texts' own "input is the escape TEXT" wording): §B note replaced verbatim
(the round-3 "accepted as a documented residual" acceptance claim and the
"sound over-approximation" candidate removed — the new-safety-claim class
the gate killed); AC-HZS-007 When/Then carries the mutant-fail clause
(maxDigits 4→2 must FAIL); plan M1(b) aligned; §F candidate bullet reduced
to the fail-closed option. NOTE: the lead's 0-hit `grep deny-safe`
self-check is unsatisfiable alongside the verbatim texts themselves
("NOT deny-safe" appears as a negation in the §B note and AC-HZS-007) —
applied verbatim and reporting the grep hits with locations instead of
rephrasing.

Gate round 5 + the glyph-poisoning discovery: (a) the doubled-backslash
form that the glyph-fix batch landed reaches no `\u` branch (gate-measured)
— all 13 double-backslash spots normalized to the NUMERIC input definition
(byte 0x5C followed by `u2287`; `\U00002287`), which no transport layer can
mangle; (b) the round-2/3 host-split conclusion ("\u renders on 3.2.57") is
SUPERSEDED — the morning probe was glyph-poisoned (the Write-tool script
carried the glyph: poisoned bytes od-verified, `$'` + e2 8a 87); the
decisive od-proven measurement (printf-assembled escape text, runtime
backslash) shows NEITHER `\u` NOR `\U` is supported on bash 3.2.57 — BOTH
render literally. The round-3 "Host split pinned" line above is superseded
accordingly. e2 8a 87 stands as the Go decoder output
(string(rune(0x2287))) and modern-bash (≥4.2) documented semantics; the
host-variance residual now covers BOTH escape families. Artifacts flipped
to both-literal (spec §B/§D/§F/table/REQ, AC-HZS-007, plan M1(b)/§D,
evidence corrections item 6 + row 5).

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
