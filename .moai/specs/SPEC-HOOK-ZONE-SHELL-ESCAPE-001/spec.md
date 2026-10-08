---
id: SPEC-HOOK-ZONE-SHELL-ESCAPE-001
title: "Protected-zone ANSI-C shell decoding judges the bytes bash executes — NUL part-terminator, raw-byte \\x, bounded no-digit escapes"
version: "0.1.0"
status: draft
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-SELF-IMPROVE-PROTECTED-ZONE-001, SPEC-HOOK-ZONE-BACKSLASH-001]
tags: "hook, security, protected-zone, ansi-c, shell-quoting, nul-truncation, raw-byte, panic, t1585"
---

# SPEC-HOOK-ZONE-SHELL-ESCAPE-001 — Protected-zone ANSI-C decoding judges the bytes bash executes

## A. History

| Date | Version | Change |
|------|---------|--------|
| 2026-10-09 | 0.1.0 | Initial draft. Authored by manager-spec for card t1585 (t1570 follow-up), from the lane-19 RED reproduction measured 2026-10-08 (`.moai/reports/t1585/red-repro.md`; instrument committed baseline-first as `9dbe40c0a` on `WT-zone-gate-defects`). Three measured defects in the ANSI-C (`$'...'`) half of the shell decoder: a P1 protected-path bypass and two P2 decoder defects. |

## B. Problem Statement

Every Bash decision funnels through `zoneWordText`
(`internal/hook/protected_zone_shell.go:244`), which assembles a word from its
parts and decodes each quoting layer: bare literals via `zoneUnescapeLit`
(`:85`), double-quoted parts via `zoneUnescapeDbl` (`:106`), and ANSI-C
(`$'...'`) parts via `zoneUnescapeAnsiC` (`:139`, added by card t1570). The
guard decides on the decoded text, so the decoded text must equal the argument
string bash actually places in the executed command — anything else splits the
judged path from the acted-on path.

Three defects in the ANSI-C half are MEASURED (not inferred): the RED
reproduction (`.moai/reports/t1585/red-repro.md`, measured 2026-10-08 in this
card worktree, exit code 1 on all three defect tests) demonstrates each. The
bash ground truth was measured first with real bash:

| Command | Output bytes | Reading |
|---|---|---|
| `$'a\x00b'X` | `61 58` ("aX") | the ANSI-C part truncates at the embedded NUL; later word parts still append |
| `zone_dir$'\x00/sub'` | `7a 6f 6e 65 5f 64 69 72` ("zone_dir") | whole word = the pre-NUL text |
| `$'a\0b'` | `61` ("a") | octal NUL truncates identically |
| `$'\xec\xa1\x80'` | `ec a1 80` | `\xHH` emits ONE RAW BYTE (no code-point re-encoding) |
| `$'⊇'` | `e2 a8 87` | `\u` renders the code point as UTF-8 (current decoder correct here) |

### Defect ① (P1) — decoded NUL kept: protected-path bypass

`zoneUnescapeAnsiC` renders `\x00` and octal NUL as a real NUL byte and keeps
decoding past it (`zoneHexEscape` `:221` returns `string(rune(0))`; the loop
continues), so the guard judges `zone_dir\x00/sub` — no zone entry matches —
while bash truncates the ANSI-C part at the NUL and the executed command
targets `zone_dir` itself. Measured end-to-end: decision `allow`, the
demonstration branch ran the allowed command under real bash, and the
protected `zone_dir` (with its marker) was REMOVED.

### Defect ② (P2) — `\x` byte vs code-point confusion: raw-byte path spelling bypass

`zoneHexEscape` renders ALL THREE hex prefixes with `string(val)` (`:221`) —
a code-point re-encode. For `\xHH` with HH ≥ 0x80, bash emits ONE RAW BYTE;
the decoder emits a 2-byte UTF-8 sequence, so a protected path spelled in raw
UTF-8 bytes decodes to text no zone entry matches. Measured: decision `allow`
and the protected marker removed under real bash.

### Defect ③ (P2) — no-digit hex escape: slice-bounds panic + text mangling

With no hex digit after `\x`/`\u`/`\U`, `zoneHexEscape` returns
`v[*i : *i+2]` (`:217`-`:218`) with `*i` at the prefix letter: when the escape
ends the string this is a slice-bounds PANIC; after a non-digit it returns the
letter and the following character with the backslash dropped, and the loop
then re-emits that character — `\xZ` decodes to `xZZ` instead of the bash
literal `\xZ`. All measured (the instrument contains the panic via a
recover-per-row helper). Guard-level consequence (code-derived, not executed
in the instrument: the panic is contained per row to keep the binary alive):
the walk path `zoneWordText → zoneUnescapeAnsiC → zoneHexEscape` carries no
`recover()` (`pre_tool.go:789` → `checkProtectedZoneShell:1125` → walker), so
any mutating command carrying `$'\x'` crashes the analysis instead of being
judged.

## C. Requirements (GEARS)

### REQ-HZS-001 — ANSI-C NUL truncation: the guard judges the word bash executes (Event-driven)

**When** a Bash command word carries an ANSI-C (`$'...'`) part whose decoded
value contains a NUL byte — from `\x00` or from an octal escape — the guard
shall judge that word with the ANSI-C part's decoded contribution ending at
the first NUL byte, later word parts still appending, so the judged text
equals the argument string bash passes to the executed command; it shall not
judge any text past the NUL. The measured shapes pin the behavior:
`zone_dir$'\x00/sub'` is judged as `zone_dir` (deny for a zone-covered
`zone_dir`), and `$'a\x00b'X` is judged as `aX`.

### REQ-HZS-002 — `\xHH` renders one raw byte (Ubiquitous)

The ANSI-C decoder shall render every `\xHH` escape as exactly one byte with
the value 0xHH — the raw-byte form bash emits. The decoded word text may then
be any byte sequence (non-UTF-8 included), and zone matching shall operate on
those same bytes the shell places in the executed argument.

### REQ-HZS-003 — `\u`/`\U` code-point rendering preserved (Unwanted)

The repair shall not change the rendering of `\u`/`\U` escapes: the code point
shall keep rendering as UTF-8 (measured row: `⊇` → `e2 a8 87`), and the
t1570 quoting matrix's ANSI-C rows (defined escape `\x2e`, escape-free text)
keep their current decisions — the `\x` raw-byte change is behavior-visible
only for values ≥ 0x80.

### REQ-HZS-004 — No-digit hex escape renders literally, never panics (Event-driven)

**When** an ANSI-C `\x`, `\u`, or `\U` prefix is followed by no hex digit —
including when the escape ends the word text — the decoder shall render the
literal backslash and the prefix letter (bash's rendering), the text following
the escape shall survive once (no doubling), and the decode shall not panic.
Measured target: `\x`, `\u`, `\U` stay `\x`, `\u`, `\U`; `\xZ` stays `\xZ`;
`a\xZb` stays `a\xZb`.

### REQ-HZS-005 — The guard's walk completes on any command text (Event-driven)

**When** a Bash call's text carries any ANSI-C escape sequence, including a
malformed or incomplete one, the guard's analysis shall run to completion and
return a decision — no escape shape shall abort the walk with a panic. A
decoder-facing defect in the escape text is judged (deny or allow per the zone
policy), never crashed on.

### REQ-HZS-006 — No behavior change outside the three defect shapes (Unwanted)

The repair shall not change the decode or decision behavior of any word
carrying none of the three defect shapes: every existing `internal/hook` test
stays green — the t1570 quoting matrix (`testZoneShellQuoting`,
`protected_zone_guard_test.go:1238`: 7 deny rows + 2 allow controls), the zone
suites, and the full package family (the family re-run is run-phase scope; the
plan-phase baseline is green except the three RED defect rows).

## D. Constraints

- **The 3-layer decoder split stays intact.** Bare (`zoneUnescapeLit`) /
  double-quoted (`zoneUnescapeDbl`) / ANSI-C (`zoneUnescapeAnsiC`) remain
  separate decoders with their separate escape semantics; the repair touches
  only the ANSI-C layer and the part assembly that consumes it.
- **The walker's policy layers are out of scope**: mutation verbs, the
  possible-directory-set walk, redirect judging, and every
  `protected_zone_shell.go` function outside the decoder family
  (`zoneUnescapeAnsiC`, `zoneHexEscape`, `zoneOctalEscape`, `zoneWordText`)
  are unchanged.
- **zsh NUL semantics are a recorded residual** (explicitly out of scope): the
  Bash tool runs bash; what a zsh ANSI-C string does with an embedded NUL is
  unmeasured here and not repaired here.
- **The RED instrument is the baseline-first evidence carrier**: committed as
  `9dbe40c0a` (`test(t1585)`, branch `WT-zone-gate-defects`) BEFORE any repair
  commit, per the ordering-attribution rule
  (`.claude/rules/moai/core/verification-claim-integrity.md` §2.3). Run phase
  FINALIZES it — adds the guard-level no-crash row and the `\u` code-point pin
  row, keeps the real-bash demonstration branches — and never weakens a row.
- Tests: `t.TempDir()` for every temporary directory; POSIX-gated rows skip
  gracefully (`t.Skip`) on windows (the guard analysis is platform-independent;
  the demonstrations exec real bash); table-driven where the surface is pure.
  Never `t.Setenv("OTEL_*", ...)` in parallel tests.
- Per repo discipline: run the affected package only —
  `go test -count=1 ./internal/hook/` — not the full suite locally; CI runs
  the full matrix.
- TRUST 5 throughout: Tested (two-cell ACs, ≥85% package coverage preserved),
  Readable, Unified (gofmt), Secured (this repair IS a security fix),
  Trackable (Conventional Commits, card id in every commit).
- Go code, comments, godoc: English. Error wrapping `fmt.Errorf("...: %w", err)`.

## E. Success Criteria

1. Two-cell adoption (`.claude/rules/moai/development/verification-completeness.md`
   §2) for AC-HZS-001/003/005: the RED cells are ALREADY measured
   (`.moai/reports/t1585/red-repro.md`, verbatim outputs + exit 1, tree
   `81786284e` + the instrument, re-runnable at the committed baseline
   `9dbe40c0a`), and the M2 repair flips them GREEN — both sides observed on
   this tree, no carry-over.
2. The mutant-probe controls (AC-HZS-002 outside-zone truncation stays
   allowed; AC-HZS-004 direct-spelling deny stays denied) are green on BOTH
   sides of the repair — a fix that blanket-denies NUL-bearing or non-ASCII
   commands fails them.
3. The run-phase-finalized instrument carries the guard-level no-crash row
   (AC-HZS-006) and the `\u` code-point pin row (AC-HZS-007), both GREEN
   after M2.
4. The affected package family is green: `go test -count=1 ./internal/hook/`
   (only the three RED defect rows flip; countable delta, no new failure), and
   `GOOS=windows go build ./...` exits 0.

## F. Scope Boundary

This SPEC repairs the three measured ANSI-C decoder defects and nothing else.

### Out of Scope — walker policy layers

- Every `protected_zone_shell.go` function outside the decoder family
  (`zoneUnescapeAnsiC`, `zoneHexEscape`, `zoneOctalEscape`, `zoneWordText`):
  mutation verbs, the possible-directory-set walk, redirect targets, function-
  call unrolling — unchanged.
- `checkProtectedZoneShell` entry, identity gating, and deny-reason rendering —
  unchanged.

### Out of Scope — other decoder layers

- `zoneUnescapeLit` (bare words) and `zoneUnescapeDbl` (double-quoted parts):
  the three defects are measured only in the ANSI-C path; the other two layers
  keep their t1570 semantics untouched.

### Out of Scope — zsh and other shells

- zsh NUL and raw-byte semantics: unmeasured, recorded as a residual
  (constraint §D). The Bash tool runs bash; no zsh behavior is repaired or
  asserted here.

### Out of Scope — sibling guard surfaces

- `protected_zone_path.go`, `pre_tool.go` file-tool branches, and every other
  protected-zone surface: unchanged and unrepaired by this SPEC.
