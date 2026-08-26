# Profiling Baseline — SPEC-HOOK-PRETOOL-PERF-001 M0

> Pre-change baseline. Captured under simulated concurrent-hook stress.

## Measurement configuration

- **Parallelism**: 8 parallel `moai hook pre-tool` invocations per batch
- **Batches**: 5
- **Total invocations**: 40
- **Platform**: darwin/arm64
- **Timestamp**: 2026-08-23T17:02:15Z

## Per-phase timing (ms)

| Phase | p50 | p99 | max |
|-------|-----|-----|-----|
| External wall-time | 95.59 | 4805.41 | 4805.41 |
| Fork+exec | 0.34 | 9.27 | 9.27 |
| Config load | 1.54 | 5.74 | 5.74 |
| Dispatch (security scan) | 0.01 | 0.03 | 0.03 |
| Internal total | 3.29 | 12.04 | 12.04 |

## Diagnosis

Config load is the dominant per-invocation cost under concurrent stress: config-load p50 is 1.54 ms vs dispatch p50 of 0.01 ms and fork+exec p50 of 0.34 ms.
The config-load phase reads ~20 per-section YAML files on every invocation, even though the PreToolUse handler only consumes a thin slice (security policy, branch-guard flag, gate config). This confirms the SPEC-HOOK-PRETOOL-PERF-001 diagnosis: the config disk cache (M1) + lazy slice (M2) attack the dominant cost.
