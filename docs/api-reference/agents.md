# Agents API Reference

MoAI-ADK 에이전트 시스템의 API 참조 문서입니다.

MoAI-ADK의 에이전트는 Claude Code의 Agent 시스템을 통해 구현됩니다. 각 에이전트는 `.claude/agents/` 디렉토리에 YAML 파일로 정의되며, Python 모듈이 아닌 Claude Code 설정입니다.

자세한 에이전트 사용법은 [Agent Ecosystem Guide](../guides/agents/overview.md)를 참조하세요.

## Agent Types

### spec-builder
SPEC 문서 작성 전문 에이전트

### code-builder
TDD 구현 전문 에이전트

### doc-syncer
문서 동기화 전문 에이전트

### tag-agent
TAG 시스템 관리 에이전트

### git-manager
Git 워크플로우 관리 에이전트

### debug-helper
오류 진단 및 해결 에이전트

### trust-checker
TRUST 5원칙 검증 에이전트

### cc-manager
Claude Code 설정 관리 에이전트

### project-manager
프로젝트 초기화 에이전트
