# 📋 Implementation Plan: UV Tool Upgrade Cache Refresh Auto-Retry

> **SPEC Reference**: @SPEC:UPDATE-CACHE-FIX-001
> **Author**: @goos
> **Created**: 2025-10-30
> **Version**: v0.0.1
> **Status**: Draft

## 개요

이 문서는 `uv tool upgrade` 명령어의 캐시 스테일 문제를 자동으로 감지하고 해결하기 위한 5단계 구현 계획입니다.

**핵심 목표**: 사용자가 `moai-adk update` 명령 1회 실행만으로 최신 버전을 설치하도록 개선

**문제 정의**:
- 현재: PyPI 캐시가 오래되면 첫 번째 `uv tool upgrade` 실행 시 "Nothing to upgrade" 메시지 표시
- 결과: 사용자가 수동으로 `uv cache clean moai-adk` 실행 후 다시 업그레이드 필요
- 목표: 이 과정을 자동화하여 1회 실행으로 완료

## 전체 아키텍처

```
[사용자 명령]
    ↓
moai-adk update
    ↓
[1단계] _execute_upgrade()
    ↓
[2단계] _detect_stale_cache()
    ↓ (stale 감지 시)
[3단계] _clear_uv_package_cache()
    ↓
[4단계] _execute_upgrade() 재시도
    ↓
[완료] 업그레이드 성공/실패 보고
```

## 구현 단계별 계획

### PHASE 1: 캐시 스테일 감지 함수 구현

**목표**: `_detect_stale_cache()` 함수 작성

**우선순위**: HIGH (핵심 로직)

**구현 상세**:

#### 1.1 함수 시그니처 정의
```python
def _detect_stale_cache(
    upgrade_output: str,
    current_version: str,
    latest_version: str
) -> bool:
    """
    Detect if uv cache is stale by comparing versions.

    Args:
        upgrade_output: Output from uv tool upgrade command
        current_version: Currently installed version
        latest_version: Latest version available on PyPI

    Returns:
        True if cache is stale, False otherwise

    @CODE:UPDATE-CACHE-FIX-001-001
    """
```

#### 1.2 감지 로직 구현
```python
from packaging.version import parse, InvalidVersion

def _detect_stale_cache(
    upgrade_output: str,
    current_version: str,
    latest_version: str
) -> bool:
    # 출력 문자열 검증
    if not upgrade_output or "Nothing to upgrade" not in upgrade_output:
        return False

    # 버전 비교
    try:
        current = parse(current_version)
        latest = parse(latest_version)
        return current < latest
    except (InvalidVersion, TypeError) as e:
        # 버전 파싱 실패 시 graceful degradation
        logger.debug(f"Version parsing failed: {e}")
        return False
```

#### 1.3 에러 처리
- 버전 파싱 실패 → False 반환 (무음 처리)
- 출력 문자열이 None → False 반환
- 예외 발생 → DEBUG 로그 기록 후 False 반환

**코드 위치**: `src/moai_adk/cli/commands/update.py`

**@TAG 참조**: @CODE:UPDATE-CACHE-FIX-001-001

**검증 방법**:
- 단위 테스트: `test_detect_stale_cache_true`, `test_detect_stale_cache_false`
- 경계값 테스트: 동일 버전, 더 낮은 버전, 잘못된 버전 문자열

---

### PHASE 2: 캐시 정리 함수 구현

**목표**: `_clear_uv_package_cache()` 함수 작성

**우선순위**: HIGH (핵심 로직)

**구현 상세**:

#### 2.1 함수 시그니처 정의
```python
def _clear_uv_package_cache(package_name: str = "moai-adk") -> bool:
    """
    Clear uv cache for specific package.

    Args:
        package_name: Package name to clear cache for

    Returns:
        True if cache cleared successfully, False otherwise

    @CODE:UPDATE-CACHE-FIX-001-002
    """
```

#### 2.2 subprocess 호출 구현
```python
import subprocess
import logging

logger = logging.getLogger(__name__)

def _clear_uv_package_cache(package_name: str = "moai-adk") -> bool:
    try:
        result = subprocess.run(
            ["uv", "cache", "clean", package_name],
            capture_output=True,
            text=True,
            timeout=10,  # 10초 타임아웃
            check=False
        )

        if result.returncode == 0:
            logger.debug(f"UV cache cleared for {package_name}")
            return True
        else:
            logger.warning(f"Failed to clear UV cache: {result.stderr}")
            return False

    except subprocess.TimeoutExpired:
        logger.warning(f"UV cache clean timed out for {package_name}")
        return False
    except FileNotFoundError:
        logger.warning("UV command not found. Is uv installed?")
        return False
    except Exception as e:
        logger.warning(f"Unexpected error clearing cache: {e}")
        return False
```

#### 2.3 에러 처리 전략

| 에러 타입 | 처리 방법 | 로그 레벨 | 반환값 |
|---------|---------|---------|-------|
| TimeoutExpired | 무음 실패 | WARNING | False |
| FileNotFoundError | uv 미설치 메시지 | WARNING | False |
| returncode != 0 | stderr 로그 | WARNING | False |
| 기타 예외 | 예외 메시지 로그 | WARNING | False |

**코드 위치**: `src/moai_adk/cli/commands/update.py`

**@TAG 참조**: @CODE:UPDATE-CACHE-FIX-001-002

**검증 방법**:
- Mock 테스트: subprocess.run 성공/실패 시나리오
- 타임아웃 테스트: 10초 초과 시 처리
- 예외 테스트: FileNotFoundError, 기타 예외

---

### PHASE 3: 재시도 로직 통합

**목표**: `_execute_upgrade()` 함수 수정 (또는 새 래퍼 함수 생성)

**우선순위**: HIGH (핵심 통합)

**구현 상세**:

#### 3.1 기존 함수 분석
```python
# 현재 구조 (간략화)
def _execute_upgrade(installer_cmd: list[str]) -> bool:
    result = subprocess.run(installer_cmd, ...)
    return result.returncode == 0
```

#### 3.2 재시도 로직 추가
```python
def _execute_upgrade_with_retry(
    installer_cmd: list[str],
    package_name: str = "moai-adk"
) -> bool:
    """
    Execute upgrade with automatic cache retry on stale detection.

    @CODE:UPDATE-CACHE-FIX-001-003
    """
    # 1단계: 첫 번째 업그레이드 시도
    result = subprocess.run(
        installer_cmd,
        capture_output=True,
        text=True,
        check=False
    )

    # 2단계: 성공 시 조기 반환
    if result.returncode == 0 and "Nothing to upgrade" not in result.stdout:
        return True

    # 3단계: 캐시 스테일 감지
    current_version = _get_current_version()
    latest_version = _get_latest_version()

    if _detect_stale_cache(result.stdout, current_version, latest_version):
        # 4단계: 사용자 피드백
        console.print("[yellow]⚠️ Cache outdated, refreshing...[/yellow]")

        # 5단계: 캐시 정리
        if _clear_uv_package_cache(package_name):
            console.print("[cyan]♻️ Cache cleared, retrying upgrade...[/cyan]")

            # 6단계: 재시도
            result = subprocess.run(
                installer_cmd,
                capture_output=True,
                text=True,
                check=False
            )

            if result.returncode == 0:
                return True
            else:
                console.print("[red]✗ Upgrade failed after retry[/red]")
                return False
        else:
            # 캐시 정리 실패
            console.print("[red]✗ Cache clear failed. Manual workaround:[/red]")
            console.print("  [cyan]uv cache clean moai-adk && moai-adk update[/cyan]")
            return False

    # 7단계: 캐시가 최신 상태이면 원래 결과 반환
    return result.returncode == 0
```

#### 3.3 기존 코드 수정
```python
# src/moai_adk/cli/commands/update.py의 main update 함수에서:

# 변경 전:
# success = _execute_upgrade(installer_cmd)

# 변경 후:
success = _execute_upgrade_with_retry(installer_cmd, "moai-adk")
```

**코드 위치**: `src/moai_adk/cli/commands/update.py`

**@TAG 참조**: @CODE:UPDATE-CACHE-FIX-001-003

**검증 방법**:
- E2E 테스트: 캐시 스테일 상태에서 1회 실행으로 업그레이드 완료
- Mock 테스트: 재시도 로직 흐름 검증
- 무한 루프 방지: 최대 1회만 재시도 확인

---

### PHASE 4: 단위 테스트 작성

**목표**: 4개 핵심 시나리오에 대한 테스트 작성

**우선순위**: HIGH (품질 보증)

**테스트 파일**: `tests/unit/test_update_uv_cache_fix.py`

#### 4.1 테스트 케이스 1: 캐시 스테일 감지 (긍정)
```python
import pytest
from moai_adk.cli.commands.update import _detect_stale_cache

@pytest.mark.parametrize("output,current,latest,expected", [
    ("Nothing to upgrade", "0.8.3", "0.9.0", True),
    ("Nothing to upgrade", "0.8.3", "0.9.1", True),
    ("Nothing to upgrade", "0.1.0", "1.0.0", True),
])
def test_detect_stale_cache_true(output, current, latest, expected):
    """
    캐시 스테일 감지 테스트 - 긍정 케이스

    @TEST:UPDATE-CACHE-FIX-001-001
    """
    result = _detect_stale_cache(output, current, latest)
    assert result is expected
```

#### 4.2 테스트 케이스 2: 캐시 최신 상태 (부정)
```python
@pytest.mark.parametrize("output,current,latest,expected", [
    ("Already up to date", "0.9.0", "0.9.0", False),
    ("Nothing to upgrade", "0.9.0", "0.9.0", False),
    ("Successfully updated", "0.9.0", "0.9.1", False),
    ("", "0.9.0", "0.9.1", False),  # 빈 출력
])
def test_detect_stale_cache_false(output, current, latest, expected):
    """
    캐시 스테일 감지 테스트 - 부정 케이스

    @TEST:UPDATE-CACHE-FIX-001-002
    """
    result = _detect_stale_cache(output, current, latest)
    assert result is expected
```

#### 4.3 테스트 케이스 3: 캐시 정리 성공
```python
def test_clear_cache_success(mocker):
    """
    캐시 정리 성공 테스트

    @TEST:UPDATE-CACHE-FIX-001-003
    """
    mock_run = mocker.patch("subprocess.run")
    mock_run.return_value.returncode = 0
    mock_run.return_value.stderr = ""

    from moai_adk.cli.commands.update import _clear_uv_package_cache
    result = _clear_uv_package_cache("moai-adk")

    assert result is True
    mock_run.assert_called_once_with(
        ["uv", "cache", "clean", "moai-adk"],
        capture_output=True,
        text=True,
        timeout=10,
        check=False
    )
```

#### 4.4 테스트 케이스 4: 캐시 정리 실패
```python
def test_clear_cache_failure(mocker):
    """
    캐시 정리 실패 테스트

    @TEST:UPDATE-CACHE-FIX-001-004
    """
    mock_run = mocker.patch("subprocess.run")
    mock_run.return_value.returncode = 1
    mock_run.return_value.stderr = "Permission denied"

    from moai_adk.cli.commands.update import _clear_uv_package_cache
    result = _clear_uv_package_cache("moai-adk")

    assert result is False
```

#### 4.5 테스트 케이스 5: 재시도 로직 통합
```python
def test_upgrade_with_retry_stale_cache(mocker):
    """
    캐시 스테일 시 재시도 로직 테스트

    @TEST:UPDATE-CACHE-FIX-001-005
    """
    # Mock subprocess calls
    mock_run = mocker.patch("subprocess.run")

    # 첫 번째 호출: "Nothing to upgrade"
    first_call = mocker.Mock()
    first_call.returncode = 0
    first_call.stdout = "Nothing to upgrade"

    # 두 번째 호출: 업그레이드 성공
    second_call = mocker.Mock()
    second_call.returncode = 0
    second_call.stdout = "Updated moai-adk 0.8.3 -> 0.9.1"

    mock_run.side_effect = [first_call, second_call]

    # Mock 다른 함수들
    mocker.patch("moai_adk.cli.commands.update._get_current_version", return_value="0.8.3")
    mocker.patch("moai_adk.cli.commands.update._get_latest_version", return_value="0.9.1")
    mocker.patch("moai_adk.cli.commands.update._clear_uv_package_cache", return_value=True)

    # 실행
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # 검증
    assert result is True
    assert mock_run.call_count == 2  # 첫 시도 + 재시도
```

**Coverage Target**: 90%+ for cache fix code

**@TAG 참조**: @TEST:UPDATE-CACHE-FIX-001

---

### PHASE 5: 문서화

**목표**: 사용자 가이드 및 변경 기록 업데이트

**우선순위**: MEDIUM (릴리즈 전 필수)

#### 5.1 README.md 업데이트

**파일 위치**: `README.md`

**추가할 섹션**:
```markdown
### Troubleshooting UV Tool Upgrade Issues

#### Problem: "Nothing to upgrade" but new version available

**Symptoms**:
- Running `moai-adk update` shows "Nothing to upgrade"
- PyPI shows a newer version is available
- Second run of `moai-adk update` successfully upgrades

**Cause**: PyPI metadata cache becomes stale between releases

**Solution** (v0.9.1+):
moai-adk automatically detects and refreshes the cache. No action needed.

**Manual Workaround** (if automatic retry fails):
```bash
uv cache clean moai-adk
moai-adk update
```

**Technical Details**:
- Cache location: `~/.cache/uv/simple-v18/pypi/`
- Auto-retry: Triggered when "Nothing to upgrade" detected but newer version exists
- Max retries: 1 (prevents infinite loops)
```

**@TAG 참조**: @DOC:UPDATE-CACHE-FIX-001-001

#### 5.2 CHANGELOG.md 업데이트

**파일 위치**: `CHANGELOG.md`

**추가할 항목** (v0.9.1 섹션):
```markdown
## [0.9.1] - 2025-10-30

### Fixed
- **UV Tool Upgrade**: Automatic cache refresh and retry when "Nothing to upgrade" detected but newer version available on PyPI
  - Root cause: PyPI metadata cache stale between first check and actual upgrade
  - Solution: Auto-detect stale cache, clear with `uv cache clean`, and retry upgrade
  - Impact: Users no longer need to run `moai-adk update` twice
  - Implementation: @SPEC:UPDATE-CACHE-FIX-001

### Technical Details
- Added `_detect_stale_cache()` function for version comparison
- Added `_clear_uv_package_cache()` function for cache management
- Modified `_execute_upgrade()` to support automatic retry logic
- Test coverage: 90%+ for cache fix code
```

**@TAG 참조**: @DOC:UPDATE-CACHE-FIX-001-002

---

## 의존성 분석

### 기존 함수 재사용
| 함수 | 위치 | 용도 |
|-----|------|-----|
| `_get_current_version()` | `update.py` | 설치된 버전 조회 |
| `_get_latest_version()` | `update.py` | PyPI 최신 버전 조회 |
| `_detect_tool_installer()` | `update.py` | 설치 도구 감지 (uv/pip) |

### 신규 함수
| 함수 | 책임 | @TAG |
|-----|------|------|
| `_detect_stale_cache()` | 캐시 스테일 감지 | @CODE:UPDATE-CACHE-FIX-001-001 |
| `_clear_uv_package_cache()` | uv 캐시 정리 | @CODE:UPDATE-CACHE-FIX-001-002 |
| `_execute_upgrade_with_retry()` | 재시도 로직 통합 | @CODE:UPDATE-CACHE-FIX-001-003 |

### 외부 의존성
- `subprocess` (Python 표준 라이브러리)
- `packaging.version` (기존 의존성)
- `uv` CLI 0.9.3+ (외부 도구)

---

## 마일스톤

| Phase | 작업 | 의존성 | 담당 에이전트 |
|-------|------|--------|-------------|
| 1 | 캐시 감지 함수 구현 | 없음 | tdd-implementer |
| 2 | 캐시 정리 함수 구현 | 없음 | tdd-implementer |
| 3 | 재시도 로직 통합 | Phase 1, 2 | tdd-implementer |
| 4 | 단위 테스트 작성 | Phase 1, 2, 3 | tdd-implementer |
| 5 | 문서화 | Phase 4 | doc-syncer |

**Critical Path**: Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5

---

## 위험 요소 & 완화 전략

| 위험 | 확률 | 영향 | 완화 전략 | 담당 |
|-----|------|------|----------|-----|
| uv 버전 호환성 | LOW | MEDIUM | 버전 체크 + graceful fallback | tdd-implementer |
| 네트워크 지연 | LOW | LOW | 타임아웃 설정 (10초) | tdd-implementer |
| 캐시 정리 실패 | LOW | MEDIUM | 무음 실패 + 수동 안내 | tdd-implementer |
| 다중 재시도 루프 | LOW | HIGH | Max retry=1 제한 | tdd-implementer |
| Windows 플랫폼 이슈 | MEDIUM | MEDIUM | CI/CD에서 Windows 테스트 | trust-checker |
| 버전 파싱 실패 | LOW | LOW | try-except + graceful degradation | tdd-implementer |

---

## 성공 기준

### 기능적 기준
- ✅ 사용자가 1회 `moai-adk update` 실행으로 업그레이드 완료
- ✅ 캐시 오래됨이 감지되어 자동으로 해결됨
- ✅ 사용자에게 명확한 진행 상황 피드백 제공
- ✅ 캐시 정리 실패 시 수동 해결 방법 안내

### 기술적 기준
- ✅ 모든 단위 테스트 통과 (90%+ coverage)
- ✅ CI/CD 파이프라인 통과 (Linux, macOS, Windows)
- ✅ Ruff 린트 통과 (no warnings)
- ✅ mypy 타입 체크 통과
- ✅ 기존 update 테스트 여전히 통과 (회귀 테스트)

### 문서적 기준
- ✅ README.md에 Troubleshooting 섹션 추가
- ✅ CHANGELOG.md에 Bug fix 기록
- ✅ SPEC 문서 (spec.md, plan.md, acceptance.md) 완성
- ✅ @TAG 참조 완전성 검증

---

## 롤백 계획

**만약 이 SPEC 구현 후 문제가 발생하면**:

### 옵션 1: 기능 플래그로 비활성화
```python
# .moai/config.json
{
  "update": {
    "enable_cache_retry": false
  }
}
```

### 옵션 2: Git revert
```bash
git revert <commit-hash>
git push origin feature/update-cache-fix-001
```

### 옵션 3: 사용자 다운그레이드
```bash
uv tool install moai-adk==0.9.0 --force
```

### 옵션 4: 수동 캐시 정리 안내
```bash
# README.md에 명시된 수동 해결 방법
uv cache clean moai-adk
moai-adk update
```

---

## 다음 단계

### STEP 2 완료 후
1. ✅ spec.md, plan.md, acceptance.md 생성 완료
2. 📋 git-manager 에이전트로 전달
3. 🌿 GitHub 브랜치 생성: `feature/update-cache-fix-001`
4. 📝 Draft PR 생성 (Personal 모드) 또는 GitHub Issue 생성 (Team 모드)

### STEP 3 (구현)
1. `/alfred:2-run SPEC-UPDATE-CACHE-FIX-001` 실행
2. TDD 사이클: RED → GREEN → REFACTOR
3. 단위 테스트 우선 작성
4. 구현 완료 후 커버리지 확인

### STEP 4 (문서 동기화)
1. `/alfred:3-sync` 실행
2. @TAG 체인 검증
3. 문서 일관성 확인
4. PR 완성 및 리뷰 요청

---

**문서 상태**: DRAFT
**승인 상태**: STEP 1 완료, STEP 2 진행 중
**다음 파일**: acceptance.md 생성 예정
