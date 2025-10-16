# MoAI-ADK Development Guide

> "No spec, no code. No test, no implementation."

This is the unified guardrail for every agent and developer using the MoAI-ADK general-purpose development toolkit. Built on TypeScript, the toolkit supports all major programming languages and follows a SPEC-first TDD methodology anchored by @TAG traceability. Korean remains the default collaboration language.

---

## SPEC-First TDD Workflow

### Core Development Loop (3 Steps)

1. **SPEC Authoring** (`/alfred:1-spec`) → No spec, no code
2. **TDD Implementation** (`/alfred:2-build`) → No tests, no implementation
3. **Documentation Sync** (`/alfred:3-sync`) → No traceability, no done

### On-Demand Support

- **Debugging**: Invoke `@agent-debug-helper` when errors occur
- **CLI Commands**: init, doctor, status, update, restore, help, version
- **System Diagnostics**: Auto-detect language tooling and verify host prerequisites

All changes comply with the @TAG system, SPEC-driven requirements, and language-specific TDD practices.

### EARS Requirements Authoring

**EARS (Easy Approach to Requirements Syntax)** provides a structured way to capture requirements.

#### Five EARS Patterns
1. **Ubiquitous**: The system SHALL provide [function].
2. **Event-driven**: WHEN [condition], the system SHALL [behavior].
3. **State-driven**: WHILE [state], the system SHALL [behavior].
4. **Optional**: WHERE [condition], the system MAY [behavior].
5. **Constraints**: IF [condition], the system MUST [constraint].

#### Example
```markdown
### Ubiquitous Requirements (Basic)
- The system MUST provide user authentication.

### Event-driven Requirements
- WHEN a user logs in with valid credentials, the system MUST issue a JWT.
- WHEN the token expires, the system MUST return a 401 error.

### State-driven Requirements
- WHILE the user remains authenticated, the system MUST allow access to protected resources.

### Optional Features
- WHERE a refresh token is supplied, the system MAY issue a new access token.

### Constraints
- IF an invalid token is supplied, the system MUST deny access.
- The access token expiration MUST NOT exceed 15 minutes.
```

---

## Context Engineering

MoAI-ADK implements efficient context management based on Anthropic's "Effective Context Engineering for AI Agents."

### 1. JIT (Just-in-Time) Retrieval

**Principle**: Load documents only when they are needed to minimize the initial context footprint.

**Alfred's JIT strategy**:

| Command | Required Load | Optional Load | Load Timing |
|---------|---------------|---------------|-------------|
| `/alfred:1-spec` | product.md | structure.md, tech.md | While exploring SPEC candidates |
| `/alfred:2-build` | SPEC-XXX/spec.md | development-guide.md | When starting TDD implementation |
| `/alfred:3-sync` | sync-report.md | TAG index | During documentation sync |

**Implementation**:
- Alfred loads only the documents needed at command execution time through the `Read` tool.
- Agents request only the documents relevant to their work.
- The five documents listed in the "Memory Strategy" section of CLAUDE.md are always loaded.

### 2. Compaction

**Principle**: Summarize long sessions (>70% token usage) and restart with a fresh conversation.

**Compaction Triggers**:
- Token usage > 140,000 (70% of the 200,000 token limit)
- Conversation exceeds 50 turns
- The user explicitly runs `/clear` or `/new`

**Compaction Procedure**:
1. **Produce a Summary**: Capture key decisions, completed work, and next steps.
2. **Start a New Session**: Use the summary as the opening message.
3. **Provide Guidance**: Recommend that the user run `/clear` or `/new`.

**Example**:
```markdown
**Recommendation**: Before continuing, run `/clear` or `/new` to start a fresh session for better performance and context management.
```

### Context Engineering Checklist

**When designing commands**:
- [ ] JIT: Are only the required documents loaded?
- [ ] Conditional Load: Are optional documents loaded based on context?
- [ ] Compaction: Do long tasks include interim summaries?

**When designing agents**:
- [ ] Minimal tools: Are only the necessary tools declared in the YAML frontmatter?
- [ ] Clear roles: Does each agent follow the single-responsibility principle?

**Managing long sessions**:
- [ ] Monitor token usage
- [ ] Recommend compaction once usage exceeds 70%
- [ ] Include guidance for `/clear` or `/new`

---

## TRUST Principles

### T - Test-Driven Development (SPEC-driven)

**SPEC → Test → Code cycle**:

- **SPEC**: Write detailed requirements annotated with `@SPEC:ID` using the EARS approach.
- **RED**: `@TEST:ID` - create failing tests derived from the SPEC and confirm they fail.
- **GREEN**: `@CODE:ID` - implement the minimal code that satisfies the SPEC and passes the tests.
- **REFACTOR**: `@CODE:ID` - improve code quality while preserving SPEC compliance, and document via `@DOC:ID`.

**Language-specific TDD execution**:

- **Python**: pytest + SPEC-aligned test cases with mypy type hints
- **TypeScript**: Vitest + SPEC-aligned test suites with strict typing
- **Java**: JUnit + SPEC annotations for behavior-driven tests
- **Go**: go test + SPEC table-driven tests enforcing interface contracts
- **Rust**: cargo test + SPEC doc tests validating traits

Each test connects `@TEST:ID` to `@CODE:ID`, linking back to the exact SPEC requirement.

### R - Requirement-Aligned Readability

**Clean code aligned to the SPEC**:

- Functions implement SPEC requirements directly (<= 50 LOC per function)
- Names reflect SPEC terminology and domain language
- Structure mirrors SPEC design decisions
- Comments only document SPEC explanations and @TAG references

**Language-specific SPEC implementation**:

- **Python**: Type hints mirror SPEC interfaces with mypy enforcement
- **TypeScript**: Strict interfaces that match SPEC contracts
- **Java**: Classes implementing SPEC components with strong typing
- **Go**: Interfaces that satisfy SPEC requirements with gofmt formatting
- **Rust**: Types that implement SPEC safety requirements with rustfmt formatting

Every code element remains traceable to the SPEC through @TAG annotations.

### U - Unified SPEC Architecture

- **SPEC-driven complexity management**: Each SPEC defines complexity thresholds. Exceeding them requires either a new SPEC or a justified waiver.
- **SPEC implementation stages**: Separate SPEC authoring from implementation; do not modify SPECs during the TDD loop.
- **Cross-language SPEC alignment**: SPECs define language boundaries (Python modules, TypeScript interfaces, Java packages, Go packages, Rust crates).
- **SPEC-based architecture**: Domain boundaries follow SPEC definitions rather than language conventions, with @TAGs preserving cross-language traceability.

### S - SPEC-Compliant Security

- **SPEC security requirements**: Explicitly document security needs, data sensitivity, and access control in every SPEC.
- **Security by design**: Implement security controls during TDD rather than bolting them on afterwards.
- **Language-agnostic security patterns**:
  - Input validation driven by SPEC interface definitions
  - Audit logging for SPEC-defined critical operations
  - Access control aligned with the SPEC authorization model
  - Secret management that honors SPEC environment requirements

### T - SPEC Traceability

- **SPEC-to-code traceability**: Every code change references the SPEC ID and specific requirement through the @TAG system.
- **Three-step workflow traceability**:
  - `/alfred:1-spec`: Author SPECs with `@SPEC:ID` tags (`.moai/specs/`)
  - `/alfred:2-build`: Execute TDD linking `@TEST:ID` (tests/) to `@CODE:ID` (src/)
  - `/alfred:3-sync`: Synchronize documentation with `@DOC:ID` (docs/) and verify TAG coverage
- **Code scan traceability**: Guarantee TAG fidelity by scanning the code directly with `rg '@(SPEC|TEST|CODE|DOC):' -n`, avoiding intermediate caches.

---

## SPEC-First Mindset

1. **SPEC-led decisions**: Base every technical decision on an existing SPEC or create a new SPEC before implementation. No work without clear requirements.
2. **SPEC context awareness**: Read related SPEC documents, analyse @TAG relationships, and verify compliance before changing code.
3. **SPEC communication**: Korean remains the default language for discussions. Write SPEC documents in clear Korean prose with English technical terms.

## SPEC-TDD Workflow

1. **SPEC first**: Create or reference a SPEC before writing code. Use `/alfred:1-spec` to clarify requirements, design, and tasks.
2. **TDD implementation**: Follow Red-Green-Refactor rigorously. Use `/alfred:2-build` with language-appropriate test frameworks.
3. **Traceability sync**: Run `/alfred:3-sync` to update documentation and maintain @TAG relationships between SPECs and code.

## @TAG System

### Core Chain

```text
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**Perfect TDD alignment**:
- `@SPEC:ID` (Plan) - Requirements authored with the EARS style
- `@TEST:ID` (RED) - Write failing tests first
- `@CODE:ID` (GREEN + REFACTOR) - Implement and refactor while staying within the SPEC
- `@DOC:ID` (Docs) - Maintain a living document

### TAG Block Template

> **📋 SPEC Metadata Standard (SSOT)**: See `spec-metadata.md`.

**Every SPEC document MUST include YAML front matter and a HISTORY section**:
- **Seven required fields**: id, version, status, created, updated, author, priority
- **Nine optional fields**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **HISTORY section**: Record all version changes (mandatory)

**Full template, field details, and validation steps**: See `spec-metadata.md`.

**Quick reference example**:
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
- **INITIAL**: Authored the specification for the JWT-based authentication system.
...
```

**Source code (`src/`)**:
```text
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

**Test code (`tests/`)**:
```text
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

### @CODE Subcategories (comment level)

Annotate implementation details within `@CODE:ID` comments:
- `@CODE:ID:API` - REST APIs, GraphQL endpoints
- `@CODE:ID:UI` - Components, views, screens
- `@CODE:ID:DATA` - Data models, schemas, types
- `@CODE:ID:DOMAIN` - Business logic, domain rules
- `@CODE:ID:INFRA` - Infrastructure, databases, integrations

### TAG Usage Rules

- **TAG ID**: `<domain>-<3 digits>` (e.g., `AUTH-003`) - immutable
- **Directory naming**: `.moai/specs/SPEC-{ID}/` (mandatory)
  - ✅ Correct: `SPEC-AUTH-001/`, `SPEC-REFACTOR-001/`, `SPEC-UPDATE-REFACTOR-001/`
  - ❌ Incorrect: `AUTH-001/`, `SPEC-001-auth/`, `SPEC-AUTH-001-jwt/`
  - **Composite domains**: Join with hyphens (e.g., `UPDATE-REFACTOR-001`)
  - **Warning**: Simplify when exceeding three hyphens
- **TAG content**: Free to change, but record every update in HISTORY
- **Version management**: Semantic Versioning (v0.0.1 → v0.1.0 → v1.0.0)
  - Detailed scheme: `spec-metadata.md#version-scheme`
- **Check for duplicates before creating a TAG**: `rg "@SPEC:{ID}" -n .moai/specs/` (mandatory)
- **TAG verification**: `rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/`
- **SPEC version consistency**: `rg "SPEC-{ID}.md v" -n`
- **Code-first principle**: The source of truth for TAGs lives in the code

### HISTORY Authoring Guidelines

**Change type tags**:
- `INITIAL`: First version (v1.0.0)
- `ADDED`: New feature/requirement → bump the minor version
- `CHANGED`: Updated content → bump the patch version
- `FIXED`: Bug/error fix → bump the patch version
- `REMOVED`: Removed features/requirements → bump the major version
- `BREAKING`: Breaking change → bump the major version
- `DEPRECATED`: Marked for future removal

**Mandatory metadata**:
- `AUTHOR`: Author/editor (GitHub ID)
- `REVIEW`: Reviewer and approval status
- `REASON`: Reason for the change (optional, recommended for significant updates)
- `RELATED`: Related issue/PR numbers (optional)

**HISTORY search examples**:
```bash
# Show the full change history for a given TAG
rg -A 20 "# @SPEC:AUTH-001" .moai/specs/SPEC-AUTH-001.md

# Extract just the HISTORY section
rg -A 50 "## HISTORY" .moai/specs/SPEC-AUTH-001.md

# Check only the most recent changes
rg "### v[0-9]" .moai/specs/SPEC-AUTH-001.md | head -3
```

---

## Development Principles

### Code Constraints

- <= 300 LOC per file
- <= 50 LOC per function
- <= 5 parameters per function
- Cyclomatic complexity <= 10

### Quality Benchmarks

- >= 85% test coverage
- Intent-revealing names
- Prefer guard clauses
- Use the standard tooling for each language

### Refactoring Rules

- **Rule of Three**: Plan refactoring when you hit the third repetition of a pattern.
- **Preparatory refactoring**: Set up the codebase to simplify the upcoming change before implementing it.
- **Clean up immediately**: Fix small issues right away; extract broader work into separate tasks.

## Exception Handling

Author a waiver when you exceed or deviate from the recommendations, and attach it to the PR/Issue/ADR.

**Waiver requirements**:

- Rationale and alternative options considered
- Risks and mitigation plan
- Temporary vs. permanent status
- Expiration criteria and approver

## Language Tooling Map

- **Python**: pytest (tests), mypy (type checking), black (formatting)
- **TypeScript**: Vitest (tests), Biome (lint + format)
- **Java**: JUnit (tests), Maven/Gradle (build)
- **Go**: go test (tests), gofmt (formatting)
- **Rust**: cargo test (tests), rustfmt (formatting)

## Variable Role Reference

| Role               | Description                         | Example                               |
| ------------------ | ----------------------------------- | ------------------------------------- |
| Fixed Value        | Constant after initialization       | `const MAX_SIZE = 100`                |
| Stepper            | Changes sequentially                | `for (let i = 0; i < n; i++)`         |
| Flag               | Boolean state indicator             | `let isValid = true`                  |
| Walker             | Traverses a data structure          | `while (node) { node = node.next; }`  |
| Most Recent Holder | Holds the most recent value         | `let lastError`                       |
| Most Wanted Holder | Holds optimal/maximum value         | `let bestScore = -Infinity`           |
| Gatherer           | Accumulator                         | `sum += value`                        |
| Container          | Stores multiple values              | `const list = []`                     |
| Follower           | Previous value of another variable  | `prev = curr; curr = next;`           |
| Organizer          | Reorganizes data                    | `const sorted = array.sort()`         |
| Temporary          | Temporary storage                   | `const temp = a; a = b; b = temp;`    |

---

This guide defines the standard for running the MoAI-ADK three-stage pipeline.
