# Profiling Baseline — SPEC-HOOK-PRETOOL-PERF-001 M0

> Pre-change baseline. Captured under simulated concurrent-hook stress.

## Measurement configuration

- **Parallelism**: 8 parallel `moai hook pre-tool` invocations per batch
- **Batches**: 5
- **Total invocations**: 40
- **Platform**: darwin/arm64
- **Timestamp**: 2026-08-12T15:34:43Z

## Per-phase timing (ms)

| Phase | p50 | p99 | max |
|-------|-----|-----|-----|
| External wall-time | 136.87 | 894.21 | 894.21 |
| Fork+exec | 0.17 | 19.99 | 19.99 |
| Config load | 1.26 | 16.95 | 16.95 |
| Dispatch (security scan) | 0.01 | 0.10 | 0.10 |
| Internal total | 3.03 | 69.99 | 69.99 |

## Diagnosis

Config load is the dominant per-invocation cost under concurrent stress: config-load p50 is 1.26 ms vs dispatch p50 of 0.01 ms and fork+exec p50 of 0.17 ms.
The config-load phase reads ~20 per-section YAML files on every invocation, even though the PreToolUse handler only consumes a thin slice (security policy, branch-guard flag, gate config). This confirms the SPEC-HOOK-PRETOOL-PERF-001 diagnosis: the config disk cache (M1) + lazy slice (M2) attack the dominant cost.
