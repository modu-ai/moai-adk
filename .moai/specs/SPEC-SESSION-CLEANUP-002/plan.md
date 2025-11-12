# Implementation Plan: @SPEC:SESSION-CLEANUP-002

<!-- @PLAN:SESSION-CLEANUP-002 -->

---

## Overview

이 구현 계획은 SPEC-SESSION-CLEANUP-002의 요구사항을 4개 Phase로 나누어 실제 커맨드 파일과 에이전트 파일에 Session Cleanup 패턴을 적용하는 방법을 정의합니다.

**목표**:
- ✅ 4개 Alfred 커맨드 파일 수정 (AskUserQuestion 패턴 추가)
- ✅ 1개 에이전트 파일 업데이트 (세션 정리 로직)
- ✅ 8개 테스트 시나리오 실행
- ✅ Git 커밋 및 문서화 완료

**영향받는 컴포넌트**:
- 4개 커맨드 파일 (`.claude/commands/alfred-*.md`)
- 1개 에이전트 파일 (`.claude/agents/agent-alfred.md`)
- 1개 문서 파일 (`CLAUDE.md` - 검증만)

---

## Implementation Phases

### Phase 2A: 커맨드 파일 수정 (4 Files)

**목표**: 4개 Alfred 커맨드 파일에 일관된 AskUserQuestion 완료 패턴 추가

**Duration**: Primary goal (첫 번째 우선순위)

**Sub-agent**: tdd-implementer (파일 수정 담당)

---

#### Step 2A-1: `/alfred:0-project` 파일 수정

**파일**: `src/moai_adk/templates/.claude/commands/alfred-0-project.md`

**작업 내용**:
1. 파일 마지막 섹션(`## Success Criteria` 다음) 위치 확인
2. 새 섹션 추가: `## Final Step: Next Action Selection`
3. AskUserQuestion 템플릿 삽입 (3 options)
4. Prohibited warning 추가

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

**검증**:
- [ ] 섹션이 올바른 위치에 추가되었는가?
- [ ] Python 문법이 올바른가?
- [ ] 3개 옵션이 모두 정의되었는가?

**@TAG**: `@CODE:CMD-0-PROJECT-IMPL`

---

#### Step 2A-2: `/alfred:1-plan` 파일 수정

**파일**: `src/moai_adk/templates/.claude/commands/alfred-1-plan.md`

**작업 내용**:
1. 파일 마지막 섹션 위치 확인
2. 새 섹션 추가: `## Final Step: Next Action Selection`
3. AskUserQuestion 템플릿 삽입 (3 options)

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

**Prohibited**: Never suggest next steps in prose.
```

**@TAG**: `@CODE:CMD-1-PLAN-IMPL`

---

#### Step 2A-3: `/alfred:2-run` 파일 수정

**파일**: `src/moai_adk/templates/.claude/commands/alfred-2-run.md`

**작업 내용**:
1. 파일 마지막 섹션 위치 확인
2. 새 섹션 추가: `## Final Step: Next Action Selection`
3. AskUserQuestion 템플릿 삽입 (3 options)

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

**Prohibited**: Never suggest next steps in prose.
```

**@TAG**: `@CODE:CMD-2-RUN-IMPL`

---

#### Step 2A-4: `/alfred:3-sync` 파일 수정

**파일**: `src/moai_adk/templates/.claude/commands/alfred-3-sync.md`

**작업 내용**:
1. 파일 마지막 섹션 위치 확인
2. 새 섹션 추가: `## Final Step: Next Action Selection`
3. AskUserQuestion 템플릿 삽입 (3 options)

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

**Prohibited**: Never suggest next steps in prose.
```

**@TAG**: `@CODE:CMD-3-SYNC-IMPL`

---

**Phase 2A Summary**:
- **Files Modified**: 4
- **Lines Added**: ~80 (20 lines per file)
- **Dependencies**: None
- **Sub-agent**: tdd-implementer
- **Priority**: Primary goal

---

### Phase 2B: 에이전트 파일 업데이트 (1 File)

**목표**: agent-alfred.md에 TodoWrite 정리 로직 및 세션 요약 생성 로직 추가

**Duration**: Secondary goal (두 번째 우선순위)

**Sub-agent**: tdd-implementer (파일 수정 담당)

---

#### Step 2B-1: TodoWrite Cleanup Logic 추가

**파일**: `src/moai_adk/templates/.claude/agents/agent-alfred.md`

**작업 내용**:
1. 새 섹션 추가: `## Session Cleanup Protocol`
2. TodoWrite 정리 로직 설명 추가
3. Pseudocode 예제 제공

**템플릿**:
```markdown
## Session Cleanup Protocol

### TodoWrite Cleanup Logic

**Trigger**: Before invoking AskUserQuestion at command completion

**Process**:
1. Extract all `completed` tasks from TodoWrite
2. Store in session context
3. Generate summary if user selects "새 세션" or "세션 완료"

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

# Step 3: If user selects "새 세션" or "세션 완료":
if user_choice in ["🔄 새 세션 시작", "✅ 세션 완료"]:
    generate_session_summary(session_context)
```

**@TAG**: `@CODE:AGENT-ALFRED-CLEANUP`
```

**검증**:
- [ ] 섹션이 올바른 위치에 추가되었는가?
- [ ] Pseudocode가 명확한가?

---

#### Step 2B-2: Session Summary Generator 추가

**파일**: `src/moai_adk/templates/.claude/agents/agent-alfred.md`

**작업 내용**:
1. 같은 섹션 내에 `### Session Summary Generator` 추가
2. Output format 정의
3. 구현 가이드 제공

**템플릿**:
```markdown
### Session Summary Generator

**Trigger**: User selects "새 세션" or "세션 완료"

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

**Implementation Guide**:
1. Call `git log --oneline` to count commits
2. Call `git diff --stat` to get file changes
3. Extract completed tasks from TodoWrite
4. Generate Markdown output directly (no Bash wrapping)

**@TAG**: `@CODE:AGENT-ALFRED-SUMMARY`
```

**검증**:
- [ ] Output format이 Markdown 형식인가?
- [ ] Git 명령어가 올바른가?

---

**Phase 2B Summary**:
- **Files Modified**: 1
- **Lines Added**: ~60
- **Dependencies**: Phase 2A (커맨드 파일 수정 완료 후)
- **Sub-agent**: tdd-implementer
- **Priority**: Secondary goal

---

### Phase 2C: 테스트 시나리오 실행 (8 Tests)

**목표**: acceptance.md에 정의된 8개 테스트 시나리오 실행 및 검증

**Duration**: Tertiary goal (세 번째 우선순위)

**Sub-agent**: tdd-implementer (테스트 실행), tag-agent (TAG 검증)

---

#### Test Group 1: AskUserQuestion 호출 검증 (4 tests)

**TEST-001**: `/alfred:0-project` 완료 후 AskUserQuestion 호출 검증
- **Given**: 프로젝트 초기화 완료
- **When**: 커맨드 종료 직전
- **Then**: AskUserQuestion이 3개 옵션과 함께 호출됨

**TEST-002**: `/alfred:1-plan` 완료 후 AskUserQuestion 호출 검증
- **Given**: SPEC 작성 완료
- **When**: 커맨드 종료 직전
- **Then**: AskUserQuestion이 3개 옵션과 함께 호출됨

**TEST-003**: `/alfred:2-run` 완료 후 AskUserQuestion 호출 검증
- **Given**: TDD 구현 완료
- **When**: 커맨드 종료 직전
- **Then**: AskUserQuestion이 3개 옵션과 함께 호출됨

**TEST-004**: `/alfred:3-sync` 완료 후 AskUserQuestion 호출 검증
- **Given**: 문서 동기화 완료
- **When**: 커맨드 종료 직전
- **Then**: AskUserQuestion이 3개 옵션과 함께 호출됨

---

#### Test Group 2: 세션 정리 검증 (2 tests)

**TEST-005**: 세션 요약 생성 (Markdown 형식 출력)
- **Given**: 사용자가 "세션 완료" 선택
- **When**: 세션 정리 시작
- **Then**: Markdown 형식의 세션 요약 출력됨

**TEST-006**: TodoWrite 정리 (completed 작업 추출)
- **Given**: 커맨드 실행 중 3개 작업 completed
- **When**: AskUserQuestion 호출 직전
- **Then**: 3개 completed 작업이 session_context에 저장됨

---

#### Test Group 3: 품질 검증 (2 tests)

**TEST-007**: Prose 패턴 검색 (검출 0건)
- **Given**: 4개 커맨드 파일 수정 완료
- **When**: `rg "You can now run" .claude/commands/alfred-*.md` 실행
- **Then**: 검색 결과 0건

**TEST-008**: Batched 디자인 (호출 횟수 = 1)
- **Given**: 각 커맨드 완료 시
- **When**: AskUserQuestion 호출
- **Then**: 호출 횟수 = 1 (batched design)

---

**Phase 2C Summary**:
- **Tests Executed**: 8
- **Test Groups**: 3 (AskUserQuestion / 세션 정리 / 품질 검증)
- **Dependencies**: Phase 2A, 2B 완료
- **Sub-agent**: tdd-implementer, tag-agent
- **Priority**: Tertiary goal

---

### Phase 2D: Git 커밋 및 문서화

**목표**: 변경사항 커밋, CHANGELOG 업데이트, CLAUDE.md 검증

**Duration**: Final goal (마지막 우선순위)

**Sub-agent**: git-manager (Git 작업), doc-syncer (문서 검증)

---

#### Step 2D-1: Git Commit 생성

**작업 내용**:
1. 변경된 5개 파일 스테이징 (`git add`)
2. Commit message 생성 (Alfred co-authorship)
3. Branch 확인: `feature/SPEC-SESSION-CLEANUP-002`

**Commit Message Template**:
```
feat(alfred): Implement Session Cleanup pattern in Alfred commands

Implement SPEC-SESSION-CLEANUP-002:
- Add AskUserQuestion completion pattern to 4 Alfred commands
- Add TodoWrite cleanup logic to agent-alfred.md
- Add session summary generator
- Remove prose suggestion patterns

Changes:
- Modified: alfred-0-project.md (add Final Step section)
- Modified: alfred-1-plan.md (add Final Step section)
- Modified: alfred-2-run.md (add Final Step section)
- Modified: alfred-3-sync.md (add Final Step section)
- Modified: agent-alfred.md (add Session Cleanup Protocol)

Tests:
- 8 test scenarios executed (TEST-001 to TEST-008)
- All AskUserQuestion patterns verified
- Prose pattern search: 0 results

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)
```

**@TAG**: `@COMMIT:SESSION-CLEANUP-002`

---

#### Step 2D-2: CHANGELOG.md 업데이트

**파일**: `CHANGELOG.md`

**작업 내용**:
1. 새 섹션 추가: `## [Unreleased]`
2. 변경 사항 기록 (feature)

**템플릿**:
```markdown
## [Unreleased]

### Added
- Session Cleanup pattern in Alfred commands (from SPEC-SESSION-CLEANUP-002)
  - `/alfred:0-project` completion with AskUserQuestion
  - `/alfred:1-plan` completion with AskUserQuestion
  - `/alfred:2-run` completion with AskUserQuestion
  - `/alfred:3-sync` completion with AskUserQuestion
  - TodoWrite cleanup protocol
  - Session summary generator

### Changed
- Removed prose suggestion patterns from all commands
- Consistent UX across all Alfred commands

### Removed
- Prose "You can now run..." patterns
```

---

#### Step 2D-3: CLAUDE.md 검증

**파일**: `CLAUDE.md`

**작업 내용**:
1. "⚡ Alfred Command Completion Pattern" 섹션 검증
2. 4개 커맨드 패턴 설명 확인
3. 예제 코드 일관성 확인

**검증 항목**:
- [ ] 4개 커맨드 모두 완료 패턴 문서화되었는가?
- [ ] AskUserQuestion 예제가 실제 구현과 일치하는가?
- [ ] Batched Design Principle 섹션이 있는가?
- [ ] Prohibited patterns가 명시되었는가?

**Sub-agent**: doc-syncer

---

**Phase 2D Summary**:
- **Git Operations**: 1 commit
- **Documents Updated**: 2 (CHANGELOG.md, CLAUDE.md 검증)
- **Dependencies**: Phase 2A, 2B, 2C 완료
- **Sub-agent**: git-manager, doc-syncer
- **Priority**: Final goal

---

## File Modification Summary

| File                                      | Type   | Lines Changed | Phase  |
| ----------------------------------------- | ------ | ------------- | ------ |
| `alfred-0-project.md`                     | UPDATE | +20           | 2A     |
| `alfred-1-plan.md`                        | UPDATE | +20           | 2A     |
| `alfred-2-run.md`                         | UPDATE | +20           | 2A     |
| `alfred-3-sync.md`                        | UPDATE | +20           | 2A     |
| `agent-alfred.md`                         | UPDATE | +60           | 2B     |
| `CHANGELOG.md`                            | UPDATE | +15           | 2D     |
| `CLAUDE.md`                               | VERIFY | 0             | 2D     |

**Total**: 5개 파일 수정 + 1개 파일 검증 + 1개 파일 업데이트 = 7 tasks

---

## Sub-agent Responsibilities

### tdd-implementer
- **Phase 2A**: 4개 커맨드 파일 수정
- **Phase 2B**: agent-alfred.md 업데이트
- **Phase 2C**: 8개 테스트 시나리오 실행

### tag-agent
- **Phase 2C**: @TAG 체인 검증
- Traceability 확인 (SPEC → CODE → TEST)

### doc-syncer
- **Phase 2D**: CLAUDE.md 검증
- CHANGELOG.md 일관성 확인

### git-manager
- **Phase 2D**: Git commit 생성
- Branch 관리
- Commit message 작성

---

## Quality Assurance Checklist

### Code Quality
- [ ] 모든 AskUserQuestion 호출이 올바른 Python 문법을 따르는가?
- [ ] 각 커맨드별로 3개 옵션이 명확히 정의되었는가?
- [ ] Prose 제안이 완전히 제거되었는가? (검색 결과 0건)

### Documentation Quality
- [ ] CLAUDE.md에 모든 커맨드 패턴이 문서화되었는가?
- [ ] 예제 코드가 실제 구현과 일치하는가?
- [ ] @TAG 체인이 올바르게 연결되었는가?

### User Experience
- [ ] 옵션 설명이 사용자의 `conversation_language` (Korean)로 작성되었는가?
- [ ] 각 옵션의 label과 description이 명확한가?
- [ ] Batched design이 1회 호출로 제한되는가?

### Testing
- [ ] 각 커맨드 완료 후 AskUserQuestion이 호출되는가?
- [ ] 사용자 선택에 따라 올바른 동작이 실행되는가?
- [ ] 세션 요약이 정확히 생성되는가?

### Git Quality
- [ ] Commit message가 Alfred co-authorship을 포함하는가?
- [ ] Branch가 올바른가? (`feature/SPEC-SESSION-CLEANUP-002`)
- [ ] CHANGELOG.md가 업데이트되었는가?

---

## Risks & Mitigation

| Risk                                   | Impact | Probability | Mitigation                                         |
| -------------------------------------- | ------ | ----------- | -------------------------------------------------- |
| Markdown 파싱 에러                     | High   | Low         | Syntax validation 추가, TDD 테스트 작성             |
| AskUserQuestion tool 호출 실패         | High   | Low         | Fallback: 임시 prose 메시지 (에러 로그)            |
| TodoWrite 상태 불일치                  | Medium | Medium      | Pre-completion validation hook 추가                |
| 기존 워크플로우 호환성 문제            | Medium | Low         | Backward compatibility test 추가                   |
| Prose 패턴 재도입 (개발자 실수)        | Low    | Medium      | Automated grep search in CI/CD                     |
| Git conflict (다른 branch와 충돌)      | Low    | Low         | Rebase before commit, conflict resolution          |

---

## Success Criteria

1. **모든 커맨드 파일 수정 완료**
   - ✅ 4개 파일에 Final Step 섹션 추가
   - ✅ AskUserQuestion 템플릿 삽입
   - ✅ Prohibited warning 추가

2. **에이전트 파일 업데이트 완료**
   - ✅ TodoWrite 정리 로직 추가
   - ✅ 세션 요약 생성 로직 추가

3. **테스트 시나리오 실행 완료**
   - ✅ 8개 테스트 모두 통과
   - ✅ Prose 패턴 검색 결과 0건

4. **Git 커밋 및 문서화 완료**
   - ✅ 1개 commit 생성
   - ✅ CHANGELOG.md 업데이트
   - ✅ CLAUDE.md 검증 완료

---

## Dependencies

### External Dependencies
- **SPEC-SESSION-CLEANUP-001**: Phase 1 완료 (documentation)
- `moai-alfred-ask-user-questions` skill: 이미 구현됨

### Internal Dependencies
- Phase 2A → Phase 2B (커맨드 파일 수정 후 에이전트 파일 업데이트)
- Phase 2B → Phase 2C (세션 정리 로직 추가 후 테스트)
- Phase 2C → Phase 2D (테스트 통과 후 Git 커밋)

---

## Timeline & Milestones

### Milestone 1: Command Files Updated (Phase 2A)
- **Goal**: 4개 커맨드 파일 수정 완료
- **Priority**: Primary goal
- **Dependencies**: None

### Milestone 2: Agent File Updated (Phase 2B)
- **Goal**: agent-alfred.md 세션 정리 로직 추가
- **Priority**: Secondary goal
- **Dependencies**: Phase 2A 완료

### Milestone 3: Tests Executed (Phase 2C)
- **Goal**: 8개 테스트 시나리오 실행 및 검증
- **Priority**: Tertiary goal
- **Dependencies**: Phase 2A, 2B 완료

### Milestone 4: Git & Docs (Phase 2D)
- **Goal**: Git 커밋 및 문서화 완료
- **Priority**: Final goal
- **Dependencies**: Phase 2A, 2B, 2C 완료

---

## Technical Approach

### Architectural Pattern
- **Template Modification**: Markdown 템플릿에 새 섹션 추가 (non-breaking change)
- **Batched Design**: 1회 AskUserQuestion 호출로 3개 옵션 제시
- **Session Management**: TodoWrite 기반 작업 추적 및 정리

### Technology Stack
- **Language**: Markdown, Python (code blocks)
- **Tools**: AskUserQuestion tool, TodoWrite system
- **Testing**: Manual verification (8 test scenarios)

### Integration Points
- `moai-alfred-ask-user-questions` skill (TUI)
- TodoWrite system (session state)
- Git workflow (commits, branch strategy)

---

## Next Steps

1. **Implementation**: `/alfred:2-run SPEC-SESSION-CLEANUP-002` 실행
2. **Testing**: acceptance.md의 8개 테스트 시나리오 실행
3. **Documentation**: CHANGELOG.md 업데이트
4. **Sync**: `/alfred:3-sync` 실행 (문서 동기화)

---

**End of Plan: SPEC-SESSION-CLEANUP-002**
