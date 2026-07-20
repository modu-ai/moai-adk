# Proactive Quality Analysis

How MoAI surfaces quality issues proactively — through gate commands,
review, and iterative fix loops. MoAI does not ship a scanner library;
proactive analysis is done by running the project's own toolchain via
`/moai gate`, `/moai review`, and `/moai loop`, then triaging the output.

## Overview

Proactive quality analysis covers:

- **Lint / type / format violations** — caught by `/moai gate` running the
  project's toolchain.
- **Security vulnerabilities** — caught by `/moai review` against the OWASP
  checklist (see moai-ref-owasp-checklist).
- **Coverage gaps** — caught by running the coverage command and comparing
  against the threshold.
- **Complexity and duplication** — surfaced during code review and by the
  project's linter (golangci-lint, eslint, clippy, etc.).
- **Technical debt** — identified via @MX:TODO / @MX:DEBT tags and triaged
  during review.

## The Gate Commands

### `/moai gate` — pre-commit quality gate

Runs lint + format + type-check + test in parallel. Applies no fixes.
Auto-detects the project language and runs the appropriate toolchain.
Tools not installed are skipped gracefully.

Use it:
- Before committing, to get a fast quality signal.
- After pulling changes, to catch regressions early.
- As the minimal/standard harness quality check.

### `/moai review` — code review

Reviews the diff with a security and @MX-tag compliance check. Deeper than
`/moai gate` — it reasons about design, not just syntax. Use it before a
PR or when a change is non-trivial.

### `/moai fix` and `/moai loop` — iterative fixing

`/moai fix` auto-detects and fixes LSP/lint/type errors. `/moai loop` runs
the fix iteratively until all issues are resolved or a max iteration count
is reached. These are the proactive remediation tools.

## Triage Process

When a gate or review surfaces findings, triage them:

1. **Classify** each finding: is it a real defect, a false positive, or a
   style preference?
2. **Severity-rank**: critical (security / data loss / crash), major
   (functional bug), minor (style / readability), info.
3. **Act**:
   - Critical/major → fix now (via `/moai fix` or direct edit).
   - Minor → fix if cheap; otherwise suppress per-line with a reason or
     defer with an @MX:TODO.
   - False positive → suppress per-line with a reason; do NOT disable the
     rule globally.
4. **Verify** the fix by re-running the gate and showing the output.

## Coverage Analysis

Coverage is a first-class quality signal. To assess it:

1. Run the project's coverage command (e.g. `go test -cover ./...`,
   `pytest --cov`, `jest --coverage`, `cargo tarpaulin`).
2. Compare against the threshold (85%+ default; 90%+ for critical packages).
3. For any uncovered code touched by the change:
   - New code → add tests (cycle_type=tdd drives this).
   - Existing untested code → add characterization tests capturing current
     behavior before modifying.

A coverage gap is a defect to close, not an accepted state. Do not claim a
coverage figure without running the command and observing the output (see
verification-claim-integrity).

## Technical Debt Identification

Technical debt is surfaced two ways:

- **@MX tags** — `@MX:TODO` marks incomplete work; `@MX:DEBT` with
  `@MX:CEILING` / `@MX:UPGRADE` marks deliberate working simplifications.
  These are the in-code debt markers. Scan for them with Grep.
- **Review findings** — `/moai review` and sync-auditor flag debt during
  assessment.

Debt identified is not debt that must be fixed immediately. Triage it:
record it (via @MX:DEBT or a SPEC), assign a priority, and schedule it.
Acting on inferred debt without running the domain's verification tool is
an unobserved defect claim.

## What NOT to Do

- Do NOT invent a scanner or analysis library. Use the project's own
  toolchain via the gate commands.
- Do NOT claim "coverage is X%" without running the coverage command and
  observing the output in this run.
- Do NOT infer a defect from a grep match alone — run the domain's
  dedicated tool (linter, type-checker, test suite) to confirm.
- Do NOT disable linter rules globally to silence warnings — suppress
  per-line with a stated reason.

## Related

- [TRUST 5 Principles](trust5-validation.md) — the dimensions this analysis feeds
- [Best Practices](best-practices.md) — documentation-grounded standards validation
- [Integration Patterns](integration-patterns.md) — quality across SPEC phases
