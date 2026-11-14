# `home-style-apply`

operation `home.style_apply` · 코드 `src/ops/home_style_apply.rs`, `src/home_style.rs` (#2875)

`home-style-apply HOME PACKAGE --plan-digest D --permission-reference R` 는 다시 계획해 digest 가 같을 때만 `create`·`append`·`merge`·`set` 을 01-B 발행 엔진으로 쓴다. 쓰기 직전 파일 sha256 이 계획 때와 다르면 멈춘다. `settings.json` 은 `outputStyle` 하나만 더하고 다른 키와 순서는 그대로 둔다(들여쓰기 2칸으로 다시 씀). 영수증 `HOME/.moai-cli/home-style/receipts/<id>/receipt.json`(pending → committed, 대상별 before/after sha256, 앞 기준)과 before 바이트 사본을 남기고, 도구별 소유 기록 `HOME/.moai-cli/home-style/<host>.json` 을 갱신한다. 충돌·blocked·skipped 대상은 쓰지 않는다. 결과: `state`(`applied`·`unchanged`), `plan_digest`, `receipt`, `written`.
