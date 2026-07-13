# Integration Patterns

How quality fits into the MoAI SPEC workflow and the harness levels. MoAI
does not ship a CI/CD integration library or a "Quality-as-a-Service" API;
quality is integrated by routing work through the SPEC phases and running
the gate commands at the right points.

## Quality Across the SPEC Phases

The SPEC workflow has three phases (plan → run → sync), with an audit
layer. Quality is enforced at each boundary.

### Plan phase (manager-spec)

- Capture the LSP baseline so the run phase can measure against it.
- Identify quality risks in the plan: untested code that will be touched,
  security-sensitive surfaces, complexity hotspots.
- Define acceptance criteria that are machine-verifiable where possible
  (test suite green, lint clean, coverage threshold met).

### Run phase (manager-develop)

- Zero errors / type-errors / lint-errors required.
- Tests must pass; coverage must meet the threshold.
- The `cycle_type` shapes the approach:
  - **tdd** — failing test first, then implementation (RED-GREEN-REFACTOR).
  - **ddd** — behavior-preserving transformation (ANALYZE-PRESERVE-IMPROVE).
  - **autofix** — diagnostic-driven fixing.
- Run `/moai gate` before committing to get a fast quality signal.

### Sync phase (manager-docs → sync-auditor)

- Lint clean (≤10 warnings), docs updated, CHANGELOG entry present.
- TRUST 5 re-affirmed across the change.
- sync-auditor provides independent 4-dimension scoring
  (Functionality / Security / Craft / Consistency) as a fresh-judgment
  audit.

## Harness Levels and When Each Runs

The harness level controls quality depth and is auto-determined by the
Complexity Estimator based on SPEC scope.

| Level | Gate | sync-auditor | Typical use |
|-------|------|--------------|-------------|
| minimal | `/moai gate` (lint + type + test) | No | Small SPECs, low risk, fast iteration |
| standard | `/moai gate` + format | No (unless invoked) | Most SPECs |
| thorough | `/moai gate` + `/moai review` | Yes — full 4-dimension scoring | Large SPECs, high risk, security-sensitive |

## CI/CD Integration Pattern

In a CI pipeline, the quality gate maps to running the project's toolchain
on every push/PR. The pattern (language-neutral — substitute the project's
commands):

1. **Lint** — run the project linter; fail on errors.
2. **Format check** — run the formatter in check mode; fail if not
   formatted.
3. **Type check** — run the type checker; fail on errors.
4. **Test** — run the test suite; fail on any failure.
5. **Coverage** — generate coverage; fail if below threshold.

`/moai gate` runs steps 1-4 in parallel locally. In CI, replicate the same
checks. Do not invent a separate quality pipeline that duplicates the gate —
use the same toolchain so local and CI signals agree.

## Review Integration Pattern

For non-trivial changes, `/moai review` goes beyond the gate:

- Security review against the OWASP checklist (Secured principle).
- @MX-tag compliance check (are invariants annotated? are danger zones
  marked?).
- Design-level review (readability, abstraction appropriateness).

Use `/moai review` before opening a PR, or when the change is too complex
for the gate alone to judge.

## Iterative Fix Pattern

When the gate or review surfaces a cluster of fixable issues:

1. Run `/moai fix` to auto-fix LSP/lint/type errors.
2. If issues remain, run `/moai loop` to iterate until resolved or the max
   iteration count is reached.
3. For issues that cannot be auto-fixed, either fix directly or defer with
   an @MX:TODO and a stated reason.

`/moai loop` exits with a ceiling-exit verdict when max iterations are
reached without full resolution — surface that verdict rather than
silently stopping.

## Team Mode Quality

In Agent Teams mode (experimental), two hooks enforce quality:

- **TeammateIdle hook** — validates a teammate's work before accepting idle
  state. If LSP errors exceed the threshold, the teammate is kept working.
- **TaskCompleted hook** — validates deliverables before completion.

These are the team-mode equivalents of the gate, applied per-teammate. See
the spec-workflow rule for the full team-quality contract.

## What NOT to Do

- Do NOT invent a CI/CD integration library or a "Quality-as-a-Service"
  REST API. Use the project's existing CI + the gate commands.
- Do NOT run a separate quality pipeline that disagrees with `/moai gate` —
  local and CI must use the same toolchain.
- Do NOT skip sync-auditor at the thorough level — its independent scoring
  is the bias-prevention check.

## Related

- [TRUST 5 Principles](trust5-validation.md) — the dimensions enforced per phase
- [Proactive Analysis](proactive-analysis.md) — gate and review triage
- [Best Practices](best-practices.md) — documentation-grounded standards validation
