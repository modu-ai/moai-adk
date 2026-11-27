---
id: SPEC-20261120-lane-screen-followups
title: "lane-screen follow-ups — other choice shapes, scan time cap, list size, reason visibility"
status: approved
revision: 3
created: 2026-11-20
parent: SPEC-20261119-lane-screen-scan
approval:
  by: user
  at: 2026-11-27
  basis: "리더 결정 노트 10779 (USER-APPROVAL C-134 SPEC-20261120-lane-screen-followups 한 줄 개정, 2026-11-27): 사용자가 직접 골랐어요(리더 세션의 질문 도구), \"yes/no 짝만 예외로 거절 (Recommended)\"; 승인 범위는 REQ-6(D-6)에 \"질문 줄 안의 긍정·부정 낱말 짝(yes/no, y/n, 대소문자 무관)이 `or` 로 이어진 것은 거절\"이라는 예외 하나를 더하는 것과 그에 따르는 revision·승인 블록 갱신뿐이에요(개정 3 본문 커밋 11516b8093), 다른 낱말 목록은 되살리지 않아요; 개정 2 승인은 노트 10710, 개정 1 승인은 노트 10680 이고 그대로예요; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요(레인은 자기 승인을 기록하지 않았어요)"
  digest: "sha256:411580f011a0ee42fcd5899de2e27661c97b9e31fed1af718d0c697451453a63"
---

# 레인 화면 후속 — 다른 선택 모양·스캔 시간 상한·목록 크기·사유 가시성 (카드 C-134)

## Intent

C-120 감사가 남긴 다섯 후속(카드 C-134 기준 1·2·3): R-N2 — 선택 줄이 `N.`·`N)` 이 아닌 `[1] Yes`·
`1: Yes`·`a. Yes` 꼴이면 지금 분류기가 일하는 레인으로 읽어요. R-N3 — `Overwrite? (y/N/a)` 같은 괄호
질문도 같습니다. C-N3 — `lane-screen --scan` 에 전체 시간 상한이 없어 느린 터미널 하나가 스캔 전체를
붙잡아요. C-N4 — 목록 크기에 실패하는 상한이 없어요(잘림 알림만 있음). C-N8 — 긴 프로그램 경로가
보고 줄에서 시간 초과 사유를 가려요. 이 조각은 SPEC 개정과 사용자 승인까지만 담아요. 승인 뒤에
RED 시험(새 시험 파일, 가짜 orca)을 붙이고, 구현은 같은 카드의 다음 조각에서 해요.

등급(2026-11-27): 이 변경은 레인 입력을 막는 안전 분류기(`classify`)와 `lane-send` 의 거절을 바꿔요.
`src/lane_screen.rs`·`src/lane_send.rs` 는 감사 floor 목록(`audit.md`)에 없어서 경로로는 tier 1 이지만,
분류기가 틀리면 확인 질문에 글자가 들어가므로 보수적으로 tier 2 로 다뤄요(SPEC 승인과 전체 리뷰).

## Non-goals

- 이 조각에서 `src/**` 편집 — SPEC 승인 뒤의 구현 조각 몫. C-120 은 끝났고(done) 두 파일은 풀려 있어요.
- 실제 orca 실행, 실제 HOME, 네트워크.
- 이미 있는 규칙(번호 Yes/No 쌍, 한도 줄, 바닥글)의 재정의 — 새 모양만 더해요.

## Interview

사용자가 고른 것(2026-11-27, 리더 세션의 질문 도구, 리더 노트 10680에 옮겨 적힘): "승인, D-7 포함 (Recommended)".

- Q1: Is revision 1 of this SPEC approved? -> A: yes, approved as written (user pick "승인, D-7 포함 (Recommended)").
- Q2: Is a terminal list cut at the output limit part of this work (D-7)? -> A: yes, option (a): a cut list output also fails with a reason that names the size. It is a new refusal applied at once, not "observe first", because observing alone does not keep text out of a confirmation question.
- Q3: What does it cost? -> A: on a screen that looks like a choice, the leader's lane sends may be refused more often.
- Q4: What changes for the person? -> A: (1) a screen showing `[1] Yes`, `1: Yes`, `a. Yes` or `(y/N/a)` is refused and nothing is sent, (2) a choice is recognised by its shape, not by a word list, (3) `--scan` gets a total time cap and a terminal-count cap, (4) a long program path no longer hides the timeout reason.

## Requirements

- REQ-1: When a screen that otherwise reads as working carries a choice list whose rows use the `[1] Yes`, `1: Yes` or `a. Yes` shapes, or the question `Overwrite? (y/N/a)`, the classifier shall answer permission_prompt or unknown, never working or idle · failure: none — an unanswered choice on the screen must never be read as a lane that is safe to type into.
- REQ-2: When lane-send is asked to send text into a screen REQ-1 refuses, the command shall refuse without calling orca send · failure: none — no `send` call may reach the Orca program.
- REQ-3: When `lane-screen --scan` runs with a total time cap, the scan shall stop once the cap has passed and report every terminal it did not read as unreachable, naming the cap in its reason · failure: none — unread terminals are reported, never silently dropped.
- REQ-4: When `orca terminal list` reports more terminals than the size cap, the scan shall fail with a reason that names the reported size · failure: none — an over-large list is never partially scanned.
- REQ-5: When a terminal's read times out and its program path is long, the report line shall keep the timeout reason visible, cutting the path instead of the reason · failure: none — the reason phrase must stay in the line.
- REQ-6: When a known question is on the screen and a short row below it is choice-shaped, the classifier shall answer permission_prompt or unknown, never working or idle · failure: none — choice-shaped is a row of two or more short words split by a delimiter (`Approve / Reject`, `OK / Cancel`, `[ Proceed ] [ Cancel ]`) or a short row starting with a bracket or a number (`[1] Approve / [2] Reject`, `1. Proceed`); the word list (yes no allow deny y n) is gone and the shape is the criterion, with one exception (user decision 2026-11-27, ledger note 10779): a question row that holds a yes/no word pair (`yes`/`no` or `y`/`n`, any case) joined by `or` is refused too, and no other word list returns.
- REQ-7: When the shape criterion of REQ-6 replaces the word list, the twelve refusals of the C-120 re-audit shall keep refusing while a question-adjacent row that is not choice-shaped stops refusing · failure: none — the twelve are the six question-line choice shapes over working and idle plus the numbered or delimiter-split choice rows under a question (`Allow / Deny`, `1. run the tests / 2. open the pull request`); the freed rows are plain prose such as `no changes yet`; the cost and the gain are both written into docs/design-review/appendix/operations/lane_send.md.

## Acceptance

- AC-1 (REQ-1, REQ-2): 표 시험 — 네 모양(`[1] Yes`·`1: Yes`·`a. Yes`·`Overwrite? (y/N/a)`) 각각 일하는 표시(esc to interrupt) 아래에서 상태는 permission_prompt 또는 unknown, lane-send 는 같은 화면에 거절하고 orca send 호출 0회.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli`
- AC-2 (REQ-3): 느린 가짜 orca — 상한 시간이 지나면 스캔이 멈추고 안 읽은 터미널이 unreachable 로, 사유에 상한이 적힘.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli scan_followups`
- AC-3 (REQ-4, REQ-5): 목록이 상한보다 크면 크기를 적은 사유로 실패 · 긴 경로 터미널의 시간 초과 줄에서 사유가 잘리지 않음.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli scan_followups`
- AC-4 (REQ-6, REQ-7): 한 표 시험 — N2 일곱 모양(아는 질문 아래)은 permission_prompt 또는 unknown, 재감사의 12가지 거절은 그대로 거절, 질문 아래 선택 모양이 아닌 줄(상태줄 산문)은 working·idle 로 풀림.
  verify: `cargo test --locked --offline --test lane_screen_followups_cli -- n2_shapes_under_a_known_question_are_never_working n1_table_keeps_the_twelve_refusals_and_frees_plain_prose`

## Scope

- write: `.moai/specs/SPEC-20261120-lane-screen-followups/spec.md`; 승인 뒤에 `tests/lane_screen_followups_cli.rs`(새 시험 파일), 그리고 구현 조각에서 `src/lane_screen.rs`, `src/lane_send.rs`, `src/ops/lane_screen.rs`, `docs/design-review/appendix/operations/lane_send.md`, `docs/design-review/appendix/operations/lane_screen.md`
- read-only: `src/ops/lane_send.rs`, `tests/lane_screen_cli.rs`, `tests/common/**`
- forbidden: SPEC 승인 전의 `src/**`·`tests/**` 편집, 실제 orca, 실제 HOME, 네트워크, 실제 원격 저장소

## Plan

1. SPEC 개정과 승인(이 조각, 2026-11-27 NeXT 팁 `e98798e41f` 기준). 옛 가지의 RED 시험 파일(옛 커밋 `5c709d26e1`·`83e0a6a3de`, 가지 `spec/lane-screen-followups-20261120`)은 승인 뒤에 새 가지로 가져와 붙여요. 팁에서 돌려 본 실제 실패(헤드 `e98798e41f`, 시험 8개 중 1 통과·7 실패): 분류 반쪽은 `[1] Yes`·`Approve / Reject` 가 working 으로 읽혀 실패하고, `lane-send` 는 `[1] Yes` 가 있는 일하는 화면에 글자를 보내 버려요(`"sent":true`) — 진짜 결함. `--scan` 반쪽은 `--max-seconds`·`--max-terminals` 가 아직 없어서 `unexpected argument` 로 실패해요. C-120 조각 B·C 는 이미 팁에 있어서 표면이 있고, 이 실패들이 기능 결함으로 자리잡았어요.
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
| D-7 | (사용자 결정 필요, 기본값: 포함) 터미널 목록 출력이 `OUTPUT_LIMIT`(1 MiB)를 넘어 잘릴 때도 "is not an Orca answer" 대신 크기를 적은 사유로 실패시킨다 | 카드 C-N4 의 원문. `--max-terminals` 만으로는 이 경우를 덮지 않음 | 터미널 수 상한만 |
| D-6 | 재감사의 12가지 거절은 모양 기준 아래에서도 그대로 거절(안전 쪽 유지), 질문 아래 선택 모양이 아닌 산문 줄만 풀림 — 대가와 이득을 lane_send.md 에 적음. 예외 하나(사용자 결정 2026-11-27, 레저 노트 10779): 질문 줄 안의 긍정·부정 낱말 짝(yes/no, y/n, 대소문자 무관)이 `or` 로 이어진 것은 거절하고, 다른 낱말 목록은 되살리지 않음 | N1 의 "어느 쪽이 보내지고 막히는지 정해" 에 대한 제안(구현 조각에서 리더 확정) | 12가지 중 일부 허용으로 |

## Risks

- 새 모양 인식이 지나치게 넓으면 일하는 레인을 unknown 으로 만들어요 — 표 시험이 일하는 화면 회귀를 지켜요(이 조각 RED 에 포함).
- 시간 상한이 너무 작으면 정상 스캔이 중단돼요 — 문서와 거부 문구에 상한값을 적어요.

## Result

이 조각(RED 까지, base 33cda595c1)의 실행: `cargo test --offline --locked --test
lane_screen_followups_cli` — 1 통과(`the_plain_working_screen_still_reads_working`, 회귀 보호)·5 실패(2026-11-20, heavy-slot agent-21-c134).

- 진짜 RED: `other_choice_shapes_and_bracket_questions_are_never_working` —
  "bracket-number: expected permission_prompt or unknown, got working"(`[1] Yes`·`1: Yes`·
  `a. Yes`·`Overwrite? (y/N/a)` 네 모양 모두 현재 working).
- (옛 기록) 표면 없음 RED: 그때는 C-120 조각 B·C 가 통합 가지에 없어서 lane-send 거절 시험 1(명령 없음)과
  --scan 시험 3(unknown argument)이 표면 부재로 실패했어요. 지금(NeXT 팁 `e98798e41f`)은 아래 "팁 위 재확인"이 이 줄을 바꿔요.
- 검사: cargo fmt --check 통과(fmt 후), `moai spec check` 0 errors 0 warnings,
  home_isolation_scan 18 통과.
- 리더 범위 추가(노트 note:134:27a3851ffd, 재감사 N1·N2): 표 시험 둘을 더해 같은 파일 8개 시험 —
  1 통과(회귀 보호)·7 실패. N2 진짜 RED: `["Approve / Reject"]: expected permission_prompt or
  unknown, got working`(일곱 모양). N1 RED: `the twelve refusals must stay refused,
  ["proceed? (yes/no)"] got working` — 열둘 유지 반쪽은 조각 B 의 질문-낱말 층이 통합 가지에
  아직 없어서 실패(표면 없음), 산문 풀림 반쪽은 지금 통과(재작성 지킴이).
- 구현은 SPEC 승인 뒤의 다음 조각이에요(두 파일은 이미 풀렸어요).

### 팁 위 재확인 (2026-11-27, NeXT 팁 `e98798e41f`)

옛 RED 시험 파일을 일시적으로 팁에서 돌려 읽었어요(`moai heavy-slot-acquire --name zoe --max 2 -- cargo nextest run --offline --locked --no-fail-fast --test lane_screen_followups_cli`, 시험 파일은 커밋하지 않고 가지에서 뺐어요). 8개 중 1 통과·7 실패:

- 분류 반쪽: `other_choice_shapes_and_bracket_questions_are_never_working` → `bracket-number: expected permission_prompt or unknown, got working`; `n2_shapes_under_a_known_question_are_never_working` → `["Approve / Reject"]: expected permission_prompt or unknown, got working`. 진짜 RED.
- `lane_send_refuses_the_new_choice_shapes_and_sends_nothing` → 일하는 화면 아래 `[1] Yes` 가 있는데 `lane-send` 가 `{"sent":true,...}` 로 글자를 보내요(종료 0). 안전 결함으로 확인된 RED. (표면은 이미 있어요.)
- `--scan` 반쪽 셋(`scan_stops_at_the_total_time_cap…`, `a_list_over_the_size_cap_fails_naming_the_size`, `a_long_program_path_does_not_hide_the_timeout_reason`): `--max-seconds`·`--max-terminals` 가 아직 없어서 `unexpected argument` 로 실패해요. 같은 이유로 기능 결함 RED.
- `n1_table_keeps_the_twelve_refusals_and_frees_plain_prose`: 열둘 유지 반쪽은 이제 통과하고(조각 B 의 질문-낱말 층이 팁에 있어요), 산문 풀림 반쪽이 `plain prose under a question is not a choice: no changes yet got unknown` 로 실패해요.

시험 파일을 붙이기 전에 고칠 두 곳(승인 뒤 RED 조각에서):

1. `a_long_program_path_does_not_hide_the_timeout_reason` 은 긴 터미널 `worktreePath` 를 쓰는데, REQ-5 의 사유를 가리는 것은 긴 프로그램 경로(`MOAI_ORCA_BIN`)예요. 그 시험은 긴 `MOAI_ORCA_BIN` 경로의 가짜 orca 로 다시 짜야 해요.
2. 카드 본문 C-N4 는 터미널 목록 출력이 1 MiB 를 넘어 `OUTPUT_LIMIT`(`src/lane_send.rs`, `1 << 20`)에서 잘려 "is not an Orca answer" 로 잘못 실패하는 일이에요. D-3 의 `--max-terminals` 는 터미널 수 상한이라 이 경우를 덮지 않아요 → 사용자 결정 D-7 (아래).

가지 이름: 이번 가지는 `spec/20261120-lane-screen-followups`(NeXT 팁에서 새로 땄어요)이고 `moai tdd red` 가 이 이름을 받아서 바꾸지 않았어요(위에서 걱정한 이름 충돌은 없었어요). 옛 가지 `spec/lane-screen-followups-20261120`(이전 레인, 미착지)은 지우지 않고 그대로예요.

### GREEN (2026-11-27, 헤드 c7e62da529, base NeXT 팁 7ec80a8155 를 머지한 위)

RED 커밋 7d0926fe03(`moai tdd red`: 두 확인 모두 test-failure, 종료 101), 구현 c7e62da529(`moai tdd green`: 두 확인 모두 통과, 종료 0). 새 시험 파일 9개 중 RED 7개가 모두 통과하고 가드 둘(`the_plain_working_screen_still_reads_working`, `a_long_program_path_does_not_hide_the_timeout_reason`)은 처음부터 통과였어요: 긴 프로그램 경로는 이유 문구가 경로보다 앞에 와서 팁에서 이미 안 가려져요(REQ-5 는 새 코드 없이 시험으로만 지켜요).

- AC-1: met. `cargo test --locked --offline --test lane_screen_followups_cli` 종료 0, 9 passed. `[1] Yes`·`1: Yes`·`a. Yes`·`Overwrite? (y/N/a)` 는 working 아래에서 permission_prompt 나 unknown 이고 lane-send 는 종료 1 에 orca send 호출 0회예요.
- AC-2: met. `cargo test --locked --offline --test lane_screen_followups_cli scan_followups` 종료 0, 4 passed(시간 상한: 안 읽은 터미널이 unreachable 이고 이유에 `--max-seconds 1`).
- AC-3: met. 같은 명령(4 passed): 터미널 수 상한은 크기 3 과 `--max-terminals` 를 적고 화면을 읽지 않으며, 출력이 1 MiB 를 넘으면 "1 MiB" 를 적고(D-7) "is not an Orca answer" 가 아니고, 긴 경로 시험은 가드로 통과해요.
- AC-4: met. 적힌 verify 명령은 그대로 돌리면 `error: unexpected argument 'n1_table_…' found` 로 종료 1 이에요(cargo 가 테스트 이름을 하나만 받아요, 2026-11-27 슬롯에서 확인). 같은 뜻의 `cargo test --locked --offline --test lane_screen_followups_cli -- n2_shapes_under_a_known_question_are_never_working n1_table_keeps_the_twelve_refusals_and_frees_plain_prose` 는 2 passed 예요. AC 글은 승인 범위라서 레인이 고치지 않았고 리더에게 승인 개정을 요청했어요. 사용자가 개정 2(verify 줄 한 줄을 `--` 꼴로)를 직접 승인했고(원장 노트 10710) 그 줄만 바꿨어요: 이제 적힌 그대로 돌려도 2 passed 예요(개정 2 헤드의 실행은 아래 노트).

범위·규칙의 어긋남과 선택(읽는 사람이 알아야 할 것):

- SPEC Scope 의 write 목록은 `tests/lane_screen_cli.rs` 를 read-only 로 적었는데 REQ-6·REQ-7 이 그 파일의 핀 셋을 바꿔요: 질문 아래 낱말 하나(`Allow`·`deny`·`y`·`N`·`yes`·`No`)는 거절하던 것이 이제 working 이고(REQ-7 이 풀어 주는 평범한 줄), 문장(`yes, go ahead`·`No - wait`·`answer: NO.`)도 같고, 번호 줄 하나(`1. maybe later`)는 working 이던 것이 unknown(REQ-6: `1. Proceed` 와 같은 모양)이에요. 카드 범위를 `card scope`(세대 3, `--by zoe`)로 그 파일까지 넓혔고, 바꾼 단언마다 주석에 REQ 를 적었어요. 같은 이유로 `src/lane_screen.rs` 단위 시험의 단언도 바뀌었어요(이름도 바뀐 셋과 새 셋).
- 시험 파일 `tests/lane_screen_followups_cli.rs` 는 RED 커밋 뒤 잠겨 있어서 GREEN 에서 손대지 않았어요.
- `--max-seconds` 기본값 120 초는 제가 정했어요(승인 글은 상한이 있다고만 해요): 상한이 없으면 C-N3 이 기본으로 남아서예요. 되돌리려면 `DEFAULT_MAX_SECONDS` 한 줄이에요.
- 질문 아래 평범한 줄의 풀림(REQ-7)은 안전 쪽에서 느슨해지는 변경이에요. 자동 보안 검토가 "낱말 목록을 다시 더하라"고 제안했지만, 그것은 REQ-6 의 "낱말 목록은 없고 모양이 기준" 과 사용자가 승인한 D-6 을 되돌리는 일이라 따르지 않았어요. 대가와 이득은 `lane_send.md` 에 적었어요.
- 질문이 없는 새 모양 목록 줄(둘 이상)은 비어 있지 않은 마지막 6줄 안에서만 세요(`LIST_ROWS`): 기존 `lane_screen_cli` 의 창 시험(질문이 9번째, 선택 줄이 7·8번째)이 그대로 맞고, 살아 있는 선택 창은 맨 아래에 그려지기 때문이에요.

Windows·Linux 는 not_run 이에요(fake orca 는 `/bin/sh` 스크립트).

### 수리 (2026-11-27, 리뷰 노트 10724 의 첫 FAIL)

kenan 의 리뷰가 헤드 d354dfe4cf 에서 하나를 찾았어요: 확인 질문 줄 자신이 끝에 선택 모양을 달고 있으면(`Do you want to proceed? Yes/No`, `… ? [Yes] [No]`) 팁은 낱말 목록으로 거절하는데 헤드는 모양 검사를 줄 전체에 걸어서(40자 이하만 모양) 못 읽고 `working`·`idle` 로 읽어 `lane-send` 가 보냈어요. 옛 재감사 R-B1 이 닫았던 구멍이 다시 열린 안전 분류기 약화예요.

- 승인된 REQ-6/REQ-7 의 글 안에서 고쳤어요(SPEC 본문은 바뀌지 않았고 승인 digest 도 그대로예요). REQ-6 은 "질문 줄이나 아래 줄에 선택 모양이 있으면 거절"이고 이 줄은 선택 모양이 있는 줄이에요; REQ-7 이 푸는 것은 선택 모양이 아닌 평범한 줄이에요.
- 수리: 질문 줄은 첫 `?` 뒤의 꼬리에 모양 검사를 해요(`?` 가 없으면 끝의 세 낱말에서부터 각각). 질문 자신의 낱말은 모양으로 읽지 않아요. 이 때문에 같은 뿌리의 반대쪽 틈도 닫혔어요: `Would you like me to commit or push?` 를 줄 전체 모양(`commit or push?`)으로 읽어 head 가 거절했는데 팁은 보냈어요(S-N2: 질문으로 대화를 끝내고 기다리는 레인). 이 줄은 이제 보내요.
- RED 커밋 9c3beb50c7(`moai tdd reopen` 으로 잠금을 풀고 사유를 기록한 뒤 새 사이클): `a_question_row_ending_in_a_choice_shape_is_never_working_or_idle`(일하는 화면·대기 화면 일곱 줄), `a_question_row_without_a_choice_shape_still_reads_working_or_idle`(가드, REQ-7 평범한 줄 포함), `lane_send_refuses_a_question_row_ending_in_a_choice_shape_and_sends_nothing`(fake orca, 보내기 호출 0회); 확인 `cargo test --offline --locked --test lane_screen_followups_cli question_row` 종료 101. 수리 b082df8988 뒤 `moai tdd green` 통과. 단위 시험 `a_question_row_is_read_by_its_tail_after_the_question_mark`(src/lane_screen.rs)도 더했어요.
- 대가: 질문 줄의 `?` 뒤에 선택처럼 생긴 꼬리가 오면(`Are you sure? continue or abort`) 일하는 레인도 거절돼요(안전한 쪽). `?` 가 없는 질문 줄은 끝의 세 낱말 중 하나가 모양이면 거절해요.

### 두 번째 수리 (2026-11-27, 델타 리뷰 노트 10736 의 논블로킹 셋과 리더 결정 노트 10737)

델타 리뷰(PASS)가 논블로킹으로 적은 세 화면이 있어요: 팁은 `unknown` 으로 거절하는데 헤드 0adf15f81c 는 `working` 으로 읽어 `lane-send` 가 보내요. (q3) `Do you want to proceed? Yes, and don't ask again / No, tell Claude what to do differently`(실제 권한 질문을 한 줄로 편 꼴), (q4) `Yes/No: do you want to proceed?`, (q6) `Do you want to delete the build folder now yes/no please confirm it`. 리더는 랜딩 뒤에 안전 분류기가 실제로 좁아지는 것이라 이 상태로 랜딩하지 않기로 했어요(노트 10737). 거절 쪽으로 기울이는 수정이라 SPEC 글은 바꾸지 않았고 승인 digest 도 그대로예요.

- 수리: 질문 줄에서는 선택 짝을 줄 어디에서든 읽어요. 낱말과 낱말 사이의 `/`·`|`(공백이 있든 없든)는 짝이에요(`slash_pair`). 점이나 콜론이 든 토큰(`src/main.rs`, `https://host/x`)은 짝이 아니에요. 낱말 목록이 아니라 모양이에요.
- RED 커밋 e3000a2f45(잠금을 `moai tdd reopen` 으로 풀고 사유 기록): `a_question_row_with_a_pair_of_alternatives_anywhere_is_never_working_or_idle`(일하는 화면·대기 화면 다섯 줄), fake orca 보내기 시험에 세 화면 추가(보내기 호출 0회), 가드 행 둘(`src/main.rs` 와 `and then`)을 `a_question_row_without_a_choice_shape_still_reads_working_or_idle` 에 추가; 확인 `… question_row` 종료 101. 수리 49be680bce 뒤 `moai tdd green` 통과. 단위 시험 `a_question_row_is_read_by_its_tail_after_the_question_mark` 에 행을 더했어요(참 다섯, 거짓 세).
- 대가와 풀지 못한 꼴: `and/or`·`TCP/IP`·`src/lib` 처럼 질문 줄 속의 낱말 짝도 거절해요(질문 줄에서만, 안전한 쪽). 긴 문장 한가운데의 `yes or no`(`/` 없이 `or`)는 이 규칙이 읽지 않아요: 같은 모양의 `Would you like me to commit or push?`(S-N2, 승인된 REQ-7 의 통과 꼴)와 낱말 목록 없이는 가를 수 없어서, 거절하면 S-N2 가 깨져요. 둘을 함께 만족시킬 수 없는 꼴이라 리더에게 알려요(`or` 는 질문 뒤 짧은 꼬리에서만 읽어요).

### 세 번째 수리 (2026-11-27, 델타 리뷰 노트 10777 과 사용자 결정 노트 10779, 개정 3)

두 번째 수리의 델타 리뷰는 PASS 였고, 풀지 못한 꼴 하나가 남았어요: `Do you want to delete the build folder now yes or no please confirm it` 을 팁은 거절하는데 헤드(0506f1bc7f)는 일하는 화면에서 `working`, 대기 화면에서 `idle` 로 읽어 `lane-send` 가 보내요. 모양으로는 `Would you like me to commit or push?`(S-N2)와 가를 수 없어서, 리더가 사용자께 여쭈었고 사용자가 직접 "yes/no 짝만 예외로 거절"을 골랐어요(노트 10779). SPEC 은 한 줄 개정(개정 3: REQ-6 과 D-6 에 예외 하나, 본문 커밋 11516b8093, 승인 블록 커밋 8cfe016870)을 했고 `moai spec show` 의 승인 digest 와 블록의 digest 가 같아요.

- 수리: 질문 줄 안에서 긍정 낱말과 부정 낱말(`yes`/`no`, `y`/`n`, 대소문자·순서 무관, 둘레 문장부호 허용)이 `or` 로 바로 이어졌으면 선택이 있는 줄로 읽어요(`yes_no_or`, `question_row_shaped` 가 부름). 다른 낱말 목록은 되살리지 않았어요. 다른 `or` 짝(`commit or push`, `yes or maybe`)과 `yes to one or no to two` 처럼 `or` 가 더 긴 구절을 잇는 줄은 읽지 않아요(가드 시험). 리뷰의 비차단 줄도 고쳤어요: `slash_pair` 의 문서 주석이 "둘째 슬래시가 든 토큰은 경로"라고 썼는데 코드는 조각이 모두 글자뿐이면 짝으로 읽어서 `crates/core/src` 도 짝이었어요. 주석과 문서를 코드에 맞췄고 단위 시험에 못박았어요.
- RED 커밋 a43d23d2f9(잠금을 `moai tdd reopen` 으로 풀고 사유 기록): `a_yes_no_pair_joined_by_or_in_a_question_row_is_never_working_or_idle`(일하는 화면·대기 화면 여섯 줄: 대소문자, 순서 바꿈, `y or n`, `Y OR N`, `(yes, or no)`), fake orca 보내기 시험에 두 줄 추가(보내기 호출 0회), 가드 행 셋 추가(다른 `or` 짝, `yes or maybe`, 떨어진 두 낱말); 확인 `cargo test --offline --locked --test lane_screen_followups_cli question_row` 종료 101(새 시험과 보내기 시험이 실패, 나머지 셋 통과). 수리 2b11e59e60 뒤 `moai tdd green` 통과. 단위 시험 `a_yes_no_pair_joined_by_or_is_read_by_its_words`, `a_path_of_plain_folder_names_reads_as_a_pair_of_alternatives`(src/lane_screen.rs)도 더했어요.
- 대가: 질문 줄 속의 긍정·부정 짝 `yes or no`·`y or n` 는 질문의 한가운데에 있어도 거절돼요(안전한 쪽; 예: `Do you want me to explain what happens on yes or no?`). 앞선 대가(`and/or`, 폴더 이름만의 경로가 든 질문 줄 거절)는 그대로이고 문서(`lane_send.md`, `lane_screen.md`)에 적었어요.
- 열린 점: 승인된 REQ-7 이 푸는 질문 아래 번호 없는 두 줄(p3)은 이 개정의 범위가 아니라 그대로예요(리더 노트 10779 와 델타 리뷰 10777 의 비차단 줄).
