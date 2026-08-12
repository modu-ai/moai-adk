# Post-Change Profiling — SPEC-HOOK-PRETOOL-PERF-001 M3

> Post-change measurement with M1 config disk cache + M2 lazy slice.
> Cache pre-warmed: one invocation ran before the parallel batches to populate
> the cache, so all subsequent invocations hit the cache (REQ-PERF-001).

## Measurement configuration

- **Parallelism**: 8 parallel `moai hook pre-tool` invocations per batch
- **Batches**: 5
- **Total invocations**: 40
- **Platform**: darwin/arm64
- **Timestamp**: 2026-08-12T15:34:57Z

## Per-phase timing (ms)

| Phase | p50 | p99 | max |
|-------|-----|-----|-----|
| External wall-time | 53.43 | 83.39 | 83.39 |
| Fork+exec | 0.16 | 1.28 | 1.28 |
| Config load | 0.81 | 1.84 | 1.84 |
| Dispatch (security scan) | 0.01 | 0.04 | 0.04 |
| Internal total | 1.94 | 3.83 | 3.83 |

## Analysis

### Config-load improvement

On a warm cache hit, config-load reads a SINGLE cache file instead of ~20
per-section YAML files. The config-load p50 is 0.81 ms, which represents
the cache-hit path (one file read + JSON unmarshal vs ~20 file reads + YAML parse + merge).

### Concurrent-stress tail (p99/max)

The external wall-time tail (p99: 83.39 ms) is dominated by OS scheduling and
disk I/O contention under concurrent stress, NOT by config-load. The cache
improves the per-invocation config-load cost but cannot eliminate the fork+exec
amplification that occurs when 8+ processes contend for CPU and memory.

The residual tail risk is addressed by REQ-PERF-010: the 10s timeout remains
in place, and narrowing it toward 5s requires the post-change measurement to
demonstrate the cost has dropped. The cache+lazy approach provides a structural
improvement (fewer file reads per invocation); a daemon (C-2, out of scope)
would be the follow-up to eliminate the fork+exec cost entirely.
