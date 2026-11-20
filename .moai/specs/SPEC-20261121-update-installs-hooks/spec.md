---
id: SPEC-20261121-update-installs-hooks
title: "moai update offers the Claude Code and Codex hooks the same way and installs them on confirmation"
status: draft
revision: 1
created: 2026-11-21
---

# moai update offers the Claude Code and Codex hooks the same way and installs them on confirmation

## Intent

지금 `moai update` 의 훅 계층은 프로젝트에 MoAI 훅이 이미 깔린 호스트만 고쳐요. 새 프로젝트에서는 Claude 훅이 아예 제안되지 않고, `project-update` 의 훅 단계는 사용자에게 "`moai hooks-plan` 을 직접 치라"고 안내해요(세션은 사용자에게 명령을 치라고 하지 않는다는 규칙과도 부딪혀요). Codex 훅은 새 프로젝트에서 이미 제안돼요. 사용자 선택(2026-11-20, 리더 기록): `moai update` 가 Claude 훅과 Codex 훅을 같은 방식으로 — 미리보기 후 확인, 영수증, 되돌리기 — 설치하고, 직접 치라는 문장은 지워요.

이 SPEC 은 그 선택을 그대로 구현해요. 설치하는 일은 이미 있는 훅 계층의 계획·적용·영수증·되돌리기가 하고, 바뀌는 것은 "깔려 있지 않은 호스트도 첫 설치로 계획한다"는 한 가지예요.

## Non-goals

- 훅 내용 자체(어떤 이벤트에 무엇을 거는지)와 `hooks-plan`·`hooks-apply` 명령의 동작.
- OMP 훅(이 변경의 대상이 아니에요).
- 호스트별 선택 화면: 미리보기에 호스트마다 동작이 보이고 한 번의 확인으로 계층 전체가 적용돼요(지금 흐름 그대로).
- 사용자가 직접 고친 훅 파일을 덮어쓰는 일: 충돌로 보이고 그대로 둬요.
- Claude `permissions.deny` 설치 제안(다른 변경의 일).

## Requirements

- REQ-1: The hook layer of `moai update` shall plan a first install for each of the Claude Code and Codex hook sets that is not installed in the git work tree and that the project profile does not leave out.
- REQ-2: If the first-install plan of a host is a conflict, then the layer shall show that host as a conflict with its reason and write nothing for it.
- REQ-3: When `moai update` runs without a confirmation, the layer shall write nothing and show each planned host with its action. · failure: none — the preview only reads the project.
- REQ-4: When `moai update` is confirmed with the plan digest, the layer shall write each planned host through the apply that `hooks-apply` uses and keep one receipt for the layer. · failure: none — REQ-5 covers a host whose apply fails.
- REQ-5: If the apply of one host fails, then the layer shall report that host's error beside the other hosts and keep writing the others.
- REQ-6: The revert of the layer receipt shall remove each file that the first install created.
- REQ-7: The plan digest of the update shall cover every planned host, so a plan that changed after the preview is refused.
- REQ-8: The hook layer shall keep updating an installed host as it did before this change.
- REQ-9: The hook layer shall skip a work tree that is the home folder or is not a git work tree, as it did before this change.
- REQ-10: The `hooks` step of `project-update` shall not tell the user to run `moai hooks-plan` or `moai hooks-apply` by hand.

## Acceptance

- AC-1 (REQ-1, REQ-3): in a git project with no MoAI hooks and no profile, the update preview lists claude and codex as `create`, and no hook file exists afterwards.
  verify: `cargo test --locked --offline --test update_installs_hooks_cli first_install_is_previewed`
- AC-2 (REQ-1): a profile that lists only claude gets no codex offer, and a profile that lists both gets both.
  verify: `cargo test --locked --offline --test update_installs_hooks_cli the_profile_hosts_decide_the_offer`
- AC-3 (REQ-2): a Claude settings file with a `hooks` key the user set is shown as a conflict with its reason, is not written, and does not stop the Codex install.
  verify: `cargo test --locked --offline --test update_installs_hooks_cli a_conflict_is_shown_and_left`
- AC-4 (REQ-4, REQ-5): a confirmed update writes `.claude/settings.json` and `.codex/hooks.json` and leaves one receipt naming both hosts.
  verify: `cargo test --locked --offline --test update_installs_hooks_cli confirm_installs_both_with_one_receipt`
- AC-5 (REQ-6): `hooks-revert` with that receipt removes both created files and the project is as it was.
  verify: `cargo test --locked --offline --test update_installs_hooks_cli revert_removes_what_the_install_created`
- AC-6 (REQ-7): a plan digest from before a change in the project is refused on confirm and nothing is written.
  verify: `cargo test --locked --offline --test update_installs_hooks_cli a_stale_digest_is_refused`
- AC-7 (REQ-8, REQ-9): the existing tests of the hook layer for an installed host, the home folder and a folder that is not a git tree still pass.
  verify: `cargo test --locked --offline --test update_hooks_cli`
- AC-8 (REQ-10): the `hooks` step of `project-update` for a project with no hook record keeps its reason without the two commands.
  verify: `cargo test --locked --offline --test project_update_hooks_cli`
- AC-9 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261121-update-installs-hooks`

## Scope

- write: `.moai/specs/SPEC-20261121-update-installs-hooks/**`, `src/update_hooks.rs` (the plan loop), `src/project_update.rs` (the one reason sentence), `src/update_tux.rs` and `tests/fixtures/update_tux/**` and `tests/update_tux_golden.rs` (the screen rows and goldens that show the hook layer), `tests/update_installs_hooks_cli.rs` (new), `tests/update_hooks_cli.rs` (the test that pins the old skip reason), `tests/project_update_hooks_cli.rs` (the test that pins the sentence)
- read-only: every other file, the real HOME
- tier 1: an independent audit of the exact head
- depends on: the files `src/update_hooks.rs`, `src/project_update.rs` and `tests/update_hooks_cli.rs` are held by other changes at the time of writing; this change takes them after they land

## Plan

1. SPEC first (this file).
2. RED: the new test file for the first install, and the edits to the two tests that pin the old words.
3. `update_hooks::plan_in`: a host that is not installed is planned as a first install with the same host plan, gate, conflict handling, digest, receipt and revert; the profile hosts decide the offer.
4. `project_update`: the reason sentence without the two commands.
5. Screen rows and goldens for the new `create` rows.
6. The unnarrowed `check affected` set, then an independent audit (tier 1).

## Risks

- A project that uses only one host would get the other one offered: the profile's hosts decide, and the preview shows every host before anything is written.
- Hooks written into a project the user did not expect: nothing is written without the confirmation, and the receipt reverts a first install by removing the files it created.
- The goldens of the update screen change: the new rows are shown for a project with no hooks, and the golden update is part of this change.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | a host is offered when the profile lists it or the project has no profile | `project_hosts` already treats an unlisted host as not offered | offer both hosts whatever the profile says |
| D-2 | the pick is the existing preview and one confirmation; no per-host selector | the user pick names the flow (preview, receipt, revert), not a selector | add a selector later |
| D-3 | OMP hooks stay outside this change | the user pick names the Claude and Codex hooks | add OMP in a later change |

## Result

(written at SYNC)
