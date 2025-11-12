---
id: LANG-DETECT-001
version: 0.1.0
status: completed
created: 2025-10-19
updated: 2025-10-19
priority: high
category: bugfix
labels:
  - language-detection
  - php
  - laravel
related_issue: "https://github.com/modu-ai/moai-adk/issues/36"
scope:
  packages:
    - src/moai_adk/core/project
  files:
    - detector.py
    - test_detector.py
---


## HISTORY

### v0.1.0 (2025-10-19)
- **TDD COMPLETED**: RED-GREEN-REFACTOR 구현 완료
- **TESTS**: 4개 Laravel 감지 테스트 추가 (24/24 통과)
- **COVERAGE**: detector.py 88.89% → 92.59% 향상
- **COMMITS**:
  - 7c69624: 🐛 FIX: Laravel 프로젝트 PHP 언어 감지 개선
  - b6b82a9: 📝 SPEC: Laravel PHP 언어 감지 개선 명세 작성
- **CHANGED**: LANGUAGE_PATTERNS에 Laravel 특화 파일 3개 추가 (artisan, app/, bootstrap/laravel.php)
- **REVIEW**: GitHub Issue #36 답변 완료

### v0.0.1 (2025-10-19)
- **INITIAL**: Laravel 프로젝트 PHP 언어 감지 오류 수정 명세 작성
- **SCOPE**: LanguageDetector 클래스 PHP 패턴 확장 및 우선순위 조정
- **CONTEXT**: GitHub Issue #36 - Laravel 프로젝트가 Python으로 잘못 인식되는 문제 해결
- **RELATED**: https://github.com/modu-ai/moai-adk/issues/36

---

## Environment (환경)

### 시스템 환경
- **Python 버전**: 3.13+
- **운영체제**: Windows, macOS, Linux (크로스 플랫폼)
- **의존성**: pathlib (표준 라이브러리)

### 전제 조건
- Laravel 프로젝트는 다음 파일/디렉토리를 포함:
  - `artisan` (PHP CLI 스크립트)
  - `composer.json` (PHP 의존성 관리)
  - `app/`, `bootstrap/`, `config/` 디렉토리 (Laravel 구조)
- Python 파일(`.py`)이 프로젝트에 혼재할 수 있음 (예: 배포 스크립트, 테스트 도구)

---

## Assumptions (가정)

### 언어 감지 전략
1. **프레임워크 우선**: 특정 프레임워크 파일이 발견되면 해당 언어를 우선 반환
2. **복합 프로젝트 처리**: 여러 언어가 혼재된 경우 주 언어 (프레임워크) 우선
3. **파일 존재 검증**: `rglob()` 패턴 매칭으로 파일 존재 확인

### Laravel 식별 기준
- **artisan 파일**: Laravel의 고유 CLI 도구
- **composer.json 의존성**: `laravel/framework` 패키지 포함
- **디렉토리 구조**: `app/`, `bootstrap/laravel.php` 등 Laravel 특화 구조

---

## Requirements (요구사항)

### Ubiquitous Requirements (필수 기능)

- 시스템은 Laravel 프로젝트를 PHP로 정확히 감지해야 한다
- 시스템은 PHP 패턴에 Laravel 특화 파일을 포함해야 한다
- 시스템은 `artisan` 파일을 PHP 프로젝트의 강력한 지표로 인식해야 한다
- 시스템은 `composer.json`을 PHP 프로젝트의 표준 파일로 인식해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN `artisan` 파일이 프로젝트 루트에 존재하면, 시스템은 즉시 PHP로 인식해야 한다
- WHEN `composer.json`에 `"laravel/framework"` 의존성이 포함되면, 시스템은 PHP로 인식해야 한다
- WHEN `app/`, `bootstrap/laravel.php` 디렉토리가 발견되면, 시스템은 Laravel 프로젝트로 판단해야 한다
- WHEN Python 파일과 PHP 파일이 혼재되고 `composer.json` 또는 `artisan`이 존재하면, 시스템은 PHP를 우선 반환해야 한다

### State-driven Requirements (상태 기반)

- WHILE 여러 언어가 감지되는 혼합 프로젝트일 때, 시스템은 프레임워크 특화 파일을 우선해야 한다
- WHILE `LANGUAGE_PATTERNS` 딕셔너리를 순회할 때, 시스템은 PHP를 Python보다 먼저 검사해야 한다
- WHILE `detect()` 메서드 실행 중일 때, 시스템은 첫 번째 매칭 언어를 즉시 반환해야 한다

### Optional Features (선택적 기능)

- WHERE `composer.json` 파일이 존재하면, 시스템은 파일 내용을 파싱하여 Laravel 프레임워크 여부를 확인할 수 있다
- WHERE Laravel 버전 정보가 필요하면, 시스템은 `composer.json`에서 버전을 추출할 수 있다

### Constraints (제약사항)

- IF `composer.json`이 존재하면, 시스템은 Python보다 PHP를 우선해야 한다
- IF `artisan` 파일이 존재하면, 시스템은 무조건 PHP로 인식해야 한다
- 언어 우선순위는 다음 순서를 따라야 한다: 프레임워크 특화 파일 > 언어별 파일 확장자
- PHP 패턴은 최소 4개 이상의 식별자를 포함해야 한다: `*.php`, `composer.json`, `artisan`, `app/`

---

## Technical Design (기술 설계)

### 현재 구현 (문제점)

```python
# src/moai_adk/core/project/detector.py:14-34
LANGUAGE_PATTERNS = {
    "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
    "typescript": ["*.ts", "tsconfig.json"],
    "javascript": ["*.js", "package.json"],
    "java": ["*.java", "pom.xml", "build.gradle"],
    "go": ["*.go", "go.mod"],
    "rust": ["*.rs", "Cargo.toml"],
    # ...
    "php": ["*.php", "composer.json"],  # ❌ Laravel 파일 누락
    # ...
}
```

**문제점**:
1. PHP 패턴에 Laravel 특화 파일 누락 (`artisan`, `app/`, `bootstrap/laravel.php`)
2. Python이 PHP보다 먼저 검사되어 `.py` 파일이 있으면 Python으로 우선 인식
3. Laravel 프로젝트에는 배포 스크립트 등 `.py` 파일이 존재할 수 있음

### 개선된 구현 (제안)

```python
LANGUAGE_PATTERNS = {
    # 프레임워크 언어를 상위로 이동
    "php": [
        "*.php",
        "composer.json",
        "artisan",           # ✅ Laravel CLI 도구
        "app/",              # ✅ Laravel 디렉토리
        "bootstrap/laravel.php"  # ✅ Laravel 부트스트랩
    ],
    "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
    # ...
}
```

**개선점**:
1. PHP 패턴 확장: Laravel 특화 파일 3개 추가
2. 우선순위 조정: PHP를 Python보다 상위 배치 (딕셔너리 순서)
3. 정확도 향상: `artisan` 파일 존재 시 Laravel 확실히 인식

---


- **TEST**: `tests/unit/test_detector.py` (Laravel 감지 테스트 추가)
- **CODE**: `src/moai_adk/core/project/detector.py` (LANGUAGE_PATTERNS 수정)
- **DOC**: `README.md` (언어 지원 목록 업데이트 선택)
- **RELATED ISSUE**: https://github.com/modu-ai/moai-adk/issues/36

---

## Success Criteria (성공 기준)

### 기능 검증
1. Laravel 프로젝트 (`artisan` 포함)를 PHP로 정확히 감지
2. `composer.json`에 `laravel/framework` 의존성 포함 시 PHP 인식
3. Python 파일이 혼재된 Laravel 프로젝트도 PHP로 우선 인식

### 품질 기준
- 테스트 커버리지: 85% 이상 유지
- 모든 새 테스트 통과
- 기존 테스트 회귀 없음
- Windows/macOS/Linux 크로스 플랫폼 동작

### 비기능 요구사항
- 성능: 언어 감지 시간 <100ms
- 호환성: Python 3.13+ 타입 힌트 사용
- 유지보수성: 주석으로 Laravel 특화 파일 설명 추가

---

## References

- **GitHub Issue**: https://github.com/modu-ai/moai-adk/issues/36
- **Laravel 공식 문서**: https://laravel.com/docs
- **Composer 공식 문서**: https://getcomposer.org/doc/
- **MoAI-ADK Development Guide**: `.moai/memory/development-guide.md`

---

_이 SPEC은 `/alfred:2-run SPEC-LANG-DETECT-001`로 TDD 구현됩니다._
