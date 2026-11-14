# `codex-doctor`

operation `codex.doctor` · 코드 `src/ops/codex_doctor.rs`

`codex-doctor PROJECT [--cwd] [--codex-home] [--home] [--fallback]… [--max-bytes]` 는 Codex 의 AGENTS 사슬과 예산, 스킬 중복, `.codex/` 층을 본다. 읽기만, config.toml 은 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R3).
