---
id: SPEC-20261123-land-scan-more-readers
title: "the reader scan of the repo-scans table also finds include_bytes! readers, tests that only spell a repository path and tests that run a checked-in script"
status: draft
revision: 1
created: 2026-11-24
---

# the reader scan of the repo-scans table also finds include_bytes! readers, tests that only spell a repository path and tests that run a checked-in script

## Intent

`SPEC-20261123-land-file-read-scans`(C-398)는 표 테스트(`tests/repo_scans_table.rs`)가 실행 때 저장소 파일을 여는 통합 테스트를 찾아 `always` 스캔이 덮는지 검사하게 했어요. 그 찾기는 두 가지 모양만 봐요: `include_str!("../..")` 와, `CARGO_MANIFEST_DIR` 를 쓰면서 읽는 호출(`read_to_string`, `fs::read`, `read_dir`, `File::open`)이 있고 저장소 경로를 글자로 적은 파일이에요. C-398 감사(노트 7531, N3)가 이 찾기가 놓치는 모양 세 가지를 적었고, 같은 감사의 기록 C-311 감사 N4(노트 7553)는 이름 붙은 패키지 파일을 복사해 읽는 `tests/e2e_temp_home_cli.rs` 가 표에서 빠져 있다고 적었어요.

1. `include_bytes!`: `tests/update_switch_layer.rs:14` 는 `.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1` 을, `tests/update_switch_cli.rs:22-23` 은 같은 파일과 `install/moai-switch.sh` 를 컴파일 때 바이트로 가져와요. `include_str_paths` 가 `include_str!(` 만 찾아서 둘 다 보이지 않고, `install/` 는 어느 스캔의 범위에도 없어요.
2. 저장소의 스크립트를 실행하는 테스트: `tests/statusline_golden.rs:24` 는 `python3 -I tools/statusline_golden.py` 를 `current_dir(CARGO_MANIFEST_DIR)` 에서 돌리는데 읽는 호출이 없어 찾기에 걸리지 않아요. `tests/atlas_three_arm.rs:14` 는 `CARGO_MANIFEST_DIR/tools` 폴더를 파이썬에 넘기고 `runpy` 로 `atlas_three_arm` 모듈을 돌려요. 글자 `tools` 는 한 조각 폴더라서 C-398 D-3 이 세지 않아요.
3. 읽기 호출 없이 저장소 경로를 바이너리에 넘기는 테스트: `tests/ui_summary_terminal_cli.rs:20` 은 `concat!(env!("CARGO_MANIFEST_DIR"), "/skills/moai")` 를 `moai` 에 인자로 넘겨요(쌍둥이 `ui_korean_terminal_cli.rs` 는 읽는 호출이 있어서 덮여요).

측정(팁 `7f3bde17c1`, 2026-11-24, 이 세션에서 직접 센 값): `tests/*.rs` 중 `CARGO_MANIFEST_DIR` 를 쓰는 파일에서 읽는 호출이 하나도 없는 파일은 `tests/statusline_golden.rs` 와 `tests/ui_summary_terminal_cli.rs` 둘뿐이에요(`grep -L`). `include_bytes!` 는 `update_switch_layer.rs`·`update_switch_cli.rs` 둘이고(`board_bundle_link_cli.rs:248` 의 글자는 검사 대상 이름이지 호출이 아니에요), 인터프리터(`"python3"`, `"bash"`, `"sh"`)를 띄우는 파일은 18개이며 그중 글자 `tools` 만 폴더로 적은 파일은 `atlas_three_arm.rs` 예요(`check_pipeline_config.rs:59` 와 `claude_python_adapter_removed.rs:192` 의 `tools` 는 임시 프로젝트나 없음 확인용 이름).

`tests/e2e_temp_home_cli.rs`(노트 7553): C-398 마지막 수리가 이 테스트를 `reads-skills`, `reads-root-files`, `reads-src` 의 규칙 파일로 올렸어요. 이름 붙은 패키지 파일(`references/home/routing.md` 같은 것)은 `skills/moai` 의 상대 이름이라 글자만으로는 저장소 경로가 아니고, 테스트가 `skills/moai` 폴더를 통째로 가리켜서 그 폴더 범위가 덮어요. 이 SPEC 은 그 사실을 표 테스트가 지키게 하는 한 검사만 더해요.

이 SPEC 은 C-398(`SPEC-20261123-land-file-read-scans`) 이 착지한 뒤의 `tests/repo_scans_table.rs` 와 `tools/repo_scans.json` 위에 올라요. 그 전에는 두 파일이 C-398 의 예약이라 손대지 않고 초안만 썼어요.

### 읽은 것 (2026-11-24, 팁 `7f3bde17c1`, C-398 head `9384f05ae8`)

- 카드 C-405 본문·기준 2개·노트 7553, 카드 C-398 의 노트 7531·7576·7623·7635(감사 비차단 N1~N6), `SPEC-20261123-land-file-read-scans` revision 3 전체.
- C-398 head 의 `tests/repo_scans_table.rs`: `include_str_paths`, `Tracked::repo_path`, `string_literals`, `file_read_paths`, `file_readers`, `uncovered_readers`, `FILE_READ_CALLS`, `FILE_READ_EXEMPT`.
- 위 다섯 테스트 파일의 해당 줄.

## Non-goals

- 경로를 코드에서 조립하는 읽기(글자로 적지 않은 경로)와 `src/` 안 단위 테스트의 읽기: C-398 D-2 와 같은 한계예요.
- 인터프리터가 실행하는 스크립트 안의 `import` 를 따라가는 일: 폴더 `tools` 전체를 읽는 것으로 세요.
- `moai repo-scans` 판정(`src/repo_scans.rs`)과 `card land`, `check affected` 의 동작을 바꾸는 일.
- 게이트를 느슨하게 하는 일, 전체 `cargo test` 를 강제하는 일, Python 테스트를 손으로 돌리는 일.

## Requirements

- REQ-1: When a top-level test file spells `include_bytes!("<literal>")` with a literal that leaves `tests/`, the reader scan shall list the repository path the literal names · failure: none — a literal that stays inside `tests/` is no repository read.
- REQ-2: When a top-level test file names `CARGO_MANIFEST_DIR`, the reader scan shall list the repository paths its string literals spell whether or not the file calls a read function · failure: none — a file that names no repository path is no reader.
- REQ-3: When a top-level test file names `CARGO_MANIFEST_DIR` and spawns an interpreter named by the literal `"python3"`, `"bash"` or `"sh"`, the reader scan shall count the literal `tools` as the repository folder `tools` · failure: none — the literal is a folder of the checked-in scripts only in a file that runs a script.
- REQ-4: The reader scan shall list `tests/update_switch_layer.rs` and `tests/update_switch_cli.rs` as readers of `.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1`, `tests/update_switch_cli.rs` as a reader of `install/moai-switch.sh`, `tests/statusline_golden.rs` as a reader of `tools/statusline_golden.py`, `tests/atlas_three_arm.rs` as a reader of `tools`, `tests/ui_summary_terminal_cli.rs` as a reader of `skills/moai` and `tests/e2e_temp_home_cli.rs` as a reader of `skills/moai` · failure: none — a finder that loses one of them fails the live-tree test.
- REQ-5: The table shall carry an `always` scan that names each reader of REQ-4 as a rule file with the path it reads in its scope and a check that runs the reader with `--test <stem>` · failure: the table test of the earlier SPEC fails on a reader with no scan.
- REQ-6: When a fixture test file plants one of the three shapes of REQ-1, REQ-2 and REQ-3 and no scan covers it, the table test shall name the file and the path · failure: none — each shape has its own fixture case.

## Acceptance

- AC-1 (REQ-1): a fixture reader with `include_bytes!("../install/x.sh")` has `install/x.sh` found, and a literal that stays inside `tests/` or has no `..` is skipped · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/include_bytes/)'`
- AC-2 (REQ-2): a fixture file that names `CARGO_MANIFEST_DIR` and spells `skills/moai` or `tools/x.py` with no read call has the path found, and the same file without `CARGO_MANIFEST_DIR` is no reader · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/no_read_call/)'`
- AC-3 (REQ-3): a fixture file with `CARGO_MANIFEST_DIR` and `"python3"` that spells `"tools"` has `tools` found, the same file without an interpreter literal does not · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/interpreter/)'`
- AC-4 (REQ-4): the reader scan on the live tree finds the six pairs · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/more_readers_in_the_live_tree/)'`
- AC-5 (REQ-5): every reader of the live tree has a covering `always` scan or an exemption, and every `always` scan runs the readers it names · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/every_file_reader/) | test(/runs_the_readers/)'`
- AC-6 (REQ-6): each of the three planted shapes without a scan makes the table test fail with the file and the path, and passes once an `always` scan covers it · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/planted_shape/)'`
- AC-7 (REQ-5): deleting the reader's entry or the new detection makes the table test fail (guard mutants of the Result) · verify: the `guard-mutant:` lines of the READY note

## Scope

- write: `.moai/specs/SPEC-20261123-land-scan-more-readers/**`, `tools/repo_scans.json`, `tests/repo_scans_table.rs`
- write only if a test needs it: `tests/fixtures/state_folder_literals/allow.tsv` (one row for the table test if its fixture text adds `.moai/specs` literals; the card scope has to be widened by the leader, as C-398 did)
- read-only: every reader test file, `src/repo_scans.rs`, `src/check_affected.rs`
- forbidden: any other path, the real HOME, `skills/**`
- order: starts after C-398 lands; `tools/repo_scans.json` and `tests/repo_scans_table.rs` are C-398's until then, and C-403 (EXEMPT line in `tests/repo_scans_table.rs`) lands one side first and the other merges NeXT.

## Plan

1. SPEC (this file), `moai spec check`, the card links it (`moai card spec`), the user approves it (never recorded here), C-398 lands, the branch merges NeXT.
2. RED: the three fixture cases and the live-tree case in `tests/repo_scans_table.rs`; each fails on the unchanged finder, the live-tree and `every_file_reader` ones with the list of missing pairs.
3. GREEN, finder: `include_str_paths` reads `include_bytes!(` too, `file_read_paths` drops the read-call gate and counts `tools` next to an interpreter literal.
4. GREEN, table: the new readers go to the `reads-` scans of `tools/repo_scans.json` (rules, scope, `--test`); `install/` joins the scope of the scan that holds `tests/update_switch_cli.rs`.
5. Evidence: nextest with `-E` over `repo_scans_table` and `repo_scans_cli`; clippy; fmt; `git diff --check`; `moai trailer-check`; one bare `check affected`; a guard-mutant line per new guard test.

## Risks

- R-1 over-approximation: with no read-call gate a file that names `CARGO_MANIFEST_DIR` and a repository path only as a fixture name becomes a reader. Measured on the tip: 2 files change, both real; later fixture-only ones go to `FILE_READ_EXEMPT` with a reason, as in C-398 R-1.
- R-2 the `tools` rule is a text match (D-3): an interpreter-spawning test that spells `"tools"` for another reason gets the folder in its scope, which adds a run and never hides one.
- R-3 `tools/repo_scans.json` and `tests/repo_scans_table.rs` collide with C-398 and C-403: this card starts after C-398 lands and merges NeXT before it writes either file.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `include_bytes!` is read by the same function as `include_str!`, no second finder | the two macros take the same relative literal; the measured difference is two files | give `include_bytes!` a function of its own |
| D-2 | The read-call gate of `file_read_paths` goes away; naming `CARGO_MANIFEST_DIR` and spelling a tracked path is enough | measured: of the files naming `CARGO_MANIFEST_DIR`, only 2 have no read call (`statusline_golden`, `ui_summary_terminal_cli`), so the gate hid exactly the two shapes of the card and covered nothing else | bring `FILE_READ_CALLS` back and list the pairs by hand |
| D-3 | The one-segment literal `tools` counts as the folder `tools` in a file that names an interpreter literal | C-398 D-3 keeps one-segment folders out because `docs`, `src`, `skills` are temporary project names; `tools` next to `"python3"` is the script folder of `atlas_three_arm`; a statement-level parser of `runpy` would be a bigger finder for one file | drop REQ-3 and exempt `tests/atlas_three_arm.rs` with a reason |
| D-4 | `tests/e2e_temp_home_cli.rs` needs a check only, no table change | its `skills/moai` read is covered by C-398 (note 7635); the open point of note 7553 was the exemption, which C-398 did not take | add an exemption pair if the live-tree test turns red after a rename |

## Result

(to be written after the work)
