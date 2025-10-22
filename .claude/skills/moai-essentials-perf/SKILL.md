---
name: moai-essentials-perf
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Performance profiling, bottleneck detection, and optimization strategies across languages.
keywords: ['performance', 'profiling', 'optimization', 'bottleneck']
allowed-tools:
  - Read
  - Bash
---

# Lang Performance Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-essentials-perf |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

Performance profiling, bottleneck detection, and optimization strategies across languages.

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ TDD workflow support

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Intel VTune** | 2025.0 | Profiling | ✅ Current |
| **perf** | 6.0 | Linux Profiler | ✅ Current |
| **Prometheus** | 2.50.0 | Monitoring | ✅ Current |
| **Grafana** | 10.0.0 | Visualization | ✅ Current |

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## References (Latest Documentation)

_Documentation links updated 2025-10-22_

---

## Changelog

- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)

---

## Best Practices

✅ **DO**:
- Profile before optimizing
- Focus on hot paths (top 10% functions)
- Measure impact of changes
- Use appropriate profiling tools

❌ **DON'T**:
- Optimize without profiling
- Ignore I/O bottlenecks
- Sacrifice readability for micro-optimizations
- Skip performance regression tests
