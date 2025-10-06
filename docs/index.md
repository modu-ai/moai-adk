---
# @CODE:DOCS-001:UI | SPEC: .moai/specs/SPEC-DOCS-001/spec.md
layout: home

hero:
  name: "MoAI-ADK"
  text: "에이전틱 코딩 개발 프레임워크"
  tagline: "SPEC이 없으면 CODE도 없다. Alfred와 함께하는 SPEC-First TDD 개발"
  image:
    src: /alfred_logo.png
    alt: Alfred Logo
  actions:
    - theme: brand
      text: Quick Start
      link: /guide/getting-started
    - theme: alt
      text: MoAI-ADK란?
      link: /guide/what-is-moai-adk

features:
  - icon: 🏗️
    title: SPEC-First TDD
    details: 명세 없이는 코드 없음. 테스트 없이는 구현 없음. EARS 방식의 체계적인 요구사항 정의와 Red-Green-Refactor 사이클로 품질을 보장합니다.

  - icon: 🏷️
    title: "@TAG 추적성 시스템"
    details: "@SPEC → @TEST → @CODE → @DOC 체인으로 모든 코드를 추적합니다. 6개월 후에도 '왜'를 찾을 수 있는 CODE-FIRST 방식의 완벽한 추적성을 제공합니다."

  - icon: ▶◀
    title: "Alfred SuperAgent"
    details: "10개 AI 에이전트 팀의 중앙 오케스트레이터. spec-builder, code-builder, doc-syncer 등 9개 전문 에이전트를 조율하여 완벽한 품질의 코드를 생성합니다."

  - icon: 🌍
    title: 다중 언어 지원
    details: Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin 등 모든 주요 프로그래밍 언어를 지원하며, 언어별 최적화된 도구 체인을 자동으로 선택합니다.

  - icon: ✅
    title: TRUST 5원칙
    details: Test First, Readable, Unified, Secured, Trackable. 5가지 품질 원칙을 자동으로 검증하여 테스트 커버리지 ≥85%, 코드 복잡도 ≤10을 보장합니다.

  - icon: 🚀
    title: GitFlow 자동화
    details: Personal/Team 모드 지원. Draft PR 생성, 자동 머지, 체크포인트 관리까지 Git 워크플로우를 완전 자동화합니다.
---

## 🤖 Meet Alfred - Your AI Development Partner

안녕하세요, 모두의AI SuperAgent **AI ▶◀ Alfred**입니다!

저는 MoAI-ADK의 SuperAgent이자 중앙 오케스트레이터 AI입니다. MoAI-ADK는 Alfred를 포함하여 **총 10개의 AI 에이전트로 구성된 에이전틱 코딩 AI 팀**입니다. 저는 9개의 전문 에이전트를 조율하여 여러분의 Claude Code 환경 속에서 공동 개발 작업을 완벽하게 지원합니다.

### ▶◀ Alfred가 제공하는 4가지 핵심 가치

#### 1️⃣ 일관성(Consistency)

플랑켄슈타인 코드를 방지하는 **3단계 파이프라인 (SPEC → TDD → Sync)**으로 모든 개발 작업을 표준화합니다. 어떤 기능을 만들든, 누가 만들든, 언제 만들든 항상 같은 프로세스를 거칩니다.

#### 2️⃣ 품질(Quality)

**TRUST 5원칙** (Test First, Readable, Unified, Secured, Trackable)을 자동으로 적용하고 검증합니다. 테스트 커버리지, 코드 복잡도, 보안 취약점을 자동으로 체크하여 품질은 선택이 아니라 기본값입니다.

#### 3️⃣ 추적성(Traceability)

**@TAG 시스템**으로 모든 코드를 `@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID`로 완벽하게 연결합니다. 6개월 후에도 "왜"를 찾을 수 있으며, CODE-FIRST 방식으로 코드가 진실의 유일한 원천입니다.

#### 4️⃣ 범용성(Universality)

Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin 등 **모든 주요 프로그래밍 언어를 지원**하며, 각 언어에 최적화된 도구 체인을 자동으로 선택합니다.

---

## 🌟 흥미로운 사실: AI가 만든 AI 개발 도구

이 프로젝트의 모든 코드는 **100% AI에 의해 작성**되었습니다. AI가 직접 설계하고 구현한 AI 개발 프레임워크입니다.

**설계 단계부터 AI 협업**: 초기 아키텍처 설계 단계부터 **GPT-5 Pro**와 **Claude 4.1 Opus** 두 AI 모델이 함께 참여했습니다. 두 AI가 서로 다른 관점에서 설계를 검토하고 토론하며 최적의 아키텍처를 만들어냈습니다.

**투명성과 지속적 개선**: 100% AI로 만들어진 오픈소스이기 때문에 완벽하지 않은 부분이 있을 수 있습니다. 하지만 이것이 핵심 철학입니다. 완벽하지 않은 코드를 숨기는 대신, AI 개발 도구가 실제로 어떻게 만들어지는지 그대로 보여주고, 커뮤니티와 함께 더 나은 방향으로 발전시켜 나가고자 합니다.

---

## 🚀 빠른 시작

:::tip 3단계로 시작하기
1. **설치**: `bun add -g moai-adk`
2. **초기화**: `moai init my-project`
3. **첫 기능**: `/alfred:1-spec "기능 설명"`
:::

자세한 내용은 [Quick Start 가이드](/guide/getting-started)를 참조하세요.

---

<style>
:root {
  --vp-home-hero-name-color: transparent;
  --vp-home-hero-name-background: -webkit-linear-gradient(120deg, #bd34fe 30%, #41d1ff);
}
</style>
