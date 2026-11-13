# `gate-loop`

operation `gate.loop` · 코드 `src/ops/gate_loop.rs`, `src/gate_loop.rs` · Issue #2674 S4

`gate-loop --once --permission-reference REF [--repo OWNER/REPO] [--base v2026.11.0] [--dir .] [--tip …] [--review-confirmed …]` 는 리더 병합 루프 한 바퀴다. 통합 가지로 가는 열린(초안 아닌) PR 마다 `gate-merge` 를 돌리고(게이트를 통과한 것만 병합), 멈춘 이유가 G6 "코드 파일 겹침"이면 head 마다 한 번 update-branch(`expected_head_sha` 고정)와 감사자 안내를, G2(판정 쓴 레인이 허용된 예약 보유자가 아님)이면 head 마다 한 번 감사자 안내를 남긴다. "head 마다 한 번"은 PR 댓글 안의 숨은 표지 `<!-- moai:gate-loop <종류> head=<sha> -->` 로 판단하고, 로컬 상태 파일은 쓰지 않는다. head 가 로컬에 없는 PR 은 건너뛴다. 반복은 호출하는 쪽(리더, 도구의 스케줄러)이 한다.
