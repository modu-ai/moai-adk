# Quality Mechanism Reference

The authoritative mapping for MoAI's quality model: harness levels,
language toolchains, agent roles, and the sync-auditor scoring model. MoAI
enforces quality through agents and gate commands — there is no quality
library to import.

## Quality Enforcement Components

| Component | Role | Invocation |
|-----------|------|------------|
| `/moai gate` | Pre-commit quality gate (lint + format + type + test) | User / orchestrator |
| `/moai review` | Security + design + @MX-tag review | User / orchestrator |
| `/moai fix` | Auto-fix LSP/lint/type errors | User / orchestrator |
| `/moai loop` | Iterative fix until resolved or max iterations | User / orchestrator |
| `manager-develop` | Run-phase implementation (cycle_type) | Agent spawn |
| `sync-auditor` | Independent 4-dimension quality scoring | Agent spawn (thorough) |

## Harness Levels

Auto-determined by the Complexity Estimator based on SPEC scope.

| Level | What runs | sync-auditor | Latency | Use when |
|-------|-----------|--------------|---------|----------|
| minimal | lint + type + test | No | Fastest | Small SPECs, low risk |
| standard | lint + format + type + test | No (unless invoked) | Default | Most SPECs |
| thorough | standard + `/moai review` + full TRUST 5 | Yes — 4-dimension scoring | Slowest | Large SPECs, high risk, security-sensitive |

Override: the harness level can be forced per-SPEC when the auto-detection
misjudges scope. See `.moai/config/sections/harness.yaml` and
`.moai/config/evaluator-profiles/` for configuration.

## Language-Aware Toolchains

`/moai gate` auto-detects the project language and runs the appropriate
toolchain. Tools that are not installed are skipped gracefully. Projects
with no recognized language marker pass the gate silently. The 16 supported
languages are treated equally — no language is "primary".

| Language | Lint | Format | Type check | Test |
|----------|------|--------|-----------|------|
| Go | go vet, golangci-lint | gofmt | (built-in) | go test |
| Python | ruff | black | mypy | pytest |
| TypeScript | eslint | prettier | tsc | jest |
| JavaScript | eslint | prettier | (none / tsc) | jest, mocha |
| Rust | cargo clippy | rustfmt | (built-in) | cargo test |
| Java | (checkstyle / spotbugs) | (google-java-format) | javac | junit |
| Kotlin | (detekt) | ktlint | (built-in) | junit |
| C# | (analyzers) | (dotnet format) | (built-in) | (xunit / nunit) |
| Ruby | rubocop | rubocop | (none) | rspec |
| PHP | phpstan, phpcs | php-cs-fixer | psalm | pest, phpunit |
| Elixir | (credo) | mix format | (dialyzer via erlang) | mix test |
| C++ | (clang-tidy) | clang-format | (built-in) | (gtest / catch2) |
| Scala | (scalafix) | scalafmt | (built-in) | (scalatest) |
| R | (lintr) | (styler) | (none) | testthat |
| flutter (Dart) | (dartanalyzer) | dart format | (built-in) | flutter test |
| Swift | (swiftlint) | swift-format | (built-in) | XCTest |

Parenthesized entries indicate the common community tool; the gate uses
whichever is configured in the project. Detection is by project markers
(go.mod, package.json, requirements.txt, Cargo.toml, etc.).

## Agent Roles in Quality

From the CLAUDE.md §4 retained-agent catalog (10 agents):

| Agent | Quality responsibility |
|-------|----------------------|
| `manager-develop` | Run-phase implementation; owns Tested + Unified via cycle_type; must hit zero lint/type errors and coverage threshold |
| `manager-docs` | Sync-phase; ensures lint clean (≤10 warnings) and docs/CHANGELOG updated |
| `sync-auditor` | Independent 4-dimension scoring (Functionality / Security / Craft / Consistency); fresh-judgment, harmonic mean |
| `plan-auditor` | Independent plan-phase audit (bias prevention, GEARS compliance) |
| `Explore` (built-in) | Read-only codebase exploration before quality assessment |

Archived agents that MUST NOT be spawned for quality work: `manager-quality`,
`expert-security`, `expert-backend`, `expert-frontend`, `expert-performance`,
`expert-refactoring`. When a paste-ready message references one, consult the
archived-agent-rejection rule for the retained-agent replacement.

## sync-auditor Scoring Model

`sync-auditor` scores a change on four dimensions as a fresh-judgment
auditor. The stance (from agent-common-protocol §Skeptical Evaluation
Stance): treat every claim as suspect until evidence is shown; demand
reproducible verification; consider the null hypothesis; score as the
harmonic mean of dimensions, not the average; reject when must-pass
criteria fail regardless of nice-to-have scores.

| Dimension | Maps to TRUST 5 | What it measures |
|-----------|-----------------|------------------|
| Functionality | Tested | Does the change do what the SPEC requires? |
| Security | Secured | Is it free of vulnerabilities? Inputs validated? |
| Craft | Readable + Unified | Is it well-structured, readable, consistent? |
| Consistency | Unified + Trackable | Does it match conventions and the traceable trail? |

**Harmonic mean, not average**: the harmonic mean penalizes lopsided
quality. A change scoring (0.9, 0.9, 0.9, 0.3) averages 0.75 but harmonics
much lower, correctly signaling that the failing dimension is a blocker.
This reflects that all four dimensions matter — you cannot trade security
for functionality.

## LSP Quality Gate Thresholds (per phase)

| Phase | Requirement |
|-------|-------------|
| plan | Capture LSP baseline |
| run | Zero errors, zero type-errors, zero lint-errors |
| sync | Zero errors, max 10 warnings, clean LSP |

These are enforced by the harness during the SPEC lifecycle. A run-phase
that leaves lint errors does not pass the run gate.

## Quality and the SPEC Lifecycle

The 3-phase close (plan → run → sync) embeds quality at each boundary:

- **plan** — LSP baseline captured; acceptance criteria defined as
  machine-verifiable where possible.
- **run** — manager-develop hits the run gate (zero errors/type/lint;
  coverage threshold; tests green).
- **sync** — manager-docs hits the sync gate (≤10 warnings; docs updated);
  sync-auditor scores independently at the thorough level.

MX Tag validation is a cross-cutting concern validated during sync, not a
separate phase. Tag types: `@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`,
`@MX:TODO`, `@MX:DEBT`.

## Best-Practices Documentation Lookup

For best-practices validation against live framework documentation, use
WebSearch / WebFetch against the official documentation:

1. Search: `WebSearch("<framework> official documentation <topic>")` to find the authoritative docs URL.
2. Fetch: `WebFetch(<url>)` to read the relevant section.

Confirm the URL at validation time — do not hardcode docs links. If a source
is unreachable, fall back to established best-practice patterns (see
agent-common-protocol §MCP Fallback Strategy).

## Coverage Thresholds

| Scope | Threshold |
|-------|-----------|
| Package-level (default) | 85% |
| Critical packages (cli, template, hook) | 90%+ |

A coverage figure is only valid when the coverage command was actually run
and its output observed in the current run — not carried over from a prior
unrelated measurement (verification-claim-integrity §2 Baseline-Integrity
Attribution).

## Cross-References

- CLAUDE.md §6 (Quality Gates) — the harness + LSP gate policy
- `.claude/rules/moai/core/moai-constitution.md` § Quality Gates — TRUST 5 principles
- `.claude/rules/moai/core/agent-common-protocol.md` § Skeptical Evaluation Stance — the auditor stance sync-auditor adopts
- `.claude/rules/moai/core/verification-claim-integrity.md` — no unobserved verification claims
- moai-ref-testing-pyramid — test-pyramid strategy and coverage targets
- moai-ref-owasp-checklist — OWASP Top 10 for the Secured principle
