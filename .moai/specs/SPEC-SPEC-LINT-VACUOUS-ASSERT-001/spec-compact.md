# spec-compact.md — SPEC-SPEC-LINT-VACUOUS-ASSERT-001 (derived digest)

Derived from spec.md v0.1.0 and acceptance.md v0.1.0. On conflict, spec.md and acceptance.md win.

## Requirement modules (4)

1. **Rule and scope** (REQ-VTA-001, 002): one rule, code `VacuousTestAssertion`, file
   `internal/spec/lint_vacuous_assertion.go`, registered in the `NewLinter` slice; reads only
   `spec.md`/`plan.md`/`acceptance.md` of its own SPEC directory — no other directory, no Go source.
2. **Detection** (REQ-VTA-003..008): (a) run pattern on a `go test` line not anchored at every
   alternation branch and subtest level → finding; (b) `--- PASS:`/`--- FAIL:` assertion name not
   followed by space/tab/`(`/`\s`/`\b`/`[[:space:]]` → finding on any line; every markdown
   context judged, blockquotes included; shell-expansion patterns not judged (stated limit);
   findings carry path, line, axis, offending text, conformant form.
3. **Gating** (REQ-VTA-009..011, 014): warning, never error; non-advisory when `created` ≥ the
   pinned cutoff constant (missing `created` counts as ≥), advisory otherwise; era/terminal
   demotion unchanged; only `lint.skip` suppresses; zero gated findings at landing.
4. **No new surfaces + E2E** (REQ-VTA-012, 013): no workflow/job/step/gate/dependency, baseline
   file untouched; the CI `spec lint --baseline` invocation exits non-zero naming the code on a
   violating fixture and 0 on its one-line-fixed twin.

## Acceptance criteria (14)

AC-VTA-001 run-pattern two arms · 002 outcome-assertion two arms · 003 markdown contexts incl.
nested blockquote and escaped table pipe · 004 D15 regression inputs · 005 artifact scope, line
numbers, shell-expansion limit · 006 cutoff split + pinned constant · 007 lint.skip only, no
marker · 008 registration, warning severity, message content · 009 end-to-end CI command red/green
on committed fixtures · 010 landing gate green, baseline untouched, measured corpus count ·
011 self-conformance · 012 no `.github/`/`go.mod`/baseline diff · 013 four mutation probes ·
014 coverage ≥85%, vet, golangci-lint.

## Files to modify

- `internal/spec/lint_vacuous_assertion.go` (new)
- `internal/spec/lint_vacuous_assertion_test.go` (new)
- `internal/spec/lint.go` (one registration entry)
- `internal/spec/testdata/vacuous_assert_e2e/{red,green}/` (new fixtures)

## Exclusions

- Axis (c): judging whether a pattern matches any existing test (needs Go source; absent tests are
  expected at plan phase).
- Remediating existing corpus findings.
- New CI workflow/job/step, baseline entry, per-line marker, frontmatter field, multi-line command
  reconstruction.
- Other prose-embedded verification command classes.
