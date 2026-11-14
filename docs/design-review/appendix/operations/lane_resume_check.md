# `lane-resume-check`

operation `lane.resume_check` · 코드 `src/ops/lane_resume_check.rs`, `src/lane_loop.rs` · Issue #2904

`lane-resume-check --file BODY.md` 는 저장한 댓글 본문에서 재개 메모(`lane-resume-note` 형식)를 읽는다. 줄 시작의 `moai-resume:` 줄이 정확히 하나여야 하고(인용 `> ` 줄은 메모가 아님, 둘이면 거부), 그 뒤로 이어지는 `next:`·`wait:` 줄까지가 메모다. 답: `ok: true` 와 `note`(필드, `next`, `wait`), 또는 `ok: false` 와 `problems`(메모 없음·둘·빠진 필드·모르는 필드·uncommitted 0 아님·짧은 head 등). 보통 파일만 읽고, 1 MiB 를 넘거나 UTF-8 이 아니면 일부만 읽지 않고 거절. 쓰기 0. compact·clear 뒤 레인의 첫 행동.
