---
id: SPEC-20261119-card-scope-widen-tests
title: "card scope widen: pin the guarantees the first slice's tests did not (C-94 follow-up, audit N4·N5)"
status: draft
revision: 1
created: 2026-11-19
---

# card scope widen — the missing pins (C-94 follow-up)

## Intent

C-94 감사(노트 2859, PASS)의 비차단 N4·N5가 가리킨 "시험 없는 보장"을 리더 DISPATCH(노트 2860)대로 시험으로 고정해요. src 는 건드리지 않고(그 파일들은 C-108·C-43 이 쓰는 중) `tests/card_scope_widen_cli.rs` 와 이 SPEC 폴더만 바꿔요. 각 시험이 어떤 변이에 실패하는지 Result 에 기록해요.

## Non-goals

- src 변경 — 없음(이 조각은 시험만).
- N1(--generation 기대값) — 리더가 후속으로 따로 결정한 것, 이 조각 밖.
- 실제 원장·실제 HOME 대상 실행.

## Requirements

- REQ-1: When the reason of a held-card widening reaches 500 characters, the command shall accept it in ASCII and Korean alike and refuse 501 characters naming the limit (the limit counts characters, not bytes). · failure: none — AC-1 의 시험이 못박아요.
- REQ-2: When someone who is not the holder widens a held card, the command shall grow the scope and record that actor as the event's (SPEC-20261119-card-scope-widen D-2 keeps it that way). · failure: none — AC-2 의 시험이 못박아요.
- REQ-3: When the held card's lease has ended, the widening shall still grow the scope at the same generation. · failure: none — AC-3 의 시험이 못박아요.
- REQ-4: When the widened reservation is handed off, reclaimed or released, each step shall process the widened row with its scope intact. · failure: none — AC-4 의 시험이 못박아요.
- REQ-5: When an added pattern differs from another card's pattern only by letter case, the widening shall be refused naming that card. · failure: none — AC-5 의 시험이 못박아요.
- REQ-6: When the card.scope event cannot be written, the widening shall refuse and leave every row of the ledger as it was. · failure: none — AC-6 의 시험이 못박아요.

## Acceptance

- AC-1 (REQ-1): reason 500자(ASCII·한글 각각) 수락, 501자 거절(한글 500자는 1500바이트 — 한도가 문자 수임을 함께 고정).
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_reason_of_500_characters_is_taken_and_501_refused`
- AC-2 (REQ-2): 비보유자 --by 로 넓히기 성공, 사건 actor 가 그 이름.
  verify: `cargo test --locked --offline --test card_scope_widen_cli someone_who_is_not_the_holder_can_widen`
- AC-3 (REQ-3): 끝난 lease 에서 넓히기 — 같은 세대로 성공.
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_widen_on_an_ended_lease_still_grows`
- AC-4 (REQ-4): 넓힌 행이 handoff→reclaim→release 를 모두 통과(세대별 상태·스코프 보존).
  verify: `cargo test --locked --offline --test card_scope_widen_cli the_widened_row_survives_handoff_reclaim_and_release`
- AC-5 (REQ-5): 대소문자만 다른 겹침 거절(양방향).
  verify: `cargo test --locked --offline --test card_scope_widen_cli an_overlap_refusal_ignores_letter_case`
- AC-6 (REQ-6): 사건 쓰기 실패(시험 측 트리거 주입) 시 거절·행 보존, 트리거 제거 뒤 같은 요청 성공.
  verify: `cargo test --locked --offline --test card_scope_widen_cli a_widen_whose_event_cannot_be_written_leaves_the_row_as_it_was`

## Scope

- write: `tests/card_scope_widen_cli.rs`, `.moai/specs/SPEC-20261119-card-scope-widen-tests/**`
- read-only: `src/ledger/reservation.rs`, `src/card/dispatch.rs`, `src/ops/card.rs`, `src/scope.rs`
- forbidden: src 쓰기, 실제 원장·HOME, 네트워크

## Plan

1. 시험 6종 추가(하네스에 쓰기 연결 헬퍼 `exec` 추가 — 시험 측 주입용).
2. 전체 스위트 통과 관찰(기존 9 + 신규 6 = 15).
3. 커밋 뒤 변이 6종(각 시험 하나씩 실패하는지) — 매번 복원·트리 무변경 확인.
4. fmt·clippy·이웃(card_work_cli 등) → push(PR 없음) → 자체 감사(tier 1, 리더 지시) → READY.

## Risks

- 시험이 실제 동작을 고정하는 핀 시험이라 "RED 먼저"가 없어요 — 대신 변이 6종이 각 시험의 검출력을 증명해요.
- 대소문자 겹침은 scope::overlap 의 어휘 규칙에 의존 — 실제 동작을 관찰해 고정해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 사건 실패 주입은 시험 측 SQLite 트리거(RAISE ABORT)로 | src 건드리지 않고 이벤트 쓰기만 실패시키는 유일한 CLI 경로 | 실패 주입 하네스 카드 |
| D-2 | lease 만료는 직접 UPDATE(lease_until = 1)로 만듦 | reclaim 전제와 같은 만료 상태, 시험 측 조작 | 시계 모의 카드 |

## Result

- AC-1: met — `a_reason_of_500_characters_is_taken_and_501_refused_in_ascii_and_korean`: 500 ASCII·한글(1500바이트) 수락, 501 거절("500" 포함), 거절분 무기록. 변이 M1(`>` → `>=`, dispatch.rs:437) KILLED — 500자가 거절됨을 관찰(FAILED 1.01초).
- AC-2: met — `someone_who_is_not_the_holder_can_widen_as_the_spec_decides`: --by another-leader 로 성공, holder 는 S_A 유지, 사건 actor 기록. 변이 M2(보유자-only 검사 추가) KILLED — 비보유자가 거절됨을 관찰(FAILED 1.25초).
- AC-3: met — `a_widen_on_an_ended_lease_still_grows_the_scope`: lease_until=1 로 나이킨 뒤 같은 세대 성공. 변이 M3(만료 lease 거절 추가) KILLED — 관찰(FAILED 1.03초).
- AC-4: met — `the_widened_row_survives_handoff_reclaim_and_release`: 넓힌 스코프가 handoff(gen2, S_B)→reclaim(gen3, S_C)→release 로 세대별 상태 released·reclaimed·released 와 스코프 보존. 변이 M4(handoff 의 스코프 복사 `AND 0` 제거) KILLED — 관찰(FAILED 1.05초).
- AC-5: met — `an_overlap_refusal_ignores_letter_case`: widen 방향(docs/ 보유 카드가 apps/web/ 소문자 추가 → C-1 의 APPS/WEB/ 와 충돌)과 reserve 방향(소문자 표기 카드의 take 거절) 모두 거절·무기록. 두 표기는 공존 불가라(reserve 가 두 번째를 막음) 한 쌍의 양문이 전부예요 — 시험 주석에 명시. 변이 M5(overlap 소문자화 제거) KILLED — 관찰(FAILED 1.18초).
- AC-6: met — `a_widen_whose_event_cannot_be_written_leaves_the_row_as_it_was`: 트리거 주입(RAISE ABORT)으로 사건 쓰기 실패 → 거절(kind invalid_input, 메시지에 주입 문구)·덤프 무변경·사유 미반복, 트리거 제거 뒤 같은 요청 성공.
  변이 M6: 진짜 변이(`let _ = add_event(…);` — 전파 `?` 제거)에서 widen 이 **ok:true 로 성공**(스코프 커밋·사건 누락, 관찰: refusal()의 left Some(0) right Some(1), 2026-11-19T08:2xZ)하고 시험이 FAILED → **KILLED**(감사자도 동일 재현). 정정: 이 SPEC 의 첫 Result 행과 커밋 48da70ec39 메시지가 적은 "SURVIVED — 등가 변이"는 잘못된 관찰이었어요 — 당시 적용한 변이는 `let _ = add_event(…)?;` 로 `?` 가 남은 **무효 변이**(오류가 그대로 전파, 동작 불변 — 그 관찰이 "커밋 단계가 같은 실패"라는 해석의 원인). 감사 B1 이 반증했고 이 행이 정정이에요.
- 이웃(리더 지시 검사): `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0(51.1초), `python3 -m unittest discover -s tests -p 'repo_scans_test.py'`(MOAI_CLI_BIN=<c94w 빌드>) 25 OK(10.6초) — 관찰 2026-11-19T08:0xZ. 전량은 묶음 검증·CI 인용.
  사고 기록: 없음 — 변이는 커밋 뒤 적용, 매번 restore 후 `git status` 로 트리 무변경 확인(C-68 의 교훈 적용).
