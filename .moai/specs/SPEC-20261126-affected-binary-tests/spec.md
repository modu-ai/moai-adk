---
id: SPEC-20261126-affected-binary-tests
title: "Select the integration tests that drive the built binary in check affected"
status: draft
revision: 4
created: 2026-11-26
---

# Select the integration tests that drive the built binary in check affected

## Intent

`moai check affected` 는 지은 실행 파일(binary)을 돌려서 시험하는 통합 테스트를 소스 변경에 이어 주지 못해요. 2026-11-26 하루에 세 번 빠졌고, 레인 영수증은 PASS 였어요. `SPEC-20261126-affected-card-readers`(head `b1e37d8aef`, 아직 착지 전)가 모듈을 읽는 테스트를 더하지만, 그 head 에서 빌드한 실행 파일로 다시 골라도 아래가 남아요. 이 SPEC 은 남은 빈틈만 막고 지금 고르는 것은 하나도 빼지 않아요.

잰 값(2026-11-26, 설치본 `moai 4.2648.53-dev` 와 `b1e37d8aef` 빌드, `--list` 와 `--list --card`, 테스트는 안 돌림):

| 범위 | 바뀐 소스 | `b1e37d8aef` 넓은/카드 | 그래도 빠지는 것 |
| --- | --- | --- | --- |
| `5a6aa857d3..12510df496` | `src/launcher_init.rs`, `src/ops/project_apply.rs` | 112 / 3 | `go_project_real_run_cli`, `go_settings_cli`, `update_end_to_end_cli`, `go_state_preview_cli` |
| `9c8b7fb763..3a6c46fc95` | `src/state_root.rs` | 83 / 83(전체 세트) | `go_project_real_run_cli`, `go_state_preview_cli`, `project_update_hooks_cli` |
| `91dd561a5a..6bb4f8a031` | `src/trailers.rs` | 59 / 2 | `trailer_check_cli` |

네 번째 사례(revision 2, 범위 `98ac09db55..32bd7c9f30`, head 에서 다시 고름): 바뀐 파일에 `tools/repo_scans.json` 과 `src/repo_scans.rs` 가 있어요. 설치본의 넓은 선택과 카드 선택(전체 세트), `b1e37d8aef` 빌드의 두 선택 모두 205개이고 `gate_registry_tier` 가 없어요. head 의 `tests/` 에서 `tools/repo_scans.json` 글자를 가진 Rust 테스트 파일 21개 중 설치본 넓은 선택에 없는 것은 7개예요: `card_land_audited_merge_cli`, `card_land_batch_cli`, `card_land_scans_cli`, `card_land_sync_cli`, `gate_registry_tier`, `statusline_card_flow_icons`, `update_tux_golden`. 바뀐 경로 글자 규칙이 `skills/` 와 `docs/` 만 보기 때문이에요.

까닭: 앞의 세 범위에서 빠진 테스트들은 모듈 경로가 아니라 명령 이름(`"project-apply"`, `"trailer-check"`)으로 실행 파일을 돌려요. 일부는 그 글자가 `#[path]` 로 가져오는 도움 파일 `tests/common/go_project_real_run.rs` 에만 있어요. `3a6c46fc95` 에서 `state_root::` 글자를 가진 `src` 파일은 65개인데 그중 `src/ops` 파일은 `leader_run.rs`, `spec.rs` 뿐이라, 명령 파일에서 한 단계만 따라가서는 `go_project_real_run_cli` 에 닿지 않아요.

다시 나타난 사례(revision 3, 리더 노트 10103):

- `trailer_check_cli`: 열두째 묶음의 넓은 검사(`c62f942eea..b3048a821a`, 247개, PASS)도 `src/trailers.rs` 가 바뀌었는데 이 모듈 9개를 돌리지 않았어요. 위 표의 세 번째 범위와 같은 까닭이고 REQ-4 가 덮는 모양이에요.
- 라이브러리 표 검사: 범위 `e946ae8102..4db9be4f90`(`src/update.rs`, `src/update_tux.rs` 변경)의 카드 선택은 라이브러리 테스트를 `update::`, `update_tux::` 로만 좁혀서 `home_layout::tests::every_rust_written_home_name_is_in_the_table` 이 빠졌어요(설치본 `--list --card` 로 다시 읽음, 전체와 레인 선택은 라이브러리 테스트를 모두 돌려요). 열셋째 묶음의 넓은 검사에서 FAIL 로 잡혔어요. 까닭은 카드 선택이 규칙 파일이 바뀐 스캔만 고르기 때문이에요(`SPEC-20261124-card-select-batch-land` D-6). 이 검사의 규칙 파일은 `src/home_layout.rs` 이고 범위(scope)는 `src/` 예요.

`SPEC-20261126-affected-card-readers` 의 지금 head `7fd6bfbea4` 를 읽으면 그 변경은 통합 테스트 대상만 더해서 위 라이브러리 표 검사는 다루지 않아요(코드 읽기, 그 head 의 실행 파일로 고르지는 않았어요).

호출 위쪽 명령의 테스트(revision 4, 감사 노트 10131, 범위 `91dd561a5a..1370d8a9d8`, `src/card/work.rs` 와 `src/identity.rs` 변경): 설치본과 `b1e37d8aef` 빌드의 카드 선택(각 10개)에 `card_wip_limit`, `card_submit_cli`, `orchestrate_cli`, `leader_brief_cli` 가 모두 없고, 설치본 전체 선택(115개)에도 `orchestrate_cli` 와 `leader_brief_cli` 가 없어요(`--list` 만, 테스트는 안 돌렸어요. 이 네 테스트가 실패했다는 기록은 못 봤어요). 네 테스트는 바뀐 함수 `card::work::next` 를 한두 단계 위에서 부르는 명령(`card next`, `card submit`, `card orchestrate plan`, `leader brief`)을 돌려요. 위의 어느 규칙도 닿지 않아서, 테스트가 기대는 파일을 스스로 적게 해요.

읽는 사람: 레인, 이 변경의 감사자, 착지 묶음을 도는 리더.

## Interview

- Q1: 바뀐 함수를 위에서 부르는 명령의 테스트가 빠지는 세 번째 사례를 어떻게 다룰까요? → A: 테스트 파일에 `// check-affected: when-changed <경로>` 한 줄을 두는 요구를 더해요(사용자가 질문 창에서 직접 고른 답, 리더가 노트 10137 에 옮겨 적었어요).

## Non-goals

- 호출 관계(call graph)를 여러 단계 따라가지 않아요.
- REQ-10 의 선언은 파일 경로 하나를 글자 그대로 읽어요. 폴더나 와일드카드 선언은 없고, 선언을 자동으로 만들지 않아요. 도우미 파일(`tests/common/`, `#[path]`)이나 영역의 `main.rs` 에 적은 선언은 읽지 않아요.
- `src/main.rs` 안의 명령 분기(`cc`, `glm` 등)는 명령 이름으로 잇지 않아요.
- 넓은 경로, 상한, 영역 규칙, `SPEC-20261126-affected-card-readers` 의 두 규칙은 바꾸지 않아요.
- 스캔 규칙은 REQ-9 의 라이브러리 검사만 더해요. 카드 세트의 통합 테스트(`--test`) 대상은 지금처럼 규칙 파일이 바뀐 스캔에서만 골라요.
- READY 관문과 요약 줄의 낱말 종류는 바꾸지 않아요.

## Requirements

<!-- 이 SPEC 은 선택에 대상을 더하기만 해요. 약해지는 검사는 없어요. -->
- REQ-1: When a range changes a path under `src/` that ends in `.rs`, the `check affected` command shall select, in the card set, the full set and the lane set, every test target whose own source holds the whole line `// check-affected: every-source-change`. · failure: none — 읽지 못한 테스트 파일은 지금과 같게 다뤄요. 표시가 사라지는 경우는 AC-4 의 저장소 테스트가 잡아요.
- REQ-2: The repository shall carry the line of REQ-1 in each of the test files `go_project_real_run_cli`, `go_settings_cli`, `go_state_preview_cli`, `project_update_hooks_cli` and `update_end_to_end_cli`. · failure: none — 하나라도 빠지면 AC-4 의 테스트가 빨개져요.
- REQ-3: When a range changes `src/ops/<name>.rs` (`mod.rs` aside), the `check affected` command shall select, in all three sets, every test target whose own source, or a helper file it loads with `#[path]`, holds the whole string literal of the command name (`<name>` with each `_` written as `-`).
- REQ-4: When a range changes a library source file with a module path, the `check affected` command shall apply the rule of REQ-3 to each `src/ops/<other>.rs` that names that module path and whose `<other>` equals the changed file's stem or starts with that stem, its trailing `s` removed, followed by `_`.
- REQ-5: The `check affected` command shall keep in every selection each test target that it selected for the same range before this change. · failure: none — 세 규칙은 합집합으로만 더해요.
- REQ-6 (REQ-3, REQ-4): IF a helper file named by `#[path]` or a file under `src/ops` that these rules read cannot be read, THEN the `check affected` command shall stop the selection with an error that names that file.
- REQ-7 (REQ-3): The repository shall name each `.rs` file under `src/ops`, `mod.rs` aside, after a command that `src/help.rs` lists (`_` written as `-`). · failure: none — 어긋난 파일이 생기면 AC-8 의 테스트가 빨개져요.
- REQ-8: When a range changes a path under `tools/`, the `check affected` command shall select, in the card set, the full set and the lane set, every test target whose own source holds that path as text. · failure: none — 글자 검색이고 읽지 못한 테스트 파일은 지금과 같게 다뤄요.
- REQ-9: When a range changes a path that a scan of `tools/repo_scans.json` holds in its scope, and a check of that scan runs library unit tests (`--lib`), the `check affected` command shall select, in the card set, the full set and the lane set, the library unit tests that the check names (every library unit test when it names none). · failure: none — 전체와 레인 세트는 지금도 라이브러리 테스트를 모두 돌려요. 읽지 못한 스캔 표는 지금처럼 전체 세트가 돼요.
- REQ-10: When a range changes a path that a test file names in its own source in a whole line `// check-affected: when-changed <path>`, the `check affected` command shall select that test file in the card set, the full set and the lane set, where a test file is a target `tests/<name>.rs` or a module `tests/<area>/<module>.rs` of an area test binary, and a module is selected as that module alone. · failure: none — 선언이 없는 테스트는 지금과 같게 골라요. 한 파일에 선언 줄이 여럿이면 하나만 맞아도 골라요.
- REQ-11: The repository shall carry the line of REQ-10 with the path `src/card/work.rs` in each of the test files `card_wip_limit`, `card_submit_cli`, `orchestrate_cli` and `leader_brief_cli`. · failure: none — 하나라도 빠지면 AC-11 의 저장소 테스트가 빨개져요.
- REQ-12 (REQ-10): The repository shall name an existing file in every line of REQ-10 that a Rust file under `tests/` carries. · failure: none — 이름이 바뀌었거나 잘못 적은 경로는 AC-12 의 테스트가 빨개져요.

## Acceptance

- AC-1 (REQ-1): 임시 저장소에서 `src/` 의 `.rs` 하나가 바뀐 범위의 카드, 전체, 레인 선택에 표시 줄을 가진 테스트가 들어 있고, 표시가 더 긴 줄의 일부인 테스트와 `docs/` 만 바뀐 범위에서는 더해지지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(a_source_change_selects_the_marked_binary_tests)'`
- AC-2 (REQ-3, REQ-6): `src/ops/project_apply.rs` 만 바뀐 범위에서, 자기 소스에 `"project-apply"` 를 가진 테스트와 `#[path]` 도움 파일에만 그 글자를 가진 테스트가 세 선택에 들어 있고, `"project-applied"` 만 가진 테스트는 더해지지 않아요. 도움 파일이 없거나 `src/ops` 의 파일을 읽지 못하면 오류에 그 파일 이름이 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(a_command_file_change_selects_the_tests_that_name_the_command)'`
- AC-3 (REQ-4): `src/trailers.rs` 가 바뀐 모양에서 `src/ops/trailers.rs`, `src/ops/trailer_check.rs` 의 명령 이름을 가진 테스트가 들어 있고, 모듈을 부르지만 이름이 다른 명령 파일의 테스트는 이 규칙으로 더해지지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(a_library_change_selects_the_tests_of_its_own_command_files)'`
- AC-4 (REQ-1, REQ-2, REQ-3, REQ-4): 저장소의 지금 파일로 고르면 `src/launcher_init.rs` 와 `src/ops/project_apply.rs` 의 선택, `src/state_root.rs` 의 선택에 `go_project_real_run_cli`, `go_settings_cli`, `update_end_to_end_cli` 가, `src/trailers.rs` 의 카드 선택과 전체 선택에 `trailer_check_cli` 가 들어 있고, REQ-2 의 다섯 대상이 모두 `src/` 변경에 들어 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(the_missed_binary_tests_of_this_repository_are_selected)'`
- AC-5 (REQ-5): 기존 선택 테스트가 기대값을 바꾸지 않고 통과해요 · verify: `cargo nextest run --locked --offline --test check_affected_card_select --test check_affected_card_areas --test check_affected_selection`
- AC-6 (REQ-5): READY 의 Result 에, 위 표의 세 범위와 `c62f942eea..b3048a821a`, `e946ae8102..4db9be4f90`, `91dd561a5a..1370d8a9d8` 를 바꾸기 전 실행 파일과 후 실행 파일로 `--list` 와 `--list --card` 를 돌린 대상 목록과 라이브러리 테스트 필터의 비교(빠진 것 0, 늘어난 수)를 적고, `c62f942eea..b3048a821a` 의 후 목록에 `trailer_check_cli` 가, `e946ae8102..4db9be4f90` 의 후 카드 목록에 `every_rust_written_home_name_is_in_the_table` 이, `91dd561a5a..1370d8a9d8` 의 후 목록에 REQ-11 의 네 테스트가 있어요. 그 범위의 head 에는 선언 줄이 없으므로 마지막 범위는 head 의 파일에 선언 줄을 임시로 더한 임시 저장소에서 재고 그 방법을 함께 적어요 · verify: `moai check affected <base>..<head> --list` 와 `moai check affected --card <base>..<head> --list`
- AC-7 (REQ-1, REQ-3, REQ-4, REQ-8, REQ-9, REQ-10): `moai check affected` 가 READY 의 head 에서 PASS 이고, 여섯 규칙의 코드를 하나씩 지운 변조에서 AC-1 에서 AC-3, AC-9, AC-10, AC-11 의 테스트가 실패해요(READY 의 guard-mutant 줄) · verify: `moai check affected --card <base>..<head>`
- AC-8 (REQ-7): 저장소의 `src/ops` 파일 이름이 모두 `src/help.rs` 의 명령 목록에 있어요(2026-11-26 `8f472c5811` 에서 174개가 모두 맞고 예외 0개) · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(every_command_file_is_named_like_a_listed_command)'`
- AC-9 (REQ-8): 임시 저장소에서 `tools/` 아래 파일 하나가 바뀐 범위의 카드, 전체, 레인 선택에 그 경로 글자를 가진 테스트가 들어 있고, 파일 이름만 가진 테스트는 더해지지 않아요. 저장소의 지금 파일로 고르면 `tools/repo_scans.json` 의 선택에 `gate_registry_tier` 가 들어 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(a_tools_change_selects_the_tests_that_name_the_path)'`
- AC-10 (REQ-9, REQ-5): 임시 저장소에서 `src/` 의 `.rs` 하나가 바뀐 범위의 카드 선택에, 범위(scope)가 `src/` 이고 규칙 파일은 바뀌지 않은 스캔의 라이브러리 테스트 이름(`--lib <이름>` 꼴과 `--lib -- <이름> <이름>` 꼴)이 들어 있고, 범위가 `docs/` 인 스캔의 이름은 없고, 그 스캔의 통합 테스트 대상은 더해지지 않고, 전체와 레인 선택은 라이브러리 테스트를 모두 돌려요. 저장소의 지금 파일로 고르면 `src/update.rs` 의 카드 선택에 `every_rust_written_home_name_is_in_the_table` 이 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(a_source_change_in_a_scan_scope_selects_its_library_tests)'`
- AC-11 (REQ-10, REQ-11, REQ-5): 임시 저장소에서 선언 줄에 적힌 파일이 바뀐 범위의 카드, 전체, 레인 선택에 그 테스트가 들어 있고(선언이 두 줄이면 어느 파일이 바뀌어도), 적히지 않은 파일의 변경에서는 이 규칙으로 더해지지 않고, 선언이 더 긴 줄의 일부이거나 `//!` 줄이거나 경로의 앞부분만 적은 줄은 더하지 않아요. 영역 테스트 실행 파일의 모듈에 적은 선언은 그 모듈을 고르고 테스트 필터도 그 모듈로 좁혀져 같은 실행 파일의 다른 모듈은 이 규칙으로 더해지지 않아요. 저장소의 지금 파일로 고르면 `src/card/work.rs` 의 세 선택에 REQ-11 의 네 테스트가 들어 있어요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(a_declared_path_change_selects_the_test_that_names_it)'`
- AC-12 (REQ-12): `tests/` 아래 Rust 파일의 모든 선언 줄이 저장소에 있는 파일을 가리키고, 없는 파일을 적은 선언은 이 검사가 이름으로 잡아요 · verify: `cargo nextest run --locked --offline --test check_affected_binary_tests -E 'test(every_declared_path_names_a_file)'`

## Scope

- write: `.moai/specs/SPEC-20261126-affected-binary-tests/**`, `src/check_affected.rs`, `tests/check_affected_binary_tests.rs`
- write (표시 한 줄씩): `tests/go_project_real_run_cli.rs`, `tests/go_settings_cli.rs`, `tests/go_state_preview_cli.rs`, `tests/project_update_hooks_cli.rs`, `tests/update_end_to_end_cli.rs`
- write (선언 한 줄씩, revision 4): `tests/card/card_wip_limit.rs`, `tests/card/card_submit_cli.rs`, `tests/card/orchestrate_cli.rs`, `tests/leader_brief_cli.rs`
- write (줄 임대): `tools/repo_scans.json`(저장소 파일을 읽는 새 테스트의 등록 줄)
- read-only: `tests/check_affected_selection.rs`, `tests/check_affected_card_select.rs`, `tests/check_affected_card_areas.rs`, `tests/areas.toml`, `src/ops/**`, `src/help.rs`
- forbidden: `skills/**`, `src/gate.rs`

## Plan

1. `SPEC-20261126-affected-card-readers` 가 착지한 뒤 그 tip 에서 시작해요. 그 전에는 새 테스트 파일(빨간 테스트)만 쓰고 `src/` 와 기존 테스트는 쓰지 않아요.
2. 빨간 테스트를 먼저 써요(AC-1 에서 AC-4, AC-9 에서 AC-12). AC-11, AC-12 는 개정 4 승인 뒤에 써요.
3. `src/check_affected.rs` 에 표시 줄 규칙, 명령 파일 규칙, 라이브러리에서 같은 낱말의 명령 파일로 한 단계 가는 규칙, `tools/` 경로 규칙, 라이브러리 스캔 검사 규칙, 선언한 경로 규칙을 합집합으로 더해요.
4. 다섯 테스트에 표시 줄을, 네 테스트에 선언 줄을 넣고, 새 테스트를 줄 임대로 스캔 표에 등록해요.
5. AC-6 의 전후 비교를 재요.

## Risks

- 영수증이 커져요. 구현 전에 손으로 센 예상(확인 안 됨): 카드 선택이 `12510df496` 범위에서 3개에서 약 33개, `6bb4f8a031` 범위에서 2개에서 약 17개, 넓은 선택은 각각 112에서 약 131, 59에서 약 74, `3a6c46fc95` 범위는 83에서 약 86. 카드 세트의 "20개 안팎" 안내를 넘어요. AC-6 에서 재요.
- `tools/repo_scans.json` 은 테스트를 더할 때마다 바뀌어서 REQ-8 의 대상도 자주 늘어요. 위 네 번째 범위에서 더해지는 대상은 7개예요.
- `src/` 를 바꾸는 모든 범위가 다섯 대상을 더 돌려요. 다른 워크트리의 영수증 로그 하나(`check-affected-60cfb37da8f1`)에서 그중 네 대상의 테스트 시간 합은 62.2초였고, `update_end_to_end_cli` 의 시간과 링크 시간은 안 쟀어요.
- 표시가 없는 실행 파일 테스트가 깊은 라이브러리 호출에만 기대면 여전히 빠져요. 그런 일이 보이면 표시 줄 하나를 더해요. 마지막 보호는 리더가 착지 다섯 번마다 도는 전체 `cargo test` 예요.
- REQ-7 은 파일 이름이 명령 목록에 있는지만 봐요. 한 파일이 다른 이름의 명령까지 처리하면 그 명령의 테스트는 REQ-3 으로 더해지지 않아요.
- 몇몇 테스트는 문자열 안에 `#[path]` 글자를 가져요(`tests/ac_command_lint.rs` 의 raw 문자열). 불러오기가 아니므로 REQ-3, REQ-6 은 실제 속성만 읽어야 해요. AC-4 가 저장소 파일로 이를 확인해요.
- AC-4 의 테스트는 저장소 파일을 읽어서, 이름이 바뀌면 빨개져요. 조용히 빠지는 것보다 낫다고 봤어요.
- REQ-9 는 `src/` 를 바꾸는 카드 영수증이 라이브러리 스캔 테스트를 함께 돌리게 해서, 묶음의 넓은 검사와 같은 테스트가 한 번 더 돌아요. 열두째 묶음 로그에서 지금 `src/` 범위인 라이브러리 스캔 테스트 10개의 시간은 하나에 0.11초에서 0.73초, 합쳐 약 4.2초였어요. 라이브러리 테스트 실행 파일은 소스 변경의 카드 선택이 이미 지어요.
- REQ-10 의 선언은 손으로 적어요. 선언하지 않은 테스트는 여전히 빠지고, 마지막 보호는 전체 `cargo test` 예요. `src/card/work.rs` 를 바꾸는 범위는 테스트 파일 넷(`card` 실행 파일의 모듈 셋과 `leader_brief_cli`)을 더 돌려요(실행 시간은 안 쟀어요).
- 여러 줄 문자열 안에 선언 줄과 똑같은 줄이 있으면 선언으로 읽혀요. 대상이 늘어나는 쪽이고, AC-12 가 없는 파일을 적은 줄은 잡아요. AC-11 의 테스트는 그런 줄이 자기 소스에 생기지 않게 문자열을 이어 붙여요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 모든 `src/` 변경에서 도는 실행 파일 테스트를 테스트 파일의 표시 줄로 정해요. 여러 단계 호출 관계를 따라가는 길은 고르지 않았어요 | `state_root` 는 `src` 파일 65개가 쓰고 그중에 `launcher_init`, `project_update` 같은 공용 모듈이 있어, 끝까지 따라가면 실행 파일을 쓰는 테스트 대부분이 걸릴 것으로 봐요(추정이에요. 폐포는 계산하지 않았어요. `3a6c46fc95` 에서 실행 파일이나 실행 도움 이름을 가진 `tests/` 파일은 316개). 카드 기준 1은 `state_root` 범위에서도 `go_project_real_run_cli` 를 요구해요 | 표시 줄과 REQ-1 의 코드를 지워요 |
| D-2 | 라이브러리 변경에서 명령 파일로 가는 길은 이름의 낱말이 같은 파일만 따라가요. 모듈을 부르는 모든 명령 파일을 따라가는 길은 고르지 않았어요 | 모두 따라가면 `state_root` 범위에서 `"spec"`, `"leader-run"` 글자를 가진 대상이 약 29개 더 늘어요(손으로 센 값). 지금 확인된 사례는 좁은 길로 모두 닿아요 | 이름 조건을 지워 넓혀요 |
| D-3 | 세 규칙 모두 카드 세트에도 상한 없이 써요 | 놓친 두 번은 카드 영수증이었어요. 리더가 앞선 SPEC 에서 상한 없이 두기로 했어요 | 카드 세트에서 REQ-3, REQ-4 를 빼요(대상이 줄어드는 쪽이라 승인이 필요해요) |
| D-4 | 표시 줄은 실행 도움 파일을 가져오는 다섯 대상에 모두 넣어요 | 카드 기준은 셋을 이름으로 말하고 나머지 둘도 같은 도움 파일로 같은 명령을 돌려요 | 두 파일의 표시 줄을 지워요 |
| D-5 | `tools/` 파일도 `skills/`, `docs/` 처럼 경로 글자를 가진 테스트를 고르고, 같은 카드에서 고쳐요(revision 2) | 리더 결정(2026-11-26, 사용자 위임에 따른 판단): 바뀐 파일을 쓰는 테스트를 선택이 모른다는 같은 원인이에요 | REQ-8 의 코드를 지워요 |
| D-6 | 라이브러리 단위 테스트로 도는 스캔 검사만, 규칙 파일이 안 바뀌어도 범위 조건으로 카드 세트에 넣어요(revision 3). 통합 테스트 대상은 `SPEC-20261124-card-select-batch-land` D-6(규칙 파일이 바뀐 스캔만)을 그대로 둬요 | 리더 요청(노트 10103). 그 D-6 의 까닭은 범위만으로 고르면 통합 대상이 63개, 모두 89개가 돼 "20개 안팎" 기준을 넘는다는 실측이에요. 라이브러리 필터는 대상 수를 늘리지 않아요 | REQ-9 의 코드를 지워요 |
| D-7 | 호출 위쪽 명령의 테스트는 테스트 파일이 기대는 파일을 적는 선언 줄로 골라요(revision 4). 호출 파일을 따라가는 규칙은 고르지 않았어요 | 사용자 선택(노트 10137). 따라가는 규칙은 두 단계가 필요하고 늘어나는 대상 수를 재지 않았어요 | 선언 줄 넷과 REQ-10 의 코드를 지워요 |

## Result

구현 전(2026-11-26, 브랜치 `spec/affected-binary-tests-20261126`).
