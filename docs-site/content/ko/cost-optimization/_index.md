---
title: 비용 최적화
weight: 70
draft: false
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>소속 가치</strong>: 🪙 토크노믹스
{{< /callout >}}
<!-- @value: tokenomics -->

MoAI-ADK 토크노믹스는 두 갈래로 추론 비용을 줄입니다. **컨텍스트
다이어트**가 항상 로드되는 컨텍스트 자체를 덜어 내는 쪽이라면, **프롬프트
캐싱**은 남은 컨텍스트를 90% 싼값에 다시 쓰는 쪽입니다. 이 섹션에서는
캐싱을 언제 켤지 가르는 손익분기 규칙과 설정 방법을 다룹니다.


## 이 섹션의 문서

- [프롬프트 캐싱 — 개념과 Claude Code에서의 동작](/ko/cost-optimization/prompt-caching) — 2개 요청 손익분기 규칙, 브레이크포인트 배치, statusline 모니터링
