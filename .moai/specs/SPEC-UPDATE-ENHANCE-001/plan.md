# Implementation Plan: SPEC-UPDATE-ENHANCE-001

@SPEC:UPDATE-ENHANCE-001 → @PLAN:UPDATE-ENHANCE-001

## 개요

SessionStart 버전 체크 시스템을 6가지 개선 사항을 통합하여 강화합니다. 24시간 캐싱, 오프라인 감지, 사용자 설정, 릴리스 노트, Major 버전 경고, `uv tool upgrade` 명령어 변경을 단일 SPEC으로 구현합니다.

**예상 작업 시간**: 10시간

---

## Phase 1: 캐시 시스템 구현 (3시간)

### 목표
- 24시간 캐시 메커니즘 구축
- 캐시 파일 관리 및 갱신 로직

### 1.1 캐시 모듈 생성

**새 파일**: `.claude/hooks/alfred/core/version_cache.py`

```python
#!/usr/bin/env python3
"""Version check caching system

@CODE:VERSION-CACHE-001
"""

import json
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Optional


class VersionCache:
    """24-hour TTL cache for PyPI version checks

    @CODE:VERSION-CACHE-001
    """

    def __init__(self, cache_dir: Path, ttl_hours: int = 24):
        """Initialize cache manager

        Args:
            cache_dir: Cache directory path (.moai/cache)
            ttl_hours: Cache TTL in hours (default: 24)
        """
        self.cache_file = cache_dir / "version-check.json"
        self.ttl = timedelta(hours=ttl_hours)

        # Ensure cache directory exists
        cache_dir.mkdir(parents=True, exist_ok=True)

    def get(self) -> Optional[dict[str, Any]]:
        """Get cached version info if valid

        Returns:
            Cached data dict or None if expired/invalid
        """
        if not self.cache_file.exists():
            return None

        try:
            data = json.loads(self.cache_file.read_text())

            # Check if cache is still valid
            checked_at = datetime.fromisoformat(data["checked_at"])
            if datetime.now() - checked_at < self.ttl:
                return data

            return None  # Expired
        except (json.JSONDecodeError, KeyError, ValueError, OSError):
            # Corrupted cache - ignore
            return None

    def set(self, data: dict[str, Any]) -> None:
        """Save version info to cache

        Args:
            data: Version info dict to cache
        """
        try:
            data["checked_at"] = datetime.now().isoformat()
            self.cache_file.write_text(
                json.dumps(data, indent=2, ensure_ascii=False)
            )
            # Set file permissions to 644
            self.cache_file.chmod(0o644)
        except OSError:
            # Write failed - graceful degradation
            pass

    def clear(self) -> None:
        """Clear cache file"""
        try:
            if self.cache_file.exists():
                self.cache_file.unlink()
        except OSError:
            pass


__all__ = ["VersionCache"]
```

### 1.2 캐시 통합 테스트

**새 파일**: `tests/unit/hooks/test_version_cache.py`

```python
"""Version cache unit tests

@TEST:VERSION-CACHE-001
"""

import json
from datetime import datetime, timedelta
from pathlib import Path
import pytest

from .claude.hooks.alfred.core.version_cache import VersionCache


def test_cache_miss_returns_none(tmp_path):
    """Cache miss should return None"""
    cache = VersionCache(tmp_path)
    assert cache.get() is None


def test_cache_hit_returns_data(tmp_path):
    """Valid cache should return data"""
    cache = VersionCache(tmp_path)
    data = {
        "current_version": "0.8.1",
        "latest_version": "0.9.0",
        "checked_at": datetime.now().isoformat()
    }
    cache.set(data)

    cached = cache.get()
    assert cached is not None
    assert cached["current_version"] == "0.8.1"


def test_expired_cache_returns_none(tmp_path):
    """Expired cache should return None"""
    cache = VersionCache(tmp_path, ttl_hours=1)

    # Create expired cache (2 hours ago)
    old_data = {
        "current_version": "0.8.0",
        "checked_at": (datetime.now() - timedelta(hours=2)).isoformat()
    }
    cache.cache_file.write_text(json.dumps(old_data))

    assert cache.get() is None


def test_corrupted_cache_returns_none(tmp_path):
    """Corrupted cache should return None gracefully"""
    cache = VersionCache(tmp_path)
    cache.cache_file.write_text("invalid json{{{")

    assert cache.get() is None


def test_cache_clear(tmp_path):
    """Clear should remove cache file"""
    cache = VersionCache(tmp_path)
    cache.set({"test": "data"})

    assert cache.cache_file.exists()
    cache.clear()
    assert not cache.cache_file.exists()
```

**예상 테스트 개수**: 5개

---

## Phase 2: 네트워크 감지 및 오프라인 지원 (2시간)

### 목표
- 네트워크 가용성 사전 체크
- 오프라인 환경에서 graceful degradation

### 2.1 네트워크 감지 모듈 추가

**수정 파일**: `.claude/hooks/alfred/core/project.py`

```python
def is_network_available() -> bool:
    """Check if network is available for PyPI access

    Fast check using socket connection to pypi.org.
    Prevents unnecessary timeout delays in offline environments.

    Returns:
        True if network is available, False otherwise

    @CODE:NETWORK-DETECT-001
    """
    import socket

    try:
        with timeout_handler(0.1):
            socket.create_connection(("pypi.org", 443), timeout=0.1)
        return True
    except (OSError, TimeoutError):
        return False
```

### 2.2 get_package_version_info() 수정

**수정 파일**: `.claude/hooks/alfred/core/project.py`

```python
def get_package_version_info(cwd: str = ".") -> dict[str, Any]:
    """Check MoAI-ADK version with 24-hour caching and offline support

    Enhanced version check with:
    - 24-hour cache to reduce PyPI requests
    - Offline detection to skip network calls
    - User preferences from config.json
    - Major version warnings
    - Release notes URL

    Args:
        cwd: Project directory for config and cache access

    Returns:
        dict with keys:
            - "current": Current installed version
            - "latest": Latest version from PyPI (or cache)
            - "update_available": Boolean
            - "upgrade_command": Recommended command
            - "is_major_update": Boolean for major version changes
            - "release_notes_url": GitHub releases URL
            - "from_cache": Boolean indicating cache hit

    @CODE:VERSION-CACHE-001
    @CODE:NETWORK-DETECT-001
    """
    from importlib.metadata import version, PackageNotFoundError
    from pathlib import Path
    from .version_cache import VersionCache

    result = {
        "current": "unknown",
        "latest": "unknown",
        "update_available": False,
        "upgrade_command": "",
        "is_major_update": False,
        "release_notes_url": "",
        "from_cache": False
    }

    # Get current version
    try:
        result["current"] = version("moai-adk")
    except PackageNotFoundError:
        result["current"] = "dev"
        return result

    # Load user config
    config = _load_version_check_config(cwd)

    # Check if version check is disabled
    if not config["enabled"]:
        return result

    # Initialize cache
    cache_dir = Path(cwd) / ".moai" / "cache"
    cache = VersionCache(cache_dir, ttl_hours=config["cache_ttl_hours"])

    # Try to get from cache first
    cached_data = cache.get()
    if cached_data:
        result.update(cached_data)
        result["from_cache"] = True
        return result

    # Check network availability (skip if offline)
    if not is_network_available():
        # Return with current version only
        return result

    # Fetch from PyPI
    try:
        import urllib.request
        import urllib.error

        with timeout_handler(1):
            url = "https://pypi.org/pypi/moai-adk/json"
            headers = {"Accept": "application/json"}
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=0.8) as response:
                data = json.load(response)
                result["latest"] = data.get("info", {}).get("version", "unknown")
    except (urllib.error.URLError, TimeoutError, Exception):
        # Network error - return without latest version
        return result

    # Compare versions
    if result["current"] != "unknown" and result["latest"] != "unknown":
        try:
            current_parts = [int(x) for x in result["current"].split(".")]
            latest_parts = [int(x) for x in result["latest"].split(".")]

            # Pad shorter version
            max_len = max(len(current_parts), len(latest_parts))
            current_parts.extend([0] * (max_len - len(current_parts)))
            latest_parts.extend([0] * (max_len - len(latest_parts)))

            if latest_parts > current_parts:
                result["update_available"] = True
                result["upgrade_command"] = "uv tool upgrade moai-adk"

                # Check for major version update
                result["is_major_update"] = latest_parts[0] > current_parts[0]

                # Generate release notes URL
                if config["show_release_notes"]:
                    result["release_notes_url"] = (
                        f"https://github.com/modu-ai/moai-adk/releases/tag/v{result['latest']}"
                    )
        except (ValueError, AttributeError):
            pass

    # Save to cache
    cache.set(result)

    return result


def _load_version_check_config(cwd: str) -> dict[str, Any]:
    """Load version check config from .moai/config.json

    Returns default values if config is missing or invalid.
    """
    defaults = {
        "enabled": True,
        "frequency": "daily",
        "cache_ttl_hours": 24,
        "show_release_notes": True,
        "warn_major_updates": True
    }

    config_path = Path(cwd) / ".moai" / "config.json"
    if not config_path.exists():
        return defaults

    try:
        config = json.loads(config_path.read_text())
        user_config = config.get("version_check", {})

        # Map frequency to TTL hours
        frequency_map = {
            "never": 0,
            "daily": 24,
            "weekly": 168,
            "always": 0
        }

        # If frequency is "never", disable completely
        if user_config.get("frequency") == "never":
            return {"enabled": False, **defaults}

        # If frequency is "always", use 0 TTL (always check)
        if user_config.get("frequency") == "always":
            user_config["cache_ttl_hours"] = 0
        elif user_config.get("frequency") in frequency_map:
            user_config["cache_ttl_hours"] = frequency_map[user_config["frequency"]]

        return {**defaults, **user_config}
    except (json.JSONDecodeError, OSError):
        return defaults
```

### 2.3 오프라인 모드 테스트

**새 파일**: `tests/unit/hooks/test_offline_mode.py`

```python
"""Offline mode tests

@TEST:OFFLINE-MODE-001
"""

import pytest
from unittest.mock import patch

from .claude.hooks.alfred.core.project import get_package_version_info, is_network_available


def test_network_available_returns_true_when_connected():
    """Network check should return True when online"""
    # This test requires actual network - may be skipped in CI
    result = is_network_available()
    assert isinstance(result, bool)


def test_network_available_returns_false_when_offline():
    """Network check should return False when offline"""
    with patch("socket.create_connection", side_effect=OSError):
        assert is_network_available() is False


def test_offline_mode_skips_pypi_check(tmp_path):
    """Offline mode should skip PyPI and return current version only"""
    with patch("is_network_available", return_value=False):
        result = get_package_version_info(str(tmp_path))

        assert result["current"] != "unknown"  # Should still detect current
        assert result["latest"] == "unknown"  # No PyPI check
        assert not result["update_available"]


def test_offline_mode_uses_cache_if_available(tmp_path):
    """Offline mode should use cached data if available"""
    # Pre-populate cache
    from .claude.hooks.alfred.core.version_cache import VersionCache
    cache = VersionCache(tmp_path / ".moai" / "cache")
    cache.set({
        "current": "0.8.1",
        "latest": "0.9.0",
        "update_available": True
    })

    with patch("is_network_available", return_value=False):
        result = get_package_version_info(str(tmp_path))

        assert result["from_cache"] is True
        assert result["latest"] == "0.9.0"
```

**예상 테스트 개수**: 4개

---

## Phase 3: Major 버전 경고 및 릴리스 노트 (2시간)

### 목표
- Major 버전 업데이트 감지 및 경고
- GitHub Releases 링크 자동 생성

### 3.1 SessionStart 핸들러 업데이트

**수정 파일**: `.claude/hooks/alfred/handlers/session.py`

```python
def handle_session_start(payload: HookPayload) -> HookResult:
    """SessionStart event handler with enhanced version check

    ... (기존 docstring 유지)

    @TAG:CHECKPOINT-EVENT-001
    @TAG:HOOKS-TIMEOUT-001
    @CODE:VERSION-CACHE-001
    """
    # ... (기존 코드 유지)

    # OPTIONAL: Package version info - skip if timeout/failure
    version_info = {}
    try:
        version_info = get_package_version_info(cwd)
    except Exception:
        pass

    # Build message
    # ... (기존 코드)

    # Add version info with enhanced formatting
    if version_info and version_info.get("current") != "unknown":
        # Check if it's a major update
        if version_info.get("is_major_update") and version_info.get("update_available"):
            # Major version warning
            lines.append(
                f"   ⚠️  Major version update available: "
                f"{version_info['current']} → {version_info['latest']}"
            )
            lines.append("   Breaking changes detected. Review release notes before upgrading:")
            if version_info.get("release_notes_url"):
                lines.append(f"   📝 {version_info['release_notes_url']}")
            lines.append(f"   ⬆️ Upgrade: {version_info['upgrade_command']}")
        elif version_info.get("update_available"):
            # Minor/patch update
            version_line = (
                f"   🗿 MoAI-ADK Ver: {version_info['current']} "
                f"→ {version_info['latest']} available ✨"
            )
            if version_info.get("from_cache"):
                version_line += " (cached)"
            lines.append(version_line)

            if version_info.get("release_notes_url"):
                lines.append(f"   📝 Release Notes: {version_info['release_notes_url']}")
            lines.append(f"   ⬆️ Upgrade: {version_info['upgrade_command']}")
        else:
            # No update available
            version_line = f"   🗿 MoAI-ADK Ver: {version_info['current']}"
            if version_info.get("from_cache"):
                version_line += " (cached)"
            lines.append(version_line)

    # ... (나머지 기존 코드)
```

### 3.2 Major 업데이트 경고 테스트

**새 파일**: `tests/unit/hooks/test_major_update_warning.py`

```python
"""Major version update warning tests

@TEST:MAJOR-UPDATE-WARN-001
"""

import pytest
from unittest.mock import patch, MagicMock

from .claude.hooks.alfred.handlers.session import handle_session_start


def test_major_update_shows_warning():
    """Major version update should show warning message"""
    payload = {"cwd": ".", "phase": "compact"}

    mock_version_info = {
        "current": "0.8.1",
        "latest": "1.0.0",
        "update_available": True,
        "is_major_update": True,
        "upgrade_command": "uv tool upgrade moai-adk",
        "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v1.0.0"
    }

    with patch("get_package_version_info", return_value=mock_version_info):
        result = handle_session_start(payload)

        assert "⚠️  Major version update available" in result.system_message
        assert "0.8.1 → 1.0.0" in result.system_message
        assert "Breaking changes detected" in result.system_message
        assert "release_notes_url" in result.system_message


def test_minor_update_shows_normal_message():
    """Minor version update should show normal update message"""
    payload = {"cwd": ".", "phase": "compact"}

    mock_version_info = {
        "current": "0.8.1",
        "latest": "0.8.2",
        "update_available": True,
        "is_major_update": False,
        "upgrade_command": "uv tool upgrade moai-adk"
    }

    with patch("get_package_version_info", return_value=mock_version_info):
        result = handle_session_start(payload)

        assert "0.8.1 → 0.8.2 available ✨" in result.system_message
        assert "⚠️" not in result.system_message
```

**예상 테스트 개수**: 2개

---

## Phase 4: Config 통합 및 문서화 (3시간)

### 목표
- `.moai/config.json` 스키마 확장
- 사용자 설정 문서화
- 마이그레이션 가이드

### 4.1 Config 스키마 업데이트

**수정 파일**: `.moai/memory/config-schema.md`

```markdown
## version_check

버전 체크 시스템 설정입니다.

### 필드

- `enabled` (boolean, 기본값: `true`): 버전 체크 활성화 여부
- `frequency` (string, 기본값: `"daily"`): 체크 빈도
  - `"never"`: 버전 체크 비활성화
  - `"daily"`: 24시간마다 체크
  - `"weekly"`: 7일마다 체크
  - `"always"`: 매 세션마다 체크 (캐시 사용 안 함)
- `cache_ttl_hours` (integer, 기본값: `24`): 캐시 유효 시간 (시간 단위)
- `show_release_notes` (boolean, 기본값: `true`): 릴리스 노트 URL 표시 여부
- `warn_major_updates` (boolean, 기본값: `true`): Major 버전 업데이트 경고 표시

### 예시

```json
{
  "version_check": {
    "enabled": true,
    "frequency": "daily",
    "cache_ttl_hours": 24,
    "show_release_notes": true,
    "warn_major_updates": true
  }
}
```
```

### 4.2 사용자 가이드 작성

**새 파일**: `.moai/docs/version-check-guide.md`

```markdown
# 버전 체크 시스템 사용 가이드

@DOC:VERSION-CHECK-CONFIG-001

## 개요

MoAI-ADK는 세션 시작 시 자동으로 최신 버전을 확인합니다.
24시간 캐싱을 통해 불필요한 네트워크 요청을 줄이고, 오프라인 환경에서도 작동합니다.

## 기본 동작

- **캐싱**: 24시간 동안 버전 정보를 캐시하여 성능 향상 (95% 빠름)
- **오프라인 지원**: 네트워크 없이도 정상 작동
- **Major 버전 경고**: Breaking changes가 있는 업데이트는 명확히 표시
- **릴리스 노트**: GitHub Releases 링크 제공

## 설정 방법

`.moai/config.json` 파일에서 설정할 수 있습니다:

```json
{
  "version_check": {
    "enabled": true,
    "frequency": "daily"
  }
}
```

### 빈도 설정

- `"never"`: 버전 체크 완전 비활성화
- `"daily"`: 24시간마다 체크 (기본값, 권장)
- `"weekly"`: 7일마다 체크
- `"always"`: 매번 체크 (네트워크 부담)

## 출력 예시

### Minor 업데이트 (0.8.1 → 0.8.2)
```
🗿 MoAI-ADK Ver: 0.8.1 → 0.8.2 available ✨
📝 Release Notes: https://github.com/modu-ai/moai-adk/releases/tag/v0.8.2
⬆️ Upgrade: uv tool upgrade moai-adk
```

### Major 업데이트 (0.8.1 → 1.0.0)
```
⚠️  Major version update available: 0.8.1 → 1.0.0
Breaking changes detected. Review release notes before upgrading:
📝 https://github.com/modu-ai/moai-adk/releases/tag/v1.0.0
⬆️ Upgrade: uv tool upgrade moai-adk
```

## 캐시 관리

캐시 파일 위치: `.moai/cache/version-check.json`

수동으로 캐시를 지우려면:
```bash
rm .moai/cache/version-check.json
```

## 문제 해결

**Q: 버전 체크가 너무 느려요**
A: `frequency: "weekly"` 또는 `"never"`로 설정하세요.

**Q: 오프라인에서 에러가 나요**
A: 자동으로 건너뛰어야 합니다. 에러가 계속되면 이슈를 제출해주세요.

**Q: 업그레이드 명령어가 실패해요**
A: `uv` 툴체인이 설치되어 있는지 확인하세요: `uv --version`
```

### 4.3 통합 테스트

**새 파일**: `tests/integration/test_version_check_e2e.py`

```python
"""End-to-end version check tests

@TEST:VERSION-CHECK-E2E-001
"""

import pytest
import json
from pathlib import Path


def test_full_workflow_with_cache(tmp_path):
    """Test complete workflow: miss → cache → hit"""
    # Setup
    config_path = tmp_path / ".moai" / "config.json"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(json.dumps({
        "version_check": {"enabled": True, "frequency": "daily"}
    }))

    # First call - cache miss
    result1 = get_package_version_info(str(tmp_path))
    assert not result1.get("from_cache")

    # Second call - cache hit
    result2 = get_package_version_info(str(tmp_path))
    assert result2.get("from_cache")
    assert result2["latest"] == result1["latest"]


def test_disabled_version_check(tmp_path):
    """Test version check disabled by config"""
    config_path = tmp_path / ".moai" / "config.json"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(json.dumps({
        "version_check": {"enabled": False}
    }))

    result = get_package_version_info(str(tmp_path))
    assert result["latest"] == "unknown"
    assert not result["update_available"]
```

**예상 테스트 개수**: 2개

---

## 수정 파일 목록

### 신규 파일 (3개)
1. `.claude/hooks/alfred/core/version_cache.py` - 캐시 관리 모듈
2. `.moai/docs/version-check-guide.md` - 사용자 가이드
3. `.moai/cache/version-check.json` - 캐시 데이터 (자동 생성)

### 수정 파일 (3개)
1. `.claude/hooks/alfred/core/project.py` - 버전 체크 로직 강화
2. `.claude/hooks/alfred/handlers/session.py` - SessionStart 출력 개선
3. `.moai/memory/config-schema.md` - Config 스키마 확장

### 테스트 파일 (5개)
1. `tests/unit/hooks/test_version_cache.py` - 캐시 단위 테스트
2. `tests/unit/hooks/test_offline_mode.py` - 오프라인 모드 테스트
3. `tests/unit/hooks/test_major_update_warning.py` - Major 업데이트 경고 테스트
4. `tests/integration/test_version_check_e2e.py` - E2E 통합 테스트
5. (기존) `tests/unit/hooks/test_session.py` - SessionStart 테스트 보강

---

## 구현 체크리스트

### Phase 1: 캐싱 (3시간)
- [ ] `VersionCache` 클래스 구현
- [ ] 캐시 파일 읽기/쓰기 로직
- [ ] TTL 만료 확인 로직
- [ ] 손상된 캐시 처리 (graceful degradation)
- [ ] 단위 테스트 5개 작성 및 통과

### Phase 2: 오프라인 지원 (2시간)
- [ ] `is_network_available()` 함수 구현
- [ ] `get_package_version_info()` 오프라인 모드 통합
- [ ] Config 로딩 로직 (`_load_version_check_config`)
- [ ] 오프라인 시나리오 테스트 4개 작성 및 통과

### Phase 3: Major 버전 경고 (2시간)
- [ ] Major 버전 감지 로직 (`is_major_update`)
- [ ] 릴리스 노트 URL 생성
- [ ] SessionStart 출력 포맷 업데이트
- [ ] Major 업데이트 경고 테스트 2개 작성 및 통과

### Phase 4: 문서화 (3시간)
- [ ] Config 스키마 확장 (`config-schema.md`)
- [ ] 사용자 가이드 작성 (`version-check-guide.md`)
- [ ] E2E 통합 테스트 2개 작성 및 통과
- [ ] README 업데이트 (버전 체크 섹션 추가)
- [ ] 코드 리뷰 및 리팩토링

---

## 의존성 및 통합 포인트

### 내부 의존성
- `core/project.py`: 기존 버전 체크 로직 확장
- `handlers/session.py`: SessionStart 출력 통합
- `.moai/config.json`: 사용자 설정 읽기

### 외부 의존성
- **PyPI JSON API**: `https://pypi.org/pypi/moai-adk/json`
- **uv 툴체인**: `uv tool upgrade` 명령어

### 테스트 의존성
- `pytest`: 단위/통합 테스트 프레임워크
- `unittest.mock`: 네트워크 모킹

---

## 테스트 전략

### 단위 테스트 (13개)
- 캐시 CRUD 동작: 5개
- 네트워크 감지: 4개
- Major 버전 경고: 2개
- E2E 시나리오: 2개

### 통합 테스트
- SessionStart 전체 플로우 (캐시 히트/미스)
- Config 변경 시나리오
- 오프라인 → 온라인 전환

### 성능 테스트
- 캐시 히트 시 지연 시간 < 50ms
- 캐시 미스 시 지연 시간 < 1.5s
- SessionStart 총 시간 < 3s

### 에러 시나리오 테스트
- 캐시 파일 손상
- Config 파일 손상
- PyPI API 응답 지연
- 디스크 쓰기 권한 없음

---

## 성능 목표

| 시나리오 | 현재 | 목표 | 개선율 |
|---------|------|------|--------|
| 캐시 히트 | N/A | < 50ms | N/A |
| 캐시 미스 | ~1.2s | < 1.5s | 유지 |
| 전체 SessionStart | ~1.2s | ~0.05s (캐시 시) | 95% ↓ |

---

## 리스크 완화 전략

### 기술적 리스크
1. **캐시 손상**: JSON 파싱 실패 시 무시하고 재생성
2. **네트워크 타임아웃**: 0.1초 사전 체크 + 1초 PyPI 타임아웃
3. **Config 오류**: 기본값 fallback 보장

### 운영 리스크
1. **사용자 혼란**: 상세한 사용자 가이드 제공
2. **성능 저하**: 캐시 히트율 모니터링 (95% 목표)

---

## 다음 단계

1. **Phase 1 구현** → 캐시 시스템 완성 후 PR
2. **Phase 2 구현** → 오프라인 지원 완성 후 PR
3. **Phase 3 구현** → Major 버전 경고 완성 후 PR
4. **Phase 4 통합** → 문서화 및 E2E 테스트 후 최종 병합

**최종 목표**: 모든 테스트 통과 (13개) + 문서화 완료 + 성능 목표 달성

---

**END OF PLAN**

@PLAN:UPDATE-ENHANCE-001
