# `plan-admit`

operation `plan.admit` · 코드 `src/ops/plan_admit.rs` · Issue #2002 (05-B, 이름은 P06 뒤 확정)

`plan-admit PROJECT` 는 승인된 plan 의 RUN admission 을 판정한다. stdin 의 typed 요청(`protocol_version` 1 만)을 받고, 프로젝트의 `HEAD^{commit}` 을 git 으로 읽는다. 쓰지 않는다. 결과는 `ready`/`stale`/`blocked` 와 사유 코드. `src/main.rs` 에서 그대로 옮김(#2834 R1).
