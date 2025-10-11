# Agents

MoAI-ADK 에이전트 시스템입니다.

## 목차

- [Architecture](architecture.md) - 에이전트 아키텍처
- [Creating Agents](creating-agents.md) - 커스텀 에이전트 작성
- [Agent Collaboration](agent-collaboration.md) - 에이전트 협업 원칙

## 에이전트 생태계

Alfred SuperAgent가 9개의 전문 에이전트를 조율합니다.

### 핵심 개발 에이전트

- **spec-builder** (🏗️): SPEC 작성 전문가
- **code-builder** (💎): TDD 구현 전문가
- **doc-syncer** (📖): 문서 동기화 전문가

### 지원 에이전트

- **tag-agent** (🏷️): TAG 시스템 관리
- **git-manager** (🚀): Git 워크플로우 자동화
- **debug-helper** (🔬): 디버깅 및 트러블슈팅
- **trust-checker** (✅): 품질 보증
- **cc-manager** (🛠️): Claude Code 설정
- **project-manager** (📋): 프로젝트 초기화

## 에이전트 협업 원칙

- **단일 책임**: 각 에이전트는 자신의 전문 영역만 담당
- **중앙 조율**: Alfred만이 에이전트 간 작업 조율
- **품질 게이트**: 각 단계 완료 시 자동 검증
