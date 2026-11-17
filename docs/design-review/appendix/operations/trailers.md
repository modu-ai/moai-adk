# `trailers`

operation `trailers.format` · 코드 `src/ops/trailers.rs`, `src/trailers.rs` · SPEC-20261117-session-trailer(SDD-6, 사용자 결정 U-B 2026-11-17)

`trailers --spec SPEC-<yyyymmdd>-<slug>` 는 에이전트 커밋이 다는 신원 trailer 줄을 낸다: `Moai-Spec: <SPEC id>` 와, host 가 이 세션의 id 를 환경에 줄 때(`session-current` 와 같은 순서 `CLAUDE_CODE_SESSION_ID`, `CODEX_THREAD_ID`; 8–64자 16진수와 `-` 만) `Moai-Session: <id>`. 결과 `text` 를 커밋 메시지 끝 trailer 블록에 그대로 붙인다. id 가 없으면 `session_missing: true` 와 이유를 내고 SPEC 줄만 낸다(감사 관문은 그 커밋의 저자를 모르는 것으로 읽는다). 새 형식 SPEC id 만 받는다 — 카드 번호는 trailer 가 되지 않는다. 환경만 읽고 아무것도 쓰지 않는다.
