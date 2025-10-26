# CLAUDE-RULES.md

> MoAI-ADK Mandatory Rules & Standards

---

## Alfred를 위해: 이 문서가 필요한 이유

Alfred가 이 문서를 읽는 시점:
1. Skill을 호출하기 직전 - "이 Skill 호출이 필수인가, 선택인가?"
2. 사용자 질문이 모호할 때 - "AskUserQuestion을 써야 할 상황인가?"
3. 코드를 검증할 때 - "TRUST 5 원칙을 모두 지켰는가?"
4. Git 커밋 전 - "이 커밋 메시지 형식이 맞는가?"
5. TAG 체인 무결성 확인 시 - "TAG 규칙을 따랐는가?"

Alfred의 의사결정:
- "이 상황에서 반드시 Skill을 호출해야 하는가?"
- "사용자의 모호한 질문에 대해 AskUserQuestion을 실행할 것인가?"
- "이 코드/커밋이 우리 규칙을 모두 준수했는가?"

이 문서를 읽으면:
- 10가지 필수 Skill 호출 시나리오 이해
- AskUserQuestion의 5가지 필수 상황 숙달
- TRUST 5의 5가지 품질 게이트 적용 가능
- TAG 규칙과 검증 방법 숙달

---
→ 관련 문서:
- [Agent 선택 기준은 CLAUDE-AGENTS-GUIDE.md](./CLAUDE-AGENTS-GUIDE.md#agent-선택-결정-트리)를 참조하세요
- [구체적 실행 예제는 CLAUDE-PRACTICES.md](./CLAUDE-PRACTICES.md#실전-워크플로우-예제)를 참조하세요

---

## 🎯 Skill Invocation Rules (English-Only)

### ✅ Mandatory Skill Explicit Invocation

**CRITICAL**: All 55 Skills in MoAI-ADK must be invoked **explicitly** using the `Skill("skill-name")` syntax. DO NOT use direct tools (Bash, Grep, Read) when a dedicated Skill exists for the task.

| **User Request Keywords** | **Skill to Invoke** | **Invocation Pattern** | **Prohibited Actions** |
|----------------------|-------------------|----------------------|-------------------|
| TRUST validation, code quality check, quality gate, coverage check, test coverage | `moai-foundation-trust` | `Skill("moai-foundation-trust")` | ❌ Direct ruff/mypy |
| TAG validation, tag check, orphan detection, TAG scan, TAG chain | `moai-foundation-tags` | `Skill("moai-foundation-tags")` | ❌ Direct rg search |
| SPEC validation, spec check, SPEC metadata, spec authoring | `moai-foundation-specs` | `Skill("moai-foundation-specs")` | ❌ Direct YAML reading |
| EARS syntax, requirement authoring, requirement formatting | `moai-foundation-ears` | `Skill("moai-foundation-ears")` | ❌ Generic templates |
| Git workflow, branch management, PR policy, commit strategy | `moai-foundation-git` | `Skill("moai-foundation-git")` | ❌ Direct git commands |
| Language detection, stack detection, framework identification | `moai-foundation-langs` | `Skill("moai-foundation-langs")` | ❌ Manual detection |
| Debugging, error analysis, bug fix, exception handling | `moai-essentials-debug` | `Skill("moai-essentials-debug")` | ❌ Generic diagnostics |
| Refactoring, code improvement, code cleanup, design patterns | `moai-essentials-refactor` | `Skill("moai-essentials-refactor")` | ❌ Direct modifications |
| Performance optimization, profiling, bottleneck analysis | `moai-essentials-perf` | `Skill("moai-essentials-perf")` | ❌ Guesswork |
| Code review, quality review, architecture review, security review | `moai-essentials-review` | `Skill("moai-essentials-review")` | ❌ Generic review |

### Skill Tier Overview (55 Total Skills)

| **Tier** | **Count** | **Purpose** | **Auto-Trigger Conditions** |
|----------|-----------|------------|--------------------------|
| **Foundation** | 6 | Core TRUST/TAG/SPEC/EARS/Git/Language principles | Keyword detection in user request |
| **Essentials** | 4 | Debug/Perf/Refactor/Review workflows | Error detection, refactor triggers |
| **Alfred** | 11 | Workflow orchestration (SPEC authoring, TDD, sync, Git) | Command execution (`/alfred:*`) |
| **Domain** | 10 | Backend, Frontend, Web API, Database, Security, DevOps, Data Science, ML, Mobile, CLI | Domain-specific keywords |
| **Language** | 23 | Python, TypeScript, Go, Rust, Java, Kotlin, Swift, Dart, C/C++, C#, Scala, Ruby, PHP, JavaScript, SQL, Shell, and more | File extension detection (`.py`, `.ts`, `.go`, etc.) |
| **Ops** | 1 | Claude Code session settings, output styles | Session start/configuration |

### Progressive Disclosure Pattern

All Skills follow the **Progressive Disclosure** principle:

1. **Metadata** (always available): Skill name, description, triggers, keywords
2. **Content** (on-demand): Full SKILL.md loads when explicitly invoked via `Skill("name")`
3. **Supporting** (JIT): Templates, examples, and resources load only when needed

### Explicit Invocation Syntax

**Standard Pattern**:
```python
Skill("skill-name")  # Invoke any Skill explicitly
```

**With Context** (recommended):
```python
# Example: Validate code quality
Skill("moai-foundation-trust")

# Example: Debug runtime error
Skill("moai-essentials-debug")
```

### Example Workflows Using Explicit Skill Invocation

**Workflow 1: Code Quality Validation (TRUST 5)**
```
User: "Check code quality"
    ↓
Invoke: Skill("moai-foundation-trust")
    → Verify Test First: pytest coverage ≥85%
    → Verify Readable: ruff lint + linter checks
    → Verify Unified: mypy type safety
    → Verify Secured: security scanner (trivy)
    → Verify Trackable: @TAG chain validation
    → Return: Quality report with TRUST 5-principles
```

**Workflow 2: TAG Orphan Detection (Full Project)**
```
User: "Find all TAG orphans in the project"
    ↓
Invoke: Skill("moai-foundation-tags")
    → Scan entire project: .moai/specs/, tests/, src/, docs/
    → Detect @CODE without @SPEC
    → Detect @SPEC without @CODE
    → Detect @TEST without @SPEC
    → Detect @DOC without @SPEC/@CODE
    → Return: Complete orphan report with locations
```

**Workflow 3: SPEC Authoring with EARS**
```
User: "Create AUTH-001 JWT authentication SPEC"
    ↓
Invoke: Skill("moai-foundation-specs")
    → Validate SPEC structure (YAML metadata, HISTORY)
    ↓
Invoke: Skill("moai-foundation-ears")
    → Format requirements using EARS syntax
    → Ubiquitous: "The system must provide JWT-based authentication"
    → Event: "WHEN valid credentials provided, THEN issue JWT token"
    → Constraints: "Token expiration ≤ 30 minutes"
    ↓
Return: Properly formatted SPEC file with @SPEC:AUTH-001 TAG
```

**Workflow 4: Debugging with Error Context**
```
User: "TypeError: Cannot read property 'name' of undefined"
    ↓
Invoke: Skill("moai-essentials-debug")
    → Analyze stack trace
    → Identify root cause: null/undefined object access
    → Check related SPEC: @SPEC:USER-003
    → Check missing test cases: @TEST:USER-003
    → Suggest fix: Add null check, update test
    → Recommend: Re-run /alfred:2-run
```

---

## 🎯 Interactive Question Rules

### Mandatory AskUserQuestion Usage

**IMPORTANT**: When the user needs to make a **choice** or **decision**, you **MUST** use AskUserQuestion. DO NOT make assumptions or implement directly.

| Situation Type | Examples | Invocation | Required |
|---------------|----------|------------|----------|
| **Multiple valid approaches exist** | Database choice (PostgreSQL vs MongoDB), state management library (Redux vs Zustand), test framework selection | `AskUserQuestion(...)` | ✅ Required |
| **Architecture/design decisions** | Microservices vs monolithic, client-side vs server-side rendering, authentication method (JWT vs OAuth) | `AskUserQuestion(...)` | ✅ Required |
| **Ambiguous or high-level requirements** | "Add a dashboard", "Optimize performance", "Add multi-language support" | `AskUserQuestion(...)` | ✅ Required |
| **Requests affecting existing components** | Refactoring scope, backward compatibility, migration strategy | `AskUserQuestion(...)` | ✅ Required |
| **User experience/business logic decisions** | UI layout, data display method, workflow order | `AskUserQuestion(...)` | ✅ Required |

### Optional AskUserQuestion Usage

You can proceed without AskUserQuestion in the following situations:

- ✅ User has already provided clear instructions
- ✅ Standard conventions or best practices are obvious
- ✅ Technical constraints allow only one approach
- ✅ User explicitly states "just implement it, I've already decided"

### Example: Correct AskUserQuestion Usage

#### ❌ **Incorrect Response** (Making assumptions)

```
User: "Add a dashboard page"

Claude: [Proceeding without asking]
- Tech: React + TypeScript
- Layout: Left sidebar + main content
- Chart library: Recharts
- Data source: Direct REST API calls

Starting implementation...
```

**Problem**: High risk of rework due to not understanding actual user intent.

#### ✅ **Correct Response** (Asking questions)

```
User: "Add a dashboard page"

Claude: I'll need to clarify a few decisions for the dashboard implementation.

AskUserQuestion(...)

[Example questions]
- "Data source?" → [REST API | GraphQL | Local state]
- "Primary chart type?" → [Time series | Category comparison | Distribution]
- "Real-time updates needed?" → [Yes | No | Every 10 seconds]
- "Access restrictions?" → [Admin only | Logged-in users | Public]

[After user responses]
Confirmed selections:
- Data: REST API
- Charts: Time series
- Real-time: No
- Access: Logged-in users

Now implementing based on these requirements...
```

**Benefits**: Clear intent understanding, minimized rework, accurate implementation.

### Best Practices for AskUserQuestion

1. **Limit to 3-5 questions**
   - ✅ "Choose from 3 mutually exclusive options"
   - ❌ "10+ options" (user fatigue)

2. **Options must be specific**
   - ✅ "PostgreSQL (ACID, JSON support)", "MongoDB (horizontal scaling, flexible schema)"
   - ❌ "Database 1", "Database 2"

3. **Always include "Other" option**
   - User's choice may not be listed
   - "Other" allows custom input

4. **Summary step after selection**
   - Display user selections summary
   - "Proceed with these choices?" final confirmation

5. **Integrate with Context Engineering**
   - Analyze existing code/SPEC before AskUserQuestion
   - Provide context like "Your project currently uses X"

### When NOT to Use AskUserQuestion

❌ When user has already given specific instructions:
```
User: "Implement state management using Zustand"
→ AskUserQuestion unnecessary (already decided)
```

❌ When only one technical choice exists:
```
User: "Improve type safety in TypeScript"
→ AskUserQuestion unnecessary (type system is fixed)
```

---

## Alfred's Next-Step Suggestion Principles

### Pre-suggestion Checklist

Before suggesting the next step, always verify:
- You have the latest status from agents.
- All blockers are documented with context.
- Required approvals or user confirmations are noted.
- Suggested tasks include clear owners and outcomes.
- There is at most one "must-do" suggestion per step.

**cc-manager validation sequence**

1. **SPEC** – Confirm the SPEC file exists and note its status (`draft`, `active`, `completed`, `archived`). If missing, queue `/alfred:1-plan`.
2. **TEST & CODE** – Check whether tests and implementation files exist and whether the latest test run passed. Address failing tests before proposing new work.
3. **DOCS & TAGS** – Ensure `/alfred:3-sync` is not pending, Living Docs and TAG chains are current, and no orphan TAGs remain.
4. **GIT & PR** – Review the current branch, Draft/Ready PR state, and uncommitted changes. Highlight required Git actions explicitly.
5. **BLOCKERS & APPROVALS** – List outstanding approvals, unanswered questions, TodoWrite items, or dependency risks.

> cc-manager enforces this order. Reference the most recent status output when replying, and call out the next mandatory action (or confirm that all gates have passed).

### Poor Suggestion Examples (❌)

- Suggesting tasks already completed.
- Mixing unrelated actions in one suggestion.
- Proposing work without explaining the problem or expected result.
- Ignoring known blockers or assumptions.

### Good Suggestion Examples (✅)

- Link the suggestion to a clear goal or risk mitigation.
- Reference evidence (logs, diffs, test output).
- Provide concrete next steps with estimated effort.

### Suggestion Restrictions

- Do not recommend direct commits; always go through review.
- Avoid introducing new scope without confirming priority.
- Never suppress warnings or tests without review.
- Do not rely on manual verification when automation exists.

### Suggestion Priorities

1. Resolve production blockers → 2. Restore failing tests → 3. Close gaps against SPEC → 4. Improve DX/automation.

---

## Error Message Standard (Shared)

### Severity Icons

- 🔴 Critical failure (stop immediately)
- 🟠 Major issue (needs immediate attention)
- 🟡 Warning (monitor closely)
- 🔵 Info (no action needed)

### Message Format

```
🔴 <Title>
- Cause: <root cause>
- Scope: <affected components>
- Evidence: <logs/screenshots/links>
- Next Step: <required action>
```

---

## Git Commit Message Standard (Locale-aware)

### TDD Stage Commit Templates

| Stage    | Template                                                   |
| -------- | ---------------------------------------------------------- |
| RED      | `test: add failing test for <feature>`                     |
| GREEN    | `feat: implement <feature> to pass tests`                  |
| REFACTOR | `refactor: clean up <component> without changing behavior` |

### Commit Structure

```
<type>(scope): <subject>

- Context of the change
- Additional notes (optional)

Refs: @TAG-ID (if applicable)

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Alfred <alfred@mo.ai.kr>
```

**Signature Standard**: All git commits created through MoAI-ADK are attributed to **Alfred** (`alfred@mo.ai.kr`), the MoAI SuperAgent orchestrating all Git operations. This ensures clear traceability and accountability for all automated workflows.

---

## @TAG Lifecycle

### Core Principles

- TAG IDs never change once assigned.
- Content can evolve; log updates in HISTORY.
- Tie implementations and tests to the same TAG.

### TAG Structure

- `@SPEC:ID` in specs
- `@CODE:ID` in source
- `@TEST:ID` in tests
- `@DOC:ID` in docs

### TAG Block Template

```
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

### HISTORY Example

```
### v0.0.1 (2025-09-15)

- **INITIAL**: Draft the JWT-based authentication SPEC.
```

### TAG Core Rules

- **TAG ID**: `<Domain>-<3 digits>` (e.g., `AUTH-003`) — immutable.
- **TAG Content**: Flexible but record changes in HISTORY.
- **Versioning**: Semantic Versioning (`v0.0.1 → v0.1.0 → v1.0.0`).
  - Detailed rules: see `@.moai/memory/spec-metadata.md#versioning`.
- **TAG References**: Use file names without versions (e.g., `SPEC-AUTH-001.md`).
- **Duplicate Check**: `rg "@SPEC:AUTH" -n` or `rg "AUTH-001" -n`.
- **Code-first**: The source of truth lives in code.

### @CODE Subcategories (Comment Level)

- `@CODE:ID:API` — REST/GraphQL endpoints
- `@CODE:ID:UI` — Components and UI
- `@CODE:ID:DATA` — Data models, schemas, types
- `@CODE:ID:DOMAIN` — Business logic
- `@CODE:ID:INFRA` — Infra, databases, integrations

### TAG Validation & Integrity

**Avoid duplicates**:
```bash
rg "@SPEC:AUTH" -n          # Search AUTH specs
rg "@CODE:AUTH-001" -n      # Targeted ID search
rg "AUTH-001" -n            # Global ID search
```

**TAG chain verification** (`/alfred:3-sync` runs automatically):
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# Detect orphaned TAGs
rg '@CODE:AUTH-001' -n src/          # CODE exists
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC missing → orphan
```

---

## TRUST 5 Principles (Language-agnostic)

> Detailed guide: `@.moai/memory/development-guide.md#trust-5-principles`

Alfred enforces these quality gates on every change:

- **T**est First: Use the best testing tool per language (Jest/Vitest, pytest, go test, cargo test, JUnit, flutter test, ...).
- **R**eadable: Run linters (ESLint/Biome, ruff, golint, clippy, dart analyze, ...).
- **U**nified: Ensure type safety or runtime validation.
- **S**ecured: Apply security/static analysis tools.
- **T**rackable: Maintain @TAG coverage directly in code.

**Language-specific guidance**: `.moai/memory/development-guide.md#trust-5-principles`.

---

## Language-specific Code Rules

**Global constraints**:
- Files ≤ 300 LOC
- Functions ≤ 50 LOC
- Parameters ≤ 5
- Cyclomatic complexity ≤ 10

**Quality targets**:
- Test coverage ≥ 85%
- Intent-revealing names
- Early guard clauses
- Use language-standard tooling

**Testing strategy**:
- Prefer the standard framework per language
- Keep tests isolated and deterministic
- Derive cases directly from the SPEC

---

## TDD Workflow Checklist

**Step 1: SPEC authoring** (`/alfred:1-plan`)
- [ ] Create `.moai/specs/SPEC-<ID>/spec.md` (with directory structure)
- [ ] Add YAML front matter (id, version: 0.0.1, status: draft, created)
- [ ] Include the `@SPEC:ID` TAG
- [ ] Write the **HISTORY** section (v0.0.1 INITIAL)
- [ ] Use EARS syntax for requirements
- [ ] Check for duplicate IDs: `rg "@SPEC:<ID>" -n`

**Step 2: TDD implementation** (`/alfred:2-run`)
- [ ] **RED**: Write `@TEST:ID` under `tests/` and watch it fail
- [ ] **GREEN**: Add `@CODE:ID` under `src/` and make the test pass
- [ ] **REFACTOR**: Improve code quality; document TDD history in comments
- [ ] List SPEC/TEST file paths in the TAG block

**Step 3: Documentation sync** (`/alfred:3-sync`)
- [ ] Scan TAGs: `rg '@(SPEC|TEST|CODE):' -n`
- [ ] Ensure no orphan TAGs remain
- [ ] Regenerate the Living Document
- [ ] Move PR status from Draft → Ready

---

**마지막 업데이트**: 2025-10-26
**문서 버전**: v1.0.0 (Option A Refactoring)
