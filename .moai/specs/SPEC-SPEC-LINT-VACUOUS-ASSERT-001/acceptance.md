---
id: SPEC-SPEC-LINT-VACUOUS-ASSERT-001
title: "Acceptance Criteria — Vacuous test-assertion lint rule"
version: "0.1.0"
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
- Every observed result is recorded in `.moai/reports/t1269/verdict.md` as command, verbatim
  output (or exit code plus bounded tail), and tree SHA.

## §D AC Matrix

### AC-VTA-001 — Run-pattern axis, two arms

**Given** table-driven inputs of single test-invocation lines **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_RunPatternAxis$' -v` runs **Then** it
exits 0 with `--- PASS: TestVacuousAssertionRule_RunPatternAxis (` present, and the test asserts:
detection arm — a pattern missing the trailing dollar, a pattern missing the leading caret, a
bare unquoted name, the `-run=` form unanchored, an alternation with the FIRST branch
unanchored, an alternation with the LAST branch unanchored, and an anchored top level with an
unanchored subtest level each yield exactly one finding; conformant arm — the single anchored
name, the grouped alternation form, the per-branch alternation form, per-level anchored subtests,
the empty-match form, a double-quoted pattern with a backslash-escaped closing dollar, and a
double-quoted pattern with a plain closing dollar each yield zero findings.
Maps: REQ-VTA-003, REQ-VTA-004.

### AC-VTA-002 — Outcome-assertion axis, two arms

**Given** single-line inputs **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_OutcomeAssertionAxis$' -v` runs
**Then** it exits 0 with `--- PASS: TestVacuousAssertionRule_OutcomeAssertionAxis (` present,
and the test asserts: detection arm — PASS and FAIL assertions whose name ends at a closing
single quote, a closing double quote, a backtick, end of line, and a dollar each yield one
finding, including on a line with no `go test` token; conformant arm — the name followed by a
space, by a tab, by `(`, by `\s`, by `\b`, and by `[[:space:]]` each yield zero findings.
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

### AC-VTA-006 — Gate cutoff split and pinning

**Given** the same defective document with `created` one day before the cutoff, equal to the
cutoff, after the cutoff, and missing **When**
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_GateCutoff$' -v` runs **Then** it
exits 0 with `--- PASS: TestVacuousAssertionRule_GateCutoff (` present; findings are advisory
only for the day-before case, and non-advisory warnings for the other three; the test also
asserts the cutoff constant's literal value, so changing it without editing the test turns the
test red.
Maps: REQ-VTA-009, REQ-VTA-010.

### AC-VTA-007 — Suppression only through lint.skip

**Given** a defective document whose frontmatter `lint.skip` lists `VacuousTestAssertion`, and
the same document without it **When** the full `Linter` runs over each via
`go test ./internal/spec/ -run '^TestVacuousAssertionRule_LintSkip$' -v` **Then** it exits 0
with `--- PASS: TestVacuousAssertionRule_LintSkip (` present; the report carries zero
`VacuousTestAssertion` findings for the first and at least one for the second. And
`grep -n "vacuous-assert-ok\|<!-- vacuous" internal/spec/lint_vacuous_assertion.go` returns
nothing (exit 1) — no per-line marker exists.
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
`.../green` (each a project root with `.moai/specs/SPEC-FIXTURE-VTA-001/` and a `baseline.json`
without a `VacuousTestAssertion` entry; the trees differ by exactly one line, which in `red` is an
unanchored run pattern inside a blockquote) **When** the binary is built from the tree under test
and the CI invocation is run from each fixture root:

```
BIN="$(mktemp -d)/moai" && go build -o "$BIN" ./cmd/moai
(cd internal/spec/testdata/vacuous_assert_e2e/red && "$BIN" spec lint --baseline baseline.json); echo "rc=$?"
(cd internal/spec/testdata/vacuous_assert_e2e/green && "$BIN" spec lint --baseline baseline.json); echo "rc=$?"
```

**Then** the `red` run prints a non-zero `rc` and its output contains `VacuousTestAssertion`
(the baseline increase line naming the code), and the `green` run prints `rc=0`. The binary is
built from the same package CI's `go run ./cmd/moai` compiles; `go run` cannot be used verbatim
because it requires the module root as working directory while the linter reads the working
directory as the project root — this is the only deviation from the CI step, and the argument
vector after the binary is identical. A unit test alone does not satisfy this criterion.
Maps: REQ-VTA-013.

### AC-VTA-010 — Landing: real corpus gate stays green with the baseline untouched

**Given** the commit that registers the rule **When**
`go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json; echo "rc=$?"` and
`git diff --exit-code e464fd5d0 -- .moai/spec-lint-baseline.json; echo "rc=$?"` run from the
worktree root **Then** both print `rc=0`, and the census
`go run ./cmd/moai spec lint --json | jq '[.[] | select(.code=="VacuousTestAssertion")] | {total: length, gated: (map(select(.advisory != true)) | length)}'`
reports `gated` equal to 0; the `total` value is recorded in `verdict.md` as the measured corpus
count replacing the plan-time estimate.
Maps: REQ-VTA-012, REQ-VTA-014.

### AC-VTA-011 — Self-conformance

**Given** this SPEC's own artifacts **When**
`go run ./cmd/moai spec lint SPEC-SPEC-LINT-VACUOUS-ASSERT-001 --json | jq '[.[] | select(.code=="VacuousTestAssertion")] | length'`
runs **Then** it prints `0` — the SPEC that removes the defect does not carry it.
Maps: REQ-VTA-004.

### AC-VTA-012 — No new gate or dependency

**Given** the run-phase diff against `e464fd5d0` **When**
`git diff --name-only e464fd5d0 -- .github/ go.mod go.sum .moai/spec-lint-baseline.json` runs
**Then** it prints nothing.
Maps: REQ-VTA-012.

### AC-VTA-013 — Mutation probes (the arms discriminate)

**Given** four temporary mutants of the rule, each reverted after its probe — (M1) `Check`
returns nil; (M2) every run pattern is reported; (M3) lines beginning with a blockquote marker are
skipped; (M4) any line containing the dollar-quote sequence is treated as anchored **When** the
tests of AC-VTA-001 through AC-VTA-004 run against each mutant **Then** M1 fails a detection arm,
M2 fails a conformant arm, M3 fails AC-VTA-003, M4 fails AC-VTA-004; each failing test name and
the restored-tree green run are recorded in `verdict.md`.
Maps: REQ-VTA-003, REQ-VTA-005, REQ-VTA-006.

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

- First-branch-unanchored alternation (AC-VTA-001, AC-VTA-004).
- Nested blockquote and markdown-escaped table pipe (AC-VTA-003).
- Missing `created` fails closed to gated (AC-VTA-006).
- Shell-expansion pattern is a stated limit, not a silent pass of a judged pattern (AC-VTA-005).

## §D.2 Definition of Done

- AC-VTA-001 … AC-VTA-014 observed and recorded in `.moai/reports/t1269/verdict.md`.
- progress.md §E.2/§E.3 populated by the run phase.
