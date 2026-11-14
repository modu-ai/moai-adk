# `lane-resume-note`

operation `lane.resume_note` · 코드 `src/ops/lane_resume_note.rs`, `src/lane_loop.rs` · Issue #2904

`lane-resume-note --lane NAME --issue N --claim ID --lease-until UTC --branch B --head SHA --uncommitted 0 --next STEP [--next STEP …] [--wait TEXT --clears WHO]` 는 카드마다 Issue 에 남기는 재개 메모의 글을 만든다(`text`). 한 세션이 오래 가는 레인(Codex `/goal`)이나 compact·clear 뒤의 레인이 이 메모부터 읽고 이어 간다.

형식: 첫 줄 `moai-resume: lane=<lane> issue=<n> claim=<claim 댓글 id> lease_until=<YYYY-MM-DDTHH:MM:SSZ> branch=<가지> head=<40hex> uncommitted=0`, 그다음 남은 단계마다 `next: <단계>`, 기다림이 있으면 `wait: <조건> clears=<leader|user|lane|author>`.

거부: `uncommitted` 가 0 아님(먼저 커밋하거나 따로 두기), head 가 소문자 40자 hex 아님, claim 이 7자리 미만, issue 0, lease 형식 틀림, 가지에 공백, `next` 없음·여러 줄, `clears` 가 넷 밖. `--wait` 와 `--clears` 는 함께만. 글을 낼 뿐 댓글을 쓰지 않는다(`writes: false`) — 게시는 레인 자기 댓글.
