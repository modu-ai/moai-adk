# `gate-schedule-plan`

operation `gate.schedule_plan` · 코드 `src/ops/gate_schedule_plan.rs`, `src/gate_schedule.rs` · Issue #2906 S2

`gate-schedule-plan HOME --os macos|linux [--repo OWNER/REPO] [--moai-bin ABS] [--dir ABS] [--path PATH] [--env NAME=VALUE …] [--merge --merge-permission-reference REF]` 는 `gate-loop --once` 를 주기적으로 돌리는 스케줄러 항목을 그려 보여 준다. 아무것도 쓰지 않는다.

- macOS: LaunchAgent plist `HOME/Library/LaunchAgents/moai.gate-loop.<owner>.<repo>.plist`. `Label` 은 파일 이름과 같고, `ProgramArguments` 는 `--moai-bin` 의 절대 경로와 `gate-loop --once --dry-run --repo … --dir …`, `WorkingDirectory` 는 `--dir`, `EnvironmentVariables` 는 PATH(와 `--env`)만, `StartInterval 90`, 표준 출력·오류는 상태 폴더의 `logs/gate-loop.log`.
- Linux: `<상태 폴더>/schedule/gate-loop.crontab` 에 주석 한 줄과 `*/2 * * * * cd <dir> && PATH=<path> flock -n <상태 폴더>/schedule/gate-loop.lock <moai> gate-loop --once … >> <log> 2>&1` 한 줄. 변수는 명령 앞에 붙여 사용자의 다른 cron 줄을 바꾸지 않는다. 상태 폴더는 `home_root` 가 고른 `~/.moai-cli` 또는 `~/.moai` 다.
- 기본은 `--dry-run` 항목이다(결정 Q3). `--merge` 는 `--merge-permission-reference REF` 와 함께일 때만 병합 항목(`--permission-reference REF`)을 만든다. 어느 쪽도 `--review-confirmed` 를 넣지 않는다 — G7 은 사람이 멈추는 곳으로 남는다.
- 토큰 같은 변수(`GH_TOKEN`·`GITHUB_TOKEN`·`MOAI_AUDIT_APP_TOKEN`, 이름이 `_TOKEN`·`_KEY`·`_SECRET` 로 끝나는 것)를 `--env` 나 `--path` 에 넣으면 `token_variable` 로 거절한다. `ghp_` 같은 GitHub 토큰 값도 거절한다.
- `--os windows` 는 `windows_not_supported` 로 거절한다(결정 Q1). 대상까지의 폴더나 영수증 폴더가 링크면 거절한다.
- 켜는 명령과 끄는 명령(`launchctl bootstrap`/`bootout`, `crontab`)은 글로만 준다(결정 Q5). 켜기 전에 쉘 루프 `auto_loop.sh` 를 먼저 멈춘다.
- 결과: `schema_version`, `os`, `home`, `label`, `target_path`, `action`(`create`·`replace`·`unchanged`), `expected_sha256`, `entry`, `entry_sha256`, `mode`(`dry-run`·`merge`), `enable`, `disable`, `plan_digest`.
