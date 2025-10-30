---
spec_id: LANG-DETECT-IMPROVED-001
title: 마커 기반 언어 감지 개선 - 구현 계획
status: draft
version: 1.0.0
created_date: 2025-10-31
updated_date: 2025-10-31
---

<!-- @PLAN:LANG-DETECT-IMPROVED-001 -->

# 구현 계획: 마커 기반 언어 감지 개선

## 📋 개요

이 문서는 SPEC-LANG-DETECT-IMPROVED-001의 구현 계획을 정의합니다. 5단계 Phase로 구성되며, 각 단계는 독립적으로 검증 가능합니다.

---

## 🎯 구현 목표

### 주요 목표
1. ✅ LanguageDetector 클래스에 Confidence Score 계산 기능 추가
2. ✅ 3단계 우선순위 감지 알고리즘 구현
3. ✅ implementation-planner Step 0과 통합
4. ✅ 테스트 커버리지 95% 이상 달성
5. ✅ 기존 시스템과의 호환성 유지

### 부목표
- 성능 최적화 (응답 시간 < 500ms)
- 확장 가능한 프레임워크 패턴 구조 설계
- 명확한 에러 처리 및 로깅

---

## 📦 Phase별 구현 계획

### Phase 1: LanguageDetector 핵심 확장

**목표**: `detect_with_confidence()` 메서드 추가 및 Confidence Score 계산 로직 구현

#### 구현 작업
1. **신규 메서드 추가**:
   ```python
   def detect_with_confidence(self, cwd: str) -> Tuple[str, float]:
       """
       프로젝트 언어 감지 및 신뢰도 점수 반환

       Returns:
           Tuple[str, float]: (언어명, confidence score 0-100)
       """
       pass
   ```

2. **Confidence 계산 구조 설계**:
   ```python
   CONFIDENCE_LEVELS = {
       "spec_document": 100,      # Tier 1
       "framework_marker": 90,    # Tier 2
       "config_file": 80,         # Tier 2
       "file_extension": 50,      # Tier 3
       "default": 0,              # 감지 실패
   }
   ```

3. **기존 함수 리팩토링**:
   ```python
   def detect_language(self, cwd: str) -> str:
       """기존 함수 - 호환성 유지"""
       language, _ = self.detect_with_confidence(cwd)
       return language
   ```

#### 검증 기준
- ✅ `detect_with_confidence()` 함수가 Tuple[str, float] 반환
- ✅ 기존 `detect_language()` 함수 동작 유지
- ✅ 타입 힌트 및 docstring 완비

#### 예상 변경 파일
- `src/moai_adk/core/project/detector.py` (확장)

---

### Phase 2: Tier 1 감지 구현 (SPEC 문서 기반)

**목표**: `.moai/specs/` 디렉터리에서 SPEC 문서를 탐색하고 `codebase_language` 필드 추출

#### 구현 작업
1. **SPEC 문서 탐색 로직**:
   ```python
   def _detect_from_spec(self, cwd: str) -> Optional[Tuple[str, float]]:
       """
       .moai/specs/SPEC-*/spec.md 파일에서 언어 추출

       Returns:
           Optional[Tuple[str, float]]: (언어명, 100.0) 또는 None
       """
       spec_dir = Path(cwd) / ".moai" / "specs"
       if not spec_dir.exists():
           return None

       # SPEC-* 디렉터리 탐색
       for spec_path in spec_dir.glob("SPEC-*/spec.md"):
           # YAML frontmatter 파싱
           # codebase_language 필드 추출
           pass

       return None
   ```

2. **YAML frontmatter 파싱**:
   - `---` 블록 감지
   - `codebase_language:` 필드 추출
   - 유효성 검증 (지원 언어 목록 확인)

3. **에러 처리**:
   - 잘못된 YAML 형식 → 무시하고 다음 Tier로
   - 지원되지 않는 언어 → 경고 로그 + 다음 Tier로

#### 검증 기준
- ✅ SPEC 문서 존재 시 100% confidence 반환
- ✅ SPEC 문서 없을 시 None 반환 (다음 Tier로)
- ✅ 잘못된 YAML 형식 처리 확인

#### 예상 변경 파일
- `src/moai_adk/core/project/detector.py` (메서드 추가)
- `tests/unit/test_detector.py` (Tier 1 테스트)

---

### Phase 3: Tier 2 감지 구현 (명시적 마커)

**목표**: 프레임워크 마커 및 설정 파일 기반 언어 감지

#### 구현 작업
1. **FRAMEWORK_PATTERNS 확장**:
   ```python
   FRAMEWORK_PATTERNS = {
       "python": {
           "django": {
               "files": ["manage.py", "settings.py"],
               "confidence": 90
           },
           "fastapi": {
               "files": ["main.py"],
               "content_check": "from fastapi import",
               "confidence": 90
           },
           "flask": {
               "files": ["app.py"],
               "content_check": "from flask import",
               "confidence": 90
           },
       },
       "javascript": {
           "react": {
               "files": ["package.json"],
               "content_check": '"react"',
               "confidence": 90
           },
           "vue": {
               "files": ["package.json"],
               "content_check": '"vue"',
               "confidence": 90
           },
       },
       "go": {
           "standard": {
               "files": ["go.mod"],
               "confidence": 90
           },
       },
       # ... 기타 언어
   }
   ```

2. **설정 파일 패턴**:
   ```python
   CONFIG_FILE_PATTERNS = {
       "python": ["requirements.txt", "pyproject.toml", "setup.py", "Pipfile"],
       "javascript": ["package.json", "tsconfig.json"],
       "go": ["go.mod"],
       "rust": ["Cargo.toml"],
       "java": ["pom.xml", "build.gradle"],
   }
   ```

3. **프레임워크 감지 로직**:
   ```python
   def _detect_from_framework(self, cwd: str) -> Optional[Tuple[str, float]]:
       """
       프레임워크 마커 기반 감지 (90점)
       """
       for lang, frameworks in FRAMEWORK_PATTERNS.items():
           for framework, pattern in frameworks.items():
               if self._check_framework_pattern(cwd, pattern):
                   return (lang, 90.0)
       return None
   ```

4. **설정 파일 감지 로직**:
   ```python
   def _detect_from_config(self, cwd: str) -> Optional[Tuple[str, float]]:
       """
       설정 파일 기반 감지 (80점)
       """
       for lang, config_files in CONFIG_FILE_PATTERNS.items():
           for config_file in config_files:
               if (Path(cwd) / config_file).exists():
                   return (lang, 80.0)
       return None
   ```

#### 검증 기준
- ✅ Django 프로젝트 → python (90%)
- ✅ FastAPI 프로젝트 → python (90%)
- ✅ React 프로젝트 → javascript (90%)
- ✅ requirements.txt만 있는 프로젝트 → python (80%)
- ✅ 여러 마커 존재 시 최고 점수 반환

#### 예상 변경 파일
- `src/moai_adk/core/project/detector.py` (패턴 추가)
- `tests/unit/test_detector.py` (Tier 2 테스트)

---

### Phase 4: Tier 3 감지 구현 (파일 구조 분석)

**목표**: 확장자 기반 통계 분석 (fallback 메커니즘)

#### 구현 작업
1. **파일 확장자 통계**:
   ```python
   def _detect_from_structure(self, cwd: str) -> Optional[Tuple[str, float]]:
       """
       파일 구조 분석 기반 감지 (50점)
       """
       extension_counts = defaultdict(int)
       total_files = 0

       for ext, count in extension_counts.items():
           percentage = (count / total_files) * 100
           if percentage > 50:  # 50% 이상
               lang = EXTENSION_TO_LANG.get(ext)
               if lang:
                   return (lang, 50.0)

       return None
   ```

2. **확장자-언어 매핑**:
   ```python
   EXTENSION_TO_LANG = {
       ".py": "python",
       ".js": "javascript",
       ".ts": "javascript",
       ".go": "go",
       ".rs": "rust",
       ".java": "java",
       # ... 기타
   }
   ```

3. **디렉터리 탐색 최적화**:
   - `.git/`, `node_modules/`, `__pycache__/` 제외
   - 심볼릭 링크 무시
   - 최대 깊이 제한 (성능 최적화)

#### 검증 기준
- ✅ `.py` 파일 > 50% → python (50%)
- ✅ `.js/.ts` 파일 > 50% → javascript (50%)
- ✅ 제외 디렉터리 무시 확인
- ✅ 성능: 중형 프로젝트 < 500ms

#### 예상 변경 파일
- `src/moai_adk/core/project/detector.py` (메서드 추가)
- `tests/unit/test_detector.py` (Tier 3 테스트)

---

### Phase 5: implementation-planner 통합

**목표**: Step 0에서 Confidence 기반 자동 결정 로직 구현

#### 구현 작업
1. **Step 0 업데이트** (이미 완료됨):
   ```markdown
   ## Step 0: 언어 감지 (On-Demand)

   **Action**: implementation-planner가 호출될 때 다음을 수행:

   1. LanguageDetector.detect_with_confidence(cwd) 호출
   2. Confidence Score >= 80% → 자동 선택
   3. Confidence Score < 80% → AskUserQuestion 호출
   ```

2. **AskUserQuestion 통합**:
   ```python
   if confidence < 80:
       # 사용자에게 분석 결과 제시 + 선택 요청
       AskUserQuestion(
           questions=[{
               "question": f"감지된 언어: {language} (신뢰도: {confidence}%). 올바른가요?",
               "options": [
                   {"label": f"✅ {language} 사용", "value": language},
                   {"label": "🔄 다른 언어 선택", "value": "manual"},
               ]
           }]
       )
   ```

3. **로깅 및 디버깅**:
   - 감지 결과 및 신뢰도 로그 출력
   - 각 Tier 시도 내역 기록
   - 최종 선택 이유 명시

#### 검증 기준
- ✅ Confidence >= 80% 시 자동 진행 확인
- ✅ Confidence < 80% 시 AskUserQuestion 호출 확인
- ✅ 사용자 선택 결과 올바르게 전달
- ✅ 로그 출력 확인

#### 예상 변경 파일
- `.claude/agents/alfred/implementation-planner.md` (이미 업데이트됨)
- `tests/integration/test_planner_integration.py` (통합 테스트)

---

## 🧪 테스트 전략

### 단위 테스트 (tests/unit/test_detector.py)

#### 기존 테스트 (23개) - 유지
- 기본 언어 감지 기능
- 에러 처리
- 엣지 케이스

#### 신규 테스트 (최소 10개)
1. **Tier 1 테스트**:
   - `test_detect_from_spec_with_valid_yaml()` - SPEC 문서 존재 시 100% 반환
   - `test_detect_from_spec_with_invalid_yaml()` - 잘못된 YAML 무시
   - `test_detect_from_spec_missing()` - SPEC 없을 시 None 반환

2. **Tier 2 테스트**:
   - `test_detect_from_django_framework()` - Django 마커 감지 (90%)
   - `test_detect_from_fastapi_framework()` - FastAPI 마커 감지 (90%)
   - `test_detect_from_react_framework()` - React 마커 감지 (90%)
   - `test_detect_from_config_file()` - requirements.txt 감지 (80%)

3. **Tier 3 테스트**:
   - `test_detect_from_structure_python()` - .py 파일 > 50% (50%)
   - `test_detect_from_structure_javascript()` - .js 파일 > 50% (50%)

4. **통합 테스트**:
   - `test_detect_priority_order()` - Tier 1 > Tier 2 > Tier 3 우선순위
   - `test_detect_with_confidence_full_flow()` - 전체 플로우 검증

### 통합 테스트 (tests/integration/)
- implementation-planner Step 0 통합 검증
- AskUserQuestion 호출 시나리오
- 실제 프로젝트 구조 테스트

### 성능 테스트
- 중형 프로젝트 (1000+ 파일) 응답 시간 측정
- 메모리 사용량 모니터링

---

## 🏗️ 아키텍처 설계

### 클래스 구조
```
LanguageDetector
├── detect_language(cwd) → str              # 기존 함수
├── detect_with_confidence(cwd) → Tuple    # 신규 함수
├── _detect_from_spec(cwd) → Optional      # Tier 1
├── _detect_from_framework(cwd) → Optional # Tier 2
├── _detect_from_config(cwd) → Optional    # Tier 2
└── _detect_from_structure(cwd) → Optional # Tier 3
```

### 감지 플로우
```
detect_with_confidence(cwd)
    ↓
1. _detect_from_spec() → Tier 1 (100점)
    ↓ (None 반환 시)
2. _detect_from_framework() → Tier 2 (90점)
    ↓ (None 반환 시)
3. _detect_from_config() → Tier 2 (80점)
    ↓ (None 반환 시)
4. _detect_from_structure() → Tier 3 (50점)
    ↓ (None 반환 시)
5. 기본값 반환: ("python", 0.0)
```

### 데이터 흐름
```
implementation-planner Step 0
    ↓
LanguageDetector.detect_with_confidence(cwd)
    ↓
(language, confidence) → Tuple[str, float]
    ↓
IF confidence >= 80:
    자동 진행
ELSE:
    AskUserQuestion(사용자 확인)
```

---

## 🚀 마일스톤

### 우선순위별 목표

#### High Priority (핵심 기능)
1. ✅ Phase 1: LanguageDetector 확장 완료
2. ✅ Phase 2: Tier 1 감지 구현 완료
3. ✅ Phase 3: Tier 2 감지 구현 완료
4. ✅ 기존 23개 테스트 통과 유지
5. ✅ 신규 10개 테스트 작성 및 통과

#### Medium Priority (확장 기능)
1. ✅ Phase 4: Tier 3 감지 구현 완료
2. ✅ Phase 5: implementation-planner 통합 완료
3. ✅ 통합 테스트 작성 및 검증
4. ✅ 성능 최적화 (응답 시간 < 500ms)

#### Low Priority (추가 개선)
1. 디버그 모드 로깅 강화
2. 다중 언어 프로젝트 지원 (미래 버전)
3. 실시간 언어 변경 감지 (미래 버전)

---

## 🔧 기술 스택

### 개발 환경
- **Python**: 3.13+
- **의존성 관리**: uv
- **테스트 프레임워크**: pytest (최신 안정 버전)
- **타입 체킹**: mypy (최신 안정 버전)
- **코드 품질**: ruff (최신 안정 버전)

### 라이브러리
- `pathlib` - 파일 시스템 탐색
- `yaml` - YAML 파싱 (SPEC 문서)
- `collections.defaultdict` - 통계 계산

---

## ⚠️ 위험 요소 및 대응 방안

### 위험 요소
1. **성능 저하**: 대형 프로젝트에서 파일 탐색 시간 증가
   - **대응**: 디렉터리 제외 규칙 강화, 최대 깊이 제한

2. **YAML 파싱 오류**: 잘못된 SPEC 문서 형식
   - **대응**: try-except 블록 + 에러 로그 + 다음 Tier로 fallback

3. **호환성 문제**: 기존 시스템과의 충돌
   - **대응**: 기존 `detect_language()` 함수 유지, 철저한 회귀 테스트

4. **다중 언어 프로젝트**: JavaScript + Python 혼합
   - **대응**: 현재 버전에서는 주요 언어만 감지, 미래 버전에서 개선

---

## 📊 성공 기준

### 기능적 성공 기준
- ✅ Confidence Score >= 80% 시 자동 결정 작동
- ✅ Tier 1/2/3 우선순위 올바르게 적용
- ✅ 기존 23개 테스트 100% 통과
- ✅ 신규 10개 테스트 100% 통과

### 비기능적 성공 기준
- ✅ 테스트 커버리지 95% 이상
- ✅ 응답 시간 < 500ms (중형 프로젝트)
- ✅ 메모리 사용 < 50MB 추가
- ✅ 코드 리뷰 승인 완료

### 사용자 경험 기준
- ✅ 명시적 마커 존재 시 자동 결정 (사용자 확인 불필요)
- ✅ 저신뢰도 시에만 AskUserQuestion 호출
- ✅ 감지 결과 및 이유 명확히 로깅

---

## 📚 참고 자료

### 내부 문서
- `SPEC-LANG-DETECT-IMPROVED-001/spec.md` - 요구사항 명세
- `SPEC-LANG-DETECT-IMPROVED-001/acceptance.md` - 수용 기준
- `.moai/memory/language-config-schema.md` - 언어 설정 스키마

### 코드 참조
- `src/moai_adk/core/project/detector.py` - 현재 구현
- `tests/unit/test_detector.py` - 기존 테스트
- `.claude/agents/alfred/implementation-planner.md` - Step 0 로직

### 관련 이슈
- GitHub Issue #131: SessionStart Hook 성능 최적화

---

## 📝 구현 체크리스트

### Phase 1
- [ ] `detect_with_confidence()` 메서드 추가
- [ ] CONFIDENCE_LEVELS 딕셔너리 정의
- [ ] 기존 `detect_language()` 리팩토링
- [ ] 타입 힌트 및 docstring 작성

### Phase 2
- [ ] `_detect_from_spec()` 메서드 구현
- [ ] YAML frontmatter 파싱 로직
- [ ] SPEC 디렉터리 탐색
- [ ] Tier 1 단위 테스트 3개 작성

### Phase 3
- [ ] FRAMEWORK_PATTERNS 확장
- [ ] CONFIG_FILE_PATTERNS 정의
- [ ] `_detect_from_framework()` 구현
- [ ] `_detect_from_config()` 구현
- [ ] Tier 2 단위 테스트 4개 작성

### Phase 4
- [ ] `_detect_from_structure()` 구현
- [ ] EXTENSION_TO_LANG 매핑
- [ ] 디렉터리 탐색 최적화
- [ ] Tier 3 단위 테스트 2개 작성

### Phase 5
- [ ] implementation-planner Step 0 검증
- [ ] AskUserQuestion 통합 테스트
- [ ] 로깅 및 디버깅 코드 추가
- [ ] 통합 테스트 1개 작성

### 최종 검증
- [ ] 기존 23개 테스트 통과
- [ ] 신규 10개 테스트 통과
- [ ] 테스트 커버리지 95% 확인
- [ ] 성능 테스트 통과 (< 500ms)
- [ ] 코드 리뷰 승인
- [ ] 문서 업데이트 완료

---

**문서 상태**: Draft
**최종 업데이트**: 2025-10-31
**작성자**: @goos
