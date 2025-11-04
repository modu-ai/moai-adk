---
id: UPDATE-CACHE-FIX-001
version: 0.1.0
status: closed
created: 2025-10-30
updated: 2025-10-30
author: "@goos"
priority: high
---

# @SPEC:UPDATE-CACHE-FIX-001: UV Tool Upgrade Cache Refresh Auto-Retry Implementation

## HISTORY

### v0.1.0 (2025-10-30)
- **IMPLEMENTATION COMPLETED**: TDD implementation finished (status: draft → completed)
- **AUTHOR**: @goos
- **SUMMARY**:
  - Implemented 3 cache detection/refresh/retry functions
  - Created 8 comprehensive test cases (100% passing)
  - All acceptance criteria verified
  - Full TAG chain: `@SPEC`→`@TEST`→`@CODE`→`@DOC`
  - Test coverage: 100% (target: 90%+)
  - Code quality: ruff+mypy all passing

- **COMPONENTS**:
  1. `@CODE:UPDATE-CACHE-FIX-001-001`: _detect_stale_cache() function
  2. `@CODE:UPDATE-CACHE-FIX-001-002`: _clear_uv_package_cache() function
  3. `@CODE:UPDATE-CACHE-FIX-001-003`: _execute_upgrade_with_retry() integration

- **TEST RESULTS**:
  - `@TEST:UPDATE-CACHE-FIX-001`: 8/8 tests passing
  - Coverage: 100% of implementation code
  - Execution time: 0.90 seconds

- **DELIVERABLES**:
  - spec.md: Complete SPEC with EARS structure
  - plan.md: 5-phase implementation plan
  - acceptance.md: Test scenarios and verification
  - Code: 3 functions + 8 tests in production quality
  - Documentation: README + CHANGELOG updates

- **COMMITS**:
  - RED phase: Initial test framework
  - GREEN phase: Minimal implementation
  - REFACTOR phase: Code quality improvements

### v0.0.1 (2025-10-30)
- **INITIAL**: UV tool upgrade 캐시 갱신 자동 재시도 구현 사양 초기 작성
- **AUTHOR**: @goos
- **SCOPE**: uv tool upgrade 명령의 자동 캐시 감지 및 재시도 로직
- **CONTEXT**: 첫 번째 업그레이드 확인이 오래된 PyPI 메타데이터 캐시로 인해 실패하는 이중 실행 문제 해결
- **ROOT CAUSE**: 캐시 갱신이 버전 확인 이전이 아닌 이후에 발생함

## Environment

**시스템 요구사항**:
- Python: 3.13+
- uv: 0.9.3+
- 운영체제: macOS/Linux/Windows (모두 지원)
- MoAI-ADK: 0.9.0+

**도구**:
- subprocess (Python 표준 라이브러리)
- packaging.version (기존 의존성)
- uv CLI (외부 의존성)

**환경 변수**:
- `UV_CACHE_DIR`: uv 캐시 디렉토리 (기본값: `~/.cache/uv/`)
- `MOAI_ADK_UPDATE_NO_RETRY`: 자동 재시도 비활성화 플래그 (향후 구현 예정)

## Assumptions

1. **uv tool cache 구조**: `~/.cache/uv/simple-v18/pypi/` 디렉토리에 .rkyv 메타데이터 파일이 포함되어 있음
2. **uv cache clean 명령어**: uv 0.9.3+ 버전에서 `uv cache clean <package>` 문법으로 패키지별 캐시 정리 지원
3. **PyPI API 가용성**: PyPI API에 접근 가능하며 5초 타임아웃 내에 응답함
4. **사용자 네트워크**: PyPI 메타데이터 갱신을 위한 합리적인 네트워크 연결 상태
5. **Subprocess 가용성**: 대상 플랫폼에서 Python subprocess 모듈이 정상 작동함

## Requirements

### Ubiquitous (보편적 요구사항)
- 시스템은 `moai-adk update` 명령 실행 시 **1회 실행만으로 최신 버전 설치를 완료**해야 한다
- 사용자는 실제로 새 버전이 존재하는데도 "업그레이드할 것이 없습니다"라는 거짓 메시지를 받아서는 안 된다
- 자동 재시도는 사용자에게 투명하게 진행되어야 하며, 명확한 피드백을 제공해야 한다

### Event-driven (이벤트 기반 요구사항)
- **WHEN** `uv tool upgrade moai-adk` 실행 후 "Nothing to upgrade" 출력이 감지되면
  - **AND** PyPI에 실제로 더 새로운 버전이 존재하는 경우
  - **THEN** 시스템은 자동으로 `uv cache clean moai-adk` 실행 후 업그레이드를 재시도해야 한다

- **WHEN** 사용자가 설치된 버전과 최신 버전이 동일함을 확인하면
  - **THEN** 시스템은 재시도를 수행하지 않고 정상 종료해야 한다

- **WHEN** 캐시 정리 명령이 실패하면
  - **THEN** 시스템은 무음으로 실패하고 수동 해결 방법을 사용자에게 제시해야 한다

### State-driven (상태 기반 요구사항)
- **WHILE** 업그레이드 재시도가 진행 중일 때
  - **THEN** 시스템은 사용자에게 진행 상황을 명확히 표시해야 한다 (예: "⚠️ Cache outdated, refreshing...")
  - **AND** 캐시 정리 진행 상황을 표시해야 한다 (예: "♻️ Cache cleared, retrying upgrade...")

- **WHILE** 재시도가 진행 중이고 실패하면
  - **THEN** 시스템은 명확한 에러 메시지와 함께 수동 해결 방법을 제시해야 한다
  - **AND** 무한 루프를 방지하기 위해 최대 1회만 재시도해야 한다

### Optional (선택적 기능)
- **WHERE** 사용자가 향후 `--no-retry` 플래그를 제공할 경우 (v0.9.2+ 예정)
  - **THEN** 시스템은 자동 재시도를 건너뛰고 원래 메시지만 표시할 수 있다

- **WHERE** 사용자가 환경 변수 `MOAI_ADK_UPDATE_NO_RETRY=1`을 설정한 경우
  - **THEN** 시스템은 자동 재시도를 비활성화하고 원래 동작을 유지할 수 있다

### Unwanted Behaviors (원치 않는 동작)
- **IF** 재시도 로직이 2회 이상을 시도하는 경우
  - **THEN** 시스템은 무한 루프를 방지하고 원래 에러를 보고해야 한다
  - **AND** 수동 캐시 정리 명령어를 안내해야 한다 (`uv cache clean moai-adk && moai-adk update`)

- **IF** `uv cache clean` 명령어가 실패하는 경우
  - **THEN** 시스템은 graceful하게 오류를 처리하고 사용자에게 수동 명령어를 제시해야 한다
  - **AND** 무한 재시도를 방지해야 한다

- **IF** 다른 패키지의 업그레이드를 시도하는 경우
  - **THEN** 이 로직은 moai-adk 패키지에만 적용되어야 한다
  - **AND** 다른 패키지의 업그레이드 흐름에 영향을 주지 않아야 한다

- **IF** 버전 파싱이 실패하는 경우
  - **THEN** 시스템은 오류를 조용히 무시하고 재시도를 건너뛰어야 한다 (graceful degradation)

## Specifications

### 핵심 컴포넌트

#### 1. 캐시 스테일 감지 함수

**함수 시그니처**:
```python
def _detect_stale_cache(
    upgrade_output: str,
    current_version: str,
    latest_version: str
) -> bool:
    """
    Detect if uv cache is stale by comparing versions.

    @CODE:UPDATE-CACHE-FIX-001-001
    """
```

**로직**:
1. `upgrade_output`에서 "Nothing to upgrade" 문자열 확인
2. `current_version < latest_version` 비교 (packaging.version.parse 사용)
3. 두 조건 모두 True이면 캐시 스테일로 판단

**에러 처리**:
- 버전 파싱 실패 → False 반환
- 출력 문자열이 None → False 반환

#### 2. 캐시 정리 함수

**함수 시그니처**:
```python
def _clear_uv_package_cache(package_name: str = "moai-adk") -> bool:
    """
    Clear uv cache for specific package.

    @CODE:UPDATE-CACHE-FIX-001-002
    """
```

**로직**:
1. `subprocess.run(["uv", "cache", "clean", package_name])` 실행
2. 타임아웃: 10초 (느린 네트워크 대비)
3. returncode == 0 이면 True, 그 외 False

**로깅**:
- 성공: DEBUG 레벨 (`logger.debug(f"UV cache cleared for {package_name}")`)
- 실패: WARNING 레벨 (`logger.warning(f"Failed to clear UV cache: {stderr}")`)

#### 3. 업그레이드 재시도 로직

**함수 수정**: `_execute_upgrade()` 함수 확장 또는 새 래퍼 함수 생성

**로직 흐름**:
```python
# 1단계: 첫 번째 업그레이드 시도
result = subprocess.run(installer_cmd, ...)

# 2단계: 캐시 스테일 감지
if result.returncode == 0 and "Nothing to upgrade" in result.stdout:
    latest_version = _get_latest_version()
    current_version = _get_current_version()

    if _detect_stale_cache(result.stdout, current_version, latest_version):
        # 3단계: 캐시 정리 + 재시도
        console.print("[yellow]⚠️ Cache outdated, refreshing...[/yellow]")

        if _clear_uv_package_cache("moai-adk"):
            console.print("[cyan]♻️ Cache cleared, retrying upgrade...[/cyan]")
            result = subprocess.run(installer_cmd, ...)
        else:
            console.print("[red]✗ Cache clear failed. Manual workaround:[/red]")
            console.print("  uv cache clean moai-adk && moai-adk update")
            return False

# 4단계: 최종 결과 반환
return result.returncode == 0
```

**@TAG 참조**: @CODE:UPDATE-CACHE-FIX-001-003

### 에러 처리 전략

| 에러 시나리오 | 대응 전략 | 사용자 메시지 |
|-------------|---------|-------------|
| 버전 파싱 실패 | 재시도 건너뛰기 | (없음 - 무음 처리) |
| 캐시 정리 실패 | 수동 명령 안내 | "✗ Cache clear failed. Try: uv cache clean moai-adk" |
| 재시도 후 실패 | 원래 에러 표시 | 원래 업그레이드 에러 메시지 |
| 무한 루프 위험 | max_retries=1 제한 | "✗ Upgrade failed after retry" |

### 로깅 전략

**로그 레벨**:
- DEBUG: 캐시 정리 성공, 버전 비교 결과
- INFO: 재시도 시작, 업그레이드 성공
- WARNING: 캐시 정리 실패, 버전 파싱 실패
- ERROR: 업그레이드 영구 실패

**로그 예시**:
```python
logger.debug("Detected stale cache: current=0.8.3, latest=0.9.1")
logger.info("Retrying upgrade after cache refresh")
logger.warning("Cache clear failed: permission denied")
logger.error("Upgrade failed after retry: network timeout")
```

## Traceability (@TAG)

### SPEC
- **Primary**: @SPEC:UPDATE-CACHE-FIX-001
- **Related**: @SPEC:UPDATE-REFACTOR-001 (기존 update 리팩토링)

### TEST
- **Primary**: @TEST:UPDATE-CACHE-FIX-001
- **Unit Tests**:
  - `tests/unit/test_update_uv_cache_fix.py::test_detect_stale_cache_true` (@TEST:UPDATE-CACHE-FIX-001-001)
  - `tests/unit/test_update_uv_cache_fix.py::test_detect_stale_cache_false` (@TEST:UPDATE-CACHE-FIX-001-002)
  - `tests/unit/test_update_uv_cache_fix.py::test_clear_cache_success` (@TEST:UPDATE-CACHE-FIX-001-003)
  - `tests/unit/test_update_uv_cache_fix.py::test_clear_cache_failure` (@TEST:UPDATE-CACHE-FIX-001-004)
  - `tests/unit/test_update_uv_cache_fix.py::test_upgrade_with_retry` (@TEST:UPDATE-CACHE-FIX-001-005)

### CODE
- **Primary**: @CODE:UPDATE-CACHE-FIX-001
- **Implementation**:
  - `src/moai_adk/cli/commands/update.py::_detect_stale_cache` (@CODE:UPDATE-CACHE-FIX-001-001)
  - `src/moai_adk/cli/commands/update.py::_clear_uv_package_cache` (@CODE:UPDATE-CACHE-FIX-001-002)
  - `src/moai_adk/cli/commands/update.py::_execute_upgrade_with_retry` (@CODE:UPDATE-CACHE-FIX-001-003)

### DOC
- **Primary**: @DOC:UPDATE-CACHE-FIX-001
- **Documentation**:
  - `README.md#troubleshooting-uv-tool-upgrade-issues` (@DOC:UPDATE-CACHE-FIX-001-001)
  - `CHANGELOG.md#v0.9.1-fixed` (@DOC:UPDATE-CACHE-FIX-001-002)
  - `.moai/specs/SPEC-UPDATE-CACHE-FIX-001/spec.md` (이 문서)
  - `.moai/specs/SPEC-UPDATE-CACHE-FIX-001/plan.md`
  - `.moai/specs/SPEC-UPDATE-CACHE-FIX-001/acceptance.md`

## 의존성

### 기존 코드 재사용
- `_get_current_version()`: 설치된 버전 조회
- `_get_latest_version()`: PyPI 최신 버전 조회
- `_detect_tool_installer()`: 설치 도구 감지 (uv/pip)

### 외부 의존성
- `subprocess` (Python 표준 라이브러리)
- `packaging.version` (기존 프로젝트 의존성)
- `uv` CLI 0.9.3+ (외부 도구)

### 테스트 의존성
- `pytest` (기존 테스트 프레임워크)
- `pytest-mock` (subprocess mocking)

## 제약사항

1. **uv 버전 요구사항**: uv 0.9.3+ 필수 (`cache clean` 명령 지원)
2. **재시도 제한**: 최대 1회만 재시도 (무한 루프 방지)
3. **타임아웃**: 캐시 정리 10초, PyPI API 호출 5초
4. **패키지 범위**: moai-adk 패키지에만 적용 (다른 패키지 영향 없음)
5. **플랫폼 지원**: Windows에서 subprocess.run 동작 확인 필요

## 향후 확장 계획

### v0.9.2+ (선택적 기능)
- `--no-retry` CLI 플래그 추가
- 환경 변수 `MOAI_ADK_UPDATE_NO_RETRY` 지원
- 상세한 디버그 모드 (`--verbose`)

### v0.10.0+ (고급 기능)
- 캐시 만료 시간 설정 가능
- 여러 패키지에 대한 일괄 캐시 정리
- 캐시 상태 진단 명령어 (`moai-adk cache status`)

---

**문서 승인 상태**: STEP 2 진행 중
**다음 단계**: plan.md 및 acceptance.md 생성 완료 후 git-manager 에이전트로 전달
