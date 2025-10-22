---
name: "Authoring CLAUDE.md Project Instructions"
description: "Design project-specific AI guidance, document workflows, define architecture patterns. Use when creating CLAUDE.md files for projects, documenting team standards, or establishing AI collaboration guidelines."
allowed-tools: "Read, Write, Edit, Glob"
---

# Authoring CLAUDE.md Project Instructions

CLAUDE.md is a Markdown file that provides Claude with project-specific context, workflows, standards, and guidance. It acts as a living document for AI collaboration patterns.

## CLAUDE.md File Location & Purpose

**Location**: `./{PROJECT_ROOT}/CLAUDE.md` or `/~/.claude/CLAUDE.md` (personal)

**Purpose**: Tell Claude about your project before every session starts.

## Core Sections

### Section 1: Project Overview

```markdown
# Project Name

**Description**: What does this project do?

**Repository**: GitHub URL
**Tech Stack**: Technologies, languages, frameworks
**Team**: Size, roles
**Status**: Active, archived, experimental
```

### Section 2: Core Workflow

```markdown
## Development Workflow

### Phase 1: Planning
- Use `/alfred:1-plan` for SPEC creation
- EARS syntax: Ubiquitous, Event-driven, State-driven, Optional, Constraints
- Store in `.moai/specs/`

### Phase 2: Implementation (TDD)
- RED: Write failing tests
- GREEN: Minimal implementation
- REFACTOR: Improve quality
- Tools: pytest, TypeScript, mypy, ruff

### Phase 3: Sync & Documentation
- Run `/alfred:3-sync` to validate
- Update Living Documents
- Verify @TAG chains
- Prepare PR
```

### Section 3: Code Standards

```markdown
## Code Standards (TRUST 5 Principles)

### T - Test First
- Target coverage: ≥ 85%
- Framework: pytest for Python, Jest for TypeScript
- Pattern: TDD (RED → GREEN → REFACTOR)

### R - Readable
- Max file: 300 LOC
- Max function: 50 LOC
- Max params: 5
- Use linters: ruff (Python), ESLint (TypeScript)

### U - Unified
- Type safety: mypy for Python, TypeScript strict mode
- Consistent patterns across domain
- Shared utilities in `src/core/`

### S - Secured
- Input validation everywhere
- No hardcoded secrets
- Use environment variables
- Run: bandit (Python), npm audit

### T - Trackable
- @TAG system: @SPEC:ID, @TEST:ID, @CODE:ID, @DOC:ID
- Document all changes in HISTORY section
- Link code to SPEC requirements
```

### Section 4: Project Architecture

```markdown
## Architecture

### Directory Structure
```
src/
├── core/          # Shared utilities, domain models
├── domain/        # Business logic, DOMAIN-specific
├── interfaces/    # API endpoints, CLI commands
├── middleware/    # Cross-cutting concerns
└── infra/         # Database, external services

tests/
├── unit/          # Unit tests for core/ and domain/
├── integration/   # API, database tests
└── e2e/          # End-to-end workflows
```

### Key Design Decisions
- Monolithic backend (for now)
- Database: PostgreSQL with migrations
- Authentication: JWT tokens
- API: REST with OpenAPI docs
```

### Section 5: AI Collaboration Patterns

```markdown
## Working with Claude Code

### When to Use Sub-agents
- `debug-helper`: Errors, test failures, exceptions
- `security-auditor`: Vulnerability assessment, OWASP checks
- `architect`: Refactoring, system design, scalability
- `code-reviewer`: Quality analysis, SOLID violations

### Commands Available
- `/alfred:1-plan "feature description"` — Create SPEC
- `/alfred:2-run SPEC-ID` — Implement (TDD)
- `/alfred:3-sync` — Sync docs and validate
- `/review-code src/**/*.ts` — Code review
- `/deploy [env]` — Deploy pipeline

### Context Engineering Tips
- Always mention relevant SPEC ID
- Provide file paths relative to project root
- Link to similar existing features
- Mention constraints or non-negotiables upfront
```

### Section 6: Known Gotchas & Decisions

```markdown
## Important Notes

### Why We Use SPEC-First
- Clarifies requirements upfront
- Prevents scope creep
- Enables parallel work (different SPECs)
- Makes code changes traceable

### Common Mistakes to Avoid
- ❌ Implementing without SPEC
- ❌ Skipping tests for "quick" fixes
- ❌ Mixing multiple features in one PR
- ❌ Ignoring @TAG system

### Team Decisions
- Use pytest fixtures for mocking (not monkeypatch)
- All API responses must include status codes
- Database migrations are CI/CD blocking
- Security audit runs on every PR
```

## CLAUDE.md Examples by Domain

### Example 1: Web API Project
```markdown
# Transaction API

**Tech Stack**: Python (FastAPI), PostgreSQL, pytest

## Phase Workflow

1. **SPEC**: Requirements in EARS syntax
   - Example: `The API must validate transaction amounts > 0`
   - Stored in: `.moai/specs/SPEC-TRANS-{###}/spec.md`

2. **TDD**: Implement with test-first approach
   - Tests: `tests/integration/test_transactions.py`
   - Code: `src/domain/transaction.py`

3. **Sync**: Verify completeness
   - Run: `/alfred:3-sync`
   - Check: TAG chain integrity
```

### Example 2: React Frontend Project
```markdown
# Competition Dashboard

**Tech Stack**: TypeScript, React 19, Vitest, Tailwind

## Development Standards

- Framework: Next.js 15 (App Router)
- Testing: Vitest + React Testing Library
- Styling: Tailwind CSS with shadcn/ui
- State: Zustand for client state, Server Components for data

## Key Patterns
- Server Components for data fetching
- Suspense boundaries for loading states
- Error boundaries for graceful failures
```

## High-Freedom: Architectural Decisions

```markdown
## Why This Architecture?

### Monolith vs Microservices
- **Chosen**: Monolithic backend
- **Reason**: Team < 10, throughput manageable
- **When to reconsider**: >10M requests/month or 3+ teams

### Database: PostgreSQL
- **Reason**: Strong ACID guarantees, mature ecosystem
- **Alternatives considered**: MongoDB (rejected: unclear schema)
```

## Medium-Freedom: Workflow Definition

```markdown
## Code Review Checklist

Before merging, reviewer must verify:
1. [ ] SPEC ID linked in PR description
2. [ ] All tests passing (>85% coverage)
3. [ ] No security issues (bandit clean)
4. [ ] Code follows TRUST 5 principles
5. [ ] @TAG chain complete (@SPEC → @TEST → @CODE → @DOC)
6. [ ] CHANGELOG updated
```

## Low-Freedom: Explicit Rules

```markdown
## Non-negotiable Rules

- ❌ No commits without @TAG references
- ❌ No merging PRs without passing tests
- ❌ No pushing secrets to repo
- ❌ No force-pushing to main/master
- ✅ All features must have SPEC document
- ✅ All code changes must have corresponding tests
- ✅ All SPECs must use EARS syntax
```

## Validation Checklist

- [ ] Project name and description clear
- [ ] Tech stack documented
- [ ] Development workflow defined (Plan → Run → Sync)
- [ ] TRUST 5 principles explained
- [ ] Architecture diagram or structure described
- [ ] AI collaboration patterns defined
- [ ] Known gotchas documented
- [ ] Examples provided for key workflows

## Best Practices

✅ **DO**:
- Keep CLAUDE.md up-to-date as standards evolve
- Include real examples (links to actual files/PRs)
- Document why (not just what)
- Link to external docs (architecture decisions, security policy)
- Review with team before finalizing

❌ **DON'T**:
- Explain general programming concepts (Claude already knows)
- List every possible workflow (focus on your specific patterns)
- Write as instruction to humans (write as guidance to Claude)
- Update CLAUDE.md only at project start (evolve as needed)

---

## 🤝 Works Well With

**Complementary Skills:**
- **Foundation Skills** (moai-foundation-trust, moai-foundation-tags, moai-foundation-specs) - Import via @references
- **Essentials Skills** (moai-essentials-debug, moai-essentials-review) - Reference quality guidelines
- **moai-cc-hooks** - SessionStart Hook loads CLAUDE.md for context
- **moai-cc-memory** - CLAUDE.md imports other memory files

**MoAI-ADK Workflows:**
- **`/alfred:0-project`** - Creates project CLAUDE.md documenting SPEC-first principles
- **`/alfred:1-plan`** - Spec-builder loads CLAUDE.md for project standards
- **`/alfred:2-run`** - References CLAUDE.md for TRUST 5 principles
- **`/alfred:3-sync`** - Doc-syncer uses CLAUDE.md metadata for Living Docs

**Example Integration (MoAI-ADK):**
```bash
# 1. Create project CLAUDE.md
@agent-cc-manager "Create CLAUDE.md for MoAI-ADK project with:
  - SPEC-first principles
  - EARS requirement syntax
  - TRUST 5 quality gates
  - @TAG traceability rules
  - Imports of moai-foundation-* Skills"

# 2. Reference in workflows
/alfred:1-plan  # Spec-builder reads CLAUDE.md for EARS patterns

# 3. Update CLAUDE.md as project evolves
@agent-cc-manager "Update CLAUDE.md to reflect new architecture patterns"
```

**Common MoAI Patterns:**
- ✅ Document SPEC-first philosophy in CLAUDE.md
- ✅ Import Foundation tier Skills (@moai-foundation-*)
- ✅ List custom slash commands for team
- ✅ Define TAG naming conventions
- ✅ Specify TRUST 5 quality gates per language
- ✅ Link to team architecture decisions

**General Claude Code Patterns:**
- ✅ Project-specific AI guidance
- ✅ 4-level hierarchy (Enterprise → Project → User → Local)
- ✅ Imports and cross-references
- ✅ Context layering and discovery

**See Also:**
- 📖 **Orchestrator Guide:** `Skill("moai-cc-guide")` → SKILL.md
- 📖 **Project Setup:** `Skill("moai-cc-guide")` → workflows/alfred-0-project-setup.md
- 📖 **Memory Management:** `Skill("moai-cc-memory")` → Context Strategies

---

**Reference**: Claude Code CLAUDE.md documentation
**Version**: 1.0.0
