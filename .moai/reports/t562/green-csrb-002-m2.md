# AC-CSRB-002 GREEN + M2 production diff record (t562 M2)

- Tree: `b78d2e425` (post STEP-0 rename) → the M2 commit (see session report for the SHA).
- Command: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntry_(SeparatorConversion|ClassifiesDeclaredFormBeforeConversion|HomeRelativeStatTargetStaysNative|EligibilityGatingPins)|TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard' -count=1 -timeout 1800s -v`
- Exit code: `0` (PASS)
- Verbatim output: `green-csrb-002-m2.log` (same directory).

## The RED→GREEN flip

`TestJudgeCodexSkillEntry_SeparatorConversion`: FAIL at M1 (recorder saw the declared form) →
**PASS at M2** (recorder sees `fromConfigPath(declared, configPathSeparator)`). The guards stayed
PASS: ClassifiesDeclaredFormBeforeConversion, HomeRelativeStatTargetStaysNative,
EligibilityGatingPins (3 arms), ShapeReadbackBaselineGuard.

## Production diff — exactly the two one-line conversions (spec.md §D.2)

1. `internal/cli/codex_skills_prune.go`, `judgeCodexSkillEntry`, `codexPathAbsolute` branch:
   `statPath = fromConfigPath(e.Path, configPathSeparator)` (was `statPath = e.Path`).
2. `internal/cli/doctor_codex.go`, `codexStaleSkillFinding`, `codexPathAbsolute` branch:
   the same one-line shape. Home-relative branches untouched in both readers.

## AC-CSRB-007 re-anchored attribution — zero osStatFn tokens added to doctor_codex.go

The file now CARRIES t563's `osStatFn` seam via the absorb (provenance `c72dc1baf`; two call
sites, `:459` and `:857`, both absorbed). This card's M2 diff to that file is the conversion line
plus its comment ONLY — measured on the staged diff before the commit:

```
$ git diff b78d2e425 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'
0
```

## Post-commit measurement (commit `c007e5409`) — and a measurement-methodology correction

The prescribed command `git show c007e5409 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'`
prints **2**, but those 2 are NOT diff lines — `git show` prints the commit MESSAGE first, and this
commit's message itself mentions "osStatFn" twice (the AC-CSRB-007 attribution sentences above
mirror into it). A measurement over `git show`'s full output counts the message's words together
with the diff's. Measured verbatim:

```
$ git show c007e5409 -- internal/cli/doctor_codex.go | /usr/bin/grep -n 'osStatFn'
21:    CARRIES t563's osStatFn seam (provenance c72dc1baf, sites :459/:857);
22:    this diff adds ZERO osStatFn tokens to that file (staged-diff count 0).
```

Both hits are message lines (no `+`/`-`/context prefix shown by `-n` over the raw output; the
diff body contributes none). Excluding the message, and counting added lines only:

```
$ git show --format='' c007e5409 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'
0        (grep exit 1 = zero matches)
$ git show c007e5409 -- internal/cli/doctor_codex.go | /usr/bin/grep -c '^+.*osStatFn'
0        (grep exit 1 = zero matches)
```

**Verdict: the M2 diff body adds ZERO osStatFn tokens to doctor_codex.go** — the seam's two
occurrences in the file are absorb-provenance `c72dc1baf`, not this card's. For any future reader
re-running the prescribed one-liner: use `--format=''` (or count `^+` lines), or the message's own
attribution sentences will masquerade as diff tokens.

## Doctor guard counters — IDENTICAL to the M1 baseline

`.moai/reports/t562/ac-csrb-006-baseline.log` vs `.moai/reports/t562/green-csrb-002-m2.log`,
diffed line-by-line: the ONLY difference is the per-run TempDir path (the fixture root is a fresh
`t.TempDir()` each run). Every counter phrase is byte-identical:

- `declares 4 [[skills.config]] entries`
- `1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean)`
- `1 relative entry (not checked: the resolution base is not observed)`
- `1 oddly-formed entry (not checked: backslash or ~other-user shape)`

Expected: on darwin `configPathSeparator = filepath.Separator = '/'`, where `fromConfigPath` is
the identity — a moved counter would have meant a darwin-visible behaviour change, contradicting
the seam's identity property. No counter moved.

## Coexistence with the absorbed t563 tests

Scoped run of the absorbed file's surface — `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_|TestInspectSkillMirror_' -count=1 -timeout 1800s -v`:
**27 PASS (26 absorbed t563 tests + this card's ShapeReadbackBaselineGuard), 0 FAIL**, `ok` in 41.9s.

## Boundary checks on the committed diff

- Classification still runs on the DECLARED form (the conversion sits INSIDE the
  `codexPathAbsolute` case, after `classifyCodexSkillPath`) — AC-CSRB-003 guard PASS confirms.
- Home-relative branch untouched in both readers (Join products never converted) — AC-CSRB-004
  guard PASS confirms.
- No blanket stat-site wrap (the conversion is at the branch, not at `osStatFn(...)`) — verified
  by reading the committed diff.
- vet rc=0; golangci-lint `0 issues.` on `./internal/cli/`.
