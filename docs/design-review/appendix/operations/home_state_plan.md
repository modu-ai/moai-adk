# `home-state-plan`

operation `home_state.plan` · 코드 `src/ops/home_state_plan.rs`, `src/home_state.rs` · Issue #2723

`home-state-plan [--home DIR] [--now UTC]` 는 `~/.moai` 가 어느 세대 것인지와, Go 세대의 `~/.moai` 를 `~/.moai-backup` 으로 옮기면 무엇이 되는지를 보여 준다. 쓰지 않는다. 판정 표는 프로젝트와 같다(`state_root::classify_with`, #2725 S2a). HOME 의 Go 표식은 `user-assets.json`(정수 `schema_version` 과 `files` 객체가 있는 JSON), `state/active-sessions.json`, `credentials.yaml` 이다.

- `action`: `migrate`(Go 이고 막는 이유 없음), `none`(없음·빈 폴더·Rust), `blocked`(섞임·모름·거부·중단, 또는 막는 이유가 있음).
- `backup`: `~/.moai-backup`, 이미 있으면 `~/.moai-backup-<UTC>`. 둘 다 있으면 `backup_name_taken` 으로 막는다. 덮어쓰지 않는다.
- `blocks`: `go_session_live`(`state/active-sessions.json.lock` 잠김, 또는 그 파일의 행 `pid` 중 살아 있는 것이 있음 — Go 는 lock 을 기록을 쓰는 순간에만 잡아서 pid 가 진짜 근거다. 끝난 세션의 pid 가 다른 프로세스로 재사용되면 계속 막힌다(안전한 쪽)), `session_lock_unknown`, `session_registry_unknown`(행 `pid` 배열로 읽을 수 없거나 pid 를 판단할 수 없음 — unix 밖에서는 행이 있으면 항상), `in_use`(그 안을 작업 폴더로 둔 프로세스), `process_check_unavailable`, `file_open_in_use`(그 안의 파일을 연 프로세스), `open_file_check_unavailable`, `nested_checkouts:N`(중첩 checkout — git worktree 가 있으면 옮길 때 등록이 끊김), `not_fully_checked`(항목 상한).
- `copy`: 기본 `none`. 맨 위 항목을 후보로 보이고(500개까지), 비밀(`credentials*`, `.env*`, `secrets`, 대소문자 무관)은 `kind: secret`, Go 표식이 든 `user-assets.json`·`state` 는 `kind: generation_marker`, 도구가 새 폴더에 쓰는 이름(`home-state-receipt.json`, `moai-cli.json`, `.tmp-write` 로 끝나는 임시 이름, 대소문자 무관)은 `kind: reserved` 로 고를 수 없게 표시한다.
- `MOAI_HOME` 이 다른 곳을 가리키면 경고한다(이 미리보기는 `~/.moai` 만 다룬다). 적용은 다음 단계이고, 실제 HOME 에서는 사용자 확인 뒤에만 한다.

## `--selective` — 항목별 plan (#2815)

통째 이동은 `~/.moai/worktrees` 에 다른 프로젝트 worktree 와 그 안의 프로세스가 있는 기계에서 늘 막힌다(리더 실측 #2815). `home-state-plan --selective` 는 맨 위 항목마다 분류 하나와 사유를 낸다. 쓰지 않는다(D3). 적용·되돌리기는 다음 카드(D4)다. 이 플래그가 없으면 위의 통째 이동 출력이 그대로다.

- 분류표는 코드 한 곳 `home_state::SELECTIVE_TABLE` 에 있다(D1). 대소문자는 무시하고, `*` 로 끝나면 접두어로 맞춘다.
  - 비밀 행이 먼저다: `.env*`·`credentials*`·`secrets`·`gateway-auth` → `secret-manual`.
  - `worktrees` → `keep-in-place`(`worktree_root`, #2811 D8), `projects` → `keep-in-place`(`work_folder`, #2811 D7/D8a, 통째 이동의 `is_work_folder` 와 같은 이름).
  - `cache`·`logs`·`run`·`state`·`releases`·`rescue`·`scratch`·`tmp-*`·`user-assets.json` → `move-to-backup`(`go_run_data`).
  - 표에 없는 이름은 `unknown`(`not_in_table`)이고 옮기지 않는다.
- `secret-manual` 은 이름만으로 정한다. 열지도, 안을 걷지도, 프로세스와 맞춰 보지도 않는다(AC-3).
- 실행 중 프로세스가 쓰는 항목은 표와 상관없이 `keep-in-place` 다(D2). 사유는 셋 중 하나다.
  - `process_cwd`: 작업 폴더가 그 안에 있음
  - `file_open`: 그 안의 파일이 열려 있음
  - `process_command_line`: 명령 줄이 그 경로나 그 안을 가리킴
  - 프로세스 목록을 읽지 못하면 `move-to-backup` 항목은 `keep-in-place` 가 되고, 사유는 `process_check_unavailable` 등이다(fail-closed). 시험용 빌드(debug)에서는 `MOAI_CLI_TEST_PROCESS_CHECK=unavailable` 로 이 경우를 흉내 낸다. 릴리스 빌드에는 없다.
- `move-to-backup` 항목에 중첩 checkout 이 있으면 `keep-in-place`(`nested_checkout`)다. 상한 안에서 다 보지 못하면 `not_fully_checked` 다.
- `destination` 은 `move-to-backup` 일 때만 `<backup>/<이름>` 이다. `backup` 고르기는 통째 이동과 같다.
- `action`: `select`(Go 이고 막는 이유가 없고 옮길 항목이 있음), `none`, `blocked`. 세션 막음(`go_session_live` 등)과 `backup_name_taken` 은 통째 이동과 같다. 항목이 500개를 넘으면 `not_fully_listed`, 이름을 UTF-8 로 읽을 수 없는 항목이 있으면 `unreadable_names:N` 으로 막는다.
- 출력 칸: `mode: "selective"`, `entries[]`(`name`·`class`·`reason`·`destination`), `counts`, `total`, `plan_digest`, `writes: false`.
