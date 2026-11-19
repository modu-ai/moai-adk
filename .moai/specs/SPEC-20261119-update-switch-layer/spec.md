---
id: SPEC-20261119-update-switch-layer
title: "moai update previews the name-switch script v2 and applies it only after a pick (C-33 S3)"
status: draft
revision: 1
created: 2026-11-19
---

# moai update previews the name-switch script v2 and applies it only after a pick (C-33 S3)

## Intent

카드 C-33 의 조각 S3(리더 결정 노트 2866). S2a 가 만든 전환 스크립트 v2 를 지금은 사람이 손으로 복사해야 해요. 이 조각은 `moai update` 가 그 단계를 **미리보기에 한 줄로 보여 주고, 사용자가 고른 뒤에만** 적용하게 해요. 실제 HOME 에는 쓰지 않아요(시험은 임시 HOME 만 써요). 실제 적용은 S4(guarded, 사용자 선택).

코드로 확인한 사실(tip d25ecb1ba1, `git grep`/`git show` 로 읽음):
1. 설치기 `install/install.sh` 의 `link_commands` 는 `<state home>/path/moai-cli`·`moai` **심볼릭 링크**만 만들어요. 이미 있는 것이 링크가 아니면 "is not a link made by this installer; left as it is" 하고 건드리지 않아요(install.sh:549-550). `src/bin_names.rs` 의 `moai` 는 버전 폴더 안 `bin/moai` 하드 링크라서 `path/moai` 와 무관해요. 그 밖에 `path/` 에 쓰는 코드는 `home_install_apply` 의 링크 다시 가리키기(`make_link`)뿐이에요.
2. 그래서 지금 `~/.moai-cli/path/moai` 의 1747 B 일반 파일 스크립트는 **설치기도 update 도 아닌 사람이 놓은 거예요**(스크립트 머리말: "installed 2026-11-10 by the ADK leader on the user's instruction"). 아무 코드도 그 파일을 고치지 않아요.
3. `moai update` 에는 이미 "자기 토큰으로만 적용되는 부분"의 틀이 있어요: `update_home`(`--confirm-home <token>`), `update_cleanup`(`--confirm-cleanup`). 미리보기에 키가 붙고, 토큰은 내용 다이제스트라 시계에 안 묶이고, `plan_digest` 밖이고, 터미널의 yes 로는 적용되지 않아요. 이 조각은 같은 틀로 `switch` 부분을 더해요.

해결: `src/update/switch.rs` 가 `path/moai` 가 **알려진 v1 스크립트(sha256 고정)** 인지 읽어서 `replace` 를 제안하고, `--confirm-switch <token>` 으로만 적용해요. 적용은 모든 조건을 먼저 확인하고(감사 N1 — 쓰기 전에 전부), 이전 파일을 `moai.before-<UTC>` 로 보존하고, 되돌리기는 v1 과 바이트가 같은 저장본만 옮겨요(감사 N2). `~/go/bin/moai` 는 건드리지 않고 안내만 해요.

## Non-goals

- 실제 HOME 에 적용(S4, guarded). 이 조각의 시험은 임시 HOME 만 써요.
- `~/go/bin/moai`(같은 스크립트) 쓰기 — update 는 읽고 안내만 해요.
- 설치기(`install.sh`)가 스크립트를 놓게 하기, Go 진입점 비활성화, 스크립트 v2 내용 바꾸기(S2a 소관).
- 다른 프로젝트의 `.codex/hooks.json`·상태줄 스크립트 수정 — 안내 문구만 더해요.
- Windows(스크립트는 `sh`): 이 부분은 unix 에서만 있어요.

## Requirements

- REQ-1: When `moai update` previews and the `path/moai` file of the HOME state folder is a regular file whose bytes are the known v1 switch script, the preview shall carry a `switch` part that names the file, the saved-copy name `moai.before-<UTC>`, the sha256 of the v1 and v2 scripts, one summary line and the confirming flag, while nothing is written. · failure: none
- REQ-2: When `path/moai` is absent, is a link, or already holds the v2 script, the preview shall carry no `switch` part, so a HOME without the old script gets exactly the preview it had before. · failure: none
- REQ-3: IF `path/moai` is a regular file whose bytes are neither the v1 nor the v2 script, THEN the preview shall show a `switch` part with action `left`, the reason and no token, and that file stays as it is in every run. · failure: none
- REQ-4: When `moai update --confirm-switch <token>` is given the token of the freshly recomputed preview, the update shall save the v1 bytes as `moai.before-<UTC>` beside the file, replace `path/moai` with the v2 script by one rename at mode 755, and read the file back to confirm it equals v2. · failure: none
- REQ-5: IF the token is not the recomputed preview's own, or `path/moai` changed after the preview, THEN the update shall write nothing and tell the person to preview again. · failure: none
- REQ-6: IF a check fails before the first write (the folder is a real folder and not a link, the file is a regular file equal to v1, the saved-copy name and the temporary name are free; all of them run before any write), or a step fails after it, THEN the apply shall leave `path/moai` whole, equal to v1 or to v2 and never partial, and name the saved copy if one exists. · failure: none
- REQ-7: When `moai-cli update --revert-switch <stamp>` names a saved copy, the revert shall move it back over `path/moai` by one rename only when the stamp is well formed, the saved copy is byte-equal to v1 and `path/moai` equals v2; in any other case nothing changes and the result says why. · failure: none
- REQ-8: The apply result shall name the stamp, the saved copy, the revert command `moai-cli update --revert-switch <stamp>` and a plain `mv` line that does the same without MoAI. · failure: none
- REQ-9: The `switch` part shall report whether `go/bin/moai` of the HOME exists and holds the v1 bytes, with guidance text for it and for the Go-generation projects that still call `moai hook` and `moai statusline`, and no run writes `go/bin/moai`. · failure: none
- REQ-10: The flag `--confirm-switch` shall apply only this part (after the layer when `--confirm <digest>` is given, alone otherwise), while a yes at the terminal question applies nothing of it, and `--revert-switch` is refused beside any confirming flag and runs without staging the package. · failure: none
- REQ-11: The v2 script text built into the program shall have the sha256 `36dd242bf1337efa9b4759ba7166fb6adea04427800334a374027cc246eb26e4` (the S2a preview `preview/moai.v2`), and the known v1 script the sha256 `3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac`. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-9): the plan offers `replace` for the known v1 script with the digests, the one-line summary and the flag; it shows no part for an absent file, a link or v2; it shows `left` with no token for an unknown file; it reports `go/bin/moai` and never writes it.
  verify: `cargo test --locked --offline --test update_switch_layer plan_offers_v2_for_the_known_v1_script`
  verify: `cargo test --locked --offline --test update_switch_layer plan_has_no_part_for_a_link_an_absent_file_or_v2`
  verify: `cargo test --locked --offline --test update_switch_layer plan_leaves_an_unknown_file_alone`
  verify: `cargo test --locked --offline --test update_switch_layer go_bin_is_reported_and_never_written`
- AC-2 (REQ-4, REQ-8): apply saves v1 as `moai.before-<UTC>` (bytes equal, mode kept), installs v2 with mode 755 by rename, reads it back, and names the stamp, the saved copy and both revert lines.
  verify: `cargo test --locked --offline --test update_switch_layer apply_saves_v1_and_installs_v2`
- AC-3 (REQ-5, REQ-6): a wrong, foreign or stale token writes nothing; every check runs before the first write; a step that fails leaves a whole file.
  verify: `cargo test --locked --offline --test update_switch_layer apply_refuses_a_wrong_or_stale_token`
  verify: `cargo test --locked --offline --test update_switch_layer apply_checks_everything_before_the_first_write`
  verify: `cargo test --locked --offline --test update_switch_layer a_failed_step_leaves_a_whole_file`
- AC-4 (REQ-7): revert puts v1 back from a saved copy equal to v1 and refuses everything else, without staging the package.
  verify: `cargo test --locked --offline --test update_switch_layer revert_puts_v1_back_and_refuses_anything_else`
- AC-5 (REQ-11): the built-in v2 and the known v1 have the pinned digests.
  verify: `cargo test --locked --offline --test update_switch_layer the_pinned_digests_are_the_s2a_scripts`
- AC-6 (REQ-10): the flags combine as written, in a temporary HOME with the installed layout: preview, `--confirm-switch`, `--confirm` plus `--confirm-switch`, a terminal yes, `--revert-switch`, and the refusals.
  verify: `cargo test --locked --offline --test update_switch_layer revert_switch_runs_alone_and_is_refused_beside_a_confirming_flag`
  verify: `cargo test --locked --offline --test update_switch_cli`
- AC-7 (REQ-4, REQ-7): the apply and revert tests write no file outside their temporary folders (the HOME is a temporary folder); the scope's `forbidden` line is checked by the repository's home isolation scan.
  verify: `cargo test --locked --offline --test home_isolation_scan`
- AC-8: neighbours (targeted; the full run is cited from the PR's CI run).
  verify: `cargo test --locked --offline --test update_home_layer_cli`
  verify: `cargo test --locked --offline --test update_tux_golden`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261119-update-switch-layer/**`, `src/update/switch.rs`, `src/update.rs`, `src/update_tux.rs`, `install/moai-switch.sh`, `tests/update_switch_layer.rs`; and, once card C-75 has merged, `src/main.rs`, `src/help.rs`, `tests/update_switch_cli.rs`, `docs/rust-product-path.md`
- read-only: `install/install.sh`, `src/bin_names.rs`, `src/home_root.rs`, `src/update_home.rs`, `.moai/specs/SPEC-20261119-switch-script-v2/**`
- forbidden: any write in the real HOME, in the Go build's folders, in the Orca profile folders or in another project; deleting anything the person owns (the program's own temporary file may be removed)

## Plan

1. SPEC (this file), then RED: `tests/update_switch_layer.rs` against a skeleton `src/update/switch.rs` that offers nothing.
2. GREEN: `install/moai-switch.sh` (byte copy of S2a's `preview/moai.v2`), the plan, apply and revert in `src/update/switch.rs`; wire the part into `src/update.rs` (`update_command_switch`; `update_command` keeps its signature and passes none) and the terminal screen in `src/update_tux.rs`.
3. After card C-75 merges: merge the new tip, add `--confirm-switch` and `--revert-switch` in `src/main.rs` and `src/help.rs`, `tests/update_switch_cli.rs`, the row in `docs/rust-product-path.md`.
4. fmt, clippy, mutants (at most 5), trailer-check, self-audit in a different context, `LANE-READY-INTEGRATION` only after step 3 (a preview that names a flag the program does not have must not land alone).

## Risks

- The branch is stacked on S2a (`spec/20261119-switch-script-v2`, head 95e02aaaeb); the leader bundles S2a first. The built-in v2 is a copy, pinned by digest, not a read of the S2a SPEC folder (SPEC folders may be archived).
- `tests/update_switch_layer.rs` reads the v1 script from the S2a preview folder at build time (`include_bytes!`); if that SPEC folder is ever archived, the test needs the 1747 B file kept somewhere else. The program itself reads nothing from a SPEC folder.
- A person who edited the old script by hand gets `left` and no token: the update never overwrites a file it did not recognise.
- The installer keeps leaving a regular `moai` file alone (install.sh:549-550); after the apply `moai` is still a regular file, so the next install does not replace it.
- `~/go/bin/moai` keeps the old script after the apply. A person who runs the Go build's `moai` from that folder keeps the v1 routing until they copy the new file there themselves; the guidance text says so.
- The apply changes where `moai statusline` goes for every Go-generation project whose `.moai/status_line.sh` runs it (S2a Result, decision 3). The guidance text names this; the project files are not touched.
- A crash between the saved-copy write and the rename leaves a harmless `moai.before-<UTC>` beside a still-v1 `moai`; the next apply uses a new stamp, and the stamp check refuses only a name that exists.
- The revert moves the saved copy, so a second revert with the same stamp finds nothing and says so.

## Result

### S3a·S3b (이 가지의 지금 상태)

| 부분 | 어디 | 상태 |
| --- | --- | --- |
| 계획·적용·되돌리기 | `src/update/switch.rs`(update.rs 의 자식 모듈 — `src/lib.rs` 를 안 건드려요) | 구현·시험 완료 |
| 내장 v2 스크립트 | `install/moai-switch.sh`(S2a `preview/moai.v2` 와 `cmp` 같음, sha256 36dd242b…) | 완료 |
| 미리보기·적용·화면 연결 | `src/update.rs`(`update_command_switch`; `update_command` 는 시그니처 그대로) · `src/update_tux.rs`(행·다음 단계·결과 줄) | 완료 |
| `--confirm-switch`·`--revert-switch` 플래그 | `src/main.rs`·`src/help.rs`·`docs/rust-product-path.md` | S3b 구현·시험 완료 |
| CLI 시험(`update_switch_cli`) | `tests/update_switch_cli.rs`(7개) · 스캔 등재 `tools/repo_scans.json`(skill-package-agents) | S3b 완료 |

### S3b 검사 (head `69a30927b7`, 전용 빌드 폴더 `moai-adk-33c`·무거운 작업 칸, 2026-11-19)

- AC-6: `cargo test --locked --offline --test update_switch_cli` 7/7, `update_switch_layer` 11/11(미리보기 → `--confirm-switch` → `--confirm` 와 함께 → 터미널 yes 는 스위치를 확인하지 않음 — 이 줄은 pty 시험이 아니라 아래 "감사 수리" 의 `layer_asks` 단위 시험이 고정해요 → `--revert-switch` → 거절들).
- AC-7: `home_isolation_scan` 통과(리더의 여섯 검사 5a 안), 시험은 임시 HOME 에서만 돌아요(실제 HOME 쓰기 0).
- 스캔 등재(RED→GREEN): 등재 전 `python3 -m unittest tests.repo_scans_test` 가 `['tests/update_switch_cli.rs'] != []` 로 실패, 등재 뒤 25개 OK.
- 이웃(AC-8): `update_home_layer_cli` 11 · `update_tux_golden` 14 · `update_route` · `update_cli` · `update_command_cli` · `help_layout` 통과, `--lib` 전체 1490 통과.
- 리더의 여섯 검사: (1) `git diff --check` 0 · (2) `cargo fmt --check` 0 · (3) `rustfmt --check --edition 2021 src/ops/*.rs` 0 · (4) `cargo clippy --locked --offline --all-targets -- -D warnings` 0 · (5) cargo test 5개 대상과 `python3 -m unittest tests.repo_scans_test tests.source_manifest_test tests.local_check_test` 0 · (6) `moai-cli trailer-check` — 내 커밋 범위(`v2026.11.0..HEAD`, 8개) 0. 원격 tip 기준 범위(109개)는 로컬 가지에 있는 남의 커밋 `d2966629e5`(trailer 없음) 때문에 1 이에요.
- 돌리지 않은 것: 저장소 전체 `cargo test`(PR CI 가 인용), macOS 이외, 실제 HOME 적용(S4, guarded — 사용자가 고르는 일).

### 감사 수리 (자체 감사 head `d99930df2a` 가 FAIL — B1 1건, N5건; 같은 감사 에이전트가 재감사)

- B1(막음): 터미널에서 `moai update --revert-switch` 가 되돌리기에 성공해도 화면이 "did not finish: reverted" 와 "up to date" 를 그렸어요. `src/update_tux.rs` 에 `revert_switch` 전용 한 줄 화면(`reverted_switch`, 영어·한국어)을 더했어요. 시험 `the_revert_switch_result_is_drawn_as_a_success_line`(`cargo test --locked --offline --test update_switch_cli` 8/8)이 한 줄·성공 문구·실패 문구 없음을 둘 다 언어로 확인해요. 변이: 분기 삭제(R1)와 상태 비교 깨기(R2)를 그 시험이 잡았어요.
- N2: 터미널 yes 가 스위치를 확인하지 않는다는 말이 시험이 아니라 코드 읽기였어요. 연결 식 `asks(cleanup, home_token.or(switch_token), writes)` 를 함수 `layer_asks` 로 뽑고 단위 시험(`update::user_layer_tests::the_summary_shows_cleanup_only_when_something_is_archived`)이 switch·home·cleanup 토큰 각각을 확인해요. 변이: `home.or(switch)` 를 `home` 으로 바꾸기(R3)를 그 시험이 잡았어요. pty 시험은 저장소에 없어요(시험 도우미가 없어서 만들지 않았어요).
- N3(비차단, 고치지 않음): 저장본 `create_new` 가 쓰기 중간에 실패하면 짧은 `moai.before-<stamp>` 가 남을 수 있어요(디스크가 가득 찰 때). `moai` 는 그대로이고 되돌리기는 v1 과 바이트가 다른 저장본을 거부해서 해가 없어요. 재현할 방법이 없어 시험 없는 코드는 넣지 않았어요.
- N5(범위 밖): `skills/moai/workflows/update.md` 5행이 새 플래그 두 개를 적지 않아요. 이 SPEC 의 쓰기 범위에 `skills/` 가 없고 다른 카드가 쥘 수 있는 파일이라 리더에게 알려요(별도 조각).
- N4(문구): 적용 뒤 화면 문구는 감사가 "정상" 으로 봤어요. N1(기준 가지가 감사 도중 움직임): `d2966629e5` 는 리더 공지대로 알려진 예외예요.

### 이 컴퓨터에서 읽은 사실 (읽기 전용, 2026-11-19, 쓰기 0)

- `shasum -a 256 ~/.moai-cli/path/moai ~/go/bin/moai` → 둘 다 `3c3446a3…6658ac`(= `V1_SHA256`), 둘 다 1747 B, 모드 `-rwxr-xr-x`. 그래서 이 컴퓨터에서 `moai update` 미리보기는 `replace` 를 제안하고, `go/bin/moai` 는 `same_as_v1: true` 로 안내만 해요.
- `~/.moai` 가 있지만 Rust 표지(`~/.moai/moai.json`·`moai-cli.json`)는 없어서 상태 폴더는 `~/.moai-cli`(`home_root::path`)예요 → 대상 파일은 `~/.moai-cli/path/moai`.
- 코드로 확인한 사실은 Intent 의 1–3(설치기·`update` 는 `path/moai` 를 쓰지 않아요).

### 검사 (이 레인, 전용 빌드 폴더 `moai-adk-33c`·무거운 작업 칸, 2026-11-19)

- RED(뼈대가 아무것도 제안하지 않을 때): `cargo test --locked --offline --test update_switch_layer` 10개 중 8개 실패, 2개 통과(`target/lane/s3red1.log`). GREEN: 10/10(`s3green1.log`), 되돌리기 시험을 더한 뒤 11/11(`s3t2.log`).
- 연결 뒤(`target/lane/v1-*.log`): `cargo fmt --check` 0 · `update_switch_layer` 10/10 · `update_home_layer_cli` 11/11 · `update_tux_golden` 14/14 · `cargo clippy --locked --offline --all-targets -- -D warnings` 0.
- 변이 5개(커밋된 `src/update/switch.rs` 에서 하나씩, `target/lane/c33-mutants.log`): S1 저장본 이름 검사 삭제 → `apply_checks_everything_before_the_first_write`, S2 되돌리기의 v1 검사 삭제 → `revert_puts_v1_back_and_refuses_anything_else`, S3 토큰 무시 → `apply_refuses_a_wrong_or_stale_token`, S4 저장본 모드 600 → `apply_saves_v1_and_installs_v2`, S5 rename 실패 뒤 임시 파일 남김 → `a_failed_step_leaves_a_whole_file` — 5/5 처음에 잡혔고 매번 `git restore` 로 되돌렸어요(끝에 `git status` 0줄).
- S3a 때 돌리지 않았던 것(S3b 에서 아래 "S3b 검사" 로 돌렸어요): `--confirm-switch`·`--revert-switch` 를 실제 명령줄로, `cargo test --lib` 전체와 `home_isolation_scan`(S3b 와 함께), 자체 감사(S3b 뒤), 실제 HOME 적용(S4, guarded), macOS 이외(`#[cfg(unix)]`; Windows 에서는 부분이 없어요).

### 안내 문구 (미리보기 `switch.notes`·`switch.go_bin.note`, 시험이 `moai hook`·`moai statusline` 낱말을 고정)

- Go 세대가 만든 프로젝트는 훅 파일에서 `moai hook ...` 을 불러요. 전환 뒤에도 `moai hook` 은 Go 빌드(moai.old)로 가요. 그 프로젝트의 파일은 바꾸지 않아요.
- 그 프로젝트의 `.moai/status_line.sh` 는 `moai statusline` 을 실행해요. 전환 뒤에는 Go 렌더가 아니라 Rust 3줄 렌더가 나와요. 그 프로젝트의 파일은 바꾸지 않아요.
- `go/bin/moai` 가 같은 옛 스크립트면: update 는 건드리지 않아요. 미리보기를 적용한 뒤 그 이름도 따라가길 원하면 새 스크립트를 직접 복사해요.
- 이 컴퓨터의 숫자(Go 훅 프로젝트 5곳, 상태줄 스크립트 30곳)는 S2a SPEC 의 Result 에 있고, 제품 문구에는 넣지 않았어요(컴퓨터마다 달라요).

### 리더 결정이 필요한 것 (되돌릴 수 있어요, 범주 none)

1. 이 가지는 S2a(`spec/20261119-switch-script-v2`) 위에 쌓였어요. 병합 순서: S2a → 이 가지. 내장 v2 는 S2a 미리보기의 복사본이고 sha256 으로 고정했어요.
2. READY 는 S3b(플래그) 뒤에 한 번만 올려요: 플래그가 이 가지에 들어 있어서 미리보기가 안내하는 플래그가 모두 있어요.
3. `--revert-switch` 의 이름·인자(`<stamp>`)와 되돌리기 명령을 `moai-cli` 로 쓴 것(스크립트가 깨졌을 때도 도는 이름): 다른 이름을 원하면 `switch.rs` 두 곳과 시험 한 곳.
