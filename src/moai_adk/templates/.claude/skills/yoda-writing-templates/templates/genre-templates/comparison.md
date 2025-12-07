# 장르 템플릿: comparison (비교 분석형)

## 📋 용도

두 가지 이상 기술/도구/방법론 비교

**구조**: Overview (10%) → Comparison Table (20%) → Option A Deep Dive (25%) → Option B Deep Dive (25%) → Decision Guide (15%) → Conclusion (5%)

**특징**:
- 객관적 비교
- 장단점 명확 구분
- 의사결정 가이드 제공
- 시나리오별 권장사항

---

## 🏗️ 5가지 구성 요소

### 1. 문서 구조

```
├─ Overview (10%, 비교 대상 소개)
│  ├─ 비교 대상 A/B 소개
│  └─ 비교 필요성
│
├─ Comparison Table (20%, 기능별 비교)
│  ├─ 주요 기능 비교
│  ├─ O/X 또는 점수 표시
│  └─ 한눈에 보는 차이점
│
├─ Option A Deep Dive (25%, 상세 분석)
│  ├─ 장점 (3-5개)
│  ├─ 단점 (2-3개)
│  ├─ 사용 시나리오
│  └─ 코드 예제
│
├─ Option B Deep Dive (25%, 상세 분석)
│  ├─ 장점 (3-5개)
│  ├─ 단점 (2-3개)
│  ├─ 사용 시나리오
│  └─ 코드 예제
│
├─ Decision Guide (15%, 선택 가이드)
│  ├─ "Choose A when..."
│  ├─ "Choose B when..."
│  └─ 의사결정 플로우차트
│
└─ Conclusion (5%, 최종 권장사항)
   └─ 종합 정리 + 추천

총 섹션: 6개
비율 검증: 10+20+25+25+15+5 = 100%
```

### 2. 문체

**어조**: 객관적이고 균형 잡힌 분석

**종결어미**: "-다", "-니다" (일관성 유지)

**능동태**: 90% 이상

**문장 길이**: 평균 18-22단어

**특징적 패턴**:
- "A는 X에 강하지만, B는 Y에 더 적합합니다"
- "Choose A when..." (선택 기준)
- "On the other hand..." (대조)

### 3. 내용 전개

**배치 순서**:
1. **Overview** (맥락 설정)
   - 비교 대상 소개
   - 왜 비교가 필요한가

2. **Comparison Table** (시각적 비교)
   - 주요 특징 5-8개
   - O/X 또는 별점 표시
   - 한눈에 파악 가능

3. **Option A Deep Dive** (심층 분석)
   - 장점 (긍정적 특징)
   - 단점 (제한 사항)
   - 적합한 사용 사례
   - 코드 예제 (10-15줄)

4. **Option B Deep Dive** (심층 분석)
   - 동일한 구조로 분석
   - A와 대조되는 특징 강조
   - 코드 예제 (10-15줄)

5. **Decision Guide** (선택 기준)
   - "Choose A when..."
   - "Choose B when..."
   - 의사결정 플로우차트 (선택)

6. **Conclusion** (종합 정리)
   - 요약
   - 개인적 추천 (객관성 유지)

### 4. 조건

**글자 수**: 2,000-2,500자

**비교 테이블**: 필수 (5-8개 항목)

**코드 예제**: 2개 (각 옵션당 1개)

**필수 요소**:
- 객관적 장단점 분석
- 구체적 사용 시나리오
- Decision Guide의 명확한 기준

**금지 사항**:
- 주관적 편향 ("A가 더 좋다")
- 불명확한 기준 ("상황에 따라")
- 장점만 나열 (단점 누락)

### 5. 형식

```markdown
# LangChain vs Google ADK: 에이전트 프레임워크 비교

## Overview

LangChain과 Google ADK는 모두 LLM 기반 애플리케이션 개발을 위한 프레임워크입니다. LangChain은 파이썬/자바스크립트 생태계에서 가장 널리 사용되는 오픈소스 프레임워크이며, Google ADK는 Google의 Gemini 모델에 최적화된 경량 프레임워크입니다. 프로젝트의 규모, 팀의 경험, 요구사항에 따라 적합한 선택이 달라집니다.

## Comparison Table

| 기능 | LangChain | Google ADK |
|------|-----------|------------|
| **학습 곡선** | 가파름 (⭐⭐⭐⭐) | 완만함 (⭐⭐) |
| **문서 품질** | 풍부 (⭐⭐⭐⭐⭐) | 제한적 (⭐⭐⭐) |
| **모델 지원** | 다중 (20+) | Gemini 중심 |
| **커뮤니티 크기** | 매우 큼 | 성장 중 |
| **병렬 처리** | RunnableParallel | ParallelAgent |
| **MCP 통합** | ✅ | ✅ |
| **프로덕션 준비도** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **성능** | 보통 | 빠름 (Gemini) |

## LangChain Deep Dive

### 장점

**1. 광범위한 생태계**: 20개 이상의 LLM 모델, 100개 이상의 도구 통합을 지원합니다. OpenAI, Anthropic, Cohere 등 모든 주요 모델을 단일 인터페이스로 사용할 수 있습니다 (1).

**2. 성숙한 프레임워크**: 2022년부터 개발되어 프로덕션 환경에서 검증되었습니다. 대부분의 엣지 케이스가 처리되어 있습니다.

**3. 풍부한 문서와 커뮤니티**: 공식 문서, 튜토리얼, 커뮤니티 예제가 매우 풍부합니다. Stack Overflow와 Discord에서 활발한 지원을 받을 수 있습니다.

**4. LCEL (LangChain Expression Language)**: 체인을 선언적으로 구성할 수 있는 강력한 DSL을 제공합니다.

### 단점

**1. 가파른 학습 곡선**: 개념(Chain, Agent, Tool, Memory 등)이 많아 초기 학습에 시간이 필요합니다.

**2. 과도한 추상화**: 간단한 작업에도 많은 보일러플레이트 코드가 필요할 수 있습니다.

**3. 버전 호환성 문제**: 빠른 개발로 인해 버전 간 breaking change가 자주 발생합니다.

### 사용 시나리오

- 다중 LLM 모델을 지원해야 하는 경우
- 복잡한 RAG 파이프라인이 필요한 경우
- 기존 LangChain 생태계의 도구를 활용하려는 경우

### 코드 예제

\`\`\`python
from langchain.agents import AgentExecutor
from langchain_core.runnables import RunnableParallel

# 병렬 처리 체인
parallel_chain = RunnableParallel({
    "research1": research_chain_1,
    "research2": research_chain_2,
    "research3": research_chain_3
})

# 실행
results = parallel_chain.invoke({"query": "..."})
\`\`\`

## Google ADK Deep Dive

### 장점

**1. 단순성**: 최소한의 개념(Agent, Tool)만으로 시작할 수 있습니다. 학습 곡선이 완만합니다 (2).

**2. Gemini 최적화**: Google의 Gemini 모델에 최적화되어 빠른 응답과 낮은 레이턴시를 제공합니다.

**3. 경량 프레임워크**: 핵심 기능에 집중하여 의존성이 적고 설치가 간단합니다.

**4. 명확한 에이전트 패턴**: ParallelAgent, SequentialAgent 등 직관적인 이름으로 패턴을 표현합니다.

### 단점

**1. 제한적인 모델 지원**: 주로 Gemini 모델에 최적화되어 있어 다른 LLM 통합이 제한적입니다.

**2. 작은 커뮤니티**: 상대적으로 신생 프레임워크라 커뮤니티 리소스가 부족합니다.

**3. 문서 한계**: LangChain에 비해 문서와 예제가 적습니다.

### 사용 시나리오

- Gemini 모델만 사용하는 경우
- 빠른 프로토타이핑이 필요한 경우
- 단순하고 명확한 코드베이스를 선호하는 경우

### 코드 예제

\`\`\`python
from adk import ParallelAgent, SequentialAgent, LlmAgent

# 병렬 에이전트
parallel_agent = ParallelAgent(
    name="ResearchTeam",
    sub_agents=[agent_1, agent_2, agent_3]
)

# 순차 실행
pipeline = SequentialAgent(
    sub_agents=[parallel_agent, merger_agent]
)

# 실행
result = pipeline.run()
\`\`\`

## Decision Guide

### Choose LangChain when:

✅ 다중 LLM 모델 지원이 필요한 경우
✅ 복잡한 RAG 파이프라인을 구축하는 경우
✅ 풍부한 도구 생태계를 활용하려는 경우
✅ 프로덕션 안정성이 최우선인 경우
✅ 팀에 LangChain 경험자가 있는 경우

### Choose Google ADK when:

✅ Gemini 모델만 사용하는 경우
✅ 빠른 프로토타이핑이 필요한 경우
✅ 단순하고 유지보수 쉬운 코드를 원하는 경우
✅ 성능과 낮은 레이턴시가 중요한 경우
✅ 학습 곡선을 최소화하려는 경우

### 의사결정 플로우

\`\`\`mermaid
graph TD
    Start{다중 LLM<br/>필요?} -->|Yes| LangChain
    Start -->|No| Gemini{Gemini만<br/>사용?}
    Gemini -->|Yes| Simple{단순한<br/>구조?}
    Gemini -->|No| LangChain
    Simple -->|Yes| ADK[Google ADK]
    Simple -->|No| LangChain[LangChain]
\`\`\`

## Conclusion

LangChain은 프로덕션 환경에서 검증된 성숙한 프레임워크로, 다양한 LLM과 도구를 지원합니다. 복잡한 요구사항과 장기 프로젝트에 적합합니다. 반면 Google ADK는 Gemini에 최적화된 경량 프레임워크로, 빠른 개발과 단순성을 제공합니다. 프로토타이핑이나 Gemini 전용 프로젝트에 권장됩니다.

대부분의 경우 LangChain을 선택하는 것이 안전하지만, Gemini만 사용하고 빠른 개발이 필요하다면 Google ADK를 고려해보세요. 두 프레임워크를 함께 사용하는 것도 가능합니다 (LangChain에서 ADK 에이전트를 도구로 래핑).

## References

[1] LangChain 공식 문서: https://python.langchain.com/
[2] Google ADK 문서: https://google.github.io/adk-docs/
```

---

## 📝 작성 프롬프트 템플릿

```
비교 분석형 섹션을 작성하세요.

**구조** (비율 준수):
1. Overview (10%) - 비교 대상 소개
2. Comparison Table (20%) - 5-8개 항목 비교표
3. Option A Deep Dive (25%)
   - 장점 3-5개
   - 단점 2-3개
   - 사용 시나리오
   - 코드 예제 (10-15줄)
4. Option B Deep Dive (25%) - 동일 구조
5. Decision Guide (15%)
   - "Choose A when..."
   - "Choose B when..."
   - 의사결정 플로우차트 (선택)
6. Conclusion (5%) - 종합 정리

**글자 수**: 2,000-2,500자

**필수 요소**:
- 객관적 분석 (편향 없이)
- 구체적 사용 시나리오
- 명확한 선택 기준
- 비교 테이블

비교 대상: [A vs B]
```

---

**마지막 수정**: 2025-11-25
**버전**: 1.0.0
**참조**: 기술 비교 문서 표준
