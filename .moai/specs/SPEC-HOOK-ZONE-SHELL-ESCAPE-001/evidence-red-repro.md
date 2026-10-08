# t1585 RED reproduction — protected-zone shell gate, 3 defects (canonical TRACKED record)

Canonical tracked copy of the card's RED record (review-gate round 2: evidence
must be readable from a fresh checkout, and `.moai/reports/*` is gitignored).
Provenance: measured 2026-10-08 by lane-19 in the card worktree
`.moai/worktrees/t1585`; the original working record
`.moai/reports/t1585/red-repro.md` (gitignored) is the demoted card-scoped
duplicate. THIS copy carries the 2026-10-09 corrections — listed explicitly so
the original's text is never silently re-read.

## Corrections (2026-10-09 — review-gate rounds 1-2 + plan-audit iteration 1)

1. The `\u` ground-truth row's bytes: the 2026-10-08 record transcribed
   `e2 a8 87` (= U+2A07, a DIFFERENT character). Re-measured with `od`:
   `$'⊇'` → `e2 8a 87` (= U+2287, ⊇). Every quoting artifact now pins
   `e2 8a 87`.
2. The row labeled "\u renders the code point" used the LITERAL character
   `$'⊇'` — no `\u` escape, a pass-through that never reaches the decoder's
   `\u` branch. Both forms are now measured and listed separately.
3. Added the `\U` row: `$'\U00002287'` renders the LITERAL text on this
   host's bash 3.2.57 (`\U` unsupported there).
4. The 2026-10-08 record labeled the ground-truth shell "bash 5.x"; the
   2026-10-09 re-measurement identifies this host's shell as bash 3.2.57
   (both `bash` and `/bin/bash`). ALL seven rows below were re-measured
   2026-10-09 on that shell — the table is anchored to the identified host
   bash, not the earlier label.
5. §Gaps updated: the two mutant-probe controls are now MEASURED (split into
   independent tests at commit `7be9f41b5`, both PASS — pre-repair green).

## Tree attribution

- HEAD at the 2026-10-08 measurement: `81786284e` (= local main tip; the base
  the card branch `WT-zone-gate-defects` was cut from), plus the UNTRACKED
  instrument `internal/hook/protected_zone_shell_repro_test.go` — committed
  baseline-first as `9dbe40c0a` AFTER the measurement (ordering attribution:
  the RED baseline precedes every repair commit).
- Bash: 3.2.57(1)-release, this host (darwin 27) — re-measurement 2026-10-09.

## Ground truth: what real bash does with ANSI-C escapes

Measured with `od -An -tx1` (rows 1-5, 7 on 2026-10-08; ALL rows re-measured
2026-10-09 on bash 3.2.57 — identical output):

| Command | Output bytes | Reading |
|---|---|---|
| `$'a\x00b'X` | `61 58` ("aX") | the ANSI-C part truncates at the embedded NUL; later word parts still append |
| `zone_dir$'\x00/sub'` | `7a 6f 6e 65 5f 64 69 72` ("zone_dir") | whole word = the pre-NUL text |
| `$'a\0b'` | `61` ("a") | octal NUL truncates identically |
| `$'\xec\xa1\x80'` | `ec a1 80` | `\xHH` emits ONE RAW BYTE (no code-point re-encoding) |
| `$'⊇'` | `e2 8a 87` | `\u` renders the code point as UTF-8 (current decoder correct here; supported even on bash 3.2.57) |
| `$'\U00002287'` | `5c 55 30 30 30 30 32 32 38 37` (`\U00002287` literal) | `\U` is NOT supported on bash 3.2.57 — rendered literally; the decoder models the modern set (host-variance residual — on such hosts a literal-named entry is a real bypass class, gate-measured; see spec.md §B, NOT claimed safe) |
| `$'⊇'` (literal char) | `e2 8a 87` | a non-ASCII literal passes through byte for byte (pass-through — NOT the `\u` branch) |

## Defect ① (P1) — NUL truncation absent: protected-path bypass

Claim: `zoneUnescapeAnsiC` keeps a decoded NUL and the text after it, so the
guard judges `zone_dir\x00/sub` while the shell removes `zone_dir`.

Command (as measured 2026-10-08; plan.md §C carries the anchored re-run form):

```
unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run 'TestCheckProtectedZoneShellAnsiCNulTruncationBypass|TestCheckProtectedZoneShellHexRawByteBypass|TestZoneUnescapeAnsiCNoDigitHexStaysLiteral' -count=1 -v
```

Verbatim output (exit code 1) — defect ① excerpt:

```
=== RUN   TestCheckProtectedZoneShellAnsiCNulTruncationBypass
    protected_zone_shell_repro_test.go:118: BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected directory is GONE (rm output: "")
    protected_zone_shell_repro_test.go:120: swept=0
--- FAIL: TestCheckProtectedZoneShellAnsiCNulTruncationBypass (0.02s)
```

The demonstration branch executed the allowed command under real bash in the
fixture root: the protected `zone_dir` (with its marker) was REMOVED. This is a
measured end-to-end bypass, not an inference.

## Defect ② (P2) — `\x` byte vs rune confusion: raw-byte path spelling bypass

Claim: `zoneHexEscape` renders `string(rune(val))`, so each `\xHH` byte ≥ 0x80
re-encodes as a 2-byte UTF-8 sequence; a protected path spelled in raw UTF-8
bytes decodes to text no zone entry matches.

Verbatim output (same run, exit code 1) — defect ② excerpt:

```
=== RUN   TestCheckProtectedZoneShellHexRawByteBypass
    protected_zone_shell_repro_test.go:185: BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected marker is GONE (rm output: "")
    protected_zone_shell_repro_test.go:187: swept=0
--- FAIL: TestCheckProtectedZoneShellHexRawByteBypass (0.02s)
```

Fixture note: the row's `$'\xNN'` escapes are derived from the bytes of the
declared directory name (`존`, measured `ec a1 b4`), so the decoded word and
the created directory are the same bytes by construction. An earlier
instrument draft hardcoded a hex spelling that did not match the declared name
and produced a fixture failure (guard correctly denied on an INVALID manifest —
fail closed); that run is superseded by the valid-manifest allow above.

## Defect ③ (P2) — no-digit hex escape: slice-bounds panic + text mangling

Claim: with no hex digit after `\x`/`\u`/`\U`, `zoneHexEscape` returns
`v[i:i+2]` with `i` at the LAST index (panic when the escape ends the string);
after a non-digit it drops the backslash and the loop re-emits that character.

Verbatim output (same run, exit code 1) — defect ③ excerpt:

```
=== RUN   TestZoneUnescapeAnsiCNoDigitHexStaysLiteral
    protected_zone_shell_repro_test.go:232: "\\x": PANICKED — slice bounds out of range on the no-digit escape
    protected_zone_shell_repro_test.go:232: "\\u": PANICKED — slice bounds out of range on the no-digit escape
    protected_zone_shell_repro_test.go:232: "\\U": PANICKED — slice bounds out of range on the no-digit escape
    protected_zone_shell_repro_test.go:236: "\\xZ": decoded "xZZ", want the bash literal "\\xZ"
    protected_zone_shell_repro_test.go:236: "a\\xZb": decoded "axZZb", want the bash literal "a\\xZb"
    protected_zone_shell_repro_test.go:242: swept=5
--- FAIL: TestZoneUnescapeAnsiCNoDigitHexStaysLiteral (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.889s
```

Guard-level consequence of ③ (not executed in the instrument to keep the panic
from aborting the binary): any mutating command carrying `$'\x'` panics inside
the walk — the analysis crashes instead of judging.

## Fix direction (plan input — validated and ADOPTED; adjudication table: plan.md §A.2)

1. ① Honor NUL as the ANSI-C part terminator: the part's decoded contribution
   ends at the first NUL byte (hex `\x00` or octal origin); later word parts
   still append (bash-measured). Implemented at the `zoneWordText` part level.
2. ② `\xHH` renders ONE RAW BYTE (`byte(val)`); `\u`/`\U` keep code-point
   rendering (`string(rune(val))`). Split `zoneHexEscape` accordingly.
3. ③ No-digit escape returns the literal backslash + prefix letter, bounded —
   no panic, following text survives once.

Constraints: t1570's 3-layer decoder split (bare / double-quoted / ANSI-C)
stays intact; the walker's policy layers are out of scope; zsh-side NUL
semantics are a noted residual (the Bash tool runs bash).

## Gaps (updated 2026-10-09)

- RESOLVED: the two mutant-probe controls (outside-zone NUL truncation;
  direct-spelling deny) are now independent tests — MEASURED at `7be9f41b5`,
  both PASS (pre-repair green). New controls owed at M1: the part-level row,
  the octal-origin command row, and the non-ASCII outside-zone allow controls
  (acceptance.md AC-HZS-009/010/011).
- The full `internal/hook` family was not run at plan time (lane-local
  verification discipline; run phase owns the family re-run).
- Windows behavior unmeasured (rows skip on windows; the guard analysis is
  platform-independent, the demonstrations are POSIX).
