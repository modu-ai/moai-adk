---
name: moai-foundation-quality
description: >
  TRUST 5 quality principles and how MoAI enforces them through agents,
  the 3-level harness, /moai gate, and sync-auditor scoring. Use for code
  review, quality gate checks, coverage targets, or TRUST 5 compliance.

when_to_use: >
  Use for code-quality guidance: TRUST 5 principles (Tested, Readable,
  Unified, Secured, Trackable), the 3-level harness (minimal/standard/
  thorough), /moai gate (lint+format+type+test), coverage targets,
  security checks, language-aware toolchains, code-smell detection, and
  technical-debt triage.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "3.0.0"
  category: "foundation"
  status: "active"
  updated: "2026-07-10"
  modularized: "true"
  tags: "foundation, quality, testing, validation, trust-5, best-practices, code-review"
  aliases: "moai-foundation-quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# TRUST 5 Quality Principles and Enforcement

This skill provides background knowledge on MoAI's quality model: the five
TRUST 5 principles, how agents enforce them, the 3-level harness, and the
language-aware toolchains that `/moai gate` runs. MoAI does NOT ship a
quality-validation library — quality is enforced through agents
(`manager-develop`, `sync-auditor`), slash commands (`/moai gate`,
`/moai review`), and the harness (minimal/standard/thorough).

## Quick Reference

**TRUST 5 Principles** (Tested, Readable, Unified, Secured, Trackable) are
quality dimensions, not code objects. Every code change is evaluated against
all five.

**Quality Mechanisms** (the real enforcement layer):

- `/moai gate` — runs lint + format + type-check + test in parallel as a
  pre-commit quality gate (<30s). Auto-detects the project language and runs
  the appropriate toolchain.
- `manager-develop` (run-phase) — implements via `cycle_type` ∈ {tdd, ddd,
  autofix}; the chosen cycle shapes how tests and behavior are produced.
- `sync-auditor` — independent skeptical quality assessment with 4-dimension
  scoring (Functionality, Security, Craft, Consistency), scored as the
  harmonic mean of dimensions, not the average.
- 3-level harness — minimal (fast validation), standard (default checks),
  thorough (full sync-auditor + TRUST 5). Auto-determined by the Complexity
  Estimator based on SPEC scope.
- LSP quality gates — phase-specific thresholds (run: zero errors/type-errors/
  lint-errors; sync: zero errors, max 10 warnings, clean LSP).

## The MoAI Quality Model

MoAI does not provide a Python SDK or any library for quality validation.
Quality is enforced through the workflow, the agents, and the gate commands.
This skill documents how those pieces fit together so a Claude invocation
can reason about quality correctly.

### How TRUST 5 is enforced per phase

| Phase | Quality check | Owner |
|-------|--------------|-------|
| plan | Capture LSP baseline; identify quality risks in the plan | manager-spec |
| run | Zero errors/type-errors/lint-errors; tests pass; coverage met | manager-develop (cycle_type shapes the approach) |
| sync | Lint clean (≤10 warnings); docs updated; TRUST 5 re-affirmed | manager-docs, then sync-auditor scores |
| audit | Independent 4-dimension scoring (Functionality/Security/Craft/Consistency) | sync-auditor |

### cycle_type and quality (manager-develop)

The run-phase `cycle_type` selects how quality is built in:

- **tdd** — Test-Driven Development (RED-GREEN-REFACTOR). Behavior is
  specified by a failing test first, then implemented. Best for new features.
- **ddd** — Domain-Driven refactoring (ANALYZE-PRESERVE-IMPROVE).
  Behavior-preserving transformation of existing code. Best for refactoring
  and debt reduction.
- **autofix** — diagnostic-driven fixing (LSP / lint / type errors). Best for
  `/moai fix` and regression recovery.

See Skill("moai-workflow-tdd"), Skill("moai-workflow-ddd"), and
Skill("moai-workflow-loop") for the per-cycle mechanics.

## TRUST 5 Principles

TRUST 5 is a mnemonic for five quality dimensions. Treat each as a question
to ask of any change, not a score to compute.

- **T — Tested**: Does the change have tests? Are they green? Is coverage at
  or above the project threshold (85%+ by default)? For existing untested
  code, are characterization tests capturing current behavior?
- **R — Readable**: Is naming clear? Are comments in English (or the
  configured code-comments language)? Could a new contributor follow the
  logic without a walkthrough?
- **U — Unified**: Does the change match the file's existing conventions
  (naming, error handling, imports)? Is it formatted with the project's
  formatter? Consistency within a file beats personal preference.
- **S — Secured**: Are all external inputs validated? Does it follow OWASP
  guidance for web security? Are credentials kept out of version control
  (environment variables instead)? See moai-ref-owasp-checklist.
- **T — Trackable**: Does the commit follow Conventional Commits? Does it
  reference the SPEC / issue it implements? Can the change be traced back to
  a requirement?

For the per-principle assessment checklist and the "not applicable" guard,
see [TRUST 5 Principles](modules/trust5-validation.md).

## Quality Gates and the 3-Level Harness

The harness level controls how deep quality validation goes. It is
auto-determined by the Complexity Estimator based on SPEC scope.

| Level | What runs | When |
|-------|-----------|------|
| minimal | Fast validation only (lint + type + test) | Small SPECs, low risk |
| standard | Default checks (lint + type + test + format) | Most SPECs |
| thorough | Full sync-auditor + 4-dimension TRUST 5 scoring | Large SPECs, high risk |

`/moai gate` is the lightweight pre-commit entry point: it runs lint +
format + type-check + test in parallel and applies no fixes. It is the
fastest way to get a quality signal. For deeper review use `/moai review`.

## Language-Aware Toolchains

The quality gate auto-detects the project language and runs the appropriate
toolchain. Tools that are not installed are skipped gracefully; projects
with no recognized language marker pass the gate silently. This skill is
language-neutral — the 16 supported languages are treated equally.

| Language | Lint | Format | Test |
|----------|------|--------|------|
| Go | go vet → golangci-lint | gofmt | go test |
| Python | ruff | black | pytest |
| TypeScript / JavaScript | eslint | prettier | jest / mocha |
| Rust | cargo clippy | rustfmt | cargo test |
| Java / Kotlin | (per project linter) | (per project) | junit |
| Ruby | rubocop | rubocop | rspec |
| PHP | phpstan / phpcs | php-cs-fixer | pest / phpunit |
| ... | (16 languages supported; auto-detected) | | |

For the full toolchain mapping and how `/moai gate` detects the language,
see [Language-Aware Toolchains](references/reference.md#language-aware-toolkchains).

## Module Reference

Each module is loaded on demand. Load the one relevant to the current task.

- [TRUST 5 Principles](modules/trust5-validation.md) — the five dimensions as
  assessment questions, per-principle checklists, and the "not applicable"
  guard.
- [Proactive Analysis](modules/proactive-analysis.md) — how `/moai gate`,
  `/moai review`, and `/moai loop` surface quality issues proactively, and
  how to triage findings.
- [Best Practices](modules/best-practices.md) — using Context7 MCP for
  up-to-date framework/library best practices, and validating against them.
- [Integration Patterns](modules/integration-patterns.md) — how quality fits
  into the SPEC workflow phases (plan/run/sync) and the harness levels.

## Reference Files

- [examples.md](references/examples.md) — worked TRUST 5 assessment examples
  and gate/review triage walkthroughs. Load when applying TRUST 5 to a
  concrete change.
- [reference.md](references/reference.md) — the quality-mechanism reference:
  harness level detail, language toolchain table, agent roles, and the
  sync-auditor scoring model. Load when you need the authoritative mapping.

## Works Well With

Agents (see CLAUDE.md §4 for the 10-agent catalog):

- `manager-develop` — run-phase implementation; owns the Tested and Unified
  principles through cycle_type.
- `sync-auditor` — independent 4-dimension quality scoring (Functionality /
  Security / Craft / Consistency).
- `Explore` (Anthropic built-in) — read-only codebase exploration before
  assessing quality.

Skills:

- `moai-foundation-core` — TRUST 5 framework cross-reference and SPEC
  workflow foundations.
- `moai-ref-testing-pyramid` — test-pyramid strategy, coverage targets, and
  test patterns.
- `moai-ref-owasp-checklist` — OWASP Top 10 security checklist for the
  Secured principle.
- `moai-workflow-tdd` / `moai-workflow-ddd` / `moai-workflow-loop` — the
  cycle_type workflows that manager-develop uses.

Commands:

- `/moai gate` — pre-commit quality gate (lint + format + type + test).
- `/moai review` — code review with security and MX-tag compliance.
- `/moai fix` — auto-detect and fix LSP/lint/type errors.
- `/moai loop` — iterative fix loop until resolved or max iterations.

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The linter warnings are false positives" | False positives should be suppressed with inline comments. Ignoring them trains the team to ignore real issues. |
| "Security scanning can wait until before release" | Security vulnerabilities compound. Late discovery means expensive rework. Scan continuously. |
| "Coverage is high enough, the remaining 15% is edge cases" | Edge cases are where production bugs live. The uncovered code is the riskiest code. |
| "Code review is subjective, automation is sufficient" | Automation catches syntax and patterns. Reviews catch design flaws, naming confusion, and missing abstractions. |
| "TRUST 5 is too bureaucratic for a hotfix" | Hotfixes without quality gates introduce the next hotfix. TRUST 5 on a hotfix is the minimum, not the maximum. |

**Chesterton's Fence**: Before removing a quality check, understand why it was added. Removing a gate without understanding its history repeats the failure it was designed to prevent.

**Shift Left**: The earlier a defect is found, the cheaper it is to fix. Quality checks belong in the development loop, not at the end of it.

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Linter or type-checker warnings suppressed globally instead of per-line
- OWASP checklist not consulted when handling user input or authentication
- Coverage report not generated for a commit that adds new functionality
- TRUST 5 dimension skipped with "not applicable" without justification
- Quality report generated but no action taken on identified issues

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Linter runs clean or remaining warnings have inline suppression comments with reasons (show the command output)
- [ ] OWASP checklist reviewed for security-relevant changes (show checklist references)
- [ ] Coverage report generated and threshold met (show tool output)
- [ ] All five TRUST 5 dimensions assessed (show assessment for each)
- [ ] Quality findings triaged with a resolution plan for each
- [ ] No global rule disabling in linter configuration

<!-- moai:evolvable-end -->
