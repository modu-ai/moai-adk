---
id: SPEC-20261124-repo-scans-targets-per-line
title: "the repo scan readers build the cargo command from a sorted test_targets array and still read the earlier checks form"
status: approved
revision: 2
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8424 (2026-11-24T09:46:34Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8423 에 답함): 개정 2 승인, head 6e91847a368f1e0305efc22aaff48125664f9a29; REQ-9 가드를 받고, 카드 기준 1·2 는 이 카드에서 not_met 으로 남겨 C-458 로 넘겨요; 레인은 그 노트를 앞머리에 옮겨 적기만 해요"
  digest: "sha256:83d045408983898ac92b2ff47712597f9970287fe495ae754a284900f1f4cf61"
---

# the repo scan readers build the cargo command from a sorted test_targets array and still read the earlier checks form

## Intent

재발 방지예요(카드 C-455, 재귀 개선 규칙). 2026-11-24 한 시간 안에 착지가 두 번 같은 줄에서 충돌했어요: `tools/repo_scans.json` 의 `reads-skills` 항목은 `checks` 에 2111 자짜리 명령 한 줄(`cargo test --offline --locked --test a --test b …`, `--test` 79 개)을 가지고 있고, 테스트를 더하는 카드마다 그 한 줄을 고쳐요(C-217 `plan_output_sections`, C-159 `project_harness_step`, C-261 `card_hold_gate_cli`). 줄이 하나라서 서로 다른 이름을 더해도 git 이 충돌로 보고, 리더가 손으로 풀었어요. 같은 모양이 `reads-docs`(30), `reads-root-files`(37), `reads-shipped-files`(22), `reads-src`(25), `reads-moai-state`(18), `reads-tools`(16) 에도 있어요.

해결의 모양은 항목의 cargo 테스트 대상(`--test <이름>`)을 JSON 배열 `test_targets`(한 줄에 하나, 정렬)로 옮기고, 읽는 코드가 그 배열에서 `cargo test --offline --locked --test a --test b …` 명령을 만드는 거예요. 두 카드가 서로 떨어진 자리에 이름을 더하면 줄이 겹치지 않아서 `git merge` 가 충돌 없이 합쳐져요.

**개정 2 (리더 결정 노트 8422, 2026-11-24T09:44:06Z)**: 이 SPEC 은 두 단계로 나뉘어요. 이 카드는 읽기 코드만 착지해요: 새 `test_targets` 와 옛 `checks` 를 둘 다 읽고, 명령은 함수 하나가 만들어요. `tools/repo_scans.json` 은 이 카드에서 한 글자도 바뀌지 않아요. 표 변환과 `no_check_is_a_plain_target_list` 강제는 후속 카드의 몫이에요(리더가 열어요). 이유: 설치된 `moai` 를 다시 빌드해 `~/.moai` 에 넣는 일은 HOME 변경이라 사용자 선택을 기다려야 해요. 구형 바이너리는 `test_targets` 를 모르기 때문에, 변환한 표로 착지하면 그 사이 모든 레인의 `check affected` 와 리더의 `card land` 가 막혀요(Risks). 그래서 두 형태를 다 읽는 빌드가 설치된 뒤에 변환해요.

읽는 곳은 셋이에요: `src/repo_scans.rs`(`moai repo-scans`, `moai card land` 의 필요한 스캔), `src/check_affected.rs`(`check affected` 의 대상 선택), `tests/repo_scans_table.rs`(표가 온전한지 지키는 테스트). 셋이 같은 함수로 명령을 만들게 해서 한 곳에서만 어긋나는 일이 없게 해요.

읽는 사람: 레인, 리더, 이 변경의 감사자.

## Interview

- Q1: 새 필드 이름은 무엇인가? -> A: `test_targets` (D-1). 기존 `tests` 는 스캔이 지키는 테스트 함수 이름이고 `checks` 는 명령이라서, 이름이 겹치지 않는 `test_targets` 로 해요.
- Q2: 명령 하나에 `--test` 를 나열하는 형태만 옮기는가? -> A: 네 (D-2). `--lib`, `--bin`, 이름 필터(`-- name`), 파이썬 `unittest` 명령은 그대로 `checks` 에 남아요. 한 항목에 하나뿐이라 충돌 원인이 아니에요.
- Q3: 한 대상뿐인 항목도 옮기는가? -> A: 네 (D-3, 후속 변환에서). 카드 기준이 "모든 항목이 배열로"이고, 한 형태만 읽는 규칙이 가장 단순해요.
- Q4: 옛 형태(`checks` 에 `--test` 나열)를 읽는 길은 남기는가? -> A: 네, 읽는 쪽은 계속 받아요(D-6). 체크인된 표에서 못 쓰게 막는 테스트는 변환과 함께 후속 카드에서 켜요.
- Q5: `tools/repo_scans.json` 을 이 카드의 범위에 넣는가? -> A: 아니요 (D-5, 개정 2). 이 카드는 그 파일을 쓰지 않아요; 변환은 후속 카드에서 `moai card line-lease` 로 해요.
- Q6: 표 변환을 이 카드에서 함께 착지하는가? -> A: 아니요 (D-8, 리더 결정 노트 8422). 읽기 코드만 착지하고 변환은 두 형태를 읽는 빌드가 설치된 뒤 후속 카드에서 해요.

## Non-goals

- `rules`, `scope`, `tests`, `rejects`, `always`, `id` 의 뜻과 모양을 바꾸지 않아요. 스캔 판정(`judge`, `own_hits`)과 `moai repo-scans` 출력 모양(`{remerge, hits, checks, reasons}`)은 그대로예요.
- 체크인된 `tools/repo_scans.json` 을 바꾸지 않아요: 항목을 `test_targets` 로 옮기는 변환, `checks` 에 `--test` 만 나열한 명령을 막는 테스트(`no_check_is_a_plain_target_list`), 레인에게 `test_targets` 를 쓰라고 말하는 안내(`check affected` 의 줄, `docs/lane-rules.md`)는 후속 카드의 몫이에요. 구형 `moai` 가 `test_targets` 를 조용히 지나치므로, 그 빌드가 설치되기 전에는 아무도 체크인된 표에 `test_targets` 를 쓰지 않아요(REQ-9 가 막아요).
- `--lib`·`--bin`·이름 필터가 붙은 `--test`·파이썬 명령을 `test_targets` 로 옮기지 않아요(Q2). 새 파이썬 테스트도 만들지 않아요.
- 서로 이웃한 이름(사이에 기존 이름이 없는 두 이름)을 두 카드가 더하는 충돌은 없애지 못해요(Risks).
- `card land` 의 `scan:<id>` 영수증 단계 이름과 요구 방식은 바꾸지 않아요(`SPEC-20261123-land-repo-scans`).
- 병합 드라이버(`hot_file_ops_merge`)나 다른 공유 파일의 형태는 건드리지 않아요.

## Requirements

- REQ-1: When a scan of a table names `test_targets`, the reader shall build the command `cargo test --offline --locked --test <t1> --test <t2> …` from the array in its order and put it first among the checks of that scan, before the entries of its `checks` · failure: none — the built command is the only output.
- REQ-2: When a scan names no `test_targets`, the reader shall give exactly the entries of its `checks` as before · failure: none — a table in the earlier form keeps working.
- REQ-3 (REQ-1): IF a `test_targets` is not an array of strings that each match `[a-z0-9_]+`, or a scan has neither `test_targets` nor `checks`, THEN the table parser shall refuse the table with a message that names the scan and the field · failure: the judge fails safe as it does for any unreadable table (`remerge: true`, `UNREADABLE` at the landing).
- REQ-4: When `moai repo-scans`, `moai card land` and `moai check affected` read a scan, they shall take its commands from the one function that applies REQ-1 to REQ-3 · failure: none — a test reads the three results for one fixture table.
- REQ-5: When a `test_targets` array of a table is not strictly ascending in byte order, repeats a name, or has an element that is not alone on its line in the file text, the test `test_targets_are_sorted_unique_and_one_per_line` shall fail · failure: none — the test message names the scan and the first offending name.
- REQ-6: When the tests `every_check_names_a_real_target_and_test` and `every_always_scan_runs_the_readers_it_names` read the checks of a scan, they shall read the built commands, so a name in `test_targets` needs `tests/<name>.rs` and counts as run for the `always` rule · failure: none — the earlier test messages stay.
- REQ-7: When two branches each add a different target to the same scan of a table in the new form at places that are not neighbours, `git merge` of the two shall finish without a conflict and keep both targets · failure: none — a test builds the two branches in a temporary repository.
- REQ-8: When the table of `596b525dbc` is converted in a copy outside the repository, the readers shall select the same targets and report the same `--test` names and the same other commands as before the conversion · failure: none — a difference is a bug of the conversion or the reader.
- REQ-9: When the checked-in `tools/repo_scans.json` names a `test_targets`, the test `the_checked_in_table_names_no_test_targets_before_the_conversion` shall fail · failure: none — the test message says the conversion waits for an installed build that reads both forms, and the follow-up removes this test.

## Acceptance

- AC-1 (REQ-1, REQ-2): 단위 테스트가 `test_targets` 배열이 `cargo test --offline --locked --test a --test b` 한 줄이 되어 `checks` 의 앞에 서고, `test_targets` 가 없는 항목은 `checks` 그대로이고, 둘 다 있으면 만든 명령이 먼저이고 나머지가 뒤따르는 것을 가려요 · verify: `cargo nextest run --locked --offline --lib repo_scans::tests::test_targets_build_one_cargo_command_before_the_listed_checks repo_scans::tests::a_scan_without_test_targets_keeps_its_checks_as_listed`
- AC-2 (REQ-3): 배열이 아니거나, 문자열이 아닌 원소나 `[a-z0-9_]+` 밖의 이름(공백, `-`, 대문자, 빈 문자열)이 있거나, `test_targets` 도 `checks` 도 없는 항목은 항목 이름과 필드를 말하며 거절돼요; `test_targets` 만 있고 `checks` 가 없는 항목은 읽혀요 · verify: `cargo nextest run --locked --offline --lib repo_scans::tests::a_malformed_test_targets_is_refused repo_scans::tests::a_scan_with_test_targets_may_omit_checks_and_one_with_neither_is_refused`
- AC-3 (REQ-4): 같은 표 픽스처(옛 형태와 `test_targets` 형태 한 쌍)를 `moai repo-scans`(`checks`), `check affected --list`(선택된 대상), `card land` 의 필요한 스캔(`needed`)이 읽어서 두 형태가 같은 대상 집합과 같은 나머지 명령을 내요 · verify: `cargo nextest run --locked --offline --test repo_scans_cli a_table_with_test_targets_gives_the_same_checks_as_the_listed_command --test check_affected_scans_cli a_table_with_test_targets_selects_the_same_targets_as_the_listed_command`
- AC-4 (REQ-5): 모양을 가리는 순수 함수의 단위 테스트가 정렬이 어긋난 배열, 같은 이름 두 번, 한 줄에 둘 이상, 한 줄짜리 배열(`[...]`)을 항목 이름이 든 줄로 말하고, 올바른 배열에는 줄이 없어요; 체크인된 표에는 배열이 없어서 통과해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table test_targets_are_sorted_unique_and_one_per_line`
- AC-5 (REQ-6): 없는 `tests/<이름>.rs` 를 `test_targets` 에 적은 픽스처 표는 `every_check_names_a_real_target_and_test` 의 판정으로 실패하고, `always` 항목의 규칙 파일이 `test_targets` 에 있으면 `every_always_scan_runs_the_readers_it_names` 의 판정이 지나가고 없으면 예전 줄로 실패해요; 체크인된 표에 대한 두 테스트는 그대로 통과해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table every_check_names_a_real_target_and_test every_always_scan_runs_the_readers_it_names`
- AC-6 (REQ-7): 임시 저장소에서 `test_targets` 형태의 표를 커밋하고 가지 둘이 같은 항목의 서로 떨어진 두 자리에 다른 이름을 더하면 `git merge` 가 종료 코드 0 이고 합친 표가 둘 다 가져요; 같은 줄에 둘을 더한 옛 형태는 충돌해요(대조) · verify: `cargo nextest run --locked --offline --test repo_scans_table two_branches_adding_different_targets_merge_without_conflict`
- AC-7 (REQ-8): `596b525dbc` 의 `tools/repo_scans.json` 을 저장소 밖 복사본에서 변환하고, 변환 전·후 표를 새 읽기 코드로 읽어서 모든 항목에서 `--test` 이름 집합과 나머지 명령이 같고, 두 표로 돌린 `moai check affected --list 596b525dbc..HEAD` 출력이 같아요; 결과와 입력의 체크섬을 Result 에 기록해요(한 번 재는 측정이고 영구 테스트가 아니에요). 변환 스크립트는 후속 카드를 위해 Result 에 남겨요 · verify: 변환 전·후 표를 읽는 항목별 비교 스크립트의 종료 코드 0 과 두 `check affected --list` 출력의 `diff` 가 빈 것
- AC-8 (REQ-9): 체크인된 표는 `test_targets` 가 없어서 통과하고, 항목 하나에 `test_targets` 를 더한 표는 항목 이름을 말하며 실패해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table the_checked_in_table_names_no_test_targets_before_the_conversion`
- AC-9 (REQ-1…REQ-9): 범위의 전체 영향 테스트가 통과해요(`moai check affected` 맨몸으로 한 번) · verify: `moai check affected <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261124-repo-scans-targets-per-line/**`, `src/repo_scans.rs`, `src/check_affected.rs`, `tests/repo_scans_cli.rs`, `tests/check_affected_scans_cli.rs`, `docs/design-review/appendix/operations/repo_scans.md`
- write, by line lease (shared file, D-5): `tests/repo_scans_table.rs`(다른 카드 C-160 이 예약을 쥐고 있어요: 새 테스트 셋과 `checks` 읽기 둘의 고침을 패치로 준비해 두고 짧은 임대 하나 안에서 적용해요)
- read-only: `tools/repo_scans.json`(이 카드는 쓰지 않아요, D-5), `docs/lane-rules.md`(후속 카드가 고쳐요), `src/card/land/scans.rs`(읽기 함수 결과만 써요), `tools/protected_checks.json`, `tests/areas.toml`, `SPEC-20261123-land-repo-scans`, `SPEC-20261124-repo-scans-precheck`
- forbidden: `tools/repo_scans.json` 을 바꾸는 것, 스캔 판정(`judge`, `own_hits`)과 `moai repo-scans` 출력 모양을 바꾸는 것, `src/lib.rs`, 새 Python 테스트, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. RED: `src/repo_scans.rs` 의 단위 테스트 넷(AC-1, AC-2), 두 CLI 테스트(AC-3)를 먼저 쓰고 코드 없이 빨간 것을 확인해요. `tests/repo_scans_table.rs` 의 새 테스트 넷(AC-4~AC-6, AC-8)은 패치 파일로 준비해 두고 임대 안에서 적용해요.
2. GREEN, `src/repo_scans.rs`: `pub fn checks_of(entry: &Value) -> Result<Vec<String>, String>`(REQ-1~3)을 더하고 `parse_table` 이 불러요. `Scan.raw` 는 그대로예요.
3. GREEN, `src/check_affected.rs`: `scan_entries` 가 같은 함수로 명령을 얻어 `--test` 낱말을 읽어요. 검사 없음 줄과 표를 못 읽는 줄의 글은 바꾸지 않아요(후속 카드).
4. `tests/repo_scans_table.rs`(임대): `checks` 읽기 둘(`every_check_names_a_real_target_and_test`, `scans_not_running_readers`)이 `moai_cli::repo_scans::checks_of` 를 쓰게 하고 새 테스트를 더해요(REQ-5~7, REQ-9).
5. 글: `repo_scans.md` 에 `test_targets` 필드와 읽는 규칙, 그리고 체크인된 표는 후속 카드까지 쓰지 않는다는 것을 적어요.
6. 측정(AC-7): 변환 스크립트(scratchpad 의 일회용, 커밋하지 않아요)로 복사본을 만들어 골든 비교를 한 번 돌리고 결과·체크섬·스크립트를 Result 에 남겨요. 또 구형 바이너리(설치된 `moai`)가 `test_targets` 가 든 표를 읽을 때의 동작을 임시 표로 재서 Risks 의 추정을 확인해요.
7. 뮤턴트: 새 가드 테스트마다 코드를 지우고 실패를 봐요. 영향 테스트 한 번, READY.

## Risks

- 구형 `moai` 는 `test_targets` 를 몰라요. 코드로 확인한 동작: `src/repo_scans.rs` `parse_table` 은 `checks` 키가 없으면 `checks is not a list` 로 거절하고, `check_affected.rs` `scan_entries` 는 `scan.get("checks")?.as_array()?` 에서 `None` 이라 표 전체를 못 읽는 것으로 봐요. 즉 `checks` 를 뺀 항목이 하나라도 있으면 크게 실패하지만, `test_targets` 와 `checks` 가 함께 있는 항목은 `test_targets` 가 조용히 무시돼 그 대상이 빠져요. 이 카드는 체크인된 표를 바꾸지 않고 REQ-9 가 그 표의 `test_targets` 를 막아서 어느 쪽도 일어나지 않아요; 변환은 두 형태를 읽는 빌드가 설치된 뒤 후속 카드가 해요.
- 이웃한 두 이름 충돌(후속 변환 뒤): 사이에 기존 이름이 없는 두 이름(예: 마지막 이름 뒤에 둘을 더함)은 줄이 이어져서 git 이 충돌로 봐요. 마지막 원소는 쉼표 때문에 앞 줄도 바뀌어요. AC-6 은 떨어진 자리만 재고, 이 한계를 Result 의 gap 으로 적어요.
- 후속 변환이 착지 직전에 표를 고친 다른 카드와 한 번 충돌해요(옛 한 줄을 지우는 변환과 그 줄을 고친 변경). 해결은 정해져 있어요: 이름을 `test_targets` 의 정렬된 자리에 더해요. 변환 스크립트는 멱등이라 합친 표에 다시 돌려도 같아요(스크립트는 Result 에 남겨요).
- 팁과 PR 두 표가 다른 형태일 때 `union` 은 원본이 다른 두 항목을 따로 담아서, 같은 대상이 순서만 다른 두 명령으로 나와 한 번씩 더 돌 수 있어요. 틀린 결과는 아니고, 후속 변환 기간에만 생겨요.
- 만든 명령의 `--test` 순서는 정렬이에요. 대상 집합은 같고 `check affected` 는 `BTreeSet` 이라 순서를 보지 않아요; `card land` 의 영수증은 `scan:<id>` 단계 이름과 종료 코드만 봐요(`src/card/land/scans.rs`), 명령 글자는 안 봐요.
- REQ-9 의 가드 테스트는 후속 카드가 변환할 때 지워야 해요: 후속 카드의 첫 줄 기준으로 적어 둬요(Result).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 필드 이름은 `test_targets`, 항목 안에서 `checks` 바로 앞에 둬요(후속 변환에서) | `tests` 는 스캔이 지키는 테스트 함수, `checks` 는 명령이라 이름이 겹치지 않아요 | 이름 바꾸기 |
| D-2 | 만든 명령은 `cargo test --offline --locked` 뒤에 `--test <이름>` 쌍을 배열 순서로 붙이고, 그 항목의 검사 목록 맨 앞에 서요; 후속 변환 뒤 `checks` 에는 `--lib`·`--bin`·이름 필터·파이썬 명령만 남고, 남는 명령이 없으면 `checks` 키를 빼요 | 키가 없으면 구형 읽기 코드가 조용히 지나치지 않고 거절해요(Risks) | 만든 명령을 목록 뒤에 두기 |
| D-3 | 후속 변환은 순수 나열 명령 40 개(한 대상뿐인 것 포함)를 모두 옮겨요 | 카드 기준 "모든 항목", 읽는 형태가 하나 | 한 대상 항목은 옛 형태로 되돌리기(읽기는 둘 다 받아요) |
| D-4 | 명령을 만드는 함수는 `repo_scans::checks_of` 하나이고 `src/check_affected.rs` 와 `tests/repo_scans_table.rs` 도 그것을 불러요 | 세 곳이 따로 만들면 한 곳만 어긋나는 일이 생겨요. 앞 SPEC `SPEC-20261124-repo-scans-precheck` D-1 은 `test_tables` 판정의 분리를 말하고 이 함수와 무관해요 | 읽는 곳마다 따로 만들기 |
| D-5 | 이 카드는 `tools/repo_scans.json` 을 쓰지 않아요; `tests/repo_scans_table.rs` 는 범위에 넣지 않고 `moai card line-lease` 로 써요 | 리더 결정 노트 8422(변환은 후속 카드). `tests/repo_scans_table.rs` 는 다른 카드가 예약으로 쥐고 있어서 `card take` 가 conflict 로 거절했어요(2026-11-24) | 없음 |
| D-6 | 읽는 쪽은 옛 형태(`checks` 에 `--test` 나열)도 계속 받아요; 체크인된 표에서 막는 테스트(`no_check_is_a_plain_target_list`)는 후속 카드가 변환과 함께 켜요 | 팁과 PR 표가 다른 형태인 전환 기간과 옛 커밋의 표를 읽는 길(`load_ref`)이 필요해요; 변환 전에 켜면 현재 표에서 실패해요 | 옛 형태를 거절하기 |
| D-7 | 이름 규칙은 `[a-z0-9_]+`, 정렬은 바이트 순서, 배열은 늘 여러 줄(원소가 하나여도) | 현재 40 개 명령의 모든 이름이 이 규칙을 지켜요(측정). 한 줄에 한 이름이어야 줄이 안 겹쳐요 | 규칙 풀기 |
| D-8 | 읽기 코드만 이 카드로 착지하고, 표 변환과 `no_check_is_a_plain_target_list` 강제는 두 형태를 읽는 빌드가 설치된 뒤 후속 카드(리더가 열어요)가 해요 | 리더 결정 노트 8422(2026-11-24T09:44:06Z, 사용자 위임 2026-11-23T13:57Z, 되돌릴 수 있음): 설치된 `moai` 를 다시 빌드해 `~/.moai` 에 넣는 것은 HOME 변경이라 사용자 선택을 기다려요. 변환을 먼저 착지하면 모든 레인의 `check affected` 와 `card land` 가 구형 바이너리에서 막혀요 | 변환을 같은 머리에 넣기(설치 뒤) |
| D-9 | 체크인된 표에 `test_targets` 가 생기면 실패하는 가드 테스트(REQ-9)를 둬요; 후속 카드가 변환할 때 지워요 | 구형 바이너리가 `test_targets` 가 든 표를 조용히 지나치는 일(Risks)을 막아요 | 테스트 지우기 |

## Result

Measured 2026-11-24 (UTC, `date -u` 10:18:32Z) on macOS in the worktree `.claude/worktrees/repo-scans-targets`, base `596b525dbcdc506d3ea89e3ec1eb43fe9fbac5e0` (the `NeXT` tip the branch was cut from after the C-456 landing; `NeXT` has moved to `221a9d07d7` since and the checks did not follow it). Approval: leader note 8424, digest `sha256:83d045408983898ac92b2ff47712597f9970287fe495ae754a284900f1f4cf61` (the front matter only copies that note). Commits: SPEC revision 2 `6e91847a36`, approval record `e10f9ed5c2`, RED `da66318458`, GREEN `5a7677a7d4`, table tests `64fbc603eb`, operation page `7516b97b9f`. RED (`da66318458`, before the reader change): 5 of the 6 new tests in `src/repo_scans.rs`, `tests/repo_scans_cli.rs` and `tests/check_affected_scans_cli.rs` failed because `test_targets` was ignored; `a_scan_without_test_targets_keeps_its_checks_as_listed` passed, as it guards the earlier form (REQ-2). The table tests of `tests/repo_scans_table.rs` were applied under `moai card line-lease` after C-160 gave the file back (leader note 8431); their RED was a compile error until `checks_of` existed.
- AC-1 met: `repo_scans::tests::test_targets_build_one_cargo_command_before_the_listed_checks` and `a_scan_without_test_targets_keeps_its_checks_as_listed` pass: the built command comes first in the array's order, `checks` as listed follow, a scan without `test_targets` gives its `checks` unchanged.
- AC-2 met: `a_malformed_test_targets_is_refused` (11 bad values: a string, an object, a number, null, an empty string, upper case, a space, `-`, `--lib`, a `::` path) and `a_scan_with_test_targets_may_omit_checks_and_one_with_neither_is_refused` pass; each message names the scan `zz-scan` and the field.
- AC-3 met, with the gap below: `repo_scans_cli::a_table_with_test_targets_gives_the_same_checks_as_the_listed_command` (the same report as the listed form: hits and checks) and `check_affected_scans_cli::a_table_with_test_targets_selects_the_same_targets_as_the_listed_command` (same `rust_targets` and `scans`) pass; nextest run of the six tests: 6 passed, 1890 skipped, exit 0.
- AC-4 met: `test_targets_are_sorted_unique_and_one_per_line` passes (unsorted, repeated, two names on a line, a one-line array and a bad character each give a line that starts with the scan id; the checked-in table gives none).
- AC-5 met: `every_check_names_a_real_target_and_test` (a `test_targets` name needs `tests/<name>.rs`: `repo_scans_table` passes, `zz_no_such_test_file` panics) and `every_always_scan_runs_the_readers_it_names` (a reader in `test_targets` counts as run, one that is not names the old line) pass.
- AC-6 met: `two_branches_adding_different_targets_merge_without_conflict` passes: in a temporary repository `b_cli` and `h_cli` added at two places with names between them merge with exit 0 and the merged text equals the table with all seven names; the same two additions to the one `--test` command of the earlier form conflict.
- AC-7 met, measured once (not a permanent test): `git show 596b525dbc:tools/repo_scans.json` (sha256 `95c3b98c2ac9dda2e8aeb9401a2d6b22246ad70eb517f66aa7362900f1925237`) was converted outside the repository by the script below into a copy (sha256 `3aa3e70bee376a050d53b41a2914113bbaeb8c014a5d744f07f572a6b2f0893a`: 40 scans got `test_targets`). Per scan, 70 of 70: the sorted set of `--test` names and the other commands equal the earlier ones, the other keys equal (0 mismatches). `moai-cli repo-scans` on empty input (the fail-safe path lists every check as the reader builds it) gives 75 and 75 commands, equal as a multiset once each `--test` list is sorted, 161 and 161 `--test` names, equal; the command lists are not byte-identical because the built `--test` lists are sorted. `moai-cli check affected <base>..<head> --list --dir <throwaway repo>` with the earlier table and the converted one gives byte-identical answers (`diff` empty; `scans` `listed:129`, 130 `rust_targets`).
- AC-8 met: `the_checked_in_table_names_no_test_targets_before_the_conversion` passes on the checked-in table and names a scan of a fixture table that has `test_targets`.
- AC-9: the `check affected` receipt of the READY head is in the READY note.
- guard mutants, each applied by exact text on the committed head `7516b97b9f` and shown applied with `git diff --stat` (1 file, 1 insertion, 1 deletion), restored with `git checkout --`, `git status --short` empty after each; the named tests then fail with exit 100: (S1) `.filter(|name| is_target_name(name))` of `checks_of` → `.filter(|_| true)`: `a_malformed_test_targets_is_refused`; (S2) `for target in list` → `list.iter().rev()`: `test_targets_build_one_cargo_command_before_the_listed_checks` and `a_table_with_test_targets_gives_the_same_checks_as_the_listed_command`; (S3) `None if targets.is_some() => {}` → `None if false`: `a_scan_with_test_targets_may_omit_checks_…`, `a_table_with_test_targets_selects_…` and the repo-scans test; (S4) the read of `checks_of` in `check_affected::scan_entries` → the `checks` key only: `a_table_with_test_targets_selects_the_same_targets_as_the_listed_command`; (S5) `parse_table` reads `checks` only: the unit test and `a_table_with_test_targets_gives_the_same_checks_…`; (T1) uniqueness `<=` → `<` in `test_targets_violations`: `test_targets_are_sorted_unique_and_one_per_line`; (T2) order check `<=` → `==`: the same test; (T3) the one-quoted-name rule dropped: the same test; (T4) `scans_naming_test_targets` sees nothing: `the_checked_in_table_names_no_test_targets_before_the_conversion`; (T5) `built_checks` reads `checks` only: `every_check_names_a_real_target_and_test` and `every_always_scan_runs_the_readers_it_names`; (T6) the fixture joins the names on one line: `two_branches_adding_different_targets_merge_without_conflict`. 11 mutants, 11 caught.
- Measured, not predicted (Risks): the installed `moai 4.2648.23 (rust)` reads the converted table worse than the SPEC predicted. `check affected --list` with the converted table answers `ok: true`, `scans: unreadable` and 2 `rust_targets` (130 with the earlier table), exit 0, no refusal; with a one-scan table that names `test_targets` and keeps a `checks` command it answers `scans: listed:0` and the `test_targets` name is missing, with no marker. Only `repo-scans` refuses: `checks is not a list`, `hold the merge`. An older build therefore runs fewer tests without a stop, which is why REQ-9 holds until the follow-up and an installed build reads both forms.
- Gap: the `moai card land` arm of AC-3 has no run of its own: `needed` copies `Scan.checks` of `parse_table` (`src/repo_scans.rs`, `needed_between`), which the CLI test covers, but no landing was run on a table in the new form.
- Gap: AC-7 ran the real table through the readers and `check affected --list` in a throwaway repository with changed paths in `src/`, `tests/`, `docs/`, `skills/`, `tools/` and `.moai/specs/`, not `--list 596b525dbc..HEAD` of this repository, because this card does not write `tools/repo_scans.json`; the real range with the converted table is the follow-up's check.
- Gap: two adjacent additions (no name between them in the array) still conflict, as in Risks; AC-6 measures separate places only.
- Gap: criteria 1 and 2 of the card are `not_met` here and move to the follow-up conversion card (leader note 8424): no scan of `tools/repo_scans.json` names `test_targets` yet and the checked-in sort check has nothing to sort.
- Gap: the whole `cargo test --locked` and the `check affected` of the READY head are not in this Result; the receipt is in the READY note. No dangerous path: nothing is deleted and no process is started beyond `git` in temporary repositories and the cargo test runs.
- Hand-off to the conversion card (follow-up of D-8): run this script on the table of the then-current tip (it is idempotent, and it refuses a table that is not in the two-space `json.dumps` form), remove `the_checked_in_table_names_no_test_targets_before_the_conversion`, add `no_check_is_a_plain_target_list`, point the `check affected` stop lines and `docs/lane-rules.md` at `test_targets`, and rebuild the installed `moai` first (the leader's step). The converter, `python3 -I convert.py IN.json OUT.json` (the module docstring of the measured file and its one-line assert message are shortened here; the measured file's sha256 is `3b439585919e9d5f085f6cc0cfcfa720c6e5873a8651c0b533c8887f6980651c`):

```python
import json
import re
import sys

PLAIN = re.compile(r"^cargo test --offline --locked((?: --test [a-z0-9_]+)+)$")


def convert(text):
    doc = json.loads(text)
    for scan in doc["scans"]:
        if "checks" not in scan:
            continue
        keep, names = [], []
        for command in scan["checks"]:
            if PLAIN.match(command):
                names += re.findall(r"--test ([a-z0-9_]+)", command)
            else:
                keep.append(command)
        if not names:
            continue
        names = sorted(set(names) | set(scan.get("test_targets", [])))
        moved = {}
        for key, value in scan.items():
            if key == "checks":
                moved["test_targets"] = names
                if keep:
                    moved["checks"] = keep
            elif key != "test_targets":
                moved[key] = value
        scan.clear()
        scan.update(moved)
    return json.dumps(doc, indent=2, ensure_ascii=False) + "\n"


if __name__ == "__main__":
    source = open(sys.argv[1], encoding="utf-8").read()
    plain = json.dumps(json.loads(source), indent=2, ensure_ascii=False) + "\n"
    assert plain == source, "the table is not in the 2-space json.dumps form"
    out = convert(source)
    assert convert(out) == out, "the conversion is not idempotent"
    open(sys.argv[2], "w", encoding="utf-8").write(out)
```
