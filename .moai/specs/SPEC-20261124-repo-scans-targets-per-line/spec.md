---
id: SPEC-20261124-repo-scans-targets-per-line
title: "tools/repo_scans.json names each scan's test targets as a sorted array, one per line, and the reader builds the cargo command"
status: draft
revision: 1
created: 2026-11-24
---

# tools/repo_scans.json names each scan's test targets as a sorted array, one per line, and the reader builds the cargo command

## Intent

재발 방지예요(카드 C-455, 재귀 개선 규칙). 2026-11-24 한 시간 안에 착지가 두 번 같은 줄에서 충돌했어요: `tools/repo_scans.json` 의 `reads-skills` 항목은 `checks` 에 2111 자짜리 명령 한 줄(`cargo test --offline --locked --test a --test b …`, `--test` 79 개)을 가지고 있고, 테스트를 더하는 카드마다 그 한 줄을 고쳐요(C-217 `plan_output_sections`, C-159 `project_harness_step`, C-261 `card_hold_gate_cli`). 줄이 하나라서 서로 다른 이름을 더해도 git 이 충돌로 보고, 리더가 손으로 풀었어요. 같은 모양이 `reads-docs`(30), `reads-root-files`(37), `reads-shipped-files`(22), `reads-src`(25), `reads-moai-state`(18), `reads-tools`(16) 에도 있어요.

이 SPEC 은 항목의 cargo 테스트 대상(`--test <이름>`)을 JSON 배열 `test_targets` 로 옮겨요: 한 줄에 하나, 정렬. 읽는 코드가 그 배열에서 `cargo test --offline --locked --test a --test b …` 명령을 만들어요. 두 카드가 서로 떨어진 자리에 이름을 더하면 줄이 겹치지 않아서 `git merge` 가 충돌 없이 합쳐져요. `rules` 목록은 그대로 둬요.

읽는 곳은 셋이에요: `src/repo_scans.rs`(`moai repo-scans`, `moai card land` 의 필요한 스캔), `src/check_affected.rs`(`check affected` 의 대상 선택), `tests/repo_scans_table.rs`(표가 온전한지 지키는 테스트). 셋이 같은 함수로 명령을 만들게 해서 한 곳에서만 어긋나는 일이 없게 해요.

읽는 사람: 레인, 리더, 이 변경의 감사자.

## Interview

- Q1: 새 필드 이름은 무엇인가? -> A: `test_targets` (D-1). 기존 `tests` 는 스캔이 지키는 테스트 함수 이름이고 `checks` 는 명령이라서, 이름이 겹치지 않는 `test_targets` 로 해요.
- Q2: 명령 하나에 `--test` 를 나열하는 형태만 옮기는가? -> A: 네 (D-2). `--lib`, `--bin`, 이름 필터(`-- name`), 파이썬 `unittest` 명령은 그대로 `checks` 에 남아요. 한 항목에 하나뿐이라 충돌 원인이 아니에요.
- Q3: 한 대상뿐인 항목도 옮기는가? -> A: 네 (D-3). 카드 기준이 "모든 항목이 배열로"이고, 한 형태만 읽는 규칙이 가장 단순해요.
- Q4: 옛 형태(`checks` 에 `--test` 나열)를 읽는 길은 남기는가? -> A: 읽는 쪽은 남기고(D-6), 체크인된 표에는 못 쓰게 테스트가 막아요.
- Q5: `tools/repo_scans.json` 을 범위에 넣는가? -> A: 아니요, 통째로는 넣지 않아요. 여러 카드가 기다리는 파일이라 구현이 끝난 뒤 `moai card line-lease` 로만 쓰고(D-5), 변환은 그때의 파일에 대고 해요.

## Non-goals

- `rules`, `scope`, `tests`, `rejects`, `always`, `id` 의 뜻과 모양을 바꾸지 않아요. 스캔 판정(`judge`, `own_hits`)과 `moai repo-scans` 출력 모양(`{remerge, hits, checks, reasons}`)은 그대로예요.
- `--lib`·`--bin`·이름 필터가 붙은 `--test`·파이썬 명령을 `test_targets` 로 옮기지 않아요(Q2). 새 파이썬 테스트도 만들지 않아요.
- 서로 이웃한 이름(사이에 기존 이름이 없는 두 이름)을 두 카드가 더하는 충돌은 없애지 못해요(Risks). 이 SPEC 은 떨어진 자리를 합쳐 줘요.
- `card land` 의 `scan:<id>` 영수증 단계 이름과 요구 방식은 바꾸지 않아요(`SPEC-20261123-land-repo-scans`).
- 병합 드라이버(`hot_file_ops_merge`)나 다른 공유 파일의 형태는 건드리지 않아요.

## Requirements

- REQ-1: When a scan of `tools/repo_scans.json` names `test_targets`, the reader shall build the command `cargo test --offline --locked --test <t1> --test <t2> …` from the array in its order and put it first among the checks of that scan, before the entries of its `checks` · failure: none — the built command is the only output.
- REQ-2: When a scan names no `test_targets`, the reader shall give exactly the entries of its `checks` as before · failure: none — a table in the earlier form keeps working.
- REQ-3 (REQ-1): IF a `test_targets` is not an array of strings that each match `[a-z0-9_]+`, or a scan has neither `test_targets` nor `checks`, THEN the table parser shall refuse the table with a message that names the scan and the field · failure: the judge fails safe as it does for any unreadable table (`remerge: true`, `UNREADABLE` at the landing).
- REQ-4: When `moai repo-scans`, `moai card land` and `moai check affected` read a scan, they shall take its commands from the one function that applies REQ-1 to REQ-3 · failure: none — a test reads the three results for one fixture table.
- REQ-5: When a `test_targets` array of the checked-in table is not strictly ascending in byte order, repeats a name, or has an element that is not alone on its line in the file text, the test `test_targets_are_sorted_unique_and_one_per_line` shall fail · failure: none — the test message names the scan and the first offending name.
- REQ-6: When a `checks` entry of the checked-in table is only `cargo test --offline --locked` followed by `--test <name>` pairs, the test `no_check_is_a_plain_target_list` shall fail · failure: none — the test message names the scan and says to add the names to its `test_targets`.
- REQ-7: When the tests `every_check_names_a_real_target_and_test` and `every_always_scan_runs_the_readers_it_names` read the checks of a scan, they shall read the built commands, so a name in `test_targets` needs `tests/<name>.rs` and counts as run for the `always` rule · failure: none — the earlier test messages stay.
- REQ-8: When two branches each add a different target to the same scan of the checked-in table at places that are not neighbours, `git merge` of the two shall finish without a conflict and keep both targets · failure: none — a test builds the two branches in a temporary repository.
- REQ-9: When a scan of the table is converted, the readers shall select the same targets and report the same `--test` names and the same other commands as before the conversion · failure: none — a difference is a bug of the conversion or the reader.
- REQ-10: When a lane is told to add a check for a reader, the `check affected` stop line and the texts of `docs/lane-rules.md` and `repo_scans.md` and the `about` of the table shall name `test_targets` instead of a `--test` check string · failure: none — a test reads the texts.

## Acceptance

- AC-1 (REQ-1, REQ-2): 단위 테스트가 `test_targets` 배열이 `cargo test --offline --locked --test a --test b` 한 줄이 되어 `checks` 의 앞에 서고, `test_targets` 가 없는 항목은 `checks` 그대로이고, 둘 다 있으면 만든 명령이 먼저이고 나머지가 뒤따르는 것을 가려요 · verify: `cargo nextest run --locked --offline --lib repo_scans::tests::test_targets_build_one_cargo_command_before_the_listed_checks repo_scans::tests::a_scan_without_test_targets_keeps_its_checks_as_listed`
- AC-2 (REQ-3): 배열이 아니거나, 문자열이 아닌 원소나 `[a-z0-9_]+` 밖의 이름(공백, `-`, 대문자, 빈 문자열)이 있거나, `test_targets` 도 `checks` 도 없는 항목은 항목 이름과 필드를 말하며 거절돼요; `test_targets` 만 있고 `checks` 가 없는 항목은 읽혀요 · verify: `cargo nextest run --locked --offline --lib repo_scans::tests::a_malformed_test_targets_is_refused repo_scans::tests::a_scan_with_test_targets_may_omit_checks_and_one_with_neither_is_refused`
- AC-3 (REQ-4): 같은 표 픽스처(옛 형태와 `test_targets` 형태 한 쌍)를 `moai repo-scans`(`checks`), `check affected --list`(선택된 대상), `card land` 의 필요한 스캔(`needed`)이 읽어서 두 형태가 같은 대상 집합과 같은 나머지 명령을 내요 · verify: `cargo nextest run --locked --offline --test repo_scans_cli a_table_with_test_targets_gives_the_same_checks_as_the_listed_command --test check_affected_scans_cli a_table_with_test_targets_selects_the_same_targets_as_the_listed_command`
- AC-4 (REQ-5, REQ-6): 체크인된 표가 통과하고, 정렬이 어긋난 배열, 같은 이름 두 번, 한 줄에 둘 이상, 한 줄짜리 배열(`[...]`)을 가진 표와 `checks` 에 `--test` 만 나열한 명령이 있는 표는 각각 항목 이름이 든 줄로 실패해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table test_targets_are_sorted_unique_and_one_per_line no_check_is_a_plain_target_list`
- AC-5 (REQ-7): 없는 `tests/<이름>.rs` 를 `test_targets` 에 적으면 `every_check_names_a_real_target_and_test` 가 그 명령으로 실패하고, `always` 항목의 규칙 파일이 `test_targets` 에 없으면 `every_always_scan_runs_the_readers_it_names` 가 예전 줄로 실패해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table every_check_names_a_real_target_and_test every_always_scan_runs_the_readers_it_names`
- AC-6 (REQ-8): 임시 저장소에서 표를 커밋하고 가지 둘이 같은 항목의 서로 떨어진 두 자리에 다른 이름을 더하면 `git merge` 가 종료 코드 0 이고 합친 표가 둘 다 가져요; 같은 줄에 둘을 더한 옛 형태는 충돌해요(대조) · verify: `cargo nextest run --locked --offline --test repo_scans_table two_branches_adding_different_targets_merge_without_conflict`
- AC-7 (REQ-9): 변환 전 표(`596b525dbc` 의 `tools/repo_scans.json`)와 변환한 표를 새 읽기 코드로 읽어서, 모든 항목에서 `--test` 이름 집합과 나머지 명령이 같고, `moai check affected --list <범위>` 출력이 같아요; 결과와 입력의 체크섬을 Result 에 기록해요(한 번 재는 측정이고 영구 테스트가 아니에요) · verify: `moai check affected --list 596b525dbc..HEAD` 를 변환 전과 후 표로 돌려 두 출력의 `diff` 가 비는 것과 항목별 비교 스크립트의 종료 코드 0
- AC-8 (REQ-10): `check affected` 의 `--test <stem>` 검사 없음 줄이 `test_targets` 를 말하고, `docs/lane-rules.md` 와 `repo_scans.md` 가 `test_targets` 를 말해요 · verify: `cargo nextest run --locked --offline --lib check_affected::tests::the_missing_check_line_names_test_targets check_affected::tests::the_guidance_says_where_a_new_test_file_registers`
- AC-9 (REQ-1…REQ-10): 범위의 전체 영향 테스트가 통과해요(`moai check affected` 맨몸으로 한 번) · verify: `moai check affected <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261124-repo-scans-targets-per-line/**`, `src/repo_scans.rs`, `src/check_affected.rs`, `tests/repo_scans_cli.rs`, `tests/check_affected_scans_cli.rs`, `docs/lane-rules.md`, `docs/design-review/appendix/operations/repo_scans.md`
- write, by line lease (shared files, D-5): `tools/repo_scans.json`(맨 끝에 한 번: 모든 항목의 `--test` 나열 명령을 `test_targets` 로 옮기고 `about` 문장을 고쳐요)와 `tests/repo_scans_table.rs`(다른 카드 C-160 이 예약을 쥐고 있어요: 새 테스트 셋과 `checks` 읽기 둘의 고침을 패치로 준비해 두고 짧은 임대 하나 안에서 적용해요)
- read-only: `src/card/land/scans.rs`(읽기 함수 결과만 써요), `tools/protected_checks.json`, `tests/areas.toml`, `SPEC-20261123-land-repo-scans`, `SPEC-20261124-repo-scans-precheck`
- forbidden: 스캔 판정(`judge`, `own_hits`)과 `moai repo-scans` 출력 모양을 바꾸는 것, `src/lib.rs`, 새 Python 테스트, 다른 카드의 `tools/repo_scans.json` 항목 내용을 바꾸는 것(변환은 모양만), 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. RED: `src/repo_scans.rs` 의 단위 테스트 넷(AC-1, AC-2), 두 CLI 테스트(AC-3), `tests/repo_scans_table.rs` 의 새 테스트 셋(AC-4, AC-6)을 먼저 쓰고 코드 없이 빨간 것을 확인해요. 표 규칙 둘(AC-4)은 변환 전 표에서 빨간 것이 정상이에요.
2. GREEN, `src/repo_scans.rs`: `pub fn checks_of(entry: &Value) -> Result<Vec<String>, String>`(REQ-1~3)을 더하고 `parse_table` 이 불러요. `Scan.raw` 는 그대로예요.
3. GREEN, `src/check_affected.rs`: `scan_entries` 가 같은 함수로 명령을 얻어 `--test` 낱말을 읽고, 검사 없음 줄(`add … to its checks`)을 `test_targets` 로 바꿔요. 표를 못 읽는 줄(`has a checks list`)도 새 모양을 말해요.
4. `tests/repo_scans_table.rs`(임대): `checks` 읽기 둘(`every_check_names_a_real_target_and_test`, `scans_not_running_readers`)이 `moai_cli::repo_scans::checks_of` 를 쓰게 하고 새 테스트 셋을 더해요(REQ-5~8). 1 의 RED 는 이 파일만 임대 안에서 적용하고, 임대는 명령 하나 동안이라 패치 파일을 미리 준비해 `git apply` 로 넣어요.
5. 글: `docs/lane-rules.md`, `repo_scans.md`.
6. 변환(마지막): `moai card line-lease` 아래서 그때의 `tools/repo_scans.json` 에 변환 스크립트(scratchpad, 커밋하지 않는 일회용)를 돌려요: 모양이 `cargo test --offline --locked (--test <이름>)+` 인 명령의 이름을 정렬·중복 제거해 `test_targets` 로 옮기고, 남는 명령이 없으면 `checks` 키를 빼요; 다른 키·항목 순서·들여쓰기는 그대로예요. 스크립트는 두 번 돌려도 같은 결과예요. 이어서 AC-7 의 측정과 표 테스트를 돌려요.
7. 뮤턴트: 새 가드 테스트마다 코드를 지우고 실패를 봐요. 영향 테스트 한 번, READY.

## Risks

- 구형 `moai` 바이너리는 변환된 표를 못 읽어요. 코드로 확인한 동작: `src/repo_scans.rs` `parse_table` 은 `checks` 키가 없으면 `checks is not a list` 로 거절하고, `check_affected.rs` `scan_entries` 는 `scan.get("checks")?.as_array()?` 에서 `None` 이라 표 전체를 못 읽는 것으로 봐요. 그래서 변환한 표를 구형 바이너리가 읽으면 조용히 빠뜨리지 않고 크게 실패해요(`card land` 는 `UNREADABLE`, `check affected` 는 "the table exists but cannot be read"). 이 SPEC 의 머리가 착지한 뒤 설치된 `moai` 를 다시 빌드하기 전까지 그 바이너리로 도는 `check affected`·착지가 막혀요. 결정 D-8 을 보세요.
- 이웃한 두 이름 충돌: 사이에 기존 이름이 없는 두 이름(예: 마지막 이름 뒤에 둘을 더함)은 줄이 이어져서 git 이 충돌로 봐요. 마지막 원소는 쉼표 때문에 앞 줄도 바뀌어요. 정렬해 두면 대부분 떨어진 자리에 들어가지만 0 은 아니에요: AC-6 은 떨어진 자리만 재고, 이 한계를 Result 의 gap 으로 적어요.
- 착지 직전에 표를 고친 다른 카드와 한 번 충돌해요(옛 한 줄을 지우는 변환과 그 줄을 고친 변경). 해결은 정해져 있어요: 이름을 `test_targets` 의 정렬된 자리에 더하고, 표 테스트의 실패 문구가 그 길을 말해요. 변환 스크립트는 멱등이라 합친 표에 다시 돌려도 같아요.
- 팁과 PR 두 표가 다른 형태일 때 `union` 은 원본이 다른 두 항목을 따로 담아서, 같은 대상이 순서만 다른 두 명령(옛 형태의 순서 없는 나열과 정렬된 새 명령)으로 나와 한 번씩 더 돌 수 있어요. 틀린 결과는 아니고, 전환 기간에만 생겨요.
- 만든 명령의 `--test` 순서가 옛 문자열과 달라져요(정렬). 대상 집합은 같고 `check affected` 는 `BTreeSet` 이라 순서를 보지 않아요; `card land` 의 영수증은 `scan:<id>` 단계 이름과 종료 코드만 봐요(`src/card/land/scans.rs`), 명령 글자는 안 봐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 필드 이름은 `test_targets`, 항목 안에서 `checks` 바로 앞에 둬요 | `tests` 는 스캔이 지키는 테스트 함수, `checks` 는 명령이라 이름이 겹치지 않아요 | 이름 바꾸기 |
| D-2 | 만든 명령은 `cargo test --offline --locked` 뒤에 `--test <이름>` 쌍을 배열 순서로 붙이고, 그 항목의 검사 목록 맨 앞에 서요; `checks` 에는 `--lib`·`--bin`·이름 필터·파이썬 명령만 남아요; 남는 명령이 없으면 `checks` 키를 빼요 | 키가 없으면 구형 읽기 코드가 조용히 지나치지 않고 거절해요(Risks) | 만든 명령을 목록 뒤에 두기 |
| D-3 | 순수 나열 명령 40 개(한 대상뿐인 것 포함)를 모두 옮겨요 | 카드 기준 "모든 항목", 읽는 형태가 하나 | 한 대상 항목은 옛 형태로 되돌리기(읽기는 둘 다 받아요) |
| D-4 | 명령을 만드는 함수는 `repo_scans::checks_of` 하나이고 `src/check_affected.rs` 와 `tests/repo_scans_table.rs` 도 그것을 불러요 | 세 곳이 따로 만들면 한 곳만 어긋나는 일이 생겨요. 앞 SPEC `SPEC-20261124-repo-scans-precheck` D-1 은 `test_tables` 판정의 분리를 말하고 이 함수와 무관해요 | 읽는 곳마다 따로 만들기 |
| D-5 | `tools/repo_scans.json` 과 `tests/repo_scans_table.rs` 는 범위에 넣지 않고 구현이 끝난 뒤 `moai card line-lease` 로 쓰며, 변환은 그때의 파일에 대고 해요 | 여러 카드가 `tools/repo_scans.json` 을 기다리고(리더 지시, `docs/lane-rules.md` 의 공유 파일 규칙) `tests/repo_scans_table.rs` 는 다른 카드가 예약으로 쥐고 있어서 `card take` 가 충돌로 거절했어요(2026-11-24). 다른 카드의 추가를 지우지 않으려면 낡은 복사본이 아니라 그때의 파일을 변환해야 해요 | 없음 |
| D-6 | 읽는 쪽은 옛 형태(`checks` 에 `--test` 나열)도 계속 받고, 체크인된 표에서는 `no_check_is_a_plain_target_list` 가 막아요 | 팁과 PR 표가 다른 형태인 전환 기간과 옛 커밋의 표를 읽는 길(`load_ref`)이 필요해요 | 옛 형태를 거절하기 |
| D-7 | 이름 규칙은 `[a-z0-9_]+`, 정렬은 바이트 순서, 배열은 늘 여러 줄(원소가 하나여도) | 현재 40 개 명령의 모든 이름이 이 규칙을 지켜요(측정). 한 줄에 한 이름이어야 줄이 안 겹쳐요 | 규칙 풀기 |
| D-8 | **리더 결정이 필요해요**: 기본(권장)은 머리 하나로 읽기 코드와 변환을 함께 착지하고 착지 직후 리더가 설치된 `moai` 를 다시 빌드해요. 구형 바이너리는 변환된 표를 크게 거절하고 조용히 빠뜨리지 않아요(Risks). 대안은 이 카드를 읽기 코드로만 착지하고 변환(AC-7 과 표 테스트)을 후속 카드로 미루는 것 | 구형 바이너리의 동작은 `parse_table` 과 `scan_entries` 를 읽어 확인했고 실행해서 재지는 않았어요(한 번 재는 일을 Plan 6 에 넣어요) | 대안으로 전환 |
