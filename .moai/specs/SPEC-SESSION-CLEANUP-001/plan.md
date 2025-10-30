# Implementation Plan: SPEC-SESSION-CLEANUP-001

<!-- @PLAN:SESSION-CLEANUP-001 -->

---

## Overview

이 구현 계획은 Alfred의 모든 커맨드(`/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`)에 일관된 세션 정리 및 다음 단계 안내 패턴을 적용하는 방법을 정의합니다.

**목표**:
- ✅ 모든 커맨드 완료 시 AskUserQuestion 패턴 적용
- ✅ TodoWrite 정리 프로토콜 구현
- ✅ 세션 요약 생성 기능 추가
- ✅ Prose 제안 금지 규칙 강제

**영향받는 컴포넌트**:
- 4개 커맨드 파일
- 1개 에이전트 파일 (agent-alfred.md)
- 1개 문서 파일 (CLAUDE.md)

---

## Implementation Phases

### Step 1: Command Completion Pattern 구현

**목표**: 4개 커맨드에 일관된 완료 패턴 추가

#### 1.1. `/alfred:0-project` 업데이트

**파일**: `.claude/commands/alfred-0-project.md`

**변경 사항**:
1. 커맨드 완료 직전에 AskUserQuestion 호출 추가
2. 옵션 3개 정의:
   - 📋 스펙 작성 진행 (`/alfred:1-plan`)
   - 🔍 프로젝트 구조 검토
   - 🔄 새 세션 시작 (`/clear`)

**템플릿**:
```markdown
## Final Step: Next Action Selection

After project initialization completes, use AskUserQuestion tool:

```python
AskUserQuestion(
    questions=[
        {
            "question": "프로젝트 초기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {"label": "📋 스펙 작성 진행", "description": "/alfred:1-plan 실행"},
                {"label": "🔍 프로젝트 구조 검토", "description": "현재 상태 확인"},
                {"label": "🔄 새 세션 시작", "description": "/clear 실행"}
            ]
        }
    ]
)
```

**Prohibited**: Never suggest next steps in prose (e.g., "You can now run `/alfred:1-plan`...")
```

**구현 위치**: 커맨드 마지막 섹션 (`## Success Criteria` 다음)

---

#### 1.2. `/alfred:1-plan` 업데이트

**파일**: `.claude/commands/alfred-1-plan.md`

**변경 사항**:
1. SPEC 생성 완료 후 AskUserQuestion 호출 추가
2. 옵션 3개 정의:
   - 🚀 구현 진행 (`/alfred:2-run`)
   - ✏️ SPEC 수정
   - 🔄 새 세션 시작

**템플릿**:
```markdown
## Final Step: Next Action Selection

After SPEC creation completes, use AskUserQuestion tool:

```python
AskUserQuestion(
    questions=[
        {
            "question": "SPEC 작성이 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {"label": "🚀 구현 진행", "description": "/alfred:2-run SPEC-XXX-001 실행"},
                {"label": "✏️ SPEC 수정", "description": "현재 SPEC 재작업"},
                {"label": "🔄 새 세션 시작", "description": "/clear 실행"}
            ]
        }
    ]
)
```
```

---

#### 1.3. `/alfred:2-run` 업데이트

**파일**: `.claude/commands/alfred-2-run.md`

**변경 사항**:
1. TDD 구현 완료 후 AskUserQuestion 호출 추가
2. 옵션 3개 정의:
   - 📚 문서 동기화 (`/alfred:3-sync`)
   - 🧪 추가 테스트/검증
   - 🔄 새 세션 시작

**템플릿**:
```markdown
## Final Step: Next Action Selection

After implementation completes, use AskUserQuestion tool:

```python
AskUserQuestion(
    questions=[
        {
            "question": "구현이 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {"label": "📚 문서 동기화", "description": "/alfred:3-sync 실행"},
                {"label": "🧪 추가 테스트/검증", "description": "테스트 재실행"},
                {"label": "🔄 새 세션 시작", "description": "/clear 실행"}
            ]
        }
    ]
)
```
```

---

#### 1.4. `/alfred:3-sync` 업데이트

**파일**: `.claude/commands/alfred-3-sync.md`

**변경 사항**:
1. 문서 동기화 완료 후 AskUserQuestion 호출 추가
2. 옵션 3개 정의:
   - 📋 다음 기능 계획 (`/alfred:1-plan`)
   - 🔀 PR 병합
   - ✅ 세션 완료

**템플릿**:
```markdown
## Final Step: Next Action Selection

After sync completes, use AskUserQuestion tool:

```python
AskUserQuestion(
    questions=[
        {
            "question": "문서 동기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {"label": "📋 다음 기능 계획", "description": "/alfred:1-plan 실행"},
                {"label": "🔀 PR 병합", "description": "main 브랜치로 병합"},
                {"label": "✅ 세션 완료", "description": "작업 종료"}
            ]
        }
    ]
)
```
```

---

### Step 2: 세션 정리 프로토콜 구현

**목표**: TodoWrite 정리 및 세션 요약 생성 로직 추가

#### 2.1. TodoWrite Cleanup Logic

**위치**: 각 커맨드의 AskUserQuestion 호출 직전

**로직**:
1. TodoWrite에서 모든 `completed` 작업 추출
2. 작업 목록을 세션 컨텍스트에 저장
3. 사용자가 "새 세션" 또는 "세션 완료" 선택 시 요약 생성

**Pseudocode**:
```python
# Extract completed tasks from TodoWrite
completed_tasks = [task for task in todos if task.status == "completed"]

# Store in session context
session_context = {
    "completed_tasks": completed_tasks,
    "command": "/alfred:X-command",
    "timestamp": datetime.now()
}

# If user selects "새 세션" or "세션 완료":
if user_choice in ["새 세션", "세션 완료"]:
    generate_session_summary(session_context)
```

---

#### 2.2. Session Summary Generator

**목표**: 세션 종료 시 작업 요약 생성

**Output Format**:
```markdown
## 🎊 세션 요약

### 완료된 작업
- ✅ 프로젝트 초기화 완료
- ✅ SPEC-AUTH-001 작성 완료
- ✅ 사용자 인증 구현 완료
- ✅ 테스트 작성 및 통과

### Git 통계
- 📝 생성된 커밋: 5개
- 📂 변경된 파일: 12개
- ➕ 추가된 라인: +450
- ➖ 삭제된 라인: -120

### 다음 권장 작업
1. `/alfred:3-sync`로 문서 동기화
2. PR 생성 및 리뷰 요청
3. 다음 SPEC 작성 시작
```

**구현 위치**: Alfred agent (agent-alfred.md)

---

### Step 3: CLAUDE.md 문서 업데이트

**파일**: `CLAUDE.md`

**변경 사항**:
1. "⚡ Alfred Command Completion Pattern" 섹션 추가 완료 (이미 존재)
2. 각 커맨드별 완료 패턴 예제 추가 완료
3. "Batched Design Principle" 섹션 추가 완료

**검증 사항**:
- ✅ 4개 커맨드 모두 패턴 설명 포함
- ✅ AskUserQuestion 예제 코드 포함
- ✅ Prohibited patterns 명시

---

## File Modification List

| File                                      | Type   | Changes                                       |
| ----------------------------------------- | ------ | --------------------------------------------- |
| `.claude/commands/alfred-0-project.md`    | UPDATE | Add AskUserQuestion completion pattern        |
| `.claude/commands/alfred-1-plan.md`       | UPDATE | Add AskUserQuestion completion pattern        |
| `.claude/commands/alfred-2-run.md`        | UPDATE | Add AskUserQuestion completion pattern        |
| `.claude/commands/alfred-3-sync.md`       | UPDATE | Add AskUserQuestion completion pattern        |
| `.claude/agents/agent-alfred.md`          | UPDATE | Add session cleanup logic                     |
| `CLAUDE.md`                               | VERIFY | Verify completion pattern documentation       |
| `.moai/specs/SPEC-SESSION-CLEANUP-001/`   | CREATE | Create SPEC documents (spec, plan, acceptance)|

**Total**: 6개 파일 수정 + 1개 디렉토리 생성

---

## Sub-agent Responsibilities

### tdd-implementer
- 4개 커맨드 파일 수정 담당
- AskUserQuestion 템플릿 삽입
- Syntax validation

### doc-syncer
- CLAUDE.md 검증 담당
- 문서 일관성 확인
- Cross-reference 업데이트

### tag-agent
- @TAG 체인 검증
- Traceability 확인
- SPEC traceability 검증

### git-manager
- 변경사항 커밋
- Branch: `feature/SPEC-SESSION-CLEANUP-001`
- Commit message 생성

---

## Quality Assurance Checklist

### Code Quality
- [ ] 모든 AskUserQuestion 호출이 올바른 Python 문법을 따르는가?
- [ ] 각 커맨드별로 3개 옵션이 명확히 정의되었는가?
- [ ] Prose 제안이 완전히 제거되었는가?

### Documentation Quality
- [ ] CLAUDE.md에 모든 커맨드 패턴이 문서화되었는가?
- [ ] 예제 코드가 실제 구현과 일치하는가?
- [ ] @TAG 체인이 올바르게 연결되었는가?

### User Experience
- [ ] 옵션 설명이 사용자의 `conversation_language`로 작성되었는가?
- [ ] 각 옵션의 label과 description이 명확한가?
- [ ] Batched design이 1-4개 질문 제한을 준수하는가?

### Testing
- [ ] 각 커맨드 완료 후 AskUserQuestion이 호출되는가?
- [ ] 사용자 선택에 따라 올바른 동작이 실행되는가?
- [ ] 세션 요약이 정확히 생성되는가?

---

## Risks & Mitigation

| Risk                                   | Impact | Probability | Mitigation                                         |
| -------------------------------------- | ------ | ----------- | -------------------------------------------------- |
| AskUserQuestion tool 호출 실패         | High   | Low         | Fallback: 임시 prose 메시지 (로그 기록)           |
| 기존 워크플로우 호환성 문제            | Medium | Medium      | Backward compatibility test 추가                   |
| TodoWrite 상태 불일치                  | Medium | Low         | Pre-completion validation hook 추가                |
| 문서 업데이트 누락                     | Low    | Medium      | Automated validation script 실행                   |

---

## Success Criteria

1. **모든 커맨드 완료 시 AskUserQuestion 호출 확인**
   - ✅ `/alfred:0-project` → AskUserQuestion with 3 options
   - ✅ `/alfred:1-plan` → AskUserQuestion with 3 options
   - ✅ `/alfred:2-run` → AskUserQuestion with 3 options
   - ✅ `/alfred:3-sync` → AskUserQuestion with 3 options

2. **Prose 제안 완전 제거**
   - ❌ "You can now run..." 패턴 검색 결과 0건

3. **세션 요약 생성 검증**
   - ✅ 완료된 작업 목록 포함
   - ✅ Git 통계 포함
   - ✅ Markdown 형식 출력

4. **문서 일관성 확인**
   - ✅ CLAUDE.md에 모든 패턴 문서화
   - ✅ @TAG 체인 검증 완료

---

## Timeline & Dependencies

### Dependencies
- **SPEC-ALF-WORKFLOW-001**: Alfred 4-Step Workflow (완료)
- `moai-alfred-interactive-questions` skill (이미 구현됨)

### Timeline
- **Phase 1**: 4개 커맨드 파일 수정 (Step 1.1-1.4)
- **Phase 2**: 세션 정리 로직 구현 (Step 2.1-2.2)
- **Phase 3**: 문서 검증 및 업데이트 (Step 3)

**예상 작업 단위**: 4개 파일 수정 + 2개 로직 추가 + 1개 문서 검증 = 7 tasks

---

## Next Steps

1. **Implementation**: `/alfred:2-run SPEC-SESSION-CLEANUP-001` 실행
2. **Testing**: acceptance.md의 Given-When-Then 시나리오 실행
3. **Documentation**: CHANGELOG.md 업데이트
4. **Sync**: `/alfred:3-sync` 실행

---
