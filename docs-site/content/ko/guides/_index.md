---
title: 가이드
description: MoAI-ADK 운영 가이드 모음 — 자율 CI/CD, GitHub 연동
weight: 85
draft: false
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>소속 가치</strong>: 토크노믹스 · 에이전틱 루프 엔지니어링
{{< /callout >}}
<!-- @value: tokenomics, agent-loop -->

MoAI-ADK를 설치하고 나면 곧바로 부딪히는 다음 질문이 있습니다. "저장소에
push하면 무슨 일이 벌어지나?" 이 섹션은 그 질문에 답하는 운영 가이드를
모아둔 곳입니다. CLI 커맨드 사전도, 개념 해설도 아닙니다 — 핵심 개념 문서를
이미 읽은 독자가 실전에서 다음 걸음을 내딛기 위해 찾는 위치입니다. 그래서
위에서 아래로 순서대로 읽기보다는, 저장소 상황에 맞춰 필요한 쪽부터 펼쳐
보는 것이 어울립니다.

두 가이드가 다루는 대상은 다르지만 한 가지 공통된 움직임이 있습니다. 둘 다
"반복 루프"를 사람 머리에서 코드로 옮깁니다. 자율 CI/CD 가이드는 로컬
세션에서 돌리던 "진단 → 수정 → 검증" 루프를 GitHub CI까지 확장하고,
GitHub 연동 가이드는 `moai github` 서브커맨드로 이슈 본문을 파싱해
SPEC(요구사항 정의 문서)과 잇는 가벼운 워크플로우를 다룹니다. 그래서 두
가이드 모두 에이전틱 루프 엔지니어링(AI 에이전트가 스스로 반복하는 작업
루프를 설계하는 기법)을 저장소 단위로 적용한 사례로 읽을 수 있고, 토큰을
아끼는 패턴도 함께 관찰할 수 있습니다.


## 가이드 목록

- [자율 CI/CD](./ci-autonomy) — pre-push 훅부터 auto-fix 루프까지, 8-Tier 품질 자동화
- [GitHub 연동](./github-integration) — `moai github` 서브커맨드로 이슈를 파싱하고 SPEC 문서와 연결
