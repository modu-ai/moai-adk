# Acceptance Criteria: SPEC-UPDATE-ENHANCE-001

@SPEC:UPDATE-ENHANCE-001 → @ACCEPTANCE:UPDATE-ENHANCE-001

## 개요

SessionStart 버전 체크 시스템 강화를 위한 상세 수락 기준 및 테스트 시나리오입니다.
모든 시나리오는 Given-When-Then (Gherkin) 형식으로 작성되며, 자동화된 테스트로 검증됩니다.

---

## AC-1: 24시간 캐싱 동작 검증

### AC-1.1: 캐시 미스 시 PyPI 호출

```gherkin
Feature: Version check with cache miss

Scenario: First session start with no cache
  Given 캐시 파일이 존재하지 않음
  And 네트워크 연결이 정상임
  When SessionStart 이벤트가 발생함
  Then PyPI API를 호출하여 최신 버전을 조회함
  And 결과를 .moai/cache/version-check.json에 저장함
  And 사용자에게 버전 정보를 표시함
  And from_cache 플래그는 False임

Acceptance:
  - 캐시 파일이 생성됨
  - 파일 권한이 644임
  - JSON 형식이 올바름 (current_version, latest_version, checked_at 포함)
  - SessionStart 지연 시간 < 1.5초
```

**테스트 코드**:
```python
def test_cache_miss_calls_pypi(tmp_path, mock_pypi):
    """AC-1.1: Cache miss should call PyPI and save cache"""
    cache_file = tmp_path / ".moai" / "cache" / "version-check.json"
    assert not cache_file.exists()

    result = get_package_version_info(str(tmp_path))

    assert cache_file.exists()
    assert oct(cache_file.stat().st_mode)[-3:] == '644'
    assert not result["from_cache"]
    assert result["latest"] != "unknown"
```

### AC-1.2: 캐시 히트 시 PyPI 건너뛰기

```gherkin
Feature: Version check with cache hit

Scenario: Session start within 24 hours of last check
  Given 캐시 파일이 존재함
  And 캐시가 24시간 이내에 생성됨
  When SessionStart 이벤트가 발생함
  Then PyPI API 호출을 건너뜀
  And 캐시된 데이터를 반환함
  And from_cache 플래그는 True임

Acceptance:
  - PyPI API 호출 횟수: 0
  - SessionStart 지연 시간 < 50ms (95% 개선)
  - 캐시 히트율 > 95% (7일간 모니터링)
```

**테스트 코드**:
```python
def test_cache_hit_skips_pypi(tmp_path, monkeypatch):
    """AC-1.2: Cache hit should skip PyPI call"""
    # Pre-populate cache
    cache = VersionCache(tmp_path / ".moai" / "cache")
    cache.set({"current": "0.8.1", "latest": "0.9.0"})

    # Mock PyPI to ensure it's not called
    pypi_called = False
    def mock_urlopen(*args, **kwargs):
        nonlocal pypi_called
        pypi_called = True
        raise AssertionError("PyPI should not be called with valid cache")

    monkeypatch.setattr("urllib.request.urlopen", mock_urlopen)

    result = get_package_version_info(str(tmp_path))

    assert not pypi_called
    assert result["from_cache"] is True
    assert result["latest"] == "0.9.0"
```

### AC-1.3: 캐시 만료 시 갱신

```gherkin
Feature: Expired cache refresh

Scenario: Session start after cache expiration (>24 hours)
  Given 캐시 파일이 존재함
  And 캐시가 24시간 이전에 생성됨 (만료)
  When SessionStart 이벤트가 발생함
  Then PyPI API를 호출하여 최신 버전을 조회함
  And 캐시를 새 데이터로 갱신함
  And from_cache 플래그는 False임

Acceptance:
  - 만료된 캐시는 무시됨
  - 새로운 캐시가 생성됨
  - checked_at 타임스탬프가 현재 시간으로 업데이트됨
```

**테스트 코드**:
```python
def test_expired_cache_refreshes(tmp_path, mock_pypi):
    """AC-1.3: Expired cache should trigger refresh"""
    from datetime import datetime, timedelta

    # Create expired cache (25 hours ago)
    cache_file = tmp_path / ".moai" / "cache" / "version-check.json"
    cache_file.parent.mkdir(parents=True)
    old_time = datetime.now() - timedelta(hours=25)
    cache_file.write_text(json.dumps({
        "current": "0.7.0",
        "latest": "0.7.5",
        "checked_at": old_time.isoformat()
    }))

    result = get_package_version_info(str(tmp_path))

    assert not result["from_cache"]
    new_cache = json.loads(cache_file.read_text())
    assert new_cache["latest"] == "0.9.0"  # Mock response
    assert datetime.fromisoformat(new_cache["checked_at"]) > old_time
```

---

## AC-2: 오프라인 감지 및 처리

### AC-2.1: 오프라인 감지 성공

```gherkin
Feature: Offline detection

Scenario: No network connectivity
  Given 네트워크 연결이 불가능함
  When is_network_available() 함수를 호출함
  Then False를 반환함
  And 0.1초 이내에 응답함

Acceptance:
  - 타임아웃: 100ms
  - DNS 조회 실패 또는 socket 연결 실패 감지
  - 예외 발생하지 않음 (graceful)
```

**테스트 코드**:
```python
def test_offline_detection(monkeypatch):
    """AC-2.1: Offline detection should work reliably"""
    import time

    # Mock socket to simulate offline
    def mock_create_connection(*args, **kwargs):
        raise OSError("Network unreachable")

    monkeypatch.setattr("socket.create_connection", mock_create_connection)

    start = time.time()
    result = is_network_available()
    duration = time.time() - start

    assert result is False
    assert duration < 0.2  # Should timeout quickly
```

### AC-2.2: 오프라인 시 PyPI 건너뛰기

```gherkin
Feature: Offline mode graceful degradation

Scenario: Network unavailable during version check
  Given 네트워크 연결이 불가능함
  And 캐시 파일이 존재하지 않음
  When SessionStart 이벤트가 발생함
  Then PyPI API 호출을 건너뜀
  And current_version만 표시함
  And latest_version은 "unknown"임
  And 에러 메시지를 표시하지 않음

Acceptance:
  - SessionStart 정상 완료
  - 사용자에게 에러 노출 안 함
  - 로그 레벨: DEBUG (선택적)
```

**테스트 코드**:
```python
def test_offline_mode_graceful_degradation(tmp_path, monkeypatch):
    """AC-2.2: Offline mode should gracefully skip PyPI"""
    monkeypatch.setattr("is_network_available", lambda: False)

    result = get_package_version_info(str(tmp_path))

    assert result["current"] != "unknown"  # Should detect installed version
    assert result["latest"] == "unknown"
    assert not result["update_available"]
    # No exception raised
```

### AC-2.3: 오프라인에서 캐시 활용

```gherkin
Feature: Offline cache usage

Scenario: Offline with valid cache
  Given 네트워크 연결이 불가능함
  And 유효한 캐시가 존재함 (24시간 이내)
  When SessionStart 이벤트가 발생함
  Then 캐시 데이터를 활용함
  And 버전 정보를 정상 표시함
  And from_cache 플래그는 True임

Acceptance:
  - 오프라인이지만 버전 정보 표시 가능
  - 사용자 경험 저하 없음
```

**테스트 코드**:
```python
def test_offline_uses_cache(tmp_path, monkeypatch):
    """AC-2.3: Offline mode should use cache if available"""
    # Pre-populate cache
    cache = VersionCache(tmp_path / ".moai" / "cache")
    cache.set({
        "current": "0.8.1",
        "latest": "0.9.0",
        "update_available": True
    })

    monkeypatch.setattr("is_network_available", lambda: False)

    result = get_package_version_info(str(tmp_path))

    assert result["from_cache"] is True
    assert result["latest"] == "0.9.0"
    assert result["update_available"] is True
```

---

## AC-3: Major 버전 경고

### AC-3.1: Major 버전 감지

```gherkin
Feature: Major version detection

Scenario: Major version update available (0.8.1 → 1.0.0)
  Given 현재 버전이 0.8.1임
  And PyPI 최신 버전이 1.0.0임
  When 버전 비교를 수행함
  Then is_major_update 플래그가 True임
  And 경고 메시지를 표시함

Acceptance:
  - Major 버전 번호 증가 감지 (0.x → 1.x, 1.x → 2.x)
  - Minor/Patch 업데이트는 False 반환
```

**테스트 코드**:
```python
def test_major_version_detection():
    """AC-3.1: Major version changes should be detected"""
    assert is_major_update("0.8.1", "1.0.0") is True
    assert is_major_update("1.5.0", "2.0.0") is True
    assert is_major_update("0.8.1", "0.9.0") is False
    assert is_major_update("1.2.3", "1.3.0") is False
```

### AC-3.2: Major 버전 경고 메시지

```gherkin
Feature: Major version warning display

Scenario: Major update warning in SessionStart
  Given Major 버전 업데이트가 감지됨
  And warn_major_updates 설정이 True임
  When SessionStart 메시지를 생성함
  Then "⚠️  Major version update available" 메시지를 포함함
  And "Breaking changes detected" 경고를 표시함
  And 릴리스 노트 URL을 포함함

Acceptance:
  - 경고 메시지가 눈에 띄게 표시됨 (⚠️ 이모지)
  - 릴리스 노트 링크 포함
  - 업그레이드 명령어 포함
```

**테스트 코드**:
```python
def test_major_update_warning_message(tmp_path, monkeypatch):
    """AC-3.2: Major update should show warning message"""
    mock_version_info = {
        "current": "0.8.1",
        "latest": "1.0.0",
        "update_available": True,
        "is_major_update": True,
        "release_notes_url": "https://github.com/modu-ai/moai-adk/releases/tag/v1.0.0",
        "upgrade_command": "uv tool upgrade moai-adk"
    }
    monkeypatch.setattr("get_package_version_info", lambda cwd: mock_version_info)

    payload = {"cwd": str(tmp_path), "phase": "compact"}
    result = handle_session_start(payload)

    assert "⚠️  Major version update available" in result.system_message
    assert "0.8.1 → 1.0.0" in result.system_message
    assert "Breaking changes detected" in result.system_message
    assert "release_notes_url" in result.system_message
```

### AC-3.3: Minor/Patch 업데이트 정상 표시

```gherkin
Feature: Non-major update display

Scenario: Minor/Patch update available (0.8.1 → 0.8.2)
  Given Minor 또는 Patch 업데이트가 감지됨
  When SessionStart 메시지를 생성함
  Then 일반 업데이트 메시지를 표시함
  And "→ 0.8.2 available ✨" 형식으로 표시함
  And Major 업데이트 경고를 표시하지 않음

Acceptance:
  - 경고 이모지(⚠️) 미사용
  - "Breaking changes" 문구 미포함
  - 릴리스 노트는 선택적으로 표시
```

**테스트 코드**:
```python
def test_minor_update_normal_message(tmp_path, monkeypatch):
    """AC-3.3: Minor update should show normal message"""
    mock_version_info = {
        "current": "0.8.1",
        "latest": "0.8.2",
        "update_available": True,
        "is_major_update": False,
        "upgrade_command": "uv tool upgrade moai-adk"
    }
    monkeypatch.setattr("get_package_version_info", lambda cwd: mock_version_info)

    payload = {"cwd": str(tmp_path), "phase": "compact"}
    result = handle_session_start(payload)

    assert "0.8.1 → 0.8.2 available ✨" in result.system_message
    assert "⚠️" not in result.system_message
    assert "Breaking changes" not in result.system_message
```

---

## AC-4: 사용자 설정 존중

### AC-4.1: 버전 체크 비활성화

```gherkin
Feature: Disable version check

Scenario: User sets frequency to "never"
  Given .moai/config.json에 frequency: "never" 설정됨
  When SessionStart 이벤트가 발생함
  Then 버전 체크를 완전히 건너뜀
  And 현재 버전만 표시함 (업데이트 정보 미표시)
  And PyPI API 호출하지 않음

Acceptance:
  - PyPI 호출 횟수: 0
  - "→ X.X.X available" 메시지 미표시
  - 업그레이드 명령어 미표시
```

**테스트 코드**:
```python
def test_disabled_version_check(tmp_path, monkeypatch):
    """AC-4.1: Disabled version check should skip entirely"""
    config_path = tmp_path / ".moai" / "config.json"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(json.dumps({
        "version_check": {"enabled": True, "frequency": "never"}
    }))

    pypi_called = False
    def mock_urlopen(*args, **kwargs):
        nonlocal pypi_called
        pypi_called = True
    monkeypatch.setattr("urllib.request.urlopen", mock_urlopen)

    result = get_package_version_info(str(tmp_path))

    assert not pypi_called
    assert result["latest"] == "unknown"
```

### AC-4.2: 주간 빈도 설정

```gherkin
Feature: Weekly check frequency

Scenario: User sets frequency to "weekly"
  Given .moai/config.json에 frequency: "weekly" 설정됨
  When 캐시 TTL을 확인함
  Then cache_ttl_hours가 168시간 (7일)임
  And 7일 이내 캐시는 유효함

Acceptance:
  - 7일간 PyPI 호출 안 함
  - 캐시 만료 후 자동 갱신
```

**테스트 코드**:
```python
def test_weekly_frequency_config(tmp_path):
    """AC-4.2: Weekly frequency should use 168-hour TTL"""
    config_path = tmp_path / ".moai" / "config.json"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(json.dumps({
        "version_check": {"frequency": "weekly"}
    }))

    config = _load_version_check_config(str(tmp_path))

    assert config["cache_ttl_hours"] == 168
```

### AC-4.3: 항상 체크 설정

```gherkin
Feature: Always check mode

Scenario: User sets frequency to "always"
  Given .moai/config.json에 frequency: "always" 설정됨
  When SessionStart 이벤트가 발생함
  Then 매번 PyPI API를 호출함
  And 캐시를 사용하지 않음 (TTL: 0)

Acceptance:
  - 캐시 TTL: 0시간
  - 매 세션마다 최신 버전 확인
  - 네트워크 부담 증가 (사용자 선택)
```

**테스트 코드**:
```python
def test_always_check_config(tmp_path):
    """AC-4.3: Always mode should disable caching"""
    config_path = tmp_path / ".moai" / "config.json"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(json.dumps({
        "version_check": {"frequency": "always"}
    }))

    config = _load_version_check_config(str(tmp_path))

    assert config["cache_ttl_hours"] == 0
```

---

## AC-5: uv tool upgrade 명령어

### AC-5.1: 명령어 변경 확인

```gherkin
Feature: Upgrade command update

Scenario: Update available
  Given 최신 버전이 감지됨
  When 업그레이드 명령어를 생성함
  Then "uv tool upgrade moai-adk" 명령어를 사용함
  And "uv pip install --upgrade" 명령어를 사용하지 않음

Acceptance:
  - 모든 업데이트 메시지에 일관된 명령어 사용
  - 이전 명령어(uv pip install) 제거됨
```

**테스트 코드**:
```python
def test_upgrade_command_uses_uv_tool(tmp_path, mock_pypi):
    """AC-5.1: Upgrade command should use 'uv tool upgrade'"""
    result = get_package_version_info(str(tmp_path))

    if result["update_available"]:
        assert result["upgrade_command"] == "uv tool upgrade moai-adk"
        assert "uv pip install" not in result["upgrade_command"]
```

### AC-5.2: SessionStart 출력 검증

```gherkin
Feature: SessionStart upgrade command display

Scenario: Display upgrade command in SessionStart
  Given 업데이트가 가능함
  When SessionStart 메시지를 생성함
  Then "⬆️ Upgrade: uv tool upgrade moai-adk" 메시지를 포함함

Acceptance:
  - 일관된 명령어 표시
  - Copy-paste 가능한 형식
```

**테스트 코드**:
```python
def test_session_start_shows_uv_tool_command(tmp_path, monkeypatch):
    """AC-5.2: SessionStart should display 'uv tool upgrade'"""
    mock_version_info = {
        "current": "0.8.1",
        "latest": "0.9.0",
        "update_available": True,
        "upgrade_command": "uv tool upgrade moai-adk"
    }
    monkeypatch.setattr("get_package_version_info", lambda cwd: mock_version_info)

    payload = {"cwd": str(tmp_path), "phase": "compact"}
    result = handle_session_start(payload)

    assert "⬆️ Upgrade: uv tool upgrade moai-adk" in result.system_message
```

---

## AC-6: 릴리스 노트 URL

### AC-6.1: GitHub Releases URL 생성

```gherkin
Feature: Release notes URL generation

Scenario: Generate GitHub Releases URL
  Given 최신 버전이 0.9.0임
  And show_release_notes 설정이 True임
  When 릴리스 노트 URL을 생성함
  Then "https://github.com/modu-ai/moai-adk/releases/tag/v0.9.0" 형식임

Acceptance:
  - GitHub URL 형식 준수
  - 버전 번호 앞에 "v" 접두사 포함
```

**테스트 코드**:
```python
def test_release_notes_url_format(tmp_path, mock_pypi):
    """AC-6.1: Release notes URL should follow GitHub format"""
    result = get_package_version_info(str(tmp_path))

    if result["release_notes_url"]:
        expected = f"https://github.com/modu-ai/moai-adk/releases/tag/v{result['latest']}"
        assert result["release_notes_url"] == expected
```

### AC-6.2: 릴리스 노트 표시 제어

```gherkin
Feature: Release notes display control

Scenario: User disables release notes
  Given show_release_notes 설정이 False임
  When 버전 정보를 조회함
  Then release_notes_url 필드가 빈 문자열임
  And SessionStart에 릴리스 노트가 표시되지 않음

Acceptance:
  - 사용자 설정 존중
  - URL 생성 건너뛰기
```

**테스트 코드**:
```python
def test_release_notes_disabled(tmp_path):
    """AC-6.2: Release notes can be disabled by config"""
    config_path = tmp_path / ".moai" / "config.json"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(json.dumps({
        "version_check": {"show_release_notes": False}
    }))

    result = get_package_version_info(str(tmp_path))

    assert result["release_notes_url"] == ""
```

---

## 성능 벤치마크

### 벤치마크 1: 캐시 히트 성능

```gherkin
Feature: Cache hit performance

Scenario: Measure cache hit latency
  Given 유효한 캐시가 존재함
  When 1000번 연속 버전 체크를 수행함
  Then 평균 지연 시간 < 50ms
  And 95th percentile < 100ms

Acceptance:
  - 평균: < 50ms
  - P95: < 100ms
  - P99: < 150ms
```

**벤치마크 코드**:
```python
def test_cache_hit_performance(tmp_path, benchmark):
    """Benchmark: Cache hit should be < 50ms"""
    cache = VersionCache(tmp_path / ".moai" / "cache")
    cache.set({"current": "0.8.1", "latest": "0.9.0"})

    result = benchmark(get_package_version_info, str(tmp_path))

    # pytest-benchmark assertion
    assert result["from_cache"] is True
    # Benchmark reports mean, stddev, min, max automatically
```

### 벤치마크 2: SessionStart 총 시간

```gherkin
Feature: SessionStart performance

Scenario: Measure total SessionStart duration
  Given 캐시가 유효함
  When SessionStart 이벤트를 처리함
  Then 총 지연 시간 < 3초
  And 버전 체크 부분 < 50ms (캐시 히트)

Acceptance:
  - SessionStart 총 시간: < 3초
  - 버전 체크 기여: < 50ms (캐시 히트 시)
```

**벤치마크 코드**:
```python
def test_session_start_performance(tmp_path, benchmark):
    """Benchmark: SessionStart should complete < 3s"""
    payload = {"cwd": str(tmp_path), "phase": "compact"}

    result = benchmark(handle_session_start, payload)

    assert result.system_message is not None
    # Benchmark reports timing automatically
```

---

## 에러 시나리오 및 복구

### 에러 1: 캐시 파일 손상

```gherkin
Feature: Corrupted cache recovery

Scenario: Invalid JSON in cache file
  Given 캐시 파일에 손상된 JSON이 있음
  When 버전 정보를 조회함
  Then 캐시를 무시하고 PyPI를 호출함
  And 새 캐시를 생성함
  And 에러를 발생시키지 않음

Acceptance:
  - Graceful degradation (에러 미노출)
  - 자동 복구 (캐시 재생성)
  - 사용자 경험 저하 없음
```

**테스트 코드**:
```python
def test_corrupted_cache_recovery(tmp_path, mock_pypi):
    """Error scenario: Corrupted cache should auto-recover"""
    cache_file = tmp_path / ".moai" / "cache" / "version-check.json"
    cache_file.parent.mkdir(parents=True)
    cache_file.write_text("invalid json{{{")

    # Should not raise exception
    result = get_package_version_info(str(tmp_path))

    assert result["current"] != "unknown"
    # Cache should be regenerated
    new_cache = json.loads(cache_file.read_text())
    assert "latest_version" in new_cache
```

### 에러 2: 디스크 쓰기 권한 없음

```gherkin
Feature: Disk write failure handling

Scenario: No write permission to .moai/cache/
  Given .moai/cache/ 디렉토리에 쓰기 권한이 없음
  When 캐시를 저장하려고 시도함
  Then 에러를 무시하고 계속 진행함
  And 사용자에게 버전 정보는 표시함
  And SessionStart는 정상 완료됨

Acceptance:
  - 권한 오류 무시 (graceful)
  - SessionStart 실패하지 않음
  - 메모리 캐시로 대체 (선택적)
```

**테스트 코드**:
```python
def test_cache_write_failure_graceful(tmp_path, monkeypatch, mock_pypi):
    """Error scenario: Cache write failure should be graceful"""
    def mock_write_text(*args, **kwargs):
        raise OSError("Permission denied")

    monkeypatch.setattr("pathlib.Path.write_text", mock_write_text)

    # Should not raise exception
    result = get_package_version_info(str(tmp_path))

    assert result["current"] != "unknown"
    assert result["latest"] != "unknown"
```

### 에러 3: PyPI API 응답 지연

```gherkin
Feature: PyPI timeout handling

Scenario: PyPI API takes > 1 second
  Given PyPI API가 응답하지 않음
  When 1초 타임아웃이 발생함
  Then TimeoutError를 발생시킴
  And 캐시 데이터를 활용하거나 건너뜀
  And SessionStart는 정상 완료됨

Acceptance:
  - 타임아웃: 1초 (hard limit)
  - SessionStart 총 시간 < 3초 (타임아웃 포함)
```

**테스트 코드**:
```python
def test_pypi_timeout_handling(tmp_path, monkeypatch):
    """Error scenario: PyPI timeout should be handled gracefully"""
    import time

    def mock_urlopen(*args, **kwargs):
        time.sleep(2)  # Simulate slow response
        raise TimeoutError("PyPI timeout")

    monkeypatch.setattr("urllib.request.urlopen", mock_urlopen)

    # Should not crash
    result = get_package_version_info(str(tmp_path))

    # Should return with limited info
    assert result["current"] != "unknown"
    assert result["latest"] == "unknown"
```

---

## 통합 테스트 (E2E)

### E2E-1: 완전한 워크플로우

```gherkin
Feature: End-to-end version check workflow

Scenario: Complete user journey
  Given 프로젝트가 초기화됨
  And 캐시가 존재하지 않음
  When 첫 번째 SessionStart 발생 (캐시 미스)
  Then PyPI에서 최신 버전 조회
  And 캐시 생성
  When 두 번째 SessionStart 발생 (1시간 후)
  Then 캐시에서 버전 정보 로드
  When 세 번째 SessionStart 발생 (25시간 후, 캐시 만료)
  Then PyPI에서 다시 조회
  And 캐시 갱신

Acceptance:
  - 전체 플로우 정상 동작
  - 캐시 히트/미스 시나리오 모두 검증
```

**테스트 코드**:
```python
def test_e2e_version_check_workflow(tmp_path, mock_pypi):
    """E2E: Complete version check workflow"""
    # First session - cache miss
    result1 = get_package_version_info(str(tmp_path))
    assert not result1["from_cache"]

    # Second session - cache hit
    result2 = get_package_version_info(str(tmp_path))
    assert result2["from_cache"]
    assert result2["latest"] == result1["latest"]

    # Simulate cache expiry
    cache_file = tmp_path / ".moai" / "cache" / "version-check.json"
    old_data = json.loads(cache_file.read_text())
    old_data["checked_at"] = (datetime.now() - timedelta(hours=25)).isoformat()
    cache_file.write_text(json.dumps(old_data))

    # Third session - cache miss (expired)
    result3 = get_package_version_info(str(tmp_path))
    assert not result3["from_cache"]
```

---

## 품질 게이트 기준

### 테스트 커버리지
- **단위 테스트**: 95% 이상
- **통합 테스트**: 모든 주요 시나리오 커버
- **엣지 케이스**: 에러 시나리오 100% 커버

### 성능 기준
- **캐시 히트**: < 50ms (평균)
- **캐시 미스**: < 1.5s (PyPI 포함)
- **SessionStart 총 시간**: < 3s

### 에러 처리
- **Graceful degradation**: 100% (모든 실패 시나리오에서 SessionStart 성공)
- **에러 로그**: 사용자에게 노출 없음 (DEBUG 레벨만 허용)

---

## Definition of Done

- [ ] 모든 수락 기준 (AC-1 ~ AC-6) 검증 완료
- [ ] 13개 단위 테스트 통과
- [ ] 2개 통합 테스트 (E2E) 통과
- [ ] 성능 벤치마크 달성 (< 50ms 캐시 히트)
- [ ] 에러 시나리오 테스트 통과 (3개)
- [ ] 코드 리뷰 완료 및 승인
- [ ] 사용자 가이드 문서 작성 완료
- [ ] Config 스키마 업데이트 완료
- [ ] 모든 테스트 CI/CD 파이프라인에서 통과

---

**END OF ACCEPTANCE CRITERIA**

@ACCEPTANCE:UPDATE-ENHANCE-001
