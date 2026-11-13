# `decision-due`

operation `decision.due` · 코드 `src/ops/decision_due.rs` · Issue #2619

`decision-due FILE [--now YYYY-MM-DDTHH:MM:SSZ]` 는 답이 없는 결정을 그 시각 기준으로 나눈다: `open`(기한 전), `default_applies`(기한이 지났고 되돌릴 수 있으며 범주 `none` 이고 user 결정), `hold`(기한이 지났지만 되돌릴 수 없거나 `guarded`·5범주). 답이 있는 결정은 `answered`. 읽기만 하고 기본값을 적용하지 않는다. 형식 오류가 있는 파일은 거부한다.
