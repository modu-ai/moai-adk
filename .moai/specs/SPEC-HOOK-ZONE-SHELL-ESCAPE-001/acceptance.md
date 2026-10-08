---
id: SPEC-HOOK-ZONE-SHELL-ESCAPE-001
title: "Acceptance — ANSI-C shell decoder repair"
version: "0.1.0"
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
> MEASURED BEFORE the SPEC was authored (lane-19, 2026-10-08) and are recorded
> in the committed evidence file `.moai/reports/t1585/red-repro.md` (commit
> `9dbe40c0a`); the ledger below cites them rather than promising future
> measurement. The re-run command reproduces the RED at the committed
> baseline. The two control ACs (002, 004) are green on BOTH sides of the
> repair by design — they are mutant probes, and their first cell records the
> pre-repair green at M1's instrument-confirmation run (the RED runs' early
> return skipped them; red-repro.md §Gaps).

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
  - **Observed stdout (verbatim)**: `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected directory is GONE (rm output: "")`
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

- **Cell 1 (pre-repair green)**: expected `allow` on the pre-repair tree too
  (the broken decode `docs\x00/zone_dir` also matches no entry); the row
  executes and records at M1's instrument-confirmation run (the RED runs'
  demonstration branch returned early — red-repro.md §Gaps row 1).
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

- **RED-now cell** (MEASURED 2026-10-08, red-repro.md §Defect ②): same
  command as AC-HZS-001; **Observed stdout (verbatim)**: `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected marker is GONE (rm output: "")`; **Exit code**: `1`; **Tree SHA**: `81786284e` + untracked instrument / committed `9dbe40c0a`.
  Fixture note (red-repro.md): the row's escapes derive from the DECLARED
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

- **Cell 1 (pre-repair green)**: denied on the pre-repair tree (measured
  inside the same RED run — the control row passed while the bypass row
  failed, red-repro.md §Defect ② fixture note); recorded at M1's
  instrument-confirmation run.
- **Cell 2 (post-repair green)**: M2 keeps it denied.

### AC-HZS-005 — No-digit escapes render literally, no panic (release-blocking; REQ-HZS-004)

Maps REQ-HZS-004

**Given** the decoder surface `zoneUnescapeAnsiC` (panic contained per row by
the instrument's helper),
**When** the rows `\x`, `\u`, `\U`, `\xZ`, `a\xZb` decode,
**Then** every row returns its input verbatim (the bash literal — backslash
and prefix letter kept, following text surviving once) and no row panics.

Two-cell adoption:

- **RED-now cell** (MEASURED 2026-10-08, red-repro.md §Defect ③): same
  command as AC-HZS-001; **Observed stdout (verbatim)**: three `PANICKED —
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
- **Green path cell**: M2 ③ flips it — the row observes a decision (deny or
  allow per policy for the non-zone target), never a panic.

### AC-HZS-007 — `\u`/`\U` code-point rendering pinned (regression pin; REQ-HZS-003)

Maps REQ-HZS-003

**Given** the decoder surface,
**When** `⊇` decodes,
**Then** the result is the three UTF-8 bytes `e2 a8 87` (the measured bash
row) — the M1 pin row makes the code-point path explicit so M2's `\x` split
cannot regress it. Green-now by design (the current decoder is correct for
`\u`/`\U`); the pin's value is POST-repair, verified again at M3.

- **Cell 1**: green on the pre-repair tree (state pinned by the ground-truth
  measurement, red-repro.md table row 5); the row executes at M1.
- **Cell 2**: green after M2 and at M3 — the split keeps the code-point arm.

### AC-HZS-008 — Family green + windows build (regression guard; REQ-HZS-006)

Maps REQ-HZS-006

**Given** the M2 repair,
**When** `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/hook/` runs,
**Then** the package is green: the t1570 quoting matrix (7 deny + 2 allow),
all zone suites, and the finalized instrument — the ONLY delta vs the M1
baseline is the three defect rows flipping green (countable, no new failure);
and `GOOS=windows go build ./...` exits 0.

- **Cell 1**: baseline recorded at M1 (the package minus the three RED rows
  is green — the RED runs' other tests passed: the solo runs show only the
  three defect failures, red-repro.md).
- **Cell 2**: M3 — full green + windows build, verbatim outputs in §E.2.

## Evidence Ledger

The binding RED observations live in the committed record
`.moai/reports/t1585/red-repro.md` (commit `9dbe40c0a`) — command, verbatim
output, exit code, tree attribution, and the bash ground-truth table per
defect. This ledger adds only what the RED runs did not carry: the M1
instrument-confirmation observations (controls + the new guard-level row) and
the GREEN flip records appended at M2/M3.

### RED confirmation at the committed baseline (appended at M1)

- **Command**: the AC-HZS-001 command, run at `9dbe40c0a`.
- **Expected**: exit 1, the three defect rows failing exactly as
  red-repro.md records them; controls green.
- **Observed**: *(populated at M1)*

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
