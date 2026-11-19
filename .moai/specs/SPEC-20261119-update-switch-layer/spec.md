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
- AC-6 (REQ-10): the flags combine as written, in a temporary HOME with the installed layout: preview, `--confirm-switch`, `--confirm` plus `--confirm-switch`, a terminal yes, `--revert-switch`, and the refusals. (S3b: needs the flags in `src/main.rs`, which card C-75 holds until its merge.)
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
- A person who edited the old script by hand gets `left` and no token: the update never overwrites a file it did not recognise.
- The installer keeps leaving a regular `moai` file alone (install.sh:549-550); after the apply `moai` is still a regular file, so the next install does not replace it.
- `~/go/bin/moai` keeps the old script after the apply. A person who runs the Go build's `moai` from that folder keeps the v1 routing until they copy the new file there themselves; the guidance text says so.
- The apply changes where `moai statusline` goes for every Go-generation project whose `.moai/status_line.sh` runs it (S2a Result, decision 3). The guidance text names this; the project files are not touched.
- A crash between the saved-copy write and the rename leaves a harmless `moai.before-<UTC>` beside a still-v1 `moai`; the next apply uses a new stamp, and the stamp check refuses only a name that exists.
- The revert moves the saved copy, so a second revert with the same stamp finds nothing and says so.

## Result

(S3a 구현 뒤에 채워요.)
