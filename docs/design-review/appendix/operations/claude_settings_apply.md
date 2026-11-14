# `claude-settings-apply`

operation `claude.settings_apply` · 코드 `src/ops/claude_settings_apply.rs`

`claude-settings-apply PROJECT --plan-digest D --permission-reference R [--scope S] [--confirm-replace TOKEN]` 는 같은 stdin 으로 미리보기한 설정 합치기를 적용한다. 프로젝트 설정 파일과 영수증만 쓰고 HOME 은 거절한다. 충돌 계획의 `confirm_token` 을 확인받으면 기존 파일을 `.moai-cli/settings-backups/` 에 두고 충돌 키를 바꾼다(#2383). `src/main.rs` 에서 그대로 옮김(#2834 R4b).
