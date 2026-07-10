# Best Practices Validation

How to validate code against up-to-date framework and library best practices
using the Context7 MCP tools. MoAI does not ship a "best practices engine";
the mechanism is: query Context7 for the latest docs, then compare the code
against what those docs prescribe.

## Overview

Best-practices validation means checking that code follows current standards
for the language and frameworks in use. Because best practices drift (a
framework's recommended pattern changes across versions), this validation
relies on live documentation rather than a fixed rule set.

The tools:
- `mcp__context7__resolve-library-id` — resolve a framework/library name to
  a Context7 library ID.
- `mcp__context7__get-library-docs` — fetch the latest docs for a resolved
  library, optionally scoped to a topic.

## Validation Process

To validate a change against best practices:

1. **Identify the relevant libraries** in the change (frameworks, key
   dependencies). Read the dependency manifest (go.mod, package.json,
   requirements.txt / pyproject.toml, Cargo.toml, pom.xml, etc.).
2. **Resolve each via Context7**:
   `mcp__context7__resolve-library-id(libraryName: "<framework>")`.
3. **Fetch the relevant docs**, scoped to the topic at hand
   (e.g. "best-practices", "error-handling", "testing",
   "performance"): `mcp__context7__get-library-docs(libraryId, query)`.
4. **Compare** the code against the documented patterns. Note where the code
   diverges and whether the divergence is justified.
5. **Report** the findings: what matches, what diverges, and the
   recommendation.

If Context7 is unavailable, fall back to official documentation via WebFetch
and established best-practice patterns. Architecture/analysis quality must
not depend on MCP availability (see agent-common-protocol §MCP Fallback
Strategy).

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

## Context7 Library Mapping (illustrative)

Context7 library IDs follow the `/org/project` convention. Resolve via the
MCP tool rather than hardcoding — the ID may change. Illustrative mappings
for common quality tooling (resolve before use):

| Tool | Likely Context7 ID |
|------|-------------------|
| eslint | /eslint/eslint |
| prettier | /prettier/prettier |
| ruff | /astral-sh/ruff |
| golangci-lint | /golangci/golangci-lint |
| clippy (rust) | /rust-lang/rust-clippy |
| jest | /jestjs/jest |
| pytest | /pytest-dev/pytest |

Do not treat this table as authoritative — always resolve via the MCP tool
at validation time.

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

- Do NOT hardcode Context7 library IDs — resolve at validation time.
- Do NOT treat one language's conventions as universal (Python's `black`
  is not relevant to a Go project; Go's `gofmt` is not relevant to a Rust
  project).
- Do NOT present cached best-practice rules as "current" without re-checking
  Context7 — best practices drift.

## Related

- [TRUST 5 Principles](trust5-validation.md) — the Unified principle this feeds
- [Proactive Analysis](proactive-analysis.md) — gate and review triage
- [Integration Patterns](integration-patterns.md) — quality across SPEC phases
