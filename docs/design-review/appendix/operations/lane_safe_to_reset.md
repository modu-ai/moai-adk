# `lane-safe-to-reset`

operation `lane.safe_to_reset` · 코드 `src/ops/lane_safe_to_reset.rs`, `src/lane_reset.rs` · Issue #2669

`lane-safe-to-reset PROJECT --lane NAME [--screen-jobs N] [--session-cwd DIR] [--repo OWNER/REPO] [--observations FILE] [--now UTC]` 는 레인 세션을 비우거나(`/clear`) 카드를 다른 레인에 넘기기 전에, 그 레인의 일이 남아 있는지 본다. 읽기만 한다.

- 레인 → worktree 지도: `issue/<n>-…` 가지가 그 레인의 열린 claim Issue 인 것, 세션 폴더(`session_cwd`)를 품은 것, 폴더 이름에 레인 이름이 한 토큰으로 든 것. 각 worktree 에 `matched_by` 로 이유를 적는다.
- worktree 마다 미커밋(`git status --porcelain`), 미push(어느 원격 추적 ref 에도 없는 커밋 — fetch 하지 않음), 그 가지의 stash 수.
- 열린 claim 과 재개 메모(#3015 C1): 도구가 GitHub 에서 직접 읽는다(GET 과 검색만, 쓰기 없음). `repo:<o>/<r> is:open in:comments "session=<레인>"` 검색(한 쪽 30건)과, 레인 이름이 든 로컬 worktree 의 `issue/<n>-…` 가지가 가리키는 Issue(검색 색인이 새 claim 을 늦게 싣는 틈을 메움)의 댓글을 끝 페이지까지 읽어서, `lane-next` 와 같은 규칙(`moai-claim: session=<레인>` 이고 `moai-claim-release:`·`moai-claim-handoff:` 로 닫히지 않은 가장 새 것)으로 열린 claim 을 찾는다. 재개 메모는 같은 Issue 에서 `moai-resume:` 줄이 `lane_loop` 의 검사를 통과하고 `lane=`·`issue=`·`claim=` 이 맞고 claim 댓글과 **같은 작성자**가 쓴 가장 새 댓글이다. 읽기가 실패하면 `unknowns` 에 `claims could not be read from GitHub: <이유>`, 검색이 일부만 보여 주면(`total_count` > 반환 수, `incomplete_results`) `claims read only in part: <이유>` — 어느 쪽도 `safe` 로 올리지 않는다. 저장소는 `--repo`, 없으면 `origin`(GitHub 가 아니면 읽지 않고 이유를 적음). 관측 파일이 `claims` 를 주면 읽지 않고 그것을 쓴다.
- 화면과 세션 폴더: `--screen-jobs N`(지금 본 배경 작업 수, 관측 시각은 `--now`, 출처 `flag`)과 `--session-cwd DIR`. 같은 부분을 관측 파일에도 주면 "given twice" 로 거절한다. 관측 파일(호출자가 이전에 읽은 값을 출처·시각과 함께 넘김)의 화면 관측은 600초 넘으면 믿지 않는다. 세션 폴더의 자동 조회는 세션 등록(#3015 C3) 뒤에 한다.
- 관측 파일 오류와 `--help` 는 필드 타입과 전체 예시를 보여 준다(`OBSERVATIONS_TYPES`, `OBSERVATIONS_EXAMPLE`; 시험이 둘을 `--help` 문서와 같게 고정한다).
- 판정: 모든 검사가 돌고 깨끗할 때만 `safe`. 하나라도 남은 일이 있으면 `not_safe`(`findings`). 돌지 못한 검사가 있으면 `unknown`(`unknowns` 에 이유).
- 주 checkout(첫 worktree)이 세션 폴더라는 이유만으로 잡혔으면 그 변경은 레인 것이 아니다 — `unattributed` 에 적고 판정에 넣지 않는다(레인은 주 checkout 에서 시작해 연결 worktree 에서 일한다, #2669 6851881955).
