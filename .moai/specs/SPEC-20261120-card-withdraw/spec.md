---
id: SPEC-20261120-card-withdraw
title: "card withdraw — a withdrawn card cannot be taken and hides from the default list"
status: completed
approval: {by: user, at: 2026-11-20, basis: "in-session selection 2026-11-20: 세션 병목 리포트 개선 'all' 선택(카드 본문 기록)이 withdraw 기능을 고름 — 리더 배차로 SPEC 작성, tier 1 감사가 요구·수용 기준과 구현의 일치를 확인", digest: "sha256:66705f2ab97240b5e56d2c7609a1af6c7f82c6cd4fde372275ea48e74bd7d6a9"}
revision: 1
created: 2026-11-20
---

# 카드 철회 — withdrawn 카드는 take 가 거절되고 기본 목록에 숨는다

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드 본문의 사용자 'all' 선택(2026-11-20)이 요구를 정했고, 별도 사용자 인터뷰는 없었어요.

## Intent

오늘 레인이 철회된 카드를 세 번 집으려 했어요. `card withdraw` 가 카드에 철회 표시를 남기면 take 가
이유를 적어 거절하고 `card list` 기본 목록이 숨겨요. 이 조각은 원장(workflow·reservation)과 읽기
층(command), 그리고 명령 배선(`src/ops/card.rs` 의 withdraw 액션·`--all` 플래그 — 그 파일의 예약이
조각 중간에 풀려 함께 넣었어요)을 담아요. `src/card/dispatch.rs` 는 여전히 다른 카드의 예약 중이라
다음 조각이에요.

## Non-goals

- 스키마 변경 — 아래 REQ-4 의 이유로 불가(사용자 결정 사항).
- `card unwithdraw` 같은 되돌림 명령 — 후속.
- `card next` 의 표시 변경 — hold 카드는 이미 제안에서 빠져요.
- `src/card/dispatch.rs` 편집 — 다른 카드의 예약 중(이 SPEC Scope 의 blocked).

## Requirements

- REQ-1: When the ledger withdraw runs on a card in state draft or ready with no live reservation, it shall move the card to hold and record one `card.withdraw` event carrying the reason and the by, with the writing session recorded as the actor · failure: none — a refusal (unknown card, another state, a live reservation, input out of bounds) leaves every row as it was.
- REQ-2: When take runs on a card carrying a `card.withdraw` event, the refusal shall name the withdrawal reason inside the shared `reservation conflict in card` message, whose envelope kind stays `conflict` as classify reads it · failure: none — a hold card without a withdraw event keeps the plain refusal wording.
- REQ-3: When card list runs without a state filter, it shall omit cards carrying a `card.withdraw` event and change nothing in the ledger · failure: none — an explicit state filter behaves as before.
- REQ-4: When a card is withdrawn, the ledger schema shall stay unchanged — the withdrawn card reuses the existing `hold` state value (one of the six values the card table CHECK pins), an existing `card_transition` row (`ready→hold`, `draft→hold`), and the free-form event kind `card.withdraw` · failure: none — a seventh state value would need a table rebuild, which is a user decision, so none is made.

## Acceptance

- AC-1 (REQ-1, REQ-4): 원장 시험 — draft·ready 는 hold 로 + 이벤트(사유·by·작성 세션), 나머지는 전부 거절·무변화, 입력 한계 N/N+1.
  verify: `cargo test --locked --offline --test ledger_workflow withdraw`
- AC-2 (REQ-2): 명령 시험 — 철회된 카드의 take 는 종료 1·kind conflict(공유 접두의 classify)·메시지에 사유; 철회 없는 hold 카드는 기존 문구 그대로.
  verify: `cargo test --locked --offline --test card_withdraw_cli`
- AC-3 (REQ-3): 명령 시험 — 같은 프로젝트에서 철회 카드는 기본 목록에 없고 다른 카드는 있으며, 목록 실행 전후 원장 파일이 같다.
  verify: `cargo test --locked --offline --test card_withdraw_cli list`
- AC-4: 이웃 검사.
  verify: `cargo test --locked --offline --test ledger_workflow`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_cli`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ledger/workflow.rs`, `src/ledger/reservation.rs`, `src/card/command.rs`, `src/ops/card.rs`(withdraw 액션·list `--all`), `tests/ledger_workflow.rs`, `tests/card_withdraw_cli.rs`, `.moai/specs/SPEC-20261120-card-withdraw/spec.md`
- read-only: `src/ledger/schema.rs`, `tests/common/**`
- blocked(다음 조각): `src/card/dispatch.rs`(비밀 화면 후속 카드의 예약 중 — 조각 중간까지 예약이던 `src/ops/card.rs` 는 풀려 write 로 옮겨졌어요)
- forbidden: 스키마·표·CHECK 변경, 실제 원장·실제 HOME, 네트워크, 카드 번호를 SPEC id·가지·커밋에 사용

## Plan

1. RED: `tests/ledger_workflow.rs` 에 withdraw 원장 시험, 새 `tests/card_withdraw_cli.rs` 에 take 거절 사유·목록 숨김 시험.
2. GREEN: `workflow.rs` withdraw 함수, `reservation.rs` 의 CardNotReady 에 철회 사유, `command.rs` list 의 기본 제외.
3. 이웃 검사(AC-4)와 가벼운 검사 — READY 노트에 운용 규칙대로 종료 코드를 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 철표시 = 기존 `hold` 상태 + 자유 event kind `card.withdraw`(사유·by 를 payload 로) | 스키마 CHECK 가 여섯 상태 값만 받음 — 일곱째는 표 재건(사용자 결정) | 상태 값 추가(스키마 변경) |
| D-2 | 철회는 draft·ready 만, 살아 있는 예약이 있으면 거절 | 철회는 집기 전 단계의 일; hold 이후는 이미 집지 않음 | review 도 허용 |
| D-3 | 목록 기본 제외는 list 질의의 NOT EXISTS(subquery), `--state` 필터가 있으면 예전대로 | 기본 목록만 요구사항 | 애플리케이션 층 필터링 |
| D-4 | take 거절 사유는 ReserveError 의 CardNotReady 가 new 필드로 갖고 온다(표시는 Display) | 거절 경로의 메시지는 Display 로 나감 | 명령층 재조회 |

## Risks

- CLI 배선이 늦으면 철회 표시를 내는 길이 원장 함수뿐이에요 — 다음 조각이 첫 일이 되도록 SPEC 에 이름을 박았어요.
- 목록 제외 subquery 가 event 표를 카드마다 읽어요 — 목록은 카드 수 규모라 지금은 문제가 안 되고, event 에 card_id 색인 문제는 통합 가지 타임라인 지표(기준 문서 5.1절)가 지켜봐요.

## Result

- AC-1 met — ledger_workflow withdraw 3종(성공 두 상태·거절 셋·입력 한계) 통과
- AC-2 met — card_withdraw_cli take 사유·kind conflict·평범 hold 문구 유지
- AC-3 met — 목록 숨김·--all·목록 전후 스냅샷 동일
- AC-4 met — 이웃 검사(fmt·clippy·ledger_workflow·card_dispatch_cli·card_cli·spec_cli·home_isolation_scan)

base `aa06a0f6ef`(로컬 v2026.11.0), worktree `~/.moai/worktrees/moai-adk/c188`(새 모양), heavy-slot `--max 4`.

- RED: 구현 전 원장·명령 시험이 `withdraw` 부재로 컴파일 실패 — `cannot find function withdraw in module moai_cli::ledger::workflow`(E0425)·`could not find WithdrawError`(E0433), 기록 뒤 복원.
- GREEN: ledger_workflow 29 passed 0 failed(신설 3: 성공 두 상태·거절 셋·입력 한계), card_withdraw_cli 3 passed 0 failed(take 사유·목록 숨김·--all·평범 hold 문구 유지).
- 계약 교정 두 번: (1) 봉투 거절 kind 는 `card_not_ready` 가 아니라 공유 접두 `reservation conflict in card` 의 classify 인 `conflict` — 시험과 REQ-2/AC-2 를 실제 계약에 맞춤; (2) `card list` 는 위치 인자.
- 이웃·폭 검사(운용 규칙 2호, 종료 코드): diff --check 0 · fmt 0(첫 검사 1 → 적용 뒤 0) · ops rustfmt 0 · bash -n 0 · clippy --all-targets -D warnings 0 · cargo build 0 · --bins 0 · **--lib 전체 0** · 영역+macos+고정 51 대상 0(Running 51줄) · python repo_scans·source_manifest·local_check·skill_references 0 · card/ledger 영역 파이썬 모듈 0개(없음) · trailer-check 0 · spec check --all 0.
- 도움말 골든: 새 하위명령·--all 이 도움말을 바꾸지만 `tests/fixtures/help/**` 는 다른 카드 예약 중 — help_layout·product_cli 가 같은 배치에서 0 으로 통과해 최상위 골든은 불변임을 실행으로 확인.
- blocked(다음 조각): `src/card/dispatch.rs`(비밀 화면 후속의 예약, tier 2 감사 진행 중).
- 감사(2026-11-20, tier 1, audit-card-withdraw) 첫 판정 **FAIL** — B-1 SPEC-코드 범위 자기모순(ops/card.rs 배선이 본문·Non-goals·blocked 와 어긋남 + 배선 액션 시험 없음), B-2 명시적 state 필터 동작 변경 무검증(--state hold 에서도 철회 카드 숨김). 기준 (1)(2) 자체는 증거로 충족 판정.
- 수리(같은 가지): B-2 — 목록 제외 subquery 를 state 필터가 없을 때만 적용(`?1 IS NULL AND NOT EXISTS …`), 명시적 필터는 이전대로 답함 + `--state hold` 로 철회 카드가 보이는 고정 시험; B-1 — SPEC 본문·Non-goals·Scope 를 실제 조각에 맞게 갱신 + 배선된 `card withdraw` 액션을 몰아보는 시험(the_wired_withdraw_action_refuses_take_and_hides_from_the_default_list) 추가 + 시험 파일 머리글 정정; NB-1 show 주석 원위 · NB-2 event actor 단언 추가.
