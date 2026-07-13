---
title: /moai goal
weight: 25
draft: false
---

완료 조건을 선언하면 세션이 그 조건을 충족할 때까지 스스로 일하는 **조건 선언형 자율 루프** 명령어입니다. `/moai goal "<조건>"`으로 완료 조건을 arm하고, 매 턴 종료 시 `stop-goal` Stop 훅이 조건 충족 여부를 평가하여 충족될 때까지 다음 턴을 자동으로 시작합니다. `/moai goal status [--all]`로 진행 상황을 확인하고, `/moai goal clear`로 루프를 해제하며, `/moai goal resume`으로 중단 시점부터 이어서 작업합니다. `/moai goal`은 전용 슬래시 커맨드 파일이 없고 `moai` 스킬 라우팅과 `moai goal` CLI를 통해 진입하는 프로그래매틱 명령어 표면입니다.
