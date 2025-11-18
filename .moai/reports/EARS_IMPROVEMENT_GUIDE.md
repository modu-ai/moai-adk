# EARS 형식 표준화 개선 가이드

**목적**: Priority별 개선 항목의 구체적 수행 방법 제시
**대상**: 개발 팀 및 유지보수 담당자

---

## Priority 1: AskUserQuestion 필드 정규화 (1주)

### 현황 분석

**문제점**:
```markdown
현재 코드:
{
  "question": "Planning is complete. Would you like to proceed?",
  "header": "SPEC Generation",  ❌ 이모지 없음 (좋음)
  "options": [
    {
      "label": "✅ Proceed with SPEC Creation",  ❌ 이모지 포함
      "description": "📋 Create SPEC files..."    ❌ 이모지 포함
    }
  ]
}
```

**영향**:
- 터미널 렌더링 불일치
- 국제화 문제 (일부 터미널에서 이모지 미표시)
- 접근성 저하 (스크린 리더 혼란)

### 개선 작업 흐름

**Step 1**: 모든 파일에서 AskUserQuestion 패턴 검색
```bash
# 검색 대상
grep -r "AskUserQuestion" .claude/commands/ .claude/agents/ | \
  grep -E "label|header|description" | \
  grep -E "[✅❌📋🔍⚠️📚]"
```

**Step 2**: 필드별 이모지 제거 및 텍스트 강화

| 필드 | 현재 | 개선 | 예시 |
|------|------|------|------|
| header | 이모지 선택적 | 순수 텍스트 | "Plan Approval" |
| label | ❌ 제거 | 명확한 액션 | "Proceed with SPEC Creation" |
| description | ❌ 제거 | 구체적 설명 | "Generate SPEC files in .moai/specs/" |

**Step 3**: 개선된 예시 코드

```python
# Before (개선 전)
AskUserQuestion({
    "questions": [{
        "question": "Planning is complete. Proceed?",
        "header": "📋 Plan Approval",
        "multiSelect": False,
        "options": [
            {
                "label": "✅ Proceed with SPEC",
                "description": "📝 Create SPEC files in .moai/"
            },
            {
                "label": "❌ Cancel",
                "description": "🔄 Return to planning"
            }
        ]
    }]
})

# After (개선 후)
AskUserQuestion({
    "questions": [{
        "question": "Planning is complete. How would you like to proceed?",
        "header": "Plan Approval",
        "multiSelect": False,
        "options": [
            {
                "label": "Proceed with SPEC Creation",
                "description": "Generate SPEC files in .moai/specs/SPEC-{ID}/"
            },
            {
                "label": "Request Modifications",
                "description": "Modify plan content before creation"
            },
            {
                "label": "Save as Draft",
                "description": "Save plan for later and exit"
            },
            {
                "label": "Cancel",
                "description": "Discard plan and return to previous step"
            }
        ]
    }]
})
```

### 수행 체크리스트

- [ ] 6개 명령어 파일 검사 및 수정
- [ ] 28개 에이전트 파일 검사 및 수정
- [ ] grep 재검색으로 모든 이모지 제거 확인
- [ ] 각 field description을 명확한 텍스트로 작성
- [ ] 테스트: Claude Code에서 실행하여 터미널 렌더링 확인

**예상 시간**: 2-3시간
**영향**: 사용자 경험 +15%

---

## Priority 2: GIVEN/AND 명시화 (1-2주)

### 개선 대상 명령어

#### `/moai:0-project`

**현재 상태**:
```markdown
명령 목적: "Initialize project metadata and documentation"
입력: "setting [tab_ID] | update"
```

**개선 제안**:
```markdown
## 🔍 전체 EARS 매트릭스

### INITIALIZATION 모드
GIVEN 프로젝트 초기화 필요 / 언어 설정 없음
WHEN /moai:0-project (인수 없음) 실행
THEN .moai/config.json 생성
  AND .claude/ 구조 자동 생성
  AND 프로젝트 문서 생성
  AND 언어 선택 UI 표시 (첫 실행 시)

### AUTO-DETECT 모드
GIVEN .moai/config.json 이미 존재
WHEN /moai:0-project (인수 없음) 실행
THEN 현재 설정 표시
  AND 수정 옵션 제시 (Settings/Language/Review/Cancel)
  AND 사용자 선택 처리

### SETTINGS 모드
GIVEN 설정 변경 필요
WHEN /moai:0-project setting [tab_ID] 실행
THEN Tab별 배치 질문 표시
  AND 사용자 응답 수집
  AND 원자적 업데이트 수행
  AND 변경 사항 보고

### UPDATE 모드
GIVEN 패키지 업데이트 후 템플릿 재적용
WHEN /moai:0-project update 실행
THEN 템플릿 비교 분석
  AND 스마트 머지 수행
  AND 기존 설정 보존
  AND 변경 발표 (다국어)
```

#### `/moai:2-run`

**추가해야 할 AND 조건**:
```markdown
원래:
GIVEN SPEC ID 지정
WHEN 구현 시작
THEN TDD 구현 (RED→GREEN→REFACTOR)

개선:
GIVEN SPEC ID 지정
WHEN /moai:2-run SPEC-XXX 실행
THEN Phase 1: 분석 및 계획 수립
  AND Phase 2: TDD 사이클 실행
  AND Phase 3: Git 커밋 생성
  AND Phase 4: 다음 단계 안내
  AND (오류 시) 오류 복구 제안
  AND (품질 게이트 실패 시) PASS/WARNING/CRITICAL 보고
```

### 수행 방법

1. **Step 1**: 각 명령어의 4가지 모드/시나리오 식별
2. **Step 2**: 각 시나리오별 EARS 매트릭스 작성
3. **Step 3**: PHASE별 AND 조건 명시
4. **Step 4**: 오류 경로 AND 조건 추가

**체크리스트**:
- [ ] /moai:0-project 개선 (GIVEN/AND 명시화)
- [ ] /moai:2-run 개선 (오류 경로 AND 추가)
- [ ] /moai:9-feedback 개선 (GIVEN 명확화)
- [ ] 문서 형식 검토 및 최종화

**예상 시간**: 1-2시간
**영향**: 명확성 +20%

---

## Priority 2-B: 컨텍스트 변수 정규화 (1-2시간)

### 패턴 표준화

**규칙**: 모든 컨텍스트 변수는 `$VARIABLE_NAME` 형식 사용

**현재 불일치**:
```markdown
/moai:1-plan에서:
- $USER_REQUEST ✅
- $CONVERSATION_LANGUAGE ✅
- {{USER_REQUEST}} ❌ (이중 중괄호 사용)
- {{CONVERSATION_LANGUAGE}} ❌

/moai:0-project에서:
- $MODE ✅
- $SPEC_ID ✅
- $TIMESTAMP ✅

/moai:3-sync에서:
- $ARGUMENTS ✅
- $PROJECT_VALIDATION_RESULTS ✅
- $SYNC_PLAN ✅
```

**개선 작업**:

```bash
# Step 1: 이중 중괄호 검색
grep -r "{{" .claude/commands/

# Step 2: 모든 {{ }} 를 $ 형식으로 치환
# Before: {{USER_REQUEST}} → After: $USER_REQUEST

# Step 3: 변수 정의 위치 명시화
```

**표준 템플릿**:
```markdown
## 변수 정의

다음 변수는 명령어 실행 중 자동으로 설정됩니다:

| 변수 | 출처 | 용도 |
|------|------|------|
| $SPEC_ID | 사용자 입력 | SPEC 파일 위치 결정 |
| $CONVERSATION_LANGUAGE | .moai/config.json | 문서 언어 설정 |
| $MODE | 명령어 분석 | 실행 모드 결정 |
| $TIMESTAMP | 시스템 시간 | 백업/리포트 파일명 |
```

**체크리스트**:
- [ ] 모든 {{ }} 형식 검색 및 수정
- [ ] 변수 정의 섹션 추가
- [ ] 변수 사용처 명시화

**예상 시간**: 30분-1시간
**영향**: 코드 추적성 +10%

---

## Priority 2-C: 에러 처리 표준화 (3-4시간)

### 표준 에러 처리 패턴

**목표**: 모든 Task() 호출에 일관된 에러 처리 추가

**현재 상태**: 부분적 에러 처리만 명시

**표준 템플릿**:

```markdown
### Step X.Y: [작업명]

Use Task tool:
- `subagent_type`: "agent-name"
- `prompt`: """..."""

**Error Handling**:
- IF Response contains error:
  - Check error type: [네트워크 / 권한 / 데이터 / 기타]
  - Log: error_type, error_message, affected_file
  - Recovery option 1: [재시도 방법]
  - Recovery option 2: [대체 방법]
  - IF critical: abort with message
  - IF non-critical: continue with warning

- IF Response is empty:
  - Check agent status
  - Provide diagnostic information
  - Ask user: retry or skip

- IF Response is incomplete:
  - Log incomplete fields
  - Provide partial results
  - Continue with warning
```

### 적용 대상

| 명령어 | Task 호출 수 | 에러 처리율 |
|--------|-----------|-----------|
| /moai:0-project | 3 | 67% |
| /moai:1-plan | 4 | 75% |
| /moai:2-run | 5 | 60% |
| /moai:3-sync | 6 | 50% |
| **합계** | **18** | **63%** |

**개선 대상**: 7개 호출 (37%)

### 수행 체크리스트

- [ ] 에러 처리 표준 템플릿 확정
- [ ] /moai:0-project 에러 처리 강화 (목표 100%)
- [ ] /moai:1-plan 에러 처리 강화 (목표 100%)
- [ ] /moai:2-run 에러 처리 강화 (목표 100%)
- [ ] /moai:3-sync 에러 처리 강화 (목표 100%)
- [ ] 테스트: 각 에러 시나리오 검증

**예상 시간**: 3-4시간
**영향**: 안정성 +25%

---

## Priority 3: Skill 명세 강화 (2-3주, 40-50시간)

### 현황

**현재**: 각 SKILL.md는 기본 정보만 포함
- 설명
- 사용 예시
- 참고 문서

**개선**: EARS 형식 명세 추가

### 표준 Skill 명세 템플릿

```markdown
# Skill 명: moai-foundation-ears

## EARS 형식 명세

### WHEN (호출 시점)
- SPEC 문서 작성 필요 시
- 요구사항을 체계적으로 표현해야 할 때
- WHEN/THEN 구조 검증 필요 시

### GIVEN (선행 조건)
- 초기 요구사항 문장 제공
- 컨텍스트 정보 (도메인, 우선순위)
- 참고 문서 (기존 SPEC, 기능 명세)

### THEN (기대 결과)
- 구조화된 EARS 형식 요구사항
- WHEN/THEN 구조로 변환된 명세
- 완전성 검증 결과

### AND (추가 결과)
- 대체 표현식 제시 (필요 시)
- 실제 SPEC에 적용 가능한 예시
- 관련 Skill 추천

## INPUT 인터페이스

### 필수 입력
```json
{
  "requirement_text": "사용자 입력 요구사항",
  "domain": "backend | frontend | database | devops",
  "context": "추가 컨텍스트"
}
```

### 선택 입력
```json
{
  "format": "simple | extended | complex",
  "language": "ko | en | ja | es | fr"
}
```

## OUTPUT 형식

### 기본 구조
```markdown
## EARS 분석 결과

### Simple Format (단순 조건)
WHEN [조건]
THEN [결과]

### Extended Format (복합 조건)
GIVEN [초기 상태]
WHEN [트리거]
THEN [결과]
AND [추가 결과]

### 완성도: [백분율]%
- WHEN 명확성: ✅/⚠️/❌
- THEN 명확성: ✅/⚠️/❌
- 테스트 가능성: ✅/⚠️/❌
```

## TRUST 준수

- **Testable**: 명확한 THEN으로 검증 가능
- **Readable**: EARS 문법으로 자동 명료화
- **Unified**: 모든 요구사항 일관된 형식
- **Secured**: 보안 요구사항 자동 강조
- **Trackable**: TAG를 통한 추적성 제공

## 사용 예시

### 예시 1: 단순 요구사항
```
INPUT: "사용자가 로그인할 수 있어야 한다"

OUTPUT:
WHEN 로그인 페이지 방문
THEN 사용자 인증 수행
AND 유효한 토큰 발급

완성도: 85%
```

## 관련 Skills
- moai-foundation-specs - SPEC 전체 구조
- moai-core-spec-authoring - SPEC 작성 자동화
- moai-core-ears-authoring - EARS 확장 문법
```

### 적용 범위

| Skill 그룹 | 수량 | 우선순위 |
|-----------|------|---------|
| 기초 (foundation-*) | 5 | P1 |
| 도메인 (*-domain-*) | 20 | P2 |
| 언어 (*-lang-*) | 15 | P2 |
| 필수 (*-essentials-*) | 8 | P1 |
| MCP 통합 | 3 | P3 |
| 유틸리티 | 75 | P3 |
| **합계** | **126** | |

### 수행 계획

**Phase 1** (1주): 기초 Skills (5개)
- moai-foundation-ears
- moai-foundation-specs
- moai-foundation-trust
- moai-foundation-git
- moai-foundation-langs

**Phase 2** (2주): 필수 Skills (8개)
- moai-essentials-debug
- moai-essentials-refactor
- moai-essentials-perf
- moai-essentials-review
- (+ 4개)

**Phase 3** (3주): 도메인 & 언어 Skills (35개)

**체크리스트**:
- [ ] SKILL 템플릿 확정
- [ ] 기초 Skills (5) 작성 완료
- [ ] 필수 Skills (8) 작성 완료
- [ ] 도메인 Skills (20) 작성 완료
- [ ] 언어 Skills (15) 작성 완료
- [ ] 정기 검증 프로세스 수립

**예상 시간**: 40-50시간 (여러 기여자 병렬 작업 권장)
**영향**: 명세 완성도 +45%

---

## Priority 3-B: 다국어 예제 확대 (2-3주, 20-30시간)

### 현황

**현재**: 영어 공식 문서 + 한국어 사용 예시 (2언어)
**목표**: 5개 언어 완전 예제 (en, ko, ja, es, fr)

### 예제 기준

| 항목 | 현황 | 개선 |
|------|------|------|
| 명령어 설명 | English | 5언어 |
| 사용 예시 | 영어 + 한국어 | 5언어 |
| 에러 메시지 | English | 5언어 |
| 팁 & 권장사항 | English | 5언어 |

### 수행 대상

**Phase 1** (1주): 명령어 (6개)
- /moai:0-project
- /moai:1-plan
- /moai:2-run
- /moai:3-sync
- /moai:9-feedback
- /moai:99-release

**Phase 2** (1주): 핵심 에이전트 (10개)
- spec-builder
- tdd-implementer
- quality-gate
- git-manager
- project-manager
- (+ 5개)

**Phase 3** (1주): 추가 자료
- 에러 메시지 표준화
- 팁 & 권장사항

### 구현 전략

```bash
# Step 1: 번역 자동화 도구 활용
# Claude API를 이용한 자동 번역 + 리뷰

# Step 2: 각 언어별 검증
# 네이티브 스피커에 의한 검증

# Step 3: 렌더링 테스트
# 각 언어의 특수 문자 표시 확인 (일본어, 아랍어 등)
```

### 체크리스트

- [ ] 번역 대상 텍스트 목록 작성 (3-5개 per 문서)
- [ ] 자동 번역 도구 설정
- [ ] 6개 명령어 × 5언어 = 30개 번역
- [ ] 핵심 에이전트 10개 × 5언어 = 50개 번역
- [ ] 네이티브 검증
- [ ] 렌더링 테스트

**예상 시간**: 20-30시간 (자동화 포함)
**영향**: 접근성 +25%

---

## 🎯 실행 로드맵

### Week 1: Priority 1 (30시간)
```
Mon-Tue: AskUserQuestion 필드 정규화 (8시간)
Wed:     GIVEN/AND 명시화 (4시간)
Thu:     컨텍스트 변수 정규화 (2시간)
Fri:     테스트 & 검증 (4시간)
        Skill 명세 Phase 1 시작 (12시간)
```

### Week 2-3: Priority 2 (20시간)
```
Mon-Tue: Skill 명세 Phase 2 (12시간)
Wed-Fri: 에러 처리 표준화 (8시간)
```

### Week 4+: Priority 3 (70시간)
```
Week 4-5:   Skill 명세 완성 (40시간)
Week 6-7:   다국어 예제 작성 (20시간)
Week 8:     최종 검증 & 정기 리뷰 (10시간)
```

---

## 📊 성공 지표

### 정량 지표

| 지표 | 현재 | 목표 | 측정 방법 |
|------|------|------|---------|
| EARS 준수도 | 96.3% | 99%+ | 자동 검증 |
| 이모지 제거율 | 0% | 100% | grep |
| GIVEN/AND 명시율 | 67% | 100% | 수동 검사 |
| 에러 처리 완성도 | 63% | 95%+ | 코드 검사 |
| Skill 명세 완성도 | 15% | 95%+ | 카운트 |
| 다국어 예제 | 2 | 5 | 언어 수 |

### 정성 지표

| 지표 | 측정 방법 |
|------|---------|
| 사용자 만족도 | 피드백 수집 |
| 온보딩 용이성 | 신규 사용자 테스트 |
| 문서 명확성 | 리뷰 의견 |
| 국제화 품질 | 네이티브 검증 |

---

## 문의 & 지원

**개선안에 대한 질문**: [GitHub Issues]
**기여 방법**: [CONTRIBUTING.md]
**정기 검증**: 분기별 (예정: 2026-02-19)

---

**작성**: Claude Code SPEC Builder
**버전**: 1.0
**최종 업데이트**: 2025-11-19
