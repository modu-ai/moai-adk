---
title: 가이드
description: MoAI-ADK 운영 가이드 모음 — 자율 CI/CD, GitHub 연동
weight: 85
draft: false
---

MoAI-ADK 운영에 도움이 되는 가이드 문서를 모았습니다. 자율 CI/CD 가이드는
루프가 품질을 지키는(에이전틱 루프 엔지니어링) 아이디어를 CI 환경으로 확장하고,
GitHub 연동 가이드는 `moai github` 서브커맨드로 이슈를 파싱해 SPEC과 연결하는
경량 워크플로우를 다룹니다.


## 가이드 목록

- [자율 CI/CD](./ci-autonomy) — pre-push 훅부터 auto-fix 루프까지, 8-Tier 품질 자동화
- [GitHub 연동](./multi-llm-ci) — `moai github` 서브커맨드로 이슈를 파싱하고 SPEC 문서와 연결
