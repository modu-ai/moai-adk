# Workflow: Document Synchronization with Claude Code (`/alfred:3-sync`)

## Objective
Verify TAG chains, generate Living Documentation, create/update PRs with automated quality gates.

## Claude Code Components Involved

| Component | Purpose | Reference |
|-----------|---------|-----------|
| **Sub-agents** | tag-agent (verify chains), doc-syncer (generate docs) | [`moai-cc-agents`](../../../skills/moai-cc-agents/SKILL.md) |
| **Skills** | moai-essentials-review (code review), moai-foundation-tags (TAG policy) | [`moai-cc-skills`](../../../skills/moai-cc-skills/SKILL.md) |
| **Hooks (PostToolUse)** | Auto-generate docs, update README/CHANGELOG | [`moai-cc-hooks`](../../../skills/moai-cc-hooks/SKILL.md) |
| **Plugins/MCP** | GitHub MCP to create/update PR | [`moai-cc-mcp-plugins`](../../../skills/moai-cc-mcp-plugins/SKILL.md) |
| **Memory** | Aggregate changes into Living Docs, cache PR metadata | [`moai-cc-memory`](../../../skills/moai-cc-memory/SKILL.md) |

## Step-by-Step Flow

### Step 1: Invoke Sync Command
```bash
/alfred:3-sync

# Scans: .moai/specs/, tests/, src/, docs/
# Verifies: TAG chains, orphan detection, quality gates
```

**Behind the scenes:**
- ✅ Loads `moai-foundation-tags` Skill (TAG policy)
- ✅ Loads `moai-essentials-review` Skill (code review checklist)
- ✅ Loads `moai-foundation-trust` Skill (TRUST 5 validation)
- ✅ References CLAUDE.md for project standards

### Step 2: 🏷️ TAG Chain Verification

**tag-agent's task:**
1. Scan all files for @TAG references:
   ```bash
   rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
   ```

2. Verify completeness for AUTH-002:
   ```
   ✓ @SPEC:AUTH-002 → .moai/specs/SPEC-AUTH-002/spec.md
   ✓ @TEST:AUTH-002 → tests/auth/test_jwt_service.py
   ✓ @CODE:AUTH-002 → src/auth/service.py
   ✓ @DOC:AUTH-002 → docs/api/auth.md
   ```

3. Check for orphans:
   ```bash
   @CODE:AUTH-002 found → Is @SPEC:AUTH-002 defined? YES ✓
   @TEST:AUTH-001 found → Is @CODE:AUTH-001 defined? NO ✗ ORPHAN
   ```

**Claude Code role:**
- Scans using `Grep` and `Glob` tools
- Reports TAG chain integrity: 95% (4/4 AUTH-002 linked)
- Flags orphans for manual cleanup

### Step 3: 📖 Living Documentation Generation

**doc-syncer's task:**
1. Extract requirements from `@SPEC:AUTH-002`
2. Extract implementation from `@CODE:AUTH-002`
3. Extract test cases from `@TEST:AUTH-002`
4. Combine into `docs/api/auth.md`:

```markdown
# @DOC:AUTH-002: JWT Authentication API

**Status**: Active (v0.1.0)
**Implementation**: @CODE:AUTH-002
**Tests**: @TEST:AUTH-002
**SPEC**: @SPEC:AUTH-002

## Overview
JWT-based authentication system for API protection.

## Requirements (from SPEC)
✓ Generate JWT tokens for authenticated users
✓ Validate tokens on protected endpoints
✓ Return 401 on expired/invalid tokens
✓ Support token refresh

## API Endpoints
### POST /auth/login
Authenticate and receive JWT token.

**Implementation**: src/auth/service.py::JWTService.generate()

**Test Coverage**:
- ✓ test_generate_token_success
- ✓ test_token_expires
- ✓ test_invalid_token

## Configuration
Token expiry: 15 minutes (configurable via JWT_SECRET)

**Last Updated**: 2025-10-23
```

**Claude Code role:**
- Reads SPEC, code comments, and test names
- Uses Memory to aggregate related TAGs
- Auto-generates: docs/api/, README updates, CHANGELOG entries

### Step 4: ✅ Code Review & Quality Validation

**Code Review Sub-agent:**
1. Run TRUST 5 validation:
   - ✅ **Test**: Coverage ≥85%? `pytest --cov` shows 92%
   - ✅ **Readable**: Functions ≤50 LOC? Max is 35 LOC
   - ✅ **Unified**: Type hints present? ✅ All functions typed
   - ✅ **Secured**: Secrets in env vars? ✅ os.environ.get()
   - ✅ **Trackable**: @TAG chain complete? ✅ 4/4 linked

2. Check for code smells:
   - ✅ No hardcoded credentials
   - ✅ No overly complex logic
   - ✅ Error handling present
   - ✅ Documentation strings present

**Claude Code role:**
- Runs `moai-essentials-review` Skill
- `quality-gate` agent validates metrics
- Flags blockers (e.g., coverage < 85%)

### Step 5: 🚀 GitHub PR Creation (if plugin enabled)

**GitHub MCP:**
1. Create branch: `feature/spec-auth-002`
2. Commit all changes:
   ```bash
   git add -A
   git commit -m "feat(AUTH-002): complete JWT auth implementation

   - RED: failing tests for JWT generation
   - GREEN: minimal JWT service implementation
   - REFACTOR: improve quality per TRUST 5

   Refs: @SPEC:AUTH-002 @TEST:AUTH-002 @CODE:AUTH-002
   "
   ```

3. Create PR with auto-generated description:
   ```markdown
   ## Summary
   - ✓ Implement JWT authentication per @SPEC:AUTH-002
   - ✓ Tests passing with 92% coverage
   - ✓ All TRUST 5 principles validated
   - ✓ TAG chain verified (@SPEC → @TEST → @CODE → @DOC)

   ## Checklist
   - [x] SPEC requirements met
   - [x] Tests pass (92% coverage)
   - [x] Code follows project conventions
   - [x] Documentation updated
   - [x] TAG chain complete

   **Related**: Closes #123 (example)
   ```

**Claude Code role:**
- Uses GitHub MCP to manage PRs
- Automatically creates Draft PR
- Sets reviewers from CLAUDE.md

### Step 6: 📊 Status Report

```
╔════════════════════════════════════════════════════╗
║        /alfred:3-sync COMPLETION REPORT            ║
╠════════════════════════════════════════════════════╣
║ TAG VERIFICATION                                   ║
║   ✓ SPEC Chain: 1 (AUTH-002)                       ║
║   ✓ TEST Chain: 1 (AUTH-002)                       ║
║   ✓ CODE Chain: 1 (AUTH-002)                       ║
║   ✓ DOC Chain: 1 (AUTH-002)                        ║
║   ✓ Integrity: 100% (4/4 linked)                   ║
║   ⚠ Orphans: 0 detected                            ║
║                                                    ║
║ QUALITY VALIDATION                                 ║
║   ✓ Test Coverage: 92% (≥85%)                      ║
║   ✓ TRUST 5: All principles passed                 ║
║   ✓ Code Smells: None detected                     ║
║   ✓ Security: No issues found                      ║
║                                                    ║
║ DOCUMENTATION                                      ║
║   ✓ Living Docs: Generated (docs/api/auth.md)     ║
║   ✓ README: Updated with AUTH-002 status          ║
║   ✓ CHANGELOG: v0.1.0 entry added                 ║
║                                                    ║
║ PR STATUS                                          ║
║   ✓ GitHub PR #124: Created (Draft → Ready)       ║
║   ✓ Branch: feature/spec-auth-002                  ║
║   ✓ Changes: +187 lines, -5 lines                  ║
║                                                    ║
╚════════════════════════════════════════════════════╝

Next: Review PR #124, merge to main
```

## Claude Code Best Practices During Sync

### ✅ DO
- Verify all TAG chains before PR creation
- Ensure Living Docs are generated from code
- Validate TRUST 5 before marking PR Ready
- Use Memory to aggregate related changes
- Document why changes were made

### ❌ DON'T
- Create PR without TAG chain verification
- Skip quality validation
- Leave orphan TAGs
- Merge without PR review
- Forget to update CHANGELOG

## Validation Checklist

### TAG Chain
- [ ] @SPEC:AUTH-002 exists and is valid
- [ ] @TEST:AUTH-002 exists with tests
- [ ] @CODE:AUTH-002 exists with implementation
- [ ] @DOC:AUTH-002 exists with documentation
- [ ] No orphan TAGs detected

### Documentation
- [ ] Living Docs generated: `docs/api/auth.md`
- [ ] README updated with new feature
- [ ] CHANGELOG.md updated with v0.1.0 entry
- [ ] All cross-references working

### Quality Validation
- [ ] Test coverage ≥85%
- [ ] TRUST 5 validation passed
- [ ] No code smells or security issues
- [ ] All requirements from SPEC met

### PR Status
- [ ] PR created on GitHub
- [ ] Branch name follows convention: `feature/spec-*`
- [ ] PR description includes @TAG references
- [ ] Status changed from Draft → Ready
- [ ] Ready for team review

## Troubleshooting

**Issue**: "TAG chain incomplete: missing @DOC"
→ Create docs/api/auth.md with @DOC:AUTH-002 tag

**Issue**: "Orphan TAG detected: @CODE:AUTH-003 without @SPEC"
→ Either create @SPEC:AUTH-003 or remove @CODE tag

**Issue**: "GitHub PR creation failed"
→ Verify GitHub MCP is configured with GITHUB_TOKEN

**Issue**: "TRUST validation failed: coverage only 78%"
→ Add more test cases to increase coverage to ≥85%

## Memory Optimization

The Memory system during `/alfred:3-sync`:
- ✅ Caches TAG inventory (avoid re-scanning)
- ✅ Stores Living Doc templates (reuse structure)
- ✅ Remembers TRUST 5 validation rules
- ✅ Tracks PR metadata (avoid manual entry)

**Accessed via**: @moai-cc-memory guide

## Next Steps
→ Review PR with team
→ Merge to main branch
→ Start next SPEC with `/alfred:1-plan`

---

**Related Guides:**
- 📖 Project Setup: [`alfred-0-project-setup.md`](./alfred-0-project-setup.md)
- 📖 Planning: [`alfred-1-plan-flow.md`](./alfred-1-plan-flow.md)
- 📖 Implementation: [`alfred-2-run-flow.md`](./alfred-2-run-flow.md)
