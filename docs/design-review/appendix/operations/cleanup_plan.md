# `cleanup-plan`

operation `cleanup.plan` · 코드 `src/ops/cleanup_plan.rs`, `src/update_cleanup.rs` · Issue #2872 (1단계, 설계 6872036842)

`cleanup-plan HOME STAGE` 는 업데이트와 Go 세대가 HOME 에 남긴 것을 보관 폴더로 옮기면 무엇이 옮겨지는지 보여 준다. 쓰지 않는다. STAGE 는 `update-plan` 과 같은 스테이징된 패키지 폴더다.

- `items`: 항목마다 `path`(HOME 기준), `class`, `action`(`archive`|`keep`), `reason`, 파일이면 `sha256`, 폴더면 `tree_digest`(정렬한 `rel/`·`rel\t<sha>` 줄의 sha256).
- `projection_extra`: `~/.claude/skills/moai`·`~/.agents/skills/moai` 안에서 패키지에 없는 파일. host receipt 가 소유하고 디스크 바이트가 소유 sha 와 같으면(`update-plan` 의 수정 없는 `remove_blocked`) `owned_unchanged`, Go manifest 가 증명하면 `go_manifest` 로 옮긴다. 아니면 `user_edited`·`receipt_pending`·`receipt_unavailable`·`go_manifest_unreadable`·`unrecorded` 로 남긴다.
- `go_skill`: `~/.claude/skills/moai-*`·`~/.agents/skills/moai-*` 폴더 하나가 항목 하나다. 안의 모든 파일이 Go manifest(`~/.moai/user-assets.json`, 키 `<slug>/<rel>`, slug `claude-skills`·`claude-agents`·`agents-skills`·`codex-agents`)의 sha256 과 맞고 링크·특수 파일·비밀이 없을 때만 옮긴다(`unproven_files` 면 남김).
- `go_agent`: `~/.claude/agents`·`~/.codex/agents` 바로 아래 파일 중 이름이 manifest 에 있는 것. 바이트가 맞으면 옮기고, 다르면 `user_edited`. manifest 에 없는 이름은 나열하지 않는다.
- `moai_leftover`: `~/.moai` 가 Go 세대일 때만. 맨 위 항목 중 `home_state::SELECTIVE_TABLE` 의 `move-to-backup` 이면서 빈 폴더이거나 이름이 `tmp-*`·`scratch`·`cache` 인 것만 옮긴다. 나머지는 모두 `keep`.
- 옮기지 않는 것: 비밀 이름(`credentials*`·`.env*`·`secrets`·`gateway-auth`, 바이트를 읽지도 해시하지도 않음, `secret`), Go 실행 데이터(`claude-profiles`·`backups`·`state`·`releases`·`db`·`handoffs`·`gpt-appserver`·`backup`, `go_runtime`), 링크·특수 파일(따라가지 않음), Go 세션이 살아 있거나 판단할 수 없을 때의 모든 Go 항목(`in_use`·`in_use_unknown`). 읽을 수 없거나 망가진 manifest 는 `go_manifest_unreadable`.
- 상한: 항목 2048, 폴더 하나의 항목 4096·바이트 64 MiB·깊이 32. 넘으면 그 무리는 `over_limit` 로 남긴다(항목 상한을 넘으면 `truncated` 이고 아무것도 옮기지 않는다).
- `confirm_token`: `cleanup:` 과 archive 항목의 `path\tdigest\n`(경로 순) sha256 앞 16자. archive 항목이 없으면 `null`.
