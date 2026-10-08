---
id: SPEC-HOOK-ZONE-SHELL-ESCAPE-001
title: "Acceptance — ANSI-C shell decoder repair"
version: "0.1.4"
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
---

# SPEC-HOOK-ZONE-SHELL-ESCAPE-001 — Acceptance Criteria

> Stateless artifact. Every AC is machine-verifiable and follows the two-cell
> adoption rule (`.claude/rules/moai/development/verification-completeness.md`
> §2): a RED-now cell (command + verbatim stdout + exit code + tree SHA) and a
> green-path cell naming what flips it. UNLIKE the t1570 predecessor — whose
> RED cells were recorded at run-phase M1 — this card's RED cells were
> MEASURED BEFORE the SPEC was authored (lane-19, 2026-10-08). The canonical
> record is the TRACKED file `evidence-red-repro.md` in this SPEC directory
> (committed — readable from a fresh checkout; it carries the 2026-10-09 byte
> corrections and the full ground-truth re-measurement on bash 3.2.57).
> Additional committed carriers: the instrument
> `internal/hook/protected_zone_shell_repro_test.go` (commit `9dbe40c0a`,
> header ground-truth comment) and the verbatim RED strings quoted in the AC
> cells below. The gitignored `.moai/reports/t1585/red-repro.md` is the
> demoted card-scoped duplicate — never the citation target. Selector forms:
> the AC cells quote the 2026-10-08 measured command verbatim (unanchored);
> plan.md §C's re-run recipe uses the anchored form — same match set today,
> the anchored form is canonical for re-runs. The control ACs (002, 004) are
> green on BOTH sides of the repair by design — they are mutant probes, and
> their first cell is the MEASURED pre-repair green at `7be9f41b5` (the
> review-gate instrument fix split the controls into independent tests; both
> PASS observed on this tree).

## D. AC Matrix

### AC-HZS-001 — NUL-truncation bypass denied (release-blocking; REQ-HZS-001)

Maps REQ-HZS-001

**Given** a project root whose manifest declares `probe_zone: paths: ["zone_dir/"]`
with a populated `zone_dir` (marker present),
**When** the Bash guard receives `rm -r zone_dir$'\x00/sub'`,
**Then** the decision is `deny` with the zone deny reason — and because bash
truncates the ANSI-C part at the NUL, the judged word is `zone_dir`, the
command's real target. The protected marker still exists after the decision.

Two-cell adoption:

- **RED-now cell** (MEASURED 2026-10-08, red-repro.md §Defect ①):
  - **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellAnsiCNulTruncationBypass|TestCheckProtectedZoneShellHexRawByteBypass|TestZoneUnescapeAnsiCNoDigitHexStaysLiteral' -count=1 -v`
  - **Observed stdout (excerpt; full 3-test run in the tracked record
    `evidence-red-repro.md` §Defect ①)**: `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected directory is GONE (rm output: "")`
  - **Exit code**: `1`
  - **Tree SHA**: measured at HEAD `81786284e` + the untracked instrument;
    committed verbatim at `9dbe40c0a` (re-run reproduces the RED there).
- **Green path cell**: M2 ① (the NUL part-terminator) flips it — exit 0,
  `wantZoneDeny` observed, marker intact.

### AC-HZS-002 — Mutant probe: NUL truncating OUTSIDE the zone stays allowed (control; REQ-HZS-001)

Maps REQ-HZS-001

**Given** the AC-HZS-001 fixture plus an unprotected `docs` directory,
**When** the Bash guard receives `rm -r docs$'\x00/zone_dir'` (truncates to
`docs` — outside the zone),
**Then** the decision is NOT deny and carries no zone sentinel: the truncation
semantics apply uniformly, and a `docs`-targeting removal is not the guard's
business.

- **Cell 1 (pre-repair green, MEASURED)**:
  `TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl` — PASS at
  `7be9f41b5` (this tree, 2026-10-09): decision allow for the outside-zone
  truncation; the review-gate instrument fix split the control out of the
  bypass test into its own function, so the pre-repair green is observed, not
  expected.
- **Cell 2 (post-repair green)**: M2 keeps it green — a fix that
  blanket-denies every NUL-bearing command fails HERE, not on AC-HZS-001.

### AC-HZS-003 — Raw-byte `\x` path bypass denied (release-blocking; REQ-HZS-002)

Maps REQ-HZS-002

**Given** a project root whose manifest declares a protected directory whose
name is the multi-byte spelling `존` (bytes `ec a1 b4`), with a marker inside,
**When** the Bash guard receives `rm $'\xec\xa1\xb4/marker.md'` (the target
spelled as one ANSI-C word of raw-byte escapes),
**Then** the decision is `deny` — the decoded word text carries the raw bytes
`ec a1 b4/marker.md`, the same bytes bash places in the executed argument, and
the zone entry matches them.

Two-cell adoption:

- **RED-now cell** (MEASURED 2026-10-08, `evidence-red-repro.md` §Defect ②): same
  command as AC-HZS-001; **Observed stdout (excerpt; full run in the tracked
  record)**: `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected marker is GONE (rm output: "")`; **Exit code**: `1`; **Tree SHA**: `81786284e` + untracked instrument / committed `9dbe40c0a`.
  Fixture note (tracked record §Defect ②): the row's escapes derive from the DECLARED
  directory name's measured bytes, so the decoded word and the created
  directory are the same bytes by construction.
- **Green path cell**: M2 ② (`\x` → `byte(val)`) flips it — exit 0, deny
  observed, marker intact.

### AC-HZS-004 — Control: the direct spelling stays denied (control; REQ-HZS-002/006)

Maps REQ-HZS-002, REQ-HZS-006

**Given** the AC-HZS-003 fixture,
**When** the Bash guard receives `rm 존/marker.md` (the same target written
literally),
**Then** the decision is `deny` — proving the manifest and the folding are not
the defect: the zone entry itself works on the pre-repair tree; defect ② is
confined to the decoder.

- **Cell 1 (pre-repair green, MEASURED)**:
  `TestCheckProtectedZoneShellHexDirectSpellingControl` — PASS at `7be9f41b5`
  (this tree, 2026-10-09): decision deny with the zone reason carrying
  `category=probe_zone` — the zone entry itself works on the pre-repair tree;
  defect ② is confined to the decoder.
- **Cell 2 (post-repair green)**: M2 keeps it denied.

### AC-HZS-005 — No-digit escapes render literally, no panic (release-blocking; REQ-HZS-004)

Maps REQ-HZS-004

**Given** the decoder surface `zoneUnescapeAnsiC` (panic contained per row by
the instrument's helper),
**When** the rows `\x`, `\u`, `\U`, `\xZ`, `a\xZb` decode,
**Then** every row returns its input verbatim (the bash literal — backslash
and prefix letter kept, following text surviving once) and no row panics.

Two-cell adoption:

- **RED-now cell** (MEASURED 2026-10-08, `evidence-red-repro.md` §Defect ③): same
  command as AC-HZS-001; **Observed stdout (excerpt; full run in the tracked
  record)**: three `PANICKED —
  slice bounds out of range on the no-digit escape` rows (`\x`, `\u`, `\U`)
  plus two mangling rows (`"\xZ": decoded "xZZ", want the bash literal
  "\xZ"`; `"a\xZb": decoded "axZZb", want ...`); **Exit code**: `1`;
  **Tree SHA**: `81786284e` + untracked instrument / committed `9dbe40c0a`.
- **Green path cell**: M2 ③ (bounded literal return) flips it — exit 0, all
  five rows literal.

### AC-HZS-006 — The walk completes on malformed escape text (release-blocking; REQ-HZS-005)

Maps REQ-HZS-005

**Given** the zone fixture of AC-HZS-001,
**When** the Bash guard receives a command carrying `$'\x'` (a no-digit
escape embedded in a real call),
**Then** the guard returns a DECISION — the walk runs to completion; the
instrument row contains the panic in its test helper (the `hzsDecodeAnsiC`
pattern extended to the guard call), so pre-fix it reports the walk panicking
instead of crashing the binary.

- **RED-now cell**: the decoder-level panic is MEASURED (AC-HZS-005's RED);
  the guard-level propagation is code-derived — `zoneWordText →
  zoneUnescapeAnsiC → zoneHexEscape → walker → checkProtectedZoneShell
  (`protected_zone_shell.go:1125`) → `pre_tool.go:789` carries no `recover()`.
  The guard-level row is AUTHORED at M1 and is expected RED there (walk
  panics); the verbatim M1 observation is recorded into the ledger below.
  Release-blocking eligibility is earned at M1 upon the recorded RED
  observation (four elements); absent a reproducing RED this criterion
  demotes to regression-guard (verification-completeness §2.1).
- **Green path cell**: M2 ③ flips it — the row observes a decision (deny or
  allow per policy for the non-zone target), never a panic.

### AC-HZS-007 — `\u`/`\U` code-point rendering pinned (regression pin; REQ-HZS-003)

Maps REQ-HZS-003

**Given** the decoder surface `zoneUnescapeAnsiC`, which receives the INNER
TEXT of a `$'...'` part (the pin therefore names the ESCAPE TEXTS, not the
literal character — a literal `⊇` would exercise only the pass-through path),
**When** the escape texts — byte 0x5C followed by `u2287`, and `\U00002287` — decode (the raw
backslash-letter TEXT forms fed to zoneUnescapeAnsiC — a literal `⊇`
character returns via the backslash-free early path and pins nothing: a
mutant that breaks the `\u` branch, e.g. maxDigits 4→2, must FAIL here),
**Then** both render the three UTF-8 bytes `e2 8a 87` (U+2287; the Go decoder
output — string(rune(0x2287)) — and the modern-bash (≥4.2) documented
expansion). Host-bash note: BOTH escape texts render LITERALLY on this
host's bash 3.2.57 (no `\u`/`\U` support — decisive od-proven
re-measurement); the divergence carries a support-boundary + residual-risk
statement (spec.md §B) — it is documented residual, not deny-safe.
Green-now by design (the current decoder is correct for
`\u`/`\U`); the pin's value is POST-repair, verified again at M3.

- **Cell 1**: green on the pre-repair tree (the `\u` arm decodes correctly
  today); the rows execute at M1.
- **Cell 2**: green after M2 and at M3 — the split keeps the code-point arm
  for both prefixes.

### AC-HZS-008 — Family green + windows build (regression guard; REQ-HZS-006)

Maps REQ-HZS-006

**Given** the M2 repair,
**When** `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/` runs,
**Then** the package is green: the t1570 quoting matrix (7 deny + 2 allow),
all zone suites, and the finalized instrument — the ONLY delta vs the M1
baseline is the three defect rows flipping green (countable, no new failure);
and `GOOS=windows go build ./...` exits 0.

- **Cell 1**: the plan-time measured scope is the 3-test selected run (exit
  1, only the three defect rows failing — tracked record
  `evidence-red-repro.md`); the package-wide green baseline is recorded at
  M1's own instrument-confirmation run (full
  `go test -count=1 ./internal/hook/`, four elements, pre-repair).
- **Cell 2**: M3 — full green + windows build, verbatim outputs in §E.2.

### AC-HZS-009 — Part-level NUL truncation shape pinned (release-blocking; REQ-HZS-001)

Maps REQ-HZS-001

**Given** a word composed of an ANSI-C part `$'a\x00b'` followed by a literal
part `X` (the bash-measured shape `$'a\x00b'X`),
**When** `zoneWordText` assembles the word,
**Then** the judged text is `aX` — the ANSI-C part's contribution ends at the
NUL and the LATER PART still appends. This row fails a word-level-truncation
mutant that satisfies the command rows of AC-HZS-001/002 (word truncation
yields the same judged text there) while violating REQ-HZS-001's pinned
part-level shape (plan-audit D4).

Two-cell adoption (verification-completeness §2):

- **RED-now cell**: the bash side is MEASURED (`evidence-red-repro.md`
  ground-truth row 1: `$'a\x00b'X` → `61 58` = `aX`; re-measured 2026-10-09
  on bash 3.2.57); the pre-repair decoder side is code-derived —
  `zoneUnescapeAnsiC` keeps the NUL (`string(rune(0))`,
  `protected_zone_shell.go:221`) and the text after it, so today's
  `zoneWordText` output is `a\x00bX` ≠ `aX`. The instrument row is AUTHORED
  at M1 and is expected RED there (decoded `a\x00bX`); the verbatim M1
  observation is recorded into the ledger below. Release-blocking eligibility
  is earned at M1 upon the recorded RED observation (four elements); absent a
  reproducing RED this criterion demotes to regression-guard (§2.1).
- **Green path cell**: M2 ① flips it — the row observes `aX`, exit 0.

### AC-HZS-010 — Octal-origin NUL command denied (release-blocking; REQ-HZS-001)

Maps REQ-HZS-001

**Given** the AC-HZS-001 fixture,
**When** the Bash guard receives `rm -r zone_dir$'\0/sub'` (the NUL spelled
in OCTAL, not hex),
**Then** the decision is `deny` — the octal escape reaches the same NUL byte
through `zoneOctalEscape`, so the truncation semantics are
origin-independent (bash ground truth: `$'a\0b'` → `61`, tracked record row
3, re-measured 2026-10-09).

Two-cell adoption:

- **RED-now cell**: the bash side is MEASURED (tracked record row 3,
  re-measured 2026-10-09); the pre-repair guard side is code-derived —
  `zoneOctalEscape` renders `byte(0)` and the decoder keeps the trailing
  text: the identical judged-text path as the measured hex case, so the
  expected pre-repair decision is allow. The instrument row is AUTHORED at M1
  and is expected RED there; the verbatim M1 observation is recorded into the
  ledger below. Release-blocking eligibility is earned at M1 upon the
  recorded RED observation (four elements); absent a reproducing RED this
  criterion demotes to regression-guard (§2.1).
- **Green path cell**: M2 ① flips it — deny observed, judged word
  `zone_dir`.

### AC-HZS-011 — Mutant probe: a non-ASCII path OUTSIDE the zone stays allowed, literal AND raw-byte (control; REQ-HZS-006)

Maps REQ-HZS-006

**Given** a project root with a non-ASCII-named file OUTSIDE the zone (e.g.
`개요.md` at the root),
**When** the Bash guard receives `rm 개요.md` (literal spelling) and
`rm $'<raw-byte \xNN escapes of the same name>'` (escapes derived from the
declared name's bytes by construction, as in AC-HZS-003),
**Then** BOTH stay allowed pre- and post-repair — a fix (or mutant) that
blanket-denies non-ASCII or raw-byte-bearing commands fails HERE. This
control closes the non-ASCII half of spec.md §E.2's mutant claim
(plan-audit D5).

- **Cell 1 (pre-repair green)**: expected allow on the pre-repair tree (the
  literal form carries no backslash — the ANSI-C decoder early-returns; the
  raw-byte form decodes to non-matching text today); the rows are AUTHORED at
  M1 and recorded there (regression-guard class — not release-blocking, no
  §2.1 RED owed).
- **Cell 2 (post-repair green)**: M2 keeps both allowed — the raw-byte fix
  makes the decoded text match the real (outside-zone) path, and the
  outside-zone decision is unchanged.

## Evidence Ledger

The binding RED observations were recorded 2026-10-08; the canonical record
is the TRACKED `evidence-red-repro.md` in this SPEC directory (committed with
the round-2 repair — readable from a fresh checkout; it carries the
2026-10-09 byte corrections and the full ground-truth re-measurement on bash
3.2.57). The gitignored `.moai/reports/t1585/red-repro.md` is the card-scoped
duplicate, demoted — verified never-committed (`git cat-file -e
9dbe40c0a:.moai/reports/t1585/red-repro.md` exit 128) and never cited
canonically. Additional committed carriers: the instrument header comment
(`9dbe40c0a`) and the verbatim RED strings quoted in the AC cells above.
This ledger adds the rest: the measured control observations (`7be9f41b5`),
the M1 instrument-confirmation observations (including AC-HZS-009/010/011's
rows), and the GREEN flip records appended at M2/M3.

### RED confirmation at the committed baseline (appended at M1)

- **Command**: the AC-HZS-001 command, run at `9dbe40c0a`.
- **Expected**: exit 1, the three defect rows failing exactly as
  red-repro.md records them; controls green.
- **Observed**: *(populated at M1)*

### CONTROLS at `7be9f41b5` — pre-repair green, MEASURED (review-gate round 1, 2026-10-09)

- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run '^(TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl|TestCheckProtectedZoneShellHexDirectSpellingControl)$' -count=1 -v`
- **Observed (verbatim tail)**:

```
--- PASS: TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl (0.01s)
--- PASS: TestCheckProtectedZoneShellHexDirectSpellingControl (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.947s
```

- **Exit code**: 0 — **Tree SHA**: `7be9f41b5` (the instrument fix only; the
  decoder is untouched — pre-repair).
- Direct-spelling deny reason observed verbatim:
  `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=존/marker.md`

### GREEN flips (appended at M2)

- **Observed**: *(populated at M2 — exit 0 on the defect-row command,
  `wantZoneDeny` lines, controls still green)*

### Family re-run (appended at M3)

- **Observed**: *(populated at M3 — package-wide `ok` line + windows build
  exit 0)*

## Quality Gates

- TRUST 5: Tested (this matrix + package coverage ≥ pre-change), Readable,
  Unified (gofmt), Secured (the repair closes a measured P1 bypass),
  Trackable (Conventional Commits carrying the card id; the baseline-first
  commit `9dbe40c0a` carries the ordering attribution).
- Definition of Done: AC-HZS-001..008 all GREEN with both cells recorded;
  ledger complete; progress.md §E.2/§E.3 populated; no PRESERVE-list file
  touched.
