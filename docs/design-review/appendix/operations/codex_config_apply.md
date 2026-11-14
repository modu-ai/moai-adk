# `codex-config-apply`

operation `codex.config_apply` · 코드 `src/ops/codex_config_apply.rs`

`codex-config-apply PROJECT --plan-digest D --permission-reference R` 는 같은 stdin·digest·권한 참조로 미리보기한 프로젝트 설정을 쓴다. 사용자 HOME 과 CODEX_HOME 의 설정은 거절하고 trust 는 바꾸지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R4a).
