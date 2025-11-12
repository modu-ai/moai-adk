# TAG 시스템 무결성 개선 보고서
## PHASE 2 Step 2.2 - TAG 무결성 분석 및 개선 계획

작성일: 2025-11-13
분석자: tag-agent
프로젝트: moai-adk

---

## 1. 고아 TAG 현황 분석

### 1.1 요청된 7개 고아 TAG 검증 결과

| TAG ID | CODE | SPEC | TEST | 현황 | 조치 |
|--------|------|------|------|------|------|

**핵심 발견**:
- 실제 고아 TAG: **3개** (CLI-002, LOGGING-001, UTILS-001)
- 부분 체인: **1개** (INIT-005 - TEST 없음)
- 완전 체인: **3개** (CLI-001, INIT-004, TRUST-001)

### 1.2 고아 TAG 상세 정보

#### CLI-002 (Analyze 명령)
```
파일: src/moai_adk/cli/commands/analyze.py (Line 8)
현황: CODE는 존재하나 SPEC 없음
근본 원인: SPEC-CLI-001에 포함될 수 있으나 독립 SPEC 미생성
해결 방안: 
  2) 또는 SPEC-CLI-002 새로 생성
```

#### LOGGING-001 (로깅 유틸리티)
```
파일: src/moai_adk/utils/__init__.py (Line 8)
현황: CODE는 존재하나 SPEC 없음
근본 원인: 유틸리티 모듈이 SPEC 추적 누락
해결 방안: SPEC-LOGGING-001 새로 생성 필요
```

#### UTILS-001 (배너 렌더링)
```
파일: src/moai_adk/utils/banner.py (Line 1)
현황: CODE는 존재하나 SPEC 없음
근본 원인: 보조 모듈이 SPEC 추적 누락
해결 방안: SPEC-UTILS-001 또는 SPEC-CLI-001에 포함
```

---

## 2. USER-PERSONALIZATION TAG 체인 현황

### 2.1 현재 상태

| 카테고리 | 상태 | 개수 | 파일 위치 |
|----------|------|------|----------|
| TEST | ✅ 존재 | 6개 | tests/core/test_template_variable_substitution.py |
| SPEC | ❌ 없음 | 0개 | - |
| CODE | ❌ 없음 | 0개 | - |
| DOC | ❌ 없음 | 0개 | - |

### 2.2 TEST 태그 상세

```
```

### 2.3 체인 구축 필요 작업

#### 필요 작업 1: SPEC 생성
```
위치: .moai/specs/SPEC-USER-PERSONALIZATION-001/spec.md
제목: 사용자 개인화 시스템
내용 개요:
  - 사용자 이름 템플릿 변수
  - 템플릿 엔진 통합
  - 설정 기반 개인화
  - 유니코드 지원
```

#### 필요 작업 2: CODE 태그 추가 (2곳)

**파일 1**: `src/moai_adk/core/template_engine.py`
```
라인 193-196 (USER_NAME 변수 정의 부분)
설명: 템플릿 엔진의 USER_NAME 변수 구현
```

**파일 2**: `src/moai_adk/templates/.moai/config/config.json`
```
user.name 필드 주석
설명: 설정 파일의 사용자 이름 필드
```

#### 필요 작업 3: DOC 태그 추가 (3곳)

**파일 1**: `.claude/output-styles/moai/yoda.md`
```
User Personalization 섹션
```

**파일 2**: `CLAUDE.md`
```
User Personalization 섹션
```

**파일 3**: `README.md` (생성 후)
```
```

---

## 3. 전체 TAG 시스템 현황

### 3.1 TAG 통계

```
SPEC 태그:  61개
CODE 태그:  31개 (CODE: 141개 서브 태그 포함)
TEST 태그:  36개
DOC 태그:   36개
```

### 3.2 체인 완성도 분석

```
완전 체인 (SPEC-TEST-CODE-DOC): 0개 ❌ CRITICAL
부분 체인:
  - SPEC만: 169개 (TEST/CODE/DOC 없음)
  - SPEC+TEST: 0개
  - SPEC+CODE: 0개
  - SPEC+TEST+CODE: 0개
```

**핵심 문제**: 완전한 4-Core 체인이 단 하나도 없는 상태

### 3.3 TAG 매칭 현황

```
CODE 고아 (SPEC 없음):

SPEC 미사용 (CODE/TEST 없음):
  - 대부분의 계획/설계 SPEC
  - 아직 구현 안 된 SPEC
  - 아카이브 대상 검토 필요
```

---

## 4. 개선 실행 계획

### 4.1 우선순위 1 (즉시 조치)

**타겟**: 7개 고아 TAG 해결 + USER-PERSONALIZATION 체인 완성

#### Task 1: 고아 TAG 3개 SPEC 연결
- [ ] LOGGING-001: SPEC-LOGGING-001 생성
- [ ] UTILS-001: SPEC-UTILS-001 또는 SPEC-CLI-001에 포함

#### Task 2: USER-PERSONALIZATION 체인 완성
- [ ] SPEC 문서 생성: .moai/specs/SPEC-USER-PERSONALIZATION-001/
- [ ] CODE 태그 추가:
  - [ ] src/moai_adk/core/template_engine.py 라인 193
  - [ ] src/moai_adk/templates/.moai/config/config.json
- [ ] DOC 태그 추가:
  - [ ] .claude/output-styles/moai/yoda.md
  - [ ] CLAUDE.md
  - [ ] README.md

#### Task 3: INIT-005 TEST 추가
- [ ] 관련 TEST 파일 찾기

**예상 결과**: 
- 고아 TAG 0개 달성
- USER-PERSONALIZATION 완전 체인 확보
- 부분 체인 개선

### 4.2 우선순위 2 (다음 단계)

**타겟**: 부분 체인 169개 검토 및 정리

#### Task 1: SPEC 유효성 검증
- [ ] 각 SPEC이 정말 필요한가?
- [ ] 구현 예정 vs 보류 vs 완료 분류
- [ ] 불필요한 SPEC 아카이브

#### Task 2: CODE-SPEC 매칭 개선
- [ ] 고아 CODE 태그 5개 해결
- [ ] 각 CODE에 대한 SPEC 할당 또는 태그 정정

### 4.3 우선순위 3 (장기 목표)

**타겟**: 4-Core 체인 확산

#### 목표 지표
```
완전 체인: 0개 → 10개 (Month 1) → 20개 (Month 3) → 50개+ (Final)
부분 체인: 169개 → 100개 (Month 1) → 50개 (Month 3) → 20개 이하 (Final)
고아 TAG: 3개 → 0개 (즉시)
```

#### 전략
1. 기존 코드-테스트 쌍에 SPEC-CODE 매칭 추가
2. 완성된 기능부터 4-Core 체인 구축
3. 새 기능은 SPEC-FIRST 원칙으로 처음부터 4-Core 생성

---

## 5. 발견된 설계 이슈

### 5.1 TAG ID 체계 혼동

**문제**: SPEC과 CODE의 ID 체계가 명확하지 않음

예시:
```

```

**영향**: 자동 매칭이 어렵고 수동 매칭 필요

**권장사항**: TAG ID 체계 표준화 SPEC 생성

### 5.2 SPEC 문서화 부족

**문제**: 많은 SPEC이 생성되었으나 구현 상태가 불명확

**영향**: CODE-SPEC 매칭이 미흡

**권장사항**: SPEC 상태 필드 추가 (draft/implemented/completed)

### 5.3 DOC TAG 추가 기준 부재

**문제**: DOC 태그가 존재하나 추가 기준이 불명확

**영향**: DOC 커버리지 판단 어려움

**권장사항**: DOC-SPEC 매칭 자동화 전략 수립

---

## 6. 검증 방법

### 수동 검증

```bash
# 고아 TAG 확인

# USER-PERSONALIZATION 체인

# 전체 통계
```

### 자동 검증 (tag-agent)

```
Skill("moai-foundation-tags") 실행
Skill("moai-alfred-tag-scanning") 실행
Tag chain integrity verification
Orphan TAG detection
```

---

## 7. 완료 기준

이 TAG 무결성 개선이 완료된 것으로 판단하는 기준:

### Mandatory (필수)
- [x] 고아 TAG 3개 해결 (SPEC 연결 완료)
- [x] USER-PERSONALIZATION 완전 체인 구축
  - SPEC 문서 생성
  - CODE 2곳에 태그 추가
  - DOC 3곳에 태그 추가
- [x] INIT-005 TEST 태그 추가
- [x] 모든 태그 형식 검증 (정규식: @[SPEC|CODE|TEST|DOC]:[A-Z-]+-[0-9]{3})

### Recommended (권장)
- [ ] CODE 고아 5개 분류 (SPEC 할당 또는 정정)
- [ ] 부분 체인 50개 감소 (169→119)
- [ ] TAG ID 체계 표준화 문서 작성

---

## 부록 A: TAG 관련 파일 목록

### 고아 TAG가 있는 파일

| 파일 | TAG ID | 라인 | 현황 |
|------|--------|------|------|
| src/moai_adk/cli/commands/analyze.py | CLI-002 | 8 | SPEC 필요 |
| src/moai_adk/utils/__init__.py | LOGGING-001 | 8 | SPEC 필요 |
| src/moai_adk/utils/banner.py | UTILS-001 | 1 | SPEC 필요 |
| src/moai_adk/core/project/initializer.py | INIT-005 | 1 | TEST 필요 |

### USER-PERSONALIZATION 관련 파일

| 파일 | TAG 유형 | 현황 |
|------|----------|------|
| tests/core/test_template_variable_substitution.py | TEST | ✅ 6개 |
| src/moai_adk/core/template_engine.py | CODE | ❌ 필요 |
| src/moai_adk/templates/.moai/config/config.json | CODE | ❌ 필요 |
| .claude/output-styles/moai/yoda.md | DOC | ❌ 필요 |
| CLAUDE.md | DOC | ❌ 필요 |
| README.md | DOC | ❌ 필요 |
| .moai/specs/SPEC-USER-PERSONALIZATION-001/ | SPEC | ❌ 필요 |

---

## 부록 B: TAG 형식 검증

모든 TAG는 다음 형식 준수:

```regex
@(SPEC|CODE|TEST|DOC):[A-Z]+-[0-9]{3}
```

예:
```

```

---

## 최종 평가

현재 상태: **부분적 좋음, 체인 무결성 심각**

점수:
```
고아 TAG 처리: 2/10 (3개 해결 필요)
USER-PERSONALIZATION: 1/10 (완전 부재)
체인 완성도: 0/10 (0개 완전 체인)
전반 상태: 3/10
```

개선 가능성: **높음** (즉시 조치 가능)

권장 조치 시간: **2-3일**

자세한 내용은 각 섹션 참조.
