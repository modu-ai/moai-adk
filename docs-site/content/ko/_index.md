---
title: MoAI-ADK 문서
weight: 99
draft: false
---

MoAI-ADK(Agentic Development Kit)는 Claude Code를 위한 전략적 오케스트레이션
프레임워크입니다.

> **현재 버전:** {{< version >}} — 버전 정보는 `hugo.toml`의 `params.version`을 단일 원천(SSOT)로 삼습니다.

> {{< icon book primary >}} **[클로드 코드로 시작하는 실전 에이전틱 코딩](/book)** — 바이브 코딩의 끝, 하네스 엔지니어링의 시작. MoAI-ADK 메인테이너가 직접 쓴 488쪽 실무 가이드(추천사 9인).

![MoAI-ADK](/og.jpg)

![문서 구조 지도](/images/sections/doc-map-ko.png)

## MoAI 3.0의 세 가지 핵심 가치

- **🪙 토크노믹스** — 컨텍스트 다이어트와 프롬프트 캐싱으로 추론 비용을 60-70% 절감합니다. [멀티 LLM](/multi-llm), [비용 최적화](/cost-optimization), [심화 학습/토크노믹스 개요](/advanced/tokenomics-overview)를 참조하세요.

- **🧠 재귀적 자가 학습** — 의사 결정 메모리와 자율 에이전트 시스템으로 자율 개선 루프를 구현합니다. [자가 진화 시스템](/advanced/self-evolving), [자율 루프](/advanced/autonomous-loops), [의사 결정 메모리](/advanced/decision-memory)를 참조하세요.

- **🛡️ 에이전틱 하네스** — 스킬, 후크, MCP로 구성 가능한 실행 환경으로 확장 가능한 에이전트 오케스트레이션을 제공합니다. [핵심 개념](/core-concepts), [워크플로우 명령어](/workflow-commands), [에이전트 가이드](/advanced/agent-guide)를 참조하세요.

## 주요 기능

- **MoAI Orchestrator**: 전문화된 에이전트를 통한 전략적 작업 위임
- **SPEC 기반 TDD/DDD**: 신규 프로젝트는 TDD, 레거시 코드는 DDD 자동 적용
- **TRUST 5 Framework**: 테스트·가독성·통일성·보안·추적성의 5가지 품질 원칙
- **Progressive Disclosure**: 3단계 스킬 로딩으로 토큰 67% 절감

## 시작하기

MoAI-ADK를 시작하려면 [시작하기](/getting-started) 섹션을 참조하세요.

## 문서 구조

- [시작하기](/getting-started) - 설치, 기본 설정, 빠른 시작
- [핵심 개념](/core-concepts) - SPEC 포맷, 에이전트, 워크플로우
- [심화 학습](/advanced) - 고급 패턴, 스킬 활용, 성능 최적화
- [깃 워크트리](/worktree) - 워크트리 CLI 완벽 가이드
- {{< icon book primary >}} [도서: 실전 에이전틱 코딩](/book) - MoAI-ADK로 배우는 실전 에이전틱 코딩 (488쪽 · 추천사 9인)
