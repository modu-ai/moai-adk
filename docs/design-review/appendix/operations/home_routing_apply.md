# `home-routing-apply`

operation `home.routing_apply` · 코드 `src/ops/home_routing_apply.rs`, `src/home_routing.rs` (#2638)

`home-routing-apply HOME PACKAGE --plan-digest D --permission-reference R` 는 다시 계획해 digest 가 같을 때만 `create`·`append`·`merge` 를 01-B 발행 엔진으로 쓴다. 관리 구간만 바뀌고 구간 밖 바이트는 그대로다. 영수증 `HOME/.moai-cli/home-routing/receipts/<id>/receipt.json`(pending → committed, 도구별 before/after sha256, 앞 기준)과 before 바이트 사본을 남기고, 도구별 소유 기록 `HOME/.moai-cli/home-routing/<host>.json` 을 갱신한다. 충돌·blocked 도구는 쓰지 않는다. 결과: `state`(`applied`·`unchanged`), `plan_digest`, `receipt`, `written`.
