# `run-verify`

operation `run.verify` · 코드 `src/ops/run_verify.rs` · Issue #2003 (05-C, 이름은 P06 뒤 확정)

`run-verify PROJECT` 는 고정된 계약(pinned contract)을 다시 확인하고 승인된 검증 명령을 PROJECT 에서 argv 로 돌려 receipt 를 돌려준다. stdin 의 typed 요청(`protocol_version` 1 만)과 프로젝트의 현재 `HEAD^{commit}` 을 쓴다. `src/main.rs` 에서 그대로 옮김(#2834 R1).
