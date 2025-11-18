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

   - **tdd-implementer** (Phase 2):
     - **RED**: Write `tests/features/dashboard.test.tsx` (failing tests)
     - **GREEN**: Implement `src/features/Dashboard.tsx` (tests pass)
     - **REFACTOR**: Clean code, separate hooks, improve reusability

4. **Document Synchronization**
   ```bash
   /alfred:3-sync
   ```

   **Alfred Internal Execution**:
   - Living Document update: README.md, CHANGELOG.md
   - PR status change: Draft → Ready

**Final Outputs**:

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

   - **TAG Chain Verification**:
     ```bash
     rg '@(SPEC|TEST|CODE):AUTH-TIMEOUT-001' -n
     ```

**Final Outputs**:
- SPEC updated
- TEST added
- CODE modified (1 line)

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

3. **TAG Integrity Check**
   ```bash
   rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
   ```

   **Results**:
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

### Scenario 4: Multilingual Workflow (Language Boundary Pattern)

**Situation**: Users request features in their preferred language. System responds in that language while keeping all internal operations in English.

**Generic Multilingual Pattern**:

```
User Input (any language):  "[Feature request in user's language]"
                             Examples:
                             - "Add user authentication feature" (English)
                             - "사용자 인증 기능 추가해줘" (Korean)
                             - "ユーザー認証機能を追加してください" (Japanese)
                             - "Agregar funcionalidad de autenticación de usuarios" (Spanish)

↓ [Alfred's Internal Translation Layer]

Alfred Internal Prompt:      "Create authentication feature SPEC with JWT strategy,
                             email+password login, 30-minute token expiry"

↓ [Invoke spec-builder with English prompt]

Task(subagent_type="spec-builder",
     prompt="Create SPEC for user authentication with JWT tokens...")

↓ [spec-builder receives English prompt]

spec-builder Execution:
1. Detects SPEC authoring task → Skill("moai-foundation-specs") ✅ [100% match]
2. Applies EARS syntax → Skill("moai-foundation-ears") ✅ [100% match]
3. Returns structured English SPEC

↓ [Alfred receives English SPEC output]

SPEC Output:
{
  "id": "AUTH-001",
  "version": "0.0.1",
  "requirements": [
    "The system must provide JWT-based authentication",
    "WHEN valid credentials provided, THEN issue JWT token with 30-minute expiry",
    "IF user not authenticated, THEN redirect to login page"
  ]
}

↓ [Alfred's Response Translation Layer]

User Response (in user's language):
- English: "I've created the authentication SPEC with JWT strategy..."
- Korean: "인증 SPEC을 작성했습니다. JWT 전략으로..."
- Japanese: "認証SPECを作成しました。JWT戦略で..."
- Spanish: "He creado la especificación de autenticación. Con estrategia JWT..."
```

**Key Principles**:

| Aspect | Implementation |
|--------|-----------------|
| **User-Facing (External)** | User's configured language (flexible) |
| **Internal Operations (Layer 2)** | English only (Task prompts, Sub-agent communication) |
| **Skills & Code (Layer 3)** | English only (Skill descriptions, code comments) |
| **Translation Points** | User Input → English (entry), English → User Language (response) |

**Why This Works**:
- ✅ **Skills remain unchanged**: English-only Skills work reliably for ANY user language
- ✅ **Zero maintenance burden**: No need to translate 55 Skills into N languages
- ✅ **Infinite scalability**: Add Korean, Russian, Mandarin, Arabic without code changes
- ✅ **Consistent quality**: English prompts guarantee 100% Skill trigger matching
- ✅ **Industry standard**: Same pattern used by Netflix, Google, AWS (localized UI + English backend)

**Estimated Duration**: Same as English (no overhead from translation layer)

---

**Last Updated**: 2025-10-27
**Document Version**: v1.0.0
