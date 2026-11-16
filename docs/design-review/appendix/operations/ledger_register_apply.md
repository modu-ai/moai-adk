# `ledger-register-apply`

operation `ledger.register.apply` · 코드 `src/ops/ledger_register_apply.rs`, `src/ledger/command.rs` · Issue #3068 P1

`ledger-register-apply PROJECT [--home DIR] --plan-digest D --permission-reference REF` 는 원장 폴더와 파일을 만드는 유일한 명령이다. 같은 입력으로 다시 계산한 미리보기의 digest 가 같을 때만 움직인다(아니면 `stale or altered plan`, 아무것도 쓰지 않음). 파일은 같은 폴더에 임시 이름으로 완성·검사한 뒤 하드 링크로 제자리에 놓고 임시 이름을 지운다. 이미 있는 유효한 원장은 건드리지 않고 `unchanged` 를 낸다. 만드는 폴더는 umask 와 상관없이 0700 이고, 이미 있는 `projects`·`projects/<id>` 가 이 사용자의 것인데 열려 있으면 0700 으로 조이고 `tightened` 에 경로를 담는다(`.moai` 는 절대 chmod 하지 않는다). 다른 사용자의 폴더는 `folder_not_ours`, 링크는 `path_link` 로 거절한다. 자세한 규칙과 남는 경쟁은 `ledger_folders.md`. 이 프로젝트의 원장이 아닌 파일(`ledger_unavailable`, `id_collision`, `wal_unexpected`, `schema_newer`)은 오류이며 대체하지 않는다.
