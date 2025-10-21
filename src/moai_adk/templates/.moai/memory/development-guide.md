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

### EARS Requirement Authoring

**EARS (Easy Approach to Requirements Syntax)** provides a disciplined method for writing requirements.

#### Five EARS Patterns
1. **Ubiquitous Requirements**: The system shall provide [capability].
2. **Event-driven Requirements**: WHEN [condition], the system shall [behaviour].
3. **State-driven Requirements**: WHILE [state], the system shall [behaviour].
4. **Optional Features**: WHERE [condition], the system may [behaviour].
5. **Constraints**: IF [condition], the system shall enforce [constraint].

#### Practical Example
```markdown
### Ubiquitous Requirements (Baseline)
- The system shall provide user authentication.

### Event-driven Requirements
- WHEN a user logs in with valid credentials, the system shall issue a JWT token.
- WHEN a token expires, the system shall return a 401 error.

### State-driven Requirements
- WHILE the user remains authenticated, the system shall allow access to protected resources.

### Optional Features
- WHERE a refresh token is present, the system may issue a new access token.

### Constraints
- IF an invalid token is supplied, the system shall deny access.
- Access tokens shall not exceed a 15-minute lifetime.
```

---

## Context Engineering

MoAI-ADK follows Anthropic's principles from “Effective Context Engineering for AI Agents” to keep context lean and relevant.

### 1. JIT (Just-in-Time) Retrieval

**Principle**: Load documents only when needed to minimize initial context load.

**Alfred’s JIT Strategy**:

| Command          | Required Load    | Optional Load                    | Timing                             |
| ---------------- | ---------------- | -------------------------------- | ---------------------------------- |
| `/alfred:1-plan` | product.md       | structure.md, tech.md            | While discovering SPEC candidates  |
| `/alfred:2-run`  | SPEC-XXX/spec.md | development-guide.md             | At the start of TDD implementation |
| `/alfred:3-sync` | sync-report.md   | TAG chain validation (`rg` scan) | During documentation sync          |

**Implementation Notes**:
- Alfred uses the `Read` tool to load only the necessary documents at command time.
- Agents request only the documents relevant to their current task.
- The five documents listed in CLAUDE.md “Memory Strategy” are always loaded.

### Context Engineering Checklist

**When designing commands**:
- [ ] JIT: Do we load only the required documents?
- [ ] Optional Load: Do we load documents conditionally?

**When designing agents**:
- [ ] Minimum tools: Does the YAML frontmatter declare only the needed tools?
- [ ] Clear roles: Does each agent maintain a single responsibility?

---

## TRUST Principles (5 Pillars)

### T – Test-Driven Development (SPEC-Aligned)

**SPEC → Test → Code Cycle**:

- **SPEC**: Author detailed specifications first with `@SPEC:ID` tags (EARS format).
- **RED**: `@TEST:ID` – write failing tests tied to SPEC requirements and confirm failure.
- **GREEN**: `@CODE:ID` – implement the minimal code that passes the tests and satisfies the SPEC.
- **REFACTOR**: `@CODE:ID` – improve the code while preserving SPEC compliance, then document with `@DOC:ID`.

**Language-Specific TDD Tooling**:

- **Python**: pytest + SPEC-based test cases (with mypy type hints)
- **TypeScript**: Vitest + SPEC-driven suites (strict typing)
- **Java**: JUnit + SPEC annotations (behaviour-driven tests)
- **Go**: go test + SPEC table-driven tests (interface adherence)
- **Rust**: cargo test + SPEC doc tests (trait validation)
- **Ruby**: RSpec + SPEC-based BDD scenarios

Each test links SPEC requirements to implementations via `@TEST:ID → @CODE:ID`.

### R – Requirement-Driven Readability

**SPEC-Aligned Clean Code**:

- Functions implement SPEC requirements directly (≤ 50 LOC per function).
- Names mirror SPEC terminology and domain language.
- Code structure reflects SPEC design decisions.
- Comments are limited to SPEC clarifications and @TAG references.

**Language-Specific SPEC Implementation**:

- **Python**: Type hints mirroring SPEC interfaces + mypy validation
- **TypeScript**: Strict interfaces that match SPEC contracts
- **Java**: Classes implementing SPEC components with strong typing
- **Go**: Interfaces that satisfy SPEC requirements + gofmt
- **Rust**: Types enforcing SPEC safety requirements + rustfmt
- **Ruby**: Behaviour reflecting SPEC narratives + RuboCop validation

Every code element must remain traceable back to the SPEC via @TAG comments.

### U – Unified SPEC Architecture

- **SPEC-Driven Complexity**: Each SPEC defines its complexity threshold. Exceeding it requires a new SPEC or a documented exception.
- **SPEC vs. Implementation**: Keep authoring and implementation separate; never edit the SPEC mid-TDD cycle.
- **Language Boundaries**: SPECs define boundaries across languages (Python modules, TypeScript interfaces, Java packages, Go packages, Rust crates, etc.).
- **SPEC-Guided Architecture**: Domain boundaries follow the SPEC, not language conventions, with the @TAG system ensuring cross-language traceability.

### S – SPEC-Compliant Security

- **Security Requirements**: Every SPEC explicitly defines security needs, data sensitivity, and access control.
- **Security by Design**: Implement security controls during the TDD cycle, not after completion.
- **Language-Agnostic Security Patterns**:
  - Input validation based on SPEC interface definitions
  - Audit logging for SPEC-defined critical operations
  - Access controls aligned with the SPEC permission model
  - Secret management tailored to SPEC environment requirements

### T – SPEC Traceability

- **Spec-to-Code Traceability**: Every code change references SPEC IDs and requirements through the @TAG system.
- **Three-Stage Workflow Trace**:
  - `/alfred:1-plan`: Write SPECs with `@SPEC:ID` tags (`.moai/specs/`)
  - `/alfred:2-run`: Implement via TDD with `@TEST:ID` (tests/) → `@CODE:ID` (src/)
  - `/alfred:3-sync`: Sync documentation using `@DOC:ID` (docs/) and validate TAG coverage
- **Code Scan Verification**: Guarantee TAG traceability by scanning the codebase directly with `rg '@(SPEC|TEST|CODE|DOC):' -n`, without intermediate caches.

---

## SPEC-First Mindset

1. **SPEC-Led Decisions**: Reference an existing SPEC or author a new one before making any technical decision. Never implement without clear requirements.
2. **SPEC Context Review**: Read the relevant SPEC documents before changing code, understand the @TAG relationships, and confirm compliance.
3. **SPEC Communication**: Default to English for collaboration. SPEC documents should use precise technical terminology and plain, unambiguous explanations.

## SPEC-TDD Workflow

1. **Start with the SPEC**: Author or reference a SPEC before writing code. Use `/alfred:1-plan` to clarify requirements, design, and tasks.
2. **Implement with TDD**: Follow the Red–Green–Refactor loop rigorously using `/alfred:2-run` with language-appropriate testing frameworks.
3. **Maintain Traceability**: Run `/alfred:3-sync` to update documentation and preserve @TAG relationships between SPECs and code.

## @TAG System

### Core Chain

```text
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**Perfect TDD Alignment**:
- `@SPEC:ID` (Preparation) – requirements authored with the EARS pattern
- `@TEST:ID` (RED) – failing tests derived from the SPEC
- `@CODE:ID` (GREEN + REFACTOR) – implementation and refactoring
- `@DOC:ID` (Documentation) – live documents that capture the outcome

### TAG Block Template

> **📋 SPEC Metadata Standard (SSOT)**: see `spec-metadata.md`

**Every SPEC document must include YAML front matter and a HISTORY section**:
- **Required Fields (7)**: id, version, status, created, updated, author, priority
- **Optional Fields (9)**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **HISTORY Section**: log every version change (mandatory)

Find the complete template, field descriptions, and validation commands in `spec-metadata.md`.

**Quick Reference Example**:
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
- **INITIAL**: Authored the JWT authentication system SPEC
...
```

**Source Code (`src/`)**:
```text
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

**Test Code (`tests/`)**:
```text
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

### @CODE Subcategories (Comment Level)

Document implementation details inside `@CODE:ID` blocks:
- `@CODE:ID:API` – REST APIs, GraphQL endpoints
- `@CODE:ID:UI` – UI components, views, screens
- `@CODE:ID:DATA` – data models, schemas, types
- `@CODE:ID:DOMAIN` – business logic, domain rules
- `@CODE:ID:INFRA` – infrastructure, databases, integrations

### TAG Usage Rules

- **TAG ID Format**: `<DOMAIN>-<3 digits>` (e.g., `AUTH-003`) – immutable once created.
- **Directory Naming**: `.moai/specs/SPEC-{ID}/` (required)
  - ✅ Valid: `SPEC-AUTH-001/`, `SPEC-REFACTOR-001/`, `SPEC-UPDATE-REFACTOR-001/`
  - ❌ Invalid: `AUTH-001/`, `SPEC-001-auth/`, `SPEC-AUTH-001-jwt/`
  - **Composite Domains**: hyphenated combinations allowed (e.g., `UPDATE-REFACTOR-001`)
  - **Guideline**: Prefer fewer than three hyphen segments for clarity
- **TAG Contents**: May evolve freely; always record the rationale in HISTORY.
- **Versioning**: Semantic Versioning (v0.0.1 → v0.1.0 → v1.0.0)
  - See `spec-metadata.md#versioning` for details.
- **Duplicate Check**: Run `rg "@SPEC:{ID}" -n .moai/specs/` before creating a new TAG.
- **TAG Validation**: `rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/`
- **Version Alignment**: `rg "SPEC-{ID}.md v" -n`
- **Code-First Principle**: The source of truth for TAGs lives in the codebase.

### HISTORY Authoring Guide

**Change Type Tags**:
- `INITIAL`: First release (v1.0.0)
- `ADDED`: New requirement or capability → increment MINOR
- `CHANGED`: Adjusted behaviour → increment PATCH
- `FIXED`: Bug or defect fix → increment PATCH
- `REMOVED`: Removed capability → increment MAJOR
- `BREAKING`: Backward-incompatible change → increment MAJOR
- `DEPRECATED`: Marked for future removal

**Required Metadata**:
- `AUTHOR`: Contributor (GitHub ID)
- `REVIEW`: Reviewer and approval status
- `REASON`: Why the change was made (optional but recommended for significant updates)
- `RELATED`: Linked issues/PRs (optional)

**HISTORY Search Examples**:
```bash
# View the full change log for a TAG
rg -A 20 "# @SPEC:AUTH-001" .moai/specs/SPEC-AUTH-001.md

# Extract only the HISTORY section
rg -A 50 "## HISTORY" .moai/specs/SPEC-AUTH-001.md

# Check the latest entries
rg "### v[0-9]" .moai/specs/SPEC-AUTH-001.md | head -3
```

---

## Development Principles

### Code Constraints

- ≤ 300 LOC per file
- ≤ 50 LOC per function
- ≤ 5 parameters per function
- Cyclomatic complexity ≤ 10

### Quality Benchmarks

- ≥ 85% test coverage
- Use intention-revealing names
- Prefer guard clauses
- Leverage language-standard tooling

### Refactoring Rules

- **Rule of Three**: Plan refactoring when the same pattern appears a third time.
- **Preparatory Refactoring**: Shape the code for easy change before applying the change.
- **Tidy as You Go**: Fix small issues immediately; when scope expands, split into a dedicated effort.

## Exception Handling

When deviating from recommendations, document a waiver and attach it to the relevant PR, issue, or ADR.

**Waiver Checklist**:

- Justification and evaluated alternatives
- Risks and mitigation plan
- Temporary vs. permanent status
- Expiry conditions and approver

## Language Tooling Map

- **Python**: pytest (tests), mypy (type checks), black (formatting)
- **TypeScript**: Vitest (tests), Biome (lint + format)
- **Java**: JUnit (tests), Maven/Gradle (build)
- **Go**: go test (tests), gofmt (format)
- **Rust**: cargo test (tests), rustfmt (format)
- **Ruby**: RSpec (tests), RuboCop (lint + format), Bundler (packages)

## Variable Role Reference

| Role               | Description                        | Example                              |
| ------------------ | ---------------------------------- | ------------------------------------ |
| Fixed Value        | Constant after initialization      | `const MAX_SIZE = 100`               |
| Stepper            | Changes sequentially               | `for (let i = 0; i < n; i++)`        |
| Flag               | Boolean state indicator            | `let isValid = true`                 |
| Walker             | Traverses a data structure         | `while (node) { node = node.next; }` |
| Most Recent Holder | Holds the most recent value        | `let lastError`                      |
| Most Wanted Holder | Holds optimal/maximum value        | `let bestScore = -Infinity`          |
| Gatherer           | Accumulator                        | `sum += value`                       |
| Container          | Stores multiple values             | `const list = []`                    |
| Follower           | Previous value of another variable | `prev = curr; curr = next;`          |
| Organizer          | Reorganizes data                   | `const sorted = array.sort()`        |
| Temporary          | Temporary storage                  | `const temp = a; a = b; b = temp;`   |

---

This guide defines the standards for executing the three-stage MoAI-ADK pipeline.
