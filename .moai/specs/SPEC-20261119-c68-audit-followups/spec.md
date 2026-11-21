---
id: SPEC-20261119-c68-audit-followups
title: "C-68 remainder: CARDS_BUSY_TIMEOUT pin and the note screen (bidi, false positives, ref/evidence, CRLF)"
status: draft
revision: 2
created: 2026-11-19
---

# C-68 remainder: the timeout pin and the note screen

> 안내(2026-11-21): src/decide/ 는 SPEC-20261120-retire-decide 에서 삭제됐어요. 아래 읽기 전용 참조는 기록이에요.

## Intent

카드 C-68 (p2) 의 네 기준 중 tip(`9bd09c2eff`) 에 이미 있는 것(AC-1 release 경쟁 검사, AC-4 help 예문 그대로 실행)은 남기고, 없는 것만 구현해요: 상태줄 원장 대기의 상한 단언(AC-2)과 노트 화면의 네 항목(AC-3) — bidi 서식 문자 거부, 감사가 나열한 가양성 통과(`sk-learn-pipeline`, 산문 `--token=VALUE`), `--ref`·`--evidence`·`--reason` 에 같은 화면 적용, CRLF 정규화.

## Non-goals

- tip 에 이미 된 AC-1(src/ledger/reservation.rs:786-811·tests/ledger_reservation.rs:911,1049)·AC-4(tests/card_cli.rs:581-627, 커밋 1d4313ff43) 재구현.
- `src/redact.rs` 공유 경로 변경 — 소비자(decide 게이트 `src/decide/outbound.rs:284`, forge 마스킹 `src/forge/public_text.rs:92`)의 고정 시험까지 범위가 늘어나요. 노트 경로에 국소 판정을 둬요(decide 의 `without_sk_names` 선례 방식).
- 실제 HOME·실제 원장 대상 실행.

## Requirements

- REQ-1: When the crate compiles, the status line's ledger wait `CARDS_BUSY_TIMEOUT` shall be pinned at 100 milliseconds or less by a compile-time assertion (a two-second value must fail the build). · failure: none — 단언은 컴파일 시간에 못 박고, 늘리는 순간 빌드가 어긋나요.
- REQ-2: When a note text or a one-line field value (ref, evidence, reason) holds a bidi or directional character (U+2028–U+202E or U+2066–U+2069), the note screen shall refuse it naming the directional characters. (선례: `src/decide/normalise.rs` `is_bidi`)
- REQ-3 (REQ-2): IF such a character reached a stored note or field, THEN the pinning test shall have failed.
- REQ-4: When a note holds a hyphenated lower-case name that starts with the `sk` prefix or a plain command-line option assignment, the note screen shall pass the text, blanking only the name part before the shared secret screen (a real key run keeps its own screening). · failure: none — 가양성 통과를 놓치면 AC-2 의 통과 고정이, 과한 블랭킹은 REQ-5 의 거부 고정이 어긋나요.
- REQ-5 (REQ-4): IF the blanking let a secret-shaped value through, THEN the note screen shall refuse the note.
- REQ-6: When a one-line field value (a ref, an evidence or a reason) is given, the value shall pass the same screen as the note text (secret shapes, directional characters, controls). · failure: none — 화면을 빠뜨리면 AC-2 의 고정이 어긋나요.
- REQ-7: When a note file uses Windows line endings, the reader shall normalise them to line feeds and refuse a lone carriage return naming the carriage return. · failure: none — 정규화·거부는 AC-2 의 CLI 고정이 못박아요.

## Acceptance

- AC-1 (REQ-1): 컴파일 시간 단언 추가; 2 초 값으로 바꾸면 빌드가 실패함을 기록.
  verify: `cargo test --locked --offline --lib statusline::state`·변이 기록(2000 ms 로 바꾸면 컴파일 실패)
- AC-2 (REQ-2…REQ-7): 노트 화면 CLI 고정 — 방향 문자(U+202E·U+2066) 거부와 메시지 이름, 감사가 나열한 두 가양성(하이픈 낱말 이름, 산문 옵션 대입) 통과, 진짜 키 모양·비밀형 값은 여전히 거부, 한 줄 필드(ref·evidence·reason) 화면 적용, CRLF 정규화 저장과 단독 CR 거부(CR 이름).
  verify: `cargo test --locked --offline --test card_work_cli`
- AC-3 (REQ-1…REQ-7): 대상 검사.
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`; `cargo fmt --check`(전량은 묶음 검증·CI 인용)
- AC-4 (REQ-2…REQ-7): tip 증거 — 카드 기준의 release 경쟁 검사와 help 예문 그대로 실행은 tip 에 이미 충족.
  verify: `git show origin/v2026.11.0:tests/ledger_reservation.rs | sed -n '911p;1049p'`; `git show origin/v2026.11.0:tests/card_cli.rs | sed -n '581p'`

## Scope

- write: `src/statusline/state.rs`, `src/card/work.rs`, `src/ledger/workflow.rs`, `tests/card_work_cli.rs`, `.moai/specs/SPEC-20261119-c68-audit-followups/**`
- read-only: `src/redact.rs`, `src/decide/outbound.rs`(`without_sk_names` 선례), `src/decide/normalise.rs`(`is_bidi` 선례), `tests/ledger_reservation.rs`, `tests/card_cli.rs`
- forbidden: `src/redact.rs` 쓰기, 실제 HOME·실제 원장, 네트워크

## Plan

1. RED: `tests/card_work_cli.rs` 에 네 항목 CLI 고정 먼저(tip 에서 실패 관찰). AC-B 단언은 컴파일 시간 몫(바로 초록; 2 초 변이로 검증력 증명).
2. GREEN: `workflow.rs`(`add_note`·`one_line` 의 방향 문자 거부 + CR 메시지), `work.rs`(노트 국소 pre-pass `sk-` 이름·`--옵션=` 이름 블랭킹 + ref/evidence/reason 화면 + `read_note_text` CRLF 정규화), `state.rs`(상수 단언).
3. 검사(대상 한정)·변이 2건(상한 2000 ms → 빌드 실패; pre-pass 제거 → 가양성 시험 실패)·push(PR 없음)·자체 감사.

## Risks

- `--name=value` 이름 블랭킹이 비밀형 값을 숨길 수 있어요 → 값은 남겨두고 이름만 블랭킹(REQ-5 시험이 구멍을 못박음).
- `workflow.rs` 는 여러 시험이 쓰는 공유 파일 — `add_note`·`one_line` 메시지 문구가 바뀌면 기존 고정이 깨질 수 있음(깨지면 그 시험 문구를 새 메시지로 갱신하고 본문에 밝힘).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | tip 에 이미 된 카드 AC-1·AC-4 는 증거로 met 표시, 재구현 없음 | 리더 배차 1단계 | 전체 재구현 카드 |
| D-2 | `redact.rs` 는 그대로; 노트 경로 국소 pre-pass | 공유 경계 변경은 decide·forge 고정까지 범위 확대(조사 권고) | redact 경계 수정 카드 |
| D-3 | CRLF 는 정규화(카드 문구), 단독 CR 은 CR 이름을 짚어 거부 | 감사 N-c "이름 없이 통째 거부" 지적 | 통째 거부 유지 |
| D-4 | 방향 문자 범위는 decide `is_bidi`(202A–202E·2066–2069)에 2028–2029 를 더한 2028–202E·2066–2069 | 선례 일관 + 감사가 202E 를 이름 | 202E 만 |

## Result

- AC-1: met — `state.rs` 컴파일 시간 단언(`CARDS_BUSY_TIMEOUT.as_millis() <= 100`) 추가; 변이 M-timeout(2000 ms)이 E0080 로 빌드를 죽여요(관찰: `cargo check` error[E0080]).
- AC-2: met — RED(가양성 줄이 "2 word(s) with the shape of a secret" 로 거부 관찰) → GREEN `card_work_cli` 25 통과: 방향 문자(U+202E·U+2066·U+2028) 거부+이름 명시, 가양성 두 종 통과(sk 하이픈 이름·산문 옵션 대입), 진짜 키 단독·옵션값 뒤 키 거부(값 단독 스크린), `--ref`·`--evidence` 화면 적용, CRLF 정규화(33바이트 저장 관찰)+단독 CR 거부("carriage" 이름). 변이 M-prepass 제거 → 가양성 고정 실패로 사망. 비고: redact 의 단어 시작 경계(`note:` 뒤 키 등)는 기존 공유 경계라 그대로(D-2).
- AC-3: met — `cargo test --lib` 전체 1450 통과(99.16 초), `--test home_isolation_scan` 18 통과, `cargo fmt --check` exit 0, `clippy --all-targets -D warnings` 0. 전량 통합 묶음·CI 인용.
- AC-4: met — tip `9bd09c2eff` 확인: `tests/ledger_reservation.rs:911,1049`(release 경쟁 검사)와 `tests/card_cli.rs:581`("Card C-68 N-d" help 예문 그대로 실행)이 이미 존재.
  사고 기록: 돌연변이 측정 후 커밋 전 `git restore` 로 work.rs 미커밋 변경을 잃어 재적용(재적용 뒤 25 통과 재확인) — 두 돌연변이 증거는 분실 전 동일 코드에서 측정됐어요.
