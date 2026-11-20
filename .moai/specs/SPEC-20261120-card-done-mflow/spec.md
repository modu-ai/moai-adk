---
id: SPEC-20261120-card-done-mflow
title: "card done accepts a local M-Flow landing as the closing evidence (no pull request)"
status: draft
revision: 1
created: 2026-11-20
---

# card done — M-Flow 에서는 로컬 착지 기록으로 닫는다

## Intent

사용자 질문(2026-11-20, "왜 카드는 완료가 안 되는가?" — statusline 에 active/review 가 많음)의
원인: `card done` 이 병합된 PR 링크를 요구하는데 M-Flow(사용자 결정 2026-11-20)는 PR 없이 로컬
통합 브랜치에 병합해요 — 로컬로 들어간 카드는 영원히 못 닫아요. m-flow 프로젝트에서는 카드의
착상 head 가 프로필의 통합 브랜치 tip 의 조상이면 done 을 허용하고, 다른 흐름은 기존 병합-PR
규칙을 그대로 유지해요.

## Non-goals

- card_link 표의 kind 확장 — CHECK 가 네 종류에 고정(blocks·split_of·relates·pr); 착상 기록은
  자유 event kind 로(card.withdraw 선례).
- 릴리스 PR 이 카드들을 소급 링크하는 절차 — 선택 사항, 후속.
- reclaim 규칙(4시간 임대·한도 3) 변경 없음.

## Requirements

- REQ-1: When a card in state review carries a `card.landed` event, the `card landed` command shall record one `card.landed` event whose payload names the head (40-hex) and the by, with the writing session as the actor · failure: none — another state, a head that is not 40-hex, or an unknown card leaves every row as it was.
- REQ-2: When `card done` runs on a project whose profile work_flow is m-flow with an integration branch and the newest `card.landed` event's head is an ancestor of that branch's tip (checked with `git merge-base --is-ancestor` at done time), it shall close without a merged pull request · failure: none — with no landed event, a non-ancestor head, or no such branch, the refusal names the rule and nothing is written.
- REQ-4: When `card done` closes a card on the local landing evidence, the `card.done` event shall record the landed head and the integration branch tip · failure: none — no mflow key when the merged-pull-request rule closed the card.
- REQ-3: When the project's profile names no m-flow work_flow, done shall keep the merged-pull-request rule exactly as before · failure: none — a profile without work_flow or with another flow behaves as before, and the done help names both rules.

## Acceptance

- AC-1 (REQ-1): 원장 시험 — review 카드의 landed 이벤트 기록(payload head·by·actor), 다른 상태·잘못된 head 거절. verify: `cargo test --locked --offline --test card_done_mflow_cli`
- AC-2 (REQ-2, REQ-4): 명령 시험(임시 git 저장소) — m-flow 프로필 + landed head 가 통합 브랜치에 포함 → done 성공(card.done 이벤트에 head·tip); 비포함 head → 규칙 문장 거절. verify: `cargo test --locked --offline --test card_done_mflow_cli`
- AC-3 (REQ-3): 명령 시험 — work_flow 없는 프로젝트는 병합-PR 규칙 유지(착상 기록 있어도 done 거절). verify: `cargo test --locked --offline --test card_done_mflow_cli`
- AC-4: 이웃 검사. verify: `cargo test --locked --offline --test ledger_workflow`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ledger/workflow.rs`(land + done 가드), `src/ledger/schema.rs`(card_done_guard 트리거 본문 — card.landed 예외 추가), `src/card/work.rs`(done 명령의 m-flow 검사 + landed 명령), `src/card/dispatch.rs`(라우팅), `src/ops/card.rs`(Land 변형·도움말), `tests/card_done_mflow_cli.rs`(신설), `tests/ledger_workflow.rs`(done 호출 시그니처), `tests/prop_reservation.rs`(동일), `.moai/specs/SPEC-20261120-card-done-mflow/**`
- read-only: `src/profile.rs`(work_flow 읽기), `tests/common/**`
- forbidden: card_link·event 스키마 변경, 실제 HOME, 카드 번호를 SPEC id·가지·커밋에 사용

## Plan

1. RED: tests/card_done_mflow_cli.rs — landed 기록·m-flow done 성공·비조상 거절·타 흐름 유지.
2. GREEN: workflow::land + done(landing 증거 인자), work.rs done 명령의 프로필·merge-base 검사, dispatch·ops 배선, 도움말 문장.
3. 이웃 검사 + READY 폭 검사.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 착상 기록은 card_link 가 아니라 자유 event kind `card.landed` | card_link kind 는 CHECK 로 4종 고정(스키마 변경은 사용자 결정) | 링크 kind 추가(스키마 변경) |
| D-2 | 조상 검사는 done 시점에 명령층이 git 으로 실행(레저는 git 을 부르지 않는 계약) | reservation.rs 머리글 "calls no git"; 증거를 done 이벤트에 기록 | 레저가 git 실행 |
| D-3 | 허용 상태는 review — READY 뒤 착상이 흐름의 순서 | MERGED-LOCAL 은 감사 PASS 뒤 일어남 | draft/ready 도 허용 |
| D-4 | card_done_guard 트리거 본문에 card.landed 이벤트 존재 예외 추가 — 테이블·칼럼·kind·상태값 무변경 | 레이어 가드만으론 DB 트리거가 merged-PR 없는 done 을 ABORT(시험이 잡음); 조상 검사는 명령층이 이미 수행 | 트리거 원복(m-flow done 불가) |
| D-4 | 거절 문장: "cannot close: no local landing is linked (`card landed` after the merge)" · "cannot close: the landed head is not contained in the integration branch <branch>" | 기준 1 — 규칙이 이유를 말함 | 문장 변경 |

## Risks

- 프로필에 m-flow+integration_branch 가 없는 프로젝트는 기존 규칙 그대로 — 기존 사용자 후퇴 없음.
- head 만으로 검사(머지 커밋은 head 의 자손이라 --no-ff·ff 모두 커버).
