# `codex-doctor`

operation `codex.doctor` · 코드 `src/ops/codex_doctor.rs`

`codex-doctor PROJECT [--cwd] [--codex-home] [--home] [--fallback]… [--max-bytes]` 는 Codex 의 AGENTS 사슬과 예산, 스킬 중복, `.codex/` 층을 본다. 읽기만, config.toml 은 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R3). Codex home 은 `--codex-home` 이 우선이고, 없으면 `src/host_home.rs` 의 규칙(카드 C-79): `--home`(기본 `$HOME`)이 이 계정의 홈이면 `CODEX_HOME`(절대 경로만; 빈 값·상대 경로는 `CODEX_HOME …` 문장으로 거절), 아니면 `HOME/.codex`.
