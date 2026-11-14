# `lane-loop-line`

operation `lane.loop_line` · 코드 `src/ops/lane_loop_line.rs`, `src/lane_loop.rs` · Issue #2904 (#2899 B 절, 사용자 결정 D-O2)

`lane-loop-line --host claude|codex|omp --lane NAME [--repo OWNER/REPO] [--duty TEXT] [--project .]` 은 레인이 리더 없이 스스로 계속 돌기 위해 붙여 넣을 줄을 호스트별로 낸다.

- Claude Code: `/loop 5m /moai --mode-agent 레인 <lane> — <duty>`. `--duty` 가 없으면 "순환 계속, self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기". 고정 5분, 세션이 쉴 때만 발화하고 7일 뒤 만료(`verified: documented`).
- Codex CLI: `codex features enable goals`(셸) 다음 `/goal <끝 조건> — 레인 <lane>: 매 회차 $moai --mode-agent 로 self-pick (1)부터. <재개 메모 규칙>`(4,000자 이하). `--repo` 필수, `--duty` 거부. `stop_hook` 에 레인이 `.codex/hooks.json` 의 `Stop` 에 둘 명령(`moai-cli lane-stop-hook --lane <lane> --repo <repo>`)과, 프로젝트 파일에서 읽은 상태 `state`(`installed`·`missing`·`no_lane_hook`·`other_lane`·`unreadable` — 보통 파일만 열고 FIFO·폴더는 열지 않음, 1 MiB 까지)를 적는다. `trust` 는 Codex 자기 기록(`/hooks`)이라 늘 `unknown`. `installed` 가 아니면 `blocked_line` 이 `LANE-BLOCKED <lane> wake codex Stop hook <state> · kind=outside owner=user next=…`. 실제 `/goal`·Stop hook 세션 동작은 `verified: not_run`(실측 단계 따로).
- 세 번째 호스트: `--goal "<끝 조건>"`(`verified: not_run`).

끝 조건은 `goal-block` 의 `loop_line` 과 같은 문장(시험이 묶어 둠): "저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(HOLD·needs:decision 제외)". O2(#2901)의 공유 판정이 들어오면 그 자리만 바뀐다. 레인 이름은 1–64자 `A-Z a-z 0-9 . _ -`, 저장소는 `OWNER/REPO`. 프로젝트의 hook 파일만 읽고 쓰기 0(`writes: false`).

Stop hook 설치는 따로 미리보기 → 확인으로 한다: `hooks-plan --host codex --lane-loop <lane> --repo <OWNER/REPO> .` 으로 보고, 같은 플래그와 그 `plan_digest` 로 `hooks-apply`. 그러면 `.codex/hooks.json` 의 `Stop` 에 영수증 그룹 다음 두 번째 그룹(`moai-cli lane-stop-hook --lane <lane> --repo <repo>`, timeout 120초)이 생기고 이 명령의 `state` 가 `installed` 가 된다. 되돌리기는 같은 두 단계를 `--lane-loop-off` 로(기본 세트 바이트로 돌아감). Codex 에서 그 hook 을 `/hooks` 로 검토·신뢰하는 것은 사람의 단계다. `lane-stop-hook` 의 판정은 `docs/hooks-generation.md` 의 "레인 Stop hook" 절.
