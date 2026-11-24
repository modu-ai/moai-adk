---
id: SPEC-20261122-next-flow-update-and-launcher
title: "The update removes skill and agent files the package no longer ships when moai's bytes are still there, shows the stale count, and the launcher drops its role flags"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-next-flow-skills-agents
---

# Stale files in the update, and the launcher without role flags

## Intent

두 가지를 묶었어요. 이 조각은 SPEC 만 써요.

(1) 갱신(`moai update`)이 패키지에서 빠진 스킬·에이전트 파일을 지금은 `remove_blocked` 로 보고만 하고 그대로 둬요. 그래서 설치본에 낡은 파일이 쌓여요(보고서 2-3: 설치된 스킬 51개 중 27개가 저장소와 달라요). 앞으로는 moai 가 쓴 바이트가 그대로인 파일만 사용자가 확인한 적용에서 지우고, 영수증에 이전 바이트를 남겨 되돌리기로 살려요. 사용자가 고친 파일은 남기고 "보존"으로 보여요. 미리보기는 낡은 파일 수를 한 키로 보여 주되, 수가 0이면 키를 내지 않아서 고정된 열세 개 키(`tests/update_user_layer_cli.rs:525-539`)가 그대로예요.

(2) 실행기의 역할 플래그(`-l`/`--leader`, `-a`/`--agent`)를 코드에서 지워요(사용자가 고른 결정, 흐름 제안 §5). 지금 이 플래그는 세션 이름과 접근 권한만 정하고, 역할은 세션 안의 `/moai --mode-*` 로 들어가요. 모드 경로가 없어지므로(부모 SPEC) 이름은 `--name` 으로, 넓은 접근은 따로 이름 붙은 옵션으로만 줘요. moai 명령이 출력하는 글도 모드 플래그로 들어가라고 말하지 않아요.

### What exists (tip `78c8a5a247`, 잰 시각 2026-11-22T07:3xZ)

- 스킬 파일: 패키지에 없는데 영수증이 소유한 파일은 `remove_blocked`, 소유하지 않은 파일은 `preserved`(`src/update.rs:1295-1318`).
- 에이전트 파일: 같은 규칙으로 `remove_blocked`(`src/agent_install.rs:259-275`). 영수증 항목의 `action` 은 `create | update` 둘뿐이에요(`src/agent_install.rs:105-114`).
- 실행기: 맨 위 `-l`/`--leader`, `-a`/`--agent [HOST]`(`src/main.rs:53-58`), 호스트별 `cc|glm|omp -l|-a`(`src/main.rs:164-181`), 공통 옵션 `--name`, `--model`, `--effort`, `--advisor`, `--no-advisor`, `--permissions`(역할 세션의 권한 질문을 남김), `--dry-run`(`src/main.rs:106-136`). 역할이 없는 세션은 권한 플래그를 받지 않아요.
- 모드 플래그를 출력하는 코드: `src/session.rs:62-65`(dry-run 노트), `src/help.rs`(루트 예시 `cc -l`, 한국어 도움말), `src/lane_idle.rs:249`, `src/lane_duty.rs:75-88`, `src/leader_round.rs:848`. 도움말 골든 파일 `tests/fixtures/help/*` 가 이 글을 담아요.
- 반대로 가는 카드가 하나 있어요(플래그 이름을 바꾸려는 카드). 리더가 "가져가지 않음"으로 표시했어요.

## Non-goals

- 설치 자리와 설정 폴더 목록(전역 설치 조각).
- 레인 띄우기·중단 명령(`moai lane`)과 레인 의무 줄의 새 글: 레인 운영 SPEC 이 정해요. 이 SPEC 은 그 줄에서 모드 플래그 문구만 빼요.
- 스킬·에이전트 밖의 낡은 파일(보고서, 빌드, 작업트리) 정리: 정리(tidy) SPEC.

## Requirements

- REQ-1: When the new package no longer ships a skill or agent file that moai wrote and the file on disk still holds the bytes moai last wrote, the update apply the user confirmed shall remove that file. · failure: none — a file whose bytes differ is kept and shown as preserved, and a file moai did not write is never read or removed.
- REQ-2: When the update apply removes a file, the receipt of that apply shall record the file's path and its bytes before the removal. · failure: none — an apply that cannot write the receipt removes nothing.
- REQ-3 (REQ-2): IF the user reverts an apply that removed a file and the path is still empty, THEN the revert shall write the recorded bytes back to that path.
- REQ-4: The update preview shall show the count of stale skill and agent files under one key that is present only when the count is above zero.
- REQ-5: The launcher shall take a session's name only from --name and its full access only from one explicit full-access option, with no -l, -a, --leader or --agent option left.
- REQ-6: No moai command shall print a text that tells the user to enter a role with a mode flag.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): in an empty temporary HOME, an install of package A then an update to package B that drops one skill file and one agent file removes both when unchanged, keeps and reports an edited one as preserved, records the removed bytes, and a revert writes them back.
  verify: `cargo test --locked --offline --test update_stale_cli stale_`
- AC-2 (REQ-4): the preview of that update holds the stale count key with the right number, and a preview with nothing stale holds exactly the thirteen pinned keys.
  verify: `cargo test --locked --offline --test update_stale_cli count_`
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-3 (REQ-5): `moai -l`, `moai -a`, `moai cc -l` and `moai cc -a` are refused as unknown options, `--name` names the session, and the full-access option alone adds the host's permission flag, in `--dry-run`.
  verify: `cargo test --locked --offline --test launcher_roleless`
  verify: `cargo test --locked --offline --test launcher_no_role_flags`
- AC-4 (REQ-6): the help goldens and every printed note hold no `--mode-agent`, `--mode-leader` or `-l|-a` role text.
  verify: `cargo test --locked --offline --test help_layout`
- AC-5: neighbouring checks.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-next-flow-update-and-launcher/**`
- later slices, write: `src/update.rs`, `src/agent_install.rs`, `src/main.rs`, `src/session.rs`, `src/help.rs`, `src/lane_idle.rs`, `src/lane_duty.rs`, `src/leader_round.rs`, `tests/update_stale_cli.rs`, `tests/launcher_no_role_flags.rs`, `tests/fixtures/help/*`, and the launcher tests that start a role session
- read-only: `src/apply.rs`, `tests/update_user_layer_cli.rs`
- forbidden: the real HOME, the network, the integration branch, `AGENTS.md` and `AGENTS.local.md` of this repository

## Plan

1. (done) SPEC (this slice).
2. Update slice: tests first in `tests/update_stale_cli.rs` (prefixes `stale_`, `count_`), run in an empty temporary HOME with two fixture packages; then the `remove` action in the skill and agent plans, the receipt entry with before bytes, the revert path and the count key.
3. Launcher slice: inventory of every test and golden that starts a role session; tests first in `tests/launcher_no_role_flags.rs`; then the flags go, the full-access option comes in, the printed notes and help lose the mode text, the goldens are regenerated.
4. The lane lines in `src/lane_idle.rs`, `src/lane_duty.rs` and `src/leader_round.rs` lose only the mode-flag words; if a lane operation slice removes those lines first, this step drops out.

## Risks

- Removing a file is the first delete the update makes in the user's config folder. The rule is narrow on purpose (moai's last bytes, the confirmed apply, a receipt with the bytes, a revert), and the test runs in a temporary HOME only.
- The skill package's own receipt may name files from an older generation that a user later edited; those show as preserved and stay.
- Lanes and the leader start sessions with `-l` and `-a` today. The launcher slice lands only after the lane operation SPEC gives lanes the new start command, or lanes lose their start line; the order goes into the READY note.
- The help goldens are wide; the regenerated files are read back and compared line by line before commit.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only files whose bytes equal moai's last write are removed. | The three-way rule the update already uses for writes; a user's edit is never lost. | Ask per file. |
| D-2 | The stale count key appears only when above zero. | The thirteen preview keys are pinned; a new always-present key would break every reader of the fixed list. | Always show it. |
| D-3 | The full-access option's name is decided in the launcher slice and written into this SPEC's next revision; the default is `--full-access`. | The user picked explicit options; the exact spelling is a reversible choice for the leader. | Rename it. |
| D-4 | `--permissions` goes with the role flags. | It only kept prompts for a role session; without roles every session keeps prompts unless the full-access option is given. | Keep it as a no-op. |

## Result

(Filled when the work is done. This is the SPEC-only slice.)
