# spec-compact.md — SPEC-SPEC-LINT-VACUOUS-ASSERT-001 (derived digest)

Derived from spec.md v0.2.0 and acceptance.md v0.2.0. On conflict, spec.md and acceptance.md win.

## Requirement modules (4)

1. **Rule and scope** (REQ-VTA-001, 002): one rule, code `VacuousTestAssertion`, file
   `internal/spec/lint_vacuous_assertion.go`, registered in the `NewLinter` slice; reads only
   `spec.md`/`plan.md`/`acceptance.md` of its own SPEC directory — no other directory, no Go source.
2. **Detection** (REQ-VTA-003..008): (a) run pattern on a `go test` line not anchored at every
   top-level branch and subtest level (leading flag group stripped; grouped form only when the
   opening paren closes at the level's end under paren matching) → finding; (b) outcome assertion
   under any of `--- PASS: `, `--- FAIL: `, `--- (PASS|FAIL): `, `--- (FAIL|PASS): ` whose name
   (Go identifier plus subtest components that may contain `-`, `.` and other non-space
   characters) is not followed by a space, tab, `\s`, or `[[:space:]]` → finding on any line; word
   boundary and bare paren are not delimiters; every markdown context judged, blockquotes
   included; shell-expansion patterns not judged (stated limit); findings carry path, line, axis,
   offending text, conformant form.
3. **Gating** (REQ-VTA-009..011, 014): warning, never error; non-advisory when `created` ≥ the
   pinned cutoff, advisory otherwise (missing `created` → advisory); cutoff = day after the newest
   existing `created` in the registering tree, so no existing SPEC is gated at landing and no
   other card's SPEC is edited; era/terminal demotion unchanged; only `lint.skip` suppresses.
4. **No new surfaces + E2E** (REQ-VTA-012, 013): no workflow/job/step/gate/dependency, baseline
   file untouched; `go run` of `cmd/moai` with the CI argument vector from a violating fixture root
   exits non-zero with only the `VacuousTestAssertion` increase line under the exceeded verdict,
   and exits 0 on its one-line-fixed twin.

## Acceptance criteria (14)

AC-VTA-001 run-pattern two arms incl. sibling groups · 002 outcome-assertion two arms incl. four
prefixes, word boundary, hyphen/dot subtests · 003 markdown contexts incl. nested blockquote and
escaped table pipe · 004 D15 regression inputs · 005 artifact scope, line numbers, shell-expansion
limit · 006 cutoff split, pinned constant, red-fixture date guard · 007 lint.skip only, trailing
comment does not suppress · 008 registration, warning severity, message content · 009 end-to-end
CI invocation red/green with increase-line assertion · 010 landing gate green, newest created <
cutoff, gated 0 and total > 0 · 011 self-conformance · 012 no `.github/`/`go.mod`/baseline diff
from the merge base · 013 five mutation probes · 014 coverage ≥85%, vet, golangci-lint.

## Files to modify

- `internal/spec/lint_vacuous_assertion.go` (new)
- `internal/spec/lint_vacuous_assertion_test.go` (new)
- `internal/spec/lint.go` (one registration entry)
- `internal/spec/testdata/vacuous_assert_e2e/{red,green}/` (new fixtures)

## Exclusions

- Axis (c): judging whether a pattern matches any existing test; exit-code-only criteria remain a
  residual risk.
- Gating new criteria added to pre-cutoff SPECs.
- Assertion spellings outside the four prefixes.
- Remediating existing corpus findings.
- New CI workflow/job/step, baseline entry, per-line marker, frontmatter field, multi-line command
  reconstruction.
- Other prose-embedded verification command classes.
