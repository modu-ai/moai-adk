# `go-settings-revert`

operation `go_settings.revert` · 코드 `src/ops/go_settings_revert.rs`, `src/go_settings_clean.rs` · Issue #2936 (조각 L2a)

`go-settings-revert --receipt ID --permission-reference REF [PROJECT]` 는 `go-settings-apply` 하나를 되돌린다. `.claude/settings.json` 이 apply 가 쓴 바이트 그대로일 때만 영수증의 `before/settings.json` 바이트로 돌리고 영수증을 `reverted` 로 바꾼다. 파일이 바뀌었거나 영수증이 `applied` 가 아니면 아무것도 쓰지 않는다. 지우지 않는다.
