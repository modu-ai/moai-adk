# `generation-plan`

operation `generation.plan` · 코드 `src/ops/generation_plan.rs` · 03-C #1997

`generation-plan PROJECT` 는 프로젝트 전용 생성 파일(에이전트·스킬)을 읽기만 해서 계획한다. stdin 의 typed 요청, 판정 봉투는 stdout. `src/main.rs` 에서 그대로 옮김(#2834 R4a).

자리표시자 프로필(#2936 R3): 프로필이 첫 실행이 쓴 그대로(`tools`·`development_methods` 가 모두 `["unspecified"]` 이고 사용자·저장소가 준 결정이 없음)이면 생성 계획을 만들기 읽기·쓰기 전에 `profile blocked: …` (`blocked`) 로 멈추고 `/moai project` 인터뷰로 보낸다. 한 축이라도 답했거나 사용자·저장소 결정이 하나라도 있으면 막지 않는다. 첫 실행 자동 초기화(`launcher_init`)는 이 명령을 거치지 않는다.
