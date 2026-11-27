# `lane-loop-line`

operation `lane.loop_line` · 코드 `src/ops/lane_loop_line.rs`, `src/lane_loop.rs` · Issue #2904 (#2899 B 절), SPEC-20261121-lane-goal-leader-codex-removed

`lane-loop-line --host claude --lane NAME [--repo OWNER/REPO] [--duty TEXT]` 은 레인이 리더 없이 스스로 계속 돌기 위해 붙여 넣을 줄을 낸다. 호스트는 Claude Code 하나다. `--host` 에 다른 값(`codex`, `omp` 포함)을 주면 명령줄 오류(`usage`, exit 2)다. 오류 문구는 `invalid argument value; see --help` 이고(거절한 값은 적지 않는다), `--help` 의 허용 값은 `claude` 하나다.

- 줄: `/loop 5m 레인 <lane> — <duty>`. `--duty` 가 없으면 "순환 계속, self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기". 고정 5분, 세션이 쉴 때만 발화하고 7일 뒤 만료(`verified: documented`). `/loop` 은 Claude Code 명령 문서(https://code.claude.com/docs/en/commands, 2026-11-26 확인)에 `/loop [interval] [prompt]` 로 실려 있다.

레인 이름은 1–64자 `A-Z a-z 0-9 . _ -`(`-` 로 시작하지 않음). `--repo` 는 주면 `OWNER/REPO` 형식인지만 검사한다. `--duty` 는 비지 않은 한 줄이다. 아무것도 읽거나 쓰지 않는다(`writes: false`).

`lane-stop-hook` 명령은 Codex 호스트와 함께 판정 기능이 없어졌다. 이전 릴리스가 `.codex/hooks.json` 에 쓴 Stop hook 이 아직 이 명령을 부르므로, 표준 오류에 한 줄을 내고 표준 출력 없이 exit 0 으로 끝난다(`docs/hooks-generation.md` 의 "레인 Stop hook" 절).
