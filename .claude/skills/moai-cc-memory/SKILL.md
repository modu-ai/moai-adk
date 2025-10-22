---
name: "Managing Claude Code Session Memory & Context"
description: "Understand session context limits, use just-in-time retrieval, cache insights, manage memory files. Use when optimizing context usage, handling large projects, or implementing efficient workflows."
allowed-tools: "Read, Write, Glob, Bash"
---

# Managing Claude Code Session Memory & Context

Claude Code operates within context windows (~100K-200K tokens). Effective memory management ensures productive sessions without hitting limits.

## Context Budget Overview

```
Total Context Budget
├── System Prompt (~2K)
├── Tools & Instructions (~5K)
├── Session History (~30K)
├── Project Context (~40K)
└── Available for Response (~23K)
```

## Just-in-Time (JIT) Retrieval Strategy

### High-Freedom: Core Principles

**Principle 1: Pull Only What You Need**
- Don't load entire codebase upfront
- Load files relevant to immediate task
- Use Glob/Grep for targeted searches
- Cache results for reuse

**Principle 2: Prefer Explore Over Manual Hunting**
```bash
# ❌ Manual approach: Search many files, load all
rg "authenticate" src/ | head -20

# ✅ JIT approach: Use Explore agent
@agent-Explore "Find authentication implementation, analyze"
```

**Principle 3: Layered Context Summaries**
```
1. High-level brief (purpose, success criteria)
   ↓
2. Technical core (entry points, domain models)
   ↓
3. Edge cases (known bugs, constraints)
```

### Example: Feature Implementation

```
Task: "Add email verification to signup"

JIT Retrieval:
├── Read: User model (src/domain/user.ts)
├── Read: Signup endpoint (src/api/auth.ts)
├── Grep: "email" in tests (understand patterns)
├── Glob: Find email service (src/infra/email.*)
└── Cache: Signup flow diagram in memory
```

## Medium-Freedom: Memory File Patterns

### Pattern 1: Session Summary Cache

**File**: `.moai/memory/session-summary.md`

```markdown
# Session Summary

## Current Task
- Feature: User email verification
- SPEC: AUTH-015
- Status: In RED phase (writing tests)

## Key Files
- Test: tests/auth/email_verify.test.ts
- Impl: src/domain/email_service.ts
- Config: src/config/email.ts

## Important Context
- Email service uses SendGrid API
- Verification tokens expire in 24h
- Already have similar flow for password reset (AUTH-012)

## Assumptions Made
- Assuming transactional emails only
- Async email sending OK
- No SMS verification needed
```

### Pattern 2: Architecture Reference

**File**: `.moai/memory/architecture.md`

```markdown
# Architecture Reference

## Data Flow for Email Verification

```
User(Browser)
    ↓ [POST /auth/signup]
Server
    ↓ [Create user + token]
DB
    ↓ [sendEmail async]
Queue
    ↓ [Process job]
Email Service (SendGrid)
    ↓
User receives email with link
User clicks link
    ↓ [GET /auth/verify?token=...]
Server validates token
    ↓ [Mark user verified]
DB
    ↓
User logged in
```

## Module Boundaries
- `domain/`: Business logic (no framework)
- `api/`: HTTP endpoints only
- `infra/`: External services (SendGrid, DB)
```

### Pattern 3: Known Gotchas Cache

**File**: `.moai/memory/gotchas.md`

```markdown
# Common Pitfalls in This Project

## Email Service
- SendGrid has rate limit: 100 emails/sec per account
- Test mode uses fake email (won't actually send)
- Async job failures don't alert (check logs)

## Database
- Migrations must be reviewed before prod deploy
- Test DB is reset after each suite
- Foreign key constraints enforced (plan deletions)

## Authentication
- JWT tokens stored in httpOnly cookies (XSRF protected)
- Refresh token rotation required (not automatic)
- Session timeout: 7 days (hardcoded, not configurable yet)
```

## Low-Freedom: Memory Management Practices

### Practice 1: Caching Key Insights

```
After reading code:
1. Note file locations (~5 min read)
2. Summarize key logic (~2 min)
3. Write to memory file (~1 min)
4. Reference in next session
```

**Example memory entry**:
```
# USER-002: Email verification flow

## Key Code Locations
- Token generation: src/domain/user.ts:generateVerificationToken()
- Email sending: src/infra/email_service.ts:sendVerificationEmail()
- Token validation: src/api/auth.ts:POST /verify

## Logic Summary
1. User submits email → server generates token (16 chars, base64)
2. Token stored in DB with 24h expiry
3. Email sent async via SendGrid
4. User clicks link → token validated → user marked verified
5. Token deleted after use (can't reuse)

## Related TESTs
- tests/auth/email_verify.test.ts (GREEN phase - needs implementation)
- Similar flow: password reset (PASSWORD-001)
```

### Practice 2: Session Boundary Management

**Before switching between tasks**:
```markdown
# Session Handoff Note

## Completed
✓ RED phase: 3 test cases for email verification
✓ GREEN phase: Minimal implementation passing tests
✓ REFACTOR: Added input validation

## Status
- Current: Ready for /alfred:3-sync
- Next action: Run full test suite, then sync docs

## Context for Next Session
- SPEC: .moai/specs/SPEC-AUTH-015/spec.md
- Tests: tests/auth/email_verify.test.ts (all passing)
- Code: src/domain/email_service.py
- Database migration: pending (see migrations/ directory)

## Assumptions
- SendGrid API key set in .env
- Test mode uses mock email service
- Database schema includes email_verified_at column
```

### Practice 3: Cleanup Before Session End

```bash
# Remove unnecessary cached files
rm .moai/memory/temp-*.md

# Archive completed memory files
mv .moai/memory/feature-x-* .moai/memory/archive/

# Keep only active session memory
ls -la .moai/memory/
# session-summary.md (current)
# architecture.md (reference)
# gotchas.md (patterns)
```

## Memory File Organization

```
.moai/
├── memory/
│   ├── session-summary.md      # Current session state
│   ├── architecture.md         # System design reference
│   ├── gotchas.md             # Common pitfalls
│   ├── spec-index.md          # List of all SPECs + status
│   ├── api-reference.md       # API endpoints quick lookup
│   └── archive/               # Completed session notes
│       ├── feature-auth-*
│       └── feature-api-*
└── specs/                      # Requirement specifications
    ├── SPEC-AUTH-001/
    ├── SPEC-USER-002/
    └── SPEC-API-003/
```

## Context Optimization Checklist

- [ ] Memory files describe architecture (not code)
- [ ] Session summary updated before handoff
- [ ] Key file locations cached (don't re-search)
- [ ] Assumptions explicitly documented
- [ ] No duplicate information between memory files
- [ ] Archive files moved after session completion
- [ ] All cached insights reference file paths
- [ ] Memory files are Markdown (human-readable)

## Best Practices

✅ **DO**:
- Use Explore for large searches
- Cache results in memory files
- Keep memory files < 500 lines each
- Update session-summary.md before switching tasks
- Reference memory files in handoff notes

❌ **DON'T**:
- Load entire src/ or docs/ directory upfront
- Duplicate context between memory files
- Store memory files outside `.moai/memory/`
- Leave stale session notes (archive or delete)
- Cache raw code (summarize logic instead)

## Commands for Memory Management

```bash
# View current session memory
cat .moai/memory/session-summary.md

# List all memory files
ls -la .moai/memory/

# Archive completed work
mv .moai/memory/feature-old-* .moai/memory/archive/

# Search memory files
grep -r "email verification" .moai/memory/

# Count context usage estimate
wc -w .moai/memory/*.md  # Total words
```

---

## 🤝 Works Well With

**Complementary Skills:**
- **moai-cc-claude-md** - CLAUDE.md implements memory hierarchy with imports (@path/file)
- **moai-cc-hooks** - SessionStart Hook seeds context from memory files
- **moai-cc-skills** - Skills can store state in memory
- **moai-cc-agents** - Agents reference cached results to speed up analysis

**MoAI-ADK Workflows:**
- **`/alfred:1-plan`** - Memory caches recent SPEC templates, naming conventions
- **`/alfred:2-run`** - Memory stores test results, import paths, type hints
- **`/alfred:3-sync`** - Memory aggregates changes for Living Doc generation
- **All phases** - Session context preserved across command invocations

**Example Integration (MoAI-ADK):**
```bash
# 1. Cache SPEC naming patterns in project memory
# .moai/memory/naming-conventions.md
ID Pattern: DOMAIN-###
Example: AUTH-001, USER-002, PAYMENT-015

# 2. Cache test fixtures
# .moai/memory/fixtures.md
import sys; sys.path.insert(0, 'src')
from tests.fixtures import app, client, db

# 3. Reference in workflows
/alfred:1-plan  # Loads cached naming conventions
/alfred:2-run   # Loads cached fixtures
```

**Common MoAI Patterns:**
- ✅ Cache SPEC naming conventions (preventing duplicates)
- ✅ Store test fixtures and helper functions
- ✅ Save import patterns and type hints
- ✅ Aggregate test results (for TRUST 5 coverage tracking)
- ✅ Store error patterns (for debug-helper guidance)

**General Claude Code Patterns:**
- ✅ JIT (Just-in-Time) retrieval
- ✅ Context layering (Enterprise → Project → User → Local)
- ✅ Session persistence across commands
- ✅ Smart caching to avoid re-computation

**See Also:**
- 📖 **Orchestrator Guide:** `Skill("moai-cc-guide")` → SKILL.md
- 📖 **CLAUDE.md Integration:** `Skill("moai-cc-claude-md")` → Memory Hierarchy
- 📖 **Implementation:** `Skill("moai-cc-guide")` → workflows/alfred-2-run-flow.md

---

**Reference**: Claude Code Context Management
**Version**: 1.0.0
