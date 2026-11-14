# `agents-apply`

operation `agents.apply` · 코드 `src/ops/agents_apply.rs`

`agents-apply PROJECT --plan-digest D --permission-reference R` 는 stdin 의 미리보기 계획대로 AGENTS.md 관리 구역을 쓴다. `src/main.rs` 에서 그대로 옮김(#2834 R4a).

자리표시자 프로필(#2936 R3): 프로필이 첫 실행이 쓴 그대로(`tools`·`development_methods` 가 모두 `["unspecified"]` 이고 사용자·저장소가 준 결정이 없음)이면 AGENTS.md 를 쓰기 읽기·쓰기 전에 `profile blocked: …` (`blocked`) 로 멈추고 `/moai project` 인터뷰로 보낸다. 한 축이라도 답했거나 사용자·저장소 결정이 하나라도 있으면 막지 않는다. 첫 실행 자동 초기화(`launcher_init`)는 이 명령을 거치지 않는다.
