# STEP 0-UPDATE 리팩토링 완료 보고서

## 작업 개요

**날짜**: 2025-01-04
**작업자**: cc-manager agent
**대상 파일**:
- `.claude/commands/alfred/0-project.md` (로컬)
- `src/moai_adk/templates/.claude/commands/alfred/0-project.md` (패키지 템플릿)

## 변경 사항

### Before: 선언적 의사 코드 (Phase 기반)

기존 STEP 0-UPDATE 섹션은 다음과 같은 문제가 있었습니다:

1. **Phase 기반 구조**: Phase 1, Phase 2로 나누어져 있어 실행 흐름이 불명확
2. **선언적 언어**: "WHAT" (무엇을 하는지)만 설명하고 "HOW" (어떻게 실행하는지) 누락
3. **실행 불가능**: Claude Code가 직접 실행할 수 없는 추상적 설명
4. **오류 처리 누락**: 실패 시 복구 절차 미흡
5. **조건부 흐름 불명확**: Preview 선택 시 분기 로직 모호

### After: 순수 명령형 단계별 지침 (STEP 기반)

새로운 구조:

```
STEP 0-UPDATE.1: Verify prerequisites and check backup
STEP 0-UPDATE.2: Load and compare templates
STEP 0-UPDATE.3: Display comparison report and ask for approval
STEP 0-UPDATE.3.1: Show detailed preview (conditional)
STEP 0-UPDATE.4: Create safety backup before merge
STEP 0-UPDATE.5: Execute smart merge
STEP 0-UPDATE.5.1: Update config.json metadata
STEP 0-UPDATE.5.2: Display completion report
STEP 0-UPDATE.6: Error recovery (merge failure)
STEP 0-UPDATE.7: Graceful exit (user skipped)
```

## 개선 사항

### 1. 명령형 언어로 전환

**Before**:
```markdown
### Phase 1: Backup analysis and comparison

1. **Make sure you have the latest backup**:
- Compare `.claude/` directory from backup with current template
```

**After**:
```markdown
### STEP 0-UPDATE.1: Verify prerequisites and check backup

**Your task**: Verify that prerequisites exist before starting template optimization.

**Steps**:
1. **Check if backup directory exists**:
   - Directory to check: `.moai-backups/`
   - IF directory does NOT exist → Show error and exit:
     ```
     ❌ Error: No backup found at .moai-backups/
     ...
     ```
```

### 2. 명확한 실행 단계

모든 작업에 대해:
- **Your task**: 이 단계에서 해야 할 일
- **Steps**: 1, 2, 3... 순차적 실행 지침
- **IF/THEN**: 조건부 분기 명시
- **Read/Write/Print**: 파일 작업 명시
- **Go to STEP X**: 다음 단계로의 전환 명시

### 3. 오류 처리 강화

**새로 추가된 오류 시나리오**:

1. **STEP 0-UPDATE.1**: 백업 디렉토리 없음
2. **STEP 0-UPDATE.1**: config.json 없음
3. **STEP 0-UPDATE.4**: 백업 생성 실패
4. **STEP 0-UPDATE.5**: 파일 쓰기 실패 → STEP 0-UPDATE.6 (복구)
5. **STEP 0-UPDATE.6**: 병합 실패 시 자동 복구 절차

### 4. 조건부 흐름 명시

**Preview 경로**:
```
STEP 0-UPDATE.3 (사용자에게 질문)
  ├─ "Proceed" → STEP 0-UPDATE.4 (백업 생성)
  ├─ "Preview" → STEP 0-UPDATE.3.1 (상세 미리보기)
  │              └─ "Proceed" → STEP 0-UPDATE.4
  │              └─ "Skip" → STEP 0-UPDATE.7 (종료)
  └─ "Skip" → STEP 0-UPDATE.7 (종료)
```

### 5. 사용자 상호작용 개선

**Before**: 단순 텍스트 설명
**After**: AskUserQuestion 도구 활용

```python
AskUserQuestion(
    questions=[
        {
            "question": "Template optimization analysis complete. How would you like to proceed?",
            "header": "📦 Template Optimization",
            "multiSelect": false,
            "options": [
                {"label": "✅ Proceed", ...},
                {"label": "👀 Preview", ...},
                {"label": "⏸️ Skip", ...}
            ]
        }
    ]
)
```

### 6. 진행 상황 출력 개선

모든 단계에서 사용자에게 명확한 피드백:

```
✅ Prerequisites verified
💾 Safety backup created
✓ CLAUDE.md merged
✓ .claude/settings.json merged
⚙️ config.json updated
✅ Template optimization completed!
```

## 통계

| 항목 | Before | After | 변화 |
|------|--------|-------|------|
| 라인 수 | 162 | 597 | +435 (+268%) |
| 섹션 수 | 4 (Phase 1-2 + 2개 부속) | 10 (STEP 0-UPDATE.1 ~ 0-UPDATE.7) | +6 |
| 오류 처리 섹션 | 1 (간략) | 2 (상세 + 복구) | +1 |
| 조건부 분기 | 1 (Preview) | 3 (Preview, Error, Skip) | +2 |
| 사용자 상호작용 | 1 (간략) | 4 (질문 + 미리보기 + 오류 + 종료) | +3 |

## 검증 체크리스트

- [x] Phase 기반 언어 제거 ("Phase 1" → "STEP 0-UPDATE.1")
- [x] 모든 섹션 명령형 언어 사용 ("Your task:", "Steps:", "Print:")
- [x] 모든 파일 작업 명시 (Read:, Write:, Update:)
- [x] 모든 사용자 상호작용 명시 (AskUserQuestion 도구)
- [x] 오류 조건 사전 식별
- [x] 병합 전략 단계별 상세화
- [x] 백업/안전 조치 명시
- [x] 완료 보고서 구조화
- [x] 의사 코드 제거 (모든 추상적 설명 제거)
- [x] 로컬 파일 업데이트 완료
- [x] 패키지 템플릿 동기화 완료

## 파일 위치

**로컬**:
- `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/0-project.md`
  - 라인 2843-3448 (597 라인)

**패키지 템플릿**:
- `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/commands/alfred/0-project.md`
  - 라인 2843-3447 (597 라인)

## 다음 단계

1. **검증 테스트**: 실제 `/alfred:0-project update` 실행하여 새 지침 검증
2. **STEP 0.1.1 리팩토링** (선택 사항): First-time Setup 섹션도 동일 패턴 적용
3. **문서 업데이트**: 변경 사항 CHANGELOG.md에 반영
4. **커밋 생성**: 리팩토링 완료 커밋

## 참고 자료

**리팩토링 패턴 문서**: 이 리팩토링은 다음 원칙을 따랐습니다:
- **명령형 언어**: "Your task", "Steps", "IF/THEN"
- **실행 가능성**: Claude Code가 직접 실행 가능한 구체적 지침
- **오류 복원력**: 모든 실패 시나리오에 복구 절차
- **투명성**: 사용자에게 모든 단계 명시적 피드백
- **추적 가능성**: 단계 번호로 진행 상황 추적
