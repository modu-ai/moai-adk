---
id: SPEC-20261123-rust-scans-before-ready
title: "The two scans every lane runs before READY are Rust tests, and check affected no longer plans a Python step"
status: draft
revision: 1
created: 2026-11-23
---

# READY 전에 도는 저장소 훑기 둘을 Rust 시험으로, check affected 에서 파이썬 단계를 걷기

## Intent

사용자 말(2026-11-23): "python tests are not used". 리더 노트(카드 C-317, 6949)가 순서를 정했어요: (1) 모든 레인이 READY 전에 돌리는 `repo_scans_test` 와 `source_manifest_test` 를 Rust 로 옮기고, (2) 그 보호가 Rust 로 옮겨진 같은 조각에서 `check affected` 가 파이썬 단계를 걷고, (3) 나머지 모듈과 `tools/*.py` 는 `SPEC-20261121-rust-only-text-scans` 와 뒤따르는 조각이 맡아요.

이 조각은 (1)과 (2)만 해요.

- `tests/repo_scans_test.py` 의 사례 6개를 `tests/repo_scans_table.rs` 로 옮기고 파이썬 파일을 지워요. 사례는 모두 `tools/repo_scans.json` 표와 저장소의 시험 파일을 읽을 뿐이라 도구가 필요 없어요.
- `tests/source_manifest_test.py` 의 "이 저장소 자체에 대한 사실"(매니페스트가 도구가 쓰는 글과 같음, 키 순서·중복 없음, 병합 드라이버 속성)을 `tests/source_manifest_facts.rs` 로 옮겨요. 도구의 동작을 임시 저장소에서 시험하는 사례는 `tools/source_manifest.py` 와 `tools/merge_driver_manifest.py` 가 `moai` 명령이 되는 조각(인벤토리 D1)에서 옮겨요.
- `check affected` 는 파이썬 모듈을 고르지 않고, `build_cli` 와 `python` 단계와 요약의 `py=` 를 없애요.

## Non-goals

- `tools/local_check.sh` 와 CI 의 `unittest` 단계, 파이썬 슬롯. 파이썬 시험이 하나라도 남아 있는 동안 CI 가 그 시험을 돌려야 해서, 마지막 파이썬 시험이 없어지는 조각이 걷어요(SPEC-20261121-rust-only-text-scans 의 D8).
- `tools/self_check.sh` 의 `py-repo-scans` 항목과 `py-source-manifest` 항목. 이 스크립트와 그 파이썬 시험은 워크플로·설치 스크립트 조각이 맡아요. `py-repo-scans` 는 파일이 없으면 건너뛰어요(`bin_python_item`).
- `tests/source_manifest_test.py` 의 삭제와 도구 동작 사례의 이식.
- 새 크레이트. 시험은 이미 있는 `serde_json`(preserve_order)과 `sha2` 만 써요.

## Requirements

- REQ-1: The file tests/repo_scans_table.rs shall hold a Rust test for each of the six cases of tests/repo_scans_test.py · failure: none — a case without a Rust test is found in the case table of the Result
- REQ-2: The file tests/source_manifest_facts.rs shall assert that the manifest of the skill package holds one entry per package file with its sha256, in sorted key order with no key twice, in the text the manifest tool writes, and that .gitattributes names the merge driver for it · failure: none — a broken fact is the failing test
- REQ-3: When check affected plans its steps, it shall plan no step that runs Python and select no Python module · failure: none — a Python module changed in a range is listed in neither the selection nor the steps
- REQ-4: When check affected gives its summary line, it shall name no `py=` count · failure: none — a reader of the old line is updated in the same slice
- REQ-5: The rules repo-scans-table and source-manifest-tool-output of tools/repo_scans.json shall name the Rust test files and a `cargo test` check for them · failure: none — a rule that names a deleted file is found by the table test
- REQ-6: The new Rust test files shall run without Python, without a shell command and without git · failure: none — a spawn is found by a scan of the two files
- REQ-7: This slice shall delete no Python file but tests/repo_scans_test.py · failure: none — another deletion is found by the check in AC-8

## Acceptance

- AC-1 (REQ-1): 표 시험 사례 6개와 면제 목록 점검 1개가 통과해요 · verify: `cargo test --offline --locked --test repo_scans_table`
- AC-2 (REQ-1): 사례를 깨뜨리는 변형이 잡혀요: 표에 없는 걷는 시험 파일, 표가 가리키는 없는 파일, 없는 시험 이름을 고르는 검사. Result 의 변형 표에 적어요 · verify: Result 절의 변형 표.
- AC-3 (REQ-2): 매니페스트 사실 시험이 통과해요 · verify: `cargo test --offline --locked --test source_manifest_facts`
- AC-4 (REQ-2): 매니페스트의 해시 한 줄을 틀리게 하거나 키 둘의 순서를 바꾸거나 한 키를 두 번 두면 해당 시험이 실패해요. Result 의 변형 표에 적어요 · verify: Result 절의 변형 표.
- AC-5 (REQ-3, REQ-4): 파이썬 시험 모듈이 바뀐 범위의 `--list` 에 파이썬 모듈도 `build_cli`·`python` 단계도 `py=` 도 없어요 · verify: `cargo test --offline --locked --test check_affected_cli`
- AC-6 (REQ-3, REQ-4): 선택·단계 단위 시험이 파이썬을 고르지 않아요 · verify: `cargo test --offline --locked --lib check_affected`
- AC-7 (REQ-5): 표 판정기가 새 규칙 이름으로 같은 판정을 해요 · verify: `cargo test --offline --locked --test repo_scans_cli`
- AC-8 (REQ-7): 이 조각이 지운 파이썬 파일은 `tests/repo_scans_test.py` 하나예요 · verify: `git diff --name-status $(git merge-base HEAD v2026.11.0)..HEAD` 의 `D` 줄이 그 파일 하나예요.

## Scope

- write: `.moai/specs/SPEC-20261123-rust-scans-before-ready/**`, `.moai/specs/SPEC-20261121-rust-only-text-scans/**`, `tests/repo_scans_table.rs`(새), `tests/source_manifest_facts.rs`(새), `tests/repo_scans_test.py`(삭제), `tools/repo_scans.json`, `src/check_affected.rs`, `src/ops/check.rs`(도움말 한 문장), `tests/check_affected_cli.rs`, `tests/check_affected_one_slot.rs`, `tests/repo_scans_cli.rs`, `tests/worktree_place_cli.rs`(한 줄), `docs/design-review/appendix/operations/check.md`
- read-only: `tests/source_manifest_test.py`, `tools/source_manifest.py`, `tools/merge_driver_manifest.py`, `tools/self_check.sh`, `skills/moai/**`, `.gitattributes`
- forbidden: 다른 파이썬 파일의 수정·삭제, 실제 HOME, 새 파이썬 시험

## Plan

1. RED: `tests/repo_scans_table.rs` 와 `tests/source_manifest_facts.rs` 를 쓰고, 표가 새 파일 이름을 가리키게 해요.
2. 변형으로 각 시험이 잡는지 확인해요(Result 의 변형 표).
3. `src/check_affected.rs` 에서 `TestFile.python`, `Selection.python`, 파이썬 고르기, `build_cli`·`python` 단계, `py=` 를 걷어내고 시험을 맞춰요.
4. `tests/repo_scans_test.py` 를 지워요.
5. 이웃 검사: 저장소 전체 단어 훑기 시험, `repo_scans_cli`, `check_affected_cli`, 문서 `check.md`.

## Decisions

결정은 되돌릴 수 있고 범주는 `none` 이에요. 결정자는 리더이고, 답이 없으면 기본값으로 진행해요.

- D10 `check affected` 가 파이썬 단계를 걷는 시점. 리더 노트 6949 는 "그 단계가 지키던 보호가 Rust 로 옮겨진 같은 조각"이라고 해요. 남은 파이썬 시험은 READY 영수증에서 빠지고 CI 의 `unittest` 단계가 계속 돌려요(D8). 기본값: 이 조각에서 걷어요.
- D11 `source_manifest_test.py` 의 저장소 사실 두 사례(`test_this_repository_manifest_is_what_the_tool_writes`, `test_this_repository_names_the_driver_for_the_manifest`)는 같은 사실을 Rust 가 단정하는 동안(D9) 파이썬에도 그대로 둬요. 기본값: 그대로 둠, 조각 4 에서 파일째 지워요.
- D12 `tools/self_check.sh` 의 `py-repo-scans` 항목은 파일이 없으면 건너뛰고(`skip (no tests/repo_scans_test.py)`), 같은 일을 하는 항목은 스크립트를 Rust 로 옮기는 조각에서 더해요. 그동안 READY 전 훑기는 표의 검사(`--test repo_scans_table`)가 맡아요. 기본값: 스크립트와 그 시험은 건드리지 않음.

## Risks

- 파이썬 정규식이 하던 훑기("걷는 시험 파일" 찾기와 `fn` 이름 찾기)를 Rust 의 문자열 함수로 다시 쓰는 것이라 두 구현이 어긋날 수 있어요. 옮긴 뒤 두 쪽이 낸 파일 집합을 같은 트리에서 비교했어요(Result).
- `skills/moai` 아래에 추적되지 않는 파일(예: `.DS_Store`)이 있으면 매니페스트 시험이 실패해요. 파이썬 도구도 디스크를 훑어서 같은 동작이에요.
- `check affected` 가 파이썬 단계를 걷으면 파이썬으로만 지켜지는 보호는 CI 의 `unittest` 단계에서만 돌아요. 그 시험이 모두 Rust 로 옮겨질 때까지의 한시적 상태예요.
- CLAUDE.md 의 "Scan tests first" 줄이 아직 `python3 -m unittest tests.repo_scans_test tests.source_manifest_test` 를 가리켜요. CLAUDE.md 는 리더의 파일이라 이 조각이 고치지 않고, 새 명령을 노트로 알려요.

## Result

구현 에이전트가 2026-11-23에 채웠어요. 감사가 다시 확인해요. 가지 `spec-impl/20261123-rust-only-repo-scans`, 기준 `178fde8beb`, 구현 커밋 `0d6b1aeb2d`, 이 worktree 의 `target/`, `CARGO_BUILD_JOBS=5`, `RUSTC_WRAPPER=sccache`, macOS arm64, cargo nextest. 같은 머신의 부하가 100 이상이었어요.

사례 대조표(`tests/repo_scans_test.py` 6개):

| 파이썬 사례 | Rust 시험(`tests/repo_scans_table.rs`) |
|---|---|
| `test_entries_are_well_formed` | `entries_are_well_formed` |
| `test_every_named_file_exists` | `every_named_file_exists` |
| `test_each_scanning_test_file_is_one_of_its_rule_files` | `each_scanning_test_file_is_one_of_its_rule_files` |
| `test_every_walking_test_file_is_listed_or_exempt` | `every_walking_test_file_is_listed_or_exempt` |
| `test_every_check_names_a_real_target_and_test` | `every_check_names_a_real_target_and_test` |
| `test_exemptions_are_not_stale` | `exemptions_are_not_stale` |
| (새) 면제 목록의 이유와 중복 | `every_exemption_has_a_reason_and_appears_once` |

`tests/source_manifest_test.py` 에서 저장소 자신을 읽는 사실만 옮겼어요(`tests/source_manifest_facts.rs`): 도구가 쓰는 글과 같음 `this_repository_manifest_is_what_the_tool_writes`, 병합 드라이버 속성 `this_repository_names_the_driver_for_the_manifest`, 키 순서 `files_are_in_sorted_key_order`, 키 중복 없음 `a_duplicate_key_is_never_read_as_last_wins`, 항목과 파일이 하나씩 짝 `every_package_file_has_one_entry_and_every_entry_a_file`, 경로의 링크 `the_manifest_path_has_no_link`, 충돌 표시 없음 `the_manifest_and_the_package_hold_no_conflict_marker`. 임시 저장소에서 `tools/source_manifest.py` 와 병합 드라이버를 돌리는 나머지 사례는 파이썬 모듈에 그대로 남아요(도구가 `moai` 명령이 되는 조각, 인벤토리 D1).

- AC-1: met — `repo_scans_table` 7 passed
- AC-2: met — 변형 표(아래) 중 표 시험 둘이 의도한 시험에서 실패
- AC-3: met — `source_manifest_facts` 7 passed
- AC-4: met — 변형 표(아래) 셋이 의도한 시험에서 실패
- AC-5: met — `check_affected_cli` 16 passed, `check_affected_one_slot` 7 passed, `check_pipeline_config` 3 passed
- AC-6: met — `--lib check_affected` 14 passed, `--bin moai-cli` 28 passed
- AC-7: met — `repo_scans_cli` 19 passed
- AC-8: met — 이 조각이 지운 파일은 `tests/repo_scans_test.py` 하나

변형 5개와 하나 더(커밋한 트리에서 돌리고 `git checkout --` 로 되돌림), 모두 의도한 시험에서 KILLED:

| 변형 | 실패한 시험 |
|---|---|
| 매니페스트의 sha256 한 글자를 바꿈 | `this_repository_manifest_is_what_the_tool_writes` |
| 매니페스트의 키 하나를 앞 키와 같게 바꿈(키 두 번) | `a_duplicate_key_is_never_read_as_last_wins` (+ 짝 시험 둘) |
| 매니페스트의 키 하나를 사전순 밖으로 바꿈 | `files_are_in_sorted_key_order` (+ 짝 시험 둘) |
| 표의 검사 이름 한 글자를 틀리게 함(없는 시험 이름) | `every_check_names_a_real_target_and_test` |
| 걷는 시험 파일의 면제 행 하나를 지움(항목도 면제도 없음) | `every_walking_test_file_is_listed_or_exempt` |
| 표의 `unittest` 검사의 `tests.<이름>` 을 다시 대상으로 읽음 | `no_python_module_is_selected_or_read_as_a_test`, `the_misses_of_the_integration_day_are_selected` |

첫 시도의 마지막 변형(`--test` 뒤 낱말만 읽는 줄에 `tests.` 시작 조건을 더함)은 살아남았어요. 변형이 약했어요: `tests.<이름>` 뒤에는 다음 낱말이 없어서 아무것도 넣지 않았어요. 옛 동작을 제대로 되살리는 변형(`tests.` 뒤 이름을 대상으로 넣기)으로 바꿔서 위 표의 결과를 얻었어요.

같은 트리에서 옛 파이썬 휴리스틱과 Rust 휴리스틱을 비교했어요: 걷는 시험 파일 73개(파이썬 쪽 값), 면제 22개가 두 목록에서 같고, 목록에 없는 파일과 낡은 면제는 양쪽 모두 0개예요.

실행: `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, `cargo fmt --check` exit 0, `rustfmt --edition 2021 --check src/ops/check.rs` exit 0, `git diff --check` exit 0, `moai spec check` 두 SPEC 오류 0·경고 0, `moai trailer-check 178fde8beb..HEAD` failing 0.

돌리지 않은 것: 전체 `cargo test`(READY 영수증이 고른 대상으로 돌려요), CI, Linux·Windows. 남은 일(이 조각 밖): `CLAUDE.md` 의 "Scan tests first" 줄과 규칙 출처 표의 같은 줄이 아직 두 파이썬 모듈을 가리켜요. `tools/self_check.sh` 의 `py-repo-scans` 항목은 파일이 없어서 건너뛰고(`skip (no tests/repo_scans_test.py)`), `py-source-manifest` 항목은 파이썬 모듈이 남아 있는 동안 그대로예요. `skills/moai/references/internal/self-check.md` 가 "the Python tests of the area" 와 `unittest` 줄을 아직 말해요(스킬 글이라 리더의 SPEC 이 맡아요). `docs/ci-rc-ci-spec.md` 의 두 `py-*` 행도 같아요.
