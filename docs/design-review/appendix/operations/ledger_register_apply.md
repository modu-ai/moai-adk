# `ledger-register-apply`

operation `ledger.register.apply` · 코드 `src/ops/ledger_register_apply.rs`, `src/ledger/command.rs` · Issue #3068 P1

`ledger-register-apply PROJECT [--home DIR] --plan-digest D --permission-reference REF` 는 원장 폴더와 파일을 만드는 유일한 명령이다. 같은 입력으로 다시 계산한 미리보기의 digest 가 같을 때만 움직인다(아니면 `stale or altered plan`, 아무것도 쓰지 않음). 파일은 같은 폴더에 임시 이름으로 완성·검사한 뒤 하드 링크로 제자리에 놓고 임시 이름을 지운다. 이미 있는 유효한 원장은 건드리지 않고 `unchanged` 를 낸다. 이 프로젝트의 원장이 아닌 파일(`ledger_unavailable`, `id_collision`, `wal_unexpected`, `schema_newer`)은 오류이며 대체하지 않는다.
