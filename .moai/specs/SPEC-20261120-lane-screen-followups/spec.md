---
id: SPEC-20261120-lane-screen-followups
title: "lane-screen follow-ups — other choice shapes, scan time cap, list size, reason visibility"
status: draft
revision: 1
created: 2026-11-20
parent: SPEC-20261119-lane-screen-scan
---

# 레인 화면 후속 — 다른 선택 모양·스캔 시간 상한·목록 크기·사유 가시성 (카드 C-134)

## Intent

C-120 감사가 남긴 다섯 후속(카드 C-134 기준 1·2·3): R-N2 — 선택 줄이 `N.`·`N)` 이 아닌 `[1] Yes`·
`1: Yes`·`a. Yes` 꼴이면 지금 분류기가 일하는 레인으로 읽어요. R-N3 — `Overwrite? (y/N/a)` 같은 괄호
질문도 같습니다. C-N3 — `lane-screen --scan` 에 전체 시간 상한이 없어 느린 터미널 하나가 스캔 전체를
붙잡아요. C-N4 — 목록 크기에 실패하는 상한이 없어요(잘림 알림만 있음). C-N8 — 긴 프로그램 경로가
보고 줄에서 시간 초과 사유를 가려요. 이 조각은 SPEC 과 RED 시험(새 시험 파일, 가짜 orca)만 담고,
구현은 `src/lane_screen.rs`·`src/lane_send.rs` 가 풀리면 범위를 넓혀 다음 조각에서 해요.

## Non-goals

- `src/lane_screen.rs`·`src/lane_send.rs`·`src/ops/*` 의 지금 편집 — C-120(agent-3)이 조각 B·C 수리 중.
- 실제 orca 실행, 실제 HOME, 네트워크.
- 이미 있는 규칙(번호 Yes/No 쌍, 한도 줄, 바닥글)의 재정의 — 새 모양만 더해요.

## Requirements

- REQ-1: When a screen that otherwise reads as working carries a choice list whose rows use the `[1] Yes`, `1: Yes` or `a. Yes` shapes, or the question `Overwrite? (y/N/a)`, the classifier shall answer permission_prompt or unknown, never working or idle · failure: none — an unanswered choice on the screen must never be read as a lane that is safe to type into.
- REQ-2: When lane-send is asked to send text into a screen REQ-1 refuses, the command shall refuse without calling orca send · failure: none — no `send` call may reach the Orca program.
- REQ-3: When `lane-screen --scan` runs with a total time cap, the scan shall stop once the cap has passed and report every terminal it did not read as unreachable, naming the cap in its reason · failure: none — unread terminals are reported, never silently dropped.
- REQ-4: When `orca terminal list` reports more terminals than the size cap, the scan shall fail with a reason that names the reported size · failure: none — an over-large list is never partially scanned.
- REQ-5: When a terminal's read times out and its program path is long, the report line shall keep the timeout reason visible, cutting the path instead of the reason · failure: none — the reason phrase must stay in the line.
- REQ-6: When a known question is on the screen and a short row below it is choice-shaped, the classifier shall answer permission_prompt or unknown, never working or idle · failure: none — choice-shaped is a row of two or more short words split by a delimiter (`Approve / Reject`, `OK / Cancel`, `[ Proceed ] [ Cancel ]`) or a short row starting with a bracket or a number (`[1] Approve / [2] Reject`, `1. Proceed`); the word list (yes no allow deny y n) is gone and the shape is the criterion.
- REQ-7: When the shape criterion of REQ-6 replaces the word list, the twelve refusals of the C-120 re-audit shall keep refusing while a question-adjacent row that is not choice-shaped stops refusing · failure: none — the twelve are the six question-line choice shapes over working and idle plus the numbered or delimiter-split choice rows under a question (`Allow / Deny`, `1. run the tests / 2. open the pull request`); the freed rows are plain prose such as `no changes yet`; the cost and the gain are both written into docs/design-review/appendix/operations/lane_send.md.

## Acceptance

- AC-1 (REQ-1, REQ-2): 표 시험 — 네 모양(`[1] Yes`·`1: Yes`·`a. Yes`·`Overwrite? (y/N/a)`) 각각 일하는 표시(esc to interrupt) 아래에서 상태는 permission_prompt 또는 unknown, lane-send 는 같은 화면에 거절하고 orca send 호출 0회.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli`
- AC-2 (REQ-3): 느린 가짜 orca — 상한 시간이 지나면 스캔이 멈추고 안 읽은 터미널이 unreachable 로, 사유에 상한이 적힘.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli scan_followups`
- AC-3 (REQ-4, REQ-5): 목록이 상한보다 크면 크기를 적은 사유로 실패 · 긴 경로 터미널의 시간 초과 줄에서 사유가 잘리지 않음.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli scan_followups`
- AC-4 (REQ-6, REQ-7): 한 표 시험 — N2 일곱 모양(아는 질문 아래)은 permission_prompt 또는 unknown, 재감사의 12가지 거절은 그대로 거절, 질문 아래 선택 모양이 아닌 줄(상태줄 산문)은 working·idle 로 풀림.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli n2_shapes_under_a_known_question_are_never_working n1_table_keeps_the_twelve_refusals_and_frees_plain_prose`

## Scope

- write: `tests/lane_screen_followups_cli.rs`, `.moai/specs/SPEC-20261120-lane-screen-followups/spec.md`
- read-only: `src/lane_screen.rs`, `src/lane_send.rs`, `src/ops/lane_screen.rs`, `src/ops/lane_send.rs`, `tests/lane_screen_cli.rs`, `tests/common/**`
- forbidden: 이 조각에서 `src/**` 편집(구현은 다음 조각), 실제 orca, 실제 HOME, 네트워크, 실제 원격 저장소

## Plan

1. RED 시험 파일(이 조각): 분류 반쪽은 지금 판정 가능(현재 코드가 working 으로 답해 실패), lane-send·--scan 반쪽은 C-120 조각 B·C 가 통합 가지에 들어온 뒤에 같은 실패가 기능 결함으로 자리잡음 — 그동안은 표면 없음으로 실패.
2. 리더 범위 추가(노트 note:134:27a3851ffd): 재감사 N1·N2 의 한 표 시험(AC-4)을 같은 파일에 더함 — N2 일곱 모양은 지금 working 이라 진짜 RED, 12가지 유지 절반은 지금도 통과(재작성 지킴이), 산문 풀림 절반은 지금 거절이라 반대 방향 RED.
3. 구현 조각(다음): `src/lane_screen.rs`(선택 모양 인식·낱말 목록 제거), `src/lane_send.rs`·`src/ops/lane_screen.rs`(상한·목록 크기·사유), `docs/design-review/appendix/operations/lane_send.md`(대가 설명) — card scope 넓혀 착수.
4. 이웃 검사 16개 접두(help_layout·lane* 15)와 가벼운 검사.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 모양 인식은 `numbered()` 옆에 별도 인식기 | 기존 `N.`·`N)` 규칙을 건드리면 C-120 수리와 겹침 | numbered 확장으로 |
| D-2 | 스캔 시간 상한은 `--max-seconds <N>`(전체 벽시계) | C-N3 의 "전체 시간" | 터미널별 상한으로 |
| D-3 | 목록 크기 상한은 `--max-terminals <N>`, 넘으면 실패 | C-N4 의 "실패하며 크기를 적은 사유" | 잘림 알림 유지로 |
| D-4 | 긴 경로는 자르고 사유는 자르지 않음 | C-N8 — 사유가 정보 | 경로 전체 출력으로 |
| D-5 | 선택 인식은 낱말 목록이 아니라 모양: 아는 질문 아래(8줄 창) 구분자로 나뉜 짧은 낱말 둘 이상, 또는 괄호·번호로 시작하는 짧은 줄 | 리더 범위 추가 노트(재감사 N1·N2) — 목록을 늘리면 계속 뒤진다 | 낱말 목록 복귀 |
| D-6 | 재감사의 12가지 거절은 모양 기준 아래에서도 그대로 거절(안전 쪽 유지), 질문 아래 선택 모양이 아닌 산문 줄만 풀림 — 대가와 이득을 lane_send.md 에 적음 | N1 의 "어느 쪽이 보내지고 막히는지 정해" 에 대한 제안(구현 조각에서 리더 확정) | 12가지 중 일부 허용으로 |

## Risks

- 새 모양 인식이 지나치게 넓으면 일하는 레인을 unknown 으로 만들어요 — 표 시험이 일하는 화면 회귀를 지켜요(이 조각 RED 에 포함).
- 시간 상한이 너무 작으면 정상 스캔이 중단돼요 — 문서와 거부 문구에 상한값을 적어요.

## Result

이 조각(RED 까지, base 33cda595c1)의 실행: `cargo test --offline --locked --test
lane_screen_followups_cli` — 1 통과(`the_plain_working_screen_still_reads_working`, 회귀 보호)·5 실패(2026-11-20, heavy-slot agent-21-c134).

- 진짜 RED: `other_choice_shapes_and_bracket_questions_are_never_working` —
  "bracket-number: expected permission_prompt or unknown, got working"(`[1] Yes`·`1: Yes`·
  `a. Yes`·`Overwrite? (y/N/a)` 네 모양 모두 현재 working).
- 표면 없음 RED(C-120 조각 B·C 가 통합 가지에 아직 없음): lane-send 거절 시험 1(명령 없음)과
  --scan 시험 3(unknown argument) — 그 가지들이 들어오면 같은 시험이 기능 결함 RED 가 돼요.
- 검사: cargo fmt --check 통과(fmt 후), `moai spec check` 0 errors 0 warnings,
  home_isolation_scan 18 통과.
- 리더 범위 추가(노트 note:134:27a3851ffd, 재감사 N1·N2): 표 시험 둘을 더해 같은 파일 8개 시험 —
  1 통과(회귀 보호)·7 실패. N2 진짜 RED: `["Approve / Reject"]: expected permission_prompt or
  unknown, got working`(일곱 모양). N1 RED: `the twelve refusals must stay refused,
  ["proceed? (yes/no)"] got working` — 열둘 유지 반쪽은 조각 B 의 질문-낱말 층이 통합 가지에
  아직 없어서 실패(표면 없음), 산문 풀림 반쪽은 지금 통과(재작성 지킴이).
- 구현은 다음 조각: `src/lane_screen.rs`·`src/lane_send.rs` 가 풀리면 card scope 넓혀 착수.
