# `goal-block`

operation `goal.block` · 코드 `src/ops/goal_block.rs`, 규칙 `src/goal_block.rs` · Issue #2708(옛 `tools/gen_goal_block.py`, #2163)

`goal-block` 은 표준 입력의 상태(`milestone`, 열린 `issues`·`prs` 각 `{number, state}`, `keep_set` 각 `{number, reason}`)로 `/goal` 붙여넣기 블록을 만든다. 결과는 리더의 milestone 조건 줄(`leader.goal_line`), 에이전트의 레인 전체 순환 줄(`agent.loop_line`, #2515)과 카드별 줄(`agent.goal_lines`), Codex 줄과 그 주의(`codex`), 수(`counts`), 제외 목록(`keep_set_excluded`)이다. 결과 객체는 옛 Python 도구가 출력하던 JSON 과 같다. 번호 순으로 정렬해 입력 순서와 무관하게 같은 블록을 낸다. 잘못된 입력(객체가 아님, 양의 정수가 아닌 `number`, 이유 없는 `keep_set` 등)은 `invalid fixture: …` 로 거부한다. 모델은 `/goal` 을 직접 입력할 수 없어 사용자가 붙여 넣는다(`model_can_input: false`). 다른 것을 읽거나 쓰지 않는다.
