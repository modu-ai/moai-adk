# `parity-check`

operation `parity.check` · 코드 `src/ops/parity_check.rs` · D6 #2192

`parity-check TEMPLATE GENERATED [--doc DOC]` 는 템플릿과 생성 파일(선택으로 세 번째 DOC 묶음)의 의미 대응을 키마다 00-A 어휘(`missing`·`extra`·`value_mismatch`·`identical`·`doc_only`)로 분류한다. 읽기만 하고 stdin 은 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R1).
