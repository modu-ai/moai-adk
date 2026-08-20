---
title: MoAI-ADK 문서
weight: 99
draft: false
---

MoAI-ADK(Agentic Development Kit)는 Claude Code를 위한 전략적 오케스트레이션
프레임워크입니다.

> **현재 버전:** {{< version >}} — 버전 정보는 `hugo.toml`의 `params.version`을 단일 원천(SSOT)로 삼습니다.

> {{< icon book primary >}} **[클로드 코드로 시작하는 실전 에이전틱 코딩](https://book.mo.ai.kr)** — 바이브 코딩의 끝, 하네스 엔지니어링의 시작. MoAI-ADK 메인테이너가 직접 쓴 488쪽 실무 가이드(추천사 9인).

![MoAI-ADK](/og.jpg)

![문서 구조 지도](/images/sections/doc-map-ko.png)

## v3.1 새 기능 — 칸반 모드 {{< new-badge v3.1 >}}

세션 하나는 컨텍스트 창 하나를 쓴다. 긴 SPEC은 그 창을 채우고, 뒤에 오는 작업은 앞의 것을 전부 지고 간다. 칸반 모드는 작업 하나를 **터미널 네 개**로 나눈다 — 리드 세션이 체인을 몰고, 세 개의 동반 세션이 `plan`·`run`·`sync` 한 칸씩을 맡아 자기 칸의 맥락만 진다. 검토 판정은 별도 칸이 아니라 sync 게이트가 흡수한다. 한도가 사라지는 것은 아니지만, 어느 세션도 세 단계치 이력을 짊어지지 않으므로 같은 예산이 훨씬 멀리 간다.

![칸반 모드 한 런 — 다섯 칸 보드와 리드·세 동반 세션이 각자의 터미널에서, 각자의 모델과 추론 강도로 돌고 있다](/images/profile/kanban-five-sessions.png)

칸마다 백엔드와 추론 강도를 다르게 둘 수 있다. 위 화면은 Plan을 Opus 5 high로, Run을 GLM 5.2 xhigh로, Sync를 GLM 5.2로 돌린다.

{{< terminal title="kanban mode" raw="true" >}}
moai cc -k                    # 리드 — run-id를 알려주고 체인을 깐다
moai cc -k --name plan        # 동반 세션, 각자 별도 터미널에서
moai cc -k --name run
moai cc -k --name sync
{{< /terminal >}}

보드는 `backlog → plan → run → sync → done` 다섯 칸이고, `backlog`에는 주인 세션이 일부러 없다. 그래서 일감은 [`/moai todo`](/ko/utility-commands/moai-todo)로 사람이 넣을 때만 보드에 들어온다. review 칸은 없다 — 검토 판정은 sync 게이트가 흡수한다. 리드는 카드의 `progress.md`에서 직접 읽은 증거로만 카드를 넘긴다 — 동반 세션의 답장으로는 넘기지 않는다.

`moai web`을 띄우면 칸반 화면에서 칸반 체인과 SPEC 파이프라인을 함께 볼 수 있다.

![moai web 콘솔 Overview 화면 — SPEC 집계, 진행 중 SPEC 목록, 세션 레지스트리](/images/profile/web-console-v31-overview.png)

자세히: [칸반 모드](/ko/advanced/kanban-mode) · [manager-lead 리드 코디네이터](/ko/advanced/manager-lead) · [`/moai todo`](/ko/utility-commands/moai-todo) · [moai web 콘솔](/ko/advanced/moai-web-console)

## MoAI 3.1의 세 가지 핵심 가치

- {{< icon database primary >}} **토크노믹스** — 컨텍스트 다이어트와 프롬프트 캐싱으로 추론 비용을 60-70% 절감합니다. [멀티 LLM](/ko/multi-llm), [비용 최적화](/ko/cost-optimization), [심화 학습/토크노믹스 개요](/ko/advanced/tokenomics-overview)를 참조하세요.

- {{< icon rotate primary >}} **에이전틱 루프 엔지니어링** — 루프가 스스로 일하고, 그렇게 쌓인 관찰이 하네스 지침을 다시 다듬는 자율 개선 사이클입니다(재귀적 자가 학습). [자가 진화 시스템](/ko/advanced/self-evolving), [자율 루프](/ko/advanced/autonomous-loops), [의사 결정 메모리](/ko/advanced/decision-memory)를 참조하세요.

- {{< icon package primary >}} **에이전틱 하네스** — 스킬·후크·MCP를 조합해 실행 환경을 직접 짜고, 에이전트 오케스트레이션을 필요한 만큼 넓힙니다. [핵심 개념](/ko/core-concepts), [워크플로우 명령어](/ko/workflow-commands), [에이전트 가이드](/ko/advanced/agent-guide)를 참조하세요.

## 주요 기능

- **MoAI Orchestrator**: 전문 에이전트에게 작업을 전략적으로 나눠 맡김
- **SPEC 기반 TDD/DDD**: 신규 프로젝트는 TDD, 레거시 코드는 DDD 자동 적용
- **TRUST 5 Framework**: 테스트·가독성·통일성·보안·추적성의 5가지 품질 원칙
- **Progressive Disclosure**: 3단계 스킬 로딩으로 토큰 67% 절감

## 시작하기

MoAI-ADK를 시작하려면 [시작하기](/ko/getting-started) 섹션을 참조하세요.

## 문서 구조

- [시작하기](/ko/getting-started) - 설치, 기본 설정, 빠른 시작
- [핵심 개념](/ko/core-concepts) - SPEC 포맷, 에이전트, 워크플로우
- [심화 학습](/ko/advanced) - 고급 패턴, 스킬 활용, 성능 최적화
- [깃 워크트리](/ko/worktree) - 워크트리 CLI 완벽 가이드
- {{< icon book primary >}} [도서: 실전 에이전틱 코딩](https://book.mo.ai.kr) - MoAI-ADK로 배우는 실전 에이전틱 코딩 (488쪽 · 추천사 9인)
