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

### M2 — decoder repair GREEN flips (2026-10-09)

Fixes applied (plan §A.2 all three adopted directions, shape call per §A.2①):
① NUL part-terminator at the `zoneWordText` ANSI-C branch — the decoded
part's contribution is truncated at its first NUL byte (`strings.IndexByte`)
BEFORE assembly, so later parts still append (part-level shape, fails the
word-level mutant); ② render split in `zoneHexEscape` — new `rawByte` param:
`\x` → `string([]byte{byte(val)})` (one raw byte), `\u`/`\U` → `string(val)`
(code point as UTF-8), the shared digit-scanning loop unchanged; ③ the
no-digit arm returns `v[*i-1 : *i+1]` (bounded backslash + prefix letter at
the prefix — never indexes past the end; the caller's loop advances past the
letter only, so a follower survives once).

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellAnsiCNulTruncationBypass|TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl|TestCheckProtectedZoneShellHexRawByteBypass|TestCheckProtectedZoneShellHexDirectSpellingControl|TestZoneUnescapeAnsiCNoDigitHexStaysLiteral|TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape|TestZoneUnescapeAnsiCCodePointRenderingPinned|TestZoneWordTextAnsiCPartTruncatesAtNul|TestCheckProtectedZoneShellOctalNulTruncationDenied|TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed' -count=1 -v`
- **Exit code**: `0`
- **Observed (verbatim tail)**:

```
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationBypass (0.01s)
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl (0.00s)
--- PASS: TestCheckProtectedZoneShellHexRawByteBypass (0.01s)
--- PASS: TestCheckProtectedZoneShellHexDirectSpellingControl (0.01s)
--- PASS: TestZoneUnescapeAnsiCNoDigitHexStaysLiteral (0.00s)
--- PASS: TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape (0.03s)
--- PASS: TestZoneUnescapeAnsiCCodePointRenderingPinned (0.00s)
--- PASS: TestZoneWordTextAnsiCPartTruncatesAtNul (0.00s)
--- PASS: TestCheckProtectedZoneShellOctalNulTruncationDenied (0.05s)
--- PASS: TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	1.878s
```

- **Tree**: M1 commit `c34021856` + the M2 source edit (uncommitted at
  measurement time; the M2 commit carries these exact bytes).
- Flip semantics observed: the bypass deny judges `zone_dir` (the truncated
  word — deny reason `path=zone_dir`); the octal row denies identically; the
  hex raw-byte deny carries `path=존/marker.md` (decoded bytes match the
  entry); the guard no-crash row observes a DECISION (`deny`, the resolver's
  lexical-arm over-approximation on `zone_dir\x/sub` — the safe direction;
  the row's requirement is a decision, never a panic); both controls and the
  code-point pin stay green.
- Post-repair builds: `gofmt -l internal/hook/` empty; `go vet
  ./internal/hook/` clean; `go build ./...` exit 0; `GOOS=windows go build
  ./...` exit 0 (this tree, M2 edit included).

The acceptance.md Evidence Ledger GREEN-flips section is populated by
manager-spec (run-phase ownership boundary — reported to the orchestrator
with the exact wording above).

### M3 — family re-run + regression confirmation (2026-10-09)

**Full package, post-repair** (slot lease `hook-suite` held for the run):

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout=25m -v ./internal/hook/`
- **Exit code**: `0`
- **Observed (verbatim tail)**:

```
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	463.797s
PACKAGE_POST_EXIT=0
```

- Counts from the same run: 3638 `=== RUN` lines, 1364 top-level `--- PASS`,
  ZERO `--- FAIL` lines. The suite COMPLETES inside the budget on a warm
  build cache (the M1-era 10m/12m cuts were cold-compile + load, not a hang —
  the only running test at the 12m cut had 0s elapsed).
- The ten instrument tests in that run (verbatim):

```
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationBypass (0.00s)
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl (0.00s)
--- PASS: TestCheckProtectedZoneShellHexRawByteBypass (0.00s)
--- PASS: TestCheckProtectedZoneShellHexDirectSpellingControl (0.00s)
--- PASS: TestZoneUnescapeAnsiCNoDigitHexStaysLiteral (0.00s)
--- PASS: TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape (0.00s)
--- PASS: TestZoneUnescapeAnsiCCodePointRenderingPinned (0.00s)
--- PASS: TestZoneWordTextAnsiCPartTruncatesAtNul (0.00s)
--- PASS: TestCheckProtectedZoneShellOctalNulTruncationDenied (0.00s)
--- PASS: TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed (0.00s)
```

- `TestStaleRunNoticeFactoryLegacyLabel` PASS (0.48s) in this run — consistent
  with the M1 isolation-green + interference classification.
- **Countable delta vs the M1 baseline**: the six M1-final RED rows flipped
  green; zero new failures; the seventh (stale-run) row green in the quiet
  window. AC-HZS-007 re-verified green post-repair (the pin held — the \x
  split did not regress the code-point arm).

**Coverage pair (family selector identical pre/post — the comparable delta):**

- Post-repair WITH the M1 skip set (same executed set as the pre-change
  figure): `go test -count=1 -cover -run 'TestProtectedZone|TestCheckProtectedZone|TestZoneUnescape|TestZoneWordText' -skip '<the six-row skip set above>' ./internal/hook/` → `ok  	github.com/modu-ai/moai-adk/internal/hook	5.030s	coverage: 12.2% of statements` — **identical to the pre-change 12.2%** (no regression).
- Post-repair WITHOUT skip (all rows executing): same selector →
  `ok  	github.com/modu-ai/moai-adk/internal/hook	3.224s	coverage: 12.4% of statements`.
- The package-wide figure is owned by the completing full-package run above
  plus remote CI; a `-cover` figure is not printable from the M1-era failing
  baseline (noted as the measurement limit of the pre/post pair — the
  family-scoped pair is the regression evidence).

**Regression guards (M3):** `gofmt -l internal/hook/` empty; `go vet
./internal/hook/` clean; `go build ./...` exit 0; `GOOS=windows go build
./...` exit 0 (post-repair tree, verbatim in the M2 record's build line).

**E4 boundary grep (post-repair tree):** `grep -rn 'AskUserQuestion'
internal/hook | grep -v "_test.go" | grep -v "// "` → **1 match**
`internal/hook/pre_tool.go:856: if input.ToolName == "AskUserQuestion" {` —
PRE-EXISTING (the plan-audit frozen tip `9b6ae0da5` carries 4 raw occurrences
in the same file; this line is the hook's AskUserQuestion observation branch,
which the adjacent comment documents as never-denying — it observes the tool
NAME, it does not invoke the question channel). Reported as a deviation from
the dispatch's 0-match expectation; not introduced by this run.

**E5 lint (post-repair tree):** `golangci-lint run internal/hook/...
--timeout=2m` → `0 issues.` exit 0 — no new warnings or lints vs the M1
baseline (`0 issues.`).

**Run-phase acceptance.md findings reported to the orchestrator (manager-spec
owns that artifact):** (1) the M1-final baseline confirmation command in the
Evidence Ledger (`-run '^(TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC)$'`)
sweeps ZERO tests — measured: `testing: warning: no tests to run` / `ok ...
[no tests to run]`, exit 0 — the anchored full-name form cannot match the
long test names; the unanchored selector forms used in §E.2 above are the
measured-working set; (2) the GREEN-flips ledger entries may cite this §E.2
per the ledger's citation convention (as the M1-final entry already does).

### Gate round 10 — M2.1 origin-scoping refinement (2026-10-09)

The review gate fired a P1 on the first M2 cut: the part-level truncation
implemented there cut at ANY decoded NUL byte, while REQ-HZS-001 scopes the
terminator to "a NUL byte — from `\x00` or from an octal escape". A
code-point-origin NUL (`\u0000`) must NOT terminate: the escape's support is
version-variant (this host's bash 3.2.57 renders the escape text literally —
the card's own both-literal measurement), so a truncating decoder judges a
SHORTER word than the pre-4.2 shell acts on. Reviewer-measured regression: a
literally-named `docs\u0000` entry plus `rm $'docs\u0000/../zone_dir/
marker.md'` — bash 3.2 resolves through the literal-named entry and deletes
the protected marker; the broad truncation judged `docs` and ALLOWED.

**New instrument row (RED under the first M2 cut `f4a0227f3`)** — command +
verbatim output + exit code + tree, the four elements: tree `f4a0227f3` +
the row edit (committed in the M2.1 row commit preceding the refinement
commit).

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellCodePointNulDoesNotTruncate' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
=== RUN   TestCheckProtectedZoneShellCodePointNulDoesNotTruncate
    protected_zone_shell_repro_test.go:499: code-point nul origin scoping: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:503: swept=1
--- FAIL: TestCheckProtectedZoneShellCodePointNulDoesNotTruncate (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.848s
FAIL
```

The row's inputs were transport-verified after authoring: `od -c` shows the
doubled-backslash form (`5c 5c 75 30 30 30 30`) on both literals and a
whole-file NUL-byte scan returns zero.

**Origin-scoping note (the one-line record the refinement owes):** the NUL
terminator truncates ONLY at `\x00` and octal-escape origins — the two
origins every bash renders as a NUL byte, named verbatim by REQ-HZS-001. A
`\u`/`\U` code point whose value is 0 stays in the decoded text (pre-fix
judgment shape): its rendering is version-variant, the divergence is the
documented `\u`/`\U` host-variance residual (spec §B), and the NUL-bearing
text under-matches toward deny on the lexical arm (Clean-collapse).

**M2.1 refinement — origin-scoped truncation (GREEN record).** Shape: the
termination moved INTO `zoneUnescapeAnsiC` where the origin is known — the
`\x` case returns the accumulated text when the escape rendered byte 0, the
octal case likewise on `byte(0)`; the `\u`/`\U` cases never terminate; the
`zoneWordText` post-truncation of the first M2 cut is REVERTED (it cut at
any-origin NULs — the gate-10 P1). Later word parts still append.

- **Command** (all 11 instrument tests on the refined tree):
  `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run '<the ten names above>|TestCheckProtectedZoneShellCodePointNulDoesNotTruncate' -count=1 -v`
- **Exit code**: `0`
- **Observed (verbatim tail)**:

```
=== RUN   TestCheckProtectedZoneShellCodePointNulDoesNotTruncate
2026/10/09 03:36:33 WARN protected zone shell violation agent_id="" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/marker.md"
    protected_zone_shell_repro_test.go:503: swept=1
--- PASS: TestCheckProtectedZoneShellCodePointNulDoesNotTruncate (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.961s
```

The three defect rows + 006/009/010 stay green; controls 002/004/011 and the
pin 007 stay green; the new row's deny reason carries
`path=zone_dir/marker.md` — the Clean-collapse judgment shape the reviewer
measured at the comparison commit.

- **Full package regression (M2.1)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	359.150s` /
  `PACKAGE_POST21_EXIT=0`; 3640 RUN lines, ZERO `--- FAIL` lines; the new row
  passes inside the package run (`--- PASS:
  TestCheckProtectedZoneShellCodePointNulDoesNotTruncate (0.00s)`). Slot
  lease `hook-suite` held for the run and released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage unchanged post-M2.1 (`12.4%`, all-rows selector).

### Gate round 13 — M2.2 dual-candidate judgment for mixed-origin NUL words (2026-10-09)

The review gate fired a P1 on the M2.1 refinement itself (reviewer confidence
1.00, real-bash repro): MIXED-origin NUL words break the origin-scoped
termination. A part carrying an EARLIER code-point-origin NUL (`\u0000` /
`\U00000000`) and a LATER hex/octal-origin NUL terminates at the latter — so
the part returns only the prefix and the judged candidate stops short of the
zone, while the pre-4.2 shell acts on the FULL literal path through a
literally-named symlink into the zone (`printf changed > $'link\u0000\x00/
../zone_dir/marker.md'` overwrites the marker; the audit baseline DENIED the
same input — the whole NUL-bearing text Clean-collapsed into the zone). The
remedy judges BOTH worlds for any word carrying `\u`/`\U` escapes: the
modern-decoded candidate AND the raw source-text candidate (the t1566 raw
arm resolves literal-named entries through symlinks; the raw text
Clean-collapses into the zone on the lexical arm) — consonant with the
guard's possible-worlds design, and narrower than failing closed every
`\u`/`\U`-bearing word.

**Four mixed-origin regression rows — RED under the M2.1 tip (`4bdc4469a` +
the rows, uncommitted at measurement).** Command, verbatim output, exit
code, tree — the four elements:

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellMixedOriginNulDenied' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
=== RUN   TestCheckProtectedZoneShellMixedOriginNulDeniedRedirect
    protected_zone_shell_repro_test.go:565: mixed origin nul redirect: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:565: swept=1
--- FAIL: TestCheckProtectedZoneShellMixedOriginNulDeniedRedirect (0.00s)
=== RUN   TestCheckProtectedZoneShellMixedOriginNulDeniedOctalTerm
    protected_zone_shell_repro_test.go:572: mixed origin nul octal term: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:572: swept=1
--- FAIL: TestCheckProtectedZoneShellMixedOriginNulDeniedOctalTerm (0.00s)
=== RUN   TestCheckProtectedZoneShellMixedOriginNulDeniedUpperHex
    protected_zone_shell_repro_test.go:579: mixed origin nul upper hex: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:579: swept=1
--- FAIL: TestCheckProtectedZoneShellMixedOriginNulDeniedUpperHex (0.00s)
=== RUN   TestCheckProtectedZoneShellMixedOriginNulDeniedUpperOctal
    protected_zone_shell_repro_test.go:586: mixed origin nul upper octal: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:586: swept=1
--- FAIL: TestCheckProtectedZoneShellMixedOriginNulDeniedUpperOctal (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.690s
FAIL
```

Row matrix (both judgment channels covered): redirect+u0000+x00 (the
reviewer's exact command), rm-arg+u0000+octal, redirect+U00000000+x00,
rm-arg+U00000000+octal; each fixture carries the literally-named symlinks
`link\u0000` / `link\U00000000` → zone_dir/marker.md. Inputs
transport-verified: whole-file NUL-byte scan zero; the escape literals carry
the doubled backslash.

**M2.2 remedy — dual-candidate judgment (GREEN record).** Shape: three new
word-text helpers — `zoneWordDual` (does the word carry `\u`/`\U` in a
Dollar-single-quoted part — the one version-variant escape family),
`zoneWordRawText` (the pre-4.2 reading: ANSI-C parts keep their source text,
Lit/DblQuoted decode as usual — identical across generations), and
`zoneWordCandidates` (the decoded text plus the raw reading when dual). The
three PATH-CANDIDATE funnels — `zonePathCandidates` (mutation-verb
arguments), `zoneRedirectTargets` (write redirections), and the git
`-C`/`--work-tree` file-argument loop — now append BOTH worlds, so a
deny on EITHER candidate denies (the guard's sound possible-worlds
over-approximation; `zoneFirstArgWord` stays single-world: it extracts the
VERB, not a path). Fail-closed-on-`\u` was the narrower alternative and is
NOT taken — it would over-block legal modern-bash paths.

- **Command** (all 15 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run '<the 11 names above>|TestCheckProtectedZoneShellMixedOriginNulDenied' -count=1 -v`
- **Exit code**: `0`
- **Observed (verbatim)**:

```
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedRedirect (0.01s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedOctalTerm (0.01s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedUpperHex (0.00s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedUpperOctal (0.00s)
--- PASS: TestCheckProtectedZoneShellCodePointNulDoesNotTruncate (0.01s)
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl (0.01s)
--- PASS: TestCheckProtectedZoneShellHexDirectSpellingControl (0.00s)
--- PASS: TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.746s
```

All 15 green: the four mixed rows flipped DENY, the eleven earlier rows
(including the M2.1 scoping row and the non-ASCII raw-byte allow controls)
stay green — the dual world adds no deny pressure on words without
`\u`/`\U`.

- **Full package regression (M2.2)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	324.657s` /
  `PACKAGE_POST22_EXIT=0`; 3644 RUN lines, ZERO `--- FAIL` lines; the four
  mixed rows pass inside the package run. Slot lease `hook-suite` held for
  the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `12.6%` (all-rows selector — the new rows exercise
  the dual-world paths; pre-change family pair baseline was 12.2%).

### Gate round 14 — M2.3 old-bash rendering correction + windows skip (2026-10-09)

Two findings on the dual-candidate v1 (reviewer, real bash): **P1** — the
old-bash (literal) candidate carried the WHOLE ANSI-C raw text, but bash 3.2
DECODES `\xHH` and octal escapes and truncates the argument at their NULs;
ONLY `\u`/`\U` stay literal (decisively measured this card: the `⊇`
escape text passes as `5c 75 32 32 38 37`). For `printf changed >
$'link\u0000\x00/../\x7aone_dir/marker.md'` the true 3.2 path truncates at
the `\x00` NUL — `link\u0000` — and the literally-named symlink resolves
INTO the zone, while the whole-raw candidate judged the hex-escaped zone
component (0x5C x 7a = "z") as an undecoded name and ALLOWED (base deny,
marker overwritten). **P2** — the four mixed rows died in `t.Fatal` on the
windows release matrix: the literal `link\u0000` name carries a backslash, a
separator on windows, so `os.Symlink` fails.

**New gate-14 regression row — RED under the dual-candidate v1 tip
(`b58eaed73` + the rows, uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
=== RUN   TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent
    protected_zone_shell_repro_test.go:615: mixed origin nul hex component: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:619: swept=1
--- FAIL: TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.875s
FAIL
```

The same commit adds the windows `t.Skip` to the shared mixed fixture
(`hzsMixedNulFixture` — one skip point covers all five literal-named rows,
P2). Inputs transport-verified: whole-file NUL-byte scan zero, doubled
backslash on the new escape literals.

**M2.3 remedy — old-bash rendering correction (GREEN record).** Shape: the
two worlds became explicit decoders of one shared loop
(`zoneUnescapeAnsiCWorld`): the MODERN world (`zoneUnescapeAnsiC`) renders
`\u`/`\U` as UTF-8 code points and ends the part at the first NUL of ANY
origin (modern bash truncates however spelled — superseding the M2.1
origin-scoped truncation; the pre-4.2 divergence it deferred is now carried
by the second candidate instead of the judged text); the PRE-4.2 world
(`zoneUnescapeAnsiCPre42`) decodes `\xHH`/octal/simple escapes exactly the
same — raw bytes, part ending at their NUL — while `\u`/`\U` are UNKNOWN
escapes (backslash + letter stay, digits are ordinary characters; measured
`5c 75 32 32 38 37`). `zoneWordRawText` (whole-raw candidate) is REPLACED by
`zoneWordTextPre42`; the candidate set per `\u`/`\U`-bearing word is
{modern, pre-4.2}, deny when either lands in the zone. The sync-phase
`@MX:DEBT` on `zoneUnescapeAnsiC` (single-modern-decode divergence,
upgrade=fail-closed) is REMOVED by this change: the simplification it
described is replaced by the two-world pair, and the fail-closed upgrade
option was rejected at gate 13 (over-blocks legal modern-bash paths).

- **Command** (all 16 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run '<the 15 names above>|TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent' -count=1 -v`
- **Exit code**: `0`
- **Observed (verbatim)**:

```
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent (0.00s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedRedirect (0.00s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedOctalTerm (0.00s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedUpperHex (0.00s)
--- PASS: TestCheckProtectedZoneShellMixedOriginNulDeniedUpperOctal (0.00s)
--- PASS: TestCheckProtectedZoneShellCodePointNulDoesNotTruncate (0.00s)
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl (0.00s)
--- PASS: TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.934s
```

All 16 green: the gate-14 row flipped DENY (the old-bash candidate resolves
`link\u0000` through the symlink arm), the fifteen earlier rows hold.

- **Full package regression (M2.3)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	377.691s` /
  `PACKAGE_POST23_EXIT=0`; 3645 RUN lines, ZERO `--- FAIL` lines; the new
  row passes inside the package run. Slot lease `hook-suite` held for the
  run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `12.8%` (all-rows selector).

### Gate round 15 — M2.4 consumer-set sweep: four P1s (2026-10-09)

Four P1s on the dual-world integration (reviewer overlay test
TestReviewDualWorldRegression, 5 shapes, base deny → current allow, real
bash): the dual worlds were wired into the PATH-CANDIDATE funnels only, and
four other consumers still read a single world — (1) an EMPTY modern
candidate discarded the whole word (`$'\u0000/../zone_dir/marker.md'`:
modern truncation yields "" and the pre-4.2 candidate was lost with it);
(2) the cd/git directory anchors read the modern world only (`cd
$'docs\u0000/../zone_dir'; rm marker.md` tracks docs while 3.2 lands inside
zone_dir); (3) the executable name truncated (`$'docs\u0000/../rm'
zone_dir/marker.md` extracted "docs", never recognizing the mutation verb);
(4) the long-option value took worlds[0] (`cp source.md
$'--target-directory=docs\u0000/../zone_dir'` extracted "docs"). The
completing principle: every consumption site of word text consumes the
candidate SET, or carries an explicitly justified single-world exception.

**Five regression rows — RED under the M2.3 tip (`d8eefcc79` + the rows,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellDualWorld' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
    protected_zone_shell_repro_test.go:680: dual world empty modern kept: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:680: swept=1
--- FAIL: TestCheckProtectedZoneShellDualWorldEmptyModernKept (0.00s)
    protected_zone_shell_repro_test.go:688: dual world cd readings: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:688: swept=1
--- FAIL: TestCheckProtectedZoneShellDualWorldCdReadings (0.01s)
    protected_zone_shell_repro_test.go:697: dual world verb recognition: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:697: swept=1
--- FAIL: TestCheckProtectedZoneShellDualWorldVerbRecognition (0.00s)
    protected_zone_shell_repro_test.go:705: dual world long option value: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:705: swept=1
--- FAIL: TestCheckProtectedZoneShellDualWorldLongOptionValue (0.00s)
    protected_zone_shell_repro_test.go:713: dual world git anchor: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:713: swept=1
--- FAIL: TestCheckProtectedZoneShellDualWorldGitAnchor (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.857s
FAIL
```

Fixture: literally-named `docs\u0000` directory + marker + source.md;
windows-skipped (backslash separator). Inputs transport-verified
(whole-file NUL scan zero, doubled backslash).

**M2.4 remedy — the consumer-set sweep (GREEN record).** Every consumption
site of word text now consumes the candidate set; the four named sites plus
one found during the fix re-run:

1. `zonePathCandidates` — empty readings filter PER WORLD (the modern "" at
   a leading code-point NUL no longer discards the word) AND the long
   option's attached value is extracted from EVERY world's spelling (sites
   1+4).
2. cd — the single argument's non-empty readings each become a POSSIBLE
   destination (`zoneNextCwd` per reading); the possible-directory set
   unions them; multi/zero-arg and dynamic shapes keep the original
   behavior (site 2-cd).
3. Executable name — new `zoneMutationVerbName`: if ANY world's base name of
   Args[0] is a mutation verb, the command is judged as that verb (site 3).
4. git anchors — `-C`/`--work-tree` values (and the `--work-tree=` attached
   prefix form) build READING SETS (`dirOpts`/`wtOpts`); the anchoring loop
   unions every combination (site 2-git).
5. `zoneRedirectTargets` — found during the fix re-run (the empty-modern
   row initially still failed): the same per-world empty filter (the word
   drops only when EVERY world is empty).

**Consumer-site sweep inventory** (grep of zoneWordText /
zoneFirstArgWord / zoneWordCandidates callers): executable name (fixed,
site 3); cd dirs (fixed, site 2); git `-C`/`--work-tree`/attached-prefix
values (fixed, site 2); git fileArgs (already dual, gate 13);
zoneRedirectTargets targets (fixed, item 5); zonePathCandidates words +
attached values (fixed, sites 1+4); zoneWordCandidates internals (dual by
construction). **Justified single-world exceptions:** (a) the git
SUBCOMMAND word — git dispatches subcommands internally on the exact
decoded string; the shell never path-resolves a subcommand name, so
cross-generation base-name matching would deny commands no generation
executes as a mutation (the modern decode covers modern bash exactly;
pre-4.2 renders \u literally and cannot form the verb from \u escapes);
(b) the sed `--in-place` option scan — the same exact-string argument;
(c) the `>&`-digits check (`isZoneDigits`) — a digits-only word contains
no backslash, so both worlds read identically (vacuously dual). **Bounded
residual documented:** the sed/git BRANCH DISPATCH (`switch name`) and the
function-declaration lookup key on the modern name — a word whose modern
reading is exactly `sed`/`git`/`cd`/a declared function while a pre-4.2
reading names a mutating executable is doubly-crafted and bounded (the
direct shapes are covered by zoneMutationVerbName; dispatch restructure
was not directed).

- **Command** (all 21 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.002s`
  (21/21 PASS — the five gate-15 rows flipped DENY, the sixteen earlier
  rows hold).
- **Full package regression (M2.4)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	300.005s` /
  `PACKAGE_POST24_EXIT=0`; 3650 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.2%` (all-rows selector).

### Gate round 17 — M2.5 per-world funnel semantics: four findings (2026-10-09)

The systemic principle extends one layer: each candidate funnel's EXISTING
semantics — emptiness filters, verb/specialized recognition, anchor
accumulation, overwrite-wins — must apply PER WORLD, not just to worlds[0].
Reviewer overlay-verified at `c125ff334`: (1) P1 the git file-arg funnel's
empty-modern filter still dropped the word's old-bash candidate
(`git rm -f $'\u0000/../zone_dir/marker.md'`); (2) P1 the specialized
analyses did not connect to the world names (`$'docs\u0000/../git' rm -f
zone_dir/marker.md` and the sed variant: the modern name truncates to
"docs", no verb, no git/sed analysis); (3) P2 each `-C` built the cartesian
product of the two bash readings — 18× `-C` produced 262,144 candidates for
2 unique paths; (4) P2 appended `--work-tree` candidates kept judging
already-overwritten anchors — git's LAST --work-tree wins, so judging the
earlier anchor is a FALSE DENY (an over-block, the inverse direction).

**Five regression rows — RED under the M2.4 tip (`c125ff334` + the rows,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellGitFunnelEmptyModernKept|TestCheckProtectedZoneShellDualWorldGitNameSpecialized|TestCheckProtectedZoneShellDualWorldSedNameSpecialized|TestCheckProtectedZoneShellGitAnchorAccumulationBounded|TestCheckProtectedZoneShellGitWorkTreeOverwriteWins' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
    protected_zone_shell_repro_test.go:727: git funnel empty modern kept: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellGitFunnelEmptyModernKept (0.00s)
    protected_zone_shell_repro_test.go:736: dual world git name specialized: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellDualWorldGitNameSpecialized (0.00s)
    protected_zone_shell_repro_test.go:743: dual world sed name specialized: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellDualWorldSedNameSpecialized (0.00s)
    protected_zone_shell_repro_test.go:756: git anchor accumulation bounded: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellGitAnchorAccumulationBounded (13.53s)
    protected_zone_shell_repro_test.go:770: git work-tree overwrite wins: decision="git --git-dir=docs/.git --work-tree=zone_dir --work-tree=docs rm -f marker.md" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/marker.md", want allowed — the last --work-tree replaces the earlier anchor
--- FAIL: TestCheckProtectedZoneShellGitWorkTreeOverwriteWins (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	14.174s
```

Rows 1-3 and 5 are decision-inverting REDs. Row 4 (the 18× -C shape) is
REGRESSION-GUARD with the §2.1 demotion stated: the boundedness defect's
mechanical in-suite reproduction is the row's own 13.53s crawl (the
reviewer's 262,144-for-2 measurement cited alongside); a decision inversion
is not constructible from -C crossings (they only deepen the anchor path),
so the row pins the deny through the per-world accumulation rather than a
pre-fix decision flip.

**M2.5 remedy — per-world funnel semantics (GREEN record).** Shape: (1) the
git file-arg funnel filters emptiness/options PER CANDIDATE (an empty modern
reading no longer discards the old-bash candidate); (2) the executable
word's possible names drive the SPECIALIZED analyses — `sedWorld`/`gitWorld`
computed from every world's base name fire the extracted `zoneSedInPlace` /
`zoneGitArgs` analyses (the old `switch name` is gone); (3) `-C`
accumulation is PER WORLD via `zoneWordWorldReadings` (world-indexed
[2]string readings): the modern reading extends the modern chain, the
pre-4.2 reading the pre-4.2 chain, never crossed, deduped — and the
CANDIDATE SET is capped (`zoneCandidateCap` = 4096) at judgment: beyond the
cap the walk is denied fail-closed (`w.unbounded`), the bounded-walk
philosophy; (4) `--work-tree` (both forms) is OVERWRITE-WINS per world — the
option's readings REPLACE the anchor slots, so an already-overwritten anchor
is never judged (the false-deny inverse is pinned by the over-block row).

Row-4 redesign disclosure (in-flight, before the M2.5 commit): the first
draft of the accumulation row pinned its deny to `/tmp`-rooted anchors,
which no world's semantics can walk into the zone — the row was rewritten
BEFORE the fix commit to the self-consistent absolute-replacement shape
above (modern reading = the project root; the deny lands through it). The
first draft's pre-fix measurement (13.53s allow) stands as the boundedness
evidence for the /tmp shape; the committed row carries the corrected shape.

- **Command** (all 26 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	0.916s`
  (26/26 PASS — the five gate-17 rows flipped: git funnel, git-name, sed-name,
  accumulation-bounded now DENY; the over-block row now ALLOWs — and the
  accumulation row dropped from 13.53s to 0.00s, the boundedness fix
  observable in-suite).
- **Full package regression (M2.5)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	283.276s` /
  `PACKAGE_POST25_EXIT=0`; 3655 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.5%` (all-rows selector).

### Gate round 19 — M2.6 dispatch order, per-generation joins, dedup-before-cap (2026-10-09)

Three findings on the in-flight M2.5 (the reviewer saw the test file change
mid-review): (1) P1 the cd BRANCH's early return skipped the dual-world verb
classification — `$'cd\u0000/../rm' zone_dir/marker.md` truncates to the
name "cd", the walker takes the cd branch and returns, and the pre-4.2
world's rm (through a literally-named `cd\u0000` entry) is never tested; a
word truncating to a DECLARED FUNCTION name is the same class; (2) P2 git
fileArgs pooled BOTH generations' readings and joined them against each
generation's directory — the CROSS-generation join (modern dir × pre-4.2
file text) is a path no generation executes and landing it is a FALSE DENY;
(3) P2 dedup ran AFTER the cap: a unicode-free command contributes the same
path from both worlds (2,050 files → 4,100 candidates → past the 4,096 cap
→ a loop-unbounded FALSE DENY on a plain command).

**Three regression rows — RED under the M2.5 tip (`474a99e92`; measured
against the pre-fix source restored from HEAD, with the rows uncommitted at
measurement).** The rows' fixture narrows the manifest to the single
protected file `zone_dir/marker.md` (gate round 20 P2: the wholesale
`zone_dir/` manifest let the modern reading trip on `zone_dir/other.txt` for
the wrong reason).

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellDualWorldVerbBeforeDispatch|TestCheckProtectedZoneShellGitFileArgsOwnGeneration|TestCheckProtectedZoneShellCandidateCapAfterDedup' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision lines; row 3's f-file enumeration
  abbreviated, carried in full by the run)**:

```
    protected_zone_shell_repro_test.go:831: dual world verb before dispatch: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellDualWorldVerbBeforeDispatch (0.01s)
    protected_zone_shell_repro_test.go:849: git file args own generation: decision="git -C $'zone\\u005fdir' rm -f $'other.txt\\u0000/../marker.md'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/other.txt/marker.md", want allowed — only the cross-generation join lands on the protected marker
--- FAIL: TestCheckProtectedZoneShellGitFileArgsOwnGeneration (0.01s)
    protected_zone_shell_repro_test.go:869: candidate cap after dedup: decision="git rm -f f0000.txt … f2049.txt" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=loop-unbounded route=human next=return-blocker-report path=loop", want allowed — the deduped candidate set (2,050) fits the cap
--- FAIL: TestCheckProtectedZoneShellCandidateCapAfterDedup (3.19s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	3.891s
```

Row 1 is a deny-miss; rows 2-3 are OVER-BLOCK inverse rows (false deny
observed verbatim — row 2's deny path is the crossed join landing on the
protected marker). Design iteration disclosed: the cross row's first two
drafts were unadoptable — the wholesale-manifest draft tripped on
`zone_dir/other.txt` (a true generation-0 path) and the `../`-climbing draft
was Clean-neutralized (the climb popped the joined dir); the committed shape
pairs the NARROWED manifest with the simple pop-to-marker word. Measured
against the pre-fix source by temporarily restoring
`HEAD:internal/hook/protected_zone_shell.go` (the M2.6 fix sat in the
working tree; restored after the capture). Inputs transport-verified:
whole-file NUL-byte scan zero, doubled backslash.

### Gate round 21 — M2.7 full dispatch pre-classification + function shadowing (2026-10-09)

Two findings on the M2.6 tip (`fd3738f42`, overlay-verified, base-PASS →
current-FAIL): (1) P1 the git/sed SPECIALIZED dispatch carried the same
early-return hole the mutation classification had just fixed —
`$'cd\u0000/../git' rm -f zone_dir/marker.md` truncates to the name "cd",
the cd branch dispatches and returns, and the pre-4.2 world's git analysis
never runs; (2) P2 a declared read-only function SHADOWS the external verb
— `rm() { printf 'read-only\n'; }; rm zone_dir/marker.md` executes only the
function in every world, but the mutation targets were registered before
the function registry was consulted (an M2.6 regression, over-block
inverse). The completing principle: ALL name-driven dispatches — mutation
verbs, git, sed, cd, AND declared functions — classify across both worlds
BEFORE any single-world branch runs, and a name SHADOWED by a declared
function executes the function in that world, dropping out of every other
dispatch.

**Two regression rows — RED under the M2.6 tip (`fd3738f42` + the rows,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellDualWorldGitNameBeforeCdDispatch|TestCheckProtectedZoneShellDeclaredFunctionShadowsVerb' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim)**:

```
    protected_zone_shell_repro_test.go:887: dual world git name before cd dispatch: decision="allow" reason="", want deny
    protected_zone_shell_repro_test.go:888: swept=1
--- FAIL: TestCheckProtectedZoneShellDualWorldGitNameBeforeCdDispatch (0.00s)
    protected_zone_shell_repro_test.go:902: declared function shadows verb: decision="rm() { printf 'read-only\\n'; }; rm zone_dir/marker.md" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/marker.md", want allowed — the function shadows the external rm and is read-only
    protected_zone_shell_repro_test.go:904: swept=1
--- FAIL: TestCheckProtectedZoneShellDeclaredFunctionShadowsVerb (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.618s
```

Row 1 is a deny-miss; row 2 is an OVER-BLOCK inverse row (false deny
observed verbatim on the protected marker). Fixture: the narrowed
marker-file manifest; inputs transport-verified (whole-file NUL-byte scan
zero, doubled backslash).

### Gate round 22 — M2.8 per-generation possible-directory set (2026-10-09)

Two P2s: (1) the GENERAL-command per-generation path relation — the cd's
directory reading and a later command's file-argument reading must keep
their generation relation (`cd $'zone_dir'; rm $'link_x'`: the
pooled join judged `zone_dir/link\u0005fx`, a path NO generation touches —
the reviewer planted a symlink only there; both a false deny and the
inverse bypass hang on it); (2) windows skip on the docs+u0000 row (the
literally-named entry parses as TWO directory levels on windows, so the
pre-4.2 target misses the protected file and wantZoneDeny fails there for
the wrong reason).

**Regression row 1 — RED under the M2.7 tip (`d28cab054` + the row,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellCdReadingGenerationRelation' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision line)**:

```
    protected_zone_shell_repro_test.go:927: cd reading generation relation: decision="cd $'zone\\u005fdir'; rm $'link\\u005fx'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/link/u005fx", want allowed — only the cross-generation cwd×file join reaches the planted symlink, and no generation executes it
--- FAIL: TestCheckProtectedZoneShellCdReadingGenerationRelation (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.902s
```

Row 2 (the windows skip) is instrument hardening with no darwin-observable
RED — the failure mode is windows-only (the two-level parse), per the
reviewer's analysis; the skip's correctness rides the windows release
matrix. Fixture: narrowed marker manifest + the planted cross-only symlink
`zone_dir/link\u0005fx` → the marker; POSIX-skipped. Inputs
transport-verified: whole-file NUL-byte scan zero, doubled backslash.

**M2.8 remedy — the per-generation possible-directory set (GREEN record).**
Shape: `w.cwds` carries its generation tag — `type zoneCwd { dir string;
gen int }` with gen -1 = generation-neutral (a generation-identical reading
put the walk there), 0/1 = a dir only that generation's reading reached.
The relation composes through the whole walk: a generation-tagged directory
moves only under its own generation's cd reading (a neutral directory
splits into per-generation entries), `zoneRelativeToSet` joins a candidate
reading of generation i only with gen -1/gen i directories (the
cross-generation cwd×file join is gone), and the git base loop binds
`fileArgs[world]` to bases of its own generation (neutral bases join both).
All the control-flow unions (fixed point, subshell, case, if/else) carry
entries with their tags mechanically. The funnel outputs are per-world
(`zonePathCandidates`/`zoneRedirectTargets` return [2][]string;
`zoneCands` takes [2][]string); the judgment's dedup-before-cap collapses
the both-worlds duplicates.

- **Command** (all 32 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.181s`
  (32/32 PASS — the cd-relation row flipped to ALLOW; the 31 earlier rows
  hold, including the t1570 matrix and every earlier gate row).
- **Full package regression (M2.8)**: on the FINAL tree (a 2-line
  dead-store fix landed after the first run; both runs green):
  `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED
  && go test -count=1 -timeout=25m -v ./internal/hook/` — exit 0, verbatim
  tail `PASS` / `ok github.com/modu-ai/moai-adk/internal/hook	264.828s` /
  `PACKAGE_POST28B_EXIT=0`; 3659 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the runs, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.5%` (all-rows selector).

### Gate rounds 23/25 — M2.9 full per-world state isolation (2026-10-09)

The isolation completion criterion: the two bash worlds are FULLY ISOLATED
interpretations — each world carries its OWN function registry, its own
name→verb binding, and its own argument readings; mutable state NEVER flows
between world iterations; generations NEVER cross within a judgment. The
ONLY place the worlds meet is the deny decision (union). Four findings, all
leaks of that isolation: (1) P1 candidate-state separation — the g
candidates' inner declarations overwrote each other on shared state
(`function g { f(){ :; }; }; g(){ f(){ rm zone_dir/marker.md; }; }; $'g';
f` — measured GREEN-now on M2.8: the body-snapshot machinery already held
for this shape, so the row pins it as a regression pin rather than a flip);
(2) P2 the verb-name's world joined the OTHER world's argument readings
(`$'rm\u0000/../printf' $'docs\u0000/../zone_dir/marker.md'` — modern rm ×
the pre-4.2 protected path = false deny); (3) P1 the function registry was
SHARED across versions (`f() { rm() { :; }; }; $'f\u0000/not_f'; rm
zone_dir/marker.md` — the modern world's no-op rm override leaked into the
pre-4.2 world, ALLOWING the real deletion: measured RED, the bypass
in-suite); (4) P2 the same name↔argument binding, noop shape.

**Four regression rows — measured under the M2.8 tip (`57c3d9428` + the
rows, uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellCandidateStateSeparated|TestCheckProtectedZoneShellVerbBindsOwnArgs|TestCheckProtectedZoneShellRegistryPerGeneration|TestCheckProtectedZoneShellVerbBindsOwnArgsNoop' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision lines)**:

```
--- PASS: TestCheckProtectedZoneShellCandidateStateSeparated (0.00s)
    protected_zone_shell_repro_test.go:970: verb binds own args: decision="$'rm\\u0000/../printf' $'docs\\u0000/../zone_dir/marker.md'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=docs/zone_dir/marker.md", want allowed — modern rm joins only the modern argument, and no world deletes
--- FAIL: TestCheckProtectedZoneShellVerbBindsOwnArgs (0.00s)
    protected_zone_shell_repro_test.go:986: registry per generation: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellRegistryPerGeneration (0.00s)
    protected_zone_shell_repro_test.go:1000: verb binds own args noop: decision="$'rm\\u0000/../noop' $'docs\\u0000/../zone_dir/marker.md'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=docs/zone_dir/marker.md", want allowed — modern rm joins only the modern argument, and no world deletes
--- FAIL: TestCheckProtectedZoneShellVerbBindsOwnArgsNoop (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.684s
```

Row 1 measured GREEN-now (the body-snapshot + mergeZoneFuncs machinery
already separated candidate state at the body level) — recorded as a
regression pin, not a flip. Rows 2 and 4 are OVER-BLOCK inverse rows
(false deny on the protected path via the cross-generation name×arg
join). Row 3 is the BYPASS: the shared registry let the pre-4.2 world see
the modern world's no-op rm override, allowing the real deletion — the
in-suite reproduction of the reviewer's finding.

**M2.9 remedy — full per-world state isolation (GREEN record).** Shape:
the walker state went per generation — `funcs`/`calling`/`calls` are
`[2]`-arrays (a function execution in one world registers only in that
world's registry; the other world's externals stay external and are judged
as the mutations they are), plus `w.world` (-1 neutral, 0/1 inside a
generation-scoped function body). `zoneCall` runs the per-world loop:
each world's name classifies against ITS OWN registry (shadowing walks
`walkFunctionBodies` entirely inside the world's state — each candidate
body from the entry snapshot, registries merged, `w.world` scoped so
nested statements resolve their own generation), the verb world binds
`pathCands[world]` (its own argument readings), git/sed/cd analyses fire
per world (`zoneGitArgs(world, cmd)`, `zoneCdMove(world, cmd)`).
`FuncDecl` registers per world (generation-identical in both, per-reading
when the name word is dual). The state-level helpers
(`cloneZoneFuncsState`/`mergeZoneFuncsState`/`zoneFuncsEqualState`) carry
the control-flow unions. `zoneWordCandidates` (the pooled variant) is
folded away — every consumer reads per-world now.

- **Command** (all 36 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.422s`
  (36/36 PASS — the registry bypass row now DENIES; the two over-block rows
  now ALLOW; the candidate-state pin holds; the 32 earlier rows hold).
- **Full package regression (M2.9)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	266.232s` /
  `PACKAGE_POST29_EXIT=0`; 3664 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.8%` (all-rows selector).

### Gate round 27 — M2.10 funnel-scoped world binding (2026-10-09)

Three findings on the M2.9 tip (`5065e16ae`, overlay 4 repros
base-PASS/current-FAIL, existing suites PASS): (1) P1 the numeric-FD
decision read the MODERN reading only — `printf changed >& $'1\u0000/../
zone_dir/marker.md'`: "1" is a descriptor in the modern world (writes
nothing) while the pre-4.2 reading is a real path old bash WRITES through
the literally-named 1u0000 entry; (2) P2 the git funnel's inner world loop
SHADOWED the passed generation, generating both worlds' candidates
(`$'git\u0000/../printf' rm -f $'docs\u0000/../zone_dir/marker.md'` —
modern git × the pre-4.2 file reading = false deny); (3) P2 function-body
redirections ignored `w.world` — a function executing only in the modern
world (`f(){ printf changed > $'docs\u0000/../zone_dir/marker.md'; };
$'f\u0000/../printf'`) still judged the pre-4.2 reading's protected path.
Findings 2-3 are OVER-BLOCK inverse rows.

**Three regression rows — RED under the M2.9 tip (`5065e16ae` + the rows,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellNumericFdPerWorld|TestCheckProtectedZoneShellGitCandidatesConfinedToCaller|TestCheckProtectedZoneShellFunctionRedirectsOwnGeneration' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision lines)**:

```
    protected_zone_shell_repro_test.go:1020: numeric fd per world: decision="allow" reason="", want deny
--- FAIL: TestCheckProtectedZoneShellNumericFdPerWorld (0.00s)
    protected_zone_shell_repro_test.go:1036: git candidates confined to caller: decision="$'git\\u0000/../printf' rm -f $'docs\\u0000/../zone_dir/marker.md'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=docs/zone_dir/marker.md", want allowed — modern git joins only the modern file reading, and no world deletes
--- FAIL: TestCheckProtectedZoneShellGitCandidatesConfinedToCaller (0.00s)
    protected_zone_shell_repro_test.go:1052: function redirects own generation: decision="f(){ printf changed > $'docs\\u0000/../zone_dir/marker.md'; }; $'f\\u0000/../printf'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=docs/zone_dir/marker.md", want allowed — the function executes only in the modern world, whose redirect lands on docs
--- FAIL: TestCheckProtectedZoneShellFunctionRedirectsOwnGeneration (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.636s
```

Row 1 is a deny-miss (the bypass); rows 2-3 are OVER-BLOCK inverse rows.
Fixture: the narrowed marker manifest; inputs transport-verified
(whole-file NUL-byte scan zero, doubled backslash).

**M2.10 remedy — funnel-scoped world binding (GREEN record).** Shape:
(1) `zoneRedirectTargets` makes the numeric-descriptor decision PER WORLD
(`>& $'1\u0000/../zone_dir/marker.md'`: the modern reading "1" dup's the
descriptor in its world, while the pre-4.2 reading is a real path that
world WRITES — only all-numeric readings skip); (2) `zoneGitArgs` binds
EVERYTHING to the passed generation — the -C/--work-tree anchors, the file
arguments, and the base loop's generation filter all read
`readings[world]`/`base.gen` (the inner world loops that shadowed the
passed generation and generated cross-generation candidates are gone); (3)
`zoneCands`/`zoneCandsWorld` respect `w.world` — during a
generation-scoped function execution only that world's redirection
candidates judge (the modern-only function no longer judges the pre-4.2
redirect reading's protected path). Own-sweep fix in the same commit: the
--work-tree overwrite-wins semantics (gate-17 P2) were re-broken by the
rewrite (append instead of replace — caught by
TestCheckProtectedZoneShellGitWorkTreeOverwriteWins failing in the
instrument re-run) and restored.

- **Command** (all 34 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.556s`
  (34/34 PASS — the numeric-FD row now DENIES; the two over-block rows now
  ALLOW; the overwrite-wins row re-verified after the in-commit sweep fix;
  the 30 earlier rows hold).
- **Full package regression (M2.10)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	335.301s` /
  `PACKAGE_POST30_EXIT=0`; 3668 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.8%` (all-rows selector).

### Gate round 28 — M2.11 sed in-place generation binding (2026-10-09)

One P2 (over-block): `zoneSedInPlace` applied the MODERN bash's option
reading to every sed dispatch — the executable word is dual such that only
the pre-4.2 world dispatches sed (the modern reading truncates to "no"),
and the option word `$'-i'` decodes to `-i` ONLY in the modern world —
the pre-4.2 reading keeps the escape text literal (an invalid option: sed
exits, touching nothing). The pooled modern scan read "-i" and false-
denied while NEITHER world's sed runs in-place.

**Regression row — RED under the M2.10 tip (`352ce8b70`; measured against
the pre-fix source restored from HEAD, with the row and the M2.11 fix
in flight):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellSedInPlaceOwnGeneration' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision line)**:

```
    protected_zone_shell_repro_test.go:1076: sed in place own generation: decision="$'no\\u0000p/../sed' $'\\u002di' 's/a/b/' zone_dir/marker.md" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/marker.md", want allowed — the executing generation's option reading is the literal escape text, an invalid option: sed touches nothing
--- FAIL: TestCheckProtectedZoneShellSedInPlaceOwnGeneration (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.808s
```

Design iteration disclosed: the first draft prefixed a literal `sed`
command word — under it the MODERN world genuinely dispatched sed with the
modern-decoded `-i` (a true mutation through the modern decode), making
the deny sound and the row unadoptable; the committed executable IS the
dual word itself. Fixture: narrowed marker manifest. Inputs
transport-verified: whole-file NUL-byte scan zero, doubled backslash.

### Gate round 34 (second) — M2.16 escape-origin option world termination (2026-10-09)

One P2 (deny-miss): when a world's git option scan hits an ESCAPE-ORIGIN
option word that world's git would refuse, that world's subcommand search
must TERMINATE. Shape: `git $'-C' rm grep --no-index AGENTS.md` —
the option word decodes to -C only in the modern world (consuming rm as
the directory, leaving the read-only grep); the pre-4.2 reading keeps the
escape text literal — git refuses the unknown option (exit 129) and
NOTHING executes in that world. The pre-4.2 scan that skipped the
unsupported option mis-read rm as the subcommand and false-denied
AGENTS.md via the docs_zone category.

**Regression row — RED under the M2.15 tip (`b4718c6d0` + the row,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellGitUnsupportedOptionTerminatesWorld' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision line)**:

```
    protected_zone_shell_repro_test.go:1214: git unsupported option terminates world: decision="git $'-\\u0043' rm grep --no-index AGENTS.md" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=baseline route=human next=return-blocker-report path=AGENTS.md", want allowed — the pre-4.2 world's git refuses the escape-origin option (exit 129) and nothing executes in that world; the modern world's -C consumes rm leaving the read-only grep
--- FAIL: TestCheckProtectedZoneShellGitUnsupportedOptionTerminatesWorld (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.738s
```

Fixture: a VALID manifest (the seven categories + a docs_zone category
covering AGENTS.md — the first draft's duplicate safety_guards key was
invalid and corrected). The deny category=baseline path=AGENTS.md
confirms the cross-generation mis-read (the baseline floor judged under
the mis-read subcommand). Inputs transport-verified: whole-file NUL-byte
scan zero, doubled backslash.

**M2.16 remedy — GREEN record.** Shape: the zoneGitArgs scan terminates a
world's subcommand search when that world's reading hits an ESCAPE-ORIGIN
option word (dual + the world's reading is the literal escape text): the
pre-4.2 world's git refuses it (exit 129) and NOTHING after executes in
that world — the world contributes no mutation analysis. The modern
world's decoded option proceeds through the existing classification.

- **Command** (all 39 instrument tests): the standard instrument selector
  — exit 0, verbatim: ok github.com/modu-ai/moai-adk/internal/hook
  12.746s (39/39 PASS — the termination row flipped to ALLOW).
- Builds: go build ./... exit 0; GOOS=windows go build ./... exit 0;
  golangci-lint run internal/hook/... --timeout=2m -> 0 issues.; gofmt
  clean; family coverage 13.8% (all-rows selector).

### Gate round 32 — M2.13 over-cap fail-closed regardless of manifest state (2026-10-09)

A REAL P1 on the M2.12 immediate cap (reviewer: base deny → current allow,
real bash deleted AGENTS.md): the over-cap fail-closed return was GATED on
`load.State != ZoneStateAbsent` — in a manifest-less project an over-cap
candidate set skipped ALL path judging, INCLUDING the compiled baseline
floor that protects AGENTS.md and the frozen instruction files WITHOUT any
manifest, and fell to the Absent allow. Repro: `rm -f AGENTS.md` + 4,096
distinct harmless paths.

**Regression row — RED under the M2.12 tip (`8dd61db7d` + the row,
uncommitted at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellOverCapBaselineFloorDenied' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision line; the 3,995-path enumeration
  abbreviated)**:

```
    protected_zone_shell_repro_test.go:1224: over cap baseline floor denied: decision="rm -f AGENTS.md p0 p1 … p3995" reason="", want deny — an over-cap candidate set is unverifiable whether or not a manifest exists, and the compiled baseline floor protects AGENTS.md without one
--- FAIL: TestCheckProtectedZoneShellOverCapBaselineFloorDenied (0.03s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.740s
```

**M2.13 remedy — the over-cap deny applies REGARDLESS of manifest state
(GREEN record).** Shape: the immediate cap site returns the fail-closed
deny directly (category `loop-unbounded`, path `over-cap`, audit row
recorded with the actual ManifestState) — before any Absent handling. The
narrower first-cap-slice shape (judge the first cap-sized slice against
the baseline floor before denying) was not taken: the unconditional
deny is the stronger sound closure and the reviewer's shape demands it.

**Sibling audit — CONFIRMED and closed in the same commit.** The OLD
unbounded-deny gate (`w.unbounded && load.State != ZoneStateAbsent`, the
pre-M2.12 fixed-point carve-out) has the SAME hole: `for ((;;)); do rm
AGENTS.md; done` in a manifest-less project → unbounded → absent → the
baseline floor unprotected. Fixed identically: the unbounded deny is now
UNCONDITIONAL (the Absent carve-out removed). Note for the orchestrator:
this alters the documented REQ-SIPZ-010 degrade-visibly behavior for
unbounded walks in manifest-less projects — the fail-closed philosophy
and the baseline-floor rationale justify it; flagged for the sync audit.

- **Command** (all 38 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	2.541s`
  (38/38 PASS — the gate-32 row flipped to DENY; the 37 earlier rows hold).
- **Full package regression (M2.13)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	267.889s` /
  `PACKAGE_POST32_EXIT=0`; 3675 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.8%` (all-rows selector).

### Gate rounds 29/30/31 — M2.12 subcommand binding, linear dedup, immediate cap (2026-10-09)

Five findings folded: (1) P1 git SUBCOMMAND word generation binding —
`git $'rm\u0000bogus' -f $'docs\u0000/../zone_dir/marker.md'`: modern
`git rm -f docs` (docs ∉ zone) vs pre-4.2 nonexistent subcommand — the
pooled modern subcommand × the pre-4.2 file reading false-denied
(superseding the earlier exact-string single-world exception, whose
cross-generation join was the defect); (2) P2 the sed in-place binding's
plain-file call shape pinned (confirmation); (3) P2 the git NAME word
dual shape — `$'printf\u0000/../git' $'rm\u0000' -f zone_dir/marker.md`:
the pre-4.2 git subcommand is the nonexistent literal text — pooled
modern "rm" false-denied; (4) P1/P2 dedup linearization (hash-set; the
per-candidate scan was O(n²): 80,000-arg repro measured 11.89s in-suite)
+ IMMEDIATE fail-closed on the unique-candidate cap (no resolution of
over-cap sets); (5) P2 zoneRedirects' mutating flag respects `w.world`
(a never-executed generation's redirect reading must not flip it).

**Rows — measured under the M2.11 tip (`103de893d` + the rows, uncommitted
at measurement):**

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellGitSubOwnGeneration|TestCheckProtectedZoneShellGitNameSubOwnGeneration|TestCheckProtectedZoneShellSedInPlacePlainFileShape|TestCheckProtectedZoneShellGitMassFileArgsBounded' -count=1 -v`
- **Exit code**: `1`
- **Observed (verbatim, decision/measurement lines)**:

```
    protected_zone_shell_repro_test.go:1097: git sub own generation: decision="git $'rm\\u0000bogus' -f $'docs\\u0000/../zone_dir/marker.md'" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=docs/zone_dir/marker.md", want allowed — modern git rm joins only the modern file reading (docs, outside the zone), and the pre-4.2 subcommand is nonexistent
--- FAIL: TestCheckProtectedZoneShellGitSubOwnGeneration (0.01s)
    protected_zone_shell_repro_test.go:1114: git name sub own generation: decision="$'printf\\u0000/../git' $'rm\\u0000' -f zone_dir/marker.md" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/marker.md", want allowed — the pre-4.2 world's git subcommand is the nonexistent literal rm\u0000 text, and the modern world's name is printf
--- FAIL: TestCheckProtectedZoneShellGitNameSubOwnGeneration (0.00s)
--- PASS: TestCheckProtectedZoneShellSedInPlacePlainFileShape (0.00s)
    protected_zone_shell_repro_test.go:1160: elapsed=11.893514125s (bounded-run measurement; before/after in progress.md §E.2)
--- PASS: TestCheckProtectedZoneShellGitMassFileArgsBounded (11.90s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	14.174s
```

Rows 1-2 are OVER-BLOCK inverse rows (false deny observed verbatim). Row 3
is a GREEN-NOW confirmation pin (the M2.11 binding covers the plain-file
shape). Row 4 is a MEASUREMENT pin: the pre-fix duration 11.89s in-suite
(the reviewer's deployment measurement: 11.70s vs 0.97s base, over the
deployed hook's 10s limit); the row asserts only the bounded deny and
logs the duration.

**M2.12 remedy — GREEN record.** Shape: (1) the git SUBCOMMAND word binds
its own generation (`sub = readings[world]` via zoneWordWorldReadings —
superseding the earlier exact-string single-world exception, whose
cross-generation join was the defect); the option-structure scan stays on
the modern reading (bounded residual: a dual word whose readings disagree
on option-vs-subcommand classification); (2) `zoneDedupStrings` is
hash-set based (linear — the O(n²) scan dominated over-cap inputs);
(3) the unique-candidate cap fires fail-closed IMMEDIATELY — an over-cap
set skips per-candidate resolution entirely; (4) `zoneRedirects`' mutating
flag respects `w.world` — during a generation-scoped function execution
only that world's redirect readings flip it.

- **Command** (all 40 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	2.354s`
  (40/40 PASS — the two git over-block rows flipped to ALLOW; the
  plain-file pin and the measurement pin hold; the 36 earlier rows hold).
- **Bounded-run measurement**: post-fix elapsed 1.43s for the 80,000-arg
  shape (was 11.89s pre-fix in-suite — 8.3×), fail-closed via
  `loop-unbounded` (the immediate cap).
- **Full package regression (M2.12, final tree)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	320.117s` /
  `PACKAGE_POST31_EXIT=0`; 3669 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.8%` (all-rows selector).

**M2.11 remedy — sed in-place generation binding (GREEN record).** Shape:
`zoneSedInPlace(world, args)` reads each option word through the PASSED
generation's reading (`readings[world]` via zoneWordWorldReadings) — the
per-world sed dispatch in zoneCall passes its own world, so a
generation-scoped sed judges only its own option spellings.

- **Command** (all 35 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.029s`
  (35/35 PASS — the sed over-block row flipped to ALLOW; the 34 earlier
  rows hold).
- **Full package regression (M2.11, final tree)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 1 with ONE environmental flake
  OUTSIDE the instrument: `TestScanWriteContentNoConfigNoTempFile`'s
  control subtest expected exactly 1 security-scan temp file and saw 3
  (the t1356-class temp-file sensitivity; the row passed the two prior
  full-package runs AND passed 3× in isolation immediately after).
  Everything in the instrument: zero failures. Slot lease `hook-suite`
  held for the run, released after; the flake re-verified green.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.8%` (all-rows selector).



**M2.7 remedy — full dispatch pre-classification + function shadowing
(GREEN record).** Shape: `zoneExecNames` (the possible base names of the
executable word, one per generation, deduped) classifies EVERY name-driven
dispatch up front — `verbName`/`cdName`/`sedName`/`gitName` plus
`funcNames` for names SHADOWED by a declared function (a shadowed world
executes the function, dropping out of every other dispatch — fixing the
M2.6-introduced over-block where the mutation targets registered before
the function registry was consulted). The analyses fire per world in
order: mutation candidates → sed in-place → git analysis (now
`zoneGitArgs` returns whether it recognized a mutating subcommand) →
function-body walks per shadowing world → the cd move LAST (its directory
change applies to subsequent statements, never to this command's other
worlds — the generalization of the gate-19 fix; the git/sed dispatch no
longer early-returns behind the cd branch). `zoneMutationVerbName` is
folded into `zoneExecNames` (removed).

- **Command** (all 31 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.340s`
  (31/31 PASS — the two gate-21 rows flipped: the git-name shape now DENIES
  through the per-world git analysis; the shadowing shape now ALLOWs).
- **Full package regression (M2.7)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	296.738s` /
  `PACKAGE_POST27_EXIT=0`; 3660 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.6%` (all-rows selector).


**M2.6 remedy — dispatch order, per-generation joins, dedup-before-cap
(GREEN record).** Shape: (1) the dual-world verb classification moved BEFORE
any branch dispatch — `zoneMutationVerbName` fires at the top of the
command analysis, so a word whose modern reading truncates to `cd` or to a
declared-function name still runs the mutation analysis for its pre-4.2
world; the modern dispatch (functions/cd/sed/git) still runs below, and the
candidate set unions both worlds' effects; (2) git file arguments keep
their OWN generation's readings (`fileArgs [2][]string` via
`zoneWordWorldReadings`) — generation i's directory joins generation i's
file arguments only, so the cross-generation join no longer exists; (3)
dedup runs BEFORE the cap at judgment (`zoneDedupStrings` then the
`zoneCandidateCap` check) — a unicode-free command's both-worlds-identical
candidates collapse first and a plain 2,050-file command fits the cap.

- **Command** (all 29 instrument tests): `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/hook -run 'TestCheckProtectedZoneShell|TestZoneUnescapeAnsiC|TestZoneWordText' -count=1`
- **Exit code**: `0`
- **Observed (verbatim)**: `ok  	github.com/modu-ai/moai-adk/internal/hook	1.413s`
  (29/29 PASS — the three gate-19 rows flipped: verb-before-dispatch now
  DENIES; the two over-block rows now ALLOW).
- **Full package regression (M2.6)**: `unset MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1
  -timeout=25m -v ./internal/hook/` — exit 0, verbatim tail `PASS` / `ok
  github.com/modu-ai/moai-adk/internal/hook	281.102s` /
  `PACKAGE_POST26_EXIT=0`; 3658 RUN lines, ZERO `--- FAIL` lines. Slot
  lease `hook-suite` held for the run, released after.
- Builds: `go build ./...` exit 0; `GOOS=windows go build ./...` exit 0;
  `golangci-lint run internal/hook/... --timeout=2m` → `0 issues.`; gofmt
  clean; family coverage `13.5%` (all-rows selector).

**Gate-31 :635 pin (green-now, appended post-M2.12):**
TestCheckProtectedZoneShellInvalidManifestRedirectAllow — with an INVALID
manifest, a never-executed generation's redirect reading must not flip the
mutating flag: the fail-closed invalid-manifest denial (REQ-SIPZ-009)
requires a MUTATING command, and the fd-duplicating function mutates
nothing in either world. Measured ALLOW on the M2.12 tip (the
numeric-descriptor skip plus the w.world filter already cover it) —
recorded as a green-now regression pin, not a flip.



### The :635 blocker RULING — DENY-as-designed (2026-10-09, leader option ii)

The leader's ruling on the excluded redirect row: the lexical deny is
SOUND. The redirect-target word's gen1 (pre-4.2) reading textually reaches
zone_dir/marker.md through Clean — and gen1 is a REAL bash generation (3.2
passes the literal text), so on a 3.2 host with docs\u0000 planted as a
directory the deny is CORRECT (the exact bypass class gate-13 measured
with real deletion). On hosts without the planting the deny is an
over-block — the guard's documented static lexical design (sound
over-approximation: deny when ANY possible reading covers the zone), the
same behavior every other lexical row encodes. An existence-sensitive
redirect-resolution upgrade is a design change beyond this run's scope —
NOT attempted; noted as possible follow-up design material, not claimed
as debt without an upgrade trigger.

**Row landed pinned DENY-as-designed:**
TestCheckProtectedZoneShellFileRedirectDenyAsDesigned — measured DENY
(green-now pin, the lexical deny as designed; the M2.12-era shape measured
the same deny as a false deny under the old want-allow framing).

### Gate round 33 — M2.14 empty-target mutating filter (2026-10-09)

One P2 (over-block): an EMPTY redirect target (printf read-only > '',
invalid-manifest project) flipped the mutating flag in both worlds — bash
cannot modify a nonexistent empty path. The per-world filter: the flag
flips only when a NON-EMPTY target candidate is added; empty readings add
nothing and flip nothing.

**Regression row — RED under the M2.13 tip (50d686b9e + the row,
uncommitted at measurement):**

- **Command**: unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR
  MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run
  'TestCheckProtectedZoneShellEmptyTargetNotMutating' -count=1 -v
- **Exit code**: 1
- **Observed (verbatim, decision line)**:

    protected_zone_shell_repro_test.go:1117: empty target not mutating: decision="printf read-only > ''" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner manifest=invalid route=human next=return-blocker-report path=.moai/config/sections/protected-zone.yaml", want allowed — bash cannot modify a nonexistent empty path, so the fail-closed invalid-manifest denial must not fire
--- FAIL: TestCheckProtectedZoneShellEmptyTargetNotMutating (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.942s

**M2.14 remedy — GREEN record.** Shape: zoneRedirectTargets sets the
mutating flag only when a NON-EMPTY target candidate is added — empty
readings add nothing and flip nothing.

- **Command** (all 40 instrument tests): the standard instrument selector
  — exit 0, verbatim: ok github.com/modu-ai/moai-adk/internal/hook
  4.171s (40/40 PASS — the empty-target row flipped to ALLOW;
  the DENY-as-designed pin holds; the 38 earlier rows hold).
- **Full package regression (M2.14)**: exit 0, verbatim tail PASS / ok
  github.com/modu-ai/moai-adk/internal/hook 469.582s /
  PACKAGE_POST33_EXIT=0; 3677 RUN lines, ZERO FAIL lines. Slot lease
  hook-suite held for the run, released after.
- Builds: go build ./... exit 0; GOOS=windows go build ./... exit 0;
  golangci-lint run internal/hook/... --timeout=2m -> 0 issues.; gofmt
  clean; family coverage 13.8% (all-rows selector).

### Gate round 34 — M2.15 git global-option generation binding (2026-10-09)

One P2: the git GLOBAL option classification and argument consumption
read zoneWordText (modern) while the file args use the passed generation —
the last zoneWordText consumer inside zoneGitArgs. Reviewer shape:
git -c color.ui=false rm -f <protected path> — modern: -c global option
consumed, rm on the modern file reading; 3.2: the literal escape text is
an unknown option (git refuses, nothing happens) — the cross-join
false-denied.

**Row landed as a GREEN-NOW pin:**
TestCheckProtectedZoneShellGitGlobalOptionOwnGeneration (the reviewer's
literal call shape, a/marker.md outside the zone) — measured ALLOW on the
M2.15 tip. Honest classification: green-now — the pre-fix classification
also allowed this literal shape (both generations read the un-escaped
-c identically); the overlay deny hinges on the zone-reaching paths not
carried in the abbreviated shape. The structural fix removes the class:
the scan loop reads readings[world] for the global-option classification,
the --work-tree= attached prefix, and the -C/--work-tree argument
consumption.

**M2.15 remedy — GREEN record.** Shape: the zoneGitArgs scan loop reads
readings[world] for the classification and consumption — the -C/
--work-tree value consumption, the --work-tree= attached prefix, and the
valued-option list all bind per world; the option-structure sequence
stays single-pass (bounded residual: a dual word whose generations
disagree on option-vs-argument classification forks the parse —
disclosed).

- **Command** (all 46 instrument tests): the standard instrument selector
  — exit 0, verbatim: ok github.com/modu-ai/moai-adk/internal/hook
  2.893s (46/46 PASS).
- Builds: go build ./... exit 0; GOOS=windows go build ./... exit 0;
  golangci-lint run internal/hook/... --timeout=2m -> 0 issues.; gofmt
  clean; family coverage 13.8% (all-rows selector).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-09
run_commit_sha:
  M1: c34021856   # instrument finalization + RED baseline re-confirmation (five new rows; verbatim §E.2)
  M2: f4a0227f3   # decoder repair — NUL part-terminator (part-level), raw-byte \x render, bounded no-digit arm
  M3: 7a914891f   # family re-run + regression confirmation (backfilled per D3 — a commit cannot cite its own SHA in its own commit)
ac_pass_count: 11/11   # AC-HZS-001..011 — 001/003/005 RED→GREEN flips; 006/009/010 RED earned at M1 then flipped; 007 pin green both sides; 002/004/011 controls green both sides; 008 family green + windows build
ac_fail_count: 0
preserve_list_post_run_count: 0   # measured: `git diff 9b6ae0da5..HEAD -- internal/hook/protected_zone_guard_test.go internal/hook/protected_zone_path.go internal/hook/pre_tool.go internal/hook/protected_zone_guard.go` EMPTY
l44_pre_commit_fetch: not-run (card worktree lane flow — origin/develop sync is the leader's batch act; the card branch merges locally via the integration window)
l44_post_push_fetch: n/a (no push from the lane)
new_warnings_or_lints_introduced: 0   # golangci-lint internal/hook 0 issues both sides; gofmt clean; vet clean
cross_platform_build:
  darwin: exit 0   # go build ./... (M1 pre-repair and M2/M3 post-repair trees)
  windows: exit 0  # GOOS=windows go build ./... (same trees)
total_run_phase_files: 3   # internal/hook/protected_zone_shell.go, internal/hook/protected_zone_shell_repro_test.go, progress.md — the spec.md draft→in-progress transition landed absorbed in manager-spec's concurrent gate round 8 commit 89d52e86f (attribution deviation reported; the transition VALUE is in-progress at HEAD)
m1_to_mN_commit_strategy: per-milestone commits (M1 instrument+evidence, M2 repair+flip record, M3 family re-run+signal); no fixup/amend; card id in every subject
gaps: >-
  (1) The M1 package-wide baseline was measured on a run CUT at 12m (cold
  build cache + load; fail inventory recorded verbatim); the completing
  package-wide green run is the M3 one above (warm cache, 463.797s, exit 0) —
  remote CI remains the integrated judge per lane protocol. (2) The E4
  boundary grep measures 1 pre-existing match (pre_tool.go:856 observation
  branch, present at the frozen tip) vs the dispatch's 0-match expectation —
  not introduced by this run. (3) acceptance.md's M1-final confirmation
  selector sweeps zero tests (measured [no tests to run]) — manager-spec's
  artifact, finding reported not edited. (4) Concurrent manager-spec gate
  rounds 8/9 (89d52e86f, 28c03f10c) committed to this tree mid-run — disjoint
  file sets, verified no absorption (the spec.md transition absorption in
  gate round 8 is the exception, itemized above).
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-10-09
sync_commit_sha: "152464221"   # backfilled per D3 — the sync commit cannot cite its own SHA (preceding commit 152464221)
changelog_entry_position: CHANGELOG.md [Unreleased] › ### Fixed › first entry (SPEC-HOOK-ZONE-SHELL-ESCAPE-001)
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed"   # single sync commit, 3-phase close; updated: 2026-10-09 (already current)
  plan_md: "no status field (stateless artifact); updated: 2026-10-09 already current — no edit"
  acceptance_md: "no status field (stateless artifact); updated: 2026-10-09 already current — no edit"
b12_self_test_a: pass   # pre-emission `grep -c 'SPEC-HOOK-ZONE-SHELL-ESCAPE-001' CHANGELOG.md` → 0 (exit 1) before append — no duplicate entry
b12_self_test_b: pass   # AC counter (manager-docs § B12 awk grammar on acceptance.md): live=11 excluded=0 ambiguous=0 — matches the entry's 11건 AC-HZS-001..011
b12_self_test_c: pass   # entry file paths verified via ls internal/hook/: protected_zone_shell.go + protected_zone_shell_repro_test.go both present; the entry cites these two only
mx_tag_validation: "added=1 removed=0 updated=0"   # @MX:DEBT(+CEILING/UPGRADE) on zoneUnescapeAnsiC — the \u/\U host-variance residual (spec §B/§F follow-up-card material); existing @MX:SPEC on checkProtectedZoneShell untouched
readme_docs_site: no-op   # internal guard repair — the protected-zone ANSI-C decoder has no README feature-list or docs-site surface
```

## §G Override and Refusal Record

- 2026-10-09T01:23:40Z SPEC-HOOK-ZONE-SHELL-ESCAPE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1585/.moai/reports/t1585/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1585/.moai/reports/t1585/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1585/.moai/reports/t1585/plan-audit.md
- 2026-10-09 (operator decision, relayed via the factory leader) SPEC-HOOK-ZONE-SHELL-ESCAPE-001 ceiling-admission outcome=ADMIT reasons="the operator approves run-phase entry on the iter-3 verdict despite the round-count ceiling: the trajectory is monotone (0.79 FAIL -> 0.84 FAIL -> 0.94 PASS-WITH-DEBT, no STOP signal), iter-3 carries blocking count 0, the single debt item is disposed in-run, and the run-phase work itself is COMPLETE and green (46 instrument tests, family exit 0, builds/lint clean) — the admission regularizes the F1 record per REQ-ACE-006's operator-decision release path" evidence=.moai/reports/t1585/plan-audit-iter3.md receipts=rcpt-dfffac88dbe5f6b32de44dcb,rcpt-dd78adc9776bd35482495236,rcpt-55dad295a91674791a673a7e recorded_by=lane-19 (relaying the operator decision via the factory leader)
