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
