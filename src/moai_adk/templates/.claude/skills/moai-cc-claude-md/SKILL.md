---
name: moai-cc-claude-md
description: "Authoring CLAUDE.md Project Instructions. Design project-specific AI guidance, document workflows, define architecture patterns. Use when creating CLAUDE.md files for projects, documenting team standards, or establishing AI collaboration guidelines."
allowed-tools: "Read, Write, Edit, Glob"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Ops |
| Auto-load | When authoring CLAUDE.md files |

## What It Does

프로젝트별 CLAUDE.md 파일 작성을 위한 표준 구조와 패턴을 제공합니다. AI 협업 가이드라인, 워크플로우 문서화, 아키텍처 패턴 정의 방법을 다룹니다.

## When to Use

- 새 프로젝트의 CLAUDE.md 파일을 생성할 때
- 기존 CLAUDE.md를 업데이트하거나 개선할 때
- 팀 표준 및 개발 워크플로우를 문서화할 때
- AI 협업 패턴을 정의할 때


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

**Reference**: Claude Code CLAUDE.md documentation
**Version**: 1.0.0
