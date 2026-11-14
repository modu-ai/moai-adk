# `codex-projection-plan`

operation `codex.projection_plan` · 코드 `src/ops/codex_projection_plan.rs`

`codex-projection-plan PACKAGE --target DIR [--display-name …] [--short-description …] [--default-prompt …] [--allow-implicit-invocation B] [--status-line ITEM]… [--disable-bundled-skills]` 는 공통 패키지를 명시한 대상(이름이 `moai` 인 절대 경로 스킬 폴더)에 Codex 사본으로 옮기는 계획을 만든다. `--status-line`·`--disable-bundled-skills` 는 HOLD 로 보고만 하고 쓰지 않는다. 쓰지 않고 stdin 도 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2a).
