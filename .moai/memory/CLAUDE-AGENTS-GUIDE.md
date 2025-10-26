# CLAUDE-AGENTS-GUIDE.md

> MoAI-ADK Agent Architecture & Decision Guide

---

## Alfred를 위해: 이 문서가 필요한 이유

Alfred가 이 문서를 읽는 시점:
1. 새로운 작업을 받았을 때 - "어떤 Sub-agent를 호출할 것인가"를 결정
2. 복합 작업이 필요할 때 - 여러 agent의 순서와 협업 방식 결정
3. 팀 구성을 재검토할 때 - 각 agent의 책임 범위 확인

Alfred의 의사결정:
- "이 작업은 spec-builder가 담당해야 하나, 아니면 code-builder가 담당해야 하나?"
- "Explore agent를 호출해야 할 때와 하지 말아야 할 때는?"
- "이 작업에 Haiku 모델이 충분한가, 아니면 Sonnet이 필요한가?"

이 문서를 읽으면:
- 19개 Sub-agent의 책임 범위를 명확히 이해
- 55개 Skills가 어떻게 Tier별로 구분되는지 파악
- Agent 협업의 원칙 (Command Precedence, Single Responsibility 등) 숙달
- Haiku vs Sonnet 모델 선택 기준 습득

---
→ 관련 문서:
- [Alfred의 의사결정 규칙은 CLAUDE-RULES.md](./CLAUDE-RULES.md#skill-invocation-rules)를 참조하세요
- [실제 Agent 호출 예제는 CLAUDE-PRACTICES.md](./CLAUDE-PRACTICES.md#실전-워크플로우-예제)를 참조하세요

---

## 4-Layer Architecture (v0.4.0)

| Layer           | Owner              | Purpose                                                            | Examples                                                                                                 |
| --------------- | ------------------ | ------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------- |
| **Commands**    | User ↔ Alfred      | Workflow entry points that establish the Plan → Run → Sync cadence | `/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`                                 |
| **Sub-agents**  | Alfred             | Deep reasoning and decision making for each phase                  | project-manager, spec-builder, code-builder pipeline, doc-syncer                                         |
| **Skills (55)** | Claude Skills      | Reusable knowledge capsules loaded just-in-time                    | Foundation (TRUST/TAG/Git), Essentials (debug/refactor/review), Alfred workflow, Domain & Language packs |
| **Hooks**       | Runtime guardrails | Fast validation + JIT context hints (<100 ms)                      | SessionStart status card, PreToolUse destructive-command blocker                                         |

---

## Core Sub-agent Roster

> Alfred + 10 core sub-agents + 6 zero-project specialists + 2 built-in Claude agents = **19-member team**
>
> **Note on Counting**: The "code-builder pipeline" is counted as 1 conceptual agent but implemented as 2 physical files (`implementation-planner` + `tdd-implementer`) for sequential RED → GREEN → REFACTOR execution. This maintains the 19-member team concept while acknowledging that 20 distinct agent files exist in `.claude/agents/alfred/`.

| Sub-agent                   | Model  | Phase       | Responsibility                                                                                 | Trigger                      |
| --------------------------- | ------ | ----------- | ---------------------------------------------------------------------------------------------- | ---------------------------- |
| **project-manager** 📋       | Sonnet | Init        | Project bootstrap, metadata interview, mode selection                                          | `/alfred:0-project`          |
| **spec-builder** 🏗️          | Sonnet | Plan        | Plan board consolidation, EARS-based SPEC authoring                                            | `/alfred:1-plan`             |
| **code-builder pipeline** 💎 | Sonnet | Run         | Phase 1 `implementation-planner` → Phase 2 `tdd-implementer` to execute RED → GREEN → REFACTOR | `/alfred:2-run`              |
| **doc-syncer** 📖            | Haiku  | Sync        | Living documentation, README/CHANGELOG updates                                                 | `/alfred:3-sync`             |
| **tag-agent** 🏷️             | Haiku  | Sync        | TAG inventory, orphan detection, chain repair                                                  | `@agent-tag-agent`           |
| **git-manager** 🚀           | Haiku  | Plan · Sync | GitFlow automation, Draft→Ready PR, auto-merge policy                                          | `@agent-git-manager`         |
| **debug-helper** 🔍          | Sonnet | Run         | Failure diagnosis, fix-forward guidance                                                        | `@agent-debug-helper`        |
| **trust-checker** ✅         | Haiku  | All phases  | TRUST 5 principle enforcement and risk flags                                                   | `@agent-trust-checker`       |
| **quality-gate** 🛡️          | Haiku  | Sync        | Coverage delta review, release gate validation                                                 | Auto during `/alfred:3-sync` |
| **cc-manager** 🛠️            | Sonnet | Ops         | Claude Code session tuning, Skill lifecycle management                                         | `@agent-cc-manager`          |

The **code-builder pipeline** runs two Sonnet specialists in sequence: **implementation-planner** (strategy, libraries, TAG design) followed by **tdd-implementer** (RED → GREEN → REFACTOR execution).

---

## Zero-project Specialists

| Sub-agent                 | Model  | Focus                                                       | Trigger                         |
| ------------------------- | ------ | ----------------------------------------------------------- | ------------------------------- |
| **language-detector** 🔍   | Haiku  | Stack detection, language matrix                            | Auto during `/alfred:0-project` |
| **backup-merger** 📦       | Sonnet | Backup restore, checkpoint diff                             | `@agent-backup-merger`          |
| **project-interviewer** 💬 | Sonnet | Requirement interviews, persona capture                     | `/alfred:0-project` Q&A         |
| **document-generator** 📝  | Haiku  | Project docs seed (`product.md`, `structure.md`, `tech.md`) | `/alfred:0-project`             |
| **feature-selector** 🎯    | Haiku  | Skill pack recommendation                                   | `/alfred:0-project`             |
| **template-optimizer** ⚙️  | Haiku  | Template cleanup, migration helpers                         | `/alfred:0-project`             |

> **Implementation Note**: Zero-project specialists may be embedded within other agents (e.g., functionality within `project-manager`) or implemented as dedicated Skills (e.g., `moai-alfred-language-detection`). For example, `language-detector` functionality is provided by the `moai-alfred-language-detection` Skill during `/alfred:0-project` initialization.

---

## Built-in Claude Agents

| Agent               | Model  | Specialty                                     | Invocation       |
| ------------------- | ------ | --------------------------------------------- | ---------------- |
| **Explore** 🔍       | Haiku  | Repository-wide search & architecture mapping | `@agent-Explore` |
| **general-purpose** | Sonnet | General assistance                            | Automatic        |

### Explore Agent Guide

The **Explore** agent excels at navigating large codebases.

**Use cases**:
- ✅ **Code analysis** (understand complex implementations, trace dependencies, study architecture)
- ✅ Search for specific keywords or patterns (e.g., "API endpoints", "authentication logic")
- ✅ Locate files (e.g., `src/components/**/*.tsx`)
- ✅ Understand codebase structure (e.g., "explain the project architecture")
- ✅ Search across many files (Glob + Grep patterns)

**Recommend Explore when**:
- 🔍 You need to understand a complex structure
- 🔍 The implementation spans multiple files
- 🔍 You want the end-to-end flow of a feature
- 🔍 Dependency relationships must be analyzed
- 🔍 You're planning a refactor and need impact analysis

**Usage**: Use `Task(subagent_type="Explore", ...)` for deep codebase analysis. Declare `thoroughness: quick|medium|very thorough` in the prompt.

**Examples**:
- Deep analysis: "Analyze TemplateProcessor class and its dependencies" (thoroughness: very thorough)
- Domain search: "Find all AUTH-related files in SPEC/tests/src/docs" (thoroughness: medium)
- Natural language: "Where is JWT authentication implemented?" → Alfred auto-delegates

---

## Claude Skills (55 packs)

Alfred relies on 55 Claude Skills grouped by tier. Skills load via Progressive Disclosure: metadata is available at session start, full `SKILL.md` content loads when a sub-agent references it, and supporting templates stream only when required.

**Skills Distribution by Tier**:

| Tier            | Count  | Purpose                                      |
| --------------- | ------ | -------------------------------------------- |
| Foundation      | 6      | Core TRUST/TAG/SPEC/Git/EARS/Lang principles |
| Essentials      | 4      | Debug/Perf/Refactor/Review workflows         |
| Alfred          | 11     | Internal workflow orchestration              |
| Domain          | 10     | Specialized domain expertise                 |
| Language        | 23     | Language-specific best practices             |
| Claude Code Ops | 1      | Session management                           |
| **Total**       | **55** | Complete knowledge capsule library           |

**Foundation Tier (6)**: `moai-foundation-trust`, `moai-foundation-tags`, `moai-foundation-specs`, `moai-foundation-ears`, `moai-foundation-git`, `moai-foundation-langs` (TRUST/TAG/SPEC/EARS/Git/language detection)

**Essentials Tier (4)**: `moai-essentials-debug`, `moai-essentials-perf`, `moai-essentials-refactor`, `moai-essentials-review` (Debug/Perf/Refactor/Review workflows)

**Alfred Tier (11)**: `moai-alfred-code-reviewer`, `moai-alfred-debugger-pro`, `moai-alfred-ears-authoring`, `moai-alfred-git-workflow`, `moai-alfred-language-detection`, `moai-alfred-performance-optimizer`, `moai-alfred-refactoring-coach`, `moai-alfred-spec-metadata-validation`, `moai-alfred-tag-scanning`, `moai-alfred-trust-validation`, `moai-alfred-interactive-questions` (code review, debugging, EARS, Git, language detection, performance, refactoring, metadata, TAG scanning, trust validation, interactive questions)

**Domain Tier (10)** — `moai-domain-backend`, `web-api`, `frontend`, `mobile-app`, `security`, `devops`, `database`, `data-science`, `ml`, `cli-tool`.

**Language Tier (23)** — Python, TypeScript, Go, Rust, Java, Kotlin, Swift, Dart, C/C++, C#, Scala, Haskell, Elixir, Clojure, Lua, Ruby, PHP, JavaScript, SQL, Shell, Julia, R, plus supporting stacks.

**Claude Code Ops (1)** — `moai-claude-code` manages session settings, output styles, and Skill deployment.

Skills keep the core knowledge lightweight while allowing Alfred to assemble the right expertise for each request.

---

## Agent Collaboration Principles

- **Command precedence**: Command instructions outrank agent guidelines; follow the command if conflicts occur.
- **Single responsibility**: Each agent handles only its specialty.
- **Zero overlapping ownership**: When unsure, hand off to the agent with the most direct expertise.
- **Confidence reporting**: Always share confidence levels and identified risks when completing a task.
- **Escalation path**: When blocked, escalate to Alfred with context, attempted steps, and suggested next actions.

---

## Model Selection Guide

| Model                 | Primary use cases                                                    | Representative sub-agents                                                              | Why it fits                                                    |
| --------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| **Claude 4.5 Haiku**  | Documentation sync, TAG inventory, Git automation, rule-based checks | doc-syncer, tag-agent, git-manager, trust-checker, quality-gate, Explore               | Fast, deterministic output for patterned or string-heavy work  |
| **Claude 4.5 Sonnet** | Planning, implementation, troubleshooting, session ops               | Alfred, project-manager, spec-builder, code-builder pipeline, debug-helper, cc-manager | Deep reasoning, multi-step synthesis, creative problem solving |

**Guidelines**:
- Default to **Haiku** when the task is pattern-driven or requires rapid iteration; escalate to **Sonnet** for novel design, architecture, or ambiguous problem solving.
- Record any manual model switch in the task notes (who, why, expected benefit).
- Combine both models when needed: e.g., Sonnet plans a refactor, Haiku formats and validates the resulting docs.

---

## Agent 선택 결정 트리

| 상황 | 추천 Agent | 이유 |
|------|-----------|------|
| 코드베이스 이해 필요 | **Explore** | 대규모 프로젝트 빠른 분석 전문. Glob + Grep 패턴으로 전체 구조 파악 |
| 새 기능의 SPEC 작성 | **spec-builder** | EARS 문법 + SPEC 구조 전문가. YAML 메타데이터 + HISTORY 자동 관리 |
| 버그 원인 분석 | **debug-helper** | 스택 트레이스 + 에러 패턴 분석 전문. fix-forward vs rollback 권장 |
| 코드 구현 (TDD) | **code-builder pipeline** | RED → GREEN → REFACTOR 자동화. implementation-planner + tdd-implementer 순차 실행 |
| 문서 동기화 필요 | **doc-syncer** | Living Document 자동화. README + CHANGELOG 생성, TAG 체인 검증 |
| Git/PR 관리 | **git-manager** | GitFlow + Draft→Ready 자동화. feature 브랜치 + PR 생성 |
| 버전 배포 | **git-manager** | 릴리즈 자동화. CHANGELOG 생성 + 태그 생성 + PR merge |
| TAG 무결성 확인 | **tag-agent** | TAG 체인 검증 전문. orphan TAG 탐지 + 수정 권장 |
| 코드 품질 검증 | **trust-checker** | TRUST 5 원칙 검증. Test/Readable/Unified/Secured/Trackable 체크 |
| 릴리즈 게이트 검증 | **quality-gate** | Coverage delta + 보안 스캔. 릴리즈 전 최종 검증 |
| 프로젝트 초기화 | **project-manager** | 메타데이터 인터뷰 + mode 선택. `/alfred:0-project` 전담 |
| Claude Code 세션 관리 | **cc-manager** | Skill lifecycle + 출력 스타일 관리. 세션 튜닝 전문 |

---

**활용 예시**:
- "사용자가 '로그인 기능 추가'를 요청" → **spec-builder** (SPEC 작성) → **code-builder pipeline** (구현) → **doc-syncer** (문서화)
- "테스트가 실패함" → **debug-helper** (원인 분석) → **code-builder pipeline** (수정) → **trust-checker** (품질 재검증)
- "릴리즈 준비" → **quality-gate** (최종 검증) → **git-manager** (PR merge + 태그)

---

**마지막 업데이트**: 2025-10-26
**문서 버전**: v1.0.0 (Option A Refactoring)
