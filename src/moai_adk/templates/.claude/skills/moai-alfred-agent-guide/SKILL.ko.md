---
name: moai-alfred-agent-guide-ko
description: "19명의 에이전트 팀 구조, 에이전트 선택 결정 트리, Haiku vs Sonnet 모델 선택, 에이전트 협업 원칙. 어떤 서브-에이전트를 호출할지, 팀 책임 범위를 이해하거나, 멀티-에이전트 오케스트레이션을 학습할 때 사용합니다."
allowed-tools: "Read, Glob, Grep"
---

## 기능

MoAI-ADK의 19개 Sub-agent 아키텍처, 어떤 agent를 선택할지 결정하는 트리, Haiku/Sonnet 모델 선택 기준을 정의합니다.

## 사용 시점

- ✅ 어떤 sub-agent를 invoke할지 불명확할 때
- ✅ Agent 책임 범위를 학습할 때
- ✅ Haiku vs Sonnet 모델 선택이 필요할 때
- ✅ Multi-agent 협업 패턴을 이해할 때

## 팀 구조 요약

### 10개 핵심 Sub-agent (Sonnet)
- spec-builder: SPEC 작성
- tdd-implementer: TDD 구현 (RED → GREEN → REFACTOR)
- doc-syncer: 문서 동기화
- implementation-planner: 구현 전략
- debug-helper: 오류 분석
- quality-gate: TRUST 5 검증
- tag-agent: TAG 체인 검증
- git-manager: Git 워크플로우
- Explore: 코드베이스 탐색
- Plan: 작업 계획

### 4개 전문가 에이전트 (Sonnet - 자동 활성화)
- **backend-expert**: Backend 아키텍처, API 설계, 데이터베이스
- **frontend-expert**: Frontend 아키텍처, 컴포넌트 설계, 상태 관리
- **devops-expert**: DevOps 전략, 배포, 인프라
- **ui-ux-expert**: UI/UX 설계, 접근성, 디자인 시스템 (Figma MCP)

### 6개 전문가 에이전트 (Haiku)
- project-manager: 프로젝트 초기화
- skill-factory: Skill 생성/최적화
- cc-manager: Claude Code 설정
- cc-hooks: Hook 시스템
- cc-mcp-plugins: MCP 서버
- trust-checker: TRUST 검증

## 에이전트 선택 결정 트리

```
작업 유형?
├─ SPEC 작성/검증 → spec-builder
├─ TDD 구현 → tdd-implementer
├─ 문서 동기화 → doc-syncer
├─ 구현 계획 → implementation-planner
├─ 오류 분석 → debug-helper
├─ 품질 검증 → quality-gate + Skill("moai-foundation-trust")
├─ 코드베이스 탐색 → Explore
├─ Git 워크플로우 → git-manager
└─ 전체 프로젝트 계획 → Plan
```

## 모델 선택

- **Sonnet**: 복잡한 추론 작업 (spec-builder, tdd-implementer, implementation-planner)
- **Haiku**: 빠른 실행 (project-manager, quality-gate, git-manager)

---

자세한 내용은 `reference.md`에서 완전한 에이전트 책임, 협업 패턴, 고급 오케스트레이션 전략을 확인하세요.

**관련 스킬**: moai-alfred-rules, moai-alfred-practices