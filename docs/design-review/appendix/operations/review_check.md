# `review-check`

operation `review.check` · 코드 `src/ops/review_check.rs` · Issue #2614 AC-3

`review-check PROJECT --head H` 는 올리기 전 검토 지적의 형식을 검사한다. stdin `{"schema_version":1,"findings":[…]}`(64 KiB 이하); 지적마다 severity, HEAD 의 `file:line`, 제목, 그 줄에서 인용한 근거, 확인된 검증이 있어야 한다. 결과: 판정(`important` 하나라도 있으면 FAIL), 셈, 올릴 본문, digest. git 만 읽는다. `src/main.rs` 에서 그대로 옮김(#2834 R1).
