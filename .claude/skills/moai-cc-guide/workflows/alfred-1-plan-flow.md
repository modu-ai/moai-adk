# Workflow: SPEC Authoring with Claude Code (`/alfred:1-plan`)

## Objective
Create requirement specifications while leveraging Claude Code context, memory, and standards.

## Claude Code Components Involved

| Component | Purpose | Reference |
|-----------|---------|-----------|
| **CLAUDE.md** | Reference EARS patterns, project standards | [`moai-cc-claude-md`](../../../skills/moai-cc-claude-md/SKILL.md) |
| **Skills** | Load Foundation tier patterns (EARS, SPEC metadata) | [`moai-cc-skills`](../../../skills/moai-cc-skills/SKILL.md) |
| **Memory** | Cache SPEC templates and naming conventions | [`moai-cc-memory`](../../../skills/moai-cc-memory/SKILL.md) |
| **Commands** | `/alfred:1-plan` entry point for spec-builder agent | [`moai-cc-commands`](../../../skills/moai-cc-commands/SKILL.md) |

## Step-by-Step Flow

### Step 1: Invoke Planning Command
```bash
/alfred:1-plan "JWT authentication system"

# Or with multiple topics:
/alfred:1-plan "User authentication" "Token refresh" "Session management"
```

**Behind the scenes:**
- ✅ Loads `.claude/CLAUDE.md` for project context
- ✅ Activates `moai-foundation-specs` Skill (SPEC metadata policy)
- ✅ Activates `moai-foundation-ears` Skill (EARS patterns)
- ✅ Retrieves recent SPEC templates from memory

### Step 2: spec-builder Agent Analysis
The spec-builder agent:
1. **Analyzes scope**: Existing code, related features, architecture
2. **Checks duplication**: "Is AUTH-001 or similar already defined?"
3. **Validates naming**: "Next ID is AUTH-002 (not AUTH-1)"
4. **Reviews context**: Recent specs, project standards from CLAUDE.md

**Claude Code role:**
- Uses `Explore` agent to search codebase
- Reads existing specs from `.moai/specs/`
- References CLAUDE.md for naming conventions

### Step 3: Interactive Clarification
```
[QUESTION 1] What is the scope of JWT authentication?
Options:
- Just token generation (narrow scope)
- Full auth cycle: generation, validation, refresh
- Enterprise: with audit logging and rate limiting

[QUESTION 2] Who should see the refresh endpoint?
Options:
- Only authenticated users
- Public (for SPA requests)
- Admin only
```

**Claude Code role:**
- `moai-alfred-interactive-questions` Skill renders TUI menu
- User navigates with arrow keys, confirms with enter
- Captures explicit intent (no guessing)

### Step 4: SPEC Generation
```bash
# spec-builder creates: .moai/specs/SPEC-AUTH-002/spec.md

---
id: AUTH-002
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @YourName
priority: high
---

# @SPEC:AUTH-002: JWT Authentication System

## Ubiquitous Requirements
- The system must provide JWT-based token generation
- The system must validate JWT tokens on protected endpoints

## Event-driven Requirements
- WHEN valid credentials are provided, the system SHOULD issue a JWT token
- WHEN a token expires, the system SHOULD return a 401 error
- WHEN an invalid token is provided, the system SHOULD deny access

## State-driven Requirements
- WHILE the user is authenticated, the system must allow access to protected resources

## Optional Features
- WHERE a refresh token is provided, the system can issue a new access token

## Constraints
- Access token expiration time must not exceed 15 minutes
- Refresh token expiration time must not exceed 7 days

## HISTORY
### v0.0.1 (2025-10-23)
- INITIAL: Draft JWT authentication SPEC
```

### Step 5: Tag Assignment
```bash
# spec-builder automatically assigns:
@SPEC:AUTH-002

# Validates uniqueness via:
rg "@SPEC:AUTH-002" -n  # Should only find this new spec
```

## Claude Code Best Practices

### ✅ DO
- Reference CLAUDE.md for project naming conventions
- Use Memory to cache SPEC templates and improve context
- Load Foundation tier Skills for EARS/SPEC guidance
- Create comprehensive SPEC before moving to TDD

### ❌ DON'T
- Skip the HISTORY section
- Forget the @SPEC:ID tag
- Create overlapping SPEC scopes
- Rush from planning to coding without explicit SPEC

## Validation Checklist

- [ ] SPEC file exists: `.moai/specs/SPEC-AUTH-002/spec.md`
- [ ] YAML frontmatter is valid (id, version, status, created, updated, author, priority)
- [ ] @SPEC:AUTH-002 TAG is present
- [ ] HISTORY section documents v0.0.1 INITIAL
- [ ] EARS syntax used (Ubiquitous, Event-driven, State-driven, Optional, Constraints)
- [ ] No duplicate TAG: `rg "@SPEC:AUTH-002" -n` returns only this file
- [ ] Status is `draft` (not `active` yet)

## Troubleshooting

**Issue**: "Spec ID already exists"
→ Use next available number: AUTH-001 exists? Use AUTH-002

**Issue**: "Interactive questions not showing"
→ Ensure `moai-alfred-interactive-questions` Skill is available and loaded

**Issue**: "EARS syntax not recognized"
→ Ensure `moai-foundation-ears` Skill is loaded via CLAUDE.md import

## Memory Optimization

The Memory system caches:
- ✅ Recent SPEC IDs (to prevent duplicates)
- ✅ EARS patterns (for faster typing)
- ✅ Team naming conventions (from CLAUDE.md)
- ✅ Related specs (for context during planning)

**Accessed via**: @moai-cc-memory guide

## Next Steps
→ Move to `/alfred:2-run` for TDD implementation with the approved SPEC

---

**Related Guides:**
- 📖 Project Setup: [`alfred-0-project-setup.md`](./alfred-0-project-setup.md)
- 📖 Implementation: [`alfred-2-run-flow.md`](./alfred-2-run-flow.md)
- 📖 Synchronization: [`alfred-3-sync-flow.md`](./alfred-3-sync-flow.md)
