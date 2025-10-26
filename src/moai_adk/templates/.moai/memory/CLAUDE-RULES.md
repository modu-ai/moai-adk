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

---

## Git Commit Message Standard

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

### TAG Core Rules

- **TAG ID**: `<Domain>-<3 digits>` (e.g., `AUTH-003`) — immutable.
- **TAG Content**: Flexible but record changes in HISTORY.
- **Versioning**: Semantic Versioning (`v0.0.1 → v0.1.0 → v1.0.0`).
- **TAG References**: Use file names without versions (e.g., `SPEC-AUTH-001.md`).
- **Duplicate Check**: `rg "@SPEC:AUTH" -n` or `rg "AUTH-001" -n`.
- **Code-first**: The source of truth lives in code.

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

Alfred enforces these quality gates on every change:

- **T**est First: Use the best testing tool per language (Jest/Vitest, pytest, go test, cargo test, JUnit, flutter test, ...).
- **R**eadable: Run linters (ESLint/Biome, ruff, golint, clippy, dart analyze, ...).
- **U**nified: Ensure type safety or runtime validation.
- **S**ecured: Apply security/static analysis tools.
- **T**rackable: Maintain @TAG coverage directly in code.

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
