---
id: SPEC-CURATION-001
title: "Tool Curation & MCP Segregation"
status: draft
priority: P1
created: "2026-04-07"
harness_pillar: "P1: Guardrails (SOLID I-Principle)"
---

# SPEC-CURATION-001: Tool Curation & MCP Segregation

## Overview

SOLID Interface Segregation 원칙에 따라 에이전트별 도구를 최소화하고 MCP 서버 노출을 명시적으로 제한.

## Requirements (EARS Format)

### REQ-CUR-001 (Ubiquitous)
settings.json.tmpl의 `enableAllProjectMcpServers` SHALL be set to `false`.

### REQ-CUR-002 (Ubiquitous)
manager-quality의 permissionMode SHALL be `acceptEdits` (not `bypassPermissions`).

### REQ-CUR-003 (Ubiquitous)
manager-git의 tools SHALL NOT include `mcp__sequential-thinking__sequentialthinking` and `mcp__context7__*`.

### REQ-CUR-004 (Ubiquitous)
expert-refactoring의 tools SHALL NOT include `Agent` (leaf-node specialist).

### REQ-CUR-005 (Ubiquitous)
expert-debug의 tools SHALL NOT include `Agent` (leaf-node specialist).

## Implementation Scope

### Modified Files (Templates)
- `internal/template/templates/.claude/settings.json.tmpl` — enableAllProjectMcpServers: false
- `internal/template/templates/.claude/agents/moai/manager-quality.md` — permissionMode change
- `internal/template/templates/.claude/agents/moai/manager-git.md` — tools reduction
- `internal/template/templates/.claude/agents/moai/expert-refactoring.md` — remove Agent
- `internal/template/templates/.claude/agents/moai/expert-debug.md` — remove Agent

### Modified Files (Local)
- Same files under `.claude/agents/moai/` for local project
