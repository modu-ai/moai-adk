# Best Practices Validation

How to validate code against up-to-date framework and library best practices.
MoAI does not ship a "best practices engine"; the mechanism is: look up the
latest official documentation via WebSearch / WebFetch, then compare the code
against what those docs prescribe.

## Overview

Best-practices validation means checking that code follows current standards
for the language and frameworks in use. Because best practices drift (a
framework's recommended pattern changes across versions), this validation
relies on live documentation rather than a fixed rule set.

The tools:
- `WebSearch` — find candidate official-documentation sources for a framework or library.
- `WebFetch` — fetch and verify the official documentation at a specific URL.

## Validation Process

To validate a change against best practices:

1. **Identify the relevant libraries** in the change (frameworks, key
   dependencies). Read the dependency manifest (go.mod, package.json,
   requirements.txt / pyproject.toml, Cargo.toml, pom.xml, etc.).
2. **Locate the official documentation** for each library via `WebSearch`
   (target the official docs site, e.g. `"<framework> official documentation"`).
3. **Fetch the relevant docs**, scoped to the topic at hand (e.g.
   "best-practices", "error-handling", "testing", "performance"), via
   `WebFetch` against the official URLs surfaced by the search.
4. **Compare** the code against the documented patterns. Note where the code
   diverges and whether the divergence is justified.
5. **Report** the findings: what matches, what diverges, and the
   recommendation.

If a documentation source cannot be reached, fall back to established
best-practice patterns from industry experience. Architecture/analysis quality
must not depend on a single source being available (see agent-common-protocol
§MCP Fallback Strategy).

## Language-Neutral Validation

This skill is language-neutral. The 16 supported languages are treated
equally — there is no "primary" language. The validation process above
applies to any of them. Examples of what to validate per language family:

| Concern | Example check |
|---------|--------------|
| Error handling | Does it use the language's idiomatic error pattern (Go's error returns, Rust's Result, exceptions, etc.)? |
| Naming | Does it follow the language's convention (camelCase vs snake_case vs PascalCase)? |
| Concurrency | Does it use the safe primitive (mutex, channel, async/await, actor)? |
| Testing | Does it use the language's test framework idiomatically? |
| Dependency management | Are dependencies pinned and minimal? |

## Documentation Source Mapping (illustrative)

Official documentation lives at well-known upstream URLs. Resolve via
`WebSearch` rather than hardcoding — the URL may change. Illustrative home
pages for common quality tooling:

| Tool | Official docs home |
|------|-------------------|
| eslint | https://eslint.org/docs/latest |
| prettier | https://prettier.io/docs |
| ruff | https://docs.astral.sh/ruff |
| golangci-lint | https://golangci-lint.run |
| clippy (rust) | https://doc.rust-lang.org/clippy |
| jest | https://jestjs.io/docs |
| pytest | https://docs.pytest.org |

Do not treat this table as authoritative — always confirm the current URL via
`WebSearch` / `WebFetch` at validation time.

## Custom Quality Checks

For project-specific quality rules that go beyond the standard toolchain,
encode them as:

- **Linter custom rules** — most linters support custom rules (golangci-lint
  custom analyzers, eslint custom rules, clippy lints). This is the
  preferred mechanism because the existing gate will run them.
- **@MX tags** — use `@MX:ANCHOR` to mark invariants the code must preserve,
  and `@MX:WARN` to mark danger zones. These surface during review.
- **A SPEC** — for a structural quality requirement, author a SPEC so it
  enters the plan/run/sync lifecycle and gets audited.

Do not invent a "custom rule engine" library. Use the project's existing
extensibility surface.

## What NOT to Do

- Do NOT treat one language's conventions as universal (Python's `black`
  is not relevant to a Go project; Go's `gofmt` is not relevant to a Rust
  project).
- Do NOT present cached best-practice rules as "current" without re-checking
  the official docs via `WebSearch` / `WebFetch` — best practices drift.

## Related

- [TRUST 5 Principles](trust5-validation.md) — the Unified principle this feeds
- [Proactive Analysis](proactive-analysis.md) — gate and review triage
- [Integration Patterns](integration-patterns.md) — quality across SPEC phases
