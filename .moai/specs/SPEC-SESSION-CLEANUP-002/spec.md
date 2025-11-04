# SPEC-SESSION-CLEANUP-002: Alfred 커맨드 파일 실제 구현 - Session Cleanup 패턴 적용

<!-- @SPEC:SESSION-CLEANUP-002 -->

---

## YAML Frontmatter

```yaml
id: SESSION-CLEANUP-002
title: Alfred 커맨드 파일 실제 구현 - Session Cleanup 패턴 적용
category: Implementation
priority: high
status: closed
author: "@GoosLab"
created: 2025-10-30
updated: 2025-10-30
version: 0.0.1
tags:
  - alfred
  - implementation
  - ux
  - session-management
dependencies:
  - SPEC-SESSION-CLEANUP-001
related_specs: []
traceability:
  parent: SPEC-SESSION-CLEANUP-001
  children: []
affected_components:
  - src/moai_adk/templates/.claude/commands/alfred-0-project.md
  - src/moai_adk/templates/.claude/commands/alfred-1-plan.md
  - src/moai_adk/templates/.claude/commands/alfred-2-run.md
  - src/moai_adk/templates/.claude/commands/alfred-3-sync.md
  - src/moai_adk/templates/.claude/agents/agent-alfred.md
risk_level: medium
review_status: pending
scope:
  packages:
    - src/moai_adk/templates/.claude/commands/
    - src/moai_adk/templates/.claude/agents/
  files:
    - alfred-0-project.md
    - alfred-1-plan.md
    - alfred-2-run.md
    - alfred-3-sync.md
    - agent-alfred.md
```

---

## HISTORY

| Version | Date       | Author    | Changes                                          |
| ------- | ---------- | --------- | ------------------------------------------------ |
| 0.0.1   | 2025-10-30 | @GoosLab  | Phase 2 SPEC creation - Implementation           |

---

## Environment

### Business Context

**Phase 1 완료 상태**: SPEC-SESSION-CLEANUP-001은 Alfred 커맨드 완료 후 세션 정리 및 다음 단계 안내 프레임워크를 문서화했습니다. Phase 2는 이 프레임워크를 실제 커맨드 파일과 에이전트 파일에 구현하는 단계입니다.

**목표**:
- ✅ 4개 Alfred 커맨드 파일에 AskUserQuestion 패턴 적용
- ✅ agent-alfred.md에 세션 정리 로직 추가
- ✅ Prose 제안 패턴 완전 제거
- ✅ 일관된 UX 확립

**Business Impact**:
- **사용자 경험 개선**: 모든 커맨드에서 동일한 인터랙션 패턴
- **학습 곡선 감소**: 예측 가능한 다음 단계 옵션
- **생산성 향상**: 명확한 세션 경계 및 작업 요약

### Technical Context

**구현 범위**:
```
src/moai_adk/templates/.claude/
├── commands/
│   ├── alfred-0-project.md    ← 3개 옵션 추가
│   ├── alfred-1-plan.md       ← 3개 옵션 추가
│   ├── alfred-2-run.md        ← 3개 옵션 추가
│   └── alfred-3-sync.md       ← 3개 옵션 추가
└── agents/
    └── agent-alfred.md        ← 세션 정리 로직 추가
```

**기술 스택**:
- Markdown 템플릿 (Claude Code command format)
- Python code blocks (AskUserQuestion syntax)
- YAML frontmatter (metadata)

**통합 포인트**:
- `moai-alfred-interactive-questions` 스킬 (TUI 인터랙션)
- TodoWrite 시스템 (작업 추적)
- Git workflow (branch 전략, commit 패턴)

### Stakeholders

- **Primary**: MoAI-ADK 개발자 (커맨드 파일 수정)
- **Secondary**: Alfred 사용자 (일관된 UX 경험)
- **Technical**: tdd-implementer, doc-syncer, git-manager (구현 담당)

---

## Assumptions

### User Assumptions

- **ASM-IMPL-001**: 개발자는 4개 커맨드 파일을 수정할 권한이 있다
- **ASM-IMPL-002**: 사용자는 AskUserQuestion 패턴이 적용된 인터랙션을 선호한다
- **ASM-IMPL-003**: 세션 요약이 Markdown 형식으로 출력되는 것을 예상한다

### Technical Assumptions

- **ASM-IMPL-004**: AskUserQuestion tool은 모든 커맨드 실행 컨텍스트에서 사용 가능하다
- **ASM-IMPL-005**: Markdown 템플릿의 Python 코드 블록은 실행 시 올바르게 파싱된다
- **ASM-IMPL-006**: TodoWrite 상태는 커맨드 간 세션 유지된다
- **ASM-IMPL-007**: 기존 커맨드 구조와 새 패턴은 호환된다

### Business Assumptions

- **ASM-IMPL-008**: 일관된 완료 패턴은 사용자 만족도를 높인다
- **ASM-IMPL-009**: Prose 제안 제거는 UX 혼란을 감소시킨다
- **ASM-IMPL-010**: 명확한 세션 경계는 장기 사용 시 생산성을 향상시킨다

---

## Requirements

### Ubiquitous Requirements (항상 적용)

- **REQ-IMPL-001**: 시스템은 각 Alfred 커맨드 완료 시 **반드시** AskUserQuestion을 호출해야 한다
  - **Rationale**: 일관된 UX, 사용자 선택권 보장
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-001`

- **REQ-IMPL-002**: 시스템은 모든 커맨드 완료 시 prose 형태의 제안을 **금지**해야 한다
  - **Rationale**: "You can now run..." 패턴 제거
  - **Priority**: MUST NOT
  - **@TAG**: `@REQ:IMPL-002`

- **REQ-IMPL-003**: 시스템은 커맨드 완료 전 TodoWrite를 정리해야 한다
  - **Rationale**: 세션 컨텍스트 명확화, 다음 커맨드 clean state 보장
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-003`

### Event-driven Requirements (특정 이벤트 발생 시)

- **REQ-IMPL-004**: WHEN `/alfred:0-project` 완료 시, 시스템은 3개 옵션을 제시해야 한다
  - Option 1: 📋 스펙 작성 진행 (`/alfred:1-plan` 실행)
  - Option 2: 🔍 프로젝트 구조 검토 (현재 상태 확인)
  - Option 3: 🔄 새 세션 시작 (`/clear` 실행)
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-004`

- **REQ-IMPL-005**: WHEN `/alfred:1-plan` 완료 시, 시스템은 3개 옵션을 제시해야 한다
  - Option 1: 🚀 구현 진행 (`/alfred:2-run SPEC-XXX-001` 실행)
  - Option 2: ✏️ SPEC 수정 (현재 SPEC 재작업)
  - Option 3: 🔄 새 세션 시작 (`/clear` 실행)
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-005`

- **REQ-IMPL-006**: WHEN `/alfred:2-run` 완료 시, 시스템은 3개 옵션을 제시해야 한다
  - Option 1: 📚 문서 동기화 (`/alfred:3-sync` 실행)
  - Option 2: 🧪 추가 테스트/검증 (테스트 재실행)
  - Option 3: 🔄 새 세션 시작 (`/clear` 실행)
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-006`

- **REQ-IMPL-007**: WHEN `/alfred:3-sync` 완료 시, 시스템은 3개 옵션을 제시해야 한다
  - Option 1: 📋 다음 기능 계획 (`/alfred:1-plan` 실행)
  - Option 2: 🔀 PR 병합 (main 브랜치로 병합)
  - Option 3: ✅ 세션 완료 (작업 종료)
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-007`

- **REQ-IMPL-008**: WHEN 사용자가 "새 세션" 또는 "세션 완료" 선택 시, 시스템은 세션 요약을 생성해야 한다
  - 완료된 작업 목록 (TodoWrite 기반)
  - 생성된 커밋 수 (git log 기반)
  - 변경된 파일 목록 (git diff 기반)
  - 다음 권장 작업 (optional)
  - **Priority**: SHOULD
  - **@TAG**: `@REQ:IMPL-008`

### State-driven Requirements (특정 상태일 때)

- **REQ-IMPL-009**: WHILE 세션이 활성 상태일 때, 시스템은 TodoWrite 상태를 유지해야 한다
  - 모든 작업은 `pending` → `in_progress` → `completed` 순서
  - 정확히 1개의 작업만 `in_progress` 상태 (parallel 승인 제외)
  - **Priority**: MUST
  - **@TAG**: `@REQ:IMPL-009`

- **REQ-IMPL-010**: WHILE 커맨드 실행 중일 때, 시스템은 모든 `completed` 작업을 추출 및 기록해야 한다
  - AskUserQuestion 호출 직전 실행
  - 세션 요약 생성 시 사용
  - **Priority**: SHOULD
  - **@TAG**: `@REQ:IMPL-010`

### Optional Requirements (선택적)

- **REQ-IMPL-011**: 시스템은 세션 메타데이터를 `.moai/memory/session-history.json`에 저장할 수 있다
  - 세션 시작/종료 시간
  - 실행된 커맨드 목록
  - 생성된 SPEC ID
  - **Priority**: MAY
  - **@TAG**: `@REQ:IMPL-011`

### Unwanted Behaviors (금지된 동작)

- **REQ-IMPL-012**: 시스템은 AskUserQuestion 없이 커맨드를 종료해서는 **안 된다**
  - **Rationale**: 사용자에게 다음 단계 선택권 제공 필수
  - **Priority**: MUST NOT
  - **@TAG**: `@REQ:IMPL-012`

- **REQ-IMPL-013**: 시스템은 옵션을 4개 이상 제시해서는 **안 된다**
  - **Rationale**: UX 혼란 방지, 선택 피로 감소
  - **Priority**: SHOULD NOT
  - **@TAG**: `@REQ:IMPL-013`

---

## Specifications

### Design Decisions

#### 1. AskUserQuestion 템플릿 구조

**Decision**: 모든 커맨드 파일의 마지막 섹션에 "Final Step: Next Action Selection" 추가

**Implementation Template**:
```markdown
## Final Step: Next Action Selection

After [command action] completes, use AskUserQuestion tool:

```python
AskUserQuestion(
    questions=[
        {
            "question": "[Command] 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {"label": "📋 Option 1", "description": "Action description"},
                {"label": "🚀 Option 2", "description": "Action description"},
                {"label": "🔄 Option 3", "description": "Action description"}
            ]
        }
    ]
)
```

**Prohibited**: Never suggest next steps in prose (e.g., "You can now run `/alfred:X`...")
```

**Rationale**:
- 일관된 섹션 이름 (`Final Step`)
- Python 코드 블록으로 명확한 구문
- Prohibited 경고로 prose 패턴 방지

#### 2. 커맨드별 옵션 정의

**`/alfred:0-project` 완료 옵션**:
```python
{
    "question": "프로젝트 초기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
    "header": "다음 단계",
    "options": [
        {"label": "📋 스펙 작성 진행", "description": "/alfred:1-plan 실행"},
        {"label": "🔍 프로젝트 구조 검토", "description": "현재 상태 확인"},
        {"label": "🔄 새 세션 시작", "description": "/clear 실행"}
    ]
}
```

**`/alfred:1-plan` 완료 옵션**:
```python
{
    "question": "SPEC 작성이 완료되었습니다. 다음으로 뭘 하시겠습니까?",
    "header": "다음 단계",
    "options": [
        {"label": "🚀 구현 진행", "description": "/alfred:2-run SPEC-XXX-001 실행"},
        {"label": "✏️ SPEC 수정", "description": "현재 SPEC 재작업"},
        {"label": "🔄 새 세션 시작", "description": "/clear 실행"}
    ]
}
```

**`/alfred:2-run` 완료 옵션**:
```python
{
    "question": "구현이 완료되었습니다. 다음으로 뭘 하시겠습니까?",
    "header": "다음 단계",
    "options": [
        {"label": "📚 문서 동기화", "description": "/alfred:3-sync 실행"},
        {"label": "🧪 추가 테스트/검증", "description": "테스트 재실행"},
        {"label": "🔄 새 세션 시작", "description": "/clear 실행"}
    ]
}
```

**`/alfred:3-sync` 완료 옵션**:
```python
{
    "question": "문서 동기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
    "header": "다음 단계",
    "options": [
        {"label": "📋 다음 기능 계획", "description": "/alfred:1-plan 실행"},
        {"label": "🔀 PR 병합", "description": "main 브랜치로 병합"},
        {"label": "✅ 세션 완료", "description": "작업 종료"}
    ]
}
```

#### 3. TodoWrite Cleanup Protocol

**Decision**: AskUserQuestion 호출 직전 TodoWrite에서 `completed` 작업 추출

**Pseudocode**:
```python
# Step 1: Extract completed tasks
completed_tasks = [task for task in todos if task.status == "completed"]

# Step 2: Store in session context
session_context = {
    "completed_tasks": completed_tasks,
    "command": "/alfred:X-command",
    "timestamp": datetime.now()
}

# Step 3: Generate summary if user selects "새 세션" or "세션 완료"
if user_choice in ["🔄 새 세션 시작", "✅ 세션 완료"]:
    generate_session_summary(session_context)
```

**Implementation Location**: `agent-alfred.md` (새 섹션 추가)

#### 4. Session Summary Generator

**Decision**: 세션 종료 시 Markdown 형식 요약 생성

**Output Format**:
```markdown
## 🎊 세션 요약

### 완료된 작업
- ✅ [Task 1 from TodoWrite]
- ✅ [Task 2 from TodoWrite]
- ✅ [Task 3 from TodoWrite]

### Git 통계
- 📝 생성된 커밋: X개
- 📂 변경된 파일: Y개
- ➕ 추가된 라인: +Z
- ➖ 삭제된 라인: -W

### 다음 권장 작업
1. [Recommendation based on current state]
2. [Optional follow-up action]
```

**Implementation Location**: `agent-alfred.md` (세션 정리 섹션)

### Non-Functional Requirements

- **NFR-IMPL-001**: AskUserQuestion 호출은 500ms 이내에 완료되어야 한다
  - **Measurement**: Response time < 500ms
  - **@TAG**: `@NFR:IMPL-001`

- **NFR-IMPL-002**: 세션 요약 생성은 1초 이내에 완료되어야 한다
  - **Measurement**: Generation time < 1000ms
  - **@TAG**: `@NFR:IMPL-002`

- **NFR-IMPL-003**: TodoWrite 상태 변경은 즉시 반영되어야 한다
  - **Measurement**: State change latency < 100ms
  - **@TAG**: `@NFR:IMPL-003`

### Constraints

- **CON-IMPL-001**: AskUserQuestion은 최대 4개 질문까지 batched 가능
  - **Rationale**: TUI 인터랙션 제한
  - **@TAG**: `@CON:IMPL-001`

- **CON-IMPL-002**: 각 옵션은 3개로 고정 (예외적으로 4개 허용)
  - **Rationale**: UX 혼란 방지, 선택 피로 감소
  - **@TAG**: `@CON:IMPL-002`

- **CON-IMPL-003**: 세션 요약은 Markdown 형식으로만 출력
  - **Rationale**: 일관된 출력 포맷, 파싱 용이성
  - **@TAG**: `@CON:IMPL-003`

- **CON-IMPL-004**: Markdown 템플릿 파일은 `.claude/commands/` 디렉토리에 위치
  - **Rationale**: MoAI-ADK 표준 구조
  - **@TAG**: `@CON:IMPL-004`

---

## Traceability

### Parent Requirements

- **@SPEC:SESSION-CLEANUP-001**: Phase 1 - Alfred 커맨드 완료 후 세션 정리 및 다음 단계 안내 프레임워크
- **@SPEC:ALF-WORKFLOW-001**: Alfred 4-Step Workflow (Intent → Plan → Execute → Report)

### Child Requirements

- **IMPL-001**: `/alfred:0-project` 완료 후 AskUserQuestion 호출 검증
- **IMPL-002**: `/alfred:1-plan` 완료 후 AskUserQuestion 호출 검증
- **IMPL-003**: `/alfred:2-run` 완료 후 AskUserQuestion 호출 검증
- **IMPL-004**: `/alfred:3-sync` 완료 후 AskUserQuestion 호출 검증
- **IMPL-005**: 세션 요약 생성 (Markdown 형식 출력)
- **IMPL-006**: TodoWrite 정리 (completed 작업 추출)
- **IMPL-007**: Prose 패턴 검색 (검출 0건)
- **IMPL-008**: Batched 디자인 (호출 횟수 = 1)

### Related Components

- `src/moai_adk/templates/.claude/commands/alfred-0-project.md` → `@CODE:CMD-0-PROJECT-IMPL`
- `src/moai_adk/templates/.claude/commands/alfred-1-plan.md` → `@CODE:CMD-1-PLAN-IMPL`
- `src/moai_adk/templates/.claude/commands/alfred-2-run.md` → `@CODE:CMD-2-RUN-IMPL`
- `src/moai_adk/templates/.claude/commands/alfred-3-sync.md` → `@CODE:CMD-3-SYNC-IMPL`
- `src/moai_adk/templates/.claude/agents/agent-alfred.md` → `@CODE:AGENT-ALFRED-IMPL`
- `moai-alfred-interactive-questions` skill → `@SKILL:INTERACTIVE-QUESTIONS`

### Test Cases

See acceptance.md for detailed test scenarios:
- 8 primary test cases (TEST-001 to TEST-008)
- Each covers 1 command or 1 quality verification
- All use Given-When-Then format

---

## Risks & Mitigation

| Risk                                      | Impact | Probability | Mitigation                                                 |
| ----------------------------------------- | ------ | ----------- | ---------------------------------------------------------- |
| AskUserQuestion 호출 실패                 | High   | Low         | Fallback: 임시 prose 메시지 (에러 로그 기록)               |
| Markdown 템플릿 파싱 에러                 | High   | Low         | Syntax validation 추가, TDD 테스트 작성                     |
| TodoWrite 상태 불일치                     | Medium | Medium      | Pre-completion validation: 모든 작업 completed 확인        |
| 세션 요약 생성 지연                        | Low    | Low         | Async 처리, 타임아웃 설정 (1초)                            |
| Batched AskUserQuestion 지원 불가         | Medium | Low         | Sequential fallback 구현                                   |
| 기존 워크플로우 호환성 문제                | Medium | Low         | Backward compatibility test 추가                           |
| Prose 패턴 재도입 (개발자 실수)            | Low    | Medium      | Automated grep search in CI/CD                             |

---

## Open Questions

1. **Q1**: 세션 메타데이터를 `.moai/memory/session-history.json`에 저장할 것인가?
   - **Status**: To be decided (Phase 3에서 논의)
   - **Owner**: @GoosLab

2. **Q2**: 세션 요약을 파일로 저장할 것인가, 아니면 출력만 할 것인가?
   - **Status**: **출력만** (파일 저장은 Optional - Phase 3)
   - **Owner**: @GoosLab

3. **Q3**: 각 커맨드별로 옵션 개수를 3개로 고정할 것인가?
   - **Status**: **3개 고정** (예외적으로 4개 허용)
   - **Owner**: @GoosLab

4. **Q4**: AskUserQuestion 호출 실패 시 어떻게 처리할 것인가?
   - **Status**: **Fallback to prose** (에러 로그 + 다음 커맨드 제안)
   - **Owner**: debug-helper

---

## Version Control

- **Current Version**: 0.0.1 (Phase 2 SPEC draft)
- **Status**: Draft
- **Next Review**: 2025-10-31
- **Approval Required**: @GoosLab
- **Implementation Target**: v0.8.0 of MoAI-ADK

---

## Implementation Checklist

### Phase 2A: 커맨드 파일 수정 (4 files)
- [ ] `alfred-0-project.md` - Final Step 섹션 추가
- [ ] `alfred-1-plan.md` - Final Step 섹션 추가
- [ ] `alfred-2-run.md` - Final Step 섹션 추가
- [ ] `alfred-3-sync.md` - Final Step 섹션 추가

### Phase 2B: 에이전트 파일 업데이트 (1 file)
- [ ] `agent-alfred.md` - TodoWrite 정리 로직 추가
- [ ] `agent-alfred.md` - 세션 요약 생성 로직 추가

### Phase 2C: 테스트 시나리오 실행 (8 tests)
- [ ] TEST-001: `/alfred:0-project` AskUserQuestion 검증
- [ ] TEST-002: `/alfred:1-plan` AskUserQuestion 검증
- [ ] TEST-003: `/alfred:2-run` AskUserQuestion 검증
- [ ] TEST-004: `/alfred:3-sync` AskUserQuestion 검증
- [ ] TEST-005: 세션 요약 생성 검증
- [ ] TEST-006: TodoWrite 정리 검증
- [ ] TEST-007: Prose 패턴 검색 (0건)
- [ ] TEST-008: Batched 디자인 (1회 호출)

### Phase 2D: Git 커밋 및 문서화
- [ ] 변경사항 커밋 (branch: `feature/SPEC-SESSION-CLEANUP-002`)
- [ ] CHANGELOG.md 업데이트
- [ ] CLAUDE.md 검증
status: closed
---

**End of SPEC-SESSION-CLEANUP-002**
