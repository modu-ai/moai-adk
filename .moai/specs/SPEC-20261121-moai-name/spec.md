---
id: SPEC-20261121-moai-name
title: "The temporary name moai-cli leaves the product: every remaining occurrence, by kind and slice"
status: draft
revision: 7
created: 2026-11-21
---

# The temporary name moai-cli leaves the product

## Intent

사용자 지시(2026-11-21, 리더 대화 원문): "moai-cli 전수 조사해서 또 수정에 빠져 있는거 찾아서 완전 제거해라". 사용자 결정(2026-11-22, 카드 노트 6286): 릴리스 태그는 `v4.<yearweek>.<build>`, 내려받는 파일은 `moai-adk-<version>-<platform>` 이고 `moai-cli` 이름은 남기지 않아요. 임시 이름 `moai-cli`(바이너리·패키지·크레이트·폴더·변수·릴리스 이름)는 제품에서 완전히 사라지고 이름은 `moai` 하나예요.

이 SPEC 은 그 전체의 목록과 순서예요. 조각마다 이 SPEC 의 개정과 착지가 따로 있고, 각 조각의 요구는 아래 Requirements 에 그 조각이 시작될 때 더해요. 개정 2 는 조각 1a 예요: 남은 발생 전부를 종류별·조각별로 나열하고, 출력되는 버전 줄이 `moai` 를 이름하기 전에 읽는 쪽이 두 이름을 모두 받게 해요(착지됨). 개정 3 은 조각 1b 예요: 실제로 찍는 이름을 `moai` 로 바꾸고, 1a 감사에서 나온 비차단 지적(N1·N2·N4·N6)을 이 문서에서 바로잡아요. 개정 4 는 조각 1b 의 수리예요: 1b 감사가 막은 것(실제 바이너리의 버전 줄을 옛 이름으로 고정한 `tests/install_candidates.rs`)과, 원격 CI 를 미리 빨강으로 만들 읽는 쪽(`tools/smoke/smoke.py`, README 두 곳, `tests/oss_docs_test.py`)을 같은 조각에서 새 서명에 맞춰요. 개정 5 는 조각 2 예요: 사용자가 읽는 거절 문구·경고·안내 문장과 도움말 골든에서 제품 이름을 `moai` 로 쓰고(파일·폴더 이름은 다른 조각), 1b 감사의 비차단 지적 N-b(낡은 순서 문장과 읽는 쪽 표)·N-c(Result 에 없는 Python 단계)를 이 문서에서 바로잡아요. 개정 6 은 조각 4 예요(heavy-slot 캐시 폴더 이름만): 슬롯 폴더의 기본값을 `moai-cli/heavy-slots` 에서 `moai/heavy-slots` 로 옮기고, 같은 폴더와 같은 잠금을 쓰는 Python 쌍둥이 도구를 한꺼번에 맞추며, 옛 폴더를 읽는 호환과 옮기는 방법을 아래 결정 D-9~D-13 으로 정해요.

### 왜 다시 쓰나 (2026-11-23 확인)

개정 1 의 조각 1 커밋 `6f0896b2f8` 은 NeXT 의 조상이지만, 병합 `b6e7f8cfcb`("통합 브랜치가 이미 하는 것은 가져온다")가 이 SPEC 파일과 `src/bin_names.rs` 의 변경을 지웠어요(`git log --full-history -m --diff-filter=D` 로 확인). 그래서 NeXT 팁 `2ade55ba2d` 에는 SPEC 도, 두 이름을 받는 같은 빌드 검사도 없어요. 이 개정이 그 일을 팁 위에서 다시 해요.

### 버전 줄을 읽는 쪽 (측정: NeXT `2ade55ba2d` 의 조각 1a 시작 전, 그 뒤 상태는 오른쪽 열)

| 읽는 쪽 | 위치 | 1a 시작 전 두 이름을 받았나 | 지금(조각 1a·1b 착지 뒤) |
| --- | --- | --- | --- |
| 설치 세대 판별 `generation_of` | `src/install.rs` `is_rust_signature` | 받았어요 (`moai-cli` 와 `moai`, 단위 테스트가 고정) | 그대로 받아요 |
| 같은 빌드 검사 (`bin-names-plan` 의 `bin/moai`) | `src/bin_names.rs` | 아니요 — 정확히 `moai-cli <sig>` 만 | 두 이름을 받아요(1a) |
| 릴리스 워크플로우 버전 검사 | `.github/workflows/release.yml` 단계 `verify` | 아니요 — 정확히 `moai-cli ${VERSION} (rust)` 만 | 두 이름을 받아요(1a) |
| 스킬 fallback 문장 | `skills/moai/SKILL.md` | 아니요 — `moai-cli <version> (rust)` 만 | 두 이름을 받아요(1a) |
| Python 도구·테스트(`tools/smoke`, `tools/statusline_golden.py`, `tests/*.py`) | Python | 이 카드 밖(Rust-only 카드) | `smoke.py` 정규식과 `oss_docs_test.py` 기대 한 줄씩만 1b 수리에서 고쳤어요 |

settings cleaner 는 버전 줄을 읽지 않아요(`~/.moai-cli/path/moai-cli` 런처 경로와 훅 명령을 지울 뿐) — 이름을 바꾸는 쪽은 조각 5 예요.

## Non-goals

- 다른 카드가 이미 쥔 것은 여기서 다시 하지 않아요: 상태 폴더 기본값과 HOME 마이그레이션(path-root 카드), 프로젝트 측 폴더 이동, Go 링크와 name-switch 스크립트 제거, Python 파일(Rust-only 카드), Codex/OMP 제거.
- 조각 4(캐시·임시 폴더 이름)와 조각 10(crate 이름)은 레인이 멈춘 뒤 리더가 예약해요.
- 키 볼트 유도선은 불변이에요(조각 9, guarded: 볼트 내용을 열지 않아요).
- 설치된 `~/.moai` 와 `~/.moai-cli` 의 링크와 폴더는 건드리지 않아요. 그 정리는 착지 뒤 사용자 선택이에요.
- NeXT 에 있는 빨강 1개(`mode_workflow_steps`)는 이 SPEC 의 범위 밖이에요(수리 중, 2026-11-23 리더 배차).

## Requirements

- REQ-1: When the same-build check of `bin-names-plan` reads the first `--version` line of a separate `bin/moai`, it shall count `moai <build version> (rust)` and `moai-cli <build version> (rust)` as this build and any other line as another program · failure: none — a line it does not know leaves `bin/moai` untouched and reports `blocked`.
- REQ-2: When the release workflow verifies the built binaries, it shall accept the first `--version` line `moai ${VERSION} (rust)` or `moai-cli ${VERSION} (rust)` for each binary · failure: none — REQ-7 handles a line that is not one of the two.
- REQ-3: The skill's fallback sentence shall name the Rust signature `moai <version> (rust)` first and accept `moai-cli <version> (rust)` as the same signature from an earlier build · failure: none.
- REQ-4: The install generation reader shall keep accepting both names · failure: none — pinned by the existing unit test `generation_of`.
- REQ-5: When `moai --version` runs, the first line shall begin with `moai ` and end with ` (rust)` for both binary names · failure: none — a build that still prints the earlier name stays readable through REQ-1 to REQ-4 until this requirement lands.
- REQ-6: The occurrence table of this SPEC shall list every kind of remaining occurrence with its slice · failure: none.
- REQ-7 (REQ-2): IF a built binary prints a first `--version` line other than those two, THEN the release workflow shall fail the job with a message naming both accepted forms before it writes the release note.
- REQ-8: When a test, a script or a document reads or quotes the first `--version` line of the built binary, it shall accept `moai <version> (rust)` and, where a binary of an earlier build can still answer, `moai-cli <version> (rust)` · failure: none — a reader that stays pinned to the earlier name fails once REQ-5 has landed.
- REQ-9: When the program tells a person about itself in a refusal, a warning, a note or a command to type, it shall call the product `moai` · failure: none — a sentence that names a file or folder on disk (`bin/moai-cli`, `~/.moai-cli/...`) keeps that name until the slice that renames the file (decisions D-6 to D-8).
- REQ-10: The help goldens under `tests/fixtures/help/` shall contain `moai-cli` only inside a state folder path or in the two `bin-names-plan` sentences that name the file on disk · failure: none — a golden line that names the old name as a command fails the golden guard.
- REQ-11: When `moai heavy-slot-acquire` or `moai heavy-slot-status` or `moai check affected` runs without a slot folder argument, it shall use the folder `moai/heavy-slots` of the user's cache folder · failure: none — a missing HOME and cache variable still refuses as before (`HOME is not set`, exit 2).
- REQ-12: When the Python twin `tools/heavy_check_slot.py` runs without `--dir`, it shall use the same default folder as REQ-11 · failure: none — two tools on two folders would stop excluding each other, so the cross-tool test fails first.
- REQ-13: When the default slot folder holds no `limits.json` and the earlier folder `moai-cli/heavy-slots` of the cache folder holds one, the limits reader shall read that earlier file by the same rules · failure: none — an earlier file that breaks those rules (symlink or folder or over 4096 bytes or not an object or an unknown key or `max` 0) refuses the run with exit 125 exactly as it would in the new folder, and nothing under the earlier folder is ever written, moved or created.
- REQ-14: When `moai heavy-slot-status` reads the default folder and the earlier folder `moai-cli/heavy-slots` exists, it shall list the slots held there under the key `earlier_folder` · failure: none — an earlier folder that cannot be read answers `null`, and a run with `--dir` carries no such key.
- REQ-15: When a migrate command opens a home folder, it shall require the marker file `.moai-synthetic-home` holding exactly `moai synthetic home v1` and a newline · failure: none — REQ-17 handles a home without it.
- REQ-16: When the launcher runs `git init` in a folder that has no repository, it shall hold its lock in the OS temporary folder under the name `moai-git-init-<16 hex of sha256(path)>.lock` · failure: none — a lock that cannot be opened safely (a symlink, another owner, an unopenable path) still gives no lock and the init runs unlocked as before, and `moai-cli-git-init-*.lock` files an earlier build left behind are never read, moved or deleted.
- REQ-17 (REQ-15): IF the home folder lacks the marker file `.moai-synthetic-home` with exactly that text, including a home that carries only the earlier marker `.moai-cli-synthetic-home`, THEN the migrate command shall refuse with `migrate refused: .moai-synthetic-home marker missing; only a prepared synthetic home is allowed` and read and write nothing in that home.

## Order that keeps installed copies working

설치된 훅·스킬·설정은 옛 이름을 부르는 채로 있어요. 순서가 그것을 지켜요.

1. **읽는 쪽을 먼저 둘 다 받게 한다(조각 1a, 착지됨).** 이 단계 동안에는 아무 바이너리도 이름을 바꿔 찍지 않아요. 같은 빌드 검사·릴리스 검사·스킬 문장이 두 이름을 받아서, 이후 단계에서 어느 쪽이 찍어도 안 깨져요.
2. **찍는 이름을 `moai` 로 바꾼다(조각 1b).** `src/main.rs` 의 clap 이름 한 줄과 그 위 주석, 그 줄을 보는 테스트. 옛 설치(`moai-cli <ver> (rust)` 를 찍는 빌드)는 읽는 쪽이 계속 받으므로 읽기가 안 깨져요. `src/main.rs` 는 리더 결정(2026-11-23 조각 1b 배차)으로 이 SPEC 쪽으로 풀렸어요.
3. **거절 문구·경고·안내·도움말 골든(조각 2, 개정 5).** 사용자에게 보이는 글만이에요: 제품 이름을 말하는 문장은 `moai`, 파일·폴더 이름을 말하는 문장은 그 이름을 바꾸는 조각(3·4·5·7)까지 그대로예요.
4. **바이너리와 패키지(조각 5).** 새 이름 `moai` 가 실제 파일이 되고 `moai-cli` 는 호환 링크로 남아, 옛 훅·설정이 부르는 `moai-cli` 가 계속 돼요. 링크는 업데이트 단계가 지워요(카드 본문 5). 임시 HOME 에서 도는 Rust 테스트가 이 호환 읽기를 증명해요.
5. **슬롯 폴더 이름(조각 4, 개정 6).** 모든 레인이 한꺼번에 새 폴더로 옮겨요: 두 빌드가 서로 다른 슬롯 폴더를 쓰면 서로를 막지 못해요. 새 빌드는 새 폴더(`moai/heavy-slots`)를 쓰고, 기계 주인의 `limits.json` 은 새 폴더에 아직 없을 때만 옛 폴더에서 읽어요(읽기만). 옛 폴더는 코드가 옮기거나 지우지 않고 주인이 한 번 `mv` 해요(결정 D-9~D-13).
- **5b. 임시 이름(조각 4b, 개정 7).** 잠금 파일과 임시 HOME 표식은 설치된 사본이 읽지 않는 이름이라 호환 읽기가 필요 없어요: 새 이름으로 바꾸고 옛 이름은 읽지도 만들지도 않아요(결정 D-15·D-16). 같은 조각이 조각 4 감사의 N1 테스트 두 개를 더해요.
6. **릴리스 이름(조각 6).** 태그 `v4.<yearweek>.<build>`, 파일 `moai-adk-<version>-<platform>`(결정 6286). 옛 설치 스크립트가 보게 되는 것은 그때 이 SPEC 이 적어요.

## Occurrence table

측정: 2026-11-23 NeXT `2ade55ba2d`, `git ls-files -z | xargs -0 rg -c -e 'moai-cli|moai_cli'` — **1061 파일, 7208 줄**. 종류별 줄 수는 리더의 조사(2026-11-21, tip `6e1bedb4f3`, 999 파일·6969 줄, `classified2.tsv`)에서 가져왔고 이 개정에서 다시 세지 않았어요(그 사이 늘었을 수 있어요; 조각 11 의 검사 테스트가 새 발생을 막아요).

| Kind | Lines (2026-11-21) | What it is | Slice |
| --- | --- | --- | --- |
| i-a state folder in tests | 1499 | tests 의 `.moai-cli` 상태 폴더 경로 | 8 (path-root 뒤, resolver 로) |
| f tracked `.moai-cli` tree + ignores | 1284 | tracked 상태 트리 157 파일·ignore·scan 설정 | 7 (path-root 착지 뒤) |
| h other docs/history | 974 | 문서·design-review·specs 이력 | 11 (이력은 이력으로 유지) |
| c1 crate path in src | 695 | `moai_cli::` crate 경로(src) | 10 (리더 예약, 레인 멈춤) |
| i-c crate path in tests | 683 | `moai_cli::`(tests) | 10 |
| k comments in src | 392 | src 주석 | 11 |
| j Python | 252 | Python 파일 | Rust-only 카드 (제외) |
| i-o binary word in tests | 348 | tests 의 바이너리 이름 그 밖 | 8 |
| i2 unit tests in src | 206 | src 내부 단위 테스트 | 10 |
| g skills/templates/agents | 392 | 스킬·템플릿 텍스트 | 1a (fallback 문장), 2, 11 |
| i-d env/file names in tests | 187 | tests 의 env·파일 이름 | 3 |
| i3 fixtures+goldens | 155 | fixture·golden | 2 |
| e installer+release | 110 | install/·.github/·scripts/·dist-workspace | 1a (릴리스 검사 한 곳), 5, 6 |
| e2 dev scripts+config | 45 | .cargo/·.config/ 개발 설정 | 3, 5 |
| c2 binary-name literal in src | 43 | src 의 `moai-cli` 리터럴 | 1a (같은 빌드 검사), 1b (clap 이름), 2 (사용자 문자열) |
| d env/file names in src | 34 | `MOAI_CLI_` 계열(src) | 3 |
| b user-visible strings | 20 | help·거절 메시지 | 2 (상태 경로분은 path-root 뒤) |
| c0 manifests | 10 | Cargo.toml·lock·.cargo | 5 |
| a state folder in src | 6 | src 의 `.moai-cli` 상태 폴더 | path-root 카드 (제외) |

## Acceptance

- AC-1 (REQ-1): 별도 `bin/moai` 가 `moai <sig>` 또는 `moai-cli <sig>` 를 찍으면 `replace_copy`, 다른 빌드 번호·`moai-adk 3.2.0`·모르는 이름은 `blocked` 예요 · verify: `cargo nextest run --test bin_names_signature_cli -E 'test(/^a_separate_moai/)'`
- AC-2 (REQ-2): 릴리스 워크플로우의 `verify` 단계 스크립트를 가짜 바이너리로 실제로 돌려, 두 이름의 네 조합은 통과하고 다른 버전·Go 줄·모르는 이름·`(rust)` 없는 줄은 실패해요 · verify: `cargo nextest run --test bin_names_signature_cli -E 'test(/^the_release_check/)'`
- AC-3 (REQ-3): 스킬의 버전 검사 문장이 `moai <version> (rust)` 를 먼저, `moai-cli <version> (rust)` 를 같은 시그니처로 적고 fallback 은 `moai-cli` 예요 · verify: `cargo nextest run --test bin_names_signature_cli -E 'test(/^the_skill_fallback/)'`
- AC-4 (REQ-4): 설치 세대 판별이 두 이름을 rust 로 받아요 · verify: `cargo nextest run --lib -E 'test(the_rust_signature_is_matched_by_shape_under_both_names)'`
- AC-5 (REQ-5): `--version` 의 첫 줄이 두 바이너리 이름 모두에서 `moai <version> (rust)` 이고 usage·help 이름이 `moai` 예요 — 조각 1b · verify: `cargo nextest run --test cli_name_moai -E 'test(both_names_print_the_same_output)'`, `cargo nextest run --test help_layout -E 'test(every_name_prints_moai_and_the_version_line_is_unchanged)'`, `cargo nextest run --bin moai-cli -E 'test(every_binary_name_renders_the_same_moai_help)'`
- AC-6 (REQ-6): 이 문서의 Occurrence table 이 종류별 줄 수·의미·조각을 전부 나열해요 · verify: `moai spec check SPEC-20261121-moai-name` 의 errors 가 0 이에요.
- AC-7 (REQ-8): 실제 바이너리가 찍은 첫 줄을 읽는 `install-inspect` 테스트가 `moai <ver> (rust)` 와 `moai-cli <ver> (rust)` 를 모두 받고, 스모크 도구의 정규식과 README 두 곳의 서명, 문서 테스트의 기대가 새 서명이에요 — 조각 1b 수리 · verify: `cargo nextest run --test install_candidates -E 'test(every_moai_is_listed_in_order_and_told_apart_without_writing)'`. Python 쪽(`tools/smoke/smoke.py`, `tests/oss_docs_test.py`)은 고치기만 했고 돌리지 않았어요: 손으로도 `moai check affected` 로도 돌지 않으므로 `not_run` 이고, 원격 CI 의 첫 실행이 그 확인이에요.
- AC-8 (REQ-9): 거절·경고·안내 문장이 `moai` 를 말하고 옛 이름의 문장은 소스에 없어요 · verify: `cargo nextest run --test help_text_names_moai -E 'test(every_refusal_and_note_sentence_names_moai)'`, `cargo nextest run --test skill_name_collision`, `cargo nextest run --test update_switch_cli --test update_switch_layer`, `cargo nextest run --lib -E 'test(/key_vault/)'`
- AC-9 (REQ-10): 도움말 골든의 `moai-cli` 는 상태 폴더 경로와 `bin-names-plan` 의 파일 문장 둘뿐이에요 · verify: `cargo nextest run --test help_text_names_moai -E 'test(the_help_goldens_keep_the_old_name_only_for_folders_and_the_alias_file)'`, `cargo nextest run --test help_layout -E 'test(moai_cli_help_names_no_command_moai_cli)'`
- AC-10 (REQ-11, REQ-12): `--dir` 없이 `heavy-slot-acquire`·`heavy-slot-status` 와 Python 쌍둥이가 임시 HOME 의 `moai/heavy-slots` 를 같은 폴더로 풀어 서로 번갈아 돌고(세 경우: 캐시 변수 있음·빈 값·없음), 옛 `moai-cli` 폴더에는 아무것도 만들지 않아요 · verify: `cargo nextest run --test heavy_slot_cli -E 'test(without_dir_both_tools_use_the_same_default_folder) | test(the_default_folder_is_moai_and_nothing_is_made_under_the_earlier_one)'`
- AC-11 (REQ-13): 새 폴더에 `limits.json` 이 없으면 옛 폴더의 것을 읽고(`heavy-slot-status` 의 `build_jobs_cap.slots`), 새 폴더에 생기면 그쪽이 이기고, 옛 파일이 규칙을 어기면 125 이고, `--dir` 로 다른 폴더를 주면 옛 파일은 읽지 않고, 옛 폴더의 바이트·목록은 그대로예요 · verify: `cargo nextest run --test heavy_slot_cli -E 'test(the_earlier_limits_file_is_read_until_the_new_folder_has_one)'`
- AC-12 (REQ-14): 옛 폴더에 쥔 슬롯이 있으면 `heavy-slot-status` 가 `earlier_folder.held` 에 보여 줘요 · verify: `cargo nextest run --test heavy_slot_cli -E 'test(status_lists_the_slots_held_in_the_earlier_folder)'`
- AC-13 (REQ-15, REQ-17): 표식 상수가 `.moai-synthetic-home` 와 `moai synthetic home v1` 줄바꿈이고, 새 표식이 있는 임시 HOME 은 열리고, 옛 표식 `.moai-cli-synthetic-home` 만 있는 임시 HOME 은 새 표식을 이름 붙인 문장으로 거절돼요; 표식을 말하는 도움말·문서 문장도 새 이름이에요 · verify: `cargo nextest run --test migrate_cli -E 'test(the_synthetic_home_marker_is_moai_and_the_earlier_marker_is_refused) | test(migrate_help_names_the_moai_marker_only)'`
- AC-14 (REQ-16): 런처의 `git init` 잠금이 임시 폴더의 `moai-git-init-<16 hex>.lock` 이고, 그 자리에 심은 심볼릭 링크는 따라가지 않고, 다른 계정 소유 잠금은 기다리지 않고, 옛 이름 잠금 파일은 읽지도 지우지도 않아요 · verify: `cargo nextest run --test launcher_auto_init -E 'test(a_symlink_planted_at_the_git_init_lock_is_not_followed) | test(a_git_init_lock_not_owned_by_the_home_owner_is_never_waited_on) | test(the_git_init_lock_is_named_moai_and_an_earlier_one_is_left_alone)'`
- AC-15 (REQ-13): 새 폴더가 있는데 `limits.json` 이 없을 때도 옛 폴더의 한도를 읽고(`build_jobs_cap.slots`), `heavy-slot-acquire` 가 그 한도를 실제로 적용해 옛 한도가 1 이면 두 번째 실행이 첫째가 놓을 때까지 기다려요 — 조각 4 감사 N1 의 두 경우 · verify: `cargo nextest run --test heavy_slot_cli -E 'test(the_earlier_limits_apply_while_the_new_folder_exists_without_limits) | test(acquire_applies_the_earlier_limits_file_to_the_run)'`

## Scope

- write (조각 1a): `.moai/specs/SPEC-20261121-moai-name/spec.md`, `src/bin_names.rs`, `skills/moai/SKILL.md`, `skills/moai/references/source-manifest.json`(SKILL.md 의 sha256 한 줄), `.github/workflows/release.yml`(단계 `verify` 만), `tests/bin_names_signature_cli.rs`
- read-only: `src/install.rs`(이미 두 이름을 받음), 조사 자료(리더 scratchpad `mcli/`)
- write (조각 1b): 이 SPEC 폴더, `src/main.rs`(clap `name` 한 줄과 그 위 주석, 단위 테스트 `every_binary_name_renders_the_same_moai_help` 의 기대값, 이름을 설명하는 주석), `src/lib.rs`·`src/bin/moai.rs`(주석만), `tests/cli_name_moai.rs`, `tests/help_layout.rs`
- write (조각 1b 수리, 개정 4): `tests/install_candidates.rs`(실제 바이너리 버전 줄 기대), `tools/smoke/smoke.py`(버전 줄 정규식 한 곳), `README.md`·`README.ko.md`(서명 인용 한 줄씩), `tests/oss_docs_test.py`(서명 기대 한 줄), 이 SPEC 폴더. Python 파일은 위 한 줄씩만 고쳐요.
- write (조각 2, 개정 5): 이 SPEC 폴더, `src/update.rs`(Go 시대 스킬 안내 문장 한 곳), `src/update/switch.rs`(되돌리기 명령 두 곳), `src/go_state.rs`(경고 두 곳), `src/session.rs`·`src/launcher_terminal.rs`(키 보관함 안내), `src/key_vault.rs`(거절 문장 한 곳과 그 단위 테스트), `src/launcher_init.rs`(첫 실행 profile 의 purpose 문장), `src/state_root.rs`(더 새로운 표식 거절 문장), `src/claude.rs`(설정 백업 폴더의 ignore 머리말), `tests/skill_name_collision.rs`, `tests/update_switch_cli.rs`, `tests/update_switch_layer.rs`, `tests/placeholder_profile_guard.rs`, `tests/help_text_names_moai.rs`(새)
- write (조각 4, 개정 6): 이 SPEC 폴더, `src/heavy_slot.rs`(기본 폴더 함수·limits 읽기·status), `src/main.rs`(`--dir` 도움말 주석 한 곳), `src/help.rs`(같은 문장), `tools/heavy_check_slot.py`(기본 폴더 한 곳과 머리말, 돌리지 않아요), `tests/heavy_slot_cli.rs`, `docs/rust-product-path.md`·`skills/moai/workflows/run.md`·`skills/moai/references/source-manifest.json`(run.md 의 sha256 한 줄)에서 기본 폴더 경로를 말하는 문장
- write (조각 4b, 개정 7): 이 SPEC 폴더, `src/migrate.rs`(표식 상수 둘과 거절 문장), `src/launcher_init.rs`(잠금 이름과 그 설명 주석), `src/help.rs`·`src/main.rs`(표식을 말하는 도움말 문장), `docs/rust-product-path.md`(표식을 말하는 문장 한 곳), `tests/launcher_auto_init.rs`(잠금 이름 두 곳과 새 테스트 하나), `tests/heavy_slot_cli.rs`(N1 테스트 둘), `tests/migrate_cli.rs`(표식 테스트 둘; 새 파일은 테스트 영역 표와 상태 폴더 글자 허용 목록에 걸려서 이미 표에 있는 이 파일로 옮겼어요, 리더 노트 7800 의 A)
- forbidden: 다른 카드가 쥔 파일(상태 폴더 resolver·상수는 path-root 카드, Python 파일은 Rust-only 카드), 조각 10 의 파일과 조각 4b 가 이름 붙이지 않은 것(파이썬 슬롯 폴더 `moai-cli/python-slots`, `tools/self_check.sh`·`tools/local_check.sh`, 임시 폴더 접두사 `moai-cli-home-`·`moai-cli-profiles-`·`moai-cli-gitconfig-`), 설치된 홈 상태 폴더(`.moai`, `.moai-cli`), `Cargo.toml` 의 패키지·bin 이름과 `install.sh`(조각 5, 이번 조각이 아니에요), 카드 번호를 SPEC id·브랜치·커밋·코드에 넣기

## Plan

1. RED: `tests/bin_names_signature_cli.rs`(같은 빌드 검사, 릴리스 스크립트, 스킬 문장)이 빨강인 것을 먼저 봐요.
2. `src/bin_names.rs`: 같은 빌드 검사를 함수로 빼고 두 이름을 받게 해요(+ 단위 테스트).
3. `.github/workflows/release.yml` 단계 `verify`: 두 이름을 받고, 메모에는 바이너리가 찍은 첫 줄을 그대로 적어요.
4. `skills/moai/SKILL.md` 문장 재작성과 `source-manifest.json` 의 해시 갱신.
5. 확인: 위 AC 의 테스트, `moai check affected`, `moai spec check`. 읽는 쪽이 끝난 뒤 조각 1b(clap 이름)를 이어요(아래 Plan 1b).

### Plan 1b (개정 3)

1. RED: `every_binary_name_renders_the_same_moai_help`·`both_names_print_the_same_output`·`every_name_prints_moai_and_the_version_line_is_unchanged` 의 기대값을 `moai <ver> (rust)` 로 먼저 고쳐 빨강인 것을 봐요(옛 이름을 찍어서 실패해야 해요).
2. `src/main.rs` 의 clap `name` 을 `moai` 로 바꾸고 그 위 주석을 고쳐요. `src/lib.rs`·`src/bin/moai.rs` 의 `--version` 설명 주석도 새 이름으로 맞춰요.
3. 같은 세 테스트가 초록이 되고, 이름을 `moai` 로 바꾸면 안 되는 읽는 쪽(조각 1a)이 그대로 초록인 것을 확인해요.
4. 1a 감사의 비차단 지적 정리: AC-6 verify 를 id 형식으로(N1), 낡은 보류 문장과 카드 번호 제거(N2), Result 영수증 head(N4), 문서 속 번역투 낱말을 개발자가 쓰는 말(테스트·브랜치)로(N6).
5. 이번 조각이 아닌 것: `Cargo.toml` 패키지·bin 이름과 `install.sh`(조각 5), 도움말 글과 골든(조각 2).

### Plan 1b 수리 (개정 4)

1. RED: 현재 head 에서 `cargo nextest run --test install_candidates` 가 실제 바이너리의 `moai <ver> (rust)` 를 옛 이름 `starts_with` 에 걸어 실패하는 것(exit 100)을 먼저 봐요.
2. `tests/install_candidates.rs` 가 `moai ` 와 `moai-cli ` 로 시작하고 ` (rust)` 로 끝나는 줄을 받게 해요. `tools/smoke/smoke.py` 정규식은 두 이름으로, README 두 곳과 `tests/oss_docs_test.py` 는 새 서명으로 고쳐요.
3. 실제 바이너리가 찍은 버전 줄을 읽는 테스트를 `moai-cli 4\.|version_line|--version` 으로 찾아, 고른 대상을 nextest 로 모두 돌려요(지난 `check affected` 는 `install_candidates` 를 고르지 않았어요).
4. 이번 수리가 하지 않는 것: `check affected` 가 `src/main.rs` 변경에 바이너리 출력을 고정한 테스트를 고르게 하는 일은 같은 문제가 한 번 더 일어날 때 카드로 열어요(리더 기록).

### Plan 2 (개정 5)

1. 찾기: 빌드한 두 바이너리의 `-h`·`--help`·`help --all` 과 모든 명령의 `--help` 출력에서 명령 이름으로 쓰인 `moai-cli` 를 찾아요(기준 출력). 소스에서는 따옴표 안의 사용자 문장만 골라요(주석·파일 경로·테스트 입력 제외).
2. RED: 기대값을 새 문장으로 먼저 고친 테스트들(`update_switch_cli`, `update_switch_layer`, `skill_name_collision`, `key_vault` 단위 테스트)과 새 `tests/help_text_names_moai.rs` 의 문장 표가 옛 이름 때문에 실패하는 것을 먼저 봐요.
3. 소스 문장을 `moai` 로 바꿔요(Scope 의 파일만). 문장 바깥의 코드는 그대로예요(키 보관함 유도선 포함).
4. 도움말 골든은 허용된 줄(상태 폴더 경로, `bin-names-plan` 의 파일 문장)만 남은 것을 `the_help_goldens_keep_the_old_name_only_for_folders_and_the_alias_file` 로 고정해요.
5. 실제 바이너리 출력을 읽는 테스트를 `moai-cli 4\.|version_line|--version|revert-switch|moai-cli` 로 찾아 nextest 로 모두 돌려요.
6. 1b 감사 N-b(낡은 순서 문장·읽는 쪽 표)와 N-c(AC-7 의 없는 Python 단계)를 이 문서에서 고쳐요(위 본문에 반영).
7. 이번 조각이 아닌 것: 파일·폴더 이름(`bin/moai-cli`, `~/.moai-cli/...`, `moai-cli-home-` 같은 임시 이름), 환경 변수(조각 3), 캐시·임시 폴더 이름(조각 4), 패키지·bin 이름(조각 5), 스킬·템플릿 문장(조각 11).

### Plan 4 (개정 6)

1. 찾기: 슬롯 기본 폴더를 푸는 곳(`heavy_slot::default_dir` 와 그 호출자 셋: `main.rs` 의 acquire·status, `check_affected.rs`)과 쌍둥이 도구의 `default_dir`, 그 폴더를 말하는 도움말·문서 문장을 `rg` 로 모두 세요.
2. RED: `tests/heavy_slot_cli.rs` 에서 `without_dir_both_tools_use_the_same_default_folder` 의 기대 폴더를 `moai/heavy-slots` 로 먼저 고치고, 새 테스트 셋(`the_default_folder_is_moai_and_nothing_is_made_under_the_earlier_one`, `the_earlier_limits_file_is_read_until_the_new_folder_has_one`, `status_lists_the_slots_held_in_the_earlier_folder`)을 써서 옛 이름·호환 없음 때문에 실패하는 것(exit 100)을 먼저 봐요. 실제 HOME·캐시는 건드리지 않고 임시 HOME·임시 `XDG_CACHE_HOME` 만 써요.
3. `src/heavy_slot.rs`: `default_dir` 를 `moai/heavy-slots` 로, `read_limits_file` 이 기본 폴더에서 파일이 없을 때만 옛 폴더의 파일을 같은 규칙으로 읽게, `status` 가 기본 폴더일 때 `earlier_folder` 를 더해요. 쌍둥이 도구는 기본 폴더 한 곳과 머리말만 바꿔요.
4. 문장: `main.rs`·`help.rs` 의 `--dir` 설명, `docs/rust-product-path.md`, `skills/moai/workflows/run.md`(와 매니페스트 해시 한 줄)의 슬롯 폴더 경로. 파이썬 슬롯 폴더 줄은 그대로예요(결정 D-12).
5. 실제 바이너리 출력을 읽는 테스트: `heavy-slots|heavy_slot|heavy-slot-status|earlier_folder` 로 찾아 nextest 명시 `--test` 로 모두 돌려요.

### Plan 4b (개정 7)

1. 찾기: `rg -n 'synthetic-home|synthetic home|moai-cli-git-init|moai-cli-home-|moai-cli-profiles-|moai-cli-gitconfig-'` 로 소스·테스트·문서·도구의 모든 줄을 세고, 이름을 만드는 곳과 읽는 곳을 나눠요(이 개정을 쓸 때 본 곳: `src/migrate.rs` 표식 상수, `src/launcher_init.rs` 잠금, `src/worktree_report.rs` 접두사, 도움말 문장 넷, 문서 한 곳).
2. RED: `tests/migrate_cli.rs` 의 새 테스트 둘(표식 상수의 글자, 새 표식은 열리고 옛 표식만 있으면 새 이름을 말하며 거절)와 `tests/launcher_auto_init.rs` 의 잠금 이름 두 곳·새 테스트 `the_git_init_lock_is_named_moai_and_an_earlier_one_is_left_alone`(임시 TMPDIR 에 옛 이름 잠금 파일을 두고 런처가 그것을 건드리지 않는 것)이 옛 이름 때문에 실패하는 것(exit 100)을 먼저 봐요. N1 테스트 둘은 이미 있는 동작을 고정하는 테스트라 처음부터 통과해요 — 그 대신 `read_limits_file` 의 옛 폴더 읽기를 빼는 뮤턴트가 둘 다 죽이는 것을 보여요.
3. 코드: `SYNTHETIC_MARKER`·`SYNTHETIC_MARKER_TEXT` 와 문장, `init_lock` 의 이름과 설명 주석만 바꿔요. 문장 바깥의 코드는 그대로예요.
4. 도움말 문장 넷과 문서 한 곳의 표식 이름을 맞추고, 도움말 골든·`help_text_names_moai` 가 그대로 초록인지 확인해요.
5. 실제 바이너리 출력을 읽는 테스트: `synthetic|git-init|migrate-|launcher_auto_init` 로 찾아 nextest 명시 `--test` 로 모두 돌려요.
6. 이번 조각이 아닌 것: 파이썬 슬롯 폴더(결정 D-17), 임시 폴더 접두사 읽기(결정 D-17), `.moai-cli` 상태 폴더 줄(조각 7), 권한 참조·키 볼트 유도선(결정 D-8).

## Risks

- clap 이름을 바꾸는 순간 옛 이름만 받는 읽는 쪽이 깨져요 — 그래서 1a 가 1b 보다 먼저예요(위 Order).
- 릴리스 워크플로우의 태그 이름 검사(`moai-adk-v${VERSION}`)는 이 조각에서 바꾸지 않아요 — 결정 6286 의 새 태그는 조각 6 이 바꿔요.
- 스킬 문장은 옛 이름의 설치가 남는 동안에도 참이어야 해요 — fallback 은 계속 `moai-cli` 예요(조각 5 의 호환 링크가 사라질 때 같이 지워요).
- 슬롯 폴더가 갈라지는 동안(옛 빌드는 옛 폴더, 새 빌드는 새 폴더)에는 두 빌드가 서로를 막지 못해 기계가 한도보다 많이 돌 수 있어요. 코드는 이를 막지 못하고 `earlier_folder` 로 보여 주기만 해요 — 그래서 이 조각은 레인이 멈춘 동안 한꺼번에 바꿔요(리더가 예약, 결정 D-11).
- 잠금 이름이 갈라지는 동안(옛 런처는 `moai-cli-git-init-…`, 새 런처는 `moai-git-init-…`)에는 두 런처가 같은 폴더를 동시에 `git init` 해도 서로를 막지 못해요 — 잠금이 없던 때와 같은 상태이고 안전하지 않은 잠금을 받아들이지 않는 규칙은 그대로예요(결정 D-16).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `VERSION_SIGNATURE` 상수는 이름을 담지 않는다(`<ver> (rust)`) | 이름은 clap 이 앞에 붙여요 — 한 줄 바꿈으로 충분해요 | 이름을 상수에 넣기 |
| D-2 | `install.rs` 의 `generation_of` 는 코드를 바꾸지 않는다 | 이미 `^(moai-cli\|moai) [0-9]\S* \(rust\)$` 를 받아요 | 한쪽으로 좁히기 |
| D-3 | settings cleaner 의 런처 경로 `$HOME/.moai-cli/path/moai-cli` 는 조각 5 | 이름 바꿈이 아니라 호환 읽기의 문제예요 | 조각 1 로 당기기 |
| D-4 | 키 볼트 유도선은 불변 | 기존 볼트 항목을 읽을 수 없게 되는 것을 막아요(guarded, 볼트 내용 미개봉) | 유도선 버전 올리기 |
| D-5 | 릴리스 워크플로우의 메모에는 바이너리가 찍은 첫 줄을 그대로 적는다 | 옛·새 이름이 섞여 있는 동안 메모가 거짓이 되지 않아요 | 고정 문구로 되돌리기 |
| D-6 | `bin-names-plan` 의 두 문장(`Preview the moai link beside moai-cli`, `should be a hard link to \`moai-cli\``)은 조각 5 까지 그대로 | 실제 파일이 아직 `bin/moai-cli` 이고(조각 5 가 바꿔요) 문장은 디스크의 그 파일을 말해요; 조각 5 는 어느 쪽이 실제 파일인지가 뒤집히므로 이 명령과 문장을 함께 다시 써요; `tests/help_layout.rs` 의 `ALIAS_FILE` 이 같은 두 문장만 허용해요 | 문장을 중립으로 바꾸기 |
| D-7 | 도움말의 `~/.moai-cli` 상태 폴더 줄(`home-install-*`, `key-*`, `hook-config`, 설정 안내 등)은 조각 7 까지 그대로 | 폴더 이름은 path-root 뒤 조각 7 이 옮기고 문장은 그때 같이 바뀌어요 | 문장만 먼저 바꾸기 |
| D-8 | 파일·폴더·권한 참조·입력 이름을 말하는 문자열은 이번 조각에서 그대로: `update.rs` 의 `PERMISSION`(`moai-cli update confirmation`, 영수증에 남는 참조 문자열), `bundle/verify.rs` 의 `no built moai-cli`(`target/debug/moai-cli` 파일), `update_hooks.rs` 의 `moai-cli hook entries`(옛 훅 항목을 가리켜요), `key_vault.rs` 의 유도선, `state_root.rs` 의 표식 `created_by` | 이름을 바꾸면 기존 영수증·볼트 항목·실제 파일 이름과 어긋나요(조각 5·9) | 해당 조각에서 바꾸기 |
| D-9 | 새 기본 슬롯 폴더는 `${XDG_CACHE_HOME:-~/.cache}/moai/heavy-slots` (선택지: 이 값 / `moai-adk/heavy-slots` / 위쪽 폴더만 두고 끝 이름만 바꾸기). 기본값: `moai/heavy-slots` | 제품 이름이 `moai` 하나이고 캐시의 앱 폴더는 그 이름을 따르는 것이 관례예요; 저장소에는 `.cache/moai/` 를 쓰는 곳이 없어요(`rg '\.cache/moai\b'` 로 확인; 이 기계의 실제 캐시는 열어 보지 않아서 `unverified`) | 한 상수와 두 도구의 기본값을 되돌리기 |
| D-10 | 옛 폴더 호환은 두 가지 읽기뿐: `limits.json` 을 새 폴더에 없을 때 옛 폴더에서 읽기(REQ-13), `heavy-slot-status` 에 옛 폴더의 쥔 슬롯 보이기(REQ-14). 선택지: 호환 없음 / 이 두 읽기(기본값) / 코드가 옛 폴더를 새 폴더로 옮기고 옛 이름에 링크를 남기기 | `limits.json` 은 기계 주인이 손으로 쓴 안전 한도(디스크 여유·스왑)라 새 폴더에서 기본값으로 풀리면 주인이 정한 한도보다 느슨해질 수 있어요; 옮기기와 링크는 사용자 상태를 코드가 바꾸고 `moai-cli` 아래 새로 만드는 일이라 하지 않아요 | 두 읽기를 지우기 |
| D-11 | 옛 폴더를 새 폴더로 옮기는 일은 기계 주인이 쥔 슬롯이 없을 때 `mv ~/.cache/moai-cli/heavy-slots ~/.cache/moai/heavy-slots` 한 번으로 해요; 코드는 옛 폴더에 쓰거나 옮기거나 지우거나 만들지 않아요. 모든 레인을 한꺼번에 바꾸는 시점은 리더가 예약해요 | 쥔 슬롯은 열린 파일의 잠금이라 폴더 이름 바꾸기에 영향이 없지만, 옛 빌드가 옛 경로를 열면 갈라져요; 이름 바꾸기는 되돌릴 수 있어요(`mv` 를 거꾸로) | 옮기지 않기: 새 폴더에서 새로 시작해요(한도는 D-10 읽기가 이어요) |
| D-12 | 이번 조각에 넣지 않는 것: 파이썬 슬롯 폴더(`moai-cli/python-slots`: `tools/self_check.sh`, `tools/local_check.sh`, 그 Python 테스트 `tests/self_check_test.py`·`tests/local_check_test.py`, `docs/ci-rc-ci-spec.md`, 스킬 문장)와 임시 이름(`moai-cli-home-`, `moai-cli-git-init-*.lock`, 표식 `.moai-cli-synthetic-home`) | 파이썬 슬롯은 셸 스크립트와 손으로 못 돌리는 Python 테스트가 같이 움직여야 하고, 임시 이름은 슬롯과 잠금을 나누지 않아 한꺼번에 바꿀 필요가 없어요 | 같은 방식으로 다음 조각(4b)에서 바꾸기 |
| D-13 | Python 쌍둥이 도구는 기본 폴더 한 곳과 머리말만 바꾸고 `limits.json` 호환 읽기는 옮기지 않는다 | 이 도구는 Rust 가 같은 잠금을 쓰는지 확인하는 교차 테스트와 옛 호출자를 위해 남아 있고, Python 도구가 없어질 때 같이 사라져요; 호환 읽기를 거기에도 두면 곧 지울 줄을 더하는 일이에요 | 쌍둥이에도 같은 읽기 더하기 |
| D-14 | 다음 조각은 4b 이고 범위는 N1 테스트 둘, `git init` 잠금 이름, 임시 HOME 표식 이름이에요 (선택지: 4b / 7 추적된 `.moai-cli` 트리 / 5 바이너리·패키지 / 9 영수증·표식 문구와 볼트 유도선 / 11 문서). 기본값: 4b | 4b 는 저장소 안에서만 끝나고 사용자 HOME·링크·설치를 건드리지 않아요. 7 은 읽는 곳이 Python 파일(`tests/github_inventory_*.py`, `tools/legacy_verdicts.py`, `tools/check_ac_commands.py`)과 `reports_move.rs` 에 흩어져 있어 Rust-only 카드와 겹쳐요. 5 는 링크·설치 전환이라 사용자를 기다려요. 9 는 키 볼트 유도선이 guarded 예요(D-4). 11 은 모든 열린 브랜치와 부딪쳐요. 조각 4 감사가 N1 두 테스트를 다음 조각 범위로 정했어요 | 다른 조각을 먼저 하기 |
| D-15 | 임시 HOME 표식은 `.moai-synthetic-home` 와 `moai synthetic home v1` 줄바꿈으로 바꾸고 옛 표식은 받지 않아요 (선택지: 새 것만 / 둘 다 받기 / 그대로) | 표식은 migrate 를 연습하려는 호출자가 직접 만드는 시험용 문턱이에요(결정 6784520717). 설치된 사본이 만들거나 읽는 파일이 아니라서 호환 읽기가 지킬 것이 없어요; 둘 다 받으면 옛 이름이 허용 목록에 남아요. 거절 문장이 새 이름을 말하므로 옛 표식을 손으로 만든 사람도 무엇을 만들지 알아요 | 상수 둘을 되돌리기 |
| D-16 | `git init` 잠금은 `moai-git-init-<16 hex>.lock` 하나만 쓰고 옛 이름 파일은 읽지도 지우지도 않아요 (선택지: 새 것만 / 옛것과 새것 둘 다 잡기 / 그대로) | 잠금은 임시 폴더의 짧은 수명 파일이고 잠금이 없을 때의 동작(`init_lock` 이 `None`)이 이미 정해져 있어요; 둘 다 잡는 일은 새 코드와 새 교착 경로를 더해요(규칙 20). 옛 파일은 크기 0 이라 두어도 해가 없고 지우기는 사용자 상태를 코드가 바꾸는 일이에요 | 이름 한 줄을 되돌리기 |
| D-17 | 그대로 두는 것: 임시 폴더 접두사 `moai-cli-home-`·`moai-cli-profiles-`·`moai-cli-gitconfig-`(`worktree_report.rs` 의 `EMPTY_PREFIXES` 와 그 테스트 둘)과 파이썬 슬롯 폴더 `moai-cli/python-slots`(`tools/self_check.sh`·`tools/local_check.sh`, 그 Python 테스트, `docs/ci-rc-ci-spec.md`) | 접두사는 옛 빌드가 11-06~11-08 에 남긴 빈 폴더를 알아보는 읽기 쪽 이름이라 지금 만드는 코드가 없어요(`docs/worktree-strategy-2026-11-10.md` 29 줄); 바꾸면 남은 것을 못 알아봐요. 파이썬 슬롯은 Python 슬롯 코드와 함께 Rust-only 카드가 걷을 예정이고(`SPEC-20261121-rust-only-inventory` 의 `python-slots` 줄) 손으로 못 돌리는 Python 테스트가 같이 움직여야 해요. 둘 다 조각 11 의 허용 목록에 이유와 함께 적어요 | 접두사에 새 이름을 더하고 옛 이름도 받기 |

## Result

### 조각 1a (읽는 쪽이 두 이름을 받게, 출력 이름은 그대로)

2026-11-23, 브랜치 `spec/20261123-moai-name`, head `6f504e999925b12baa969425fdebabf9c5aed557`(NeXT `0f7711788f` 로 착지), base `2ade55ba2d1821d993de72d836f60bbebb140769`.

- AC-1: met — `cargo nextest run --test bin_names_signature_cli` 의 `a_separate_moai_of_this_build_is_the_same_build_under_either_printed_name` 통과(`moai <sig>`·`moai-cli <sig>` 는 `replace_copy`, 다른 빌드 번호·Go 줄·모르는 이름은 `blocked`). 뮤턴트: 같은 빌드 검사에서 `moai` 이름을 뺐더니 이 테스트가 실패(exit 100).
- AC-2: met — `the_release_check_accepts_both_printed_names_and_nothing_else` 통과(워크플로우 `verify` 단계의 실제 스크립트를 가짜 바이너리로 돌림: 두 이름의 네 조합 통과, 다른 버전·Go 줄·모르는 이름·`(rust)` 없는 줄 실패). 뮤턴트: `case` 에서 `moai` 줄을 뺐더니 실패(exit 100).
- AC-3: met — `the_skill_fallback_sentence_names_both_signatures` 통과. 뮤턴트: 문장을 옛 문구로 되돌렸더니 실패(exit 100). `source-manifest.json` 의 SKILL.md sha256 도 같은 커밋에서 갱신했어요.
- AC-4: met(기존 동작) — `install::generation_tests::the_rust_signature_is_matched_by_shape_under_both_names` 가 `moai check affected` 의 test 단계에서 통과.
- AC-5: not_met — `--version` 첫 줄을 `moai <version> (rust)` 로 바꾸는 clap 이름 한 줄(`src/main.rs`)은 조각 1b 예요. 리더 결정으로 `src/main.rs` 가 이 SPEC 쪽으로 풀렸어요(개정 3 에서 처리).
- AC-6: met — 위 Occurrence table, `moai spec check SPEC-20261121-moai-name` errors 0 (경고 2개는 형식이 모르는 `##` 절 이름).
- 같은 head 영수증: `moai check affected 2ade55ba2d..6f504e9999`(위 head) — 대상 50개, diff_check=0 fmt=0 clippy=0 test=0, 2432개 통과 (앞선 head `a2933f2c6d` 의 첫 실행은 새 테스트가 상태 폴더 글자를 직접 써서 `state_folder_literal_scan` 1개 실패 → 리졸버로 고쳐 재실행).
- 범위 밖으로 둔 것: NeXT 의 빨강 1개 `mode_workflow_steps`(수리 중) — 이번 실행에서는 대상에 들지 않아 관찰되지 않았어요(`not_run`).

### 조각 1b (찍는 이름을 moai 로)

2026-11-23, 브랜치 `spec/20261123-moai-name-print`, base NeXT `0f7711788f7ac8c8e89a64e37141f86f362bdd2d`.

- AC-5: met — `src/main.rs` 의 clap `name` 이 `moai` 가 됐어요. 기대값을 새 이름으로 먼저 고친 세 테스트(`both_names_print_the_same_output`, `every_name_prints_moai_and_the_version_line_is_unchanged`, `every_binary_name_renders_the_same_moai_help`)가 옛 이름을 찍어 빨강(exit 100)인 것을 본 뒤, 이름 한 줄을 바꿔 초록이 됐어요(1a 의 세 테스트 포함 11개 통과). 뮤턴트: `name` 을 `moai-cli` 로 되돌렸더니 세 테스트가 모두 실패(exit 100).
- AC-1~AC-4: 1a 의 결과 그대로예요 — 이름을 바꾼 뒤에도 같은 테스트가 초록이에요(위 11개에 들어 있어요).
- AC-6: met — 이 개정에서 verify 줄을 id 형식으로 고쳤고(N1) `moai spec check SPEC-20261121-moai-name` errors 0.
- 1a 감사 비차단 정리: N1(AC-6 verify id 형식)·N2(낡은 보류 문장과 카드 번호 제거)·N4(위 1a 영수증의 head)·N6(문서 속 번역투 낱말 → 테스트·브랜치) 를 이 문서에서 고쳤어요. N3(약속한 단위 테스트)·N5·N7 은 이번 조각 밖이에요(`not_run`).
- 같은 head 영수증은 카드의 READY 노트에 있어요(이 문서를 담는 head 에서 돈 값이라 해시를 여기에 적지 않아요).
- 이번 조각이 손대지 않은 것(수리 전 기록): `Cargo.toml` 패키지·bin 이름과 `install.sh`(조각 5), 도움말 글과 골든(조각 2), 설치된 `~/.moai`·`~/.moai-cli`·PATH 링크. 관찰: `main` 이 처리기 반환값을 `println!` 으로 찍어 끊긴 파이프에서 exit 101 이 되는 일은 이 조각에서 다루지 않았어요(리더가 정할 일, READY 노트에 적어요).


### 조각 1b 수리 (개정 4)

2026-11-23, 브랜치 `spec/20261121-moai-name-r1`(조각 1b 의 head `8eaf203c102d8a776ae62caf1e9daeaa5a3136cd` 에서 이어요), base NeXT `0f7711788f7ac8c8e89a64e37141f86f362bdd2d`.

- AC-7: met — 현재 head 에서 `cargo nextest run --test install_candidates` 가 실제 바이너리의 `moai <ver> (rust)` 를 옛 이름 `starts_with` 에 걸어 실패(exit 100)하는 것을 먼저 봤고, `moai ` 와 `moai-cli ` 를 모두 받게 고친 뒤 통과했어요. 뮤턴트: 새 단언에서 `moai ` 갈래를 빼면 같은 테스트가 실패(exit 100).
- 실제 바이너리 출력을 읽는 테스트: `moai-cli 4\.|version_line|--version` 으로 찾은 `moai-cli` 바이너리 단위 테스트와 18개 통합 테스트 파일을 nextest 로 모두 돌려 285개 통과(exit 0). 파일 목록은 카드의 REPAIRED 노트에 있어요.
- Python 쪽: `tools/smoke/smoke.py` 정규식은 `^(moai|moai-cli) \S+ \(rust\)$`, `tests/oss_docs_test.py` 의 서명 기대는 `moai `, README 두 곳의 인용은 `moai 2026.11.0-rc.1 (rust)` 로 고쳤어요. 돌리지 않았어요(`not_run`): 손으로 돌리지 않았고, 그 실행의 `moai check affected` 에는 Python 단계가 없었어요. 원격 CI 의 첫 실행이 확인이에요(개정 5 에서 이 문장을 바로잡았어요).
- N5·N6: 감사 노트에는 이름만 있고 본문이 없어서, 이 SPEC 에서 읽히는 어긋남(AC-7·REQ-8·범위·계획에 수리 내용이 없던 것)을 개정 4 로 바로잡았어요. 리더가 가리킨 다른 문구가 있으면 알려 주세요.
- 이번 수리가 하지 않은 것: `check affected` 가 `src/main.rs` 변경에 바이너리 출력을 고정한 테스트를 고르게 하는 일(반복 후보, 리더 기록)과 위 1b 의 관찰.

### 조각 2 (거절·경고·안내 문장, 개정 5)

2026-11-23, 브랜치 `spec/moai-name-20261121`, base NeXT `568bc8956c41be8cae7bbcd518e37fcf5239d5c3`.

- 먼저 본 것: 두 바이너리의 `-h`·`--help`·`help --all` 과 모든 명령의 `--help`(최상위 237개 단어, `bundle verify` 등 일부 하위 명령 포함)에서 명령 이름으로 쓰인 `moai-cli` 는 `bin-names-plan` 의 파일 문장과 `bundle verify` 의 플래그 `--moai-cli`(이름 자체가 인터페이스, 조각 5 이후) 뿐이에요. 도움말 골든(`tests/fixtures/help/*.txt` 8개)에 남은 `moai-cli` 는 상태 폴더 경로(조각 7)와 `bin-names-plan` 파일 문장(결정 D-6) 뿐이라 골든 파일 내용은 바꾸지 않았어요.
- AC-8: met — 문장 11곳(Go 시대 스킬 안내 `update.rs`, 되돌리기 명령 `update/switch.rs` 2곳, 상태 이전 경고 `go_state.rs` 2곳, 키 보관함 안내 `session.rs`·`launcher_terminal.rs`, 거절 `key_vault.rs`, 첫 실행 purpose `launcher_init.rs`, 더 새로운 표식 `state_root.rs`, 설정 백업 머리말 `claude.rs`)을 `moai` 로 바꿨어요. RED: 기대값을 먼저 새 문장으로 고친 테스트와 새 `tests/help_text_names_moai.rs` 가 옛 문장 때문에 7개 실패(exit 100, `target/moai-name-help/red.log`), 바꾼 뒤 초록 52/52(exit 0).
- AC-9: met — `the_help_goldens_keep_the_old_name_only_for_folders_and_the_alias_file` 와 기존 `moai_cli_help_names_no_command_moai_cli` 통과.
- 뮤턴트: 소스 문장 4곳(`session.rs`·`update/switch.rs`·`update.rs`·`key_vault.rs`)과 골든 한 줄을 옛 이름으로 되돌렸더니 8개 테스트가 실패(exit 100): 문장 표 테스트, 골든 guard, `skill_name_collision` 2개, `key_vault` 단위 테스트, `update_switch_layer`·`update_switch_cli` 3개. 커밋 뒤에 넣었고 `git checkout` 으로 되돌린 뒤 `git status` 가 깨끗한 것을 확인했어요.
- 실제 바이너리 출력을 읽는 테스트: 바꾼 문장을 읽는 곳(`update_*`, `go_state_*`, `key_*`, `launcher_*`, `state_root_cli`, `claude_*`, `session_*`, `home_install_*`, `project_update_*`)과 1b 수리 때 찾은 버전 줄을 읽는 곳을 합친 83개 통합 테스트 대상을 nextest 명시 `--test` 로 heavy-slot 4 에서 돌려 1256개 통과(exit 0, `target/moai-name-help/real-output.log`).
- 결정 D-6~D-8: `bin-names-plan` 파일 문장, `~/.moai-cli` 상태 폴더 줄, 권한 참조·파일 이름 문자열은 해당 조각까지 그대로예요(위 Decisions).
- 1b 감사 비차단 N-b(낡은 순서 문장과 읽는 쪽 표)·N-c(AC-7 의 없는 Python 단계)는 이 개정에서 고쳤어요. 같은 head 영수증은 카드의 READY 노트에 있어요.
- 이번 조각이 하지 않은 것: 스킬·템플릿 문장(조각 11), `bundle verify` 의 `--moai-cli` 플래그와 `target/debug/moai-cli` 설명(조각 5), 환경 변수(조각 3), 캐시 폴더 이름(조각 4). `README` 의 옛 문장(1b 감사 N-a)은 조각 11 이에요.

### 조각 4 (heavy-slot 캐시 폴더 이름, 개정 6)

2026-11-24, 브랜치 `spec/20261121-moai-name-c4`, base NeXT `1a8a7357a8a8004923bd3d0869627ca8296978ae`, 구현 커밋 `a3c7d8911630bd0eb39e9ff53af71c7beb0b50be`. 리더가 개정 6 의 결정 D-9~D-13 을 카드 노트 7594 에 대한 답으로 기본값 그대로 받았어요(승인 기록은 그 노트예요; 이 문서의 status 는 draft 그대로 두었어요).

- AC-10: met — RED: 기대 폴더를 `moai/heavy-slots` 로 고친 `without_dir_both_tools_use_the_same_default_folder` 와 새 `the_default_folder_is_moai_and_nothing_is_made_under_the_earlier_one` 이 옛 폴더를 풀어 실패(exit 100, `target/moai-name-cache/red.log`), 바꾼 뒤 통과. 세 경우(캐시 변수 있음·빈 값·없음) 모두 Rust 와 Python 쌍둥이가 같은 새 폴더에서 번갈아 돌고 `<cache>/moai-cli` 는 생기지 않아요. 뮤턴트: 기본 폴더 이름을 `moai-cli` 로 되돌렸더니 두 테스트가 실패(exit 100).
- AC-11: met — `the_earlier_limits_file_is_read_until_the_new_folder_has_one` 통과: 옛 폴더만 `limits.json` 을 가지면 `build_jobs_cap.slots` 가 그 값(1), 새 폴더에 생기면 새 값(2), `--dir` 로 다른 폴더를 주면 기본 3, 옛 폴더의 바이트·목록은 읽은 뒤에도 그대로, 규칙을 어긴 옛 파일(`{"bogus":1}`)은 exit 125 이고 명령은 돌지 않아요. 뮤턴트: 옛 파일 읽기를 `Ok(None)` 으로 바꾸면 실패(exit 100).
- AC-12: met — `status_lists_the_slots_held_in_the_earlier_folder` 통과: 옛 폴더에 쥔 슬롯(임시 폴더의 `old-build`)이 `earlier_folder.held` 에 보이고 새 폴더의 `held` 는 비어 있고, 옛 폴더가 없거나 `--dir` 이면 키가 없고, 놓은 뒤에는 빈 목록이에요. 뮤턴트: `earlier_folder` 채우기를 빼면 실패(exit 100; 위 AC-11 뮤턴트와 한 번에 넣어 두 테스트가 각각 실패하는 것을 봤어요).
- 실제 바이너리 출력을 읽는 테스트: `heavy-slots|heavy_slot|heavy-slot-status|earlier_folder|heavy-slot-acquire|heavy_check_slot` 로 찾은 테스트 파일 17개와 run.md·매니페스트·`rust-product-path.md` 를 읽는 테스트 7개, `--lib` 를 nextest 명시 `--test` 로 heavy-slot 4 에서 돌려 2167개 통과(exit 0, `target/moai-name-cache/real-output.log`).
- 실제 HOME·캐시: 건드리지 않았어요. 모든 테스트는 임시 HOME 과 임시 `XDG_CACHE_HOME` 만 써요. 옛 폴더를 옮기는 `mv`(결정 D-11)는 하지 않았고 사용자 몫이에요.
- Python 쪽: `tools/heavy_check_slot.py` 의 기본 폴더 한 곳과 머리말만 고쳤어요. 돌리지 않았어요(`not_run`): 손으로도 `moai check affected` 의 Python 단계로도; Rust 교차 테스트가 그 도구를 자식 프로세스로 돌려 같은 폴더를 쓰는 것을 확인해요.
- 이번 조각이 하지 않은 것(결정 D-12): 파이썬 슬롯 폴더 `moai-cli/python-slots`(`tools/self_check.sh`·`tools/local_check.sh`·그 Python 테스트·`docs/ci-rc-ci-spec.md`)와 임시 이름 `moai-cli-home-`·`moai-cli-git-init-*.lock`·표식 `.moai-cli-synthetic-home` → 조각 4b. 스킬 문장의 나머지(조각 11).
- 같은 head 영수증은 카드의 READY 노트에 있어요(이 문서를 담는 head 에서 돈 값이라 해시를 여기에 적지 않아요).

### 조각 4b (임시 이름과 N1 테스트, 개정 7)

2026-11-23, base NeXT `ec6233e7b4f9ec3c35b1f8f50918c21b25181890`, 구현 커밋 `e3ab2165e9cfb14cbd19a2efa9688a859e3e9a90`. 리더가 개정 7 의 결정 D-14~D-17 을 카드 노트 7787 에 기본값 그대로 받았어요(승인 기록은 그 노트예요; 리더 판단이고 사용자 직접 선택이 아니에요; 이 문서의 status 는 draft 그대로예요). 브랜치는 SPEC 우선 훅이 처음 이름을 막아 `spec/20261121-moai-name` 으로 바꿨어요(예약 기록의 브랜치 칸에는 처음 이름이 남아요; 커밋 전이라 잃은 것은 없어요).

- AC-13: met — 이 두 테스트는 처음 새 파일 `tests/temp_names_moai.rs` 에서 RED·GREEN 을 봤고(`target/moai-name-4b/red.log`·`green2.log`), `moai check affected` 가 그 파일을 테스트 영역 표(`area_table_maps_every_real_test_file`)와 상태 폴더 글자 허용 목록(`the_repository_tests_stay_within_the_state_folder_baseline`)이 거절해서 `tests/migrate_cli.rs` 로 옮겼어요(리더 노트 7800 의 A; 두 표는 건드리지 않았어요). 옮긴 뒤 `migrate_cli`·`test_area_layout`·`state_folder_literal_scan` 34개가 통과했고(`green3.log`) 표식 뮤턴트 둘을 옮긴 테스트로 다시 돌려 둘 다 exit 100 이었어요(`mutants2.out`). 처음 RED: 표식 상수를 글자 그대로 `.moai-synthetic-home`·`moai synthetic home v1` 줄바꿈과 견준 `the_synthetic_home_marker_is_moai_and_the_earlier_marker_is_refused` 와 도움말 테스트 `migrate_help_names_the_moai_marker_only` 가 옛 이름 때문에 실패(exit 100, `target/moai-name-4b/red.log`), 바꾼 뒤 통과(`green2.log`). 실제 바이너리가 임시 HOME·임시 TMPDIR 에서: 새 표식이 있는 HOME 은 `migrate-plan` 이 열고(writes 0, 트리 digest 그대로), 옛 표식 `.moai-cli-synthetic-home`(+옛 문구)만 있는 HOME 은 `.moai-synthetic-home marker missing` 으로 거절하고 HOME 트리 digest 가 그대로이고, 이름만 새것·문구만 새것인 두 경우도 거절해요. 뮤턴트: 상수를 옛 이름으로 되돌리면 실패(exit 100); 거절 분기(`if !ok`)를 `if false` 로 지우면 옛 표식 HOME 이 열려 같은 테스트가 실패(exit 100) — 안전 문턱을 지우는 뮤턴트가 죽는 것을 봤어요. 도움말 네 명령(`migrate-plan`·`-apply`·`-recovery-plan`·`-recover`)의 `--help` 가 새 표식만 말해요. 문서 `docs/rust-product-path.md` 의 한 문장은 고쳤지만 저장소 파일을 읽는 테스트를 두지 않아 고정하는 테스트가 없어요(`not_run`, 눈으로 확인).
- AC-14: met — 새 `the_git_init_lock_is_named_moai_and_an_earlier_one_is_left_alone` 가 옛 코드에서 임시 폴더에 `moai-git-init-<hex>.lock` 이 없어 실패(exit 100)하고, 바꾼 뒤 통과: 새 이름 파일이 생기고 옛 이름 파일은 바이트·길이·mtime 이 그대로예요. 기존 두 테스트(심볼릭 링크, 다른 계정 소유 잠금)는 새 이름을 기대하도록 고쳤고 통과해요 — 단 이 둘은 옛 코드에서도 통과해서 이름을 고정하지는 못하고, 이름은 새 테스트가 고정해요. 뮤턴트: 잠금 이름을 옛 이름으로 되돌리면 새 테스트 실패(exit 100).
- AC-15: met — `the_earlier_limits_apply_while_the_new_folder_exists_without_limits`(새 폴더는 있고 `slot-0` 만 있어도 `build_jobs_cap.slots` 가 옛 값 1, 옛 폴더 바이트 그대로)와 `acquire_applies_the_earlier_limits_file_to_the_run`(옛 `max` 1 이면 두 번째 실행이 `waiting 1 of 1 for 1 slot(s)` 로 첫째가 놓을 때까지 기다림)이 이미 있던 동작이라 처음부터 통과했어요. 뮤턴트: `read_limits_file` 의 옛 폴더 읽기를 `Ok(None)` 으로 바꾸면 둘 다 실패(exit 100; 둘째는 대기가 안 생겨 hang guard 로 끝나요).
- REQ-15~17 의 실제 HOME·임시 폴더: 건드리지 않았어요. 모든 자식은 임시 HOME·임시 `TMPDIR`·임시 `XDG_CACHE_HOME` 만 써요.
- 그대로 둔 것(결정 D-17): `src/worktree_report.rs` 의 `moai-cli-home-` 등 접두사와 그 테스트 둘, 파이썬 슬롯 폴더. 더해서 `src/lib.rs` 의 상태 폴더 글자 스캔 테스트 입력(`const S: &str = ".moai-cli-synthetic-home";` 와 그 위 설명 한 줄)은 스캐너가 다른 이름을 세지 않는 것을 보이는 입력이라 그대로예요(그 파일은 이 조각의 쓰기 범위 밖이에요); 조각 11 허용 목록에 이유와 함께 적어요.
- 이번 조각이 하지 않은 것: 조각 5·7·9·11, `moai-cli/python-slots`, 임시 폴더 접두사 읽기.
- 같은 head 영수증은 카드의 READY 노트에 있어요(이 문서를 담는 head 에서 돈 값이라 해시를 여기에 적지 않아요).
