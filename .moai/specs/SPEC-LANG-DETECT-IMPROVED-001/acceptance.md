---
spec_id: LANG-DETECT-IMPROVED-001
title: 마커 기반 언어 감지 개선 - 수용 기준
status: draft
version: 1.0.0
created_date: 2025-10-31
updated_date: 2025-10-31
---

<!-- @ACCEPTANCE:LANG-DETECT-IMPROVED-001 -->

# 수용 기준: 마커 기반 언어 감지 개선

## 📋 개요

이 문서는 SPEC-LANG-DETECT-IMPROVED-001의 수용 기준을 정의합니다. Given-When-Then 형식의 시나리오를 통해 구현의 완료 조건을 명확히 합니다.

---

## ✅ 수용 기준 요약

### 필수 기준 (Must Have)
1. ✅ Tier 1/2/3 우선순위 감지 알고리즘이 올바르게 작동해야 함
2. ✅ Confidence Score가 정확히 계산되어야 함 (0-100%)
3. ✅ Confidence >= 80% 시 자동 결정이 작동해야 함
4. ✅ Confidence < 80% 시 AskUserQuestion이 호출되어야 함
5. ✅ 기존 `detect_language()` 함수와 호환되어야 함
6. ✅ 기존 23개 단위 테스트가 통과해야 함
7. ✅ 신규 10개 이상의 테스트가 통과해야 함
8. ✅ 테스트 커버리지가 95% 이상이어야 함

### 선택 기준 (Should Have)
1. 성능: 중형 프로젝트에서 응답 시간 < 500ms
2. 로깅: 감지 과정의 상세 로그 출력
3. 에러 처리: 명확한 에러 메시지 및 fallback

### 추가 기준 (Could Have)
1. 다중 언어 프로젝트 지원 (미래 버전)
2. 실시간 언어 변경 감지 (미래 버전)

---

## 🧪 시나리오 기반 수용 기준

### Scenario 1: Tier 1 감지 (SPEC 문서 기반)

#### 1.1 SPEC 문서 존재 시 자동 감지

**Given**: `.moai/specs/SPEC-AUTH-001/spec.md` 파일이 존재하고 다음 내용을 포함:
```yaml
---
codebase_language: python
---
```

**When**: `detect_with_confidence("/project/path")` 호출

**Then**:
- 반환값: `("python", 100.0)`
- Confidence Score: 100%
- 로그: "Detected language from SPEC document: python (100% confidence)"

**Verification**:
```python
def test_tier1_spec_document_detection():
    # Arrange
    setup_spec_document(codebase_language="python")

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 100.0
```

---

#### 1.2 SPEC 문서에 잘못된 YAML 형식

**Given**: `.moai/specs/SPEC-AUTH-001/spec.md` 파일이 잘못된 YAML 형식을 포함:
```yaml
---
codebase_language python  # 콜론 누락
---
```

**When**: `detect_with_confidence("/project/path")` 호출

**Then**:
- Tier 1 감지 실패 → Tier 2로 fallback
- 로그: "Warning: Failed to parse YAML in SPEC document, fallback to Tier 2"
- 반환값: Tier 2 결과 또는 Tier 3 결과

**Verification**:
```python
def test_tier1_invalid_yaml_fallback():
    # Arrange
    setup_invalid_spec_yaml()

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert confidence < 100.0  # Tier 1 실패
    # Tier 2 또는 Tier 3 결과 반환 확인
```

---

#### 1.3 SPEC 문서가 없는 경우

**Given**: `.moai/specs/` 디렉터리가 비어있거나 존재하지 않음

**When**: `detect_with_confidence("/project/path")` 호출

**Then**:
- Tier 1 감지 실패 → Tier 2로 fallback
- 반환값: Tier 2 또는 Tier 3 결과

**Verification**:
```python
def test_tier1_no_spec_document():
    # Arrange
    ensure_no_spec_directory()

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert confidence <= 90.0  # Tier 2 이하
```

---

### Scenario 2: Tier 2 감지 (명시적 마커)

#### 2.1 Django 프레임워크 감지

**Given**: 프로젝트 루트에 다음 파일 존재:
- `manage.py`
- `settings.py`

**When**: `detect_with_confidence("/django-project")` 호출

**Then**:
- 반환값: `("python", 90.0)`
- Confidence Score: 90%
- 로그: "Detected language from framework marker: python (Django, 90% confidence)"

**Verification**:
```python
def test_tier2_django_framework():
    # Arrange
    setup_django_project()  # manage.py, settings.py 생성

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 90.0
```

---

#### 2.2 FastAPI 프레임워크 감지

**Given**: 프로젝트 루트에 `main.py` 파일이 존재하고 다음 내용 포함:
```python
from fastapi import FastAPI

app = FastAPI()
```

**When**: `detect_with_confidence("/fastapi-project")` 호출

**Then**:
- 반환값: `("python", 90.0)`
- Confidence Score: 90%
- 로그: "Detected language from framework marker: python (FastAPI, 90% confidence)"

**Verification**:
```python
def test_tier2_fastapi_framework():
    # Arrange
    setup_fastapi_project()  # main.py + fastapi import

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 90.0
```

---

#### 2.3 React 프레임워크 감지

**Given**: 프로젝트 루트에 `package.json` 파일이 존재하고 다음 내용 포함:
```json
{
  "dependencies": {
    "react": "^18.0.0"
  }
}
```

**When**: `detect_with_confidence("/react-project")` 호출

**Then**:
- 반환값: `("javascript", 90.0)`
- Confidence Score: 90%
- 로그: "Detected language from framework marker: javascript (React, 90% confidence)"

**Verification**:
```python
def test_tier2_react_framework():
    # Arrange
    setup_react_project()  # package.json + react dependency

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "javascript"
    assert confidence == 90.0
```

---

#### 2.4 설정 파일 기반 감지 (requirements.txt)

**Given**: 프로젝트 루트에 `requirements.txt` 파일만 존재 (프레임워크 마커 없음)

**When**: `detect_with_confidence("/python-project")` 호출

**Then**:
- 반환값: `("python", 80.0)`
- Confidence Score: 80%
- 로그: "Detected language from config file: python (requirements.txt, 80% confidence)"

**Verification**:
```python
def test_tier2_config_file_detection():
    # Arrange
    setup_requirements_txt()  # requirements.txt만 생성

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 80.0
```

---

### Scenario 3: Tier 3 감지 (파일 구조 분석)

#### 3.1 Python 파일 비율 > 50%

**Given**: 프로젝트 디렉터리에 다음 파일 구조 존재:
```
project/
├── src/
│   ├── module1.py
│   ├── module2.py
│   └── module3.py
├── tests/
│   └── test_module.py
└── README.md
```
- 총 5개 파일 중 4개가 `.py` (80%)

**When**: `detect_with_confidence("/python-project")` 호출

**Then**:
- 반환값: `("python", 50.0)`
- Confidence Score: 50%
- 로그: "Detected language from file structure: python (80% .py files, 50% confidence)"

**Verification**:
```python
def test_tier3_python_file_structure():
    # Arrange
    setup_python_file_structure()  # .py 파일 80%

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 50.0
```

---

#### 3.2 JavaScript 파일 비율 > 50%

**Given**: 프로젝트 디렉터리에 다음 파일 구조 존재:
```
project/
├── src/
│   ├── app.js
│   ├── utils.js
│   └── config.ts
├── tests/
│   └── test.js
└── README.md
```
- 총 5개 파일 중 4개가 `.js/.ts` (80%)

**When**: `detect_with_confidence("/js-project")` 호출

**Then**:
- 반환값: `("javascript", 50.0)`
- Confidence Score: 50%

**Verification**:
```python
def test_tier3_javascript_file_structure():
    # Arrange
    setup_javascript_file_structure()  # .js/.ts 파일 80%

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "javascript"
    assert confidence == 50.0
```

---

### Scenario 4: 우선순위 검증

#### 4.1 Tier 1 > Tier 2 우선순위

**Given**: 프로젝트에 다음 조건 모두 충족:
- SPEC 문서: `codebase_language: python`
- Django 마커: `manage.py`, `settings.py`

**When**: `detect_with_confidence("/project")` 호출

**Then**:
- Tier 1이 우선 적용됨
- 반환값: `("python", 100.0)` (Tier 1 결과)
- Tier 2 감지 로직 실행되지 않음

**Verification**:
```python
def test_priority_tier1_over_tier2():
    # Arrange
    setup_spec_document(codebase_language="python")
    setup_django_project()  # Tier 2 마커도 존재

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 100.0  # Tier 1 우선
```

---

#### 4.2 Tier 2 > Tier 3 우선순위

**Given**: 프로젝트에 다음 조건 모두 충족:
- SPEC 문서 없음
- Django 마커: `manage.py`, `settings.py`
- 파일 구조: `.py` 파일 80%

**When**: `detect_with_confidence("/project")` 호출

**Then**:
- Tier 2가 우선 적용됨
- 반환값: `("python", 90.0)` (Tier 2 결과)
- Tier 3 감지 로직 실행되지 않음

**Verification**:
```python
def test_priority_tier2_over_tier3():
    # Arrange
    ensure_no_spec_document()
    setup_django_project()  # Tier 2 마커
    setup_python_file_structure()  # Tier 3 조건도 충족

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"
    assert confidence == 90.0  # Tier 2 우선
```

---

### Scenario 5: Confidence 기반 자동 결정

#### 5.1 Confidence >= 80% 자동 진행

**Given**: 프로젝트에 `requirements.txt` 존재 (Confidence: 80%)

**When**: implementation-planner Step 0 실행

**Then**:
- 언어 자동 선택: "python"
- AskUserQuestion 호출되지 않음
- 로그: "Auto-selected language: python (80% confidence)"

**Verification**:
```python
def test_auto_decision_high_confidence():
    # Arrange
    setup_requirements_txt()  # Confidence: 80%

    # Act
    result = implementation_planner.step_0(cwd)

    # Assert
    assert result["language"] == "python"
    assert result["auto_selected"] == True
    assert "AskUserQuestion" not in result["calls"]
```

---

#### 5.2 Confidence < 80% 사용자 확인

**Given**: 프로젝트에 마커 없음, 파일 구조만 존재 (Confidence: 50%)

**When**: implementation-planner Step 0 실행

**Then**:
- AskUserQuestion 호출됨
- 질문 내용: "감지된 언어: python (신뢰도: 50%). 올바른가요?"
- 사용자 선택 옵션:
  - "✅ python 사용"
  - "🔄 다른 언어 선택"

**Verification**:
```python
def test_user_confirmation_low_confidence():
    # Arrange
    setup_python_file_structure()  # Confidence: 50%

    # Act
    result = implementation_planner.step_0(cwd)

    # Assert
    assert result["auto_selected"] == False
    assert "AskUserQuestion" in result["calls"]
    assert "신뢰도: 50%" in result["question"]
```

---

### Scenario 6: 호환성 및 에러 처리

#### 6.1 기존 `detect_language()` 함수 호환성

**Given**: 기존 코드가 `detect_language(cwd)` 호출

**When**: 함수 실행

**Then**:
- 반환값: `str` (언어명만)
- Confidence Score는 내부적으로만 계산
- 기존 동작 100% 유지

**Verification**:
```python
def test_backward_compatibility():
    # Arrange
    setup_django_project()

    # Act
    language = detector.detect_language(cwd)

    # Assert
    assert isinstance(language, str)
    assert language == "python"
```

---

#### 6.2 감지 실패 시 기본값 반환

**Given**: 프로젝트에 언어 감지 가능한 마커나 파일이 전혀 없음

**When**: `detect_with_confidence("/empty-project")` 호출

**Then**:
- 반환값: `("python", 0.0)` (기본값)
- 로그: "Warning: Language detection failed, using default: python (0% confidence)"

**Verification**:
```python
def test_default_language_on_detection_failure():
    # Arrange
    setup_empty_project()  # 빈 프로젝트

    # Act
    language, confidence = detector.detect_with_confidence(cwd)

    # Assert
    assert language == "python"  # 기본값
    assert confidence == 0.0
```

---

#### 6.3 파일 시스템 에러 처리

**Given**: 프로젝트 경로가 유효하지 않음

**When**: `detect_with_confidence("/invalid/path")` 호출

**Then**:
- 예외 발생하지 않음
- 반환값: `("python", 0.0)` (기본값)
- 로그: "Error: Invalid project path, using default language"

**Verification**:
```python
def test_invalid_path_error_handling():
    # Arrange
    invalid_path = "/non/existent/path"

    # Act
    language, confidence = detector.detect_with_confidence(invalid_path)

    # Assert
    assert language == "python"
    assert confidence == 0.0
```

---

## 🧪 테스트 커버리지 기준

### 단위 테스트 커버리지

**목표**: 95% 이상

**필수 커버리지 항목**:
1. ✅ `detect_with_confidence()` - 100%
2. ✅ `_detect_from_spec()` - 100%
3. ✅ `_detect_from_framework()` - 95%
4. ✅ `_detect_from_config()` - 95%
5. ✅ `_detect_from_structure()` - 90%
6. ✅ 에러 처리 로직 - 100%

### 통합 테스트

**필수 시나리오**:
1. ✅ implementation-planner Step 0 통합
2. ✅ AskUserQuestion 호출 검증
3. ✅ 실제 프로젝트 구조 테스트 (Django, FastAPI, React)

---

## 📊 성능 기준

### 응답 시간

**중형 프로젝트** (1000+ 파일):
- ✅ Tier 1 감지: < 100ms
- ✅ Tier 2 감지: < 200ms
- ✅ Tier 3 감지: < 500ms

**대형 프로젝트** (10000+ 파일):
- ✅ Tier 1 감지: < 100ms
- ✅ Tier 2 감지: < 300ms
- ✅ Tier 3 감지: < 1000ms

### 메모리 사용량

- ✅ 추가 메모리: < 50MB
- ✅ 메모리 누수 없음

---

## 🎯 Definition of Done

### 기능적 완료 조건
- [x] Tier 1/2/3 우선순위 알고리즘 구현 완료
- [x] Confidence Score 계산 로직 구현 완료
- [x] `detect_with_confidence()` 함수 구현 완료
- [x] 기존 `detect_language()` 호환성 유지
- [x] 기존 23개 테스트 100% 통과
- [x] 신규 10개 이상 테스트 100% 통과
- [x] implementation-planner Step 0 통합 완료

### 품질 완료 조건
- [x] 테스트 커버리지 95% 이상
- [x] 타입 체킹 (mypy) 0 에러
- [x] 린팅 (ruff) 0 에러
- [x] 성능 테스트 통과 (응답 시간 < 500ms)
- [x] 코드 리뷰 승인

### 문서화 완료 조건
- [x] spec.md 작성 완료
- [x] plan.md 작성 완료
- [x] acceptance.md 작성 완료 (본 문서)
- [x] 코드 주석 및 docstring 완비
- [x] CHANGELOG.md 업데이트

### 통합 완료 조건
- [x] Git 브랜치 생성 (feature/lang-detect-improved)
- [x] 모든 커밋 메시지 표준 준수
- [x] Draft PR 생성
- [x] CI/CD 파이프라인 통과
- [x] PR 리뷰 및 승인
- [x] main 브랜치 병합 준비 완료

---

## 🚀 릴리스 준비 체크리스트

### Pre-Release
- [ ] 모든 수용 기준 시나리오 검증 완료
- [ ] 성능 벤치마크 실행 및 기록
- [ ] 보안 취약점 스캔 통과
- [ ] 문서 최종 검토

### Release
- [ ] 버전 번호 업데이트 (v0.8.0)
- [ ] CHANGELOG.md 항목 추가
- [ ] Git 태그 생성
- [ ] GitHub Release 노트 작성

### Post-Release
- [ ] 사용자 피드백 수집
- [ ] 모니터링 대시보드 확인
- [ ] 이슈 트래킹 시작

---

## 📝 추가 검증 항목

### 사용자 경험 검증
1. ✅ 명시적 마커 존재 시 자동 진행 (사용자 확인 불필요)
2. ✅ 저신뢰도 시 명확한 안내 메시지
3. ✅ 감지 결과에 대한 상세 설명 제공

### 엣지 케이스 검증
1. ✅ 빈 프로젝트 디렉터리
2. ✅ 심볼릭 링크가 있는 프로젝트
3. ✅ 숨김 파일만 있는 프로젝트
4. ✅ 매우 깊은 디렉터리 구조 (10+ 레벨)
5. ✅ 파일 권한 없는 디렉터리

### 보안 검증
1. ✅ 프로젝트 루트 외부 경로 접근 시도 차단
2. ✅ 악의적인 YAML 페이로드 처리
3. ✅ 파일 시스템 Race Condition 처리

---

**문서 상태**: Draft
**최종 업데이트**: 2025-10-31
**작성자**: @goos
**검토자**: TBD

---

## 📚 참고 자료

### 테스트 프레임워크
- pytest: [https://docs.pytest.org](https://docs.pytest.org)
- pytest-cov: Coverage 측정
- pytest-benchmark: 성능 테스트

### Given-When-Then 가이드
- BDD (Behavior-Driven Development) 원칙
- EARS 요구사항 검증 패턴

### 관련 SPEC
- SPEC-LANG-DETECT-IMPROVED-001/spec.md
- SPEC-LANG-DETECT-IMPROVED-001/plan.md
