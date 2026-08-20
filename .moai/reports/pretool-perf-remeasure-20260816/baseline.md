# Profiling Baseline — SPEC-HOOK-PRETOOL-PERF-001 M0

> Pre-change baseline. Captured under simulated concurrent-hook stress.

## Measurement configuration

- **Parallelism**: 8 parallel `moai hook pre-tool` invocations per batch
- **Batches**: 5
- **Total invocations**: 40
- **Platform**: darwin/arm64
- **Timestamp**: 2026-08-16T12:53:40Z

## Per-phase timing (ms)

| Phase | p50 | p99 | max |
|-------|-----|-----|-----|
| External wall-time | 251.10 | 1256.28 | 1256.28 |
| Fork+exec | 0.30 | 82.96 | 82.96 |
| Config load | 1.75 | 102.81 | 102.81 |
| Dispatch (security scan) | 0.01 | 0.05 | 0.05 |
| Internal total | 4.36 | 109.39 | 109.39 |

## Diagnosis

Config load is the dominant per-invocation cost under concurrent stress: config-load p50 is 1.75 ms vs dispatch p50 of 0.01 ms and fork+exec p50 of 0.30 ms.
The config-load phase reads ~20 per-section YAML files on every invocation, even though the PreToolUse handler only consumes a thin slice (security policy, branch-guard flag, gate config). This confirms the SPEC-HOOK-PRETOOL-PERF-001 diagnosis: the config disk cache (M1) + lazy slice (M2) attack the dominant cost.
