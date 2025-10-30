---
spec_id: LANG-DETECT-IMPROVED-001
title: 마커 기반 언어 감지 개선
status: draft
priority: high
category: feature/enhancement
author: "@goos"
created_date: 2025-10-31
updated_date: 2025-10-31
version: 1.0.0
related_specs:
  - SPEC-LANG-DETECT-001
  - SPEC-LANG-FIX-001
tags:
  - language-detection
  - confidence-score
  - implementation-planner
  - performance
---

<!-- @SPEC:LANG-DETECT-IMPROVED-001 -->

# SPEC: 마커 기반 언어 감지 개선

## 📋 개요

### 목적
프로젝트 언어 감지 정확도를 향상시키고, Confidence Score 기반 자동 결정 시스템을 구축하여 사용자 경험을 개선합니다.

### 배경
- **Issue #131 해결**: SessionStart Hook에서 언어 감지 로직 제거 (성능 최적화)
- **책임 이전**: implementation-planner의 Step 0에서 on-demand 언어 감지 수행
- **개선 필요성**: 명시적 마커(프레임워크, 설정 파일) 기반 감지 로직 부족
- **사용자 경험**: 저신뢰도 시에만 사용자 확인 요청 (자동화 강화)

### 범위
- **포함**:
  - LanguageDetector 클래스 확장 (Confidence Score 계산)
  - 3단계 우선순위 감지 알고리즘 구현
  - implementation-planner Step 0 통합
  - 단위 테스트 및 통합 테스트 추가
- **제외**:
  - 다중 언어 프로젝트 동시 감지 (미래 버전)
  - 실시간 언어 변경 감지 (현재 범위 외)

---

## 🎯 요구사항 (EARS 형식)

### Environment (환경)

**E1**: 시스템은 Python 3.13+ 환경에서 실행됩니다.

**E2**: 시스템은 `/Users/goos/MoAI/MoAI-ADK/` 프로젝트 구조를 기반으로 합니다.

**E3**: 시스템은 `.claude/agents/alfred/implementation-planner.md`의 Step 0와 통합됩니다.

**E4**: 시스템은 기존 23개 단위 테스트와 호환되어야 합니다.

### Assumptions (가정)

**A1**: 프로젝트 루트 디렉터리는 유효한 경로로 제공됩니다.

**A2**: SPEC 문서가 존재하는 경우 `codebase_language` 필드를 포함합니다.

**A3**: 명시적 마커(프레임워크, 설정 파일)는 프로젝트 루트 또는 하위 디렉터리에 위치합니다.

**A4**: 사용자는 Confidence Score >= 80% 시 자동 결정을 신뢰합니다.

### Ubiquitous Requirements (항상 적용)

**U1**: 시스템은 프로젝트 언어를 3단계 우선순위로 감지해야 합니다.
- **Tier 1**: SPEC 문서 명시 (`codebase_language`) → 100점
- **Tier 2**: 명시적 마커 (프레임워크: 90점, 설정 파일: 80점)
- **Tier 3**: 파일 구조 분석 (확장자: 50점)

**U2**: 시스템은 Confidence Score (0-100%)를 반환해야 합니다.
- **계산 방식**: `(detected_score / 100) * 100%`
- **예시**: Tier 1 감지 → 100%, Tier 2 감지 → 80-90%

**U3**: 시스템은 기존 `detect_language(cwd)` 함수와 호환되어야 합니다.
- **신규 함수**: `detect_with_confidence(cwd) -> Tuple[str, float]`
- **기존 함수**: 기존 동작 유지 (호환성 보장)

**U4**: 시스템은 감지 실패 시 기본값 "python"을 반환해야 합니다.

### Event-driven Requirements (이벤트 기반)

**EV1**: WHEN `/alfred:2-run` 실행 AND implementation-planner Step 0 진입
THEN 시스템은 `detect_with_confidence(cwd)`를 호출하여 언어와 신뢰도를 계산합니다.

**EV2**: WHEN Confidence Score >= 80%
THEN 시스템은 자동으로 언어를 선택하고 사용자 확인을 건너뜁니다.

**EV3**: WHEN Confidence Score < 80%
THEN 시스템은 AskUserQuestion을 호출하여 분석 결과와 함께 사용자에게 언어 선택을 요청합니다.

**EV4**: WHEN SPEC 문서에 `codebase_language` 필드 존재
THEN 시스템은 즉시 Tier 1 감지를 수행하고 100% 신뢰도를 반환합니다.

**EV5**: WHEN 명시적 마커 (예: `package.json`, `requirements.txt`) 발견
THEN 시스템은 Tier 2 감지를 수행하고 80-90% 신뢰도를 반환합니다.

### State-driven Requirements (상태 기반)

**ST1**: WHILE LanguageDetector 초기화 상태
THEN `LANGUAGE_PATTERNS` 및 `FRAMEWORK_PATTERNS` 딕셔너리가 메모리에 로드되어야 합니다.

**ST2**: WHILE 프로젝트에 명시적 마커 존재
THEN 시스템은 파일 구조 분석(Tier 3)보다 마커 기반 감지(Tier 2)를 우선 수행합니다.

**ST3**: WHILE 감지 프로세스 진행 중
THEN 시스템은 가장 높은 점수를 가진 언어를 후보로 선택합니다.

### Optional Requirements (선택 요구사항)

**O1**: IF 사용자가 여러 언어 프로젝트인 경우
THEN 시스템은 주요 언어(highest confidence)를 제시하고 대체 언어 목록을 제공할 수 있습니다.

**O2**: IF 디버그 모드 활성화
THEN 시스템은 감지 과정의 상세 로그를 출력할 수 있습니다.

### Unwanted Behaviors (금지 동작)

**UW1**: IF SPEC 문서에 명시된 언어를 무시
THEN 시스템은 불정확한 감지 결과를 반환합니다 (버그).

**UW2**: IF 명시적 마커 존재 시에도 저신뢰도(<80%)로 사용자 확인 요청
THEN 사용자 경험이 저하됩니다.

**UW3**: IF 동일한 프로젝트에서 반복 호출 시 다른 결과 반환
THEN 시스템 신뢰성이 저하됩니다 (일관성 위반).

---

## 🏗️ 시스템 설계

### 핵심 컴포넌트

#### 1. LanguageDetector 클래스
```python
class LanguageDetector:
    """프로젝트 언어 감지 및 신뢰도 계산"""

    def detect_language(self, cwd: str) -> str:
        """기존 함수 (호환성 유지)"""
        language, _ = self.detect_with_confidence(cwd)
        return language

    def detect_with_confidence(self, cwd: str) -> Tuple[str, float]:
        """언어 감지 + 신뢰도 점수 반환"""
        # Tier 1: SPEC 문서 확인 (100점)
        # Tier 2: 명시적 마커 확인 (80-90점)
        # Tier 3: 파일 구조 분석 (50점)
        pass
```

#### 2. 감지 우선순위 알고리즘

**Tier 1: SPEC 문서 기반 (100점)**
- `.moai/specs/SPEC-*/spec.md` 파일 검색
- YAML frontmatter에서 `codebase_language` 필드 추출
- 예시: `codebase_language: python` → 100% 신뢰도

**Tier 2: 명시적 마커 (80-90점)**
- **프레임워크 마커** (90점):
  - Django: `manage.py`, `settings.py`
  - FastAPI: `main.py` + `from fastapi import`
  - React: `package.json` + `"react"` 의존성
- **설정 파일** (80점):
  - Python: `requirements.txt`, `pyproject.toml`, `setup.py`
  - JavaScript: `package.json`, `tsconfig.json`
  - Go: `go.mod`

**Tier 3: 파일 구조 분석 (50점)**
- 확장자 기반 통계:
  - `.py` 파일 > 50% → python (50점)
  - `.js/.ts` 파일 > 50% → javascript (50점)

### 데이터 구조

#### Confidence Score 계산
```python
confidence_mapping = {
    "spec_document": 100,      # Tier 1
    "framework_marker": 90,    # Tier 2
    "config_file": 80,         # Tier 2
    "file_extension": 50,      # Tier 3
}
```

#### FRAMEWORK_PATTERNS 확장
```python
FRAMEWORK_PATTERNS = {
    "python": {
        "django": ["manage.py", "settings.py", "wsgi.py"],
        "fastapi": ["main.py + fastapi import"],
        "flask": ["app.py + flask import"],
    },
    "javascript": {
        "react": ["package.json + react dependency"],
        "vue": ["package.json + vue dependency"],
    },
    # ... 기타 언어
}
```

---

## 🔗 추적성 (Traceability)

### TAG 체인
- **SPEC**: @SPEC:LANG-DETECT-IMPROVED-001
- **CODE**: @CODE:LANG-DETECT-IMPROVED-001
- **TEST**: @TEST:LANG-DETECT-IMPROVED-001
- **DOC**: @DOC:LANG-DETECT-IMPROVED-001

### 관련 파일
- **구현**: `src/moai_adk/core/project/detector.py`
- **테스트**: `tests/unit/test_detector.py`
- **통합**: `.claude/agents/alfred/implementation-planner.md`
- **문서**: `.moai/specs/SPEC-LANG-DETECT-IMPROVED-001/`

### 의존성
- **이전 SPEC**: SPEC-LANG-DETECT-001, SPEC-LANG-FIX-001
- **관련 Issue**: #131 (SessionStart Hook 성능 최적화)
- **통합 지점**: implementation-planner Step 0

---

## 📊 품질 기준

### 테스트 커버리지
- **목표**: 95% 이상
- **기존 테스트**: 23개 유지
- **신규 테스트**: 10개 이상 추가
  - Confidence Score 계산 검증
  - Tier 1/2/3 우선순위 검증
  - Edge case 처리

### 성능 요구사항
- **응답 시간**: < 500ms (중형 프로젝트 기준)
- **메모리**: < 50MB (추가 메모리 사용)

### 보안 요구사항
- **경로 탐색**: 프로젝트 루트 외부 접근 금지
- **파일 읽기**: 읽기 전용 (수정 금지)

---

## 🚀 구현 전략

### Phase 1: LanguageDetector 확장
- `detect_with_confidence()` 메서드 추가
- Confidence Score 계산 로직 구현

### Phase 2: Tier 1 감지 (SPEC 문서)
- `.moai/specs/` 디렉터리 탐색
- YAML frontmatter 파싱

### Phase 3: Tier 2 감지 (명시적 마커)
- FRAMEWORK_PATTERNS 확장
- 설정 파일 탐지 로직

### Phase 4: implementation-planner 통합
- Step 0에서 `detect_with_confidence()` 호출
- Confidence 기반 자동 결정 로직

### Phase 5: 테스트 및 검증
- 단위 테스트 작성
- 통합 테스트 실행
- Edge case 검증

---

## 📚 참고 자료

### 기술 스택
- Python 3.13+
- pytest (최신 안정 버전)
- mypy (최신 안정 버전)
- ruff (최신 안정 버전)

### 관련 문서
- `.moai/memory/language-config-schema.md`
- `CLAUDE-RULES.md` (Language Handling 섹션)
- `CLAUDE-PRACTICES.md` (Context Engineering)

---

## ✅ Definition of Done

1. ✅ `detect_with_confidence()` 함수 구현 완료
2. ✅ 3단계 우선순위 알고리즘 동작 확인
3. ✅ Confidence Score 계산 정확도 검증
4. ✅ 기존 23개 테스트 통과
5. ✅ 신규 10개 테스트 작성 및 통과
6. ✅ implementation-planner Step 0 통합 완료
7. ✅ 테스트 커버리지 95% 이상
8. ✅ 코드 리뷰 승인
9. ✅ 문서 업데이트 완료
10. ✅ GitHub Issue #131 연계 검증

---

**문서 상태**: Draft
**최종 업데이트**: 2025-10-31
**작성자**: @goos
**검토자**: TBD
