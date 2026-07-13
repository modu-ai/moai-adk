# TRUST 5 Principles

TRUST 5 is a mnemonic for five quality dimensions: **T**ested, **R**eadable,
**U**nified, **S**ecured, **T**rackable. These are assessment questions to ask
of any change, not a library or class to instantiate. MoAI enforces them
through agents, the harness, and gate commands.

## Overview

| Principle | Core question | Primary enforcement |
|-----------|--------------|---------------------|
| Tested | Does the change have tests, and are they green? | manager-develop (cycle_type), `/moai gate` |
| Readable | Is the code clear to a new contributor? | Code review, `/moai review` |
| Unified | Does it match existing file conventions? | Formatter + linter via `/moai gate` |
| Secured | Are inputs validated and credentials protected? | OWASP checklist, `/moai review` |
| Trackable | Is the commit traceable to a requirement? | Conventional Commits, SPEC references |

## Per-Principle Assessment

Apply each principle as a checklist when reviewing a change. Do not skip a
dimension with "not applicable" — if it genuinely does not apply, state why.

### T — Tested

- Does the change include tests for the new or modified behavior?
- Are the tests green (run the test command and show the output)?
- Is coverage at or above the project threshold (85%+ default)?
- For existing untested code touched by the change, are characterization
  tests capturing current behavior before the change?
- For bug fixes, was a failing reproduction test written first and then
  fixed (Reproduction-First Bug Fix)?

Thresholds: package-level 85% minimum, critical packages (cli, template,
hook) 90%+. If coverage is below threshold, the gap is a defect to close,
not an accepted state.

### R — Readable

- Is naming clear and intention-revealing (no abbreviations that need
  decoding)?
- Are comments present where the "why" is non-obvious, in the configured
  code-comments language (English by default)?
- Could a new contributor follow the logic without a walkthrough?
- Is the change the simplest correct implementation (no unnecessary
  abstraction — see Enforce Simplicity)?

### U — Unified

- Does the change match the file's existing naming, error-handling, and
  import conventions? (Consistency within a file beats personal preference.)
- Is the code formatted with the project's formatter (gofmt, black,
  prettier, rustfmt, etc.)?
- Does it pass the project linter (golangci-lint, ruff, eslint, clippy,
  etc.)?
- Are there no global rule suppressions in linter config (per-line
  suppression with a reason is acceptable)?

### S — Secured

- Are all external inputs validated at trust boundaries?
- For web-facing code, does it follow OWASP guidance (injection, XSS, CSRF,
  authn/authz)? See moai-ref-owasp-checklist.
- Are credentials / secrets kept out of version control (environment
  variables or a secrets manager instead)?
- Are dangerous operations (e.g. `rm -rf`, force-push, `DROP TABLE`)
  gated behind explicit confirmation?

### T — Trackable

- Does the commit follow Conventional Commits (`feat`, `fix`, `docs`,
  `refactor`, `test`, `chore`)?
- Does it reference the SPEC or issue it implements (e.g. `feat(SPEC-{DOMAIN}-001):`)?
- Can the change be traced back to a requirement through the commit
  message or linked issue?
- Are @MX annotations present where appropriate (`@MX:NOTE`,
  `@MX:ANCHOR`, `@MX:WARN`, `@MX:TODO`)?

## The "Not Applicable" Guard

Skipping a TRUST 5 dimension with a bare "not applicable" is a red flag.
If a dimension genuinely does not apply to a change, the assessment MUST
state why. Examples of justified skips:

- **Tested** on a docs-only change: "No executable code touched; no tests
  required."
- **Secured** on an internal refactor with no I/O change: "No trust
  boundary or input surface changed."

An unjustified skip is treated as an unverified dimension — it is a gap,
not a pass.

## How sync-auditor Scores

`sync-auditor` provides independent skeptical assessment with 4-dimension
scoring:

| Dimension | What it measures |
|-----------|------------------|
| Functionality | Does the change do what the SPEC requires? (maps to Tested) |
| Security | Is it free of vulnerabilities? (maps to Secured) |
| Craft | Is the code well-structured and readable? (maps to Readable + Unified) |
| Consistency | Does it match conventions and the trackable trail? (maps to Unified + Trackable) |

The overall score is the **harmonic mean** of the four dimensions, not the
average — a single failing dimension drags the score down hard. This
penalizes lopsided quality (e.g. functional but insecure) and reflects that
all four matter. sync-auditor operates as a fresh-judgment auditor: it
treats every claim as suspect until evidence is shown, and rejects when
must-pass criteria fail regardless of nice-to-have scores.

## Relationship to the Harness Level

The harness level determines whether sync-auditor runs at all:

- **minimal** — fast validation only; no sync-auditor scoring.
- **standard** — default checks; no sync-auditor scoring unless invoked.
- **thorough** — full sync-auditor + 4-dimension TRUST 5 scoring.

At minimal/standard, TRUST 5 is enforced by the gate commands and the
manager-develop cycle. At thorough, it is additionally scored by
sync-auditor as an independent check.

## Related

- [Proactive Analysis](proactive-analysis.md) — how gates and reviews surface issues
- [Best Practices](best-practices.md) — documentation-grounded standards validation
- [Integration Patterns](integration-patterns.md) — quality across SPEC phases
