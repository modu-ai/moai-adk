# `ledger-status`

operation `ledger.status` · 코드 `src/ops/ledger_status.rs`, `src/ledger/command.rs` · Issue #3068 P1

`ledger-status PROJECT [--home DIR]` 는 프로젝트 원장의 상태를 읽기만 하고 보인다. 결과: `project_id`, `ledger_path`, `exists`, `state`(`absent`|`ready`|`unavailable`), `schema_version`(파일의 `user_version` 또는 null), `wal`(항상 false, 헤더가 WAL 이면 `unavailable` 과 `problem: wal_unexpected` 와 함께 true), `cards`(`total` 과 `by_state`), `problem`, `folder_modes`(`{"projects": "0755", "project": "0700"}` 8진 글자, 없는 폴더는 null; 읽기만 한다), `writes: false`. `.moai`·`projects`·`<id>` 가 링크이면 `state: unavailable`, `problem: path_link` 이고 그 너머는 읽지 않는다. 원장이 없으면 `exists: false` 로 정상 종료(exit 0)하고 폴더도 파일도 만들지 않는다. 읽기 전용 연결로 열며, 쓸 수 없는 파일도 고치지 않는다. `stale_temp`: 죽은 `register` 가 프로젝트 폴더에 남긴 `.creating-<pid>-<n>` 파일 중 그 pid 의 프로세스가 이 컴퓨터에 없는 것(보고만 하고 지우지 않는다; 같은 폴더가 없으면 빈 목록).
