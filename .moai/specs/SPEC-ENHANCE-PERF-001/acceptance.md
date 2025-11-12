---
id: ENHANCE-PERF-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
---

# Acceptance Criteria: Hook Performance Optimization


> **Objective**: Define measurable acceptance criteria for hook performance optimization in Given-When-Then format.

---

## Overview

This document specifies testable acceptance criteria for SPEC-ENHANCE-PERF-001. All criteria must be validated before marking the SPEC as `completed`.

**Success Definition**: All hooks execute in <50ms on average across 100 invocations on all three platforms (macOS, Linux, Windows).

---

## AC1: SessionStart Hook Performance

### Scenario 1.1: First Invocation (Cold Start)

**Given**:
- Fresh Claude Code session (no cache)
- Project directory with standard MoAI-ADK structure
- `.moai/config.json` exists and is valid

**When**:
- SessionStart hook is invoked for the first time

**Then**:
- Hook execution completes in <80ms (allows for cache initialization overhead)
- Cache is initialized with project config
- Skill index is built (55 skill names mapped)
- Command metadata is loaded (4 files)
- Return value includes `config`, `skill_index`, `commands` keys

**Validation**:
```python
def test_session_start_cold_start():
    clear_cache()  # Ensure cold start
    start = time.perf_counter()
    result = on_session_start()
    elapsed = (time.perf_counter() - start) * 1000

    assert elapsed < 80, f"Cold start: {elapsed:.2f}ms (expected <80ms)"
    assert "config" in result
    assert "skill_index" in result
    assert "commands" in result
    assert len(result["skill_index"]) == 55
```

### Scenario 1.2: Subsequent Invocations (Warm Cache)

**Given**:
- SessionStart has been invoked once (cache populated)
- No file changes in `.moai/` or `.claude/`

**When**:
- SessionStart hook is invoked again

**Then**:
- Hook execution completes in <40ms (cache hit)
- Cached data is served without file I/O
- Cache hit rate increments

**Validation**:
```python
def test_session_start_warm_cache():
    on_session_start()  # Populate cache
    clear_metrics()

    times = []
    for _ in range(100):
        start = time.perf_counter()
        on_session_start()
        elapsed = (time.perf_counter() - start) * 1000
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    assert avg_time < 40, f"Warm cache average: {avg_time:.2f}ms (expected <40ms)"

    hit_rate = _hook_cache.get_hit_rate()
    assert hit_rate > 0.95, f"Cache hit rate: {hit_rate:.2%} (expected >95%)"
```

### Scenario 1.3: Cache Invalidation on File Change

**Given**:
- SessionStart has been invoked (cache populated)
- `.moai/config.json` is modified

**When**:
- SessionStart hook is invoked after file modification

**Then**:
- Cache detects file change via `mtime` check
- Fresh data is loaded from file system
- Cache is updated with new data
- Execution time <80ms (cache miss + reload)

**Validation**:
```python
def test_session_start_cache_invalidation():
    on_session_start()  # Populate cache

    # Modify config file
    config_path = Path(".moai/config.json")
    config = json.loads(config_path.read_text())
    config["test_key"] = "test_value"
    config_path.write_text(json.dumps(config, indent=2))

    start = time.perf_counter()
    result = on_session_start()
    elapsed = (time.perf_counter() - start) * 1000

    assert elapsed < 80, f"Cache miss: {elapsed:.2f}ms (expected <80ms)"
    assert result["config"]["test_key"] == "test_value"
```

---

## AC2: PreToolUse Hook Performance

### Scenario 2.1: Tool with Skills Validation

**Given**:
- Tool `code-builder` requires Skills validation
- Skills metadata is cached

**When**:
- PreToolUse hook is invoked for `code-builder`

**Then**:
- Hook execution completes in <30ms
- Skills validation uses cached data
- No file I/O performed

**Validation**:
```python
def test_pre_tool_use_with_skills():
    on_session_start()  # Populate cache

    times = []
    for _ in range(100):
        start = time.perf_counter()
        on_pre_tool_use("code-builder", {"language": "python"})
        elapsed = (time.perf_counter() - start) * 1000
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    assert avg_time < 30, f"PreToolUse average: {avg_time:.2f}ms (expected <30ms)"
```

### Scenario 2.2: Tool without Skills Requirement

**Given**:
- Tool `Read` does not require Skills validation

**When**:
- PreToolUse hook is invoked for `Read`

**Then**:
- Hook execution completes in <10ms
- Skills validation is skipped
- Minimal processing overhead

**Validation**:
```python
def test_pre_tool_use_no_skills():
    times = []
    for _ in range(100):
        start = time.perf_counter()
        on_pre_tool_use("Read", {"file_path": "/tmp/test.txt"})
        elapsed = (time.perf_counter() - start) * 1000
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    assert avg_time < 10, f"PreToolUse (no skills) average: {avg_time:.2f}ms (expected <10ms)"
```

---

## AC3: PostToolUse Hook Performance

### Scenario 3.1: SPEC Metadata Update (Buffered)

**Given**:
- PostToolUse hook has buffered metadata updates

**When**:
- PostToolUse is invoked with SPEC metadata

**Then**:
- Hook execution completes in <30ms
- Metadata is appended to in-memory buffer
- No immediate disk write (buffered)

**Validation**:
```python
def test_post_tool_use_buffered_writes():
    times = []
    for i in range(100):
        start = time.perf_counter()
        on_post_tool_use("spec-builder", {"spec_id": f"TEST-{i:03d}"})
        elapsed = (time.perf_counter() - start) * 1000
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    assert avg_time < 30, f"PostToolUse average: {avg_time:.2f}ms (expected <30ms)"

    # Verify buffer contains 100 entries
    assert len(_metadata_buffer) == 100
```

### Scenario 3.2: Buffer Flush on Session End

**Given**:
- PostToolUse has buffered 50 metadata updates

**When**:
- Claude Code session ends (or buffer reaches limit)

**Then**:
- All buffered updates are flushed to disk
- No data loss
- Flush completes in <200ms

**Validation**:
```python
def test_post_tool_use_buffer_flush():
    # Simulate 50 updates
    for i in range(50):
        on_post_tool_use("spec-builder", {"spec_id": f"TEST-{i:03d}"})

    start = time.perf_counter()
    _flush_metadata_updates()  # Manual flush
    elapsed = (time.perf_counter() - start) * 1000

    assert elapsed < 200, f"Flush time: {elapsed:.2f}ms (expected <200ms)"

    # Verify all 50 entries written
    metadata_file = Path(".moai/memory/spec-metadata.md")
    content = metadata_file.read_text()
    assert content.count("TEST-") == 50
```

---

## AC4: Notification Hook Performance

### Scenario 4.1: Buffered Log Writes

**Given**:
- Notification hook uses buffered writes (8KB buffer)

**When**:
- 10 notification messages are sent

**Then**:
- Each notification completes in <20ms
- Messages are buffered in memory
- No disk write until buffer full or flush

**Validation**:
```python
def test_notification_buffered_writes():
    times = []
    for i in range(10):
        start = time.perf_counter()
        on_notification(f"Test message {i}")
        elapsed = (time.perf_counter() - start) * 1000
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    assert avg_time < 20, f"Notification average: {avg_time:.2f}ms (expected <20ms)"
```

### Scenario 4.2: Log Flush on Exit

**Given**:
- 50 notifications buffered in memory

**When**:
- Process exits (or explicit flush)

**Then**:
- All 50 messages written to log file
- No message loss
- Flush completes in <100ms

**Validation**:
```python
def test_notification_flush_on_exit():
    for i in range(50):
        on_notification(f"Message {i}")

    start = time.perf_counter()
    _flush_log_buffer()  # Simulate exit
    elapsed = (time.perf_counter() - start) * 1000

    assert elapsed < 100, f"Flush time: {elapsed:.2f}ms (expected <100ms)"

    # Verify all 50 messages in log
    log_file = Path(".moai/logs/session.log")
    content = log_file.read_text()
    assert content.count("Message") == 50
```

---

## AC5: Cross-Platform Performance

### Scenario 5.1: macOS Performance

**Given**:
- Running on macOS (Intel or Apple Silicon)

**When**:
- All hooks are benchmarked (100 invocations each)

**Then**:
- SessionStart: <40ms average
- PreToolUse: <30ms average
- PostToolUse: <30ms average
- Notification: <20ms average

**Validation**:
```bash
# Run on macOS CI runner
pytest tests/hooks/benchmark_hooks.py --platform=macos
```

### Scenario 5.2: Linux Performance

**Given**:
- Running on Ubuntu 22.04 (GitHub Actions)

**When**:
- All hooks are benchmarked (100 invocations each)

**Then**:
- SessionStart: <40ms average
- PreToolUse: <30ms average
- PostToolUse: <30ms average
- Notification: <20ms average

**Validation**:
```bash
# Run on Ubuntu CI runner
pytest tests/hooks/benchmark_hooks.py --platform=linux
```

### Scenario 5.3: Windows Performance

**Given**:
- Running on Windows Server 2022 (GitHub Actions)

**When**:
- All hooks are benchmarked (100 invocations each)

**Then**:
- SessionStart: <50ms average (allows 10ms overhead for Windows I/O)
- PreToolUse: <35ms average
- PostToolUse: <35ms average
- Notification: <25ms average

**Validation**:
```bash
# Run on Windows CI runner
pytest tests/hooks/benchmark_hooks.py --platform=windows
```

---

## AC6: Cache Effectiveness

### Scenario 6.1: Cache Hit Rate in Typical Session

**Given**:
- Typical session with 50 hook invocations
- No file changes during session

**When**:
- Session completes

**Then**:
- Cache hit rate >80%
- Cache misses <10 (initial loads)

**Validation**:
```python
def test_cache_hit_rate_typical_session():
    # Simulate typical session
    on_session_start()
    for _ in range(10):
        on_pre_tool_use("code-builder", {})
    for _ in range(10):
        on_post_tool_use("spec-builder", {})
    for _ in range(30):
        on_notification("Progress update")

    hit_rate = _hook_cache.get_hit_rate()
    assert hit_rate > 0.80, f"Cache hit rate: {hit_rate:.2%} (expected >80%)"

    misses = _hook_cache.get_miss_count()
    assert misses < 10, f"Cache misses: {misses} (expected <10)"
```

### Scenario 6.2: Cache Size Limit Enforcement

**Given**:
- Cache size limit is 10MB
- Large project with 100+ SPEC files

**When**:
- Cache grows beyond 10MB

**Then**:
- LRU eviction triggers
- Oldest entries are evicted
- Cache size stays below 10MB

**Validation**:
```python
def test_cache_size_limit():
    # Populate cache with large data
    for i in range(200):
        large_data = {"spec_id": f"LARGE-{i:03d}", "content": "x" * 50000}
        _hook_cache.set(Path(f".moai/specs/SPEC-{i:03d}/spec.md"), large_data)

    cache_size = _hook_cache.get_size_bytes()
    assert cache_size < 10 * 1024 * 1024, f"Cache size: {cache_size / 1024 / 1024:.2f}MB (expected <10MB)"
```

---

## AC7: Backward Compatibility

### Scenario 7.1: Existing Tests Pass Unchanged

**Given**:
- Existing test suite for hooks (`tests/hooks/test_*.py`)

**When**:
- All existing tests are run

**Then**:
- 100% of existing tests pass
- No test modifications required
- Same behavior as before optimization

**Validation**:
```bash
pytest tests/hooks/ -v --ignore=tests/hooks/benchmark_hooks.py
```

### Scenario 7.2: Hook Interface Unchanged

**Given**:
- External code calling hook functions

**When**:
- Hook functions are invoked

**Then**:
- Function signatures unchanged
- Return values identical to baseline
- No breaking changes

**Validation**:
```python
def test_hook_interface_unchanged():
    # Verify function signatures
    import inspect
    sig = inspect.signature(on_session_start)
    assert sig.parameters == {}  # No parameters

    result = on_session_start()
    assert isinstance(result, dict)
    assert "config" in result
    assert "skill_index" in result
```

---

## AC8: Cache Invalidation Behavior

### Scenario 8.1: Manual Cache Disable

**Given**:
- Environment variable `MOAI_DISABLE_CACHE=1` is set

**When**:
- Hooks are invoked

**Then**:
- Cache is bypassed completely
- Fresh data loaded on every invocation
- Performance degrades to baseline (~100ms)

**Validation**:
```python
def test_cache_disable_env_var(monkeypatch):
    monkeypatch.setenv("MOAI_DISABLE_CACHE", "1")

    times = []
    for _ in range(10):
        start = time.perf_counter()
        on_session_start()
        elapsed = (time.perf_counter() - start) * 1000
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    # Without cache, should be close to baseline
    assert 80 < avg_time < 150, f"Cache disabled average: {avg_time:.2f}ms (expected 80-150ms)"
```

### Scenario 8.2: Cache Clear API

**Given**:
- Cache is populated with data

**When**:
- `clear_cache()` is called

**Then**:
- All cache entries are removed
- Next invocation triggers cache miss
- Fresh data loaded

**Validation**:
```python
def test_cache_clear_api():
    on_session_start()  # Populate cache
    assert _hook_cache.get_size() > 0

    clear_cache()
    assert _hook_cache.get_size() == 0

    # Next call should be cache miss
    start = time.perf_counter()
    on_session_start()
    elapsed = (time.perf_counter() - start) * 1000

    assert elapsed > 60, f"Should be cache miss: {elapsed:.2f}ms"
```

---

## AC9: Error Handling and Fallback

### Scenario 9.1: Graceful Degradation on Cache Failure

**Given**:
- Cache initialization fails (e.g., memory error)

**When**:
- Hooks are invoked

**Then**:
- Hooks fall back to non-cached behavior
- No exceptions raised
- Performance degrades to baseline (acceptable)

**Validation**:
```python
def test_cache_failure_fallback(monkeypatch):
    def mock_cache_init(*args, **kwargs):
        raise MemoryError("Simulated cache failure")

    monkeypatch.setattr("moai_adk.core.cache.HookCache.__init__", mock_cache_init)

    # Should not raise exception
    result = on_session_start()
    assert result is not None
    assert "config" in result
```

### Scenario 9.2: Cache Corruption Detection

**Given**:
- Cached data is corrupted (e.g., incomplete JSON)

**When**:
- Cache attempts to serve corrupted data

**Then**:
- Cache detects corruption
- Fresh data loaded from file system
- Warning logged

**Validation**:
```python
def test_cache_corruption_detection():
    # Simulate corrupted cache entry
    _hook_cache._cache["corrupted.json"] = (time.time(), {"incomplete": True})

    with pytest.warns(UserWarning, match="Cache corruption detected"):
        result = on_session_start()

    assert result is not None
```

---

## AC10: Performance Regression Prevention

### Scenario 10.1: CI/CD Performance Gate

**Given**:
- GitHub Actions CI/CD pipeline

**When**:
- PR is submitted with hook changes

**Then**:
- Performance benchmarks run automatically
- PR blocked if performance degrades >10%
- Benchmark results commented on PR

**Validation**:
```yaml
# .github/workflows/performance-check.yml
- name: Run performance benchmarks
  run: pytest tests/hooks/benchmark_hooks.py --benchmark-only

- name: Compare with baseline
  run: |
    python scripts/compare_benchmarks.py \
      --baseline=benchmarks/baseline.json \
      --current=benchmarks/current.json \
      --threshold=10
```

---

## Summary Checklist

Before marking SPEC-ENHANCE-PERF-001 as `completed`, verify:

- ✅ **AC1**: SessionStart executes in <40ms (warm cache)
- ✅ **AC2**: PreToolUse executes in <30ms
- ✅ **AC3**: PostToolUse executes in <30ms
- ✅ **AC4**: Notification executes in <20ms
- ✅ **AC5**: Performance validated on macOS, Linux, Windows
- ✅ **AC6**: Cache hit rate >80% in typical sessions
- ✅ **AC7**: All existing tests pass unchanged
- ✅ **AC8**: Cache invalidation works correctly
- ✅ **AC9**: Graceful fallback on cache failures
- ✅ **AC10**: CI/CD performance gate enforced

---

## Definition of Done

1. ✅ All acceptance criteria pass (10/10)
2. ✅ Test coverage ≥85% for cache module
3. ✅ Performance benchmarks documented in `.moai/docs/performance-tuning.md`
4. ✅ Code reviewed and approved
5. ✅ PR merged to `main` branch
6. ✅ Release notes updated

---

_Generated by spec-builder agent_
_Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)_
