# SPEC-HOOK-ZONE-SHELL-ESCAPE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (plan-audit iteration 3 PASS-WITH-DEBT 0.94 — Tier M threshold 0.80 passed, must_pass_failed 0, blocking 0; verdict `.moai/reports/t1585/plan-audit-iter3.md`, receipts rcpt-dfffac88dbe5f6b32de44dcb / rcpt-dd78adc9776bd35482495236 / rcpt-55dad295a91674791a673a7e, frozen-tip artifact hash 0c7cf355fb6262461598019bf27dcb8c81e5e3d2ab682f658f604a78034369c2)
plan_complete_at: 2026-10-09
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M set)

### Plan→run Kickoff decision record (autonomous form, 2026-10-09)

decided_by=lane-19 (orchestrator) evidence_refs=[.moai/reports/t1585/plan-audit-iter3.md, frozen tip `9b6ae0da5`, artifact hash 0c7cf355…, receipts ×3] ladder_path=auto-semantics §9.1 (independent audit PASS-class + per-tier score ≥ 0.80 + artifact-hash unchanged since verdict + no blockers). Debt ADMITTED, not pre-fixed: acceptance.md:308 ledger Expected row still names the demoted record — a one-word fix now would invalidate the fresh verdict's artifact hash and force an iter-4 re-audit; M1-era ledger work touches that section anyway → dispose_in=run, carried into the run delegation. Known mid-flight state, by design: the three RED rows fail at the frozen tip (the two-cell observation window, verification-completeness §2 + verification-claim-integrity §2.3 ordering attribution); M2's decoder repair flips them GREEN before any push/merge, so the branch merges green (gate round-6 P1 disposition: `.moai/state/watchdog/t1585.json`). Proceeding to run phase M1 → M2 → M3.

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

### M1 — RED baseline re-confirmation + instrument finalization (2026-10-09)

**Pre-flight re-run at the committed baseline — the three defect rows + two
controls (E8 pre-GREEN evidence).** Tree: `9b6ae0da5` (`WT-zone-gate-defects`;
the instrument rows below were uncommitted at measurement time — the M1 commit
carries these exact bytes).

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellAnsiCNulTruncationBypass|TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl|TestCheckProtectedZoneShellHexRawByteBypass|TestCheckProtectedZoneShellHexDirectSpellingControl|TestZoneUnescapeAnsiCNoDigitHexStaysLiteral' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
=== RUN   TestCheckProtectedZoneShellAnsiCNulTruncationBypass
    protected_zone_shell_repro_test.go:128: BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected directory is GONE (rm output: "")
    protected_zone_shell_repro_test.go:130: swept=0
--- FAIL: TestCheckProtectedZoneShellAnsiCNulTruncationBypass (0.02s)
=== RUN   TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl
    protected_zone_shell_repro_test.go:174: swept=1
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl (0.02s)
=== RUN   TestCheckProtectedZoneShellHexRawByteBypass
    protected_zone_shell_repro_test.go:216: BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected marker is GONE (rm output: "")
    protected_zone_shell_repro_test.go:218: swept=0
--- FAIL: TestCheckProtectedZoneShellHexRawByteBypass (0.02s)
=== RUN   TestCheckProtectedZoneShellHexDirectSpellingControl
2026/10/09 02:34:14 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=존/marker.md"
    protected_zone_shell_repro_test.go:258: swept=1
--- PASS: TestCheckProtectedZoneShellHexDirectSpellingControl (0.01s)
=== RUN   TestZoneUnescapeAnsiCNoDigitHexStaysLiteral
    protected_zone_shell_repro_test.go:281: "\\x": PANICKED — slice bounds out of range on the no-digit escape
    protected_zone_shell_repro_test.go:281: "\\u": PANICKED — slice bounds out of range on the no-digit escape
    protected_zone_shell_repro_test.go:281: "\\U": PANICKED — slice bounds out of range on the no-digit escape
    protected_zone_shell_repro_test.go:285: "\\xZ": decoded "xZZ", want the bash literal "\\xZ"
    protected_zone_shell_repro_test.go:285: "a\\xZb": decoded "axZZb", want the bash literal "a\\xZb"
    protected_zone_shell_repro_test.go:291: swept=5
--- FAIL: TestZoneUnescapeAnsiCNoDigitHexStaysLiteral (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.990s
FAIL
```

**M1 new-row pre-repair observation (AC-HZS-006 / 007 / 009 / 010 / 011).**
Tree: `9b6ae0da5` + the M1 instrument rows (uncommitted at measurement time;
the M1 commit carries these exact bytes). The `\u` pin inputs were byte-verified
after authoring: `grep -c 'u2287'` = 2, `od -c` on the literal lines shows the
doubled-backslash form (`5c 5c 75 32 32 38 37` and `5c 5c 55 30 30 30 30 32 32
38 37`) — no glyph bytes (`e2 8a 87`) in the input literals.

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape|TestZoneUnescapeAnsiCCodePointRenderingPinned|TestZoneWordTextAnsiCPartTruncatesAtNul|TestCheckProtectedZoneShellOctalNulTruncationDenied|TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
=== RUN   TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape
    protected_zone_shell_repro_test.go:341: guard walk PANICKED on the no-digit escape in "rm -r zone_dir$'\\x'/sub" — the walk must complete and return a decision
--- FAIL: TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape (0.00s)
=== RUN   TestZoneUnescapeAnsiCCodePointRenderingPinned
    protected_zone_shell_repro_test.go:378: swept=2
--- PASS: TestZoneUnescapeAnsiCCodePointRenderingPinned (0.00s)
=== RUN   TestZoneWordTextAnsiCPartTruncatesAtNul
    protected_zone_shell_repro_test.go:402: zoneWordText($'a\x00b'X) = "a\x00bX", want "aX" — the ANSI-C part ends at its first NUL byte and the later part still appends
    protected_zone_shell_repro_test.go:404: swept=1
--- FAIL: TestZoneWordTextAnsiCPartTruncatesAtNul (0.00s)
=== RUN   TestCheckProtectedZoneShellOctalNulTruncationDenied
    protected_zone_shell_repro_test.go:424: octal nul truncation: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:428: swept=1
--- FAIL: TestCheckProtectedZoneShellOctalNulTruncationDenied (0.01s)
=== RUN   TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed
    protected_zone_shell_repro_test.go:459: swept=2
--- PASS: TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.070s
FAIL
```

Row states at M1: guard no-crash RED (walk panics — AC-HZS-006 release-blocking
eligibility earned: command + verbatim stdout + exit 1 + tree above);
code-point pin green-now ×2 (AC-HZS-007 cell 1); part-level RED (decoded
`a\x00bX` — AC-HZS-009 RED earned); octal-origin RED (allow — AC-HZS-010 RED
earned); non-ASCII outside-zone control green ×2 (AC-HZS-011 cell 1).

**M1 package-wide pre-repair baseline (AC-HZS-008 cell 1).** The full-package
run does NOT complete inside the default 10m budget on this host: an
instrument-confirmation run cut at the 10m default (FAIL 601.064s) and a
re-run at `-timeout=12m -v` cut at 12m with the suite still mid-flight
(657 top-level PASS, 7 FAIL named below, then `panic: test timed out after
12m0s`, the only running test at the cut `TestSyncGateFailState_AC006c_RetryBoundAndNotice (0s)`
— progressing, not hung). Full hook package = 1375 test functions; the suite
exceeds 12 minutes wall on this machine. This is the t1542-class structural
condition (in-card-worktree suite scale); the package-wide verdict owner is
remote CI (`origin/develop`).

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout=12m -v ./internal/hook/`
- **Exit code**: `1` (timeout cut)
- **Tree**: `9b6ae0da5` + the M1 instrument rows
- **Observed fail inventory (verbatim names from the run)**:

```
--- FAIL: TestCheckProtectedZoneShellAnsiCNulTruncationBypass (0.02s)
--- FAIL: TestCheckProtectedZoneShellHexRawByteBypass (0.03s)
--- FAIL: TestZoneUnescapeAnsiCNoDigitHexStaysLiteral (0.00s)
--- FAIL: TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape (0.01s)
--- FAIL: TestZoneWordTextAnsiCPartTruncatesAtNul (0.00s)
--- FAIL: TestCheckProtectedZoneShellOctalNulTruncationDenied (0.13s)
--- FAIL: TestStaleRunNoticeFactoryLegacyLabel (1.47s)
```

The first six are exactly the expected M1 RED set (three defects + three
authored rows). The seventh is OUTSIDE this card's scope and interference- or
env-flavored: `TestStaleRunNoticeFactoryLegacyLabel` (stale-run notice text)
PASSES in isolation (0.63s, measured this tree) and its in-suite failure body
is `factory messaging degraded: context deadline exceeded` — the countable
M2/M3 delta is therefore "the six zone rows flip green; zero new failures;
the stale-run row is tracked as a known non-delta (isolation-green)". No
coverage figure is printable from a failing run; the pre-repair comparable
baseline was measured on the affected family scope instead:

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -cover -run 'TestProtectedZone|TestCheckProtectedZone|TestZoneUnescape|TestZoneWordText' -skip '^TestCheckProtectedZoneShellAnsiCNulTruncationBypass$|^TestCheckProtectedZoneShellHexRawByteBypass$|^TestZoneUnescapeAnsiCNoDigitHexStaysLiteral$|^TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape$|^TestZoneWordTextAnsiCPartTruncatesAtNul$|^TestCheckProtectedZoneShellOctalNulTruncationDenied$' ./internal/hook/`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	5.244s	coverage: 12.2% of statements`
- The M3 comparison re-runs the identical selector, pre- and post-repair
  figures on the same executed set.

**Pre-flight builds (M1)**: `go build ./...` exit 0; `GOOS=windows go build
./...` exit 0 (this tree, `9b6ae0da5`). Lint baseline: `golangci-lint run
internal/hook/... --timeout=2m` — `0 issues.` (this tree, M1 rows included).


## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
