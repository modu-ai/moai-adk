# `leader-run`

operation `leader.run` · 코드 `src/ops/leader_run.rs` · SPEC-20261121-leader-run-lease

`leader-run begin|end|status <project>` 는 프로젝트마다 감독 중인 리더 실행을 하나로 지키는 임대(lease)를 다뤄요. 임대는 프로젝트 상태 폴더의 `state/leader-run.lease` 파일이고, 한 번에 한 호출만 판단하도록 같은 폴더의 `leader-run.guard` 파일 잠금을 호출 동안 잡아요. 설정·승인·권한은 건드리지 않고 이 두 파일 밖에는 쓰지 않아요.

- `begin --session S --ready N [--hold] [--lease-secs 600] [--now UNIX]`: 다른 세션의 유효한 임대가 없고 준비 카드가 1장 이상이며 HOLD 가 없으면 `taken`(같은 세션의 유효한 임대는 `renewed: true` 로 연장, 끝난 임대는 인수하고 `replaced` 에 옛 임대). 다른 세션의 임대가 유효하면 `skipped` 와 `holder`·`age_secs`·`lease_until`, 아무것도 바꾸지 않아요. `--ready 0` 또는 `--hold` 면 임대를 잡거나 건드리지 않고 `idle` 과 `wake_on`(`card_ready` 또는 `hold_lifted`). 임대 끝 시각 `lease_until` 과 같은 순간까지는 유효해요.
- `end --session S`: 그 세션이 보유한 임대만 지워요(`released: true`). 다른 세션의 임대는 거절하고 그대로 둬요(`invalid_input`). 임대가 없으면 오류가 아니라 `released: false`.
- `status [--now UNIX]`: 읽기만. `lease`(`session`·`started`·`lease_until`·`age_secs`·`ended`) 또는 null. 잠금 파일도 만들지 않아요.
- 입력 검사: 세션은 1–80자·제어 문자 없음, 임대 길이는 60–86400초(기본 600). 읽을 수 없는 임대 파일은 덮어쓰지 않고 거절해요.
- 중복 배차는 카드 원장의 세대 울타리가, 중복 병합은 병합 잠금과 fast-forward 전용이 이미 막아요. 이 명령은 실행 자체의 임대만 맡아요.
