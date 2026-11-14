# `hooks-plan`

operation `hooks.plan` · 코드 `src/ops/hooks_plan.rs` · rc.2 I1 #2238

`hooks-plan PROJECT --host claude|codex [--lane-loop L --repo R | --lane-loop-off]` 는 한 호스트의 MoAI hook 묶음 미리보기를 낸다. 쓰지 않고, 생성될 hook 명령이 PATH 에서 풀리는지와 지금 hook 상태(`hook_status`, digest 밖)를 덧붙인다. `src/main.rs` 에서 그대로 옮김(#2834 R4b).
