---
id: SPEC-SKILL-001
version: "1.0.0"
status: completed
created: "2026-03-20"
updated: "2026-03-20"
author: GOOS
priority: medium
issue_number: 0
---

# SPEC-SKILL-001: effort frontmatter + worktree sparsePaths 설정

## 1. 개요

Claude Code v2.1.76~v2.1.80에서 추가된 `effort` skill frontmatter 필드와 `worktree.sparsePaths` 설정을 MoAI-ADK 템플릿에 적용한다.

## 2. 구현 요약

### effort frontmatter (v2.1.80)
- `moai-workflow-thinking`: effort: high (심층 분석)
- `moai-foundation-philosopher`: effort: high (전략 분석)
- `moai-workflow-loop`: effort: low (반복 실행, 속도 중시)
- `skill-authoring.md`: effort 필드 문서화

### worktree sparsePaths (v2.1.76)
- `workflow.yaml` 템플릿에 `worktree` 섹션 추가 (SPEC-WORKTREE-002 기반)
- `sparse_paths` 설정 주석으로 예시 제공 (사용자가 프로젝트에 맞게 활성화)
