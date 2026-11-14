# `go-hooks-plan`

operation `go_hooks.plan` · 코드 `src/ops/go_hooks_plan.rs`, `src/go_hooks.rs` · Issue #2877

`go-hooks-plan [PROJECT]` 은 Go 세대가 만든 프로젝트 hook 을 바꾸면 어떻게 되는지 보여 준다. 쓰지 않는다. `.claude/settings.json` 의 hook 중 `.claude/hooks/moai/<이름>` 을 돌리는 것(`go_entries`)은 빠지고, 다른 것(`kept_entries`)은 남고, MoAI hook 묶음(`added`)이 들어가고, 스크립트(`scripts`)는 지우지 않고 보관 폴더로 옮긴다. 다른 도구의 hook 이 `hooks` 에 있으면 지금은 `blocked`(`other_hooks_present`) — MoAI 쓰기가 `hooks` 키 전체를 소유해서, 남기려면 결정이 먼저다. `plan_digest` 는 settings 바이트·스크립트·결과를 묶는다.
