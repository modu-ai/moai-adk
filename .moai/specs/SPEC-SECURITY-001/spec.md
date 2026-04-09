---
id: SPEC-SECURITY-001
title: "Security Hardening"
status: draft
priority: P2
created: "2026-04-07"
harness_pillar: "P1: Guardrails"
---

# SPEC-SECURITY-001: Security Hardening

## Overview

DefaultSecurityPolicy()의 하드코딩된 보안 패턴을 security.yaml 외부 설정으로 이동하고
위험 명령어 블록리스트를 확장하여 SOLID O원칙(확장 가능성) 준수.

## Requirements (EARS Format)

### REQ-SEC-001 (Ubiquitous)
`.moai/config/sections/security.yaml` SHALL define additional dangerous Bash patterns that extend (not replace) the hardcoded defaults.

### REQ-SEC-002 (Ubiquitous)
The security config SHALL support `extra_deny_patterns`, `extra_ask_patterns`, and `extra_dangerous_bash_patterns` lists.

### REQ-SEC-003 (Event-Driven)
When security.yaml is loaded, the extra patterns SHALL be appended to the DefaultSecurityPolicy patterns.

### REQ-SEC-004 (Ubiquitous)
The default hardcoded patterns SHALL NOT be removed (defense in depth — config extends, never replaces).

### REQ-SEC-005 (Ubiquitous)
The extra_dangerous_bash_patterns SHALL include: `curl.*\|.*sh`, `wget.*\|.*sh`, `chmod\s+777`, `rm\s+-rf\s+/[^.]` (targeting root paths).

## Implementation Scope

### New Files
- `internal/template/templates/.moai/config/sections/security.yaml` — Extra security patterns
- `internal/hook/security/config.go` — Load and merge security config
- `internal/hook/security/config_test.go` — Tests

### Modified Files
- `internal/hook/pre_tool.go` — Load extra patterns from config
- `internal/cli/deps.go` — Wire security config loading
