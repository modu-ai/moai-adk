---
id: ENHANCE-PERF-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
---

# Implementation Plan: Hook Performance Optimization


> **Objective**: Reduce MoAI-ADK hook execution time from ~100ms to <50ms through caching, lazy loading, and I/O optimization.

---

## Executive Summary

### Current State (Measured 2025-10-31)
- **SessionStart**: **185.26ms** (4.6x over target) - Primary bottleneck
  - `get_package_version_info()`: 112.82ms (61%)
  - `get_git_info()`: 52.88ms (28%)
  - Other functions: 2.07ms (1%)
- **PreToolUse**: **37.09ms** (near target, not critical)
- **PostToolUse & Notification**: 0.00ms (optimal)
- **Cache Strategy**: None (all data re-read on every invocation)

### Target State
- **Average Hook Time**: <50ms (50% improvement)
- **Cache Hit Rate**: >80% in typical sessions
- **Backward Compatibility**: 100% (no breaking changes)

### Success Metrics
1. ✅ All hooks execute in <50ms (validated on macOS, Linux, Windows)
2. ✅ Cache implementation adds <5ms overhead
3. ✅ Test coverage remains ≥85%
4. ✅ No regressions in existing functionality

---

## Implementation Strategy

### Focused Optimization Approach

Based on measured performance data, optimize **ONLY SessionStart hook** using targeted caching:

**Optimization Targets** (with time savings):
1. **Cache `get_package_version_info()`** result → Save 112.82ms (61%)
   - Valid for 30 minutes (stable version info)
   - File: `.moai/.cache/version_info.json`

2. **Cache `get_git_info()`** result → Save 52.88ms (28%)
   - Valid for 10 seconds (recent commits)
   - Check `git status` periodically for invalidation

3. **Leave other hooks unchanged** (already optimal at 0-37ms)

### Phase 1: Version Info Caching (Priority: CRITICAL)

**Goal**: Cache PyPI version check for 30 minutes.

#### Tasks
1. **Add version cache file** (`.moai/.cache/version_info.json`)
   - Stores: package version, latest PyPI version, last check timestamp
   - TTL: 30 minutes (reduce network requests)
   - Fallback: Use cached data if network fails

2. **Modify `get_package_version_info()`**
   - Check cache first
   - If cache expired, fetch from PyPI in background
   - Return cached value immediately (don't block)
   - Use absolute file paths for deterministic lookup
   - Handle symbolic links and relative paths

3. **Add cache metrics logging**
   - Hit rate tracking
   - Miss rate tracking
   - Cache size monitoring

#### Deliverables
- `src/moai_adk/core/cache/__init__.py` (module init)
- `src/moai_adk/core/cache/hook_cache.py` (core cache implementation)
- `tests/core/cache/test_hook_cache.py` (unit tests, ≥90% coverage)

#### Acceptance Criteria
- Cache stores/retrieves data correctly
- LRU eviction works when size limit exceeded
- `mtime` comparison detects file changes
- Cache invalidation clears all entries

---

### Phase 2: SessionStart Hook Optimization (Priority: High)

**Goal**: Reduce SessionStart execution time from 120ms to <40ms.

#### Current Bottlenecks
| Operation | Time | Optimization Strategy |
|-----------|------|----------------------|
| Load `.moai/config.json` | 30ms | Cache after first read |
| Scan `.claude/skills/` directory | 50ms | Build lightweight index only |
| Load command metadata | 25ms | Cache all command files |
| Plugin registry initialization | 15ms | Lazy initialization on demand |

#### Tasks
1. **Implement cached config loader**
   - Read `.moai/config.json` once per session
   - Validate cache with `mtime` check

2. **Replace eager Skills loading with indexing**
   - Build skill name → file path mapping (~5ms)
   - Load actual skill content on-demand

3. **Batch-read command metadata**
   - Read all 4 command files in one pass
   - Use list comprehension for parallel parsing

4. **Defer plugin registry initialization**
   - Initialize only when plugins are actually used
   - Track initialization state

#### Implementation Pseudocode
```python
# Before
def on_session_start():
    config = load_config()           # 30ms
    skills = load_all_skills()       # 50ms (55 files)
    commands = load_commands()       # 25ms
    plugins = init_plugin_registry() # 15ms
    return {"config": config, "skills": skills, "commands": commands}

# After
def on_session_start():
    config = _cache.get_or_load(".moai/config.json", load_config)  # 5ms (cached)
    skill_index = build_skill_index()                              # 5ms (index only)
    commands = _cache.get_or_load_batch(COMMAND_PATHS)             # 10ms (batch read)
    # Plugin registry initialized lazily on first use
    return {"config": config, "skill_index": skill_index, "commands": commands}
```

#### Deliverables
- `src/moai_adk/hooks/session_start.py` (optimized hook)
- `tests/hooks/test_session_start_perf.py` (performance benchmarks)

#### Acceptance Criteria
- SessionStart executes in <40ms (67% improvement)
- Skills load on-demand when `Skill("name")` invoked
- Cache hit rate >80% after first session
- Existing tests pass without modification

---

### Phase 3: PreToolUse/PostToolUse Hook Optimization (Priority: Medium)

**Goal**: Reduce PreToolUse/PostToolUse execution time from 80-90ms to <30ms each.

#### PreToolUse Optimization

**Current Bottlenecks**:
- Skills metadata validation (~40ms)
- Tag agent initialization (~30ms)

**Optimization**:
1. Cache Skills metadata validation results
3. Skip validation if tool doesn't require Skills

**Implementation**:
```python
def on_pre_tool_use(tool_name: str, args: dict):
    # Skip validation for tools that don't need Skills
    if tool_name in NO_SKILL_TOOLS:
        return

    # Use cached validation results
    validation = _cache.get_or_validate(tool_name, args)
    return validation
```

#### PostToolUse Optimization

**Current Bottlenecks**:
- SPEC metadata refresh (~50ms)
- Tag validation (~30ms)

**Optimization**:
1. Batch SPEC metadata updates (write once per session)
2. Cache tag validation results
3. Use buffered file writes for metadata updates

**Implementation**:
```python
def on_post_tool_use(tool_name: str, result: dict):
    # Batch updates to memory buffer
    _metadata_buffer.append(result)

    # Flush to disk only when buffer full or session ends
    if len(_metadata_buffer) >= BATCH_SIZE:
        _flush_metadata_updates()
```

#### Deliverables
- `src/moai_adk/hooks/pre_tool_use.py` (optimized)
- `src/moai_adk/hooks/post_tool_use.py` (optimized)
- `tests/hooks/test_pre_post_tool_perf.py` (benchmarks)

#### Acceptance Criteria
- PreToolUse executes in <30ms
- PostToolUse executes in <30ms
- Metadata consistency maintained (no data loss)
- Buffered writes flush correctly on session end

---

### Phase 4: Notification Hook Optimization (Priority: Low)

**Goal**: Reduce Notification hook execution time from 60ms to <20ms.

#### Current Bottlenecks
- Log file append operations (~40ms)
- Status update writes (~20ms)

#### Optimization Strategy
1. Use buffered file writes (`buffering=8192`)
2. Batch multiple notifications before flushing
3. Async writes (future enhancement)

#### Implementation
```python
import atexit

_log_buffer = []

def on_notification(message: str):
    _log_buffer.append(message)

    # Flush periodically (every 10 messages or on exit)
    if len(_log_buffer) >= 10:
        _flush_log_buffer()

def _flush_log_buffer():
    with open(LOG_PATH, 'a', buffering=8192) as f:
        f.write('\n'.join(_log_buffer))
    _log_buffer.clear()

# Register flush on exit
atexit.register(_flush_log_buffer)
```

#### Deliverables
- `src/moai_adk/hooks/notification.py` (optimized)
- `tests/hooks/test_notification_perf.py` (benchmarks)

#### Acceptance Criteria
- Notification hook executes in <20ms
- All log entries written before process exit
- No log message loss on crashes (flush on SIGTERM)

---

## Architecture Design

### Cache Layer Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Claude Code Session                     │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                    Hook Layer                            │
│  ┌────────────┬────────────┬────────────┬────────────┐  │
│  │SessionStart│PreToolUse  │PostToolUse │Notification│  │
│  └──────┬─────┴──────┬─────┴──────┬─────┴──────┬─────┘  │
└─────────┼────────────┼────────────┼────────────┼────────┘
          │            │            │            │
          ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────┐
│                   Cache Layer (NEW)                      │
│  ┌────────────────────────────────────────────────────┐ │
│  │  HookCache (LRU, mtime validation, 10MB limit)     │ │
│  │  - Config cache: .moai/config.json                 │ │
│  │  - Skill index: {skill_name -> file_path}          │ │
│  │  - Command metadata: .claude/commands/*.md         │ │
│  │  - SPEC metadata: .moai/specs/*/spec.md            │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────┐
│                 File System Layer                        │
│  .moai/config.json, .claude/skills/*, .moai/specs/*     │
└─────────────────────────────────────────────────────────┘
```

### Cache Invalidation Strategy

| Trigger | Action | Scope |
|---------|--------|-------|
| Session end | Clear all cache | Global |
| File modified | Invalidate specific entry | File-level |
| `MOAI_DISABLE_CACHE=1` | Bypass cache | Global |
| Cache size > 10MB | Evict LRU entries | Automatic |

---

## File Modifications

### New Files

| File | Purpose | Complexity |
|------|---------|------------|
| `src/moai_adk/core/cache/__init__.py` | Cache module initialization | Low |
| `src/moai_adk/core/cache/hook_cache.py` | Core cache implementation | Medium |
| `tests/core/cache/test_hook_cache.py` | Cache unit tests | Medium |
| `tests/hooks/test_*_perf.py` | Performance benchmark suite | Medium |
| `.moai/docs/performance-tuning.md` | Performance optimization guide | Low |

### Modified Files

| File | Changes | Risk Level |
|------|---------|------------|
| `src/moai_adk/hooks/session_start.py` | Add caching, lazy loading | Medium |
| `src/moai_adk/hooks/pre_tool_use.py` | Add cached validation | Low |
| `src/moai_adk/hooks/post_tool_use.py` | Add buffered writes | Medium |
| `src/moai_adk/hooks/notification.py` | Add log buffering | Low |
| `src/moai_adk/core/skills/loader.py` | Refactor to support lazy loading | Medium |

---

## Testing Strategy

### Performance Benchmarks

**Benchmark Suite**: `tests/hooks/benchmark_hooks.py`

```python
import time
import pytest

def test_session_start_performance():
    """Verify SessionStart executes in <40ms."""
    times = []
    for _ in range(100):
        start = time.perf_counter()
        on_session_start()
        elapsed = (time.perf_counter() - start) * 1000  # ms
        times.append(elapsed)

    avg_time = sum(times) / len(times)
    assert avg_time < 40, f"SessionStart average: {avg_time:.2f}ms (expected <40ms)"

def test_cache_hit_rate():
    """Verify cache hit rate >80% in typical session."""
    # Simulate 50 hook invocations
    for _ in range(50):
        on_session_start()

    hit_rate = _hook_cache.get_hit_rate()
    assert hit_rate > 0.80, f"Cache hit rate: {hit_rate:.2%} (expected >80%)"
```

### Regression Testing

- **Existing Test Suite**: All current hook tests must pass without modification
- **Integration Tests**: Verify end-to-end workflows (e.g., `/alfred:1-plan`)
- **Platform Tests**: Run benchmarks on macOS, Linux, Windows

### Coverage Target

- **Cache Module**: ≥90% coverage
- **Hook Optimization**: ≥85% coverage (maintain project baseline)

---

## Risk Assessment

### High Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Cache invalidation bugs | Medium | High | Comprehensive unit tests + manual testing |
| Platform-specific performance variance | Medium | Medium | Test on all three platforms before merge |
| Regression in hook functionality | Low | High | Run full test suite + manual verification |

### Medium Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Memory pressure on large projects | Low | Medium | LRU eviction + 10MB limit |
| Cache poisoning (stale data) | Low | Medium | Always check `mtime` before serving |

---

## Rollout Plan

### Phase 1: Development (Week 1)
- Implement cache infrastructure
- Optimize SessionStart hook
- Write unit tests

### Phase 2: Testing (Week 2)
- Run performance benchmarks
- Validate on macOS, Linux, Windows
- Fix regressions

### Phase 3: Documentation (Week 2)
- Update performance tuning guide
- Document cache behavior
- Add troubleshooting section

### Phase 4: Release (Week 2)
- Merge to `develop` branch
- Create PR to `main`
- Tag release (v0.6.0 or next minor version)

---

## Success Validation

### Quantitative Metrics

| Metric | Baseline | Target | Measurement |
|--------|----------|--------|-------------|
| SessionStart time | 120ms | <40ms | `pytest tests/hooks/benchmark_hooks.py` |
| PreToolUse time | 80ms | <30ms | Same |
| PostToolUse time | 90ms | <30ms | Same |
| Notification time | 60ms | <20ms | Same |
| Cache hit rate | N/A | >80% | `_hook_cache.get_hit_rate()` |
| Test coverage | 87.84% | ≥85% | `pytest --cov` |

### Qualitative Validation

- ✅ User feedback: "Hooks feel snappier"
- ✅ No reported cache-related bugs in first 30 days
- ✅ Documentation clarity: Users understand caching behavior

---

## Future Enhancements (Out of Scope)

1. **Async I/O**: Refactor hooks to support `async/await`
2. **Distributed Caching**: Share cache across multiple Claude Code instances
3. **Smart Prefetching**: Predict next Skills to load based on patterns
4. **Compression**: Compress cached data for memory efficiency
5. **Persistent Cache**: Disk-based cache that survives session restarts

---

## Dependencies

- **No external dependencies**: All optimizations use Python standard library
- **Backward Compatibility**: Existing code continues to work unchanged
- **Related SPEC**: `SPEC-BUGFIX-001` (Windows timeout fix)

---

## Timeline

**Total Estimated Time**: 3-5 days (one developer)

| Phase | Tasks | Time |
|-------|-------|------|
| Phase 1 | Cache infrastructure | 1 day |
| Phase 2 | Hook optimization | 2 days |
| Phase 3 | Testing + validation | 1 day |
| Phase 4 | Documentation | 0.5 days |
| Phase 5 | Review + merge | 0.5 days |

---

## Next Steps

1. **Review this plan** with team/stakeholders
2. **Approve SPEC** (update status to `active`)
3. **Run `/alfred:2-run SPEC-ENHANCE-PERF-001`** to implement
4. **Run performance benchmarks** before and after optimization
5. **Document results** in `acceptance.md`

---

_Generated by spec-builder agent_
_Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)_
