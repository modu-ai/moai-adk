# t488 mutant falsification log — how to read the anchors

Every `m*.txt` / `f*.txt` file here is the verbatim output of a `go test` run
with one named mutation applied to the working tree.

**The line numbers resolve.** They are anchors into
`internal/kanban/settings_drift_test.go` and
`internal/cli/integration_settings_drift_exitcode_test.go` **as delivered in
this card's final commit**. Every mutation below is applied to a
non-test source file, and a source-file edit does not shift a test file's
lines, so the cited line is the line.

That was not true of the first generation of these logs. They were produced
before the test file grew by ~30 lines, so their anchors were off by +14/+16
against the delivered tree and resolved to the wrong assertion. The RED was
real and the anchor was not, which is the same defect class as citing a commit
that no longer exists: a record claiming verifiability has to cite something
that still resolves. Every log here was regenerated after the test files were
final.

**A stable anchor is given alongside the line number**, so the citation
survives a future edit that does shift lines: each entry names the test
function and quotes the assertion text that fired.

| Log | Mutation (one line, in a source file) | Test that fired | Assertion text |
|---|---|---|---|
| `m1-predicate-deleted.txt` | `DetectSettingsDrift`'s final `return count, raw, nil` → `return 0, "", nil` | `TestSettingsDriftPredicateHit` | `match count: got 0, want 1` |
| `m2-verdict-inverted-gate.txt` | `AssessSettingsDrift`'s `if count > 0 {` → `if count == 0 {` | `TestAcquireRefusesOnDriftWithGateEnabled` (internal/cli) | `acquire succeeded on a drifted tree with the refusal layer on` |
| `m3-path-misspecified.txt` | argv path literal → `.claude/settings.local.json` | `TestSettingsDriftPredicateHit`, `…IgnoresOtherFile`, `…ArgvRecordedAtExecutionBoundary` | `recorded 0 occurrences of …` / `recorded argv … got/want` |
| `m4-pathspec-dropped.txt` | argv `--` and the path removed | `TestSettingsDriftPredicateScopedToWatchedPath` | `got 2, want 1 (2 means the pathspec was dropped)` |
| `m5-no-optional-locks-removed.txt` | argv `--no-optional-locks` removed | `TestSettingsDriftPredicateArgvRecordedAtExecutionBoundary` | `--no-optional-locks absent from executed argv` |
| `m6-executor-bypassed.txt` | a direct `exec.Command("git","commit",…)` added on the preserve path, bypassing the runner | `TestAssessSettingsDriftPreservesAndLedgers` | `primary-root HEAD moved` |
| `m7-exitcode-sweep-control.txt` | the token `ExitCode` written into a comment in `settings_drift.go` | `TestSettingsDriftVerdictNeverReadsAnExitCode` (internal/cli) | `reads an exit code ("ExitCode")` |
| `f1-double-read.txt` | the two-read shape restored: `data, _ = os.ReadFile(result.Path)` before the preserve call | `TestLedgerDigestDescribesThePreservedBytes` | `the ledger digest does not describe the preserved bytes` |

`m6` is the only entry whose mutation adds lines rather than replacing one. It
still does not move the cited anchor, because the mutation is in
`settings_drift.go` and the assertion that fires is in `settings_drift_test.go`.

Two of these are worth reading for what they say about the assertions rather
than about the code:

- **`m6`** fails on ONE line. The record-reading assertions `(d-1)` and `(d-2)`
  passed — the mutant routes only the commit around the runner, so the recorded
  command list stays clean. Only `(d-3)`, which reads no record at all, caught
  it, and it caught it on the PRIMARY root's HEAD while the target tree's HEAD
  never moved. Measuring one root would have missed it.
- **`f1`** is the falsification of a fix, not of the original implementation.
  A static fixture cannot tell one read from two, so the interleaving is
  constructed with a test hook that writes to the file between the measurement
  and the copy.

Other files here:

- `budget-baseline-before-m6.txt` — `TestAlwaysLoadedTokenBudget` measured
  before this card touched any always-loaded file (overflow 123, inherited).
- `budget-after-m6.txt` — the same measurement after the doctrine line
  (overflow 201; +78 is this card's).
