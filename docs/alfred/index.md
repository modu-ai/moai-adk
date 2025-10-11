# Alfred SuperAgent

▶◀ **Alfred**: MoAI-ADK의 공식 SuperAgent

## 목차

- [Overview](overview.md) - Alfred 개요
- [Commands](commands.md) - 커맨드 가이드
- [Orchestration](orchestration.md) - 오케스트레이션 전략

## Alfred란?

Alfred는 MoAI-ADK의 중앙 오케스트레이터로, 9개의 전문 에이전트를 조율합니다.

## 핵심 커맨드

- `/alfred:1-spec` - SPEC 작성
- `/alfred:2-build` - TDD 구현
- `/alfred:3-sync` - 문서 동기화
- `/alfred:8-project` - 프로젝트 초기화

## 9개 전문 에이전트

| 에이전트 | 전문 영역 | 페르소나 |
|---------|----------|---------|
| spec-builder | SPEC 작성 | 시스템 아키텍트 |
| code-builder | TDD 구현 | 수석 개발자 |
| doc-syncer | 문서 동기화 | 테크니컬 라이터 |
| tag-agent | TAG 시스템 | 지식 관리자 |
| git-manager | Git 워크플로우 | 릴리스 엔지니어 |
| debug-helper | 오류 진단 | 트러블슈팅 전문가 |
| trust-checker | TRUST 검증 | 품질 보증 리드 |
| cc-manager | Claude Code 설정 | 데브옵스 엔지니어 |
| project-manager | 프로젝트 초기화 | 프로젝트 매니저 |
