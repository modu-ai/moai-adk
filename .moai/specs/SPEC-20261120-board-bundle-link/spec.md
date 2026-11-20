---
id: SPEC-20261120-board-bundle-link
title: "BOARD 04 — link bundle runs, bundle SHAs and PR heads; two audit layers"
status: draft
revision: 1
created: 2026-11-20
---

# BOARD 04 — link bundle runs, bundle SHAs and PR heads; two audit layers

## Intent

협업 보드 설계(`~/.moai/reports/moai-adk/collab-board-final-design-2026-11-20.md` 3판 5.4·6.1절, 사용자 승인)의 넷째 겹이에요. 지금은 묶음(bundle)을 만든 기록(`run.json`), 묶음의 소스 SHA, 묶음 PR 의 head 가 서로 다른 곳에 따로 있어서, 수리로 PR head 가 바뀐 뒤에도 앞선 감사 통과가 아직 유효해 보여요(`src/audit_gate.rs:13` 은 정확히 그 head 의 PASS 만 인정하지만 카드에서는 그 연결이 안 보여요). 이 SPEC 은 그 셋을 카드 안에서 잇고, head 가 바뀌면 앞선 통과를 "효력 없음"으로 보여 주고, 두 감사 층(SPEC 감사, 묶음 PR 감사)을 따로 보여 줘요. 기록은 BOARD 02의 관측 저장에 두고, 원장 표는 건드리지 않아요. 밖으로 나가는 글에는 승인된 SPEC id 와 링크만 써요.

## Non-goals

- 관측 저장 파일과 수집기: BOARD 02 영역이에요. 이 SPEC 은 그 읽기 API(`open_read`, `read_after`)로 읽을 뿐이에요.
- 타임라인 화면: BOARD 01 영역이에요. 이 SPEC 은 화면이 가져다 쓸 계산 결과를 만들어요.
- GitHub 쓰기(댓글·리뷰·병합)와 감사 게이트 규칙 바꾸기: 하지 않아요. GitHub 는 읽기만 해요.
- 원장 v2, 호스트 훅 설치, BOARD 06(두 호스트 관측·요약 전송): 승인 범위 밖이에요.

## Requirements

- REQ-1: When a bundle's run record and a pull request's stored values name the same bundle, the link view shall show the bundle's source SHA, its target SHA, the run record's command results and the pull request's exact head together.
- REQ-2: When the newest stored head of a pull request differs from the source SHA of a run record attached to it, the link view shall mark every earlier pass for that pull request as no longer in effect and keep it as history.
- REQ-3: When a lane audit and a bundle pull request audit both exist for one bundle, the link view shall show them as two separate layers that each carry their own head SHA · failure: none — a layer with no verdict is shown as having none and never takes the other layer's pass.
- REQ-4: When text is rendered for a place outside the repository, the renderer shall write only approved SPEC ids and links.
- REQ-5 (REQ-4): IF the text that `bundle body` renders still holds a card number or a session id after the machine-made arguments of the run record's commands are replaced, THEN the renderer shall refuse and print nothing.
- REQ-6: While the link view runs, it shall write nothing to the ledger, the observation store or the host.
- REQ-7 (REQ-1): IF the observation store is missing or a stored row cannot be read, THEN the link view shall report the link as unknown and name the unreadable row.
- REQ-8 (REQ-2): IF the head of a stored pass cannot be read, THEN the link view shall show that pass as not in effect.

## Acceptance

- AC-1 (REQ-1): 시험 저장에 묶음 실행 행(소스 SHA, 대상 SHA, 명령 결과)과 PR 행(번호, 상태, 리뷰 판정, head)을 넣으면 한 묶음의 보기에 다섯 값이 함께 나와요 · verify: `cargo test --locked --offline --test board_bundle_link_cli the_link_view_shows_the_run_the_shas_and_the_pr_head_together`
- AC-2 (REQ-2, REQ-8): 같은 PR 에 새 head 행을 더하면 앞선 통과가 "효력 없음"과 이력으로 보이고, 새 head 의 통과는 효력이 있어요. head 를 읽을 수 없는 통과는 효력 없음으로 나와요. 읽기 비교 조건을 `==` 에서 `!=` 로 뒤집은 변이가 이 시험을 실패시켜요 · verify: `cargo test --locked --offline --test board_bundle_link_cli a_head_change_marks_the_earlier_pass_as_no_longer_in_effect`
- AC-3 (REQ-3): 같은 묶음의 SPEC 감사 통과와 묶음 PR 감사 통과가 두 층으로 나뉘어 각자의 head 와 함께 나오고, 한 층의 통과가 다른 층의 통과가 되지 않아요 · verify: `cargo test --locked --offline --test board_bundle_link_cli the_two_audit_layers_stay_apart`
- AC-4 (REQ-4): `bundle body` 가 만든 PR 본문에는 카드 번호 `C-<숫자>` 가 한 줄도 없고 세션 id 도 없어요: 표에 카드 열이 없고, 감사 칸은 판정과 head(와 tier)만 보이고, SPEC 별 결과 줄은 SPEC id 로 시작하고, 실행 기록의 명령 글에 든 카드 번호는 자리표시자 `<lane>` 으로 바뀌어요 · verify: `cargo test --locked --offline --test bundle_body_cli the_body_carries_no_card_number_and_no_session_id`
- AC-7 (REQ-5): 실행 기록·Self-Audit 의 자유 글 값에 카드 번호(`C-` 또는 `c-` 와 숫자)나 세션 id 가 있으면 본문 전체를 거절하고 아무것도 쓰지 않아요(`--out` 파일도 만들지 않아요). 거절 글은 어느 필드인지(예: ``the Self-Audit field `host` ``)만 말하고 값은 되풀이하지 않아요 · verify: `cargo test --locked --offline --test bundle_body_cli a_card_number_or_a_session_id_in_a_free_text_value_refuses_the_body`
- AC-8 (REQ-4, REQ-5): 카드 해석기는 `c-12` 도 받으니 소문자도 같아요: 명령 글 속 `c-12` 는 `<lane>` 으로 바뀌고, 실행 기록의 not-run 사유나 가지 이름 속 `c-12` 는 본문 전체를 거절하고 필드(`the not-run reasons`, `the integration line (the branch name)`)를 말해요. 시험의 판정 함수는 구현의 사본이 아니라 단어 단위로 따로 읽어요 · verify: `cargo test --locked --offline --test bundle_body_cli a_lower_case_card_number_is_replaced_in_a_command_and_refused_anywhere_else`
- AC-5 (REQ-6): 보기는 파일·데이터베이스·프로세스·네트워크를 이름으로 쓰지 않고, 호출해도 입력이 그대로이며 같은 답을 두 번 줘요(원장·관측 저장 파일의 바이트 비교는 저장 모듈이 착지한 뒤 마감 조각에서) · verify: `cargo test --locked --offline --test board_bundle_link_cli the_link_view_writes_nothing`
- AC-6 (REQ-7): 관측 저장이 없거나 행 하나가 읽히지 않으면 연결은 "알 수 없음"이고 읽히지 않은 행이 이름으로 나와요 · verify: `cargo test --locked --offline --test board_bundle_link_cli a_missing_store_or_an_unreadable_row_is_unknown`

## Scope

- write(조각 2 이후): 이 SPEC, `src/board/link.rs`(새), `tests/board_bundle_link_cli.rs`(새), 운영 문서의 연결 문단. 조각 3(리더 결정 3636 의 D-4 (나)): `src/bundle.rs` 의 `body` 와 그 시험 `tests/bundle_body_cli.rs`, `bundle body` 를 설명하는 `docs/design-review/appendix/operations/bundle.md` 문단.
- read-only: `src/board/observation.rs`(BOARD 02 의 읽기 API), `src/audit_gate.rs`, `src/bundle.rs`, `src/bundle/verify.rs`(`run.json` 모양), `src/card/work.rs`(`card pr` 의 `head`)
- forbidden: 원장 스키마·표·색인·journal_mode, 읽기 명령의 어떤 쓰기, 호스트 훅 설치, GitHub 쓰기, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 타인 회수 3회 한도·4시간 임대 바꾸기, 실제 원장·실제 HOME

## Plan

1. 조각 1(이번, 코드 없음): SPEC + PLAN 노트. 리더가 읽고 D-1..D-5 를 확인해요.
2. 조각 2(연결 계산): 저장 행의 내용 모양(`bundle-run@1`, `pr-state@1` 이 읽는 필드)과 `link.rs` 의 순수 함수. 효력 판정은 읽을 때 계산해서 아무것도 쓰지 않아요(D-1). 시험은 시험 저장으로 AC-1·2·3·5·6.
3. 조각 3(밖으로 나가는 글): 렌더러와 거절 규칙, AC-4. D-4 에서 정한 대로 `bundle body` 를 고치거나 새 렌더러만 더해요.
4. 조각 4(마감): 변이 시험, 이웃 검사, 운영 문서, 두 층 표기를 BOARD 01 화면이 쓰는 방법(계약 정렬).

## Risks

- **지금 밖으로 나가는 글에 카드 번호가 있어요.** `src/bundle.rs:954` 의 `body` 가 PR 본문 표에 `Card` 열과 `card note <id>` 를 넣고, 그 안에 `AUDIT C-n …` 판정 줄을 그대로 옮겨요(읽은 코드, 시험으로 돌려 보지는 않았어요). "카드 번호 0건" 기준은 이것을 고쳐야 지켜져요 → D-4.
- `pr-state@1`·`bundle-run@1` 내용 모양은 한 번 정하면 바뀌지 않아요(`SPEC-20261120-board-observation-store` 의 REQ-3, D-5). 수집기(BOARD 02 조각 3)와 이 SPEC 이 같은 필드 이름을 써야 해서, 조각 2 전에 BOARD 02 담당과 카드 노트로 맞춰요 → D-2.
- `src/board/` 는 BOARD 02 가 만든 폴더이고 `pub mod board;` 한 줄도 BOARD 02 몫이에요. 같은 폴더에 `link.rs` 를 더하려면 모듈 선언 순서를 맞춰야 해요(조각 2 착수 전 그 카드의 `card show`·`card check` 로 겹침 확인).
- 묶음 PR 감사의 통과는 GitHub 의 PR 댓글에 있어요(`src/audit_gate.rs` 가 읽는 판정 댓글). 설계의 PR 네 값(번호, 상태, 리뷰 판정, head)에는 그 감사 판정이 따로 없어요 → D-3.
- 같은 head 의 FAIL 과 PASS 는 둘 다 "효력 있음"으로 나와요(이 조각은 head 만 비교하고 `src/audit_gate.rs` 의 순서·감사자 규칙은 보지 않아요: 같은 감사자의 더 나중 PASS 가 FAIL 을 푸는지, 저자 제외). "효력 있음"을 게이트와 같은 뜻으로 읽으면 오해라서, 이 규칙은 BOARD 01 화면과 합치는 마감 조각에서 정해요.
- 읽을 수 없는 행이 있어도 나머지 행으로 계산해요: 가장 새로운 PR 행이 읽히지 않으면 그 직전 행이 지금 head 로 쳐져 `run_effect` 가 `standing` 으로 나올 수 있어요. 그래서 이 보기를 쓰는 쪽(BOARD 01 화면)은 `state` 를 먼저 보고 `unknown` 이면 효력 표시를 믿지 않아요(REQ-7 이 알리는 몫).
- 사소한 모서리 셋을 이 조각에서는 두었어요: 같은 SPEC id 가 한 묶음의 `specs` 에 두 번 있으면 레인 층이 두 번 나와요; head 를 읽을 수 없는 서로 다른 통과 둘은 한 항목으로 합쳐져요(둘 다 효력 없음이라 표시는 같아요); 같은 가지의 수리 전후 실행은 서로 다른 묶음으로 나오고 최신 표시가 없어서 쓰는 쪽이 `run_effect` 로 가려요.
- **밖으로 나가는 글의 거절은 알려진 모양만 막아요**(조각 3 감사 N-1·N-2·N-5). 막는 것: 카드 번호 `C-<숫자>`·`c-<숫자>`(앞이 글자·숫자가 아닐 때)와 세션 id(`src/check_tree.rs` 의 `uuid_hits` 와 같은 모양). 막지 못하는 것: 비슷한 하이픈(`‐`·`‑`·`−`)이나 전각 글자로 쓴 카드 번호, 글자에 붙은 세션 id(`ab<uuid>`), 자유 글 속 노트 번호(`note 3636` 은 도구가 만들지 않는 손글이라 막지 않아요), 타임라인 상태와 로그 원문(REQ-5 에서 뺐어요: `bundle body` 는 그런 필드를 받지 않아요; 연결 보기를 화면에 내는 마감 조각의 렌더러가 같은 거절을 가져야 해요). 손으로 쓰는 값은 그래서 쓰는 쪽(리더)의 책임이 남아요.
- 효력 판정을 읽을 때 계산하니 행이 늘수록 읽기가 길어져요. 조각 2 시험에서 행 수에 따른 시간을 재요(BOARD 02 의 읽기 API 는 `limit` 로 끊어 읽어요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "효력 없음"은 저장하지 않고 읽을 때 계산해요: 통과의 head 가 그 층의 지금 head 와 같을 때만 효력이 있어요 | 설계 5.4(이력으로만 남기고 "효력 없음" 표시)와 5.7 불변식 8(저장을 지우고 다시 모아도 같아야 해요), 읽기 명령은 아무것도 쓰지 않아요 | 효력 표시를 행으로 저장(읽기·쓰기가 섞여서 하지 않아요) |
| D-2 | 연결 열쇠는 SHA 하나예요: 묶음 실행 행의 소스 SHA 와 같은 head 의 PR 행이 그 묶음의 PR 이고(그런 PR 이 둘 이상이면 지금도 그 head 인 PR 을 먼저, 없으면 그 head 를 가장 늦게 보인 PR), 같은 PR 번호의 더 새로운 행이 지금 head 예요. 새 종류 없이 기존 `bundle-run@1`·`pr-state@1`·`local-check@1` 을 읽어요 | 설계 5.3(종류 여섯 고정)과 `run.json`(`pinned`, `tip`, `commands`), `card pr` 가 읽는 `head`(`src/card/work.rs`) | 새 종류 추가(되돌릴 수 있지만 버전은 영구) |
| D-3 | 두 층의 출처가 달라요: SPEC 감사는 원장의 판정 노트(첫 줄 `AUDIT C-n head … verdict: PASS`)를 읽기 전용으로, 묶음 PR 감사는 `pr-state@1` 의 리뷰 판정과 head 를 읽어요. 묶음 PR 감사 판정 댓글을 수집기가 `pr-state@1` 에 싣는지는 BOARD 02 조각 3 에 맞춰요 | 설계 5.4 표(두 층), `src/audit_gate.rs` 헤더(판정 댓글 규칙) | 두 층을 한 줄로 합치기(설계가 금지: 레인 통과가 묶음 감사를 대신하지 않아요) |
| D-4 | **리더가 고르는 결정**: 밖으로 나가는 글의 범위. (가) 새 렌더러와 거절 규칙만 더하고 `bundle body` 는 그대로 둔다(빠르지만 지금 PR 본문의 카드 번호는 남아요, 기준 "0건"을 이 SPEC 이 못 지켜요). (나) `bundle body` 의 `Card` 열과 `card note` id, 판정 줄의 카드 번호를 SPEC id 와 링크로 바꿔요(기준을 지키고 `src/bundle.rs` 한 곳과 그 시험을 고쳐요). 권장 (나) | 설계 승인 결정 3(밖에는 SPEC id 와 링크만), 카드 기준 2("카드 번호 0건") | (나)를 되돌리면 이전 본문 형식 |
| D-5 | 이번 조각에서 코드는 쓰지 않아요: PLAN 노트를 리더가 읽은 뒤 조각 2 를 정해요 | 리더 배차(조각 1 = SPEC 과 PLAN) | 없음 |
| D-6 | 밖으로 나가는 글의 렌더러는 지금 있는 `bundle body`(`src/bundle.rs`)예요(리더 결정 3636 의 (나)). 바꾸는 곳 넷: 머리 표의 `Card` 열 삭제, 감사 칸을 `card note <id>` 와 판정 줄 인용 대신 `` `PASS` on `<head>` (tier N) `` 꼴로, SPEC 별 결과 줄을 `<SPEC id>: …` 로, 실행 기록의 명령 글 속 `C-<숫자>` 를 `<lane>` 으로(기계가 만든 인자라서 거절하지 않고 바꿔요). 그다음 만든 본문 전체를 한 번 더 훑어 카드 번호나 세션 id(UUID 꼴)가 남아 있으면 본문 전체를 거절해요(리더가 손으로 쓰는 Self-Audit 값·가지 이름 등 자유 글이 몫이에요) | 설계 승인 결정 3, 리더 결정 노트 3636; 사실은 head 와 실행 기록에서만 온다는 `bundle body` 의 기존 규칙 | 이전 본문 형식(표의 `Card` 열과 `card note` 인용) |
| D-7 | 조각 3 감사(노트 3818, 막음 1)의 수리: 카드 번호는 `C-` 와 `c-` 둘 다 읽어요(카드 해석기가 `c-12` 도 받아서). 세션 id 찾기는 `src/check_tree.rs` 의 `uuid_hits` 를 쓰고 사본을 지웠어요. 거절 글은 줄 종류로 필드 이름을 말하고 값은 되풀이하지 않아요. REQ-5 는 `bundle body` 몫으로 좁혔고 타임라인 상태·로그 원문은 Risks 로 옮겼어요. 시험의 판정 함수는 구현과 다른 방식(단어 단위)으로 다시 쓰고 알려진 문자열로 고정했어요 | 감사 노트 3818 과 리더의 "같이 고치기" 결정(가역, none) | 소문자 읽기를 끄면 되돌려요(시험 둘이 막아요) |

## Result

조각 2 (연결 계산, 명령 없음; 리더 결정 노트 3636 의 D-4 (나)·필드 이름 확정):

- 내용 모양을 확정했어요(BOARD 02 수집기와 맞추려고 그쪽 카드에 노트 3638 로 알렸어요). `pr-state@1` = `{number, state, head, review: null | {verdict: PASS|FAIL, head}}` (`review.head` 는 그 판정이 말하는 정확한 head), `bundle-run@1` = `{branch, source_sha, target_sha, commands[{check, exit}], not_run[], run_digest, specs[]}`. 둘 다 참조와 종료 코드뿐이고 명령 줄·로그 원문·세션 id·카드 번호가 없어요.
- `src/board/link.rs`: `link(rows, lane) -> LinkView` 순수 함수(표준 라이브러리·serde 만 쓰고 `crate::` 를 쓰지 않아서 `src/lib.rs` 선언 없이 시험돼요). 연결 열쇠는 SHA 하나: 묶음 실행의 소스 SHA 와 같은 head 를 보인 적 있는 PR 이 그 묶음의 PR 이고, 같은 번호의 가장 새 행이 지금 head 예요. 효력은 저장하지 않고 읽을 때 계산해요(통과가 말하는 head 가 그 층의 지금 head 와 같을 때만).
- RED 먼저: `todo!()` 껍데기로 시험 5개가 모두 실패(`0 passed; 5 failed`), 구현 뒤 5개 통과.
- 독립 감사(head 53335a3dba, 차단 0)가 구현을 일부러 망가뜨린 변이 8개가 모두 살아남는다고 알렸어요(seq 정렬 삭제, FAIL 을 PASS 로 읽기, 레인 층의 SPEC 구분 삭제, 읽지 못한 head 의 효력 판정, 같은 head 의 FAIL·PASS 합치기, 소속 PR 을 가장 오래된 행으로, 같은 묶음의 뒤 실행 기록 무시, PR 상태를 첫 행에서). 시험 여섯을 더해 11개가 됐고 그 8개 변이를 모두 시험이 잡아요(변이마다 복원 뒤 `git status` 깨끗).
- 같은 감사의 그 밖의 비차단 둘을 고쳤어요: (N-2) 같은 head 를 보인 PR 이 둘이면 지금도 그 head 인 PR 을 먼저 붙여요(시험 `a_pull_request_that_still_shows_the_head_is_preferred`, 그 선택을 끈 변이를 잡아요). (N-4) AC-5 의 순수성 검사를 금지어 목록에서 허용 목록으로 바꿨어요(`use` 줄이 정확히 serde·serde_json·BTreeMap 셋이고 `std::` 가 그 한 곳뿐이며 `crate::`·`println!` 등이 없어요; `std::fs` 가져오기·`std::io` 경로·`println!` 변이를 잡아요). 시험은 12개예요.
- AC-1·2·3·5·6 met — 변이 5개를 시험이 잡았어요(`!=` 로 뒤집은 PR head 비교 → AC-1·2, `==` 를 `!=` 로 뒤집은 효력 판정 → AC-2·3, 한 층의 통과를 항상 무효로 → AC-3, 읽지 못한 행 이름 빼기 → AC-6, 같은 판정 중복 제거 빼기 → AC-2). 변이마다 `git status` 깨끗하게 복원했어요.
- 돌리지 않은 것(조각 2): 실제 관측 저장과의 연결(저장 모듈은 아직 가지에 없어요; `Row` 변환 한 줄은 마감 조각), `src/lib.rs` 의 모듈 선언(`pub mod board;` 는 BOARD 02 와 순서 맞추기), CI.

조각 3 (밖으로 나가는 글, tier 2, 리더 결정 노트 3636 의 D-4 (나), D-6):

- RED 먼저: `bundle_body_cli` 에 새 시험 둘을 더하고 옛 코드로 돌리면 둘 다 실패(`7 passed; 2 failed`), 구현 뒤 9개 모두 통과(`FAILED` 0). 단위 시험 둘(`card_numbers_are_found_at_word_starts_only`, `a_session_id_is_exactly_five_hex_groups`)이 단어 시작·세션 id 모양 경계를 고정해요.
- 고친 곳(`src/bundle.rs` `body`): 머리 표의 `Card` 열 삭제, 감사 칸은 `` `PASS` on `<head>` `` + `(tier N)`(노트 번호·판정 줄 인용 삭제), SPEC 별 결과 줄은 `<SPEC id>: …`, 실행 기록의 명령 글 속 `C-<숫자>` 는 `<lane>` 으로 치환, 만든 본문 전체를 한 번 더 훑어 카드 번호나 세션 id 꼴이 남으면 본문 전체를 거절(`bundle body refused: …`, `--out` 파일 없음). 옛 형식을 기대하던 시험 네 곳과 `docs/design-review/appendix/operations/bundle.md` 문단을 새 형식에 맞췄어요.
- AC-4·AC-7 met — 변이 6개를 시험이 잡았어요(명령 치환 끄기 → 시험 6개 실패, 최종 훑기 끄기 → `a_card_number_or_a_session_id_in_a_free_text_value_refuses_the_body`, 단어 시작 검사 삭제 → 단위 시험, 세션 id 모양을 `8-4-4-4-11` 로 → 단위 시험, 세션 id 검사 끄기 → 시험 2개, 감사 칸을 옛 문구로 → 시험 3개). 변이마다 복원 뒤 `git status` 깨끗해요.
- 첫 변이 시험은 단위 시험의 클로저 수명 오류로 원본이 컴파일되지 않아 6개 모두 무효였어요(`--lib` 를 커밋 전에 안 돌린 제 실수): 중첩 함수로 고치고(`ea5cb6c41b`) 변이를 처음부터 다시 돌렸어요.
- 영향: 리더가 손으로 쓰는 `--self-audit` 값·`not_run` 사유·가지 이름에 카드 번호나 세션 id 꼴이 있으면 이제 `bundle body` 가 거절해요. 리더가 손으로 쓰는 릴리즈 PR 본문도 같은 규칙을 따라요.
- 돌리지 않은 것: 실제 GitHub 에 올린 PR 본문 확인, CI(PR CI 는 리눅스 전용), macOS 전용 시험.

조각 3 수리 (독립 감사 노트 3818, head 9cb1ab7ab0, 판정 FAIL, 막음 1·권고 4):

- B-1(막음): 카드 번호를 대문자 `C` 만 읽어서 `c-12` 가 본문에 그대로 나갔어요(명령 글·Self-Audit 값·가지 이름 모두). 시험의 판정 함수가 구현 코드를 베껴서 같은 빈틈이 있었어요. 수리: `C`/`c` 둘 다 읽어요; 시험 판정은 단어 단위로 다시 쓰고(`the_judges_of_these_tests_see_both_cases_and_the_session_shape` 가 알려진 문자열로 판정 함수를 고정해요), 소문자 시험(`a_lower_case_card_number_is_replaced_in_a_command_and_refused_anywhere_else`: 명령 속 `c-12 c-13` → `<lane> <lane>`, not-run 사유와 가지 이름 `lane2/c-12` → 거절)을 더했어요.
- N-4: 거절 글이 어느 필드인지 말해요(``the Self-Audit field `host` ``, `the not-run reasons`, `the integration line (the branch name)` …) 값은 되풀이하지 않아요(시험이 값 조각이 글에 없음을 확인해요). N-5: REQ-5 를 `bundle body` 몫으로 좁히고 나머지를 Risks 에 적었어요. N-6: 세션 id 는 `check_tree::uuid_hits` 를 써요(사본 제거; 글자에 붙은 세션 id 는 그 모양 규칙대로 막지 않고 Risks 에 적었어요). N-1: 비슷한 하이픈·전각 표기는 Risks 에 적었어요. N-2(자유 글 속 노트 번호)는 막지 않아요(리더 결정). N-7·N-8 은 대상이 아니거나 후속이에요.
