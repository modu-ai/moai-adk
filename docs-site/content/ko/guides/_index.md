---
title: 가이드
description: MoAI-ADK 운영 가이드 모음 — 자율 CI/CD, GitHub 연동
weight: 85
draft: false
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>소속 가치</strong>: 🪙 토크노믹스 · 🧠 에이전틱 루프 엔지니어링
{{< /callout >}}
<!-- @value: tokenomics, agent-loop -->

MoAI-ADK를 운영할 때 곁에 두면 좋은 가이드를 모았습니다. 자율 CI/CD 가이드는
루프가 스스로 품질을 지킨다는 에이전틱 루프 엔지니어링의 발상을 CI까지 넓히고,
GitHub 연동 가이드는 `moai github` 서브커맨드로 이슈를 파싱해 SPEC과 잇는
가벼운 워크플로우를 다룹니다.


## 가이드 목록

- [자율 CI/CD](./ci-autonomy) — pre-push 훅부터 auto-fix 루프까지, 8-Tier 품질 자동화
- [GitHub 연동](./github-integration) — `moai github` 서브커맨드로 이슈를 파싱하고 SPEC 문서와 연결
