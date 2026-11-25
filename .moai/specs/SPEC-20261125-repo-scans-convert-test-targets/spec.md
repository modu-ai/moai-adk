---
id: SPEC-20261125-repo-scans-convert-test-targets
title: "convert tools/repo_scans.json to per-line test_targets with a jq tool, and turn on the guard that refuses a plain target list"
status: draft
revision: 1
created: 2026-11-25
---

# convert tools/repo_scans.json to per-line test_targets with a jq tool, and turn on the guard that refuses a plain target list

## Intent

`SPEC-20261124-repo-scans-targets-per-line` 이 읽기 코드만 착지했어요(노트 8422 의 (B), 병합 `aa0a009e8b`). 이 SPEC 은 그 후속이에요(후속 작업): 설치된 `moai` 가 두 형태를 읽는 것을 먼저 재고, `tools/repo_scans.json` 의 순수 `--test` 나열 명령을 정렬된 `test_targets` 배열(한 줄에 하나)로 옮겨서, 테스트를 더하는 두 카드가 착지 머지에서 같은 줄로 충돌하지 않게 해요(2026-11-25 에도 네 번 충돌했어요, `docs/later.md`).

측정(2026-11-25, 이 저장소 `71d819ce8a` 의 표, 설치된 `moai 4.2648.46-dev`): 변환한 표(43 항목이 `test_targets` 를 얻어요)를 설치본이 읽어요. 같은 범위 `HEAD~60..HEAD` 의 `moai check affected --list` 출력이 변환 전후 바이트까지 같고(`scans: listed:179`, 대상 408), `moai repo-scans` 가 만든 명령은 `--test` 이름을 정렬한 뒤 35 개와 35 개, 이름 258 개와 258 개로 같아요(정규화한 목록의 sha256 이 같아요). 그래서 앞 SPEC 이 기다리던 조건(설치된 빌드가 두 형태를 읽는다)이 충족돼요.

이 SPEC 은 같은 카드의 세 번째 기준도 닫아요(앞 SPEC 의 감사 후속): 새 형태의 표로 `needed` 를 돌리는 테스트(N1), `check affected` 의 낡은 문구(N2), 빈 `test_targets` 를 받아들이는 일(N5).

## Interview

- Q1: 설치된 `moai` 가 두 형태를 읽는지 어떻게 아는가? -> A: 변환한 표를 던져 버릴 복제본에 넣고 설치본으로 돌려서 쟀어요(Intent의 숫자). `strings` 로 `test_targets` 글자를 센 것은 가설이었고 결정에 쓰지 않았어요.
- Q2: 변환은 무엇이 하는가? -> A: `jq` 프로그램 `convert.jq`(이 SPEC 폴더) 하나가 해요 (D-2). 이 프로젝트는 Python 을 쓰지 않고(사용자 지시), 앞 SPEC 의 변환기는 Python 이었어요. 같은 규칙을 `jq` 로 다시 썼고, 입력이 이미 변환돼 있으면 같은 출력을 내는 멱등이에요.
- Q3: 한 대상뿐인 항목도 옮기는가? -> A: 네 (D-1). 읽는 형태가 하나가 되고 카드 기준이 "모든 항목"이에요.
- Q4: 변환은 착지에서 어떻게 안전한가? -> A: 다른 레인이 이 파일에 줄을 더하고 있어서, 변환은 `moai card line-lease` 안의 짧은 한 커밋이고, 착지 직전 팁의 표에 `jq -f convert.jq` 를 다시 돌려 같은 결과를 얻는 방식이에요 (D-3).
- Q5: 읽는 쪽에서 옛 형태를 거절하는가? -> A: 아니요 (D-4). 옛 커밋의 표를 읽는 길(`load_ref`)과 전환 기간의 두 표 합집합이 필요해서, 읽기는 계속 둘 다 받아요. 막는 것은 체크인된 표뿐이에요(`no_check_is_a_plain_target_list`).
- Q6: 앞 SPEC 의 가드(REQ-9)는 어떻게 되는가? -> A: 지우고 새 가드로 바꿔요 (D-5). 설치된 빌드가 두 형태를 읽으므로 `test_targets` 를 막을 이유가 없어요.

## Non-goals

- 읽기 코드의 동작을 바꾸지 않아요. 예외는 빈 `test_targets` 를 거절하는 한 가지예요(REQ-6).
- `tools/repo_scans.json` 의 `rules`, `scope`, `rejects`, `tests` 필드와 항목 순서를 바꾸지 않아요.
- `source-manifest.json` 해시 충돌(같은 later 줄의 뒷부분)은 이 SPEC 의 범위가 아니에요. 해시를 다시 계산하는 `moai` 명령은 따로 카드가 필요해요.
- 이웃한 두 이름을 두 카드가 더하는 충돌은 없애지 않아요(앞 SPEC 의 Risks 그대로).
- 설치된 `moai` 를 다시 빌드하거나 `~/.moai` 를 쓰지 않아요.

## Requirements

- REQ-1: When `convert.jq` runs on a table, the converter shall move every plain `cargo test --offline --locked --test <t> …` command of a scan's `checks` into that scan's `test_targets` as a sorted, de-duplicated array, in the place of `checks`, keep every other command in `checks` and every other key in its place, and leave a scan with no plain command unchanged · failure: none — the output is the only effect.
- REQ-2: When `convert.jq` runs on its own output, the converter shall give the same bytes as that output, which is also the text `jq .` prints for it · failure: none — the `diff` of AC-1 shows the difference.
- REQ-3: When a `checks` command of the checked-in `tools/repo_scans.json` is a plain `--test` list, the test `no_check_is_a_plain_target_list` shall fail with the scan id and the command · failure: none — the test message names the way out (`test_targets`).
- REQ-4: When the converted table is read, the readers shall select the same targets and give the same `--test` names and the same other commands per scan as the table before the conversion · failure: none — a difference is a bug of the conversion.
- REQ-5: When two branches each add a different target name to the `test_targets` of the same scan of the checked-in table at places that are not neighbours, `git merge` of the two shall finish without a conflict and keep both names · failure: none — a test builds the two branches in a temporary repository.
- REQ-6: IF a scan's `test_targets` is an empty array, THEN `checks_of` shall refuse the table with a message that names the scan and the field · failure: the judge fails safe as for any unreadable table.
- REQ-7: When `needed` reads a repository whose table is in the new form, the function shall name the same scans and the same commands as for the table in the earlier form · failure: none — a test reads both forms of one fixture.
- REQ-8: When `moai check affected` names a rule file that no check runs or says the table cannot be read, the printed line shall name `test_targets` as the place to add the target · failure: none — a test reads the printed line; `docs/lane-rules.md` says the same (AC-7).
- REQ-9: When the guard test `the_checked_in_table_names_no_test_targets_before_the_conversion` of the earlier SPEC exists, this change shall remove it in the same commit as the conversion · failure: none — a leftover guard fails every landing.

## Acceptance

- AC-1 (REQ-1, REQ-2): 변환한 `tools/repo_scans.json` 에 `jq -f .moai/specs/SPEC-20261125-repo-scans-convert-test-targets/convert.jq` 를 다시 돌린 출력이 파일과 같고 `jq .` 의 출력과도 같아요 · verify: `jq -f .moai/specs/SPEC-20261125-repo-scans-convert-test-targets/convert.jq tools/repo_scans.json | diff - tools/repo_scans.json`
- AC-2 (REQ-3): 체크인된 표는 순수 `--test` 나열이 없어서 통과하고, `checks` 에 한 줄을 더한 표는 항목 이름과 명령을 말하며 실패해요; 정렬·중복·한 줄에 하나를 보는 테스트도 체크인된 표에서 통과해요 · verify: `cargo nextest run --locked --offline --test repo_scans_table no_check_is_a_plain_target_list test_targets_are_sorted_unique_and_one_per_line`
- AC-3 (REQ-4): 변환 전 표(`71d819ce8a`)와 변환한 표를 읽어서 76 항목마다 `--test` 이름 집합과 나머지 명령이 같고, 모든 파일이 바뀐 요청에 대한 `moai repo-scans --table` 의 정규화한 명령 목록이 같고, 같은 범위의 `moai check affected --list` 두 출력이 같아요; 설치본과 이 머리를 빌드한 `moai` 둘로 재요. 체크섬과 결과를 Result 에 적어요(한 번 재는 측정) · verify: 항목별 비교 `jq` 의 종료 코드 0, 두 `check affected --list` 출력의 `diff` 가 빈 것
- AC-4 (REQ-5): 체크인된 표의 `reads-skills` 항목의 서로 떨어진 두 자리에 가지 둘이 다른 이름을 더하면 `git merge` 가 종료 코드 0 이고 합친 표가 둘 다 가져요 · verify: `cargo nextest run --locked --offline --test repo_scans_table two_branches_adding_targets_to_the_checked_in_table_merge_without_conflict`
- AC-5 (REQ-6): 빈 `test_targets` 는 항목 이름과 필드를 말하며 거절돼요 · verify: `cargo nextest run --locked --offline --lib repo_scans::tests::an_empty_test_targets_is_refused`
- AC-6 (REQ-7): 임시 저장소에 옛 형태와 새 형태의 표를 각각 커밋해서 `needed(repo, tip, head)` 가 같은 항목과 같은 명령을 내요 · verify: `cargo nextest run --locked --offline --test repo_scans_cli needed_reads_a_table_with_test_targets_like_the_listed_form`
- AC-7 (REQ-8): `check affected` 의 규칙 파일 줄과 표를 읽을 수 없다는 줄이 `test_targets` 를 말하고, `docs/lane-rules.md` 의 해당 문장이 같은 말을 해요 · verify: `cargo nextest run --locked --offline --test check_affected_scans_cli a_rule_file_without_a_check_names_test_targets`
- AC-8 (REQ-9): 앞 SPEC 의 가드 테스트 이름이 저장소에 없어요 · verify: `grep -rn the_checked_in_table_names_no_test_targets_before_the_conversion src tests` 가 아무것도 내지 않고 종료 코드 1
- AC-9 (REQ-1…REQ-9): 범위의 전체 영향 테스트가 통과해요(`moai check affected` 맨몸으로 한 번) · verify: `moai check affected <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261125-repo-scans-convert-test-targets/**`, `src/repo_scans.rs`, `src/check_affected.rs`, `tests/repo_scans_cli.rs`, `tests/check_affected_scans_cli.rs`, `tests/repo_scans_table.rs`, `docs/lane-rules.md`, `docs/design-review/appendix/operations/repo_scans.md`
- write, by line lease (shared file, D-3): `tools/repo_scans.json`(구현 끝에 한 번, 변환기를 돌리는 한 명령)
- read-only: `docs/later.md`(해당 줄을 지우는 일은 착지하는 쪽이 해요), `src/card/land/scans.rs`, `tests/areas.toml`
- forbidden: `rules`·`scope`·`rejects`·`tests` 필드와 항목 순서를 바꾸는 것, `judge`·`own_hits` 와 `moai repo-scans` 출력 모양을 바꾸는 것, `src/lib.rs`, 새 Python(테스트·스크립트·측정 모두), 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 설치된 `moai` 를 다시 쓰는 것, 다른 레인의 파일

## Plan

1. RED: `src/repo_scans.rs` 의 `an_empty_test_targets_is_refused`, `tests/repo_scans_cli.rs` 의 `needed_reads_a_table_with_test_targets_like_the_listed_form`, `tests/check_affected_scans_cli.rs` 의 `a_rule_file_without_a_check_names_test_targets`, `tests/repo_scans_table.rs`(임대 없이 예약 경로) 의 `no_check_is_a_plain_target_list` 와 `two_branches_adding_targets_to_the_checked_in_table_merge_without_conflict` 를 먼저 써서 빨간 것을 확인해요.
2. GREEN, `src/repo_scans.rs`: 빈 배열을 거절해요. `src/check_affected.rs`: 두 문구(규칙 파일 줄, 표를 못 읽는 줄)가 `test_targets` 를 말해요. `docs/lane-rules.md` 와 `repo_scans.md` 가 같은 말을 해요.
3. 앞 SPEC 의 가드 테스트를 지워요(REQ-9). 이 시점의 체크인된 표는 아직 옛 형태라 `no_check_is_a_plain_target_list` 는 빨개요: 변환 커밋과 같은 머리에서 처음 초록이 돼요.
4. 변환: `moai card line-lease <card> --generation <g> --path tools/repo_scans.json -- jq -f … ` 한 번(결과는 임시 파일을 거쳐 옮기는 한 줄 스크립트), 한 커밋.
5. 측정(AC-3): 변환 전 표와 변환한 표로 설치본과 이 머리를 빌드한 `moai` 둘이 재요. 한 번 재는 측정이고 영구 테스트가 아니에요.
6. 뮤턴트: 새 가드마다 코드를 지우고 실패를 봐요. 맨몸 `check affected --card` 한 번, READY.
7. 착지 직전 팁의 표가 바뀌었으면 착지하는 쪽이 `jq -f convert.jq` 를 팁의 표에 다시 돌려 그 결과를 쓰고(멱등), 두 가지 병합 충돌은 이름을 `test_targets` 의 정렬된 자리에 더해서 풀어요.

## Risks

- 설치된 `moai` 가 `test_targets` 를 읽지 못하는 빌드(앞 SPEC 이 측정한 `4.2648.23`)는 변환한 표에서 말없이 대상을 빼요. 이번 측정은 `4.2648.46-dev` 에서만 했어요. 그 사이 빌드가 설치돼 있어도 `moai --version` 이 `4.2648.43` 이상이어야 안전해요(앞 SPEC 착지 `aa0a009e8b` 이후, 이 하한은 읽기 코드가 들어간 빌드 번호에 대한 추정이에요: 번호 43 은 문자열 확인만 했고 읽기는 46 에서 쟀어요). CI 의 릴리스 빌드는 머리에서 빌드해서 읽어요.
- 변환 커밋 직전에 다른 카드가 옛 한 줄을 고친 채 착지하면 변환과 한 번 충돌해요. 해결은 정해져 있어요: `convert.jq` 를 팁의 표에 다시 돌려요(멱등).
- 한 항목의 `--test` 이름이 `[a-z0-9_]+` 밖이면 `convert.jq` 의 정규식에 걸리지 않아 `checks` 에 그대로 남아요. 변환 뒤 `no_check_is_a_plain_target_list` 가 그 줄을 잡아요(AC-2).
- 이웃한 두 이름 충돌은 그대로예요(Non-goals).
- `jq` 의 출력 모양(들여쓰기 2칸, 한 줄에 한 원소)이 표의 현재 모양과 같다는 것은 이 저장소의 표로 쟀어요(`jq .` 와 같음). 다른 모양의 표는 `diff` 가 먼저 알려요(AC-1).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 순수 나열 명령 43 개(한 대상뿐인 것 포함)를 모두 `test_targets` 로 옮겨요 | 카드 기준 "모든 항목", 읽는 형태가 하나 | 한 대상 항목은 옛 형태로 되돌리기(읽기는 둘 다 받아요) |
| D-2 | 변환기는 `jq` 프로그램 `convert.jq` 이고 이 SPEC 폴더에 둬요 | Python 금지(사용자 지시), 멱등, 팁에서 다시 돌릴 수 있어요. `moai` 하위 명령으로 만드는 일은 한 번뿐인 변환에는 과해요(되풀이되면 따로 카드) | `moai` 하위 명령으로 바꾸기 |
| D-3 | `tools/repo_scans.json` 은 범위에 넣지 않고 `moai card line-lease` 안의 한 커밋으로 변환해요 | 공유 파일이에요(`shared_files`); 다른 레인이 이 파일에 줄을 더하고 있어요 | 없음 |
| D-4 | 읽는 쪽은 옛 형태도 계속 받아요; 막는 것은 체크인된 표뿐이에요 | 옛 커밋의 표(`load_ref`)와 전환 기간의 팁·PR 두 표가 필요해요 | 옛 형태를 거절하기 |
| D-5 | 앞 SPEC 의 가드 `the_checked_in_table_names_no_test_targets_before_the_conversion` 를 지우고 `no_check_is_a_plain_target_list` 로 바꿔요 | 앞 SPEC D-9 가 후속 카드에서 지우기로 했어요 | 가드를 되살리고 변환을 되돌리기 |
| D-6 | 빈 `test_targets` 배열은 거절해요 | 앞 SPEC 감사 N5: 빈 배열이 있는 항목은 명령이 하나도 없어서 그 스캔이 말없이 아무 것도 돌리지 않아요 | 빈 배열 받기 |
| D-7 | 두 문구를 `test_targets` 로 고치고 `docs/lane-rules.md` 도 같이 고쳐요; `docs/later.md` 의 해당 줄은 착지하는 쪽이 지워요 | 레인이 옛 형태를 다시 쓰지 않게 안내가 바로 있어야 해요; `docs/later.md` 는 다른 카드가 같이 쓰는 파일이에요 | 문구 되돌리기 |

## Result

Measured 2026-11-25 (UTC) on macOS in the worktree `.claude/worktrees/scans-per-line`, base `71d819ce8a` (the `NeXT` tip the branch was cut from; `NeXT` has moved to `64b948a105` since and the checks did not follow it). Approval: leader delegated approval, note 9390 (`approval_digest sha256:fddb5c749c453374086206d1d6b9d48922231fafc3fb119472316e759d3cc27e`, `digest sha256:77fea51f4aedf2b020a3ff3037b5c35cac3d485cf1f46f286e19824a7e2f5a6d`, revision 1). Commits on the base: SPEC draft `5d37a7416f`, RED `32806b9c98`, GREEN `1383a9ca4e`, documents `b200103b83`, conversion `9d10a180ff` (the conversion, run by `convert.jq` inside `moai card line-lease`, and the removal of the earlier guard in one commit).
- AC-1 met: `jq -f .moai/specs/SPEC-20261125-repo-scans-convert-test-targets/convert.jq tools/repo_scans.json | diff - tools/repo_scans.json` is empty at `9d10a180ff`, and `jq . tools/repo_scans.json | diff - tools/repo_scans.json` is empty too. 43 scans got `test_targets`; no `checks` command is a plain `--test` list any more.
- AC-2 met: `cargo nextest run --locked --offline --test repo_scans_table --no-fail-fast` at `9d10a180ff`: 34 tests run, 34 passed, exit 0 (`no_check_is_a_plain_target_list`, `test_targets_are_sorted_unique_and_one_per_line` and the rest). Before the conversion (`b200103b83`) the same target had exactly the two new table tests failing, as the RED expected.
- AC-3 met, measured once (not a permanent test), `target/scans-conv/golden.sh` (a shell and `jq` script, no Python; the script is not kept): the table of `71d819ce8a` (sha256 `f7afab88…79b1`) against the converted one (sha256 `16bd645c…a8df`), read by the installed `moai 4.2648.48-dev` and by the `moai-cli` built from `9d10a180ff`. Per scan, 76 of 76: the sorted set of `--test` names, the other commands and the other keys are equal (0 mismatches). `moai repo-scans --table` with a request that changes every tracked file gives 35 commands and 258 `--test` names for both tables and both binaries, and the sha256 of the commands (each plain `--test` list sorted) is `e40f6ee4ec642f000f9ae33660f48086da3e2ce3b7eab961859275266bb4045e` in all four runs. `moai check affected HEAD~60..HEAD --list` in a throwaway clone of the base gives byte-identical answers for the four combinations (sha256 `d51129d5…56bd`, `rust_targets` 408, `scans: listed:179`).
- AC-4 met: `two_branches_adding_targets_to_the_checked_in_table_merge_without_conflict` passes (two additions to `reads-skills`, one after its first name and one after its last, merge with exit 0 and both names stay).
- AC-5 met: `repo_scans::tests::an_empty_test_targets_is_refused` passes (with `checks` empty, listed, absent, and through `parse_table`).
- AC-6 met: `repo_scans_cli::needed_reads_a_table_with_test_targets_like_the_listed_form` passes. It passed before the GREEN commit too: the reader of the earlier SPEC already read the new form, so it pins REQ-7 rather than driving a change.
- AC-7 met: `check_affected_scans_cli::a_rule_file_without_a_check_names_test_targets` passes: the rule-file line, the invalid-UTF-8 line and the table-shape line name `test_targets`; `docs/lane-rules.md` line 70 says the reader's test name goes into the entry's `test_targets` (reviewed in the diff of `b200103b83`, no test reads that sentence). One logic addition went with it: the generic unreadable-table line of `tools/repo_scans.json` carries the same hint (`src/check_affected.rs`, `table_step`).
- AC-8 met: `grep -rn the_checked_in_table_names_no_test_targets_before_the_conversion src tests` prints nothing and exits 1.
- AC-9: the `check affected` receipt of the READY head is in the READY note.
- Seven guard mutants, one edit each on a copy restored with `git checkout`, tree clean after each, each run is one named test (nextest exit 100 = the test failed, the logs show no compile error): `an_empty_test_targets_is_refused` (`if list.is_empty()` to `if false`), `needed_reads_a_table_with_test_targets_like_the_listed_form` (`checks_of` ignores `test_targets`), `a_rule_file_without_a_check_names_test_targets` three times (rule-file line back to "to its checks", the unreadable-table hint branch to `if false`, the table-cannot-be-read line without `test_targets`), `no_check_is_a_plain_target_list` (`pairs > 0` to `pairs > 100`), `two_branches_adding_targets_to_the_checked_in_table_merge_without_conflict` (the second addition next to the first). All seven exit 100.
- Gap: `moai card land` was not run with the converted table; `needed` reads it through the same `checks_of` (AC-6 covers that path), and the landing's own scan receipt is the first real run.
- Gap: the installed build measured here is `4.2648.48-dev`; no older build was measured on the converted table (the earlier SPEC measured `4.2648.23` dropping targets silently). A build older than the one that holds the reader of the earlier SPEC (`aa0a009e8b`) must not run `check affected` or `card land` on this table.
- Gap: `NeXT` has moved since the base and its table differs (one `checks` command and one rule file line of `card_leader_release_cli`); the conversion was not run on that table. The lander runs `jq -f .moai/specs/SPEC-20261125-repo-scans-convert-test-targets/convert.jq tools/repo_scans.json` on the merged table (idempotent: a plain list added meanwhile moves into `test_targets`) and checks that `no_check_is_a_plain_target_list` passes, and deletes the line of `docs/later.md` that names the conflict on the long `--test` line.
- Gap: two adjacent additions to one `test_targets` array still conflict (Risks); `source-manifest.json` hash conflicts are not part of this SPEC.
- Gap: the whole `cargo test --locked` is not in this Result; the READY note carries the `check affected` receipt. No dangerous path: nothing was deleted beyond the throwaway clone `target/scans-conv/` of this worktree (removed by a relative path before READY); git ran in that clone and in the worktree; the only write outside the SPEC scope is `tools/repo_scans.json` under the line lease.
