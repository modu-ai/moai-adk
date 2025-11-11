# Detailed Skill Analysis Report
Generated: Tue Nov 11 22:17:34 KST 2025

## Critical Issues (Missing SKILL.md)

## moai-document-processing

### File Structure
```
total 0
drwxr-xr-x@   7 goos  staff   224 Nov 11 21:25 .
drwxr-xr-x@ 106 goos  staff  3392 Nov 11 18:33 ..
drwxr-xr-x@   8 goos  staff   256 Nov 11 18:33 docx
drwxr-xr-x@   9 goos  staff   288 Nov 11 21:25 moai-document-processing-unified
drwxr-xr-x@   7 goos  staff   224 Nov 11 18:33 pdf
drwxr-xr-x@   8 goos  staff   256 Nov 11 18:33 pptx
drwxr-xr-x@   5 goos  staff   160 Nov 11 18:33 xlsx
```

❌ **SKILL.md file missing**

---

## Metadata Issues (Sample)

## moai-alfred-agent-guide

### File Structure
```
total 56
drwxr-xr-x@   5 goos  staff    160 Nov 11 12:30 .
drwxr-xr-x@ 106 goos  staff   3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff   1149 Nov 11 12:30 examples.md
-rw-r--r--@   1 goos  staff  18485 Nov 11 12:30 reference.md
-rw-r--r--@   1 goos  staff   2497 Nov 11 12:30 SKILL.md
```

### SKILL.md Preview (first 20 lines)
```
---
name: moai-alfred-agent-guide
description: "19-agent team structure, decision trees for agent selection, Haiku vs Sonnet model selection, and agent collaboration principles. Use when deciding which sub-agent to invoke, understanding team responsibilities, or learning multi-agent orchestration."
allowed-tools: "Read, Glob, Grep"
---

## What It Does

MoAI-ADK의 19개 Sub-agent 아키텍처, 어떤 agent를 선택할지 결정하는 트리, Haiku/Sonnet 모델 선택 기준을 정의합니다.

## When to Use

- ✅ 어떤 sub-agent를 invoke할지 불명확
- ✅ Agent 책임 범위 학습
- ✅ Haiku vs Sonnet 모델 선택 필요
- ✅ Multi-agent 협업 패턴 이해

## Agent Team at a Glance

### 10 Core Sub-agents (Sonnet)
```

**File size:**       70 lines

---

## moai-alfred-code-reviewer

### File Structure
```
total 24
drwxr-xr-x@   5 goos  staff   160 Nov 11 12:30 .
drwxr-xr-x@ 106 goos  staff  3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff  2936 Nov 11 12:30 examples.md
drwxr-xr-x@   3 goos  staff    96 Nov  6 09:42 scripts
-rw-r--r--@   1 goos  staff  7482 Nov 11 12:30 SKILL.md
```

### SKILL.md Preview (first 20 lines)
```
---
name: moai-alfred-code-reviewer
description: "Systematic code review guidance and automation. Apply TRUST 5 principles, check code quality, validate SOLID principles, identify security issues, and ensure maintainability. Use when conducting code reviews, setting review standards, or implementing review automation."
allowed-tools: "Read, Write, Edit, Glob, Bash"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Quality |
| Auto-load | When conducting code reviews or quality checks |

## What It Does

체계적인 코드 리뷰 프로세스와 자동화 가이드를 제공합니다. TRUST 5 원칙 적용, 코드 품질 검증, SOLID 원칙 준수, 보안 이슈 식별, 유지보수성 보장을 다룹니다.

## When to Use

```

**File size:**      212 lines

---

## moai-baas-foundation

### File Structure
```
total 24
drwxr-xr-x@   3 goos  staff     96 Nov 11 12:30 .
drwxr-xr-x@ 106 goos  staff   3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff  11110 Nov 11 12:30 SKILL.md
```

### SKILL.md Preview (first 20 lines)
```
# Skill: moai-baas-foundation

## Metadata

```yaml
skill_id: moai-baas-foundation
skill_name: BaaS Platform Foundation & 9-Platform Decision Framework (Ultra-comprehensive)
version: 2.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["BaaS", "backend-as-a-service", "platform selection", "architecture", "9 platforms", "Convex", "Firebase", "Cloudflare", "Auth0"]
  - contexts: ["/alfred:1-plan", "platform-selection", "architecture-decision", "pattern-a-h"]
agents:
  - spec-builder
  - backend-expert
  - database-expert
  - devops-expert
  - security-expert
```

**File size:**      348 lines

---

## moai-foundation-trust

### File Structure
```
total 192
drwxr-xr-x@   7 goos  staff    224 Nov 11 12:30 .
-rw-r--r--@   1 goos  staff  18667 Nov 11 12:30 .!22407!examples.md
-rw-r--r--@   1 goos  staff  18667 Nov 11 12:30 .!59026!examples.md
drwxr-xr-x@ 106 goos  staff   3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff  19507 Nov  5 21:35 examples.md
-rw-r--r--@   1 goos  staff  22313 Nov 11 12:30 reference.md
-rw-r--r--@   1 goos  staff  10054 Nov 11 12:30 SKILL.md
```

### SKILL.md Preview (first 20 lines)
```
---
name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable). Use when aligning with TRUST governance.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
version: 2.0.0
created: 2025-10-22
---

# Foundation: TRUST Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Version | 2.0.0 |
| Created | 2025-10-22 |
```

**File size:**      307 lines

---

## Missing Supporting Files (Sample)

## moai-cc-agents

### File Structure
```
total 48
drwxr-xr-x@   4 goos  staff    128 Nov 11 12:30 .
drwxr-xr-x@ 106 goos  staff   3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff  21273 Nov 11 21:41 SKILL.md
drwxr-xr-x@   3 goos  staff     96 Nov 11 12:30 templates
```

### SKILL.md Preview (first 20 lines)
```
---
name: moai-cc-agents
version: 4.0.0
created: 2025-11-11
updated: 2025-11-11
status: active
description: AI-powered enterprise Claude Code sub-agents orchestrator with intelligent agent coordination, predictive optimization, ML-based collaboration patterns, and Context7-enhanced workflow automation. Use when creating smart agent systems, implementing AI-driven agent discovery, optimizing agent performance with machine learning, or building enterprise-grade multi-agent architecture with automated governance and coordination.
keywords: ['ai-claude-code-agents', 'enterprise-agent-orchestration', 'predictive-optimization', 'ml-collaboration-patterns', 'context7-workflows', 'intelligent-agent-design', 'automated-governance', 'smart-agents', 'enterprise-architecture']
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# AI-Powered Enterprise Claude Code Sub-Agents Orchestrator v4.0.0

```

**File size:**      580 lines

---

## moai-alfred-workflow

### File Structure
```
total 16
drwxr-xr-x@   3 goos  staff    96 Nov 11 12:30 .
drwxr-xr-x@ 106 goos  staff  3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff  7125 Nov 11 12:30 SKILL.md
```

### SKILL.md Preview (first 20 lines)
```
---
name: moai-alfred-workflow
version: 1.0.0
created: 2025-11-02
updated: 2025-11-02
status: active
description: Guide 4-step workflow execution with task tracking and quality gates
keywords: ['workflow', 'execution', 'planning', 'task-tracking', 'quality']
allowed-tools:
  - Read
---

# Alfred 4-Step Workflow Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-workflow |
| **Version** | 1.0.0 (2025-11-02) |
```

**File size:**      288 lines

---

## moai-artifacts-builder

### File Structure
```
total 64
drwxr-xr-x@   5 goos  staff    160 Nov 11 18:32 .
drwxr-xr-x@ 106 goos  staff   3392 Nov 11 18:33 ..
-rw-r--r--@   1 goos  staff  11357 Nov 11 21:25 LICENSE.txt
drwxr-xr-x@   5 goos  staff    160 Nov 11 18:32 scripts
-rw-r--r--@   1 goos  staff  18941 Nov 11 21:25 SKILL.md
```

### SKILL.md Preview (first 20 lines)
```
---
name: moai-artifacts-builder
description: AI-powered enterprise artifact development orchestrator with Context7 integration, intelligent component generation, automated React/Vue/TypeScript development, modern frontend stack optimization, and enterprise-grade artifact deployment patterns
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
version: 4.0.0 Enterprise
created: 2025-11-11
updated: 2025-11-11
status: active
keywords: ['ai-artifacts-development', 'context7-integration', 'react-vue-typescript', 'frontend-automation', 'enterprise-artifacts', 'intelligent-components', 'modern-webstack', 'artifact-deployment', 'component-intelligence']
---

# AI-Powered Enterprise Artifacts Builder Skill v4.0.0
```

**File size:**      497 lines

---

