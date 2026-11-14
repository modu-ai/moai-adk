# `claude-settings-plan`

operation `claude.settings_plan` · 코드 `src/ops/claude_settings_plan.rs` · #2239

`claude-settings-plan PROJECT [--scope project|local]` 는 stdin 의 관리 키(`outputStyle`·`statusLine`·`permissions`)를 PROJECT/.claude/settings.json(또는 `local` 이면 settings.local.json, #2581)에 세 갈래로 합친 미리보기를 낸다. 다른 키는 거절, 쓰지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R4b).
