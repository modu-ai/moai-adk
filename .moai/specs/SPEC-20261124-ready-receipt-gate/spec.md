---
id: SPEC-20261124-ready-receipt-gate
title: "A READY or REPAIRED note is refused without a PASS check-affected receipt of the same head"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7687 (2026-11-23T16:31:42Z, 사용자 위임 2026-11-23T13:57Z, 노트 7686 의 결정 1 (a) 판 1 승인과 FAIL 때도 summary.txt·같은 head 워크트리 여럿 허용의 다듬기, 결정 2 (a), 결정 3 미리 승인)과 노트 7693 (2026-11-23T16:42:32Z, 영수증 요약 줄의 base 가 노트 head 의 조상이어야 한다는 한 줄을 판 1 안의 다듬기로 승인: REQ-11, AC-11, D-7); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:7dd9648060976b200f6f1ad5b6bf6d7119bfcfcb62f94d264d292f7dd2b11817"
---

# A READY or REPAIRED note is refused without a PASS check-affected receipt of the same head

## Intent

카드 C-414 의 재발 방지예요. READY 노트가 그 head 의 `check affected` PASS 영수증 없이 올라오는 일이 같은 종류로 두 번 이상 있었어요. C-409 조각 1 에서는 영수증이 FAIL 이었던 head `16b01f6d05` 뒤에 좁은 테스트만 다시 돌리고 READY `8594066284` 를 올렸고, 그 head 의 영수증도 FAIL 이었어요(`state_dir_tests` 한 번, `home_root` 스캔 테스트 한 번, 감사 #409 의 N4 와 B1). 같은 규칙("READY 는 같은 head 의 PASS 영수증과 함께")은 이미 기억 메모와 CLAUDE.md 의 Flow 6 에 적혀 있는데도 되풀이됐어요. 글로 적은 규칙은 같은 일을 막지 못했으니 명령이 거절하게 소스를 바꿔요.

읽어서 확인한 사실(코드, 2026-11-24):

- `moai check affected` 는 단계 로그와 `heavy-slot.log` 만 `target/check-affected-<head12>/` 에 써요. 요약 줄(`result=PASS`)은 답으로 화면에만 나오고 파일로 남지 않아요(`src/check_affected.rs:1085-1097`, 폴더 목록으로 확인). 그래서 영수증을 "읽기만" 해서는 PASS 를 알 수 없고, 요약 줄을 파일로 남기는 한 줄이 먼저 필요해요(D-1).
- `moai card note` 는 한 곳(`src/card/work.rs` 의 `note`)에서 노트를 받고, 첫 줄 모양은 `LANE-READY-INTEGRATION <카드> head=<40 hex>` 이에요(`src/bundle.rs` 의 `parse_ready` 가 같은 모양을 읽어요). `REPAIRED <카드> head=<40 hex>` 도 같은 모양으로 쓰여 왔어요.
- 문서만 바꾼 head 에서도 `check affected --list` 는 전체 대상(151개)을 골라요(`c1db8741f8~1..c1db8741f8`, SPEC 파일 하나만 바뀐 범위에서 측정). 그래서 영수증을 만들 수 없는 head 가 따로 있지 않아요(D-4).

## Interview

- Q1: 거절은 어디서 하나요? → A: `card note` 한 곳이에요. 레인이 READY 를 올리는 길이 이 명령 하나뿐이라 가장자리가 하나예요(D-2). `card ready` 는 카드를 `draft` 에서 `ready` 로 바꾸는 명령이라 READY 노트와 관계가 없어요.
- Q2: 영수증은 어디서 찾나요? → A: 노트의 head 를 HEAD 로 둔 워크트리의 `target/check-affected-<head12>/summary.txt` 예요. 워크트리는 `git worktree list --porcelain` 로 찾아요(D-3). 카드 예약에 적힌 워크트리는 쓰지 않아요 — 예약은 한 번 적히면 바뀌지 않고 레인이 다른 워크트리에서 돌릴 수도 있어요.
- Q3: 영수증을 손으로 만들면요? → A: 막지 못해요. 이 검사는 "돌리지 않고 올림"을 막는 장치이고 위조를 막는 보안 경계가 아니에요. 감사자가 같은 head 에서 다시 돌리는 일은 그대로 남아요(Risks).
- Q4: 우회 플래그는요? → A: 없어요(카드 본문). 문서만 바꾼 head 도 같은 영수증이 있어야 해요(D-4).

## Non-goals

- 영수증 위조 방지, 영수증의 base 가 통합 브랜치 끝과 같은지 확인(NeXT-Flow 는 팁을 쫓지 않아요).
- `LANE-READY-INTEGRATION` 이나 `REPAIRED` 가 아닌 노트의 거절(HEAD-MOVED, LANE-BLOCKED, AUDIT, LEADER 노트 등은 그대로 받아요).
- `moai card land` 나 `moai bundle` 이 영수증을 다시 읽는 일. 그 둘은 이미 감사 verdict 와 SPEC 검사를 읽어요.
- `target/` 이 아닌 곳에 영수증을 두거나 `--out` 으로 옮긴 폴더를 찾는 일(D-3).
- `src/check_affected.rs` 의 단계 선택·실행 바꾸기(C-406 이 그 파일을 쥐어요, D-1).

## Requirements

- REQ-1: WHEN the first line of a note starts with `LANE-READY-INTEGRATION ` or `REPAIRED `, the `card note` command shall read the card and the head from that line as `<marker> <card> head=<40 hex>` followed by the end of the line or a space.
- REQ-2 (REQ-1): IF the line does not read so or names another card than the one the note is on, THEN the `card note` command shall refuse the note and write nothing to the ledger.
- REQ-3 (REQ-1): IF none of the worktrees of the project that have the named head as their HEAD holds a PASS receipt of that head, THEN the `card note` command shall refuse the note and write nothing to the ledger.
- REQ-4: The refusal of the `card note` command shall name the head, each worktree it looked in with its receipt path and what that receipt says (missing, or its result), and the command that makes the receipt.
- REQ-5: The receipt of a head shall be the file `summary.txt` in `target/check-affected-<head12>/` of a worktree, whose first line is one `check-affected` summary line naming that head and ending in `result=PASS`.
- REQ-6 (REQ-5): IF a receipt is missing, unreadable, longer than one line, names another head or ends in `result=FAIL`, THEN the `card note` command shall count it as no receipt.
- REQ-7: WHEN `moai check affected` runs and does not only list, it shall write its `summary` line, whether the result is PASS or FAIL, to `summary.txt` in its step-log folder.
- REQ-8 (REQ-7): IF `summary.txt` cannot be written, THEN `moai check affected` shall refuse with the path and the reason.
- REQ-9: WHEN the first line of a note starts with neither marker, the `card note` command shall accept or refuse it exactly as before. · failure: none — such a note takes the path it took before this SPEC
- REQ-10: The `card note` command shall have no flag, word or kind that skips the receipt check.
- REQ-11 (REQ-5): IF the `base` of the summary line of a receipt is not an ancestor of the head the note names, THEN the `card note` command shall count it as no receipt and say so.

## Acceptance

- AC-1 (REQ-7): 작은 크레이트가 든 임시 저장소에서 `moai check affected` 를 돌리면(임시 HOME, 임시 슬롯 폴더) `target/check-affected-<head12>/summary.txt` 가 있고 그 첫 줄이 답의 `summary` 와 같아요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(check_affected_leaves_a_summary_file)'`
- AC-2 (REQ-3, REQ-5): 같은 저장소에서 PASS 영수증이 있는 head 로 `LANE-READY-INTEGRATION C-1 head=<head>` 노트를 쓰면 받고, `REPAIRED C-1 head=<head>` 도 받아요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_ready_and_a_repaired_note_with_a_pass_receipt_are_kept)'`
- AC-3 (REQ-3, REQ-4): 영수증이 없는 head 의 READY 노트는 거절되고 거절 문장이 head, 찾은 워크트리와 영수증 경로, `moai check affected` 명령을 적고 장부의 노트 수가 그대로예요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_ready_note_without_a_receipt_is_refused_with_the_way_out)'`
- AC-4 (REQ-6): `result=FAIL` 영수증, 다른 head 를 적은 영수증, 두 줄짜리 영수증, 읽을 수 없는(UTF-8 아님) 영수증은 모두 영수증이 없는 것과 같이 거절돼요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_bad_receipt_counts_as_no_receipt)'`
- AC-5 (REQ-3): 영수증이 head A 에 있는데 노트가 head B 를 적으면 거절돼요. 영수증이 링크된 다른 워크트리(그 head 를 HEAD 로 둔 곳)에 있으면 받고, 어느 워크트리도 그 head 를 HEAD 로 두지 않으면 영수증 폴더가 남아 있어도 거절돼요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(the_receipt_is_looked_for_in_the_worktree_that_has_the_head)'`
- AC-6 (REQ-1, REQ-2): 마커로 시작하지만 모양이 틀린 첫 줄(head 짧음, 카드 이름이 다름, `head=` 없음)은 거절돼요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_marker_line_of_the_wrong_shape_is_refused)'`
- AC-7 (REQ-9, REQ-10): `LANE-BLOCKED`, `HEAD-MOVED`, `AUDIT`, 마커 낱말을 둘째 줄에만 가진 노트는 영수증 없이 받아요. 거절을 건너뛰는 플래그가 `card note --help` 에 없어요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(other_notes_are_kept_and_there_is_no_skip_flag)'`
- AC-8 (REQ-8): `summary.txt` 자리에 폴더가 이미 있어서 쓸 수 없으면 `write_summary` 가 그 경로와 이유를 적은 오류를 답해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(write_summary_names_the_path_and_the_reason)'`
- AC-9 (REQ-7, REQ-4, REQ-6): 첫 단계(`git diff --check`)가 실패하는 커밋에서 `moai check affected` 를 돌리면 `summary.txt` 가 있고 `result=FAIL` 이에요. 그 head 의 READY 노트는 거절되고 거절 문장이 그 워크트리의 영수증이 `result=FAIL` 이라고 말해요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_failed_run_leaves_a_fail_receipt_and_the_refusal_says_so)'`
- AC-10 (REQ-3): 같은 head 를 HEAD 로 둔 워크트리가 둘이고(하나는 영수증 없음, 하나는 PASS 영수증) READY 노트를 받아요. 둘 다 없거나 FAIL 이면 거절 문장이 둘 다 적어요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(any_worktree_with_the_head_and_a_pass_receipt_is_enough)'`
- AC-11 (REQ-11): 요약 줄의 `base` 가 노트 head 의 조상이 아닌 영수증(head 와 이어지지 않는 커밋)은 PASS 여도 거절되고 거절 문장이 그 base 가 head 의 조상이 아니라고 말해요. 조상인 base 면 받아요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(a_receipt_whose_base_is_not_an_ancestor_of_the_head_is_refused)'`
- AC-12 (REQ-1..REQ-11): 이 SPEC 이 쓰는 파일 밖의 `src` 파일 diff 가 0이에요. · verify: `git diff --stat <base>..<head> -- src`

## Scope

- write: `src/card/ready_receipt.rs`(새 파일), `src/card/mod.rs`, `src/card/work.rs`, `src/ops/check.rs`, `tests/card_ready_receipt_cli.rs`(새 파일), `tests/common/ready_receipt.rs`(새 파일), `.moai/specs/SPEC-20261124-ready-receipt-gate/spec.md`
- 범위 확장 요청(RED 에서 실제로 깨진 파일만, 그때 리더 노트로 받아요): READY 노트를 `card note` 로 쓰는 기존 테스트 `tests/bundle_ready_cli.rs`, `tests/bundle_body_cli.rs`, `tests/bundle_check_cli.rs`, `tests/bundle_build_cli.rs`, `tests/bundle_verify_cli.rs`, `tests/e2e_temp_home_cli.rs` (이름은 `grep` 로 읽은 후보이고 어느 것이 깨지는지는 unverified)
- read-only: `src/bundle.rs`(`parse_ready` 모양), `src/check_affected.rs`(`target_dir`, 요약 줄 모양), `src/ledger/workflow.rs`
- forbidden: 실제 HOME 과 실제 장부(임시 HOME 만), 네트워크, `CLAUDE.md`, `src/check_affected.rs`(C-406 범위), 영수증을 우회하는 플래그

## Plan

1. RED: 위 AC 의 테스트 이름으로 `tests/card_ready_receipt_cli.rs` 와 `tests/common/ready_receipt.rs`(영수증을 쓰는 도우미) 를 먼저 써서 실패를 봐요.
2. `src/card/ready_receipt.rs`: 마커 줄 읽기, `git worktree list --porcelain` 읽기, 영수증 읽기, 거절 문장, `write_summary`.
3. `src/card/work.rs` 의 `note` 가 장부를 열기 전에 검사하고, `src/ops/check.rs` 가 `run` 의 답을 받아 `write_summary` 를 불러요.
4. GREEN, 그다음 기존 테스트 중 `card note` 로 READY 를 쓰는 것이 깨지면 도우미를 써서 영수증을 먼저 만들게 고치고(범위 확장 요청), `check affected` 를 READY head 에서 한 번.

## Risks

- 문서만 바꾼 head 도 `check affected` 전체 대상을 돌려야 READY 를 올릴 수 있어요(측정: 151개 대상). 이것이 부담이면 문서 전용 head 의 표기를 따로 정해야 하는데, 그 표기가 곧 우회 길이 되어서 이 SPEC 은 만들지 않아요(D-4).
- 영수증 파일은 손으로 쓸 수 있어요(Q3). 이 검사는 같은 head 의 영수증을 "돌렸는지" 를 습관 대신 명령으로 묻는 장치예요.
- 이미 쓰인 영수증 폴더(이 SPEC 이전에 돌린 것)에는 `summary.txt` 가 없어요. 그 head 의 READY 는 `check affected` 를 한 번 더 돌려야 해요(앞으로의 노트에만 적용해요).
- `CARGO_TARGET_DIR` 로 옮긴 `target` 과 `--out` 으로 옮긴 폴더의 영수증은 찾지 않아요(`CLAUDE.md` 가 `CARGO_TARGET_DIR` 를 쓰지 않게 정해요).
- 기존 테스트가 READY 노트를 가짜 head 로 쓰고 있으면 깨져요. 장부에 직접 쓰는 테스트(`card_inbox_cli` 등)는 영향이 없고, `card note` 명령을 부르는 테스트만 도우미로 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 요약 줄을 파일로 남기는 일은 `src/ops/check.rs` 가 해요(`ready_receipt::write_summary` 를 불러요). `src/check_affected.rs` 는 건드리지 않아요 | 그 파일은 C-406 의 쓰기 범위예요(공유 인터페이스는 그 카드가 바꿔요). `check_affected::run` 을 부르는 곳이 `ops/check.rs` 하나뿐이에요(`grep` 로 확인) | `write_summary` 호출 한 줄과 새 파일을 지우기 |
| D-2 | 거절은 `card note`(`src/card/work.rs` 의 `note`)에서만 해요. 마커 두 개: `LANE-READY-INTEGRATION`, `REPAIRED`. 마커로 시작하는 줄이 모양이 틀리면 그것도 거절해요 | 레인이 READY 를 올리는 길이 `card note` 하나예요. 모양이 틀린 줄을 받으면 오타 하나가 우회 길이 돼요 | 마커 목록에서 하나 빼기 |
| D-3 | 영수증 자리는 head 를 HEAD 로 둔 워크트리(`git worktree list --porcelain`)의 `target/check-affected-<head12>/summary.txt` 하나예요 | 영수증 폴더 이름은 `check affected` 가 이미 `<워크트리>/target/check-affected-<head12>` 로 정해요(`src/check_affected.rs:997`). 카드 본문이 "그 head 를 HEAD 로 둔 워크트리" 로 적었어요 | 찾는 자리를 바꾸기 |
| D-4 | 문서 전용 head 의 표기를 만들지 않아요. 문서 전용 head 도 `check affected` 영수증이 있어야 해요 | 문서만 바꾼 범위에서 `--list` 가 전체 대상 151개를 골라 영수증을 만들 수 있다고 측정했어요. 별도 표기는 우회 길이 돼요(카드 본문은 "SPEC 이 정하는 표기"를 허용했고 이 SPEC 은 "없음"으로 정해요) | 표기를 SPEC 개정으로 더하기 |
| D-5 | 영수증 파일 이름은 `summary.txt`, 내용은 답의 `summary` 줄 한 줄과 줄바꿈 하나예요 | 답에 이미 있는 줄을 그대로 써서 새 형식을 만들지 않아요 | 파일 이름 바꾸기 |
| D-6 | `summary.txt` 은 FAIL 일 때도 써서 거절 문장이 FAIL 을 말해요. 같은 head 를 HEAD 로 둔 워크트리가 여럿(감사 슬롯 등)이면 PASS 영수증이 있는 곳 아무 것이나 받아요 | 리더 결정 노트 7687 (판 1 안의 다듬기) | 두 줄(REQ-7 의 `whether…` 와 REQ-3 의 `any`)을 되돌리기 |
| D-7 | 받는 영수증의 요약 줄 `base` 는 노트 head 의 조상이어야 해요(`git merge-base --is-ancestor <base> <head>`, 줄이 적은 12자리 그대로). head 와 같아도 조상으로 봐요. 조상이 아니거나 이 저장소에 없는 커밋이면 영수증이 없는 것과 같아요 | 리더 결정 노트 7693 (같은 갈래의 두 번째 재발: 레인이 적은 base 가 head 의 조상이 아니었어요, 감사 #403 델타 B-7) | REQ-11 과 `ready_receipt.rs` 의 조상 검사 한 곳을 지우기 |

## Result

구현은 브랜치 `spec/ready-receipt-gate-20261124`(기준 NeXT `05fec214a8`)의 커밋 RED `f4e768c054` → GREEN `1e00d3adbb` → 기존 테스트 보정 `735c581d0e` → 판 1 다듬기(노트 7693) `efc4786233` → RED `13ae27677d` → GREEN `c237d13118` 순서예요. RED 는 새 테스트 9개 중 7개 실패(`target/c414/red.log`), 조상 검사 RED 는 10개 중 1개 실패(`target/c414/red2.log`)였어요. 깨끗한 head 에서 `cargo nextest run --offline --locked --test card_ready_receipt_cli --lib -E 'binary(card_ready_receipt_cli) | test(ready_receipt)'` → exit 0, 13 passed(통합 10 + 라이브러리 3, `target/c414/final-new.log`).

- AC-1: 충족. `check_affected_leaves_a_summary_file` passed — 첫 단계가 실패하는 임시 저장소 실행에서 `summary.txt` 첫 줄이 답의 `summary` 와 같아요.
- AC-2: 충족. `a_ready_and_a_repaired_note_with_a_pass_receipt_are_kept` passed.
- AC-3: 충족. `a_ready_note_without_a_receipt_is_refused_with_the_way_out` passed — 거절 문장이 head, 영수증 경로, `missing`, `moai check affected` 를 적고 노트 수가 0 이에요.
- AC-4: 충족. `a_bad_receipt_counts_as_no_receipt` passed — FAIL, 다른 head, 두 줄, UTF-8 아님 네 가지가 모두 거절돼요.
- AC-5: 충족. `the_receipt_is_looked_for_in_the_worktree_that_has_the_head` passed.
- AC-6: 충족. `a_marker_line_of_the_wrong_shape_is_refused` passed — 여섯 가지 틀린 모양.
- AC-7: 충족. `other_notes_are_kept_and_there_is_no_skip_flag` passed — 다섯 노트가 영수증 없이 받아지고 `card note --help` 에 건너뛰는 플래그 낱말이 없어요.
- AC-8: 충족. `write_summary_names_the_path_and_the_reason`(라이브러리) passed.
- AC-9: 충족. `a_failed_run_leaves_a_fail_receipt_and_the_refusal_says_so` passed — 거절 문장이 `result=FAIL` 을 말해요.
- AC-10: 충족. `any_worktree_with_the_head_and_a_pass_receipt_is_enough` passed — 두 워크트리가 문장에 다 나오고, 하나가 PASS 가 되면 받아요.
- AC-11: 충족. `a_receipt_whose_base_is_not_an_ancestor_of_the_head_is_refused` passed — 부모 없는 커밋과 이 저장소에 없는 커밋은 거절, 진짜 base 는 받아요.
- AC-12: 충족. `git diff --stat 05fec214a8..HEAD -- src` → `src/card/ready_receipt.rs`(새 파일), `src/card/mod.rs`, `src/card/work.rs`, `src/ops/check.rs` 네 파일이고 `src/check_affected.rs` 는 건드리지 않았어요.
- 뮤턴트(적용을 `git diff --stat` 으로 확인하고 되돌린 뒤 `git status` 가 깨끗함을 확인): ① `work.rs` 의 `check_note` 호출 삭제 → exit 100(6개 실패) ② `Some("PASS")` 를 `Some(_)` 로 → exit 100(3개 실패) ③ 조상 검사를 `|| true` 로 → exit 100(1개 실패).
- 기존 테스트 영향: 구현 직후 후보 열 개 파일 125개 중 57개가 실패했어요(`target/c414/neighbors.log`). 여섯 파일(`bundle_body_cli`, `bundle_build_cli`, `bundle_check_cli`, `bundle_verify_cli`, `bundle_ready_cli`, `e2e_temp_home_cli`)이었고 리더 노트 7687 결정 3 의 후보 안이에요. 보정 뒤 일곱 파일(새 테스트 포함) 85개 모두 통과(`target/c414/neighbors2.log`). bundle 다섯 파일은 프로젝트 폴더와 저장소가 따로라서 가짜 영수증을 둘 워크트리가 없어 노트만 장부 라이브러리로 직접 쓰는 도우미(`add_note_unchecked`)를 써요(결정 3 의 "도우미 호출로 가짜 영수증" 과 방식이 다르다는 점은 READY 노트에 적었어요). e2e 카드 순환은 실제 워크트리가 있어 `write_receipt` 로 PASS 영수증을 써요.
- 위험한 길(프로세스를 시작하는 변경): `card note` 가 `git worktree list` 와 `git merge-base` 를 임시 저장소에서 실제로 불러요(링크된 워크트리 둘, 부모 없는 커밋, 없는 커밋). 멈춘 git 자식은 테스트하지 않았어요(not_run).
- 범위 밖·미검증: `card note --help`(`src/ops/card.rs`)의 설명은 이 카드 범위 밖이라 그대로예요. `CARGO_TARGET_DIR` 로 옮긴 `target` 은 영수증을 찾지 않아요(Risks). 실제 cargo 가 도는 PASS 실행이 `summary.txt` 를 남기는지는 READY 의 `check affected` 한 번으로 봐요.
