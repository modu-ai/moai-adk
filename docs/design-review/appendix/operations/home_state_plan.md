# `home-state-plan`

operation `home_state.plan` · 코드 `src/ops/home_state_plan.rs`, `src/home_state.rs` · Issue #2723

`home-state-plan [--home DIR] [--now UTC]` 는 `~/.moai` 가 어느 세대 것인지와, Go 세대의 `~/.moai` 를 `~/.moai-backup` 으로 옮기면 무엇이 되는지를 보여 준다. 쓰지 않는다. 판정 표는 프로젝트와 같다(`state_root::classify_with`, #2725 S2a). HOME 의 Go 표식은 `user-assets.json`(정수 `schema_version` 과 `files` 객체가 있는 JSON), `state/active-sessions.json`, `credentials.yaml` 이다.

- `action`: `migrate`(Go 이고 막는 이유 없음), `none`(없음·빈 폴더·Rust), `blocked`(섞임·모름·거부·중단, 또는 막는 이유가 있음).
- `backup`: `~/.moai-backup`, 이미 있으면 `~/.moai-backup-<UTC>`. 둘 다 있으면 `backup_name_taken` 으로 막는다. 덮어쓰지 않는다.
- `blocks`: `go_session_live`(`state/active-sessions.json.lock` 잠김, 또는 그 파일의 행 `pid` 중 살아 있는 것이 있음 — Go 는 lock 을 기록을 쓰는 순간에만 잡아서 pid 가 진짜 근거다. 끝난 세션의 pid 가 다른 프로세스로 재사용되면 계속 막힌다(안전한 쪽)), `session_lock_unknown`, `session_registry_unknown`(행 `pid` 배열로 읽을 수 없거나 pid 를 판단할 수 없음 — unix 밖에서는 행이 있으면 항상), `in_use`(그 안을 작업 폴더로 둔 프로세스), `process_check_unavailable`, `file_open_in_use`(그 안의 파일을 연 프로세스), `open_file_check_unavailable`, `nested_checkouts:N`(중첩 checkout — git worktree 가 있으면 옮길 때 등록이 끊김), `not_fully_checked`(항목 상한).
- `copy`: 기본 `none`. 맨 위 항목을 후보로 보이고(500개까지), 비밀(`credentials*`, `.env*`, `secrets`, 대소문자 무관)은 `kind: secret`, Go 표식이 든 `user-assets.json`·`state` 는 `kind: generation_marker`, 도구가 새 폴더에 쓰는 이름(`home-state-receipt.json`, `moai-cli.json`, `.tmp-write` 로 끝나는 임시 이름, 대소문자 무관)은 `kind: reserved` 로 고를 수 없게 표시한다.
- `MOAI_HOME` 이 다른 곳을 가리키면 경고한다(이 미리보기는 `~/.moai` 만 다룬다). 적용은 다음 단계이고, 실제 HOME 에서는 사용자 확인 뒤에만 한다.
