---
title: 가이드
description: MoAI-ADK 운영 가이드 모음 — 자율 CI/CD, 멀티 LLM CI
weight: 85
draft: false
---

MoAI-ADK 운영에 도움이 되는 가이드 문서를 모았습니다. 두 문서 모두 v3.0의
핵심 아이디어 — 루프가 품질을 지키고(에이전틱 루프 엔지니어링), 모델 배정이
비용을 지킨다(토크노믹스) — 를 CI 환경으로 확장한 내용입니다.

## 가이드 목록

- [자율 CI/CD](./ci-autonomy) — pre-push 훅부터 auto-fix 루프까지, 8-Tier 품질 자동화
- [멀티 LLM CI](./multi-llm-ci) — GitHub Actions에서 여러 AI 모델로 코드 리뷰 자동화
