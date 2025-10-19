# Implementation Plan: LANG-DETECT-001

## 구현 계획 개요

이 문서는 SPEC-LANG-DETECT-001(PHP/Laravel 언어 감지 개선)의 TDD 구현 계획을 정의합니다.

---

## TDD 구현 전략

### RED → GREEN → REFACTOR

1. **RED**: Laravel 감지 실패하는 테스트 작성 및 확인
2. **GREEN**: 최소한의 코드로 테스트 통과시키기
3. **REFACTOR**: 코드 품질 개선 및 주석 추가

---

## Phase 1: RED - 실패하는 테스트 작성

### 작업 목록

#### 1.1. Laravel artisan 파일 감지 테스트
```python
# tests/unit/test_detector.py

def test_detect_laravel_from_artisan_file(tmp_project_dir: Path):
    """Should detect Laravel project as PHP from artisan file"""
    # Given: Laravel artisan file
    (tmp_project_dir / "artisan").write_text("#!/usr/bin/env php")
    (tmp_project_dir / "composer.json").write_text('{"require": {"laravel/framework": "^11.0"}}')

    # When: detect language
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then: should return "php", not "python"
    assert result == "php"
```

**예상 실패 이유**: 현재 PHP 패턴에 `artisan` 없음 → 감지 실패 또는 Python 우선 인식

---

#### 1.2. Laravel 디렉토리 구조 감지 테스트
```python
def test_detect_laravel_from_directory_structure(tmp_project_dir: Path):
    """Should detect Laravel from app/ and bootstrap/ directories"""
    # Given: Laravel directory structure
    (tmp_project_dir / "app").mkdir()
    (tmp_project_dir / "bootstrap").mkdir()
    (tmp_project_dir / "bootstrap" / "laravel.php").write_text("<?php")
    (tmp_project_dir / "composer.json").write_text('{}')

    # When
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then
    assert result == "php"
```

**예상 실패 이유**: PHP 패턴에 `app/`, `bootstrap/laravel.php` 없음

---

#### 1.3. 혼합 프로젝트에서 PHP 우선 인식 테스트
```python
def test_detect_php_over_python_in_mixed_project(tmp_project_dir: Path):
    """Should prioritize PHP when both Python and PHP exist"""
    # Given: Mixed Python + PHP project with Laravel markers
    (tmp_project_dir / "deploy.py").write_text("import os")
    (tmp_project_dir / "index.php").write_text("<?php")
    (tmp_project_dir / "artisan").write_text("#!/usr/bin/env php")
    (tmp_project_dir / "composer.json").write_text('{}')

    # When
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then: PHP should be detected first
    assert result == "php"

    # Bonus: check multiple languages
    multiple = detector.detect_multiple(tmp_project_dir)
    assert multiple[0] == "php"
```

**예상 실패 이유**: Python이 PHP보다 먼저 검사되어 `"python"` 반환

---

#### 1.4. composer.json Laravel 의존성 확인 테스트
```python
def test_detect_php_from_composer_laravel_dependency(tmp_project_dir: Path):
    """Should detect PHP from composer.json with laravel/framework"""
    # Given
    import json
    composer_content = {
        "require": {
            "php": "^8.2",
            "laravel/framework": "^11.0"
        }
    }
    (tmp_project_dir / "composer.json").write_text(json.dumps(composer_content))
    (tmp_project_dir / "index.php").write_text("<?php")

    # When
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then
    assert result == "php"
```

**예상 실패 이유**: 현재는 `composer.json` 파일 존재만 확인, 내용 파싱 안 함

---

### 1.5. 테스트 실행 및 실패 확인

```bash
# RED 단계: 모든 테스트가 실패해야 함
pytest tests/unit/test_detector.py::test_detect_laravel_from_artisan_file -v
pytest tests/unit/test_detector.py::test_detect_laravel_from_directory_structure -v
pytest tests/unit/test_detector.py::test_detect_php_over_python_in_mixed_project -v
pytest tests/unit/test_detector.py::test_detect_php_from_composer_laravel_dependency -v

# 예상 결과: 4 FAILED
```

**커밋 메시지** (Locale: ko):
```
🔴 RED: Laravel 언어 감지 테스트 추가

@TAG:LANG-DETECT-001-RED
- test_detect_laravel_from_artisan_file
- test_detect_laravel_from_directory_structure
- test_detect_php_over_python_in_mixed_project
- test_detect_php_from_composer_laravel_dependency
```

---

## Phase 2: GREEN - 테스트 통과시키기

### 작업 목록

#### 2.1. LANGUAGE_PATTERNS 수정 (detector.py)

**현재 코드** (`src/moai_adk/core/project/detector.py:13-34`):
```python
LANGUAGE_PATTERNS = {
    "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
    # ...
    "php": ["*.php", "composer.json"],
    # ...
}
```

**개선된 코드**:
```python
LANGUAGE_PATTERNS = {
    # PHP를 Python보다 먼저 검사 (우선순위 조정)
    "php": [
        "*.php",
        "composer.json",
        "artisan",                # Laravel CLI tool
        "app/",                   # Laravel app directory
        "bootstrap/laravel.php"   # Laravel bootstrap file
    ],
    "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
    "typescript": ["*.ts", "tsconfig.json"],
    # ... (나머지 동일)
}
```

**변경 사항**:
1. PHP를 딕셔너리 최상단으로 이동 (우선순위 상승)
2. Laravel 특화 파일 3개 추가: `artisan`, `app/`, `bootstrap/laravel.php`
3. 주석 추가: Laravel 식별자 설명

---

#### 2.2. 테스트 재실행 및 통과 확인

```bash
# GREEN 단계: 모든 테스트가 통과해야 함
pytest tests/unit/test_detector.py::test_detect_laravel_from_artisan_file -v
pytest tests/unit/test_detector.py::test_detect_laravel_from_directory_structure -v
pytest tests/unit/test_detector.py::test_detect_php_over_python_in_mixed_project -v
pytest tests/unit/test_detector.py::test_detect_php_from_composer_laravel_dependency -v

# 예상 결과: 4 PASSED
```

**커밋 메시지** (Locale: ko):
```
🟢 GREEN: PHP/Laravel 언어 감지 로직 구현

@TAG:LANG-DETECT-001-GREEN
- LANGUAGE_PATTERNS에 Laravel 특화 파일 추가
- artisan, app/, bootstrap/laravel.php 패턴 추가
- PHP 우선순위 상승 (Python보다 먼저 검사)
```

---

## Phase 3: REFACTOR - 코드 품질 개선

### 작업 목록

#### 3.1. 주석 및 문서화 개선

```python
class LanguageDetector:
    """Automatically detect up to 20 programming languages.

    Prioritizes framework-specific files (e.g., Laravel, Django) over
    generic language files to improve accuracy in mixed-language projects.
    """

    LANGUAGE_PATTERNS = {
        "php": [
            "*.php",
            "composer.json",
            "artisan",                # Laravel: CLI tool (unique identifier)
            "app/",                   # Laravel: application directory
            "bootstrap/laravel.php"   # Laravel: bootstrap file
        ],
        # ... (주석 추가)
    }
```

**개선 사항**:
- 클래스 docstring에 우선순위 전략 설명 추가
- Laravel 패턴에 인라인 주석 추가 (각 파일의 역할 명시)

---

#### 3.2. 타입 힌트 강화 (선택사항)

```python
from pathlib import Path
from typing import Optional, List

def detect(self, path: str | Path = ".") -> Optional[str]:
    """Detect a single language (in priority order).

    Args:
        path: Directory to inspect (default: current directory).

    Returns:
        Detected language name (lowercase) or None if no match.

    Priority:
        Framework-specific files > Generic language files
    """
    # ... (기존 코드)
```

---

#### 3.3. 기존 테스트 회귀 확인

```bash
# 전체 테스트 스위트 실행
pytest tests/unit/test_detector.py -v

# 커버리지 측정
pytest tests/unit/test_detector.py --cov=src/moai_adk/core/project/detector --cov-report=term-missing

# 기대 결과:
# - 모든 테스트 PASSED
# - Coverage >= 85%
```

---

#### 3.4. 성능 확인

```bash
# 테스트 실행 시간 측정
pytest tests/unit/test_detector.py --durations=10

# 기대 결과:
# - test_detect_laravel_* 함수들이 각각 <100ms
```

---

**커밋 메시지** (Locale: ko):
```
♻️ REFACTOR: 언어 감지 로직 주석 및 문서화 개선

@TAG:LANG-DETECT-001-REFACTOR
- LanguageDetector 클래스 docstring 업데이트
- Laravel 패턴에 인라인 주석 추가
- 우선순위 전략 문서화
```

---

## 예상 산출물

### 코드 변경
- **수정**: `src/moai_adk/core/project/detector.py`
  - LANGUAGE_PATTERNS 딕셔너리: PHP 항목 확장 및 순서 조정
  - 주석 추가 (Laravel 특화 파일 설명)
- **추가**: `tests/unit/test_detector.py`
  - 4개 테스트 함수 추가 (Laravel 감지)

### 커밋 이력
1. **RED**: 실패하는 테스트 4개 추가
2. **GREEN**: PHP 패턴 확장 및 우선순위 조정
3. **REFACTOR**: 주석 및 문서화 개선

---

## 검증 체크리스트

- [ ] RED: 4개 테스트 작성 및 실패 확인
- [ ] GREEN: `LANGUAGE_PATTERNS` 수정 및 테스트 통과
- [ ] REFACTOR: 주석 및 docstring 개선
- [ ] 기존 테스트 회귀 없음 (전체 테스트 통과)
- [ ] 커버리지 85% 이상 유지
- [ ] 성능: 언어 감지 <100ms
- [ ] 크로스 플랫폼 동작 확인 (pathlib 사용)

---

## 다음 단계

1. `/alfred:2-build SPEC-LANG-DETECT-001` 실행
2. TDD 사이클 진행 (RED → GREEN → REFACTOR)
3. `/alfred:3-sync` 실행하여 문서 동기화 및 TAG 검증
4. GitHub Issue #36 종료

---

_이 계획은 `/alfred:2-build SPEC-LANG-DETECT-001` 실행 시 참조됩니다._
