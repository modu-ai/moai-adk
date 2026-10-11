---
id: SPEC-HOOK-ZONE-SHELL-ESCAPE-001
title: "Plan — ANSI-C shell decoder repair (NUL part-terminator, raw-byte \\x, bounded no-digit escapes)"
version: "0.1.4"
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
---

# SPEC-HOOK-ZONE-SHELL-ESCAPE-001 — Implementation Plan

## §A Context

- Card t1585 (Class C: plan → run → sync), branch `WT-zone-gate-defects`,
  worktree `.moai/worktrees/t1585`.
- Baseline HEAD: `9dbe40c0a` — the baseline-first commit
  (`test(t1585): RED reproduction for protected-zone shell escape defects`)
  carrying the instrument `internal/hook/protected_zone_shell_repro_test.go`
  (its header comment is the COMMITTED carrier of the bash ground truth), on
  top of `81786284e` (= local main tip). The evidence record
  `.moai/reports/t1585/red-repro.md` is MACHINE-LOCAL (`.moai/reports/*` is
  gitignored — verified `git cat-file -e 9dbe40c0a:...` exit 128); it is the
  working record, never a citation target for fresh checkouts. The CANONICAL
  record is the TRACKED `evidence-red-repro.md` in the SPEC directory
  (round-2 repair). Ordering attribution satisfied: the RED baseline
  precedes every repair commit.
- Development mode: `tdd` (quality.yaml `constitution.development_mode`) —
  RED is already in place; run phase is GREEN-first on the M1-final
  baseline's six RED rows (the three measured defects + the M1-authored
  three).
- SPEC artifacts: `.moai/specs/SPEC-HOOK-ZONE-SHELL-ESCAPE-001/{spec,plan,acceptance,progress}.md`.
- Evidence: `evidence-red-repro.md` in the SPEC directory (canonical TRACKED
  record, committed — corrected and fully re-measured 2026-10-09 on bash
  3.2.57) — bash ground-truth table, per-defect verbatim test output
  (exit 1), tree attribution, fix directions, gaps. The gitignored
  `.moai/reports/t1585/red-repro.md` is the demoted card-scoped duplicate;
  the instrument header comment (`9dbe40c0a`) is the committed in-code
  carrier.

### §A.1 The surface (read, verified against the RED record)

- `zoneUnescapeAnsiC` (`protected_zone_shell.go:139`) — ANSI-C escape decode;
  defect ① lives in what it KEEPS (NUL + trailing text), defects ②③ in
  `zoneHexEscape`.
- `zoneHexEscape` (`:193`) — one function serves all three hex prefixes
  (`\x` maxDigits 2, `\u` 4, `\U` 8) and returns `string(val)` (`:221`)
  unconditionally; the no-digit arm returns `v[*i : *i+2]` (`:217`-`:218`).
- `zoneOctalEscape` (`:226`) — octal decode, renders `byte(val)`; a `\0`
  origin NUL reaches the word text the same way a `\x00` does.
- `zoneWordText` (`:244`) — part assembly; the ANSI-C branch (`:258`-`:259`)
  is where a decoded part joins the word — the part-level NUL terminator
  lands here.

### §A.2 Fix-direction validation (plan input adjudication)

The dispatch supplied three fix directions (evidence-red-repro.md §Fix
direction) as
plan INPUT. Each was validated against the code; all three are ADOPTED with
no deviation:

| # | Direction (input) | Code reading | Verdict |
|---|---|---|---|
| ① | Honor NUL as the ANSI-C part terminator at the `zoneWordText` part level; later parts still append | Ground-truth row `$'a\x00b'X` → "aX" pins the truncation to the PART (the `X` is a separate Lit part that must survive) — a word-level truncation would wrongly drop it. Truncating the decoded ANSI-C part's contribution at its first NUL byte covers both origins (`\x00` via `zoneHexEscape` val 0, `\0` via `zoneOctalEscape` byte 0). Equivalent placement inside `zoneUnescapeAnsiC` (truncate its return) is acceptable — `zoneWordText:259` is the sole caller — run phase picks the shape | ADOPTED |
| ② | `\xHH` renders ONE RAW BYTE (`byte(val)`); `\u`/`\U` keep code-point rendering; split `zoneHexEscape` accordingly | `:221` `return string(val)` is the single return serving all three prefixes — the split is real and minimal. Behavior-visible only for val ≥ 0x80 (`byte(c)` == `string(rune(c))` below 0x80), so the t1570 matrix row `\x2e` (val 0x2e) is untouched. Raw-byte output may be non-UTF-8; zone matching is byte-wise, which is exactly what must match | ADOPTED |
| ③ | No-digit escape returns the literal backslash + prefix letter, bounded — no panic, following text survives once | `:217`-`:218` returns `v[*i : *i+2]`: `*i` is the letter index, so end-of-string panics and a non-digit follower drops the backslash, after which the loop's `i++` re-emits the follower (doubling — measured `xZZ`). Rendering `\` + letter and leaving `*i` on the letter (the loop advances past it) yields the bash literal and survives the follower exactly once | ADOPTED |

### §A.3 PRESERVE list (scope discipline)

- `internal/hook/protected_zone_shell.go` — every function OUTSIDE the decoder
  family (`zoneUnescapeLit`, `zoneUnescapeDbl`, the walker, redirect judging,
  `checkProtectedZoneShell`).
- `internal/hook/protected_zone_guard_test.go` — the t1570 quoting matrix
  (`testZoneShellQuoting`) unchanged.
- `internal/hook/protected_zone_path.go`, `internal/hook/pre_tool.go` — untouched.
- The instrument's existing rows and real-bash demonstration branches — run
  phase may only ADD rows (guard-level no-crash, `\u` pin), never weaken one.

## §B Known Issues (auto-injection, relevant subset)

- **B1 Cross-platform**: the demonstration rows skip on windows (`t.Skip`,
  POSIX bash); the decoder itself is platform-independent — `GOOS=windows go
  build ./...` must stay exit 0 (AC-HZS-008).
- **B6 spec-lint headings**: the Out of Scope section uses `### Out of Scope —
  <topic>` H3 sub-headings with `-` bullets (OutOfScopeRule).
- **B8 Working-tree hygiene**: stage by explicit pathspec; the instrument file
  is already committed — no stray untracked files expected.
- **B10 Scope discipline**: touch ONLY the decoder family + the instrument
  file (§A.3 PRESERVE list).
- **B11 Blocker protocol**: run phase returns structured blocker reports; no
  AskUserQuestion.

## §C Pre-flight (run-phase, before M1)

```bash
git branch --show-current && git rev-parse --short HEAD   # WT-zone-gate-defects @ 9dbe40c0a (or a descendant carrying only plan docs)
go build ./... && GOOS=windows go build ./...              # both exit 0
unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -run '^(TestCheckProtectedZoneShellAnsiCNulTruncationBypass|TestCheckProtectedZoneShellHexRawByteBypass|TestZoneUnescapeAnsiCNoDigitHexStaysLiteral)$' -count=1   # exit 1 — RED re-confirmed at the committed baseline
```

## §D Constraints

- As spec.md §D verbatim: 3-layer split intact; policy layers out of scope;
  zsh residual; instrument finalize-only; lane-local verification
  (`go test -count=1 ./internal/hook/`); TRUST 5; English comments.
- The repair must keep every existing `internal/hook` test green — the ONLY
  expected flips are the M1-final baseline's six RED rows going green
  (①②③ + the M1-authored 006/009/010).
- **The `\u`/`\U` host-variance bypass class is NOT this card's repair
  scope**: on a pre-4.2 bash — this host renders BOTH `\u` and `\U` literally
  (decisive od-proven re-measurement) —
  the literal-named entry (symlink included)
  reaches the zone unjudged — documented residual (spec.md §B/§F) for a
  follow-up card, with candidate closures noted there WITHOUT decision. Run
  phase may not repair beyond the RED rows.
- No blanket-deny mutant: REQ-HZS-001's fix must truncate, not reject — the
  outside-zone control (AC-HZS-002) fails a mutant that denies every
  NUL-bearing command.

## §E Self-Verification (run-phase deliverables)

Per the attribution discipline (manager-develop-prompt-template §E): every
item names command + verbatim output + tree SHA.

- **E1** AC binary PASS/FAIL matrix (AC-HZS-001..011) with the verbatim test
  output per row.
- **E2** `go build ./...` and `GOOS=windows go build ./...` — both exit 0.
- **E3** `go test -cover ./internal/hook/` — package coverage ≥ the
  pre-change figure (85%+ maintained).
- **E8** RED failure output verbatim — already carried by the tracked record
  `evidence-red-repro.md` (measured 2026-10-08); M1 re-confirms it
  at the committed baseline `9dbe40c0a` and records the re-run in progress.md
  §E.2 (the pre-GREEN evidence citation).

## §F Milestones

Ordered RED-first (the RED rows gate every repair), then by
decision-reversibility (the decoder shape decisions in M2 lead the mechanical
re-runs of M3).

### M1 — Instrument finalization + RED baseline re-confirmation

- Re-run the RED rows at the committed baseline; record the verbatim
  output in progress.md §E.2 (pre-GREEN evidence).
- ADD instrument rows: (a) guard-level no-crash — a Bash call carrying
  `$'\x'` judged with the panic contained in the test helper
  (`hzsDecodeAnsiC` pattern extended to the guard call): RED pre-fix (the
  walk panics), decision post-fix; (b) `\u`/`\U` code-point pin — the
  ESCAPE TEXTS — numerically: byte 0x5C followed by the ASCII characters
  u, 2, 2, 8, 7 (six bytes; in Go test source the literal is written
  `\\u2287`, where the doubled backslash is Go source syntax denoting the
  single 0x5C byte at runtime), and
  `\U00002287` (the decoder receives the inner text, so the rows name the
  escape TEXT, not the glyph) — the current decoder returns bytes `e2 8a 87`
  (the `\u` branch, maxDigits 4), and a maxDigits 4→2 mutant returns a
  different byte sequence and FAILS the row (pins fix ②'s split from
  regressing the code-point arm); (c) part-level NUL row — `zoneWordText`
  over `$'a\x00b'X` yields `aX` (RED pre-fix: the current decode keeps the
  NUL, `a\x00bX`; fails a word-level-truncation mutant that satisfies the
  command rows); (d) octal-origin command row — `rm -r zone_dir$'\0/sub'`
  denied (RED pre-fix: allow — the same judged-text path as the hex case
  through `zoneOctalEscape`'s `byte(0)`); (e) non-ASCII outside-zone allow
  controls — a non-ASCII-named file outside the zone stays ALLOWED in both
  the literal and the raw-byte spelling (fails a deny-all-non-ASCII mutant;
  plan-audit D5).
- No repair in this milestone.

### M2 — Decoder repair (the three defect fixes)

- ① NUL part-terminator: the ANSI-C part's decoded contribution ends at its
  first NUL byte — implement at the `zoneWordText` ANSI-C branch (or the
  decoder return; §A.2 ① — run phase's shape call), later parts still append.
- ② Split the hex rendering: `\x` → `byte(val)` (one raw byte); `\u`/`\U` →
  `string(rune(val))` (code point as UTF-8). Keep the shared digit-scanning
  loop; the split is in the RENDER, not the scan.
- ③ No-digit arm: render `\` + prefix letter, advance over the letter only,
  never index past the end — all three prefixes covered.
- Flip the M1-final baseline's six RED rows green — ①②③ (AC-HZS-001/003/005)
  + the M1-authored guard no-crash (006), part-level (009), octal-origin
  (010); the controls (002/004/011) and the code-point pin (007) stay green.

### M3 — Family re-run + regression confirmation

- `go test -count=1 ./internal/hook/` — full package green (countable delta:
  the M1-final baseline's six RED rows flip — 001/003/005 + 006/009/010;
  the controls 002/004/011 and the pin 007 stay green; no new failure).
- `GOOS=windows go build ./...` exit 0; gofmt clean; coverage ≥ pre-change.
- progress.md §E.2 evidence + §E.3 audit-ready signal.

## §G Anti-Patterns

- **Blanket-deny mutant**: rejecting every NUL-bearing or non-ASCII command
  instead of truncating/decoding — AC-HZS-002/004 exist to fail exactly this.
- **Regressing `\u`**: "simplifying" to one render for all three prefixes
  re-breaks the code-point path — REQ-HZS-003 and the M1 pin row guard it.
- **Weakening the instrument**: deleting a demonstration branch or a control
  row to go green — never; rows only accumulate.
- **Scope creep into the policy layers**: the walk, verbs, and redirects are
  not defective here and are not touched.
- **Full-suite local runs**: lane-local discipline — affected package only.

## §H Cross-References

- SPEC-HOOK-ZONE-BACKSLASH-001 (t1570) — the ANSI-C decoder's origin SPEC;
  its quoting matrix is the family regression floor.
- SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 — the protected-zone guard root SPEC.
- `evidence-red-repro.md` (in this SPEC directory) — the canonical TRACKED
  RED record (corrected + fully re-measured 2026-10-09 on bash 3.2.57). The
  gitignored `.moai/reports/t1585/red-repro.md` is the demoted card-scoped
  duplicate; the instrument header (`9dbe40c0a`) is the committed in-code
  carrier.
- `.claude/rules/moai/development/verification-completeness.md` §2 — the
  two-cell adoption rule every AC follows.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 — the
  ordering attribution the baseline-first commit satisfies.
