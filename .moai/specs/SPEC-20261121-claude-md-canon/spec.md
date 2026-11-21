---
id: SPEC-20261121-claude-md-canon
title: "Personal guidance moves from AGENTS.local.md to CLAUDE.local.md, with preview and rollback"
status: draft
revision: 1
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
- REQ-5: Where the project has an `AGENTS.md` and no `CLAUDE.md`, the plan shall warn that Claude Code stops reading `AGENTS.md` once `CLAUDE.local.md` exists · failure: none — the warning is part of the preview the user confirms.
- REQ-6 (REQ-2): IF the target is a link, THEN the migration shall refuse and write nothing.
- REQ-7 (REQ-2): IF the plan differs from the confirmed preview at apply time, THEN the apply shall write nothing and report a stale plan.
- REQ-8 (REQ-4): IF the file changed since the apply, THEN the revert shall write nothing and say so.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): 임시 프로젝트에서 `AGENTS.local.md` 글이 `CLAUDE.local.md` 로 가는 미리보기가 글·위치·깃 제외 파일임을 보여 주고, 다시 하면 `unchanged`, `AGENTS.md` 만 있는 프로젝트에선 경고가 붙어요 · verify: `cargo test --offline --locked --test agents_local_merge_cli`
- AC-2 (REQ-2, REQ-4, REQ-6, REQ-7, REQ-8): 적용은 없던 `CLAUDE.local.md` 를 만들고 영수증에 적용 전 바이트와 원본 사본을 남기며, 되돌리기는 만든 파일을 지우고 뒤에 고친 파일은 건드리지 않아요 · verify: `cargo test --offline --locked --test agents_local_merge_cli`

## Scope

- write: `src/agents_local_merge.rs`, `tests/agents_local_merge_cli.rs`, this SPEC folder
- read-only: `src/apply.rs`, `src/go_state.rs`, `src/claude.rs`
- forbidden: the user's real home and projects, `src/project_update.rs` (held), any release or deploy action

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

- `src/agents_local_merge.rs`: 이관 대상을 `AGENTS.md` 에서 `CLAUDE.local.md` 로 바꿨어요. `AGENTS.local.md` 글이 `CLAUDE.local.md` 끝에 표시 줄 사이로 한 번 들어가고, 파일이 없으면 만들고(되돌리기는 만든 파일을 지워요), 있으면 덧붙여요. 미리보기에 `warnings` 가 생겼어요: `AGENTS.md` 가 있고 `CLAUDE.md`·`.claude/CLAUDE.md`·`CLAUDE.local.md` 가 없을 때 이 이관이 첫 `CLAUDE.local.md` 를 만들면 Claude Code 가 `AGENTS.md` 를 읽지 않게 된다고 알려요. 대상이 링크이거나 읽을 수 없으면 계획이 `blocked` 예요. 영수증에 `target_path`·`target_existed` 가 생겼고, 옛 영수증(필드 없음)은 그대로 `AGENTS.md` 를 되돌려요.
- `tests/agents_local_merge_cli.rs`: 기존 18개를 새 대상에 맞게 바꾸고 3개를 더했어요(없던 파일 만들기와 되돌리기, 경고, 옛 영수증).

### 증거 (2026-11-21, 이 작업 트리에서 돌린 것)

- `cargo test --offline --locked --test agents_local_merge_cli`: 21 passed(`target/red.log`). `--lib agents_local_merge::`: 15 passed, `--test project_update_plan_cli`: 60 passed(`target/nb.log`; 이 시험이 쓰는 `commit_notice` 는 "not committed" 라는 말에 `committed` 가 들어 있어 그대로 통과해요).
- 변형 5개 모두 killed(`target/mutants.py`): 경고 끄기, 되돌리기가 만든 파일을 안 지우기, 옛 영수증이 새 파일을 가리키기, 영수증이 항상 "있었다"고 적기, 경고가 기존 `CLAUDE.md` 를 무시하기. 변형 뒤 `git diff` 가 의도한 변경만 보여요.
- 돌리지 않은 것: 시험을 구현보다 먼저 돌려 실패를 보는 RED 관측은 못 했어요(시험과 구현을 같은 작업에서 썼고, 대신 변형으로 시험이 잡아내는지 확인했어요). 병합 뒤 `check affected`(fmt·clippy 포함)는 병합 때 돌려요.

### 수용 기준 결과

- AC-1: met — `agents_local_merge_cli`(미리보기의 글·위치·`target.path`·"not committed" 알림, 다시 하면 `unchanged`, `AGENTS.md` 만 있을 때 경고, `CLAUDE.md`·`.claude/CLAUDE.md`·기존 `CLAUDE.local.md`·`AGENTS.md` 없음일 때 경고 없음).
- AC-2: met — `agents_local_merge_cli`(없던 파일 생성, 영수증에 적용 전 복사본 없음과 원본 사본, 뒤에 고친 파일은 되돌리기 거절, 만든 파일 삭제, 링크 거절, 낡은 계획 거절, 옛 영수증은 `AGENTS.md` 복원).

### 남은 것 (후속)

- 조각 B(생성이 `CLAUDE.md`·`CLAUDE.local.md` 를 쓴다)와 C(`AGENTS.md` 이관): 자기 SPEC. `src/agents.rs`·`src/generation.rs` 는 다른 카드가 잡고 있어요.
- `src/project_update.rs` 가 새 `warnings` 필드를 화면에 전하기, `tests/project_update_plan_cli.rs` 의 `committed` 단언 고치기, `src/ops/agents_merge_*.rs` 도움말 문구와 도움말 고정 파일: 그 파일들이 풀린 뒤.
