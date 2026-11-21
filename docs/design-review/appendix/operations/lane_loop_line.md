# `lane-loop-line`

operation `lane.loop_line` · 코드 `src/ops/lane_loop_line.rs`, `src/lane_loop.rs` · Issue #2904 (#2899 B 절, 사용자 결정 D-O2)

`lane-loop-line --host claude --lane NAME [--repo OWNER/REPO] [--duty TEXT]` 은 레인이 리더 없이 스스로 계속 돌기 위해 붙여 넣을 줄을 낸다. 호스트는 Claude Code 하나뿐이다. `--host` 는 꼭 줘야 하고, `claude` 가 아닌 낱말은 `claude` 만 호스트라고 알리는 한 줄 오류(`unknown host …: claude is the only host`)와 0 이 아닌 종료 코드로 거부한다.

- Claude Code: `/loop 5m /moai --mode-agent 레인 <lane> — <duty>`. `--duty` 가 없으면 "순환 계속, self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기". 고정 5분, 세션이 쉴 때만 발화하고 7일 뒤 만료(`verified: documented`).

끝 조건은 `goal-block` 의 `loop_line` 과 같은 문장(시험이 묶어 둠): "저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(HOLD·needs:decision 제외)". O2(#2901)의 공유 판정이 들어오면 그 자리만 바뀐다. 레인 이름은 1–64자 `A-Z a-z 0-9 . _ -`, 저장소는 `OWNER/REPO`. 아무것도 읽거나 쓰지 않는다(`writes: false`).
