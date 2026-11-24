# `leader`

operation `leader.brief` · 코드 `src/ops/leader.rs`, `src/leader_brief.rs` · SPEC-20261124-leader-brief

`leader brief [<project>] [--home <dir>]` 는 리더가 깨어날 때 읽는 짧은 상태 하나를 찍어요. 원장과 레인 등록부만 읽고 아무것도 쓰지 않아요(`writes: false`). 새 계산은 없어요. `card orchestrate plan` 의 답에서 착지 대기열과 막힘과 결정만 골라 줄여요. 그래서 판정과 막힘 목록이 plan 과 어긋나지 않아요.

- 출력: `counts`(`queue`·`audit`·`blocked`·`decisions` 의 전체 개수), `queue`, `blocked`, `decisions`, `omitted`(잘려 나간 개수), `unread`(읽지 못한 원천 이름), `ledger`, `writes`. 봉투까지 합쳐 4096바이트 이하예요.
- `queue` 항목: `card`, 40자리 `head`, `ref`(그 head 를 통과시킨 판정, `note:<번호>` 또는 `event:<번호>`), 최신 head 가 아니면 `old: true`. 순서는 plan 의 착지 순서(우선순위, 카드 번호)예요.
- `blocked`·`decisions` 항목: `card`, `note`(노트 번호), `text`(첫 줄에서 표지 낱말을 뺀 글, 96바이트에서 글자 경계로 잘라요). 답이 없는 `LANE-BLOCKED` 노트 가운데 `kind=decision` 이 적힌 것과 `DECISION-QUESTION` 노트는 `decisions`, 나머지는 `blocked` 예요. 새 노트부터 보여 줘요. 항목을 열려면 `moai card show <카드>` 를 써요.
- 안 맞으면: 가장 긴 목록의 마지막 항목부터 덜어 내요. `counts` 는 전체 개수를 그대로 두고 `omitted` 가 덜어 낸 개수를 말해요.
- `unread`: 통합 브랜치(`integration`)나 레인 등록부(`registry`)를 읽지 못하면 그 이름이 들어가요. 읽지 못한 원천은 빈 대기열로 보이지 않아요.
- 원장이 없는 프로젝트는 빈 목록과 `ledger: "absent"` 로 끝나고(종료 0) 아무것도 만들지 않아요. 시계 값이 없어서 같은 원장은 같은 바이트를 찍어요.
- 착지 대기열 실행기(묶음·둘로 나누기)는 이 명령에 없어요. C-470 의 `card land-batch` 가 착지한 뒤 2조각이에요.
