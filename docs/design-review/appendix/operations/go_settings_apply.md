# `go-settings-apply`

operation `go_settings.apply` · 코드 `src/ops/go_settings_apply.rs`, `src/go_settings_clean.rs` · Issue #2936 (조각 L2a)

`go-settings-apply --plan-digest D --permission-reference REF [PROJECT]` 는 확인한 `go-settings-plan` 그대로 바꾼다. 계획을 다시 만들어 digest 가 다르거나 `held`·`blocked` 이거나 권한 참조가 없으면 아무것도 쓰지 않는다. 순서: 영수증 폴더(`<상태>/receipts/go-settings-<id>/`)에 옛 `.claude/settings.json` 바이트를 `before/settings.json` 으로 남기고(영수증 폴더와 `before/` 는 0700, 사본은 0600, 영수증 폴더에 `*` 한 줄짜리 `.gitignore` 를 먼저 두어 관리 `.gitignore` 블록이 아직 없는 프로젝트에서도 `git add -A` 가 사본을 잡지 않는다), `plansDirectory`·`subagentStatusLine` 키를 빼고 `statusLine` 을 Rust launcher 직접 경로 줄로 바꾼다(다른 키는 값도 자리도 그대로). 쓰기가 실패하면 파일은 그대로이고 영수증은 `rolled_back`. `.claude/settings.local.json` 은 쓰지 않는다.
