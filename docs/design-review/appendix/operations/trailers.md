# `trailers`

operation `trailers.format` · 코드 `src/ops/trailers.rs`, `src/trailers.rs` · SPEC-20261117-session-trailer(SDD-6, 사용자 결정 U-B 2026-11-17)

`trailers --spec SPEC-<yyyymmdd>-<slug>` 는 에이전트 커밋이 다는 신원 trailer 줄을 낸다: `Moai-Spec: <SPEC id>` 와, host 가 이 세션의 id 를 환경에 줄 때(`session-current` 와 같은 순서 `CLAUDE_CODE_SESSION_ID`, `CODEX_THREAD_ID`; 8–64자 16진수와 `-` 만) `Moai-Session: <id>:<worktree key>`. worktree key 는 지금 폴더가 든 worktree 의 정식 git dir(`git rev-parse --absolute-git-dir`, git 환경을 지우고 읽음)의 sha256 앞 12자리다 — Claude Code 하위 에이전트는 부모의 session id 를 물려받으므로 id 만으로는 저자와 그 감사 에이전트를 가르지 못한다(SDD-6 재감사 B3, 리더 결정). git worktree 가 아니면 이 줄도 내지 않는다. 결과 `text` 를 커밋 메시지 끝 trailer 블록에 그대로 붙인다. id 가 없으면 `session_missing: true` 와 이유를 내고 SPEC 줄만 낸다(감사 관문은 그 커밋의 저자를 모르는 것으로 읽는다). 새 형식 SPEC id 만 받는다 — 카드 번호는 trailer 가 되지 않는다. 환경만 읽고 아무것도 쓰지 않는다.
