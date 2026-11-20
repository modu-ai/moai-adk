# `audit-queue`

operation `audit.queue` · 코드 `src/ops/audit_queue.rs` · SPEC-20261120-audit-request-notes

`audit-queue [--project .] [--home DIR] [--stale-secs N] [--all]` 는 원장에서 **끝나지 않은 감사 요청**을 읽기만 해서 나이·상태·단계별 시간과 함께 돌려준다. 감사 요청은 라이브 통로가 아니라 카드의 로컬 노트이고, 노트의 **첫 줄** 모양이 넷이다: `AUDIT-REQUEST <card> head=<40-hex> tier=<0|1|2> scope=<text>`(그 head 의 감사 요청), `AUDIT-CLAIM <card> head=<40-hex> by=<name>`(감사자가 맡음), `AUDIT-VERDICT <card> head=<40-hex> verdict=<PASS|FAIL>`(판정), `REPAIR <card> head=<40-hex> of=<40-hex>`(FAIL 을 고친 새 head: 그 head 의 재감사 요청이기도 하다). `<card>` 는 노트가 달린 카드의 이름(`C-12`)이어야 하고 head 는 소문자 16진수 40자다.

각 요청의 `state` 는 `waiting`(맡은 감사자 없음), `running`(맡았고 판정 없음), `repair`(FAIL 인데 고친 head 가 아직 없음), `stale`(`waiting`·`running` 이 `--stale-secs`(1–604800, 기본 1800)보다 오래됨, `was` 가 떠난 상태를 말함)이다. 이미 판정이 있는 head 에 대한 뒤의 요청은 `reuse` 다: 같은 head 의 감사는 다시 돌리지 않는다. 카드의 감사자 `auditor` 는 그 카드의 **첫 claim** 의 `by` 이고, 고친 head 의 재감사는 같은 감사자에게 간다. `--all` 은 판정이 난 head(`passed`·`repaired`)도 보여 준다.

`times` 는 요청→claim, claim→판정, 판정→repair, repair→다음 claim 의 초를 **원장이 노트에 찍은 시각**(`card_note.created_at`)으로 계산한다(노트 글에 적은 시각은 읽지 않는다). 한쪽 끝 노트가 없으면 그 구간은 `null` 이다. `age_secs` 는 요청 노트부터, `state_age_secs` 는 지금 상태가 시작된 노트부터다. 네 낱말로 시작하지만 모양이 틀린 첫 줄은 `malformed`(`card`·`note_id`·`line`)에 올리고 세지 않는다. 쓰지 않는다(`writes: false`); 원장이 없으면 `ledger: "absent"`, `count: 0`, 종료 0, 아무것도 만들지 않는다. 낱말 규칙은 누가 감사하는가를 강제하지 않는다: 저자가 감사하지 않고, 정확한 head 를 감사하고, 고침은 같은 감사자에게 간다는 규칙은 워크플로 글이 말한다.
