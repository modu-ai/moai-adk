# `forge-report`

operation `forge.report` · 코드 `src/ops/forge_report.rs`

`forge-report PROJECT --permission-reference R` 는 run 의 보고 댓글 하나를 그 Issue 에 만들거나 고친다. stdin 의 typed payload(64 KiB 넘으면 `input exceeds 64 KiB; put large output in an artifact reference, not the payload`), 봉투는 stdout. 큰 출력은 본문이 아니라 artifact 참조로. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
