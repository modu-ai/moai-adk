---
title: CLI 레퍼런스
weight: 45
draft: false
description: "터미널에서 실행하는 moai CLI 커맨드의 상세 레퍼런스."
---

{{< callout type="info" >}}{{< icon flash primary >}} <strong>소속 가치</strong>: 토크노믹스 · 에이전틱 루프 엔지니어링 · 에이전틱 하네스
{{< /callout >}}
<!-- @value: tokenomics, self-learning, agentic-harness -->

터미널에서 실행하는 `moai`(Go 바이너리) 커맨드의 상세 레퍼런스입니다. 각 커맨드의 플래그, 하위 명령어, 사용 예시를 다룹니다.

Claude Code 대화창에서 입력하는 슬래시 `/moai` 커맨드(워크플로우/유틸리티 명령어)와는 완전히 다른 도구입니다. 터미널 `moai` 가 프로젝트 초기화와 템플릿 배포, SPEC 라이프사이클 관리, 하네스(harness) 라우팅 같은 파일시스템 작업을 맡고, 슬래시 `/moai` 는 Claude Code 안에서 에이전트(스스로 일하는 AI)를 지시하는 대화 커맨드이기 때문에, 두 축을 한 색인에서 명확히 나눠 둡니다. 이 분리가 사용자의 가장 흔한 혼란을 줄여 주기 때문에, 각 커맨드가 어느 쪽에 속하는지부터 확인하고 본문으로 내려가는 것을 권합니다.

워크플로우 명령어는 [워크플로우 명령어](/ko/workflow-commands), 유틸리티 명령어는 [유틸리티 명령어](/ko/utility-commands) 섹션을 참조하세요. `moai` CLI 전체 개요는 [CLI 개요](/ko/getting-started/cli)에서 확인할 수 있습니다.
