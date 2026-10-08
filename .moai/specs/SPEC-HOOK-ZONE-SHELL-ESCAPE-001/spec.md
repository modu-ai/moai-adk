---
id: SPEC-HOOK-ZONE-SHELL-ESCAPE-001
title: "Protected-zone ANSI-C shell decoding judges the bytes bash executes — NUL part-terminator, raw-byte \\x, bounded no-digit escapes"
version: "0.1.4"
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
| 2026-10-09 | 0.1.1 | Review-gate rounds 1-2 + plan-audit iter-1 repair: `\u` pin bytes corrected `e2 a8 87` → `e2 8a 87` (the old value is U+2A07, a different character), pin input restated as the escape texts (byte 0x5C + `u2287`; `\U00002287`), host-bash 3.2.57 footnote + full ground-truth re-measurement (all seven rows, `od`), evidence custody corrected and the canonical record TRACKED as `evidence-red-repro.md` (gitignored duplicate demoted), controls cited MEASURED at `7be9f41b5`, AC-HZS-006 §2.1 conditional-demotion sentence, new AC-HZS-009 (part-level shape) / AC-HZS-010 (octal origin) / AC-HZS-011 (non-ASCII outside-zone control), AC-HZS-008 cell 1 re-scoped to the recorded run. |
| 2026-10-09 | 0.1.2 | Gate round 3: the prior deny-safe claim about `\U` on old bash DELETED (false — a literal-named entry, symlink included, is live and reaches the zone unjudged while the guard judges the decoded path; gate-measured: allow + protected file deleted); replaced with a support-boundary + residual statement (decoder models bash ≥ 4.2 semantics; version-variance family alongside zsh; follow-up-card material; candidate closures noted without decision). Host split pinned: `\u` renders / `\U` literal on `/bin/bash` 3.2.57 (three pinned re-measurements). |
| 2026-10-09 | 0.1.3 | Gate round 4 (verbatim gate-shaped texts): ground-truth `\u` row labeled "input is the escape TEXT, bash-measured"; §B host note replaced verbatim (SUPPORT BOUNDARY + RESIDUAL RISK — the literal-named-entry bypass measured twice; fail-closed alternative recorded as an undecided run-phase design option); the round-3 "accepted as a documented residual" acceptance claim and the "sound over-approximation" candidate removed; AC-HZS-007 When/Then carries the mutant-fail clause (maxDigits 4→2 must FAIL); plan M1(b) aligned. |
| 2026-10-09 | 0.1.4 | Gate round 5 + glyph-poisoning root cause: the morning probe for the `\u` row was GLYPH-POISONED (parameter-transport decode — poisoned script bytes od-verified), so "\u renders on 3.2.57" was false; decisive od-proven re-measurement (printf-assembled escape text): NEITHER `\u` NOR `\U` is supported on bash 3.2.57 — BOTH escape texts render literally; modern-bash (≥4.2) expansion recorded as documented semantics, not host-measured. The residual now covers BOTH escape families. Pin inputs specified NUMERICALLY (byte 0x5C + `u2287`; `\U00002287`) — the doubled-backslash transport artifact removed (it reaches no `\u` branch, gate-measured). |

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
reproduction (the tracked record `evidence-red-repro.md` in this directory,
measured 2026-10-08 in this card worktree, exit code 1 on all three defect
tests) demonstrates each. The bash ground truth — fully re-measured
2026-10-09 on this host's bash 3.2.57 (`od`, all seven rows):

| Command | Output bytes | Reading |
|---|---|---|
| `$'a\x00b'X` | `61 58` ("aX") | the ANSI-C part truncates at the embedded NUL; later word parts still append |
| `zone_dir$'\x00/sub'` | `7a 6f 6e 65 5f 64 69 72` ("zone_dir") | whole word = the pre-NUL text |
| `$'a\0b'` | `61` ("a") | octal NUL truncates identically |
| `$'\xec\xa1\x80'` | `ec a1 80` | `\xHH` emits ONE RAW BYTE (no code-point re-encoding) |
| escape text: byte 0x5C + `u2287` (ANSI-C quoting) | `5c 75 32 32 38 37` (literal) | `\u` is NOT supported on this host's bash 3.2.57 — rendered literally (decisive od-proven re-measurement, superseding the glyph-poisoned probe); modern bash (≥4.2) documents `e2 8a 87`; the Go decoder renders `e2 8a 87` |
| `$'⊇'` (literal char) | `e2 8a 87` | a non-ASCII literal passes through byte for byte |

Host-bash note (decisive re-measurement 2026-10-09: printf-assembled escape
text, script bytes od-verified — immune to the glyph substitution that
poisoned the earlier probes; `bash` and `/bin/bash` identical): NEITHER
`\u` NOR `\U` is supported on this host's bash 3.2.57 — BOTH escape texts
render literally (the byte-0x5C+`u2287` probe outputs `5c 75 32 32 38 37`;
the `\U00002287` probe outputs `5c 55 30 30 30 30 32 32 38 37`; the poisoned
earlier probe's `e2 8a 87` was the glyph passing through). Modern bash
(≥4.2) documents expansion to the code point's UTF-8 `e2 8a 87` —
documented semantics, not host-measured (no ≥4.2 bash on this host). The
decoder models the modern `\u`/`\U` set. SUPPORT BOUNDARY + RESIDUAL RISK —
NOT deny-safe: where the executing bash renders the escape texts literally,
the guard judges the decoded code-point path while the shell touches the
LITERAL escape-text name, and a filesystem entry with that literal name
(e.g. a symlink into the protected zone) is then reachable unjudged — the
review gate measured exactly this bypass on this host (guard `allow`,
protected file deleted by real bash, twice). This divergence is a documented
residual of modeling bash ≥4.2 ANSI-C semantics; a fail-closed design
alternative (deny on any `\u`/`\U` escape) is recorded as a run-phase design
option, not decided here.

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
shall keep rendering as UTF-8 — decoder-pinned rows: the escape texts (byte
0x5C followed by `u2287`, and `\U00002287`) both decode to `e2 8a 87` — the
Go output string(rune(0x2287)) and the modern-bash (≥4.2) documented
expansion. Host-bash note in §B: this host's bash 3.2.57 renders BOTH escape
texts literally, so the decoder's output diverges from that shell's argument
for BOTH families — a documented residual, not claimed safe — and the
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
- **Host `\u`/`\U` version variance is a recorded residual** (explicitly out
  of scope, §B/§F): on a bash lacking `\u`/`\U` (this host: BOTH render
  literally) the decoder's modern-set rendering
  diverges from the executing shell's argument, and a literal-named entry —
  symlink included — is a real bypass class (gate-measured). Not repaired by
  this card — follow-up-card material.
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
   §2): AC-HZS-001/003/005's RED cells are ALREADY measured (the tracked
   record `evidence-red-repro.md` in this directory — verbatim outputs +
   exit 1, tree `81786284e` + the instrument, re-runnable at the committed
   baseline `9dbe40c0a`), and the M2 repair flips them GREEN. AC-HZS-009/010
   carry MEASURED bash-side ground truth with their decoder rows authored at
   M1 under the §2.1 conditional-demotion sentence.
2. The mutant-probe controls — AC-HZS-002 (outside-zone NUL truncation stays
   allowed), AC-HZS-004 (direct-spelling deny stays denied), AC-HZS-011
   (non-ASCII outside-zone stays allowed, literal AND raw-byte spellings) —
   are green on BOTH sides of the repair: a fix that blanket-denies
   NUL-bearing or non-ASCII commands fails them.
3. The run-phase-finalized instrument carries the M1 rows — the guard-level
   no-crash row (AC-HZS-006), the `\u`/`\U` code-point pin rows (AC-HZS-007,
   escape texts: byte 0x5C + `u2287`, and `\U00002287`), the part-level NUL row (AC-HZS-009:
   `$'a\x00b'X` → `aX`), the octal-origin command row (AC-HZS-010), and the
   non-ASCII outside-zone allow controls (AC-HZS-011) — all GREEN after M2
   (AC-HZS-006/009/010 under their §2.1 conditionals).
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

### Out of Scope — host shell-version variance (`\u`/`\U`)

- On hosts whose bash lacks `\u`/`\U` (measured: 3.2.57 renders BOTH
  literally — decisive od-proven re-measurement, superseding the
  glyph-poisoned earlier probe),
  the modern-set decode diverges from the executed argument and a
  literal-named entry — symlink included — reaches the zone unjudged
  (gate-measured bypass class, review-gate round 3). Documented residual,
  explicitly NOT claimed safe; repair is follow-up-card material.
- A fail-closed design alternative (deny on any `\u`/`\U` escape) is
  recorded as a run-phase design option, not decided here.

### Out of Scope — sibling guard surfaces

- `protected_zone_path.go`, `pre_tool.go` file-tool branches, and every other
  protected-zone surface: unchanged and unrepaired by this SPEC.
