# {{PROJECT_NAME}} Development Guide

> "No spec, no code. No tests, no implementation."

This unified guardrail applies to every agent and developer who uses the MoAI-ADK universal development toolkit. The Python-based toolkit supports all major programming languages and enforces a SPEC-first TDD methodology with @TAG traceability. English is the default working language.

---

## SPEC-First TDD Workflow

### Core Development Loop (3 Steps)

1. **Write the SPEC** (`/alfred:1-plan`) → no code without a spec
2. **Implement with TDD** (`/alfred:2-run`) → no implementation without tests
3. **Sync Documentation** (`/alfred:3-sync`) → no completion without traceability

### On-Demand Support

- **Debugging**: summon `@agent-debug-helper` when failures occur
- **CLI Commands**: init, doctor, status, update, restore, help, version
- **System Diagnostics**: auto-detect language tooling and verify prerequisites

Every change must comply with the @TAG system, SPEC-derived requirements, and language-specific TDD practices.

---

## Skills Index (Progressive Loading)

MoAI-ADK provides **55 Skills** across 6 tiers. Agents invoke Skills explicitly using `Skill("skill-name")` syntax for JIT (Just-in-Time) loading.

### Foundation Tier (Core Principles)

| Skill Name | Purpose | When to Use |
|------------|---------|-------------|
| `moai-foundation-trust` | TRUST 5 principles validation (Test ≥85%, Readable, Unified, Secured, Trackable) | Quality gates, TDD cycle verification, pre-merge checks |
| `moai-foundation-tags` | TAG inventory management (@SPEC/@TEST/@CODE/@DOC chain) | TAG validation, traceability verification, orphan detection |
| `moai-foundation-specs` | SPEC YAML frontmatter validation (7 required fields + HISTORY) | SPEC authoring, metadata validation, versioning |
| `moai-foundation-ears` | EARS requirement authoring (5 patterns: Ubiquitous, Event-driven, State-driven, Optional, Unwanted) | Requirements writing, SPEC creation |
| `moai-foundation-langs` | Language detection and tooling map (Python, TypeScript, Go, Rust, etc.) | Project setup, tool selection |

### Essentials Tier (Common Tasks)

| Skill Name | Purpose | When to Use |
|------------|---------|-------------|
| `moai-essentials-debug` | Error analysis, stack trace interpretation, resolution procedures | Debugging failures, test errors |
| `moai-essentials-refactor` | Refactoring patterns, variable roles, code quality | Code improvement, REFACTOR phase |
| `moai-essentials-perf` | Performance optimization strategies | Performance issues, bottleneck analysis |

### Alfred Tier (Workflow Integration)

| Skill Name | Purpose | When to Use |
|------------|---------|-------------|
| `moai-alfred-tag-scanning` | CODE-FIRST TAG inventory scan via `rg` | Full TAG validation, orphan detection |
| `moai-alfred-spec-metadata-validation` | SPEC metadata and HISTORY enforcement | SPEC quality checks |
| `moai-alfred-trust-validation` | TRUST workflow integration for Alfred commands | `/alfred:3-sync`, quality gates |
| `moai-alfred-ears-authoring` | EARS-based SPEC writing workflow | `/alfred:1-plan`, requirement authoring |
| `moai-alfred-interactive-questions` | TUI-based user interaction patterns | User decision collection |
| `moai-alfred-language-detection` | Project language detection for Alfred | `/alfred:0-project`, tooling setup |

### Domain Tier (Specialized Knowledge)

| Skill Name | Purpose | When to Use |
|------------|---------|-------------|
| `moai-domain-backend` | Backend architecture patterns (REST, GraphQL, microservices) | Backend SPEC, API design |
| `moai-domain-frontend` | Frontend patterns (React, Vue, Angular) | Frontend SPEC, UI design |
| `moai-domain-database` | Database design, ORMs, migrations | Data modeling, schema design |
| `moai-domain-security` | Security patterns, SAST, secret management | Security requirements, threat modeling |
| `moai-domain-ml` | ML workflows, model training, evaluation | ML projects, data pipelines |
| `moai-domain-mobile-app` | Mobile app patterns (iOS, Android, Flutter) | Mobile SPEC, app architecture |
| `moai-domain-cli-tool` | CLI design patterns, argument parsing | CLI tool SPEC |

### Language Tier (Language-Specific Guidance)

| Skill Name | Purpose | When to Use |
|------------|---------|-------------|
| `moai-lang-python` | Python TDD (pytest, mypy, ruff) | Python projects |
| `moai-lang-typescript` | TypeScript TDD (Vitest, Biome, tsc) | TypeScript projects |
| `moai-lang-go` | Go TDD (go test, golangci-lint) | Go projects |
| `moai-lang-rust` | Rust TDD (cargo test, clippy) | Rust projects |
| `moai-lang-ruby` | Ruby TDD (RSpec, RuboCop) | Ruby projects |
| `moai-lang-kotlin` | Kotlin TDD (JUnit, ktlint) | Kotlin projects |
| `moai-lang-php` | PHP TDD (PHPUnit, PHPStan) | PHP projects |
| `moai-lang-r` | R TDD (testthat, lintr) | R projects |
| `moai-lang-shell` | Shell scripting patterns (ShellCheck) | Bash/shell scripts |

### Claude Code Tier (Claude Code Integration)

| Skill Name | Purpose | When to Use |
|------------|---------|-------------|
| `moai-cc-hooks` | Hook architecture and implementation | Hook development |
| `moai-cc-agents` | Agent template structure | Agent creation |
| `moai-cc-commands` | Command orchestration patterns | Command design |
| `moai-cc-skills` | Skill authoring guidelines | Skill creation |
| `moai-cc-memory` | Memory strategy and session management | Context management |
| `moai-cc-settings` | Claude Code configuration | Settings optimization |
| `moai-cc-claude-md` | CLAUDE.md structure and best practices | Project documentation |
| `moai-cc-mcp-plugins` | MCP plugin integration | External tool integration |

### Ops Tier (CI/CD & Automation)

Covers GitHub Actions workflows, Docker, deployment strategies, and infrastructure automation.

---

## Context Engineering Strategy

MoAI-ADK follows Anthropic's "Effective Context Engineering for AI Agents" principles to keep context lean and relevant.

### JIT (Just-in-Time) Loading Rules

**Commands load documents progressively**:

| Command | Required Load | Optional Load | Timing |
|---------|---------------|---------------|--------|
| `/alfred:1-plan` | product.md | structure.md, tech.md | During SPEC discovery |
| `/alfred:2-run` | SPEC-XXX/spec.md | (Skills on-demand) | At TDD start |
| `/alfred:3-sync` | sync-report.md | (TAG scan via `rg`) | During doc sync |

**Skills load on-demand**:
- Agents invoke Skills explicitly: `Skill("moai-foundation-trust")`
- Skills are <500 words, loaded only when needed
- No preloading of all Skills

**Five core documents always loaded** (per CLAUDE.md):
- CLAUDE.md (project guidance)
- product.md (product vision)
- config.json (project settings)
- spec-metadata.md (SPEC metadata schema)
- development-guide.md (this file - Skills Index)

### Context Budget Optimization

**Priority Levels**:
1. **CRITICAL**: config.json, SPEC files, current code being modified
2. **HIGH**: TRUST validation, TAG chain, test results
3. **MEDIUM**: Architecture docs, related modules
4. **LOW**: Historical context, deprecated code

**Loading Strategy**:
- Load CRITICAL immediately
- Load HIGH on first reference
- Load MEDIUM on explicit request
- Defer LOW until needed

---

## TRUST Principles (5 Pillars)

**For complete details**: `Skill("moai-foundation-trust")`

### T – Test-Driven Development (SPEC-Aligned)

- **SPEC → Test → Code Cycle**: @SPEC:ID → @TEST:ID (RED) → @CODE:ID (GREEN + REFACTOR)
- **Coverage Target**: ≥85% line coverage, ≥80% branch coverage
- **Language-Specific Tooling**: pytest (Python), Vitest (TypeScript), go test (Go), cargo test (Rust), RSpec (Ruby)

### R – Requirement-Driven Readability

- **Code Constraints**: ≤300 LOC per file, ≤50 LOC per function, ≤5 parameters, cyclomatic complexity ≤10
- **Naming**: Functions mirror SPEC terminology, intention-revealing names
- **Linting**: ruff (Python), Biome (TypeScript), golangci-lint (Go), clippy (Rust)

### U – Unified SPEC Architecture

- **SPEC-Driven Complexity**: Each SPEC defines complexity threshold
- **Language Boundaries**: SPECs define module/package boundaries (not language conventions)
- **Type Safety**: TypeScript strict mode, mypy strict, Go vet, cargo check

### S – SPEC-Compliant Security

- **SAST Tools**: trivy, semgrep, detect-secrets, bandit (Python), gosec (Go)
- **Security Requirements**: No hardcoded secrets, input validation, proper error handling
- **Dependency Scans**: npm audit, cargo audit, trivy fs

### T – SPEC Traceability

- **TAG Chain**: @SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
- **Code-First Principle**: Source code is the source of truth for TAGs
- **Validation**: `rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/`

---

## @TAG System

**For complete details**: `Skill("moai-foundation-tags")`

### Core Chain

```text
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**Perfect TDD Alignment**:
- `@SPEC:ID` (Preparation) – requirements with EARS format
- `@TEST:ID` (RED) – failing tests
- `@CODE:ID` (GREEN + REFACTOR) – implementation
- `@DOC:ID` (Documentation) – live docs

### TAG Block Template

**SPEC Metadata** (see `SPEC-METADATA.md`):
- **Required Fields (7)**: id, version, status, created, updated, author, priority
- **Optional Fields (9)**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **HISTORY Section**: Mandatory version change log

**Quick Example**:
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @Goos
priority: high
---

# @SPEC:AUTH-001: JWT Authentication System

## HISTORY
### v0.0.1 (2025-09-15)
- **INITIAL**: Authored JWT authentication SPEC
```

**Source Code (`src/`)**:
```text
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

**Test Code (`tests/`)**:
```text
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

### TAG Usage Rules

- **TAG ID Format**: `<DOMAIN>-<3 digits>` (e.g., `AUTH-003`) – immutable
- **Directory Naming**: `.moai/specs/SPEC-{ID}/` (required)
  - ✅ Valid: `SPEC-AUTH-001/`, `SPEC-REFACTOR-001/`
  - ❌ Invalid: `AUTH-001/`, `SPEC-001-auth/`
- **Duplicate Check**: `rg "@SPEC:{ID}" -n .moai/specs/` before creating TAG
- **TAG Validation**: `rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/`

### HISTORY Authoring Guide

**Change Type Tags**:
- `INITIAL`: First release (v1.0.0)
- `ADDED`: New requirement → increment MINOR
- `CHANGED`: Adjusted behavior → increment PATCH
- `FIXED`: Bug fix → increment PATCH
- `REMOVED`: Removed capability → increment MAJOR
- `BREAKING`: Backward-incompatible → increment MAJOR
- `DEPRECATED`: Marked for future removal

---

## EARS Requirement Authoring

**For complete details**: `Skill("moai-foundation-ears")`

### Five EARS Patterns

1. **Ubiquitous Requirements**: The system shall provide [capability].
2. **Event-driven Requirements**: WHEN [condition], the system shall [behavior].
3. **State-driven Requirements**: WHILE [state], the system shall [behavior].
4. **Optional Features**: WHERE [condition], the system may [behavior].
5. **Unwanted Behaviors**: IF [condition], the system shall enforce [constraint].

**Example**:
```markdown
### Ubiquitous Requirements
- The system shall provide user authentication.

### Event-driven Requirements
- WHEN a user logs in with valid credentials, the system shall issue a JWT token.

### State-driven Requirements
- WHILE the user remains authenticated, the system shall allow access to protected resources.

### Optional Features
- WHERE a refresh token is present, the system may issue a new access token.

### Unwanted Behaviors
- IF an invalid token is supplied, the system shall deny access.
```

---

## Language Tooling Map

**For complete details**: `Skill("moai-foundation-langs")`

Quick reference:

| Language | Testing | Linting | Formatting | Type Checking |
|----------|---------|---------|------------|---------------|
| Python | pytest | ruff | ruff | mypy |
| TypeScript | Vitest | Biome | Biome | tsc |
| Go | go test | golangci-lint | gofmt | go vet |
| Rust | cargo test | clippy | rustfmt | cargo check |
| Ruby | RSpec | RuboCop | RuboCop | - |
| Java | JUnit | Checkstyle | - | javac |

---

## Development Principles

### Code Quality Benchmarks

- ≥ 85% test coverage
- Use intention-revealing names
- Prefer guard clauses (early return)
- Leverage language-standard tooling

### Refactoring Rules

- **Rule of Three**: Refactor when pattern appears third time
- **Preparatory Refactoring**: Shape code before change
- **Tidy as You Go**: Fix small issues immediately

**For refactoring patterns**: `Skill("moai-essentials-refactor")`

---

## Exception Handling

When deviating from guidelines, document a waiver and attach to PR/issue/ADR.

**Waiver Checklist**:
- Justification and evaluated alternatives
- Risks and mitigation plan
- Temporary vs. permanent status
- Expiry conditions and approver

---

This guide provides the Skills Index and core workflow overview. For detailed guidance on any topic, invoke the appropriate Skill explicitly.
