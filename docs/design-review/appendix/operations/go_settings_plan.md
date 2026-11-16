# `go-settings-plan`

operation `go_settings.plan` · 코드 `src/ops/go_settings_plan.rs`, `src/go_settings_clean.rs` · Issue #2936 (조각 L2a)

`go-settings-plan [PROJECT]` 는 `.claude/settings.json` 에서 Go 세대가 만든 `.moai/` 참조를 걷으면 어떻게 되는지 보여 준다. 쓰지 않는다. Go 세대 프로젝트(`state_root` 가 `go`)에서만, 그리고 값이 Go 도구가 쓴 것과 **한 글자도 다르지 않을 때만** Go 가 만든 값으로 본다: `statusLine`·`subagentStatusLine` 이 정확히 `{"type":"command","command":X}` 이고 X 가 `$CLAUDE_PROJECT_DIR/.moai/status_line.sh`(앞에 `${…}` 또는 따옴표가 붙은 네 가지 표기), `plansDirectory` 가 정확히 `.moai/plans`. 한 글자라도 다른 값은 사용자 값이라 건드리지 않고 `kept` 에 키 이름만 올린다. 값은 어디에도 찍지 않는다.

필드: `action` (`replace` | `none` | `held` | `blocked`), `go_keys` (`{key, change: remove|replace}`), `kept`, `launcher` (`{path, state: rust|missing|not_rust|unprobed|not_needed}`), `plans_files` (`.moai/plans` 안 일반 파일 수; 링크는 따라가지 않고, 항목 1000개까지만 보며 넘으면 `more_than_limit`; 지우지 않고 go-state 백업과 함께 옮겨진다), `plan_digest` (settings 바이트·launcher 상태·결과를 묶는다), `next_actions`.

`statusLine` 은 PATH 의 `moai` 이름이 아니라 Rust 실행 파일의 직접 경로(`"$HOME/.moai-cli/path/moai-cli" statusline`)로 바뀐다(리더 결정 D2 = C). 그 줄을 쓰기 전에 `$HOME/.moai-cli/path/moai-cli` 가 실행 가능한 일반 파일(또는 그리로 풀리는 링크)이고 `statusline --help` 첫 줄이 `Render the Claude Code status line` 임을 보인다(`status_line` 의 안전한 탐침). 보일 수 없으면 `held`: 아무것도 쓰지 않고 `reason` 과 `next`(확인 명령)를 낸다. `.claude/settings.local.json` 은 쓰지 않는다.
