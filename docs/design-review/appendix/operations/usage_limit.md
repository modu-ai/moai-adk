# `usage-limit`

operation `usage.limit` · 코드 `src/ops/usage_limit.rs`, `src/usage_limit.rs`, 기록 쓰기 `src/statusline/mod.rs` (#2668 S1)

`usage-limit HOME` 은 Claude Code 상태줄이 그릴 때마다 남기는 `HOME/.moai-cli/state/usage-limits.json`(창별 `used_percentage`·`resets_at`, 그린 시각)을 읽어 판정한다. `state`: `reached`(어느 창이 100), `near`(90 이상), `ok`, `unknown`(기록 없음·읽을 수 없음·열린 창 없음·폴더가 링크). `near`·`reached` 면 그 창 가운데 가장 늦게 풀리는 것의 `window`·`resets_at` 과 `renew_until`(= `resets_at` + 15분, 레인이 자기 claim lease 를 늘릴 시각)을 함께 준다. 지난 창(`resets_at` ≤ 지금)은 열린 창으로 치지 않는다. 아무것도 쓰지 않는다. 기록은 claude.ai 구독 세션의 첫 응답 뒤에만 생긴다(상태줄 입력에 `rate_limits` 가 그때부터 온다). 결과: `state`, `record`, `observed_at`, `near_percent`, `windows`, 그리고 `near`·`reached` 일 때 `window`, `resets_at`, `renew_until`; `unknown` 이면 `reason`.
