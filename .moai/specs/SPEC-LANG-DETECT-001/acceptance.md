# Acceptance Criteria: LANG-DETECT-001

## 인수 기준 개요

이 문서는 SPEC-LANG-DETECT-001(PHP/Laravel 언어 감지 개선)의 인수 기준을 정의합니다.

---

## AC-1: Laravel 프로젝트 기본 감지

### Given (전제 조건)
- 임시 디렉토리에 Laravel 프로젝트 구조 생성
- `artisan` 파일 존재 (Laravel CLI 도구)
- `composer.json` 파일 존재

### When (실행)
- `LanguageDetector().detect(project_path)` 호출

### Then (기대 결과)
- 반환값이 `"php"`이어야 함
- Python이 아닌 PHP로 정확히 인식
- 감지 시간이 100ms 이하

### 검증 방법
```python
def test_detect_laravel_from_artisan_file(tmp_project_dir):
    """Should detect Laravel project as PHP from artisan file"""
    # Given
    (tmp_project_dir / "artisan").write_text("#!/usr/bin/env php")
    (tmp_project_dir / "composer.json").write_text('{"require": {"laravel/framework": "^11.0"}}')

    # When
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then
    assert result == "php"
    assert result != "python"
```

---

## AC-2: Laravel 디렉토리 구조 감지

### Given (전제 조건)
- Laravel 특화 디렉토리 존재:
  - `app/` 디렉토리 (애플리케이션 로직)
  - `bootstrap/laravel.php` 파일 (부트스트랩)
  - `composer.json` 파일

### When (실행)
- `LanguageDetector().detect(project_path)` 호출

### Then (기대 결과)
- 반환값이 `"php"`이어야 함
- Laravel 디렉토리 구조를 정확히 인식

### 검증 방법
```python
def test_detect_laravel_from_directory_structure(tmp_project_dir):
    """Should detect Laravel project from directory structure"""
    # Given
    (tmp_project_dir / "app").mkdir()
    (tmp_project_dir / "bootstrap").mkdir()
    (tmp_project_dir / "bootstrap" / "laravel.php").write_text("<?php")
    (tmp_project_dir / "composer.json").write_text('{"name": "laravel/laravel"}')

    # When
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then
    assert result == "php"
```

---

## AC-3: 혼합 프로젝트에서 PHP 우선 인식

### Given (전제 조건)
- Python 파일 존재: `deploy.py` (배포 스크립트)
- PHP 파일 존재: `index.php`
- Laravel 식별자 존재: `artisan`, `composer.json`

### When (실행)
- `LanguageDetector().detect(project_path)` 호출

### Then (기대 결과)
- 반환값이 `"php"`이어야 함 (Python이 아님)
- `artisan` 또는 `composer.json` 존재 시 PHP를 우선 반환
- `detect_multiple()` 호출 시 `["php", "python"]` 순서로 반환

### 검증 방법
```python
def test_detect_php_over_python_in_mixed_project(tmp_project_dir):
    """Should prioritize PHP when both Python and PHP files exist"""
    # Given
    (tmp_project_dir / "deploy.py").write_text("import os")
    (tmp_project_dir / "index.php").write_text("<?php echo 'Hello';")
    (tmp_project_dir / "artisan").write_text("#!/usr/bin/env php")
    (tmp_project_dir / "composer.json").write_text('{}')

    # When
    detector = LanguageDetector()
    result = detector.detect(tmp_project_dir)

    # Then
    assert result == "php"

    # Bonus: detect_multiple should return PHP first
    multiple = detector.detect_multiple(tmp_project_dir)
    assert multiple[0] == "php"
    assert "python" in multiple
```

---

## AC-4: composer.json Laravel 의존성 확인

### Given (전제 조건)
- `composer.json` 파일에 `laravel/framework` 의존성 포함
- PHP 파일 존재 (`*.php`)

### When (실행)
- `LanguageDetector().detect(project_path)` 호출

### Then (기대 결과)
- 반환값이 `"php"`이어야 함
- Laravel 프레임워크 의존성을 정확히 식별

### 검증 방법
```python
def test_detect_php_from_composer_laravel_dependency(tmp_project_dir):
    """Should detect PHP from composer.json with laravel/framework"""
    # Given
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

---

## AC-5: 기존 테스트 회귀 방지

### Given (전제 조건)
- 기존 모든 언어 감지 테스트 (`test_detector.py`)
- Python, TypeScript, Java, Go, Rust 등 다른 언어 테스트

### When (실행)
- 전체 테스트 스위트 실행: `pytest tests/unit/test_detector.py`

### Then (기대 결과)
- 모든 기존 테스트 통과 (회귀 없음)
- 새 테스트 4개 추가 통과
- 테스트 커버리지 85% 이상 유지

### 검증 방법
```bash
# 전체 테스트 실행
pytest tests/unit/test_detector.py -v

# 커버리지 측정
pytest tests/unit/test_detector.py --cov=src/moai_adk/core/project/detector --cov-report=term-missing

# 기대 결과:
# - 모든 테스트 PASSED
# - Coverage >= 85%
```

---

## 비기능 요구사항 검증

### 성능
- **기준**: 언어 감지 시간 <100ms
- **검증**: `pytest --durations=10` 실행 후 `test_detect_*` 함수 소요 시간 확인

### 크로스 플랫폼 호환성
- **기준**: Windows, macOS, Linux 모두 동작
- **검증**: `pathlib.Path` 사용으로 경로 구분자 자동 처리 확인

### 타입 안전성
- **기준**: Python 3.13+ 타입 힌트 사용
- **검증**: `mypy src/moai_adk/core/project/detector.py --strict` (향후)

---

## 체크리스트

- [ ] AC-1: Laravel 프로젝트 기본 감지 테스트 통과
- [ ] AC-2: Laravel 디렉토리 구조 감지 테스트 통과
- [ ] AC-3: 혼합 프로젝트에서 PHP 우선 인식 테스트 통과
- [ ] AC-4: composer.json Laravel 의존성 확인 테스트 통과
- [ ] AC-5: 기존 테스트 회귀 없음
- [ ] 성능: 언어 감지 <100ms
- [ ] 커버리지: >=85%
- [ ] 크로스 플랫폼: Windows/macOS/Linux 동작 확인

---

_이 인수 기준은 `/alfred:2-build SPEC-LANG-DETECT-001` TDD 구현 시 검증됩니다._
