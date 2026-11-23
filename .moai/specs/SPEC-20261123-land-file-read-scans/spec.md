---
id: SPEC-20261123-land-file-read-scans
title: "the repo-scans table covers every integration test that reads a checked-in file at run time, not only those that use include_str!"
status: draft
revision: 2
created: 2026-11-23
---

# the repo-scans table covers every integration test that reads a checked-in file at run time, not only those that use include_str!

## Intent

`SPEC-20261123-land-repo-scans` 는 착지 명령이 `tools/repo_scans.json` 의 스캔을 요구하게 만들고, `include_str!` 로 저장소 파일을 읽는 테스트를 표에 넣게 했어요(`always` 스캔). 그 검사는 컴파일 때 읽는 테스트만 찾아요. 같은 하루에 NeXT 가 네 번 빨개진 마지막 원인이 그 밖에 있어요: `d382fa3af5` 가 `skills/` 파일을 고쳤는데, `tests/mode_workflow_steps.rs:10` 은 `fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows").join(name))` 으로 실행 때 읽어서 표에도 없고 어떤 필요한 명령도 그 테스트를 돌리지 않았어요. C-392 감사 노트 N1 이 이 구멍(D-5 로 뺀 것)을 적었어요.

측정한 크기(팁 `5a61d95ba2`, 2026-11-23): `tests/*.rs` 464개 가운데 `CARGO_MANIFEST_DIR` 를 쓰는 파일은 152개이고, 그중 읽는 호출(`read_to_string`, `fs::read`, `read_dir`, `File::open`)이 있고 저장소에 있는 경로를 글자로 적은 파일이 107개예요. 그 107개 중 표에 범위 항목이 아예 없는 파일이 76개, 이름은 있어도 읽는 경로가 범위 밖인 파일까지 합치면 105개가 어떤 스캔의 범위에도 덮이지 않아요.

이 SPEC 은 세 가지를 바꿔요.

1. 표 테스트(`tests/repo_scans_table.rs`)가 "실행 때 저장소 파일을 읽는 통합 테스트 파일"을 찾고, 읽는 경로마다 그 파일을 규칙 파일로 가진 `always` 스캔의 범위가 덮는지 검사해요. 덮이지 않으면 파일과 경로를 이름으로 빨강이에요.
2. 표에 읽기 스캔을 영역별 묶음으로 더해요(`always`): 한 묶음이 영역(skills, docs, src 의 이름 붙은 파일, 루트 문서, 설정, `.github` 와 설치 템플릿, tools, `.moai`)의 읽기 테스트를 규칙 파일로 갖고, 범위는 그 테스트들이 읽는 경로의 합이며, 검사는 묶음의 테스트를 모두 `--test` 로 돌리는 cargo 명령 한 줄이에요. 착지 때 영수증 단계가 영역마다 하나(`scan:<id>`)여서 한 줄 고친 워크플로 문서가 수십 단계를 요구하지 않아요.
3. 스캔이 규칙 파일로 올린 통합 테스트는 그 스캔의 검사가 실제로 돌려요(`--test <stem>`): 이름만 올리고 돌리지 않는 항목은 빨강이에요.

### 읽은 것 (2026-11-23, 팁 `5a61d95ba2`)

- 카드 C-398 본문·기준 3개, C-392 의 감사 노트(N1, 반복 문제 네 번째 원인).
- `src/repo_scans.rs`(판정: 표의 모양, `always`, `own_hits`), `tools/repo_scans.json`(55개 스캔, 그중 `always` 14개), `tests/repo_scans_table.rs`(`flagged()` 걷기 검사, `include_str!` 검사, 정적 검사 검증).
- `src/check_affected.rs:98-167`: `check affected` 는 스캔의 범위나 규칙 파일이 바뀐 범위에 닿으면 그 스캔 검사의 `--test` 타깃을 자기가 고르는 테스트에 더해요. 그래서 새 항목은 READY 때 같은 head 의 영수증에 자동으로 들어가요.
- `tests/mode_workflow_steps.rs`, `tests/product_cli.rs`, `tests/spec_cli.rs`, `tests/audit_admission.rs`: 저장소 파일을 읽는 모양(상수에 담긴 경로, 도우미 함수의 인자, 임시 폴더에 쓰는 같은 이름의 글자)을 직접 읽었어요.

## Non-goals

- 게이트를 느슨하게 하는 일, 전체 `cargo test` 를 강제하는 일, Python 테스트를 돌리는 일(`CLAUDE.md` "Tests are Rust only").
- `src/` 안의 단위 테스트가 저장소 파일을 읽는 경우: 걷기 검사(`flagged()`)가 직접 걷는 것만 봐요. 이 SPEC 의 대상은 `tests/*.rs` 최상위 파일이에요.
- 경로를 코드에서 조립하는 읽기(글자로 적지 않은 경로): 이 검사는 글자로 적힌 경로만 봐요. 놓치는 쪽이어서 D-2 가 이유와 대응을 적어요.
- `moai repo-scans` 판정(`src/repo_scans.rs`)이나 `card land` 의 동작을 바꾸는 일: 표의 데이터와 표 테스트만 바꿔요.
- 테스트 파일을 지역 crate 로 합치는 일(SPEC-20261123-test-area-crates): 합친 뒤 `--test <stem>` 이름이 바뀌면 그 카드가 이 표도 같이 고쳐요(R-3).

## Requirements

- REQ-1: When the table test runs, the table test shall list each top-level integration test file that names `CARGO_MANIFEST_DIR`, calls `read_to_string`, `fs::read`, `read_dir` or `File::open`, and spells a repository path as a string literal, together with the paths it spells · failure: none — a file that spells none is no reader.
- REQ-2 (REQ-1): When a string literal of a reader has no space and names an existing file or an existing folder of two segments or more outside `tests/`, the reader scan shall count it as a spelled repository path · failure: none — the cut before a placeholder and the folder rule are D-2 and D-3, and a one-segment folder name such as `docs` is a fixture folder name, not a read.
- REQ-3 (REQ-1): IF a path a reader spells is not in the scope of an `always` scan that names the reader in its rules or its readers and the pair of the reader and the path is not in the exemption list, THEN the table test shall fail and name the reader and the path · failure: none — the message names the two ways out, a table entry or an exemption with its reason.
- REQ-4: The table shall carry `always` scans that name every reader of the current tree in their rules or readers with the paths it spells in their scopes except the pairs of the exemption list · failure: REQ-3 makes a missing entry fail.
- REQ-5: When an `always` scan names a top-level test file in its readers or else in its rules, the table test shall require one check of that scan to run that test target with `--test <stem>` · failure: none — a scan that names a reader and does not run it would cover nothing.
- REQ-6: IF the exemption list names a file that the reader scan does not find or names a path the file does not spell or names a file twice or gives a file no reason THEN the table test shall fail · failure: none — the list keeps no stale line.
- REQ-7: The table test shall keep the `include_str!` check of `SPEC-20261123-land-repo-scans` and cover its paths through the same function as the new reader paths · failure: none — the earlier test names stay.
- REQ-8: When a fixture test file spells a repository path and has no covering scan, the table test shall name it · failure: none — the fixture case runs on every table test.
- REQ-9: The table test shall find `tests/mode_workflow_steps.rs` as a reader of `skills/moai/workflows` in the current tree · failure: none — a reader scan that finds nothing fails this one.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9): the reader scan finds `tests/mode_workflow_steps.rs` with `skills/moai/workflows`, finds a fixture file that reads `skills/moai/SKILL.md` through `fs::read_to_string` with `CARGO_MANIFEST_DIR`, skips a one-segment folder literal, a literal with a space, a literal outside the repository and a path inside `tests/` · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/file_read_/)'`
- AC-2 (REQ-3, REQ-8): a reader with no covering scan is named with its path, and the same reader passes after an `always` scan names it in its rules or readers with the path in its scope · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/uncovered_file_reader/)'`
- AC-3 (REQ-3, REQ-4): every reader of the current tree has a covering `always` scan or an exemption with its reason · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/every_file_reader/)'`
- AC-4 (REQ-5): every `always` scan runs the test targets it names in its readers (else in its rules), and a fixture scan that names a test target its checks do not run fails · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/runs_the_readers/)'`
- AC-5 (REQ-6): every exemption is a reader the scan still finds, has a reason and appears once · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/file_read_exemptions/)'`
- AC-6 (REQ-7): the `include_str!` tests of the earlier SPEC still pass through the shared function · verify: `cargo nextest run --locked --offline --test repo_scans_table -E 'test(/include_str/)'`
- AC-7 (REQ-4): the whole table test passes and `moai repo-scans --refs` on a head that edits `skills/moai/workflows/work.md` lists a reads scan in `own_hits` and its check · verify: `cargo nextest run --locked --offline --test repo_scans_table --test repo_scans_cli`; the case is `always_reads_scan_needs_its_checks_for_a_head_that_edits_a_file_its_readers_open`
- AC-8 (REQ-8): deleting the reader scan makes the table test fail (guard mutants M1 to M4 of the Result) · verify: the `guard-mutant:` lines of the READY note

## Scope

- write: `.moai/specs/SPEC-20261123-land-file-read-scans/**`, `tools/repo_scans.json`, `tests/repo_scans_table.rs`, `tests/repo_scans_cli.rs` (two no-hit cases moved off `README.md`, one new case; the card scope was widened by this one file)
- write only if a test needs it: `src/repo_scans.rs`, `skills/moai/references/source-manifest.json` (hash only)
- read-only: `src/check_affected.rs`, `src/card/land/scans.rs`, `tests/areas.toml`, every reader test file
- forbidden: `src/card/land.rs`, `src/bundle.rs`, `src/profile.rs`, `skills/**` other than the manifest hash, any other path, the real HOME

## Plan

1. SPEC (this file), then `moai spec check`; the card links it (`moai card spec`); the user approves it (never recorded here).
2. RED: the reader scan, its fixture cases and the three table checks in `tests/repo_scans_table.rs`; each new test fails on the unchanged table, the live-tree ones with the list of uncovered readers.
3. GREEN, table: the reads scans in `tools/repo_scans.json`, written from the reader scan's output and read against each reader file; fixture-only readers go to the exemption list with their reason.
4. GREEN, runs-the-readers check over every `always` scan.
5. Evidence: nextest with a filterset over `repo_scans_table`, `repo_scans_cli`, `check_affected*`; clippy; fmt; `git diff --check`; `moai trailer-check`; one bare `check affected`; a guard-mutant line per new guard test.

## Risks

- R-1 over-approximation: a reader that spells `src/lib.rs` only as the name of a file in its temporary project gets that path in its scope, so a head that touches `src/lib.rs` runs its test too. A wider scope only adds a run; it never hides one. Fixture-only readers (every spelled path is a fixture name) go to the exemption list, each with the reason read from its file.
- R-2 groups are big: a head that edits a workflow text runs every test of the `skills` group (70 readers when written). One cargo command compiles them once; the receipt is one step per group. The leader may split a group in a later card when the run time shows it. Editing one reader does not run the group (D-6).
- R-3 the test-area-crates card renames test targets: it must change the `--test <stem>` names of this table in the same card; AC-4 turns a missed one red.
- R-4 the exemption list can hide a real read added later to an exempt file: the list is a reviewed list with reasons, like `EXEMPT` for the walkers, and REQ-6 keeps it from going stale.
- R-5 landings run in observe mode (`work_flow.land_verdict`): the new scans change what the landing would require, not what it refuses today. The would-refuse count with the new scans is recorded in the Result; the switch to enforce stays the user's recorded decision.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | One `always` scan per area with all its readers in a `readers` list and one multi-`--test` check, not one scan per reader | the measured 105 readers would be 105 receipt steps for one workflow edit (`scan:<id>` is one step per scan, `SPEC-20261123-land-repo-scans` D-3); `check_affected.rs:138-147` already reads every `--test` of a check | split a group into scans by listing the same readers in two entries |
| D-2 | The reader scan reads string literals only and takes every tracked-looking one (over-approximation), with a reviewed exemption list for fixture-only files | the paths come through constants, helper arguments and `.join` chains, so a statement-level parser would miss reads; a text match is a hypothesis (CLAUDE.md), so a wrong match costs an exemption line, a missed read costs a red NeXT | tighten the literal rule in `file_read_paths` |
| D-3 | Folder literals count at two segments or more, file literals at any depth | `docs`, `src`, `skills` alone are the names of temporary project folders in 60+ fixtures | count single folders too and add them to the exemption list |
| D-4 | The check is `tests/*.rs` top level only, in the table test, no product code | the card names that set (147 files); `src/repo_scans.rs` and the landing stay unchanged, so the landing's observe mode and receipts need no change | move `file_read_paths` into `src/repo_scans.rs` in a later card |
| D-5 | REQ-5 asks every `always` scan to run the test targets it names | a scan that lists a reader without running it would pass REQ-3 and still miss the case | drop REQ-5 and its test |
| D-6 (revision 2) | The readers of a `reads-` scan sit in a new `readers` key, not in `rules`; its one rule file is `tests/repo_scans_table.rs`, the test that keeps the scan whole | `check affected` selects a scan's targets when any rule file changed (`check_affected.rs:109-117`) and the judge cross-hits on a rule change: with 70 readers as rule files, editing one test of the `skills` group would run all 70 and a lane editing one reader would force a re-merge check against any `skills/` change on the tip. The parser ignores unknown keys (`src/repo_scans.rs` reads `id`, `rules`, `scope`, `checks`, `always`), so no product code changes | move the names from `readers` to `rules` and drop the `readers` lines of `names_file` and `scans_not_running_readers` |
| D-7 (revision 2) | The exemption is `(file, paths, reason)` per path, and lists only the names that change often: `src/lib.rs`, `src/main.rs`, `README.md` and `.moai/specs` | measured on the tree: 24 files spell those four as names inside the temporary projects they build, and `src/lib.rs`, `README.md` and `.moai/specs` change in most cards, so each would make an unrelated test run on every card; rarer fixture names (`Cargo.toml`, `.gitignore`, `.moai-cli/..`) stay in scopes (D-2) | list more names, or empty the list and accept the extra runs |

## Result

(draft: filled when the work is done)
