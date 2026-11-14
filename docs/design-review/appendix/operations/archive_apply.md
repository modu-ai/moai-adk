# `archive-apply`

operation `archive.apply` · 코드 `src/ops/archive_apply.rs` · D2 #2188

`archive-apply PROJECT --plan-digest D --permission-reference R [--grace-days N]` 는 계획을 다시 계산해 `plan_digest` 가 같을 때만 보관할 수 있는 SPEC 을 옮긴다(REQ-D2-01). SPEC 마다 해시·복사·되읽기·색인 뒤에 원본을 지우고, 하나라도 어긋나면 옮김을 멈추고 원본을 둔다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).
