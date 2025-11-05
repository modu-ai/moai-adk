# SPEC-SESSION-CLEANUP-001: Alfred 커맨드 완료 후 세션 정리 및 다음 단계 안내 프레임워크

<!-- @SPEC:SESSION-CLEANUP-001 -->

---

## YAML Frontmatter

```yaml
id: SESSION-CLEANUP-001
title: Alfred 커맨드 완료 후 세션 정리 및 다음 단계 안내 프레임워크
category: Enhancement
priority: high
status: completed
author: "@GoosLab"
created: 2025-10-30
updated: 2025-11-05
version: 1.0.0
related_issue: "138"
tags:
  - alfred
  - workflow
  - ux
  - session-management
dependencies:
  - SPEC-ALF-WORKFLOW-001
related_specs: []
traceability:
  parent: null
  children: []
affected_components:
  - .claude/commands/alfred-0-project.md
  - .claude/commands/alfred-1-plan.md
  - .claude/commands/alfred-2-run.md
  - .claude/commands/alfred-3-sync.md
  - .claude/agents/agent-alfred.md
risk_level: medium
review_status: pending
```

---

## HISTORY

| Version | Date       | Author    | Changes                                          |
| ------- | ---------- | --------- | ------------------------------------------------ |
| 0.1.0   | 2025-10-30 | @GoosLab  | Initial SPEC creation                            |

---

## Environment

### Business Context

Alfred는 MoAI-ADK의 SuperAgent로서 4개의 핵심 커맨드(`/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`)를 통해 SPEC → TDD → Sync 워크플로우를 오케스트레이션합니다. 현재 커맨드 완료 후 다음 단계 안내가 일관되지 않아 사용자 경험에 혼란이 발생하고 있습니다.

**문제점**:
- 커맨드 완료 시 prose로 제안하는 경우가 있음 ("You can now run `/alfred:1-plan`...")
- AskUserQuestion 사용이 불일치함
- TodoWrite 정리가 누락되는 경우가 있음
- 세션 컨텍스트가 다음 커맨드로 전달되지 않음

### Technical Context

**현재 아키텍처**:
- Alfred (SuperAgent) → 4 Commands → 10 Sub-agents → 55 Skills
- AskUserQuestion tool: `moai-alfred-ask-user-questions` 스킬 기반 TUI 인터랙션
- TodoWrite: 작업 진행 상황 추적
- Task() 호출: Sub-agent 간 컨텍스트 전달

**영향받는 컴포넌트**:
- `.claude/commands/alfred-*.md` (4 files)
- `.claude/agents/agent-alfred.md` (1 file)
- `CLAUDE.md` (documentation)

### Stakeholders

- **Primary**: MoAI-ADK 사용자 (개발자)
- **Secondary**: Alfred Sub-agents (워크플로우 일관성)
- **Technical**: GoosLab (시스템 아키텍트)

---

## Assumptions

### User Assumptions

- **ASM-SESSION-001**: 사용자는 커맨드 완료 후 명확한 다음 단계 옵션을 원한다
- **ASM-SESSION-002**: 사용자는 자유 텍스트 입력보다 선택지를 선호한다 (3-4 options)
- **ASM-SESSION-003**: 사용자는 세션 종료 시 작업 요약을 확인하길 원한다

### Technical Assumptions

- **ASM-SESSION-004**: AskUserQuestion tool은 1-4개의 질문을 batched로 처리할 수 있다
- **ASM-SESSION-005**: TodoWrite는 모든 커맨드 실행 중 유지된다
- **ASM-SESSION-006**: 각 커맨드는 독립적으로 완료 패턴을 구현할 수 있다

### Business Assumptions

- **ASM-SESSION-007**: 일관된 UX는 사용자 학습 곡선을 감소시킨다
- **ASM-SESSION-008**: 명확한 세션 경계는 생산성을 향상시킨다

---

## Requirements

### Ubiquitous Requirements

- **REQ-SESSION-001**: Alfred는 모든 커맨드 완료 시 **반드시** AskUserQuestion을 사용하여 다음 단계를 물어야 한다
  - **Rationale**: 일관된 UX, prose 제안 금지
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-001`

- **REQ-SESSION-002**: Alfred는 커맨드 완료 시 TodoWrite를 정리해야 한다
  - **Rationale**: 세션 컨텍스트 명확화
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-002`

### Event-driven Requirements

- **REQ-SESSION-003**: WHEN `/alfred:0-project` 완료 시, 시스템은 3가지 옵션을 제시해야 한다
  - Option 1: 📋 스펙 작성 진행 (`/alfred:1-plan` 실행)
  - Option 2: 🔍 프로젝트 구조 검토 (현재 상태 확인)
  - Option 3: 🔄 새 세션 시작 (`/clear` 실행)
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-003`

- **REQ-SESSION-004**: WHEN `/alfred:1-plan` 완료 시, 시스템은 3가지 옵션을 제시해야 한다
  - Option 1: 🚀 구현 진행 (`/alfred:2-run` 실행)
  - Option 2: ✏️ SPEC 수정 (현재 SPEC 재작업)
  - Option 3: 🔄 새 세션 시작 (`/clear` 실행)
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-004`

- **REQ-SESSION-005**: WHEN `/alfred:2-run` 완료 시, 시스템은 3가지 옵션을 제시해야 한다
  - Option 1: 📚 문서 동기화 (`/alfred:3-sync` 실행)
  - Option 2: 🧪 추가 테스트/검증 (테스트 재실행)
  - Option 3: 🔄 새 세션 시작 (`/clear` 실행)
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-005`

- **REQ-SESSION-006**: WHEN `/alfred:3-sync` 완료 시, 시스템은 3가지 옵션을 제시해야 한다
  - Option 1: 📋 다음 기능 계획 (`/alfred:1-plan` 실행)
  - Option 2: 🔀 PR 병합 (main 브랜치로 병합)
  - Option 3: ✅ 세션 완료 (작업 종료)
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-006`

- **REQ-SESSION-007**: WHEN 사용자가 "새 세션" 또는 "세션 완료" 선택 시, 시스템은 세션 요약을 생성해야 한다
  - 완료된 작업 목록
  - 생성된 커밋 수
  - 변경된 파일 목록
  - 다음 권장 작업 (optional)
  - **Priority**: SHOULD
  - **@TAG**: `@REQ:SESSION-007`

### State-driven Requirements

- **REQ-SESSION-008**: WHILE 커맨드 실행 중일 때, 시스템은 TodoWrite 상태를 유지해야 한다
  - 모든 작업은 `pending` → `in_progress` → `completed` 순서를 따름
  - 정확히 1개의 작업만 `in_progress` 상태 (parallel 승인 제외)
  - **Priority**: MUST
  - **@TAG**: `@REQ:SESSION-008`

- **REQ-SESSION-009**: WHILE 세션 정리 중일 때, 시스템은 모든 completed 작업을 기록해야 한다
  - TodoWrite에서 `completed` 상태의 모든 작업 추출
  - 세션 요약에 포함
  - **Priority**: SHOULD
  - **@TAG**: `@REQ:SESSION-009`

### Unwanted Behaviors

- **REQ-SESSION-010**: 시스템은 prose로 "You can now run..."과 같은 제안을 해서는 **안 된다**
  - **Rationale**: 일관성 없는 UX, AskUserQuestion 우회
  - **Priority**: MUST NOT
  - **@TAG**: `@REQ:SESSION-010`

- **REQ-SESSION-011**: 시스템은 AskUserQuestion 없이 커맨드를 종료해서는 **안 된다**
  - **Rationale**: 사용자에게 다음 단계 선택권 제공 필수
  - **Priority**: MUST NOT
  - **@TAG**: `@REQ:SESSION-011`

### Optional Requirements

- **REQ-SESSION-012**: 시스템은 세션 메타데이터를 `.moai/memory/session-history.json`에 저장할 수 있다
  - 세션 시작/종료 시간
  - 실행된 커맨드 목록
  - 생성된 SPEC ID
  - **Priority**: MAY
  - **@TAG**: `@REQ:SESSION-012`

---

## Specifications

### Design Decisions

#### 1. AskUserQuestion Batched Design

**Decision**: 모든 커맨드 완료 시 1개의 AskUserQuestion 호출로 다음 단계를 물어본다.

**Rationale**:
- 사용자 인터랙션 턴 감소 (UX 개선)
- 일관된 패턴 유지
- 3-4개 옵션으로 명확한 선택지 제공

**Implementation**:
```python
AskUserQuestion(
    questions=[
        {
            "question": "커맨드가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {"label": "📋 Option 1", "description": "..."},
                {"label": "🚀 Option 2", "description": "..."},
                {"label": "🔄 Option 3", "description": "..."}
            ]
        }
    ]
)
```

#### 2. TodoWrite Cleanup Protocol

**Decision**: 커맨드 완료 시 모든 `completed` 작업을 최종 요약에 포함하고 TodoWrite를 초기화한다.

**Rationale**:
- 세션 컨텍스트 명확화
- 다음 커맨드 실행 시 clean state 보장

**Implementation**:
- 커맨드 완료 직전: TodoWrite에서 모든 `completed` 작업 추출
- 세션 요약 생성 시 포함
- AskUserQuestion 호출 전 TodoWrite 상태 확인

#### 3. Session Summary Generation

**Decision**: 사용자가 "새 세션" 또는 "세션 완료" 선택 시 자동으로 요약 생성.

**Rationale**:
- 작업 기록 보존
- 다음 세션 컨텍스트 제공

**Content**:
- ✅ 완료된 작업 목록 (TodoWrite 기반)
- 📝 생성된 커밋 수 (git log 기반)
- 📂 변경된 파일 목록 (git diff 기반)
- 🚀 다음 권장 작업 (optional)

### Non-Functional Requirements

- **NFR-SESSION-001**: AskUserQuestion 호출은 500ms 이내에 완료되어야 한다
- **NFR-SESSION-002**: 세션 요약 생성은 1초 이내에 완료되어야 한다
- **NFR-SESSION-003**: TodoWrite 상태 변경은 즉시 반영되어야 한다

### Constraints

- **CON-SESSION-001**: AskUserQuestion은 최대 4개 질문까지 batched 가능
- **CON-SESSION-002**: 각 옵션은 3-5개로 제한 (UX 혼란 방지)
- **CON-SESSION-003**: 세션 요약은 Markdown 형식으로 출력

---

## Traceability

### Parent Requirements

- **@SPEC:ALF-WORKFLOW-001**: Alfred 4-Step Workflow (Intent → Plan → Execute → Report)

### Child Requirements

- (TBD: 구현 단계에서 생성될 테스트 케이스)

### Related Components

- `.claude/commands/alfred-0-project.md` → `@CODE:CMD-0-PROJECT`
- `.claude/commands/alfred-1-plan.md` → `@CODE:CMD-1-PLAN`
- `.claude/commands/alfred-2-run.md` → `@CODE:CMD-2-RUN`
- `.claude/commands/alfred-3-sync.md` → `@CODE:CMD-3-SYNC`
- `.claude/agents/agent-alfred.md` → `@CODE:AGENT-ALFRED`
- `moai-alfred-ask-user-questions` skill → `@SKILL:INTERACTIVE-QUESTIONS`

### Test Cases

See acceptance.md for detailed test case definitions:
- `/alfred:0-project` 완료 후 AskUserQuestion 호출 검증
- `/alfred:1-plan` 완료 후 AskUserQuestion 호출 검증
- `/alfred:2-run` 완료 후 AskUserQuestion 호출 검증
- `/alfred:3-sync` 완료 후 AskUserQuestion 호출 검증
- 세션 요약 생성 검증
- TodoWrite 정리 검증

---

## Risks & Mitigation

| Risk                                      | Impact | Mitigation                                                 |
| ----------------------------------------- | ------ | ---------------------------------------------------------- |
| AskUserQuestion 호출 실패                 | High   | Fallback: prose 메시지로 안내 (임시)                      |
| TodoWrite 상태 불일치                     | Medium | Pre-completion validation: 모든 작업 completed 확인        |
| 세션 요약 생성 지연                        | Low    | Async 처리, 타임아웃 설정 (1초)                            |
| Batched AskUserQuestion 지원 불가         | Medium | Sequential fallback 구현                                   |

---

## Open Questions

1. **Q1**: 세션 메타데이터를 `.moai/memory/session-history.json`에 저장할 것인가?
   - **Status**: To be decided
   - **Owner**: @GoosLab

2. **Q2**: 각 커맨드별로 옵션 개수를 3개로 고정할 것인가, 아니면 유연하게 할 것인가?
   - **Status**: **3-4개로 제한** (UX 최적화)
   - **Owner**: @GoosLab

3. **Q3**: 세션 요약을 파일로 저장할 것인가, 아니면 출력만 할 것인가?
   - **Status**: To be decided
   - **Owner**: @GoosLab

---

## Version Control

- **Current Version**: 0.1.0
- **Status**: Draft
- **Next Review**: 2025-10-31
- **Approval Required**: @GoosLab

---
