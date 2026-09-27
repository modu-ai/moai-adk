---
id: SPEC-SPEC-LINT-VACUOUS-ASSERT-001
title: "Acceptance Criteria — Vacuous test-assertion lint rule"
version: "0.2.0"
created: 2026-09-26
updated: 2026-09-26
---

# acceptance.md — SPEC-SPEC-LINT-VACUOUS-ASSERT-001

Conventions binding every criterion below:

- Every `go test` command runs with `-v` and a `-run` pattern anchored at both ends. A test
  criterion passes only when the exit code is 0, the output contains the named outcome line
  followed by a space and `(`, and the output does not contain `no tests to run`.
- Every detection axis has a **detection arm** (defective input → the rule fires) and a
  **conformant arm** (conformant input → the rule is silent). A criterion whose arms are not both
  observed is not passed.
- Detection-arm inputs are described in prose here and written literally only inside the Go test
  fixtures, so this document stays conformant under its own rule (AC-VTA-011).
- Every observed result is recorded in `.moai/reports/t1269/verdict.md` as command, verbatim
  output (or exit code plus bounded tail), and tree SHA. Base-relative criteria (AC-VTA-010,
  AC-VTA-012) record the merge base they were measured against, recomputed at verification time
  as `git merge-base HEAD develop`, never a SHA pinned in this document.

## §D AC Matrix

### AC-VTA-001 — Run-pattern axis, two arms

**Given** table-driven inputs of single test-invocation lines **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_RunPatternAxis$' -v` runs **Then** it
exits 0 with `--- PASS: TestVacuousAssertionRule_RunPatternAxis (` present, and the test asserts:
detection arm — a pattern missing the trailing dollar, a pattern missing the leading caret, a
bare unquoted name, the `-run=` form unanchored, an alternation with the FIRST branch
unanchored, an alternation with the LAST branch unanchored, an anchored top level with an
unanchored subtest level, and **two sibling groups joined by a top-level pipe**, i.e. the level
`^(TestA)|(TestB)$` (which Go reads as a caret-anchored first group or a dollar-anchored second
group, matching longer names at either end) each yield exactly one finding; conformant arm —
the single anchored name, the grouped alternation form, the per-branch alternation form,
per-level anchored subtests, the empty-match form, a leading `(?i)` flag group before an anchored
name, a `(?i)` flag group at the head of an anchored subtest level, a double-quoted pattern with a backslash-escaped closing dollar, and a double-quoted
pattern with a plain closing dollar each yield zero findings.
Maps: REQ-VTA-003, REQ-VTA-004.

### AC-VTA-002 — Outcome-assertion axis, two arms

**Given** single-line inputs **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_OutcomeAssertionAxis$' -v` runs
**Then** it exits 0 with `--- PASS: TestVacuousAssertionRule_OutcomeAssertionAxis (` present,
and the test asserts:
detection arm — for each of the four prefixes (`--- PASS: `, `--- FAIL: `, `--- (PASS|FAIL): `,
`--- (FAIL|PASS): `), a name ending at a closing single quote, a closing double quote, a
backtick, end of line, and a dollar each yield one finding, including on a line with no `go test`
token; a name followed by the regex word-boundary escape yields one finding; a name followed
directly by `(` yields one finding; a subtest name containing a hyphen that ends at a closing
single quote yields one finding whose reported name includes the hyphenated component;
conformant arm — the name followed by a space, by a tab, by `\s`, and by `[[:space:]]` each
yield zero findings, and so do the whitespace-delimited subtest names
`--- PASS: TestX/fail-open_path (` and `--- PASS: TestX/spec-workflow.md (`, a subtest name
followed by `[[:space:]]` (`--- PASS: TestX/case[[:space:]]`), and a subtest name carrying a
regex-escaped dot followed by a space (`--- PASS: TestX/spec-workflow\.md `). On a markdown
table row, the `(PASS|FAIL)` prefix written with a markdown-escaped pipe is recognized: an
undelimited name after it yields one finding and a space-delimited one yields zero.
Maps: REQ-VTA-005.

### AC-VTA-003 — Markdown context, including blockquotes

**Given** one defective and one conformant instance of each axis placed in a fenced code block,
an inline code span, a markdown table row, plain prose, a single-level blockquote, and a nested
blockquote **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_MarkdownContexts$' -v` runs **Then**
it exits 0 with `--- PASS: TestVacuousAssertionRule_MarkdownContexts (` present, every defective
instance fires (blockquote instances included), every conformant instance is silent, and a table
row whose alternation pipe is markdown-escaped is judged as an alternation in both arms.
Maps: REQ-VTA-006, REQ-VTA-004.

### AC-VTA-004 — D15 regression inputs

**Given** the t1243 iter-3 probe shapes: a line that carries the D15 exclusion-filter text
together with an unanchored run pattern; a line whose first alternation branch is unanchored and
last branch anchored (so the line contains the dollar-quote sequence); and a line carrying only
an anchored run pattern **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_D15Regression$' -v` runs **Then** it
exits 0 with `--- PASS: TestVacuousAssertionRule_D15Regression (` present, the first two inputs
each fire exactly once and the third is silent — the presence of the dollar-quote sequence
elsewhere on a line never makes a pattern conformant.
Maps: REQ-VTA-003, REQ-VTA-004.

### AC-VTA-005 — Artifact scope and shell-expansion limit

**Given** a temporary SPEC directory whose `spec.md`, `plan.md`, and `acceptance.md` each carry
one defective line, whose `progress.md`, `spec-compact.md`, `research.md`, and `design.md` carry
defective lines, and a sibling SPEC directory carrying defective lines **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_ArtifactScope$' -v` runs **Then** it
exits 0 with `--- PASS: TestVacuousAssertionRule_ArtifactScope (` present and the rule reports
exactly 3 findings whose paths are the three scanned artifacts with correct 1-based line numbers;
a run pattern containing a shell variable expansion yields 0 findings.
Maps: REQ-VTA-002, REQ-VTA-007, REQ-VTA-008.

### AC-VTA-006 — Gate cutoff split, pinning, and fixture date

**Given** the same defective document with `created` one day before the cutoff, equal to the
cutoff, after the cutoff, missing, and present but malformed (a single-digit month, and non-date
text) **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_GateCutoff$' -v` runs **Then** it
exits 0 with `--- PASS: TestVacuousAssertionRule_GateCutoff (` present; findings are
non-advisory warnings for the equal, after, and malformed cases and advisory for the day-before
and missing cases; the test asserts the cutoff constant's literal value, so changing it without editing the
test turns the test red; and the test reads the `created` value of the committed red fixture
(AC-VTA-009) and asserts it is on or after the cutoff, so a later cutoff move cannot silently
turn the end-to-end proof advisory.
Maps: REQ-VTA-009, REQ-VTA-010.

### AC-VTA-007 — Suppression only through lint.skip

**Given** a defective document whose frontmatter `lint.skip` lists `VacuousTestAssertion`, the
same document without it, and a defective line followed on the same line by an HTML comment of
arbitrary text **When** the full `Linter` runs over each via
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_LintSkip$' -v` **Then** it exits 0
with `--- PASS: TestVacuousAssertionRule_LintSkip (` present; the report carries zero
`VacuousTestAssertion` findings for the first, at least one for the second, and exactly one for
the commented line — no comment text suppresses a finding.
Maps: REQ-VTA-011.

### AC-VTA-008 — Registration, severity, and finding content

**Given** the linter built by `NewLinter` **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_Registered$' -v` runs **Then** it exits
0 with `--- PASS: TestVacuousAssertionRule_Registered (` present, asserting: the rule slice
contains exactly one rule whose `Code()` is `VacuousTestAssertion`; a defective document linted
through `Linter.Lint` yields findings of severity warning (never error) whose message names the
axis, quotes the offending text, and states the conformant form.
Maps: REQ-VTA-001, REQ-VTA-008, REQ-VTA-014.

### AC-VTA-009 — End to end: a violating document turns the CI command red

**Given** the committed fixture trees `internal/spec/testdata/vacuous_assert_e2e/red` and
`.../green`. Each is a project root holding `.moai/specs/SPEC-FIXTURE-VTA-001/` (a testdata-only
id, never placed under the repository's own `.moai/specs/`) and `.moai/spec-lint-baseline.json`
with no `VacuousTestAssertion` entry. The two trees differ by exactly one line, which in `red` is
an unanchored run pattern inside a blockquote.
**When** the CI step's own invocation — `go run` of the `cmd/moai` package with the argument
vector `spec lint --baseline .moai/spec-lint-baseline.json`, identical to
`.github/workflows/spec-lint.yml` after the package path — runs from each fixture root:

```
(cd internal/spec/testdata/vacuous_assert_e2e/red && go run ../../../../../cmd/moai spec lint --baseline .moai/spec-lint-baseline.json); echo "rc=$?"
(cd internal/spec/testdata/vacuous_assert_e2e/green && go run ../../../../../cmd/moai spec lint --baseline .moai/spec-lint-baseline.json); echo "rc=$?"
```

**Then** the `red` run prints a non-zero `rc`, its output contains the line
`baseline: EXCEEDED — .moai/spec-lint-baseline.json` followed by the increase line
`  VacuousTestAssertion: recorded 0 -> current 1 (+1)`, no other rule appears as an increase
line under that verdict, and no `baseline: ERROR-GATED` line appears; the `green` run prints
`rc=0`. The only difference from the CI step is the relative package path, which is how
`go run` reaches the module package from a working directory that is the fixture root. A unit
test alone does not satisfy this criterion.
Maps: REQ-VTA-013.

### AC-VTA-010 — Landing: real corpus gate green, no existing SPEC gated

**Given** the commit that registers the rule **When** these run from the worktree root:
`go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json; echo "rc=$?"`;
the newest-created measurement
`grep -h -m1 '^created:' .moai/specs/*/spec.md | sed 's/created: *//; s/"//g' | sort | tail -1`;
and the census
`go run ./cmd/moai spec lint --json | jq '[.[] | select(.code=="VacuousTestAssertion")] | {total: length, gated: (map(select(.advisory != true)) | length)}'`
**Then** the gate prints `rc=0`; the newest `created` value is strictly earlier than the cutoff
constant in the rule's source; the census reports `gated` equal to 0 and `total` greater than 0
(a non-zero total is the positive control that the rule is registered and reading the corpus);
`total` is recorded in `verdict.md` as the measured corpus count replacing the plan-time estimate.
Maps: REQ-VTA-012, REQ-VTA-014.

### AC-VTA-011 — Self-conformance

**Given** this SPEC's own artifacts **When**
`go run ./cmd/moai spec lint SPEC-SPEC-LINT-VACUOUS-ASSERT-001 --json | jq '[.[] | select(.code=="VacuousTestAssertion")] | length'`
runs **Then** it prints `0`. The positive control that the same invocation can report the code is
the red arm of AC-VTA-009.
Maps: REQ-VTA-004.

### AC-VTA-012 — No new gate or dependency

**Given** the run-phase branch **When**
`BASE=$(git merge-base HEAD develop) && git diff --name-only "$BASE" HEAD -- .github/ go.mod go.sum .moai/spec-lint-baseline.json`
runs at verification time **Then** it prints nothing, and `verdict.md` records the `BASE` SHA it
was measured against. Measuring from the merge base keeps changes that arrive by absorbing
`develop` out of the comparison.
Maps: REQ-VTA-012.

### AC-VTA-013 — Mutation probes (the arms discriminate)

**Given** five temporary mutants of the rule, each reverted after its probe — (M1) `Check`
returns nil; (M2) every run pattern is reported; (M3) lines beginning with a blockquote marker are
skipped; (M4) any line containing the dollar-quote sequence is treated as anchored; (M5) the
grouped form is recognized by a prefix-and-suffix check alone, without paren matching **When**
the tests of AC-VTA-001 through AC-VTA-004 run against each mutant **Then** M1 fails a detection
arm, M2 fails a conformant arm, M3 fails AC-VTA-003, M4 fails AC-VTA-004, and M5 fails AC-VTA-001
on the sibling-groups input; each failing test name and the restored-tree green run are recorded
in `verdict.md`.
Maps: REQ-VTA-003, REQ-VTA-004, REQ-VTA-005, REQ-VTA-006.

### AC-VTA-014 — Quality gates

**Given** the final tree **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_(RunPatternAxis|OutcomeAssertionAxis|MarkdownContexts|D15Regression|ArtifactScope|GateCutoff|LintSkip|Registered)$' -coverprofile=<evidence dir>/cover.out`
is followed by `go tool cover -func=<evidence dir>/cover.out | grep lint_vacuous_assertion.go`,
and `go vet ./internal/spec/...` and `golangci-lint run ./internal/spec/...` run **Then** every
function of the rule file reports coverage, the file's statement total is at least 85%, and
`go vet` and `golangci-lint` exit 0. (A new rule test added later joins this alternation; the
selector stays anchored.)
Maps: REQ-VTA-001.

## §D.1 Edge cases covered

- First-branch-unanchored alternation and sibling groups joined by a top-level pipe (AC-VTA-001,
  AC-VTA-004, AC-VTA-013 M5).
- Subtest names carrying `-` and `.`; word-boundary and bare-paren non-delimiters; the
  `(PASS|FAIL)` spelling (AC-VTA-002).
- Nested blockquote and markdown-escaped table pipe (AC-VTA-003).
- Missing `created` falls to advisory (AC-VTA-006).
- Shell-expansion pattern is a stated limit, not a silent pass of a judged pattern (AC-VTA-005).

## §D.2 Definition of Done

- AC-VTA-001 … AC-VTA-014 observed and recorded in `.moai/reports/t1269/verdict.md`.
- progress.md §E.2/§E.3 populated by the run phase.
