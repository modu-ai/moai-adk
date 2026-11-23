---
id: SPEC-20261123-home-state-retire
title: "The earlier HOME state folder <home>/.moai-cli is never read: one confirmed step archives it whole under <home>/.moai, and the HOME resolver answers <home>/.moai only"
status: draft
revision: 1
created: 2026-11-23
---

# 옛 HOME 상태 폴더 `<home>/.moai-cli` 는 읽지 않고, 확인 한 번으로 `<home>/.moai/archive/` 아래에 통째로 치워요

## Intent

사용자 결정 6504(2026-11-22)와 리더 답(2026-11-23, 카드 C-199 노트)이에요.

- moai 는 `<home>/.moai-cli` 를 다시는 읽지 않아요. 대체 읽기도, 이동하려는 읽기도 없어요. 볼트는 다시 입력해요(`moai key-set glm --prompt`).
- 옛 폴더는 파일을 열지 않고 이름 바꾸기 한 번으로 `<home>/.moai/archive/moai-cli-<UTC>/` 아래로 옮겨요(같은 파일 시스템, 권한·바이트·볼트 0700/0600 그대로). 미리보기가 모든 항목과 막힘을 보이고, 확인 토큰 뒤에만 적용하고, 되돌리기는 이름을 되돌려요. 아무것도 지우지 않아요. 보관 폴더를 지우는 일은 사용자가 따로 골라요.
- HOME 리졸버는 `<home>/.moai` 만 답해요(`<home>/.moai-cli` 가 있든 없든). 기존 카드 기준의 "옮긴 뒤 옛 경로도 읽음" 은 6504 에 따라 폐기해요.

왜 이동·병합이 아닌가: 옛 폴더에는 설치 id 가 바이너리 경로 해시라서 폴더가 옮겨지면 영수증이 낯선 것이 되고(`install.rs`, `update.rs` 의 별칭 읽기), 백업 영수증이 절대 경로를 품고, 이름이 겹치는 항목(`backups`, `bin`, `projects`, `state`)은 Go 세대와 한 폴더를 나눠 써요. 이 모든 읽기를 되살리는 이동은 6504 가 일부러 없앤 읽기를 되살려요. 보관 이동은 파일을 읽지 않아서 이 문제들이 생기지 않아요.

### 읽은 것(2026-11-23, NeXT 끝 `4266b0586c`, 읽기만)

- 이미 있는 짝: `src/home_retire.rs`(Go 상태를 `<home>/.moai-go-retired/<UTC>/` 로 치움)와 `src/move_journal.rs`(이름 바꾸기 한 번씩, 쓰기 전 저널, 재개, 같은 항목만 되돌림, 같은 장치만, 링크는 따라가지 않음). 새 단계는 이 둘의 뼈대를 그대로 써요.
- 이동기 `home-install-plan/apply/revert`(`src/home_install*.rs`)는 옛 폴더를 일부러 읽어요(`home_install.rs:40-41, 712-713, 792`). 이 읽기를 지우는 일은 `update.rs`·`tests/update_moved_receipts.rs`(다른 카드가 쥠)와 얽혀서 이번 조각 밖이에요(Non-goals).
- HOME 리졸버(`src/home_root.rs`)는 지금 "옛 폴더가 있고 표식이 없으면 옛 폴더" 규칙이에요(slice 1).

## Non-goals

- 이동기 `home-install-*` 와 별칭 읽기(`committed_aliases`, `earlier_ids_of`)를 지우는 일, 78개 테스트 픽스처를 `<home>/.moai` 로 옮기는 일: 다른 카드가 쥔 파일(`src/update.rs`, `tests/update_moved_receipts.rs`)과 얽혀서 뒤 조각이에요.
- `moai update` 에 이 단계를 거는 한 줄(`src/update.rs`): 그 파일을 쥔 카드가 착지한 뒤에 해요. 이 조각은 독립 명령 `home-state-retire-*` 로 제공해요.
- 보관 폴더를 지우는 일, 셸 파일(`.zshrc` 등) 수정, 설치 스크립트 수정, 볼트 내용 읽기·복사.
- 프로젝트 쪽 `.moai-cli`(리졸버 `state_root`, `clone_root`)는 그대로예요.
- 실제 HOME: 모든 테스트와 위험 경로 실행은 임시 HOME 이에요.

## Requirements

- REQ-1: The preview shall list each top-level entry of `<home>/.moai-cli` with its kind and mode read by `lstat` alone, opening no file and following no link · failure: none — a read changes nothing
- REQ-2: When `<home>/.moai-cli` is a real folder and nothing blocks, the preview shall plan one move of the whole folder to `<home>/.moai/archive/moai-cli-<stamp>/` and give a `plan_digest` and a `content_digest` · failure: none — a preview writes nothing
- REQ-3: IF `<home>/.moai-cli` is absent, THEN the preview shall report the action `none`
- REQ-4: IF `<home>/.moai-cli` is a link or not a folder, or `<home>/.moai` or `<home>/.moai/archive` exists and is not a folder, or the archive name is taken, or the two paths are on different file systems, THEN the preview shall report the action `blocked` with one named reason each · failure: refused — the apply writes nothing
- REQ-5: IF a process uses a path under `<home>/.moai-cli`, or the running binary lies inside it, or the process table cannot be read, THEN the preview shall report the action `blocked` · failure: refused — the apply writes nothing
- REQ-6: The preview shall list the `PATH` entries of the calling process that name `<home>/.moai-cli` and print the guidance to run `moai update` and use the folder it names, without reading or writing any shell file · failure: none — guidance text only
- REQ-7: When the apply is given a permission reference and the `content_digest` or `plan_digest` of a preview recomputed at that moment, it shall write the manifest and the move journal, then rename `<home>/.moai-cli` to the archive name once
- REQ-8: The apply shall leave no `<home>/.moai-cli` and an archive folder that holds the same entries with the same modes and bytes as before, the vault folder at 0700 and its files at 0600 included · failure: partial — the journal lets the revert or a rerun finish
- REQ-9: The apply shall change nothing in `<home>/.moai` outside `archive/`, so that every Go-owned name stays as it was · failure: refused — a block names the cause
- REQ-10: When the revert is given a permission reference, it shall rename the archive folder back only when it is still the recorded entry and `<home>/.moai-cli` is free, and delete nothing
- REQ-14 (REQ-7): IF the digest differs, the reference is missing or the action is not `retire`, THEN the apply shall write nothing
- REQ-15 (REQ-10): IF the archive entry was replaced or `<home>/.moai-cli` exists again, THEN the revert shall leave the entry where it is and report `revert_partial`
- REQ-11: The apply and the revert shall be safe to run again after a crash at any journal step: a rerun with the same clock finishes the move, and a finished journal is left unchanged · failure: `needs_review` — a rename whose result cannot be proven is not repeated
- REQ-12: The HOME resolver shall answer `<home>/.moai` for every HOME, whatever `<home>/.moai-cli` holds, without looking at `<home>/.moai-cli` · failure: none — a read changes nothing
- REQ-13: IF `<home>/.moai-cli` is absent or archived, THEN a command that writes HOME state shall create no `<home>/.moai-cli` · failure: none — the earlier name is no longer written

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): in a temporary HOME with `<home>/.moai-cli` (vault, versions, path, receipts, a file and a link inside) and a Go `<home>/.moai`, the preview lists every top-level entry with kind and mode, plans one move, names the archive folder and writes nothing (a before and after snapshot of the HOME is equal); a HOME without the folder reports `none` · verify: `cargo nextest run --test home_state_retire_cli -E 'test(the_preview_)'`
- AC-2 (REQ-4, REQ-5): each blocked case — a link, a plain file, a link as `<home>/.moai`, a file as `archive`, a taken name, a held-open file (a child holding a file under the folder), a running binary inside the folder, an unreadable process table — reports `blocked` with its reason and the apply refuses with the HOME unchanged · verify: `cargo nextest run --test home_state_retire_cli -E 'test(blocked_)'`
- AC-3 (REQ-6): a `PATH` that names `<home>/.moai-cli/path` is listed in the preview with the guidance text, and a `.zshrc` in the temporary HOME is byte-identical afterward · verify: `cargo nextest run --test home_state_retire_cli -E 'test(path_guidance_)'`
- AC-4 (REQ-7, REQ-8, REQ-9): with a valid reference and token the apply leaves no `<home>/.moai-cli`; the archive folder holds a byte-identical tree with identical modes (vault 0700, its files 0600), a file under the Go `<home>/.moai` names is byte-identical, a stale or altered token and a missing reference write nothing · verify: `cargo nextest run --test home_state_retire_cli -E 'test(the_apply_)'`
- AC-5 (REQ-10): the revert restores the folder with the same tree hash and modes; it refuses (and `revert_partial`) when `<home>/.moai-cli` was recreated or the archive entry was replaced; a second revert changes nothing · verify: `cargo nextest run --test home_state_retire_cli -E 'test(the_revert_)'`
- AC-6 (REQ-11): a crash injected after the journal, after marking the move, after the rename and before the commit leaves a journal that a rerun finishes and that the revert undoes · verify: `cargo nextest run --test home_state_retire_cli -E 'test(a_crash_)'`
- AC-7 (REQ-12, REQ-13): `home_root::dir` and `path` answer `.moai` for an empty HOME, a HOME with only `<home>/.moai-cli`, a HOME with both, a Go `<home>/.moai`, a schema-1 and a schema-2 marker and an unreadable `<home>/.moai-cli`; no command of `install-inspect` or `key-set` creates `<home>/.moai-cli` in an empty temporary HOME · verify: `cargo nextest run --test path_root_default -E 'test(home_)'`
- AC-8 (REQ-1, REQ-2, REQ-7, REQ-10, REQ-12): the new files pass the repository scans and the help text of the three commands exists in both languages · verify: `cargo nextest run --test help_all_cli --test state_folder_literal_scan`
- AC-9 (REQ-7, REQ-8): the dangerous path runs once with a binary built from the head in a throwaway HOME (preview, apply, a held-open child, revert), table in the READY note · verify: `moai home-state-retire-plan --home <temp>` with the built binary

## Scope

- write: `.moai/specs/SPEC-20261123-home-state-retire/**`
- write: `src/home_state_retire.rs`, `src/ops/home_state_retire_plan.rs`, `src/ops/home_state_retire_apply.rs`, `src/ops/home_state_retire_revert.rs`, `src/lib.rs` (one `pub mod`), `src/help.rs` (the three commands)
- write: `src/home_root.rs` (the resolver and its unit tests)
- write: `tests/home_state_retire_cli.rs`, `tests/path_root_default.rs`
- write: `docs/design-review/appendix/operations/home_state_retire_plan.md`, `…_apply.md`, `…_revert.md`
- read-only: `src/update.rs`, `src/host_home.rs`, `src/bin_names.rs` (held by another change), `src/home_retire.rs`, `src/move_journal.rs`, `src/home_state.rs`, `src/home_layout.rs`
- forbidden: the real HOME and anything under `<home>/.moai-cli`, shell profile files, the Go repository, a card number in a file name, SPEC id, branch or commit

## Plan

1. RED: `tests/home_state_retire_cli.rs` and the changed `path_root_default` home cases, run on the tip so they fail for the missing command and the old resolver rule.
2. GREEN part 1: `src/home_state_retire.rs` (plan, apply, revert on `move_journal` with kind `state-retire`, token, in-use and running-binary checks) and the three op files. `<home>/.moai/archive` gets a row in the HOME name table if the table test asks for it.
3. GREEN part 2: `home_root::dir` and `path` answer `.moai` always; `is_new_home` no longer asks for `<home>/.moai-cli`; run the library and measure which test targets fail because their fixture builds `<home>/.moai-cli` as the HOME state. The list goes in the READY note; the fixture move is the next slice.
4. Dangerous path: build the binary in the worktree and run preview, apply with a held-open child, revert in a throwaway HOME.

## Risks

- A HOME whose running `moai` lives under `<home>/.moai-cli/versions` loses its launcher when archived. The preview blocks while the running binary is inside the folder and warns when `<home>/.moai` holds no version folder.
- The resolver change makes every test that builds `<home>/.moai-cli` as the HOME state fail until its fixture moves. The slice measures and lists them; landing waits for the fixture slice or the leader's call.
- A held-open file under the folder is judged by the process table once at the preview and again at the apply; a program that opens a file between the check and the rename is a race the rename itself does not see. The journal's own in-use check narrows it; the apply is a user-confirmed step.
- `<home>/.moai/archive` is new to a Go `<home>/.moai`; the name is not a Go name in the HOME table, and the apply writes nothing else there.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Archive by one rename of the whole folder, not move-and-merge | User decision 6504 (never read `<home>/.moai-cli`) and leader answer 2026-11-23 | Replace the step with the mover by a new decision |
| D-2 | The step is its own command pair, not a `moai update` layer yet | `src/update.rs` is held by another card; the hook-up is one line after it lands | Add the layer later |
| D-3 | The old criterion "old paths are still read" is dropped | User decision 6504 | none |
| D-4 | The move stops while a process uses the folder or the running binary is inside it | A launcher inside the archived folder would break the session that runs the step | none |

## Result

(not yet — filled after the checks run)
