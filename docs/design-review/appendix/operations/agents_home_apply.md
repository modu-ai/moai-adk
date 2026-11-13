# `agents-home-apply`

operation `agents.home_apply` · 코드 `src/ops/agents_home_apply.rs`, `src/agent_install.rs` (#2592 S2)

`agents-home-apply HOME PACKAGE --plan-digest D --permission-reference R` 는 다시 계획해 digest 가 같을 때만 `create`·`update` 를 01-B 발행 엔진으로 쓴다. 영수증 `HOME/.moai-cli/agents/receipts/<id>/receipt.json`(pending → committed, 파일별 before/after sha256)과 before 바이트 사본을 남기고, host 별 소유 기록 `HOME/.moai-cli/agents/<host>.json` 을 갱신한다. 사용자 수정·충돌 파일은 쓰지 않는다. 결과: `state`(`applied`·`unchanged`), `plan_digest`, `receipt`, `written`.
