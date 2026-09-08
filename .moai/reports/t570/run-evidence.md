# SPEC-CODEX-DOCTOR-PATH-GUARD-001 — Run-phase Evidence (card t570)

Worktree `.claude/worktrees/t570` · branch `WT-codex-doctor-guard` · card base
`a4855f0b2834f5179d07ebcd4c72ff0b452d6aa3` (= `git merge-base origin/develop HEAD`).
Every measurement below was taken in this tree at plan HEAD `076847abab17064a65a89b3e7424b9ec1bf5e4a8`
(the tree the tests were authored on; the guard test file was the only source added).

`grep` on this host is a ugrep wrapper that skips files silently, so every cited grep is `/usr/bin/grep`.

---

## Claim

Three new non-parallel tests in `internal/cli/doctor_codex_path_guard_test.go` pin the doctor-side
declared-path conversion at `codexStaleSkillFinding`'s `codexPathAbsolute` arm. Reverting
`internal/cli/doctor_codex.go` to `statPath = e.Path` now turns the package red; before this card it
did not. No production file was changed.

## Evidence — AC matrix

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-CDPG-001 | PASS | `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_(Absolute\|ExtendedLength)StatTargetIsConvertedForm' -timeout 1200s -v` | `--- PASS: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)` (`m1-green.log`) |
| AC-CDPG-002 | PASS | same invocation | `--- PASS: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)` (`m1-green.log`) |
| AC-CDPG-003 | PASS | `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_ClassificationPrecedesConversion' -timeout 1200s -v` | `--- PASS: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)` (`m2-green.log`) |
| AC-CDPG-004 | PASS | mutants M-1, M-2, M-3 executed and reverted; M-4 executed as a boundary probe | see § Mutant table |
| AC-CDPG-005 | PASS | `/usr/bin/grep -rn 'type statRecorder' internal/cli/ \| wc -l` and the same for `type pruneReadbackStatRecorder` | `1` and `1` — unchanged from the `a4855f0b2` baseline |
| AC-CDPG-006 | discharged by constraint (no AC by design) | read the three test functions | none calls `t.Parallel()`; the separator restores through `overrideSeparator`'s `t.Cleanup`, the stat seam through `stubStatRecording`'s `t.Cleanup`, `CODEX_HOME` / `codexUserHomeDir` through `stubCodexHome`'s `t.Setenv` + `t.Cleanup`. This card introduces no new override of its own. |
| AC-CDPG-007 | PASS | `git diff a4855f0b2834f5179d07ebcd4c72ff0b452d6aa3 -- internal/cli/doctor_codex.go` | empty (0 bytes) — see § Scope |

**Swept-count check (an empty sweep is not a pass).** The `-v` runs print one `=== RUN` line per
matched test: 2 for the AC-001/002 selector, 1 for the AC-003 selector, 3 for the combined selector
used in the M-4 probe. Sibling packages under `./internal/cli/...` correctly print
`testing: warning: no tests to run` — an empty sweep, visible rather than silent.

### Full package

```
$ go test ./internal/cli/... -timeout 1200s
ok  	github.com/modu-ai/moai-adk/internal/cli	611.738s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	1.253s
… (17 packages, all ok)
rc=0
```

`611.738s` re-confirms in this run why `-timeout 1200s` is mandatory: the 600 s default would have
failed the package spuriously. Full log: `full-package.log`.

`go vet ./internal/cli/` exits 0 with no output (`m3-green.log`).

## Mutant table (AC-CDPG-004)

Each mutation window was one test run long. `git diff -- internal/cli/doctor_codex.go` was confirmed
empty after every production-file mutant before the card proceeded.

| # | Mutation | Result | FAIL log | PASS-after-revert log | Revert verified |
|---|---|---|---|---|---|
| M-1 | `statPath = fromConfigPath(e.Path, configPathSeparator)` → `statPath = e.Path` | **CAUGHT** — AC-001 and AC-002 both FAIL | `m1-red.log` | `m1-green.log` | `m1-revert-check.txt` (0 bytes) |
| M-2 | classify the converted string: `classifyCodexSkillPath(fromConfigPath(e.Path, configPathSeparator))` | **CAUGHT** — AC-003 FAILs on the Detail string | `m2-red.log` | `m2-green.log` | `m2-revert-check.txt` (0 bytes) |
| M-3 | a second `type statRecorder` declared in the new test file | **CAUGHT** — the count prints `2`; `go vet` reports `statRecorder redeclared in this block` | `m3-red.log` | `m3-green.log` (count `1`, vet silent) | test-file-only; the production file was untouched throughout |
| M-4 | `statPath = strings.ReplaceAll(e.Path, "/", "\\")` — seam bypassed, separator hard-coded | **MISSED** — all three tests PASS | `m4-missed.log` (all PASS) | n/a | `m4-revert-check.txt` (0 bytes) |

### RED evidence, verbatim (M-1, captured before the guard could pass)

```
=== RUN   TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm
    doctor_codex_path_guard_test.go:52: stat target = "/Users/u/skills/probe/SKILL.md", want the converted form "\\Users\\u\\skills\\probe\\SKILL.md"
    doctor_codex_path_guard_test.go:55: stat target = "/Users/u/skills/probe/SKILL.md" — the DECLARED form: the conversion is not applied
--- FAIL: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)
=== RUN   TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm
    doctor_codex_path_guard_test.go:93: stat target = "//?/C:/Users/u/skills/probe/SKILL.md", want the converted extended-length form "\\\\?\\C:\\Users\\u\\skills\\probe\\SKILL.md"
--- FAIL: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.186s
```

This is the exact revert that left the package green before this card. It no longer does.

### M-2 RED, verbatim

```
=== RUN   TestCodexStaleSkillFinding_ClassificationPrecedesConversion
    doctor_codex_path_guard_test.go:124: detail does not report the entry as relative:
        …/.codex/config.toml declares 1 [[skills.config]] entry; 1 oddly-formed entry (not checked: backslash or ~other-user shape)
    doctor_codex_path_guard_test.go:127: detail reports the entry as oddly-formed — the declared string was converted BEFORE classification:
        …/.codex/config.toml declares 1 [[skills.config]] entry; 1 oddly-formed entry (not checked: backslash or ~other-user shape)
--- FAIL: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)
```

(The `…` elides only the absolute `t.TempDir()` prefix, which differs per run; the full line is in
`m2-red.log`.)

### The AC-CDPG-003 non-discriminating statement (required verbatim by acceptance.md)

> a zero-stat-call assertion does NOT discriminate the two classification orders

Both the `codexPathRelative` and the `codexPathOddlyFormed` branches `continue` BEFORE reaching
`osStatFn`, so the recorder counts 0 under the correct order and under the forbidden one alike. The
M-2 run above confirms this directly rather than merely arguing it: under the forbidden order the
two Detail assertions fired while the `len(rec.paths) != 0` assertion stayed silent. The zero-count
assertion is retained in the test only as a supporting check, and is labelled non-discriminating in
the test's own comment.

### Missed mutants — the guard's boundary, recorded rather than omitted

**M-4 was missed, and that is the honest reach of this guard.** Substituting
`strings.ReplaceAll(e.Path, "/", "\\")` for `fromConfigPath(e.Path, configPathSeparator)` produces
byte-identical stat targets while the separator is pinned to `'\\'`, so all three tests pass. The
guard therefore pins **the converted value under a pinned separator**, not the *use of the
`fromConfigPath` / `configPathSeparator` seam*. A refactor that hard-codes the separator at this call
site — discarding the seam that lets a host with a different separator behave correctly — is not
caught here. Catching it would need a second separator value (a `'/'`-pinned run asserting the
identity), which is outside this card's scope and is stated as a gap below, not as coverage.

No other mutant was attempted.

## Baseline-attribution

- Tree: worktree `.claude/worktrees/t570`, branch `WT-codex-doctor-guard`, plan HEAD
  `076847abab17064a65a89b3e7424b9ec1bf5e4a8`; card base `a4855f0b2834f5179d07ebcd4c72ff0b452d6aa3`.
- Every command above was run in this tree in this run. No figure is carried from another tree,
  another package, or another point in time.
- The `a4855f0b2` `type statRecorder` → 1 baseline is quoted from `progress.md` §E.1 (plan-phase,
  same worktree) and re-measured this run to `1`. The AC asserts the count is UNCHANGED, and both
  ends of that comparison were observed.

## Scope (AC-CDPG-007)

```
$ git diff a4855f0b2834f5179d07ebcd4c72ff0b452d6aa3 -- internal/cli/doctor_codex.go
(empty)
```

The `git diff --stat` measurement against the same base is recorded in `final-scope.txt`.

## Gaps — explicitly NOT observed

- **Windows resolution is not measured.** The guard pins the argument handed to `osStatFn`, never the
  filesystem outcome. That `\\?\C:\…` resolves on Windows while `//?/C:/…` does not is cited from
  documented platform behaviour in `spec.md` §B.2; nothing here measured it. Out of scope per
  `spec.md` §D.
- **The seam itself is unguarded** (M-4, above). A `'/'`-pinned identity companion test would close
  it; not written under this card.
- **The home-relative branch is untouched and unasserted.** t562 REQ-CSRB-002 left it unconverted
  deliberately; no fixture here exercises it, so nothing in this card constrains it.
- **`golangci-lint` was not run.** No lint measurement is claimed for this card; the only source
  added is a test file, and `go vet ./internal/cli/` exits 0.
- **The full run was not `-v`**, so the three guard tests' individual PASS lines come from the
  targeted `-v` invocations, not from `full-package.log`. `full-package.log` establishes only that
  the package as a whole is green (`rc=0`).
- **Cross-platform build (`GOOS=windows go build`) was not run.** This card adds no production code
  and no build tags, so no cross-platform claim is made in either direction.

## Residual risk

- All three tests pin `configPathSeparator` to `'\\'`. Should `overrideSeparator` ever stop pinning
  (or be changed to pin `'/'`), all three would go vacuously green on darwin — the exact ceiling this
  card lifted, reintroduced one level up inside the shared helper. Nothing mechanical prevents it.
- AC-CDPG-003's discriminator is the rendered Detail string. A later change collapsing the `relative`
  and `oddly-formed` renderings into one wording removes the discriminator silently and the test
  still passes. `spec.md` §E A-2 names this and requires re-anchoring, not weakening.
- The M-1 revert is the mutation this card was written against. A *different* way of losing the
  conversion — untried here beyond M-4 — may still slip through.
