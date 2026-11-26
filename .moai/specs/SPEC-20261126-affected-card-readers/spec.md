---
id: SPEC-20261126-affected-card-readers
title: "check affected selects the integration tests that read a changed source module"
status: draft
revision: 1
created: 2026-11-26
---

# check affected selects the integration tests that read a changed source module

## Intent

`moai check affected` 가 바뀐 소스를 읽는 통합 테스트를 고르지 못한 일이 두 번 있었어요. 두 번 모두 레인 영수증은 PASS 였고, 빠진 테스트는 이름으로 따로 돌려야 보였어요. 이 SPEC 은 그 두 길을 막되 지금 고르는 것은 하나도 빼지 않아요.

확인한 사실(2026-11-26, NeXT `5a6aa857d3`, 설치본 `moai 4.2648.53-dev` 의 `--list` 와 `src/check_affected.rs` 읽기):

- 범위 `9c8b7fb763..3a6c46fc95`(`SPEC-20261121-resurrected-marker` 의 앞 절반, 바뀐 소스 `src/state_root.rs`): `--list --card` 는 감사 하한이라 전체 세트로 답하고(`card_wide.reason` = `audit floor`) 대상 79개에 `update_end_to_end_cli` 가 없어요. 전체 세트(`select_pick`)는 이름 첫 낱말, 스캔 표, OS 표시만 읽고 테스트 소스가 모듈을 가리키는지는 보지 않아요. `update_end_to_end_cli` 는 `moai_cli::state_root` 를 가져오지 않고 명령 답의 열쇠 `"state_root"` 만 읽어요.
- 범위 `252a6e3be0..ab2d27e1fd`(파일 잠금 제거, 바뀐 소스에 `src/ledger/reservation.rs`): `--list --card` 244개와 `--list` 268개 어느 쪽에도 `prop_reservation` 이 없어요. 이 테스트는 `use moai_cli::ledger::{reservation, workflow};` 로 묶어 가져와서 `names_path` 의 글자 검색에 걸리지 않아요. 묶음 가져오기만 알아봐도 모자라요: 이 모듈에 닿는 대상이 10개(`COMMON_WORD_CAP`)를 넘어 영역 안 대상만 남기는데, `prop` 은 `card` 영역이 아니에요.

읽는 사람: 레인, 이 변경의 감사자, 착지 묶음을 도는 리더.

## Non-goals

- 호출 관계(call graph)를 따라 테스트를 고르지 않아요.
- 넓은 경로(`wide_class`), `NARROWED_FLOOR`, `COMMON_WORD_CAP`, 영역 규칙, 스캔 규칙은 바꾸지 않아요.
- 모듈 경로를 가리키는 모든 테스트를 전체 세트에 넣지 않아요(D-2).
- `--full` 의 뜻(앞선 실행을 읽지 않음), READY 관문, 요약 줄의 낱말 종류는 바꾸지 않아요.

## Requirements

<!-- 이 SPEC 은 선택에 대상을 더하기만 해요. 새로 막는 검사는 없고 약해지는 검사도 없어요. -->
- REQ-1: When `check affected` reads a changed `src/**/*.rs` path that names a library module, the command shall select every test target whose name carries the file's own word (`reservation` for `src/ledger/reservation.rs`) and whose source names that module path, written in one piece (`ledger::reservation`) or as an item of a brace group directly after the parent segment (`ledger::{reservation, workflow}`, on one line, over several lines, or nested inside an outer group), whatever the number of targets that reach the module and whatever its area. · failure: none — 글자 검색을 하는 순수 함수이고 읽지 못한 테스트 파일은 지금과 같게 다뤄요.
- REQ-2: When `check affected` reads a changed `src/**/*.rs` path whose file name without `.rs` holds an underscore (`src/state_root.rs`; `mod.rs`, `lib.rs` and `main.rs` name no module), the command shall select every test target whose source holds that file word as a whole string literal (`"state_root"`). · failure: none — REQ-1 과 같아요.
- REQ-3: The rules of REQ-1 and REQ-2 shall apply to the card set, the full set and the lane set, so that a card run that falls back to the full set still holds them and the lane set stays a subset of the full set. · failure: none — 아래 AC-3 이 기존 선택 테스트로 지켜요.
- REQ-4: The command shall keep in every selection each test target that it selected for the same range before this change. · failure: none — 두 규칙은 합집합으로만 더해요.
- REQ-5: When REQ-1 or REQ-2 adds a target to the card set, the answer shall name the path's reason with the existing word `module:<word>`. · failure: none — 이유 낱말은 표시일 뿐이고 관문이 읽지 않아요.

## Acceptance

- AC-1 (REQ-2, REQ-3): 임시 저장소에서 `src/state_root.rs` 만 바뀐 범위의 `--list --card`(전체 세트로 답해요)와 맨 `--list` 에, 모듈을 가져오지 않고 `"state_root"` 글자만 가진 테스트가 들어 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_selection -E 'test(a_changed_module_selects_the_test_that_reads_its_answer_key)'`
- AC-2 (REQ-1, REQ-3, REQ-5): 임시 저장소에서 `src/ledger/reservation.rs` 가 바뀌고 모듈에 닿는 대상이 10개를 넘을 때, 다른 영역의 테스트 셋(한 줄 묶음, 여러 줄 묶음, 바깥 묶음 안의 묶음으로 가져오는 `prop_reservation` 모양)이 카드 세트, 전체 세트, 레인 세트에 모두 들어 있고, 카드 답의 그 경로 이유에 `module:` 낱말이 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_selection -E 'test(a_grouped_import_of_a_common_module_is_selected)'`
- AC-3 (REQ-1, REQ-2, REQ-3, REQ-4): 기존 선택 테스트가 기대값을 바꾸지 않고 통과해요. 새 테스트 파일의 `a_look_alike_name_or_literal_adds_nothing` 은 이름만 같고 모듈을 가리키지 않는 테스트, `"state_roots"` 글자, 한 낱말 모듈의 글자, 그리고 `src/ledger/reservation.rs` 변경에서 `"reservation"` 이나 `"ledger_reservation"` 글자만 가진 테스트가 더해지지 않는 것을 확인해요 · verify: `cargo nextest run --locked --offline --test check_affected_card_select --test check_affected_card_areas --test check_affected_selection`
- AC-4 (REQ-1, REQ-2): 저장소의 지금 `tests/` 파일로 고르면 `src/state_root.rs` 의 전체 세트에 `update_end_to_end_cli` 가, `src/ledger/reservation.rs` 의 카드 세트와 전체 세트에 `prop_reservation` 이 들어 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_selection -E 'test(the_two_missed_tests_of_this_repository_are_selected)'`
- AC-5 (REQ-4): READY 의 Result 에, 지은 실행 파일로 `9c8b7fb763..3a6c46fc95` 와 `252a6e3be0..ab2d27e1fd` 를 다시 고른 결과와 최근 착지 20건의 대상 수(전, 후)를 적어요. 어느 범위에서도 전에 있던 대상이 빠지지 않아요 · verify: `moai check affected <base>..<head> --list` 와 `--list --card` 를 바꾸기 전 실행 파일과 후 실행 파일로 돌려 대상 목록을 비교
- AC-6 (REQ-1, REQ-2, REQ-4): `moai check affected` 가 READY 의 head 에서 PASS 이고, REQ-1 과 REQ-2 의 코드를 하나씩 지운 변조에서 AC-1, AC-2 의 테스트가 실패해요(READY 의 guard-mutant 줄) · verify: `moai check affected --card <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261126-affected-card-readers/**`, `src/check_affected.rs`, `tests/check_affected_selection.rs`
- write (줄 임대): `tools/repo_scans.json`(저장소의 `tests/` 를 읽는 새 테스트의 등록 줄)
- read-only: `tests/check_affected_card_select.rs`, `tests/check_affected_card_areas.rs`, `tests/areas.toml`, `src/gate.rs`
- forbidden: `src/check_fix.rs`, `skills/**`

## Plan

1. 착지는 `SPEC-20261125-repo-scans-convert-test-targets` 다음이에요(리더 결정 2026-11-26). 구현은 `5a6aa857d3` 에서 하고, 그 SPEC 이 착지했으면 READY 전에 tip 을 합쳐 스캔 표 등록 줄을 tip 의 형식으로 써요. 소스에서는 겹치지 않아요: 그 변경(`08647f0d7e`)은 `scan_violations`·`table_step` 을, `SPEC-20261126-check-affected-trailers-flaky`(`02005c41fe`)는 `table_step`·`plan` 둘레와 `list`·`run_with` 를 바꿔요.
2. 빨간 테스트를 먼저 써요(AC-1, AC-2, AC-4).
3. `src/check_affected.rs` 에 묶음 `use` 를 읽는 함수와 두 규칙의 합집합을 더해요. 손대는 자리는 `names_path` 둘레, `select_pick`, `select_card_with_registry` 의 소스 경로 되풀이예요. 단계 계획과 실행(`table_step`, `plan`, `list`, `run_with`)은 건드리지 않아요.
4. 새 테스트를 스캔 표에 줄 임대로 등록해요.
5. AC-5 의 전후 비교를 재요.

## Risks

- 이 두 규칙으로도 모듈을 이름으로도 글자로도 가리키지 않고 동작만 기대는 테스트는 여전히 빠져요. 착지 묶음의 넓은 세트도 같은 선택 함수를 써서, 그 마지막 보호는 리더가 착지 다섯 번마다 도는 전체 `cargo test` 예요.
- 영수증이 조금 커져요. REQ-2 는 `project_id` 처럼 흔한 밑줄 낱말의 모듈이 바뀌면 많은 테스트를 더할 수 있어요. AC-5 에서 재고, 크면 개정으로 알려요.
- AC-4 의 테스트는 저장소 파일을 읽어서, 두 테스트의 이름이나 글자가 바뀌면 빨개져요. 조용히 빠지는 것보다 낫다고 봤어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 선택에 더하기만 하고 넓은 경로, 상한, 영역 규칙은 그대로 둬요 | 관문을 약하게 만들지 않는 규칙(CLAUDE.md Standing principles). 상한 계산을 바꾸면 어떤 범위에서는 대상이 줄어요 | 두 합집합 줄을 지워요 |
| D-2 | 묶음 가져오기 규칙은 "이름에 파일 낱말 + 모듈 경로" 두 조건을 모두 요구해요. 경로를 가리키는 모든 테스트를 넣는 길은 고르지 않았어요 | `5a6aa857d3` 에서 `moai_cli::state_root` 를 가리키는 테스트 파일이 95개인데 위 `state_root` 범위의 전체 세트는 79개라 두 배 넘게 커질 수 있어요(겹치는 수는 안 쟀어요). 고르려면 리더 결정이에요 | 두 번째 조건만 쓰도록 규칙을 넓혀요 |
| D-3 | 글자 규칙은 밑줄이 있는 파일 낱말에만(폴더를 붙인 낱말이 아니라), 상한 없이 써요 | 한 낱말(`card`, `gate`)은 명령 인자로 흔해서 거의 모든 테스트가 걸려요. 상한을 두면 글자를 가진 파일이 늘 때 `update_end_to_end_cli` 가 조용히 빠져요(`3a6c46fc95` 에서 7개, `5a6aa857d3` 에서 6개) | 상한을 더해요 |

## Result

2026-11-26, 가지 `spec/affected-card-readers-20261126`, 구현 base `5a6aa857d3`.

커밋: 빨간 테스트 `766f301026`(고치기 전 4개 중 3개 실패, exit 100), 스캔 표 등록 `fd5d8d01a7`(`reads-src` 의 규칙 한 줄과 검사 명령 한 곳, 줄 임대), 구현 `b1e37d8aef`.

구현: `src/check_affected.rs` 에 `source_readers`(REQ-1, REQ-2)와 `names_module`·`names_member`(묶음 `use`; `names_common_sub` 는 `names_member` 를 부르고 동작은 같아요)를 더하고, `select_pick` 과 `select_card_with_registry` 에서 합집합으로만 더해요.

테스트(`b1e37d8aef` 의 소스와 `fd5d8d01a7` 의 표): `cargo nextest run --locked --offline --lib --test check_affected_selection --test check_affected_card_select --test check_affected_card_areas --test repo_scans_table -E 'test(check_affected::) | binary(check_affected_selection) | binary(check_affected_card_select) | binary(check_affected_card_areas) | binary(repo_scans_table)'` → 119 run, 119 passed, exit 0. 등록 전 실행은 86 중 85 통과, 1 실패(`the_repository_tests_pass_the_scan_step_as_added_files`: 새 테스트의 스캔 표 등록이 없다는 알림)였어요.

guard-mutant(`--test check_affected_selection`, 하나씩 지우고 되돌림, 뒤의 작업 트리는 깨끗함):

- guard-mutant: `a_changed_module_selects_the_test_that_reads_its_answer_key` | REQ-2 의 글자 검색 | exit 100 (137행 단언)
- guard-mutant: `a_grouped_import_of_a_common_module_is_selected` | REQ-1 의 묶음 읽기 | exit 100 (191행)
- guard-mutant: `a_look_alike_name_or_literal_adds_nothing` | REQ-1 의 이름 낱말 조건 | exit 100 (231행)
- guard-mutant: `the_two_missed_tests_of_this_repository_are_selected` | 전체 세트의 합집합 줄 | exit 100 (270행)
- guard-mutant: `a_grouped_import_of_a_common_module_is_selected` | 카드 세트의 합집합 블록 | exit 100 (191행)

AC-5 전후 비교: 바꾸기 전 실행 파일 sha256 `17361363000507076c82a837c93c62686a9e7541a0678b239d87f064bfaab31a`(`5a6aa857d3` 의 소스), 후 `2e6c8bb54f0f4e274b7daa1f85e2dec881e60df412280bf23d3898bd883dde6c`(`b1e37d8aef` 의 소스). 22개 범위 × `--list`·`--list --card` 44번 모두 빠진 대상 `[]`. 앞의 두 줄이 놓친 범위이고 나머지는 최근 착지 20건(NeXT `91dd561a5a`)이에요.

| 범위 | `--list` 전 → 후 | `--list --card` 전 → 후 |
| --- | --- | --- |
| `9c8b7fb763..3a6c46fc95` | 80 → 84 | 80 → 84 |
| `252a6e3be0..ab2d27e1fd` | 269 → 270 | 244 → 245 |
| `6f874cffa9..de1e2c4e29` | 114 → 114 | 80 → 80 |
| `372dcd9509..6f874cffa9` | 191 → 191 | 191 → 191 |
| `7b284c9d21..372dcd9509` | 209 → 209 | 119 → 119 |
| `5a6aa857d3..7b284c9d21` | 113 → 113 | 2 → 2 |
| `af3f43ad3a..979c176905` | 114 → 174 | 118 → 183 |
| `3dd2ecb6cb..af3f43ad3a` | 101 → 101 | 5 → 5 |
| `016b944397..3dd2ecb6cb` | 56 → 56 | 0 → 0 |
| `e3450e0ede..016b944397` | 164 → 164 | 164 → 164 |
| `120876d7b4..be608d5dfd` | 177 → 177 | 156 → 156 |
| `ded201bde2..120876d7b4` | 82 → 82 | 82 → 82 |
| `b0b0afaaad..ded201bde2` | 166 → 166 | 166 → 166 |
| `398b66bda0..b0b0afaaad` | 92 → 92 | 40 → 40 |
| `3b355d7a64..90ef2cedb9` | 90 → 90 | 90 → 90 |
| `d77e5a70bc..3b355d7a64` | 86 → 86 | 86 → 86 |
| `9c8b7fb763..d77e5a70bc` | 498 → 498 | 154 → 154 |
| `7d79cc6d7d..9c8b7fb763` | 57 → 57 | 1 → 1 |
| `3c55cc18b7..7d79cc6d7d` | 191 → 191 | 177 → 177 |
| `0833534d07..807e2014c1` | 266 → 266 | 236 → 236 |
| `3366e0539e..0833534d07` | 229 → 230 | 164 → 165 |
| `e946ae8102..3366e0539e` | 89 → 89 | 89 → 89 |

비용(Risks): `af3f43ad3a..979c176905` 의 늘어난 대상은 `src/project_id.rs` 의 `"project_id"` 글자 때문이에요. 밑줄 있는 소스 파일 331개를 그 글자를 가진 테스트 대상 수(`tests/` 의 글자 검색)로 세면 0개 310, 1~5개 17, 6~10개 2, 10개 초과 2(`project_id` 67, `work_flow` 15)예요. `state_root` 는 7이에요. `src/project_id.rs` 를 바꾼 것은 NeXT 최근 first-parent 100 중 1이에요. 리더 결정 9874(위임 판단)는 승인대로 두는 것이에요. 비용이 실제로 문제가 되면 `docs/later.md` 에 한 줄을 남기고 다시 봐요.

통합 tip 합치기: `c62f942eea`(열한째 묶음)를 `8e7f4c765b827d96649f46deb43f9f22f0761243` 에 합쳤어요. 충돌은 `tools/repo_scans.json` 한 파일이고, tip 의 `test_targets` 형식에 `reads-src` 의 대상 한 줄과 규칙 경로 한 줄을 줄 임대로 다시 넣었어요. `src/check_affected.rs` 는 충돌 없이 합쳐졌고 tip 과의 차이는 이 SPEC 의 추가분뿐이에요.

SPEC verify(head `8e7f4c765b`, 적힌 그대로, 무거운 작업 칸): AC-1 1 passed, AC-2 1 passed, AC-3 22 passed, AC-4 1 passed, 모두 exit 0. `cargo fmt --all` 뒤 바뀐 파일 없음.

AC-5 재확인(합친 head 로 지은 실행 파일 sha256 `4b1caa49e6fd2834af0ab47ecb2431cba94bb8197aae88f55f3a6759b44d6a41`, 놓쳤던 두 범위만): 처음의 "전" 목록에서 빠진 대상 `[]`. `9c8b7fb763..3a6c46fc95` 는 `--list` 80 → 84, `--list --card` 80 → 84 이고 `update_end_to_end_cli` 가 있어요. `252a6e3be0..ab2d27e1fd` 는 269 → 276, 244 → 251 이고 `prop_reservation` 이 있어요. 위 표의 후 값보다 7개 많은 까닭은 확인하지 않았어요. 최근 착지 20건은 합친 head 로 다시 재지 않았어요.

guard-mutant 는 `b1e37d8aef` 에서 잰 값이고 합친 head 에서 다시 돌리지 않았어요.

아직 안 한 것(not_run): 전체 `cargo test`, Windows·Linux. READY 영수증(`check affected --card`)의 결과는 READY 노트에 적어요.
