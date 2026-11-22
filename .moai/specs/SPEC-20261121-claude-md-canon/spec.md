---
id: SPEC-20261121-claude-md-canon
title: "Personal guidance moves from AGENTS.local.md to CLAUDE.local.md, with preview and rollback"
status: draft
revision: 2
created: 2026-11-21
---

# Personal guidance moves from AGENTS.local.md to CLAUDE.local.md, with preview and rollback

## Intent

MoAI targets Claude Code only (the GLM path stays; Codex support is withdrawn). Install and update used to write the managed guidance section into `AGENTS.md`, the personal file as `AGENTS.local.md`, and offered an `@AGENTS.md` line for a `CLAUDE.md` that shadowed it. The user confirmed the final design on 2026-11-21: the team guidance canon is `CLAUDE.md`, the personal file is `CLAUDE.local.md` (git-ignored), MoAI never creates `AGENTS.md` or `AGENTS.local.md`, and never adds an `@AGENTS.md` bridge. An existing `AGENTS.md` or `AGENTS.local.md` is offered for migration with a preview, a backup and a rollback.

Facts read from https://code.claude.com/docs/en/memory on 2026-11-21 (saved copy, line numbers of that copy):

- 355: Claude Code reads `AGENTS.md` as the project instructions without a `CLAUDE.md`, an import or a setting (Claude Code 2.1.277 or later, line 367).
- 360: with an `AGENTS.md` and a `CLAUDE.md` or `CLAUDE.local.md` in the working directory or above, Claude reads "Your `CLAUDE.md` files only".
- 381: `AGENTS.local.md` and `AGENTS.override.md` are not read.
- 385: `CLAUDE.local.md` counts as a `CLAUDE.md` for that check, so adding one to a project that relies on `AGENTS.md` stops Claude reading `AGENTS.md`.
- 125: `CLAUDE.local.md` is for private preferences and belongs in `.gitignore`.

The design is therefore sound, but one order matters for real users: a project with an `AGENTS.md` and no `CLAUDE.md` loses that file from Claude's context the moment MoAI creates `CLAUDE.md` or `CLAUDE.local.md`. The migration of the `AGENTS.md` text into `CLAUDE.md` has to come with, or before, the first creation, and the preview names the effect.

## Non-goals

- Slices B (generation writes `CLAUDE.md`) and C (migrating `AGENTS.md`): their own SPECs.
- Deleting or moving a user's `AGENTS.md` or `AGENTS.local.md`: both stay as history and for other tools.
- The user-level files under the Claude config folder (`home_routing`, `home_style`): unchanged.
- Codex files and the Codex adapter: removed by other work, left alone here.
- `src/project_update.rs` and `tests/agents_choices_step_cli.rs` while another card holds them.
- Writing into a user's real home or project during tests: every test uses a temporary HOME and project.

## Requirements

- REQ-1: When the personal-file migration is previewed, the plan shall show the text of `AGENTS.local.md` that would go into `CLAUDE.local.md`, its position, and that the target is a git-ignored personal file · failure: none — a preview writes nothing.
- REQ-2: When the personal-file migration is applied after a confirmed preview, the apply shall create `CLAUDE.local.md` when it is missing or add one marked block to it, and keep a receipt with the before bytes and a copy of the source · failure: the receipt cannot be written, so nothing is written to the target.
- REQ-3: When the migration runs again over an unchanged source, the plan shall be `unchanged` and write nothing · failure: none — an unchanged plan needs no confirmation.
- REQ-4: When the migration is reverted from its receipt, the revert shall restore the earlier bytes of `CLAUDE.local.md` or remove the file the apply created, only while the file still holds the bytes the apply wrote.
- REQ-5: Where the project has an `AGENTS.md` or a `.claude/AGENTS.md` and none of `CLAUDE.md`, `.claude/CLAUDE.md` and `CLAUDE.local.md`, the plan shall warn in plain words that Claude Code stops reading the `AGENTS.md` files once `CLAUDE.local.md` exists, and name the official setting that keeps both read (Project instructions `claude-md-and-agents-md`) · failure: none — the warning is part of the preview the user confirms.
- REQ-6 (REQ-2): IF the target is a link, THEN the migration shall refuse and write nothing.
- REQ-7 (REQ-2): IF the plan differs from the confirmed preview at apply time, THEN the apply shall write nothing and report a stale plan.
- REQ-8 (REQ-4): IF the file changed since the apply, THEN the revert shall write nothing and say so.
- REQ-9 (REQ-5): IF the plan carries that warning, THEN the apply shall refuse and write nothing unless the caller passes the confirmation token printed with the plan.
- REQ-10: Where git does not ignore `CLAUDE.local.md`, the plan shall add the one line `CLAUDE.local.md` to `.gitignore` through the same preview, receipt and revert, without rewriting any other line of that file · failure: none — a project that is not a git repository gets no line and a note.
- REQ-11 (REQ-10): IF git already tracks `CLAUDE.local.md`, THEN the plan shall warn that the file stays tracked and that this move does not untrack it.
- REQ-12 (REQ-1): IF the text of `AGENTS.local.md` is already in a block an earlier generation merged into `AGENTS.md`, THEN the plan shall be `unchanged` and say where the text is.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): 임시 프로젝트에서 `AGENTS.local.md` 글이 `CLAUDE.local.md` 로 가는 미리보기가 글·위치·깃 제외 파일임을 보여 주고, 다시 하면 `unchanged`, `AGENTS.md` 만 있는 프로젝트에선 경고가 붙어요 · verify: `cargo test --offline --locked --test agents_local_merge_cli`
- AC-2 (REQ-2, REQ-4, REQ-6, REQ-7, REQ-8): 적용은 없던 `CLAUDE.local.md` 를 만들고 영수증에 적용 전 바이트와 원본 사본을 남기며, 되돌리기는 만든 파일을 지우고 뒤에 고친 파일은 건드리지 않아요 · verify: `cargo test --offline --locked --test agents_local_merge_cli`
- AC-3 (REQ-5, REQ-9): `AGENTS.md` 나 `.claude/AGENTS.md` 만 있는 프로젝트의 계획에 "더 이상 읽지 않는다" 경고와 공식 설정 이름이 붙고, 확인 토큰 없이는 적용이 거절돼 프로젝트가 안 바뀌며, `/moai update` 미리보기가 같은 경고와 토큰을 담아요 · verify: `cargo test --offline --locked --test agents_local_merge_cli` and `--test project_update_plan_cli`
- AC-4 (REQ-10, REQ-11): 깃이 `CLAUDE.local.md` 를 무시하지 않으면 계획이 `.gitignore` 한 줄 추가를 미리보기·영수증·되돌리기로 하고(다른 줄은 그대로, 파일이 없으면 만들고 되돌리기가 지움), 이미 추적 중이면 경고만 해요 · verify: `cargo test --offline --locked --test agents_local_merge_cli`
- AC-5 (REQ-12): 이전 세대가 `AGENTS.md` 에 합쳐 둔 같은 글이 있으면 계획이 `unchanged` 로 그 위치를 말해요 · verify: `cargo test --offline --locked --test agents_local_merge_cli`

## Scope

- write: `src/agents_local_merge.rs`, `tests/agents_local_merge_cli.rs`, `src/ops/agents_merge_plan.rs`, `src/ops/agents_merge_apply.rs`, `src/ops/agents_merge_revert.rs`, this SPEC folder, and the four help goldens that carry the help text of `agents-merge-*` (`tests/fixtures/help/all.txt`, `all-80.txt`, `all-60.txt`, `all-dumb.txt`); and by pair note with the holder of each file: the function `agents_local_merge_step` and one constant of `src/project_update.rs`, and the helper `apply_step` plus two new tests beside the merge tests of `tests/project_update_plan_cli.rs`
- read-only: `src/apply.rs`, `src/go_state.rs`, `src/claude.rs`
- forbidden: the user's real home and projects, any other line of `src/project_update.rs`, any release or deploy action

## Plan

This SPEC is slice A. The design has three slices; B and C get their own SPECs (parent: this one) so that each stays under twelve requirements.

1. A (now, no overlap): re-target the personal-file migration from `AGENTS.local.md` into `AGENTS.md` to `AGENTS.local.md` into `CLAUDE.local.md`; a missing target is created and its revert removes it; the plan warns when `AGENTS.md` would stop being read. RED tests first.
2. B (own SPEC; after the held files are free or by pair note): generation writes `CLAUDE.md` and `CLAUDE.local.md`, the ignore block lists the new name, the bridge pick becomes a no-op, the launcher skips when either file exists; mechanical renames in the tests that name the old files.
3. C (own SPEC): the `AGENTS.md` migration module (`agents_migrate`) with its plan, apply and revert commands, modelled on A.
4. Help, skills and docs wording is `moai-docs` work after the code lands.

## Risks

- The doc comments of `src/ops/agents_merge_*.rs` are the help text of the commands and are pinned by the help goldens; they keep the old wording until the goldens are free, and the wording follows in the docs slice.
- `src/project_update.rs` does not forward the new `warnings` field yet, and `tests/project_update_plan_cli.rs` still asserts the word `committed` in the notice; the notice therefore says the text is "not committed", and the held file gets the field in a later change.
- The warning of REQ-5 is the only guard for a project that has an `AGENTS.md` and no `CLAUDE.md` until slice C exists; slice A therefore shows it in the preview, and slice B does not reach users without C.
- `src/agents.rs` and `src/generation.rs` are reserved by another card; the lane widens by pair note over different functions or waits.
- Many tests name the old files (`product_cli` 75 lines, `agents_local_merge_cli` 81); a rename slice touches several shared test files.
- `local_file: tracked` conflicts with the documentation's advice to ignore `CLAUDE.local.md`; it stays allowed with a warning.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Keep the function and command names `agents-merge-*` for the personal file and add `agents-migrate-*` for `AGENTS.md` (reversible, category none) | `project_update` already calls the merge by those names and is held | Rename in a later change |
| D-2 | The bridge pick `claude-import` stays parseable and becomes a no-op until the held choices test is free (reversible, category none) | `tests/agents_choices_step_cli.rs` is held | Delete the dormant code later |
| D-3 | A credential-shaped line in `AGENTS.md` text blocks the plan, as for the personal file (reversible, category none) | The text goes into a committed file | Change to a warning |
| D-4 | `AGENTS.md` and `AGENTS.local.md` are never deleted (reversible, category none) | Other tools and history still read them | Add an optional cleanup step |

## Result

### 만든 것

- `src/agents_local_merge.rs`: 이관 대상이 `CLAUDE.local.md` 예요(없으면 만들고 되돌리기가 지워요, 있으면 표시 줄 사이로 한 번 덧붙여요). 감사 FAIL 수리로 더해진 것:
  - 경고: `AGENTS.md` 나 `.claude/AGENTS.md` 가 있고 `CLAUDE.md`·`.claude/CLAUDE.md`·`CLAUDE.local.md` 가 없으면, 두 파일을 이름으로 대며 "더 이상 읽지 않는다"와 공식 설정(Project instructions `claude-md-and-agents-md`)을 알려요. 그 계획에는 `needs_confirmation` 과 `confirm_token` 이 붙고, 적용은 `--confirm-warnings <토큰>` 없이는 아무것도 쓰지 않고 거절해요(토큰은 경고 글에서 만들어 계획 지문에 묶여요).
  - `.gitignore`: 깃 저장소이고 `CLAUDE.local.md` 가 무시되지 않으면 한 줄만 덧붙여요(미리보기의 `gitignore` 필드, 영수증의 `gitignore_*`, 되돌리기는 파일이 안 바뀐 동안만 되돌리고 바뀌었으면 `kept`). 깃 저장소가 아니면 줄 없이 "not a git repository" 설명을 남겨요. 이미 추적 중이면 경고만 하고 추적을 풀지 않아요(`commit_notice` 도 "tracks" 로 바뀌어 사실과 맞아요).
  - 이전 세대가 `AGENTS.md` 에 합쳐 둔 같은 글(표시 줄의 sha256 과 본문이 모두 같을 때)이면 `unchanged` 로 그 위치를 말해요. 줄 끝이 CRLF 면 미리보기에 그 한 줄이 나와요.
- `src/ops/agents_merge_{plan,apply,revert}.rs`: 도움말 문구를 새 동작에 맞췄고, 적용 명령에 `--confirm-warnings` 가 생겼어요.
- `src/project_update.rs`(함수 `agents_local_merge_step` 와 상수 하나, 같은 파일을 고치는 다른 변경과 pair 증거 노트): `/moai update` 미리보기가 `warnings`·`needs_confirmation`·`confirm_token`·`gitignore`·`line_endings` 를 전하고, 경고가 있는 계획에만 `--confirm-warnings <warning-confirmation>` 자리표시자를 권한 참조 앞에 찍어요(그 밖의 단계는 자리표시자가 그대로 하나뿐이에요).
- 시험: `tests/agents_local_merge_cli.rs` 21 → 28개, `tests/project_update_plan_cli.rs` 에 시험 도우미와 새 시험 2개.

### 증거 (2026-11-21, 이 작업 트리에서 돌린 것)

- RED 관측: 엔진 시험 7개는 구현 전에 실패했어요(`target/red4.log`: 21 passed, 7 failed). update 시험은 소스 패치를 되돌린 상태에서 새 시험 2개와 옛 병합 시험이 실패했어요(`target/red5.log`: 3 failed).
- GREEN: `--test agents_local_merge_cli` 28 passed, `--test agents_migrate_cli` 11 passed(`target/green6.log`), `--lib agents_local_merge::` 18 passed(`target/green5.log`), `--test project_update_plan_cli` 62 passed(`target/green7.log`).
- 변형 5개 모두 killed(`target/mut3.log`, `target/mutants3.py`): 확인 토큰 검사 끄기, `.claude/AGENTS.md` 무시, 되돌리기가 바뀐 `.gitignore` 를 덮어쓰기, 마지막 줄에 붙여 쓰기, "이미 합쳐짐" 이 표시 줄 sha256 을 무시. 변형 뒤 변형 흔적 0(`grep`).
- 돌리지 않은 것: `/moai update` 단계의 변형(수리 감사 상한 5개를 엔진에 썼어요), 도움말 고정 파일 재생성과 병합 뒤 `check affected`(fmt·clippy 포함)는 병합 때 돌려요.

### 수용 기준 결과

- AC-1: met — `agents_local_merge_cli`(글·위치·`target.path`·알림, 다시 하면 `unchanged`, 경고 유무).
- AC-2: met — `agents_local_merge_cli`(생성, 영수증, 고친 파일 되돌리기 거절, 링크·낡은 계획 거절, 옛 영수증).
- AC-3: met — `agents_local_merge_cli`(`the_warning_names_the_files_and_the_setting_and_needs_a_token`, `an_apply_without_the_token_writes_nothing`), `project_update_plan_cli`(`the_merge_preview_warns_and_the_apply_waits_for_the_confirmation`).
- AC-4: met — `agents_local_merge_cli`(`a_git_project_gets_one_ignore_line_through_the_same_receipt`, `an_ignored_file_gets_no_line_and_a_folder_without_git_gets_a_note`, `a_tracked_personal_file_is_warned_about_and_not_untracked`), `project_update_plan_cli`(`the_merge_step_adds_the_ignore_line_in_a_git_project`).
- AC-5: met — `agents_local_merge_cli`(`text_already_merged_into_agents_md_is_unchanged`, 표시 줄 sha256 을 속인 경우 포함).

### 남은 것 (후속)

- 조각 B(생성이 `CLAUDE.md`·`CLAUDE.local.md` 를 쓴다)와 C(`AGENTS.md` 이관): 자기 SPEC.
- `skills/moai/workflows/update.md` 의 미리보기 문장(경고와 확인 자리표시자 설명)은 문서 담당(moai-docs)의 몫이에요.
- 감사자가 말한 변형 M4·M5 가 정확히 무엇인지는 리더에게 물었고(노트 5646) 아직 답이 없어요. 답이 오면 그 시험을 더해요.
