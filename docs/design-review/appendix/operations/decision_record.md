# `decision-record`

operation `decision.record` · 코드 `src/ops/decision_record.rs`

`decision-record PROJECT --run RUN --permission-reference R` 는 stdin 의 기본 결정 기록 하나를 `.moai-cli/decisions/<RUN>.jsonl` 에 덧붙인다. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
