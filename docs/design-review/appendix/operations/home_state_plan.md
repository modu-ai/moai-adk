# `home-state-plan`

operation `home_state.plan` · 코드 `src/ops/home_state_plan.rs`, `src/home_state.rs` · Issue #2723

`home-state-plan [--home DIR] [--now UTC]` 는 `~/.moai` 가 어느 세대 것인지와, Go 세대의 `~/.moai` 를 `~/.moai-backup` 으로 옮기면 무엇이 되는지를 보여 준다. 쓰지 않는다. 판정 표는 프로젝트와 같다(`state_root::classify_with`, #2725 S2a). HOME 의 Go 표식은 `user-assets.json`(정수 `schema_version` 과 `files` 객체가 있는 JSON), `state/active-sessions.json`, `credentials.yaml` 이다.

- `action`: `migrate`(Go 이고 막는 이유 없음), `none`(없음·빈 폴더·Rust), `blocked`(섞임·모름·거부·중단, 또는 막는 이유가 있음).
- `backup`: `~/.moai-backup`, 이미 있으면 `~/.moai-backup-<UTC>`. 둘 다 있으면 `backup_name_taken` 으로 막는다. 덮어쓰지 않는다.
- `blocks`: `go_session_live`(`state/active-sessions.json.lock` 잠김), `session_lock_unknown`, `in_use`(그 안을 작업 폴더로 둔 프로세스), `process_check_unavailable`, `nested_checkouts:N`(중첩 checkout — git worktree 가 있으면 옮길 때 등록이 끊김), `not_fully_checked`(항목 상한).
- `copy`: 기본 `none`. 맨 위 항목을 후보로 보이고(500개까지), 비밀(`credentials*`, `.env*`, `secrets`)은 `kind: secret` 으로 고를 수 없게 표시한다.
- `MOAI_HOME` 이 다른 곳을 가리키면 경고한다(이 미리보기는 `~/.moai` 만 다룬다). 적용은 다음 단계이고, 실제 HOME 에서는 사용자 확인 뒤에만 한다.
