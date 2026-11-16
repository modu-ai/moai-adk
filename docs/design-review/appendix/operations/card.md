# `card`

operation `card` · 코드 `src/ops/card.rs`, `src/card/command.rs` · Issue #3070 C1 (PLAN 댓글 6905829690, 명령 모양 6898149938)

`card <동작>` 은 프로젝트 로컬 원장(#3068)의 카드를 읽고 쓴다. 낱말 하나에 동작이 중첩돼 있고 봉투의 `operation` 은 항상 `card`, 동작은 결과의 `action` 으로 나간다. 모든 동작의 도움말 첫 문장은 `Reads the ledger only.` 또는 `Writes the ledger.` 이다(시험이 묶는다). **원장 파일과 폴더는 `ledger-register-apply` 만 만든다**; `card` 의 어떤 동작도 만들지 않는다.

| 동작 | 원장을 | 입력 | 결과 |
| --- | --- | --- | --- |
| `card list [PROJECT] [--state S] [--home DIR]` | 읽기만(`open_read_only`) | 상태는 `draft`·`ready`·`active`·`review`·`done`·`hold` | `ledger`(`ready`\|`absent`), `count`, `cards[]`(`name`=`C-n`, `card_id`, `title`, `kind`, `state`, `priority`, `rev`, `updated_at`, `source`). 원장이 없으면 `cards: []`, `ledger: "absent"`, 종료 0, 아무것도 만들지 않음 |
| `card show C-n\|#번호 [--project .] [--home DIR]` | 읽기만 | `#번호` 는 호스트의 번호이고 여러 카드가 같으면 거절 | 카드 전체 + `criteria[]`(seq 순) + `notes[]`(note_id 순) + `links[]`(link_id 순) |
| `card add --title T [--body B] [--kind K] [--priority P] [--scope PATH]... [--ac TEXT]... [--by NAME] [--project .] [--home DIR]` | 씀(`with_write_tx` 한 번) | 검사는 `card::input::validate_add` (제목 1–300자 한 줄, 종류·우선순위 목록, 쓰기 범위는 `/` 상대 경로, 기준은 한 줄·1부터 번호) | `name`, `card_id`, `state: "draft"`, `criteria`, `rev`, `created_by`, `events: 1`, `writes: true`. 카드·기준·`card.add` 사건이 한 트랜잭션이라 하나라도 거절되면 아무것도 안 남음 |

- `rev` 는 기준을 하나 넣을 때마다 스키마 트리거(`card_touch_ac_insert`)가 올려서, 기준이 n 개인 새 카드는 `1 + n` 이다.
- 시각은 모두 트랜잭션의 DB 시계 값(`unixepoch()`)이다. `--by` 기본값 `user`(1–80자 한 줄).
- 오류: 원장이 없으면 `not_found`(안내 문장에 `ledger-register-plan` → `ledger-register-apply`), 쓸 수 없는 원장·WAL·낯선 파일은 `blocked`, 거절된 입력은 `card <동작> refused: …` (`invalid_input`), 쓰기 잠금을 오래 못 잡으면 `busy`. 새 `kind` 를 만들지 않는다.
| `card ready C-n [--by NAME]` | 씀(트랜잭션 한 번) | `draft`·`hold` 카드만 | `from`, `state: "ready"`, `rev`, `events: 1`. 카드 상태 변경과 `card.ready` 사건이 한 트랜잭션 |
| `card take C-n --holder H [--lease-secs N] [--scope P]... [--worktree W] [--branch B]` | 씀(`reservation::reserve`) | `--scope` 를 안 주면 카드의 `write_scope`; lease 60–86400초(기본 14400) | `reservation`(`generation`·`holder`·`lease_until`·`write_scope`…), `state: "active"`. `ready` 가 아니거나 이미 잡혔거나 범위가 다른 카드와 겹치면 `conflict` |
| `card release C-n --holder H --generation G` | 씀(`reservation::release`) | 세대는 `take`·`handoff` 가 답한 값 | `reservation.status: "released"`, `state`(`active` 였으면 `ready`). 옛 세대·다른 holder 는 `conflict`(사건 `fence_reject` 한 줄) |
| `card handoff C-n --holder H --generation G --to H2` | 씀(`reservation::handoff`) | 살아 있는 lease 필요, `--to` 는 holder 와 달라야 함 | 새 `reservation`(세대 + 1, 같은 범위·worktree·branch·lease 끝), `from_holder`, `state: "active"` |

- 예약 규칙(세대·lease·쓰기 범위·울타리)은 #3068 P3(`src/ledger/reservation.rs`, `ledger_reservation.md`)의 것이다. `ready`·`take`·`release`·`handoff` 는 `C-n` 으로만 부르고(`#번호` 는 거절), P3 의 거절 문장이 그대로 나가 계약 분류가 `not_found`(`card not found`)·`conflict`(`conflict in`)·`invalid_input`(`invalid reservation input`)을 정한다.
- 아직 없는 동작: `edit`·`note`·`renew`·`review`·`done`·`export`·`restore`·`backup`.
