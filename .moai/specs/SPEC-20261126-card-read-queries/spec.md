---
id: SPEC-20261126-card-read-queries
title: "리더의 장부 읽기를 읽기 전용 moai card query 명령 하나로 바꿔요"
status: draft
revision: 1
created: 2026-11-26
---

# 리더의 장부 읽기를 읽기 전용 moai card query 명령 하나로 바꿔요

## Intent

리더는 라운드마다 장부(ledger) 파일을 sqlite3 로 직접 읽어요: 여러 카드의 최신 판정과 head, 예약을 쥔 카드와 워크트리, 노트 번호로 본문, 카드별 열린 기준이에요. `moai card list`, `show`, `timeline` 은 한 카드씩이거나 이 값을 주지 않아요. 원시 SQL 은 제품 계약 밖이라 스키마가 바뀌면 조용히 틀리고, 셸 명령 꼴 때문에 워크트리 가드 거절도 자주 나요. 읽기 전용 명령 `moai card query` 하나가 이 네 가지를 JSON 으로 답해요.

## Non-goals

- MCP 서버로 감싸지 않아요. 이 명령이 생긴 뒤 따로 정해요.
- 장부에 아무것도 쓰지 않고, 기존 `card` 명령의 출력은 바꾸지 않아요.
- 판정 이력 전체는 답하지 않아요. 한 카드의 이력은 `card timeline` 이 이미 줘요.
- 노트 본문의 비밀값을 가리지 않아요. `card show` 처럼 저장된 그대로 내요(쓸 때 검사해요).
- `skills/moai/**` 의 글은 바꾸지 않아요.
- 설치된 `moai` 갱신은 이 SPEC 밖이에요. 착지 뒤의 후속 일이에요.

## Requirements

- REQ-1: When `moai card query` 가 `--card C-n[,C-m]` 로 실행되면, the query command shall 이름을 준 카드마다 `card`, `card_id`, `title`, `state`, `head`(카드의 지금 일에 속한 가장 새 hand-back 노트가 말하는 head, 없으면 null), `verdict`(head 와 상관없이 가장 새 `card.verdict` 사건, 없으면 null), `head_verdict`(head 가 `head` 와 같은 가장 새 판정, 없으면 null), 기준 상태별 개수 `criteria`, 열린 기준 `open_criteria`(`seq`, `text`, `verify`)를 JSON 으로 답한다.
- REQ-2: When `--reservations` 를 주면, the query command shall 상태가 `held`, `expired_candidate`, `held_for_review` 인 지금 예약을 카드 번호 순으로 모두 답하고, 줄마다 `card`, `title`, 카드 `state`, 예약 `status`, `session`, `generation`, `acquired_at`, `lease_until`, `lease_ended`, `worktree`, `branch` 를 담는다 · failure: none — 살아 있는 예약이 없으면 빈 목록이다.
- REQ-3: When `--note <번호>[,<번호>]` 를 주면, the query command shall 노트마다 잘리지 않은 본문 전체와 `note_id`, `card`, `card_id`, `kind`, `author`, `origin`, `session`, `ref`, `created_at`, `bytes` 를 답한다.
- REQ-4: When `--notes N` 을 `--card` 와 함께 주면, the query command shall 카드마다 가장 새 노트 N 개를 새것부터, 본문 없이 머리 항목과 200자에서 자른 첫 줄 `line` 으로 담는다 · failure: none — 노트가 없는 카드는 빈 목록이다.
- REQ-5 (REQ-1): IF 한 카드의 `card.verdict` 행 가운데 판정으로 읽히지 않는 것이 있으면, THEN the query command shall 그 카드의 `verdict` 와 `head_verdict` 를 null 로 하고 `verdicts_unreadable` 에 못 읽은 행의 수를 담아, 옛 판정이 최신으로 보이지 않게 한다.
- REQ-6: The query command shall 아무것도 쓰지 않는다: 장부를 읽기 전용으로 열어 읽기 트랜잭션 하나에서 모든 값을 모으고, 도움말 첫 문장은 `Reads the ledger only.` 이며, 장부가 없으면 다른 `card` 읽기 명령과 같은 안내로 거절하고 파일을 만들지 않는다.
- REQ-7 (REQ-1, REQ-3): IF 입력이 틀리면(고르는 옵션이 하나도 없음, 없는 카드나 노트, `#n` 꼴, `--card` 없는 `--notes`, 카드 50개·노트 번호 10개·`--notes` 20 초과, `--notes` 가 0 미만이거나 수가 아님), THEN the query command shall 호출 전체를 0 이 아닌 종료 코드로 거절하고 부분 답을 내지 않는다.
- REQ-8: The leader rules (`docs/leader-rules.md` 의 "What a round reads") shall 장부 읽기에 `moai card query` 와 그 옵션 `--card`, `--reservations`, `--note` 를 이름으로 적고 리더가 장부 파일을 직접 열지 않는다고 적으며, 명령 표 `docs/design-review/appendix/operations/card.md` 에 `card query` 줄을 둔다.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 한 번의 호출이 카드 둘의 head 와 판정, 열린 기준, 지금 예약, 노트 하나의 본문을 함께 답해요 · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::one_query_/)'`
- AC-2 (REQ-1, REQ-5): 옛 head 에 늦게 적힌 판정이 있으면 `verdict` 와 `head_verdict` 가 서로 다르게 나오고, 못 읽는 판정 행이 있으면 둘 다 null 이며 수가 나와요 · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::verdict_/)'`
- AC-3 (REQ-2): `held`, `expired_candidate`, `held_for_review` 예약은 나오고 `released`, `reclaimed` 예약은 나오지 않아요 · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::reservations_/)'`
- AC-4 (REQ-3, REQ-4): 번호로 고른 노트는 본문 전체가 바이트 그대로 나오고, `--notes N` 은 새것부터 N 개의 머리 항목과 자른 첫 줄만 담아요 · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::note_/)'`
- AC-5 (REQ-7): REQ-7 의 틀린 입력마다 거절하고 표준 출력에 부분 답이 없어요 · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::refusal_/)'`
- AC-6 (REQ-6): 호출 앞뒤로 임시 HOME 의 파일이 그대로이고, 장부가 없으면 파일을 만들지 않고 거절하며, 도움말이 읽기 전용이라고 말해요 · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::readonly_/)'`
- AC-7 (REQ-8): "What a round reads" 에 `moai card query` 와 세 옵션이 있고 `docs/leader-rules.md` 에 `sqlite` 라는 낱말이 없으며, 명령 표에 `card query` 줄이 있어요 · verify: `grep -n -- "moai card query" docs/leader-rules.md`
- AC-8 (REQ-1, REQ-6): 이 명령의 테스트는 모두 임시 HOME 의 임시 장부에서 돌아요(실제 HOME 의 장부를 열지 않아요) · verify: `cargo nextest run --offline --locked --test card -E 'test(/^card_query_cli::/)'`

## Scope

- write: `.moai/specs/SPEC-20261126-card-read-queries/spec.md`, `src/card/query.rs` (new), `src/card/mod.rs`, `src/card/command.rs`, `src/card/wip_limit.rs`, `src/ledger/reservation.rs`, `tests/card/card_query_cli.rs` (new), `tests/card/main.rs`, `docs/leader-rules.md`, `src/ops/card.rs` (shared, line lease), `docs/design-review/appendix/operations/card.md` (shared, line lease)
- read-only: `src/ledger/workflow.rs`, `src/card/timeline.rs`, `src/card/inbox.rs`, `skills/moai/references/audit.md`
- forbidden: `skills/moai/**`, `tools/repo_scans.json`, the real HOME and its `.moai/projects` folder

## Plan

1. 테스트를 먼저 써요(RED): `tests/card/card_query_cli.rs`.
2. 기존 읽기를 함께 쓸 수 있게 열어요: 예약 목록, 기준과 노트 세션 읽기, 현재 head.
3. `src/card/query.rs` 를 쓰고 `src/ops/card.rs` 에 하위 명령을 이어요.
4. 문서 두 곳을 고쳐요.
5. `cargo fmt --all`, 이름을 지정한 테스트, 영수증, READY.

## Risks

- 장부 스키마가 바뀌면 `query.rs` 의 SQL 이 틀릴 수 있어요. 테스트가 실제 스키마의 임시 장부를 써서 드러내요.
- 노트 본문을 자르지 않아 출력이 클 수 있어요. 한 호출의 노트 번호를 10개로 묶어요.
- 카드마다 head 를 구하려고 노트를 훑어요. 카드를 50개로 묶고, 속도는 재지 않았어요.
- 설치된 `moai` 에는 갱신 전까지 이 명령이 없어요. 문서가 먼저 닿으면 리더는 갱신까지 기존 읽기를 써요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 명령은 여럿이 아니라 `card query` 하나이고 옵션으로 부분을 골라요 | 기준 1 "한 읽기 전용 명령"; 구현 레인의 범위 안 결정 | 커밋 되돌리기 |
| D-2 | 옵션은 `--card`, `--reservations`, `--note`, `--notes N`, `--project`, `--home` 이고 앞의 셋 중 하나는 꼭 필요해요 | 위와 같음 | 커밋 되돌리기 |
| D-3 | `verdict`(가장 새 판정)와 `head_verdict`(지금 head 의 판정)를 둘 다 내요. 착지 여부는 `card land` 가 정해요 | 옛 head 의 늦은 FAIL 과 지금 head 의 PASS 를 구별해야 해요 | 커밋 되돌리기 |
| D-4 | head 는 작업 상한이 쓰는 정의(`wip_limit` 의 current head)를 그대로 써요 | 같은 값을 두 방식으로 정하지 않아요 | 커밋 되돌리기 |
| D-5 | 못 읽는 판정 행이 있으면 판정을 null 로 해요 | 조용히 옛 판정을 최신으로 보이지 않아요 | 커밋 되돌리기 |
| D-6 | 예약 목록은 지금 살아 있는 세 상태이고 쓰기 범위는 담지 않아요 | 리더가 필요한 값(카드·상태·워크트리·세션)만 | 커밋 되돌리기 |
| D-7 | 상한은 카드 50, 노트 번호 10, `--notes` 20 이고 넘으면 거절해요 | 출력 크기와 읽기 시간 | 커밋 되돌리기 |
| D-8 | 판정 이력 옵션(`--verdicts N`)은 넣지 않아요 | `card timeline` 이 이력을 줘요; 줄이기가 먼저예요 | 옵션 추가 |
| D-9 | 문서 글을 고정하는 테스트는 두지 않고 verify 명령과 감사 읽기로 확인해요 | 공용 파일 `tools/repo_scans.json` 을 건드리지 않아요 | 테스트와 scan 줄 추가 |
| D-10 | 기준 2는 "What a round reads" 에 `card query` 를 더하고 `sqlite` 낱말이 없게 두는 것으로 채워요(그 문서에는 원래 sqlite 줄이 없어요) | 읽은 사실; 리더에게 확인을 요청했어요 | 리더 답에 따라 고침 |
