# SPEC-CLI-COVER-SUPPRESS-001 — Acceptance Criteria

> 4 ACs (AC-CSS-001..004), each independently verifiable by a file check or a recorded command
> output. This is a report-only investigation card: no AC requires — or permits — a code or test
> change (REQ-CSS-006). Every PASS row names its evidence path and, where a command produced it,
> the command and its verbatim output (VCI §2 attribution). "Current state" lines separate what
> is already measured at plan time from what the run phase must observe.

## §1 Mechanism verdict

**AC-CSS-001 — Suppression claim carries a measured verdict (both shapes, go version pinned)**
Given the claim "`go test -cover` prints no coverage line for a failing package" (t237
progress.md §E.2), When the verdict is recorded, Then it rests on the minimal two-package
experiment (package `ok` passes, package `bad` fails deliberately) measured on
`go1.26.4 darwin/arm64`, observing BOTH shapes: (a) the plain `-cover` stdout, where the failing
package prints `coverage: 100.0% of statements` on its own line above its package line
`FAIL	t481mech/bad`, and (b) the `-coverprofile` data, which contains the failing package's
line (`t481mech/bad/bad.go:4.24,6.2 1 1`).
Verification: `.moai/reports/t481/mechanism-plain.txt` (8 lines, run exit 1) +
`.moai/reports/t481/mechanism-coverprofile.txt` +
`.moai/reports/t481/mechanism-coverprofile-data.txt`; fixture at `.moai/cache/t481-mech/`.
Current state: MEASURED at plan time — verdict **REFUTED** at the toolchain level; the
run-phase act is recording it with per-row attribution (M1).

## §2 Post-repair measurement

**AC-CSS-002 — Post-t477 sweep observed; main-package number recorded (or absence recorded as an
observed fact)**
Given the absorbed develop tree containing t477's repair (`b2f98f6aa`, verified ancestor of
this HEAD), When the settled output of
`go test -count=1 -cover -timeout 900s ./internal/cli/...` (at
`.moai/reports/t481/cli-cover-post-t477-sweep.txt`) is read, Then the main `internal/cli`
package's coverage number is recorded — or, where no number appears for the package, that
absence is recorded as an observed fact with the verbatim output lines around the package
result.
Verification: the settled output file + the recorded number/absence in progress.md §E.2 with
attribution. Current state: **PENDING** — the file exists at 0 bytes at plan authoring (sweep
in flight, main-session-owned); the AC is undischarged until the sweep settles and is read, and
its value is never assumed (REQ-CSS-002). RED-now cell: the plan-phase fact that no post-t477
main-package number is recorded anywhere is the pre-work state this AC flips.

## §3 Target judgment and report-only invariant

**AC-CSS-003 — Verdict table judges every measured number against §6 targets; the card makes NO
code or test changes**
Given the measured package numbers (the sweep's rows and the mechanism files), When the verdict
table is composed, Then each measured package number is judged against its CLAUDE.local.md §6
target class — 85% minimum per package, 90%+ for the critical set, with `internal/cli` named
critical — each row carrying the measured number, the target class, and the judgment; AND the
card's total diff is confined to `.moai/reports/t481/**` +
`.moai/specs/SPEC-CLI-COVER-SUPPRESS-001/**` (zero code/test/template/config changes).
Verification: the table in the card report / progress.md §E.2; zero-diff check via the card's
commit file list (or `git diff --stat <base>..<close> -- ':!.moai/reports/t481'
':!.moai/specs/SPEC-CLI-COVER-SUPPRESS-001'` → empty). Current state: PENDING (depends on
AC-CSS-002's sweep); the zero-diff half holds by construction at plan time and must still hold
at close.

## §4 Claim correction record

**AC-CSS-004 — Refutation recorded with the verbatim original quote next to the measured
counter-evidence**
Given the refuted claim, When the correction is recorded (card report + progress.md §E.2), Then
the record carries the original sentence verbatim — with its file and tree — placed directly
next to the measured counter-evidence and the pinned go version:

> "**Unmeasured** — the package FAILs on the inherited red (see AC-PVM-010), and `go test -cover`
> prints no coverage line for a failing package. Named gap, not a claim."
> — `.moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/progress.md` §E.2 coverage row (lane-10
> authoring tree `32319621b`; sentence read directly on this branch at HEAD `8b391bc8c`)

Verification: `grep -c 'prints no coverage line for a failing package'` ≥ 1 over the card's
record, AND the counter-evidence paths (`mechanism-plain.txt`,
`mechanism-coverprofile-data.txt`) appear within the same record. Current state: quote verified
verbatim at plan time; the run-phase act is the paired recording (M1).

## §5 Quality gates

- Report-only invariant: no run-phase commit touches any path outside the two permitted roots.
- Attribution: every number in the final record names its command + verbatim output + evidence
  path; relayed values (t237's cells) are labeled as lane-10's, not re-measured here.
- The sweep under judgment keeps the exact command form
  `go test -count=1 -cover -timeout 900s ./internal/cli/...` with its exit code observed
  unpiped.
- No time estimates in any artifact (phase ordering only).
