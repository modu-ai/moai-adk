# `cost-report`

operation `cost.report` · 코드 `src/ops/cost_report.rs` · D5 #2191

`cost-report` 는 run/verify 영수증들의 비용 항목을 합친다(REQ-D5-01). stdin 의 typed 요청(영수증 배열, `protocol_version` 은 안에서 확인), 내역 봉투는 stdout, 쓰지 않는다. 외부 backend 호출은 구조적으로 0, 토큰은 없으면 없다고 하고 추정하지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
