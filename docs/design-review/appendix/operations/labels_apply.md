# `labels-apply`

operation `labels.apply` · 코드 `src/ops/labels_apply.rs` · #2434

`labels-apply PROJECT --plan-digest D --permission-reference R` 는 stdin 의 승인된 라벨 동기화 계획을 적용한다(POST/PATCH 만, 결정 40). 권한 참조가 비면 거절한다. `src/main.rs` 에서 그대로 옮김(#2834 R4a).
