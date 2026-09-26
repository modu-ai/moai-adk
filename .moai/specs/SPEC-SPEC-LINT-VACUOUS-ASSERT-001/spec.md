---
id: SPEC-SPEC-LINT-VACUOUS-ASSERT-001
title: "Vacuous test-assertion lint rule: move the prose-grep acceptance judgment into internal/spec"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "spec-lint, acceptance-criteria, go-test, vacuous-assertion, anchoring, ci-gate, baseline"
tier: M
related_specs: [SPEC-SPECLINT-GATE-SIGNAL-001, SPEC-MOVING-REF-GUARD-001, SPEC-INSTRUCTION-FILES-UNIFY-001]
---

# SPEC-SPEC-LINT-VACUOUS-ASSERT-001 — Vacuous test-assertion lint rule

## History

- 2026-09-26 — v0.1.0 — Initial plan-phase draft (manager-spec, card t1269). Base tree
  develop `e464fd5d0`, branch `WT-vacuous-lint-rule`. Origin: t1243 decision record
  `.moai/reports/t1243/decision-scope-reduction.md` (scope reduction after the plan-audit
  iteration cap) — this card receives the pattern judgment that SPEC removed from its prose.
- 2026-09-26 — v0.2.0 — Plan-audit iter-1 (FAIL 0.82) repair: test-name alphabet widened to Go
  subtest names (D1); word boundary and bare `(` removed from delimiters (D2); cutoff redefined
  as the day after the newest existing `created`, so no existing SPEC is gated at landing (D3);
  grouped form defined by paren matching (D4); `--- (PASS|FAIL): ` prefixes added (D5); axis-(c)
  justification narrowed with exit-code-only residual risk (D6); E2E runs the literal CI
  invocation form from each fixture root (D7) and asserts the increase line (D8); pre-cutoff
  amendment gap disclosed (D9); AC-010/012 measured against the merge base at verification time
  (D10); optional D11-D16 addressed.

## §A Background and Problem Statement

### §A.1 The defect class

Acceptance criteria in this repository verify behavior by running `go test` with a `-run`
pattern and asserting on the verbose output. Two shapes of that verification pass without the
behavior existing:

1. **Unanchored `-run` pattern.** `-run` is an unanchored regular expression. A pattern that is
   not closed with `^…$` at every alternation branch and every subtest level selects every test
   whose name merely *starts with* (or contains) the pattern. t1243 plan-audit iter-2 (D9)
   measured one criterion's pattern selecting 19 pre-existing tests and printing 74 matching
   lines against a tree where the required behavior was not implemented.
2. **Undelimited outcome assertion.** An assertion that the output contains the `--- PASS:` (or
   `--- FAIL:`) line for a test name, written without a delimiter after the name, is also
   satisfied by any longer test name sharing that prefix. Go prints the line as the name
   followed by a space and the elapsed time in parentheses, so the delimiter is the space.

A third shape — a `-run` pattern naming no test at all (t1243 D1) — is the same class seen from
the other side; §E explains why this SPEC does not judge it directly.

### §A.2 Why the prose detector failed three times

t1243 tried to guard the class with grep commands written into its own acceptance document. The
class recurred three times (D1 → D9 → D15/D19). The E1 consult recorded in the decision record
named four structural deficits of a detector that lives as prose inside the document it inspects:

1. **No execution record** — the command is a string; nothing records that it ran or what it
   printed.
2. **No positive control** — prose has no defective input, so a detector that catches nothing
   and one that catches the right thing are indistinguishable.
3. **Scope lives in the command** — the recorded glob reached every SPEC directory (941, D16)
   instead of the two it claimed.
4. **The document is its own input** — the blockquote exclusion filter needed to stop the
   detector matching its own documentation opened a blind spot for decision rules written inside
   blockquotes.

iter-3 also measured the detector's exclusion filter as a regex no-op (D15): the filter meant to
exclude lines carrying the literal two-character sequence dollar-quote was passed to `grep -E`,
where the dollar is an end-of-line anchor, so it excluded nothing; with the filter repaired, a
pattern whose FIRST alternation branch was unanchored slipped through because the line contained
the sequence elsewhere.

### §A.3 Measured seam (this tree, `e464fd5d0`)

| Fact | Observation |
|---|---|
| Rule interface | `internal/spec/lint.go` `type Rule interface { Code() string; Check(doc *SPECDoc, all []*SPECDoc) []Finding }` — single-document scope |
| Registration | plain slice `l.rules = []Rule{` in `NewLinter` (`internal/spec/lint.go:157`) |
| Sibling-artifact precedent | `MovingRefUnpinnedRule` reads its SPEC's own directory via `filepath.Dir(doc.Path)` |
| Demotion | `applyEraDemotion` marks every warning of a grandfather-era or terminal-status SPEC `Advisory` |
| Baseline gate | `CompareBaseline`: any error fails; a rule's non-advisory warning count above its recorded count fails; an absent rule records 0 |
| CI | `.github/workflows/spec-lint.yml` strict path runs `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` |
| Current baseline | `.moai/spec-lint-baseline.json` records only `TierArtifactMissing: 4` |

The end-to-end path from a violating document to a red CI step has **not** been exercised by
anyone (decision record, Gaps). This SPEC requires it measured (REQ-VTA-013).

## §B Requirements

### §B.1 Definitions (normative)

- **Scanned artifacts** — `spec.md`, `plan.md`, and `acceptance.md` in the directory of the
  SPEC under lint. These are the artifacts that carry decision rules (Tier S inline criteria live
  in `spec.md`).
- **Test-invocation line** — a line of a scanned artifact containing the token `go test`.
- **Run pattern** — the argument of a `-run` flag on a test-invocation line, in any of the forms
  `-run 'P'`, `-run "P"`, `-run P`, `-run='P'`, `-run=P`. Inside double quotes, a
  backslash-escaped dollar is read as a dollar. On a markdown table row, a backslash-escaped
  pipe is read as a pipe (markdown escaping, not regex escaping).
- **Top-level** — outside any parenthesized group and any `[...]` character class, and not
  backslash-escaped. `|` and `/` inside a group, inside a class, or escaped are not separators.
- **Anchored** — a leading RE2 flag group of the form `(?letters)` is removed first. The pattern
  is then split on top-level `/` into subtest levels, and each level on top-level `|` into
  branches (a level with no top-level `|` is one branch). The pattern is anchored when every
  level is one of:
  1. **Grouped form** — the level is `^(` + inner + `)$`, where the `(` at the second character
     is closed by the `)` at the second-to-last character under paren matching that skips
     escaped parens and character classes; the group therefore spans the whole level. A level
     that only starts with `^(` and ends with `)$` but whose opening paren closes earlier (two
     sibling groups joined by a top-level `|`) is NOT the grouped form.
  2. **Branch form** — every top-level branch begins with `^` and ends with an unescaped `$`.
  The empty-match form `^$` is anchored.
- **Outcome assertion** — an occurrence of one of the four prefixes `--- PASS: `, `--- FAIL: `,
  `--- (PASS|FAIL): `, `--- (FAIL|PASS): ` followed by a non-empty test name. A test name is a
  Go identifier (a letter or underscore, then letters, digits, underscores) optionally followed
  by `/`-separated subtest components; a subtest component is one or more characters other than
  whitespace, single quote, double quote, backtick, backslash, and `(`. This admits the `-`,
  `.`, `#`, `=`, `+` and non-ASCII characters Go keeps in subtest names.
- **Delimited** — an outcome assertion whose name is immediately followed by a whitespace
  delimiter: a space, a tab, `\s`, or `[[:space:]]`. Nothing else is a delimiter: the regex
  word boundary is NOT one, because it holds between a parent name and the `/` of a passing
  subtest printed under a failing parent; and `(` directly after the name is NOT one, because
  Go always prints a space before the elapsed time, so such an assertion can never match.
- **Gate cutoff** — a single named date constant in the rule's source file. Selection rule: at
  run phase, measure the newest `created` value over every `.moai/specs/*/spec.md` in the tree
  that will register the rule, and set the cutoff to the day after it (strictly later than
  every existing `created`). Plan-time measurement: newest `created` is 2026-09-26 (14 SPECs),
  zero SPECs carry a later date, so the value would be 2026-09-27; plan.md M4 re-measures.

### §B.2 GEARS requirements

- REQ-VTA-001 — The `internal/spec` linter shall register one rule with the finding code
  `VacuousTestAssertion`, implemented in `internal/spec/lint_vacuous_assertion.go`, as an entry of
  the existing rule slice built in `NewLinter`.
- REQ-VTA-002 — The rule shall read only the scanned artifacts of the SPEC under lint, located
  through that SPEC's own `spec.md` path, and shall not read any other directory, any Go source
  file, `progress.md`, `spec-compact.md`, `research.md`, or `design.md`.
- REQ-VTA-003 — When a test-invocation line carries a run pattern that is not anchored, the rule
  shall emit one finding for that pattern.
- REQ-VTA-004 — The rule shall treat every anchored run pattern as conformant, including the
  grouped alternation form, the per-branch alternation form, per-level anchored subtest patterns,
  a leading flag group, a double-quoted pattern whose closing dollar is backslash-escaped, and a
  table-row pattern whose alternation pipe is markdown-escaped; and it shall NOT treat as grouped
  a level whose opening group closes before the level's final `$`.
- REQ-VTA-005 — When a scanned-artifact line carries an outcome assertion (any of the four
  prefixes) that is not delimited, the rule shall emit one finding for that assertion, whether or
  not the line is a test-invocation line; a subtest name containing `-`, `.`, or other non-space
  characters Go preserves shall be read whole before the delimiter is judged.
- REQ-VTA-006 — The rule shall judge every line of a scanned artifact irrespective of markdown
  context — fenced code block, inline code span, table row, prose, and blockquote at any nesting
  depth — and shall not exclude a line because it begins with a blockquote marker.
- REQ-VTA-007 — When a run pattern contains a shell expansion (a dollar followed by a letter,
  underscore, `{`, or `(`), the rule shall not emit a finding for it, and the rule's source
  comment shall state this as a known detection limit.
- REQ-VTA-008 — Each finding shall carry the artifact path, the 1-based line number within that
  artifact, the axis (`run-pattern` or `outcome-assertion`), the offending text, and the
  conformant form to use instead.
- REQ-VTA-009 — Where a SPEC's `created` frontmatter date is on or after the gate cutoff, the
  rule shall emit its findings as non-advisory warnings; where the date is earlier, the rule
  shall emit them as advisory warnings. A missing or unparseable `created` value shall be treated
  as earlier than the cutoff (advisory): on a modern-era SPEC that absence is already an error
  from `FrontmatterSchemaRule`, so gating it here adds no signal, and the one corpus SPEC lacking
  a parseable `created` (a legacy `created_at:` alias, measured at plan time) must not become
  gated at landing. The existing era and terminal-status demotion shall apply unchanged on top of
  this.
- REQ-VTA-010 — The gate cutoff shall be a single named constant whose value a unit test pins,
  so that moving the cutoff requires editing the pinning test in the same change.
- REQ-VTA-011 — The only suppression mechanism shall be the existing frontmatter `lint.skip`
  list; the rule shall not introduce a per-line exemption marker.
- REQ-VTA-012 — The change shall not add a CI workflow, a CI job, a CI step, a new gate, or a
  module dependency, and shall not modify `.moai/spec-lint-baseline.json`.
- REQ-VTA-013 — When a SPEC tree contains a non-advisory `VacuousTestAssertion` finding and the
  baseline file records no count for that code, the CI lint invocation — `go run` of the
  `cmd/moai` package with the argument vector `spec lint --baseline
  .moai/spec-lint-baseline.json`, run from that tree's root — shall exit non-zero and print the
  baseline increase line for `VacuousTestAssertion` as the only rule listed under the exceeded
  verdict; for the same tree with the one violating line made conformant, it shall exit 0.
- REQ-VTA-014 — The rule shall not raise any error-severity finding, and at the commit that
  registers it no SPEC in the repository corpus shall carry a `created` date on or after the
  gate cutoff, so the corpus carries zero non-advisory `VacuousTestAssertion` findings and the
  unchanged baseline gate stays green on landing without editing any other card's SPEC.

### §B.3 Design decisions (normative summary; rationale in plan.md §B)

| # | Decision | Reason in one line |
|---|---|---|
| DD-1 | Axes (a) run-pattern and (b) outcome-assertion are in scope; axis (c) "pattern matches no test" is out | (c) needs Go source a single-document rule does not hold, and at plan phase absent tests are the expected state; for a criterion that also carries a delimited outcome assertion, (a)+(b) make (c) fail visibly at execution time — an exit-code-only criterion is not covered (§E residual risk) |
| DD-2 | Scan `spec.md`, `plan.md`, `acceptance.md` only | decision rules live there; `progress.md` records commands that already ran, `spec-compact.md` duplicates `spec.md` and would double-count |
| DD-3 | Severity `warning`, non-advisory only for SPECs created on or after the gate cutoff | advisory cannot turn the baseline gate red (REQ-VTA-013 would be unsatisfiable); error is unabsorbable; warning-plus-rebaseline absorbs hundreds of legacy findings as permanent headroom |
| DD-4 | No per-line exemption marker | `lint.skip` already exists. Measured cost accepted: a SPEC that quotes the defect as documentation (2 of the 4 non-terminal SPECs created 2026-09-26 do, per plan-audit iter-1) must either use whole-code `lint.skip` — which also silences its real criteria — or move the examples into Go fixtures. The cutoff (DD-3) keeps every existing such SPEC advisory, so no current SPEC pays this cost |

## §C Non-Functional Constraints

- Simplicity: the rule file is expected near 150 lines; above 450 lines (3× the estimate) the
  implementation stops and is simplified before proceeding.
- Performance: the rule reads each scanned artifact once per lint run, spawns no process, and
  performs no network or git access.
- Coverage: the rule file reaches at least 85% statement coverage.

## §D Acceptance Criteria

The full criteria, with their commands, live in `acceptance.md` §D. This index carries the REQ
traceability the linter reads; every detection axis has a detection arm and a conformant arm.

- AC-VTA-001: Given single test-invocation lines, When the run-pattern test runs, Then every unanchored form fires once, including two sibling groups joined by a top-level pipe, and every anchored form is silent (maps REQ-VTA-003, REQ-VTA-004)
- AC-VTA-002: Given single lines, When the outcome-assertion test runs, Then undelimited names under all four prefixes fire, the word-boundary escape fires, and whitespace-delimited names including subtest names with hyphen or dot are silent (maps REQ-VTA-005)
- AC-VTA-003: Given both arms placed in fenced, inline, table, prose, blockquote and nested-blockquote contexts, When the markdown-context test runs, Then every defective instance fires and every conformant one is silent (maps REQ-VTA-006, REQ-VTA-004)
- AC-VTA-004: Given the t1243 iter-3 probe shapes, When the D15 regression test runs, Then the dollar-quote sequence elsewhere on a line never makes a pattern conformant (maps REQ-VTA-003, REQ-VTA-004)
- AC-VTA-005: Given defective lines in scanned, unscanned and sibling-directory artifacts, When the scope test runs, Then exactly the three scanned artifacts report with correct line numbers and shell expansions are not judged (maps REQ-VTA-002, REQ-VTA-007, REQ-VTA-008)
- AC-VTA-006: Given created dates around the cutoff, When the cutoff test runs, Then only on-or-after findings are non-advisory, the cutoff literal is pinned, and the red fixture's created is on or after the cutoff (maps REQ-VTA-009, REQ-VTA-010)
- AC-VTA-007: Given a defective document with and without lint.skip, and a defective line carrying a trailing HTML comment, When the linter runs, Then only lint.skip suppresses (maps REQ-VTA-011)
- AC-VTA-008: Given NewLinter, When the registration test runs, Then exactly one VacuousTestAssertion rule is registered and its findings are warnings with axis, text and conformant form (maps REQ-VTA-001, REQ-VTA-008, REQ-VTA-014)
- AC-VTA-009: Given committed red and green fixture trees, When the CI spec lint baseline invocation runs from each, Then red exits non-zero with only the VacuousTestAssertion increase line under the exceeded verdict and green exits 0 (maps REQ-VTA-013)
- AC-VTA-010: Given the registering commit, When the real-corpus baseline gate and census run, Then the gate exits 0, the newest corpus created is before the cutoff, gated findings are zero and total findings are non-zero (maps REQ-VTA-012, REQ-VTA-014)
- AC-VTA-011: Given this SPEC's own artifacts, When the linter runs on it, Then it reports zero VacuousTestAssertion findings (maps REQ-VTA-004)
- AC-VTA-012: Given the run-phase diff, When workflow, module and baseline paths are diffed, Then nothing changed (maps REQ-VTA-012)
- AC-VTA-013: Given five rule mutants including a prefix-and-suffix grouped-form check, When the axis tests run against each, Then each mutant fails a named arm (maps REQ-VTA-003, REQ-VTA-004, REQ-VTA-005, REQ-VTA-006)
- AC-VTA-014: Given the final tree, When coverage, vet and golangci-lint run, Then the rule file reaches 85% and both linters exit 0 (maps REQ-VTA-001)

## §E Exclusions (What NOT to Build)

### Out of Scope — axis (c), patterns that match no test

- Judging whether a run pattern selects any existing `Test` function. This needs the package's
  Go source, which a single-document `Rule` does not receive, and a plan-phase SPEC legitimately
  names tests that do not exist yet. An anchored pattern that matches nothing makes Go print
  `no tests to run`, and a delimited outcome assertion then fails — so for a criterion that
  carries a delimited outcome assertion, axes (a) and (b) convert the (c) defect into a visible
  execution failure.
- Residual risk accepted: a criterion that runs an anchored pattern naming a non-existent test
  and asserts only the exit code still passes vacuously (`go test` exits 0 on `no tests to run`)
  and is judged by neither axis. Requiring an outcome assertion on every test-invocation
  criterion is a different rule, not built here.

### Out of Scope — gating edits to pre-cutoff SPECs

- The gate keys on `created`, so a SPEC created before the cutoff that is still active can gain
  new vacuous criteria after the rule lands and is never gated — including in-flight SPECs whose
  run phase or follow-up cards will edit them. Those findings stay visible as advisory; gating
  them would need an edit-time signal this rule does not have. Accepted residual risk.

### Out of Scope — assertion spellings outside the four prefixes

- An outcome assertion written as a regex group other than the two `(PASS|FAIL)` orders, or as a
  pattern whose name starts with `(` or a character class, is not recognized. Stated limit.

### Out of Scope — remediation of existing SPECs

- Rewriting the unanchored patterns already present in the corpus. Findings on SPECs created
  before the gate cutoff remain visible as advisory warnings; remediation is a separate card if
  the operator wants one.

### Out of Scope — new enforcement surfaces

- A new CI workflow, job, step, pre-commit hook, or baseline entry.
- A per-line exemption marker or a new frontmatter field.
- Multi-line command reconstruction (a `-run` flag on a backslash-continued line without the
  `go test` token is not judged — known limit, stated in the rule's source).

### Out of Scope — other prose-embedded verification commands

- The general observation that every prose-embedded verification command shares three of the four
  structural deficits (decision record, Residual-risk). Only the `go test` assertion class is
  moved into code here.
