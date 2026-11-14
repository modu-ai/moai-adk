# `audit-delta`

operation `audit.delta` · 코드 `src/ops/audit_delta.rs`

`audit-delta [PROJECT] --pass-head P --tip T --head H`(PROJECT 기본 `.`, 세 값은 40자 소문자 hex)는 로컬 두 부모 병합 하나를 통제된 재연(replay)과 비교한다. 증거일 뿐 AUDIT PASS 가 아니다. stdin 은 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R1).
