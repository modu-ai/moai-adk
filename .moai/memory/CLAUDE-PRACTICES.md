# CLAUDE-PRACTICES.md

> MoAI-ADK Practical Workflows & Examples

---

## For Alfred: Why This Document Matters

When Alfred reads this document:
1. When performing actual tasks - "How specifically should I execute this?"
2. When context management is needed - "How can I use Explore efficiently?"
3. When solving problems - "How do I diagnose and resolve this error/problem?"
4. When onboarding new developers - "Learn MoAI-ADK workflows through practice"

Alfred's Decision Making:
- "What are the specific steps to perform this task?"
- "How can I collect the necessary context minimally?"
- "Where should I diagnose problems when they occur?"

After reading this document:
- Master JIT (Just-in-Time) context management strategies
- Learn how to use the Explore agent efficiently
- Master specific commands for SPEC → TDD → Sync execution
- Reference solutions for frequently occurring problems

---
→ Related Documents:
- [For rules verification, see CLAUDE-RULES.md](./CLAUDE-RULES.md#skill-invocation-rules)
- [For Agent selection, see CLAUDE-AGENTS-GUIDE.md](./CLAUDE-AGENTS-GUIDE.md#agent-selection-decision-tree)

---

## Context Engineering Strategy

### 1. JIT (Just-in-Time) Retrieval

- Pull only the context required for the immediate step.
- Prefer `Explore` over manual file hunting.
- Cache critical insights in the task thread for reuse.

#### Efficient Use of Explore

- Request call graphs or dependency maps when changing core modules.
- Fetch examples from similar features before implementing new ones.
- Ask for SPEC references or TAG metadata to anchor changes.

### 2. Layered Context Summaries

1. **High-level brief**: purpose, stakeholders, success criteria.
2. **Technical core**: entry points, domain models, shared utilities.
3. **Edge cases**: known bugs, performance constraints, SLAs.

### 3. Living Documentation Sync

- Align code, tests, and docs after each significant change.
- Use `/alfred:3-sync` to update Living Docs and TAG references.
- Record rationale for deviations from the SPEC.

---

## On-demand Agent Usage

### Debugging & Analysis

- Use `debug-helper` for error triage and hypothesis testing.
- Attach logs, stack traces, and reproduction steps.
- Ask for fix-forward vs rollback recommendations.

### TAG System Management

- Assign IDs as `<DOMAIN>-<###>` (e.g., `AUTH-003`).
- Update HISTORY with every change.
- Cross-check usage with `rg '@TAG:ID' -n` searches.

### Backup Management

- `/alfred:0-project` and `git-manager` create automatic safety snapshots (e.g., `.moai-backups/`) before risky actions.
- Manual `/alfred:9-checkpoint` commands have been deprecated; rely on Git branches or team-approved backup workflows when additional restore points are needed.

---

## Practical Workflow Examples

### Scenario 1: Implementing a New Feature (USER-DASHBOARD-001)

**Situation**: User requests "Add user dashboard"

**Alfred's Execution Sequence**:

1. **Detect Ambiguity → Execute AskUserQuestion**
   ```
   Question 1: What is the data source?
   Options: [REST API | GraphQL | Local State]

   Question 2: What is the refresh cycle?
   Options: [Real-time (WebSocket) | Every 10 seconds | Manual refresh]

   Question 3: Is access restriction needed?
   Options: [Admin only | Logged-in users | Public]

   User Answers:
   - Data source: REST API
   - Refresh cycle: Manual refresh
   - Access restriction: Logged-in users
   ```

2. **Write SPEC (Based on User Answers)**
   ```bash
   /alfred:1-plan "User Dashboard Feature - Display user stats with manual refresh, authenticated access only"
   ```

   **Output**: `.moai/specs/SPEC-USER-DASHBOARD-001/spec.md`
   - YAML metadata: id, version: 0.0.1, status: draft
   - @SPEC:USER-DASHBOARD-001 TAG
   - EARS syntax requirements:
     - "The system must display user statistics dashboard"
     - "WHEN user clicks refresh button, THEN fetch latest data from REST API"
     - "IF user not authenticated, THEN redirect to login page"

3. **TDD Implementation (RED → GREEN → REFACTOR)**
   ```bash
   /alfred:2-run USER-DASHBOARD-001
   ```

   **Alfred Internal Execution**:
   - **implementation-planner** (Phase 1):
     - Establish implementation strategy: React component + fetch API + auth guard
     - Library selection: react-query (data fetching), @tanstack/react-query (caching)
     - TAG design: @CODE:USER-DASHBOARD-001:UI, @CODE:USER-DASHBOARD-001:API

   - **tdd-implementer** (Phase 2):
     - **RED**: Write `tests/features/dashboard.test.tsx` (failing tests)
     - **GREEN**: Implement `src/features/Dashboard.tsx` (tests pass)
     - **REFACTOR**: Clean code, separate hooks, improve reusability

4. **Document Synchronization**
   ```bash
   /alfred:3-sync
   ```

   **Alfred Internal Execution**:
   - TAG chain verification: @SPEC ↔ @TEST ↔ @CODE
   - Living Document update: README.md, CHANGELOG.md
   - PR status change: Draft → Ready

**Final Outputs**:
- SPEC: `.moai/specs/SPEC-USER-DASHBOARD-001/spec.md` (@SPEC:USER-DASHBOARD-001)
- TEST: `tests/features/dashboard.test.tsx` (@TEST:USER-DASHBOARD-001)
- CODE: `src/features/Dashboard.tsx` (@CODE:USER-DASHBOARD-001:UI)
- CODE: `src/api/dashboard.ts` (@CODE:USER-DASHBOARD-001:API)
- DOCS: `docs/features/USER-DASHBOARD-001.md` (@DOC:USER-DASHBOARD-001)

**Estimated Duration**: 30-45 minutes (SPEC 10min + TDD 20min + Sync 10min)

---

### Scenario 2: Bug Fix (BUG-AUTHENTICATION-TIMEOUT)

**Situation**: User reports "Authentication automatically disconnects after 5 minutes" bug

**Alfred's Execution Sequence**:

1. **Error Analysis (debug-helper)**
   ```bash
   @agent-debug-helper "Authentication timeout after 5 minutes - expected 30 minutes"
   ```

   **debug-helper Analysis Results**:
   - Which function causes timeout? → `src/auth/token.ts:validateToken()`
   - What is current timeout value? → `300000 ms` (5 minutes)
   - What should the normal value be? → `1800000 ms` (30 minutes)
   - Cause: JWT token expiration time incorrectly configured

2. **Write SPEC (For Bug Fix)**
   ```bash
   /alfred:1-plan "Fix AUTH-TIMEOUT-001: JWT token expiration should be 30 minutes, not 5 minutes"
   ```

   **Output**: `.moai/specs/SPEC-AUTH-TIMEOUT-001/spec.md`
   - Bug description: Fix JWT expiration from 5min → 30min
   - Root cause: `expiresIn` value error (change `300` → `1800`)
   - Test case: Verify token validity for 30 minutes

3. **TDD Implementation (RED → GREEN → REFACTOR)**
   ```bash
   /alfred:2-run AUTH-TIMEOUT-001
   ```

   **Alfred Internal Execution**:
   - **RED**: Add `tests/auth/token.test.ts`
     ```typescript
     it('should keep token valid for 30 minutes', () => {
       const token = generateToken();
       const now = Date.now();
       const futureTime = now + 30 * 60 * 1000;
       expect(isTokenValid(token, futureTime)).toBe(true);
     });
     ```

   - **GREEN**: Modify `src/auth/token.ts`
     ```typescript
     const JWT_EXPIRATION = 1800; // 30 minutes (was 300)
     ```

   - **REFACTOR**: Constantize
     ```typescript
     const JWT_EXPIRATION_MINUTES = 30;
     const JWT_EXPIRATION = JWT_EXPIRATION_MINUTES * 60;
     ```

4. **Verification**
   - **TRUST 5 Check**:
     - Test First: ✅ New test case added
     - Readable: ✅ ruff lint passed
     - Unified: ✅ mypy type safety passed
     - Secured: ✅ trivy security scan passed
     - Trackable: ✅ @TAG chain normal

   - **TAG Chain Verification**:
     ```bash
     rg '@(SPEC|TEST|CODE):AUTH-TIMEOUT-001' -n
     ```
     - @SPEC:AUTH-TIMEOUT-001 → `.moai/specs/SPEC-AUTH-TIMEOUT-001/spec.md`
     - @TEST:AUTH-TIMEOUT-001 → `tests/auth/token.test.ts`
     - @CODE:AUTH-TIMEOUT-001 → `src/auth/token.ts`

**Final Outputs**:
- SPEC updated
- TEST added
- CODE modified (1 line)
- Git commit: `fix(auth): Extend JWT expiration to 30 minutes (was 5 minutes) - Refs: @AUTH-TIMEOUT-001`

**Estimated Duration**: 15-20 minutes (Analysis 5min + SPEC 5min + TDD 5min + Verification 5min)

---

### Scenario 3: Document Synchronization (Automatic)

**Situation**: Keep documents up to date after code modifications

**Alfred's Execution Sequence**:

1. **Check Changed Files**
   ```bash
   git diff develop...HEAD
   ```

   **Results**:
   - `src/features/Dashboard.tsx` (modified)
   - `src/api/dashboard.ts` (new)
   - `tests/features/dashboard.test.tsx` (new)

2. **Living Document Verification**
   ```bash
   /alfred:3-sync status
   ```

   **doc-syncer Analysis**:
   - README.md update needed: Add "User Dashboard" to Features section
   - CHANGELOG.md creation needed: v0.4.2 release notes
   - TAG integrity verified: All @CODE linked to @SPEC

3. **TAG Integrity Check**
   ```bash
   rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
   ```

   **Results**:
   - ✅ @SPEC:USER-DASHBOARD-001 → @TEST:USER-DASHBOARD-001 ✅
   - ✅ @TEST:USER-DASHBOARD-001 → @CODE:USER-DASHBOARD-001:UI ✅
   - ✅ @CODE:USER-DASHBOARD-001:UI → @DOC:USER-DASHBOARD-001 ✅
   - 🎉 No orphan TAGs detected

4. **PR Status Change (Draft → Ready)**
   ```bash
   @agent-git-manager "Move PR #42 from Draft to Ready"
   ```

   **git-manager Execution**:
   - PR verification: All tests passed, coverage ≥85%
   - PR label update: `draft` → `ready-for-review`
   - Auto-assign reviewer: GOOS오라버니
   - PR description update: Reflect CHANGELOG.md content

**Final Outputs**:
- README.md auto-updated (Features section)
- CHANGELOG.md auto-generated (v0.4.2 entry)
- TAG chain verification completed
- PR #42 status: Draft → Ready for Review

**Estimated Duration**: 5-10 minutes (automated)

---

### Scenario 4: Multilingual Workflow (Direct Language Pass-through)

**Situation**: Users request features in their preferred language. System works directly in that language while keeping infrastructure (Skills, code) in English.

**Direct Pass-through Pattern**:

```
User Input (any language):  "[Feature request in user's language]"
                             Examples:
                             - "Add user authentication feature" (English)
                             - "사용자 인증 기능 추가해줘" (Korean)
                             - "ユーザー認証機能を追加してください" (Japanese)
                             - "Agregar funcionalidad de autenticación de usuarios" (Spanish)

↓ [Alfred passes directly to Sub-agent]

Task(subagent_type="spec-builder",
     prompt="사용자 인증 기능을 추가해주세요. JWT 전략 사용, 이메일+비밀번호 로그인, 30분 토큰 만료",
     conversation_language="ko")

↓ [spec-builder receives Korean prompt directly]

spec-builder Execution:
1. Recognizes SPEC authoring task
2. Explicitly invokes: Skill("moai-foundation-specs") ✅
3. Explicitly invokes: Skill("moai-foundation-ears") ✅
4. Reads English Skill content (technical documentation)
5. Generates Korean SPEC document

↓ [spec-builder generates output in user's language]

SPEC Output (Korean with English technical elements):
{
  "id": "AUTH-001",  ← English (technical identifier)
  "version": "0.0.1",
  "requirements": [  ← Korean (narrative)
    "시스템은 JWT 기반 인증을 제공해야 한다",
    "WHEN 유효한 자격 증명이 제공될 때, THEN 30분 만료 시간의 JWT 토큰을 발급한다",
    "IF 사용자가 인증되지 않았을 때, THEN 로그인 페이지로 리다이렉트한다"
  ]
}

↓ [Alfred receives Korean output]

User Response (in user's language):
- English: "I've created the authentication SPEC with JWT strategy..."
- Korean: "인증 SPEC을 작성했습니다. JWT 전략으로..."
- Japanese: "認証SPECを作成しました。JWT戦略で..."
- Spanish: "He creado la especificación de autenticación. Con estrategia JWT..."
```

**Key Principles**:

| Aspect | Implementation |
|--------|-----------------|
| **User-Facing** | User's configured language (conversations, documents) |
| **Task Prompts** | User's language (passed directly to Sub-agents) |
| **Skill Invocation** | Explicit: `Skill("moai-foundation-*")` (works with any language) |
| **Static Infrastructure** | English only (Skills, agents, commands, code comments) |
| **Translation Points** | None - direct pass-through |

**Why This Works**:
- ✅ **Explicit invocation**: `Skill("name")` works regardless of prompt language
- ✅ **Zero maintenance**: No translation logic to maintain
- ✅ **Infinite scalability**: Add any language without code changes
- ✅ **Simplified architecture**: Direct language flow, no overhead
- ✅ **Industry standard**: Technical docs in English (Skills), UI in user's language

**Estimated Duration**: Same as English (no translation overhead)

---

**Last Updated**: 2025-10-27
**Document Version**: v1.0.0
