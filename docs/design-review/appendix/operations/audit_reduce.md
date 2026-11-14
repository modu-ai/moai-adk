# `audit-reduce`

operation `audit.reduce` · 코드 `src/ops/audit_reduce.rs` · Issue #2004 (05-D, 이름은 P06 뒤 확정)

`audit-reduce PROJECT` 는 PROJECT 의 독립 감사다. stdin 의 typed 요청(`protocol_version` 1 만, 다르면 `unsupported protocol_version; supported: 1`)을 받아 승인된 검증을 스스로 다시 돌리고 admission 을 정한다. stdin 의 receipt 는 결코 세지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R1).
