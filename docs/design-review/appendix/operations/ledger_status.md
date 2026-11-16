# `ledger-status`

operation `ledger.status` · 코드 `src/ops/ledger_status.rs`, `src/ledger/command.rs` · Issue #3068 P1

`ledger-status PROJECT [--home DIR]` 는 프로젝트 원장의 상태를 읽기만 하고 보인다. 결과: `project_id`, `ledger_path`, `exists`, `state`(`absent`|`ready`|`unavailable`), `schema_version`(파일의 `user_version` 또는 null), `wal`(항상 false, 헤더가 WAL 이면 `unavailable` 과 `problem: wal_unexpected` 와 함께 true), `cards`(`total` 과 `by_state`), `problem`, `writes: false`. 원장이 없으면 `exists: false` 로 정상 종료(exit 0)하고 폴더도 파일도 만들지 않는다. 읽기 전용 연결로 열며, 쓸 수 없는 파일도 고치지 않는다.
