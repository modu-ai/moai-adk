# `codex-projection-apply`

operation `codex.projection_apply` · 코드 `src/ops/codex_projection_apply.rs` · I4 #2241

`codex-projection-apply PACKAGE --target DIR --plan-digest D --permission-reference R [--display-name …] [--short-description …] [--default-prompt …] [--allow-implicit-invocation B]` 는 계획한 Codex 사본을 그 대상에 설치한다. 낡은 계획·소스 manifest 불일치·충돌하는 대상 파일은 쓰기 전에 거부한다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).
