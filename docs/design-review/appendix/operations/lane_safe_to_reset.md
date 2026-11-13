# `lane-safe-to-reset`

operation `lane.safe_to_reset` · 코드 `src/ops/lane_safe_to_reset.rs`, `src/lane_reset.rs` · Issue #2669

`lane-safe-to-reset PROJECT --lane NAME [--observations FILE] [--now UTC]` 는 레인 세션을 비우거나(`/clear`) 카드를 다른 레인에 넘기기 전에, 그 레인의 일이 남아 있는지 본다. 읽기만 한다.

- 레인 → worktree 지도: `issue/<n>-…` 가지가 그 레인의 열린 claim Issue 인 것, 세션 폴더(`session_cwd`)를 품은 것, 폴더 이름에 레인 이름이 한 토큰으로 든 것. 각 worktree 에 `matched_by` 로 이유를 적는다.
- worktree 마다 미커밋(`git status --porcelain`), 미push(어느 원격 추적 ref 에도 없는 커밋 — fetch 하지 않음), 그 가지의 stash 수.
- 관측 파일(호출자가 forge·터미널에서 읽어 출처와 함께 넘김): 열린 claim 과 재개 메모, 세션 폴더, 화면의 배경 작업 수와 관측 시각(600초 넘으면 믿지 않음).
- 판정: 모든 검사가 돌고 깨끗할 때만 `safe`. 하나라도 남은 일이 있으면 `not_safe`(`findings`). 돌지 못한 검사가 있으면 `unknown`(`unknowns` 에 이유).
