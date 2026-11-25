# `flow-measure`

operation `flow.measure` · 코드 `src/ops/flow_measure.rs`, `src/flow_measure.rs` · SPEC-20261125-flow-measure

`flow-measure [C-<번호>...] [--hours <N>] [--project <dir>] [--home <dir>] [--now <unix 초>]` 는 흐름의 구간 시간을 원장에서 읽기 전용으로 찍어요(`writes: false`). 카드 이름도 `--hours` 도 없으면 거절해요. 감사 보고서의 정의를 그대로 써요: 제출은 PASS 판정 앞(같은 시각 포함)의 가장 새 제출 노트(`LANE-READY-INTEGRATION`, `REPAIRED`, `READY `; 리더가 쓴 노트와 `decision` 노트는 제외), 판정은 `verdict` 가 `pass` 인 `card.verdict` 이벤트(행위자가 감사자 세션), done 은 그 판정 뒤 첫 `card.finish` 이벤트, 반납은 예약이 `held` 를 떠난 시각이에요.

- 출력: `ledger`(`present` 또는 `absent`), `now`, `cards`, `pass_to_done`, `window`, `table`(사람이 읽는 줄), `unreadable_verdicts`(읽지 못한 판정 이벤트 수), `writes`. 시간은 초 정수이고 `table` 은 분(소수 한 자리)이에요. 끝이 관찰되지 않은 구간은 `null` 이에요(0 이 아니에요).
- `cards[]`: `card`, `state`, `submit_note`·`submit_at`, `verdict_at`, `done_at`, `submit_to_verdict_sec`, `pass_to_done_sec`, `fail_count`, `auditor`, 40자리 `head`(PASS 판정의 head), `reservation`(`generation`, `status`, `left_held_at`).
- `pass_to_done`: 이름 준 카드들의 `{n, p50_sec, p90_sec, values_sec}`. P50/P90 은 nearest rank(올림한 `p×n/100` 번째 값)이고 표본이 없으면 `null` 이에요.
- `window`(`--hours N`, 끝은 `--now` 또는 원장 시계, 양 끝 포함): `issued`(`card.add`), `landed`(`card.landed`), `finished`(`card.finish`)를 따로 세고, `stages` 의 `take_to_submit`·`submit_to_verdict`·`pass_to_landed`·`pass_to_done` 은 끝 이벤트가 창 안에 있는 구간의 `{n, p50_sec, p90_sec}` 이에요. `first_verdict` 는 첫 판정이 창 안에 있는 카드의 `n`·`fail`·`fail_rate`, `lane_blocked` 는 창 안 `LANE-BLOCKED` 노트의 `total` 과 `kind=` 별 수(`kind` 가 없으면 `unspecified`)예요.
- 원장이 없는 프로젝트는 빈 목록과 `ledger: "absent"` 로 끝나고(종료 0) 아무것도 만들지 않아요. 영수증 폴더와 리더 노트 길이는 읽지 않아요.
