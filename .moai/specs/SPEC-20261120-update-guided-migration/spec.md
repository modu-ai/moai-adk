---
id: SPEC-20261120-update-guided-migration
title: "moai update shows one grouped preview, asks one proceed question, then migrates in a fixed order with plain progress"
status: draft
revision: 1
created: 2026-11-20
---

# moai update shows one grouped preview, asks one proceed question, then migrates in a fixed order with plain progress

## Intent

지금 `moai update` 는 일을 여러 세계로 나눠 해요. 셸의 `moai update` 는 HOME 층(스킬·에이전트·routing·style·명령 이름·현재 프로젝트 훅·키)만 일곱 부분 묶음 하나(`--confirm`)로 적용하고, Go 잔재 정리·Go HOME 정리·이름 전환은 각자 토큰으로 따로 돌려요. HOME 폴더 이동(`~/.moai-cli → ~/.moai`)은 `moai update` 가 아예 부르지 않고(`home-install-*`), 프로젝트 쪽(`.moai-cli → .moai`, Go 설정·상태, Go 훅 → Rust 훅)은 `moai project-update-plan` 이 단계별 명령을 인쇄해서 `/moai update` 세션이 하나씩 돌려요. 사용자는 "무엇이 바뀌는지 한 번에 보고, 한 번 답하고, 끝까지 가는" 길이 없어요.

이 SPEC 은 사용자 요청 하나를 설계해요: `/moai update`(와 셸 `moai update`)가 미리보기를 만든 바로 뒤에 진행 여부를 한 번 묻고, 승낙하면 이전을 빠르고 안내하며 끝까지 해요.

1. 감지는 한 번에, 읽기만 해요(survey). 업데이트가 바꿀 수 있는 모든 것이 한 번에 모여요.
2. 미리보기는 쉬운 말로, "바뀜 / 그대로 / 선택 필요" 세 묶음이에요. 묶음마다 개수, 보호 여부, 되돌리는 명령(또는 "되돌리는 명령 없음")이 보여요.
3. 미리보기 바로 뒤에 질문은 하나예요: 추천대로 진행(기본) / 묶음별로 고르기 / 멈추기.
4. 적용은 고정된 순서로, 단계마다 진행 줄(n/N, 끝남/건너뜀/실패와 이유)을 보이고 첫 실패에서 멈춰요. 실패하면 그 단계의 되돌리는 명령이 나와요.
5. 마지막 요약은 바뀐 것, 그대로 둔 것, 되돌리는 명령(거꾸로 순서), 다음 명령이에요. 다시 돌리면 "적용할 것 없음"이에요.

이 SPEC 의 첫 조각(slice 1)은 설계와 감지 통합 표(`detection-table.md`)만이에요. 코드는 바뀌지 않아요. 표는 팁 `6a8fa9cacf4634f772b3c13dc628d0c1e4ec7450` 의 코드를 읽은 결과이고, 읽은 사실과 빈틈(G1~G12)이 거기 있어요.

## Non-goals

- 이 조각에서의 코드 변경. 설계와 표뿐이에요.
- 새 이전 로직. HOME 이동, 프로젝트 상태 폴더 이동, Go 훅 교체는 각각 기존 계획/적용/되돌리기 코드를 그대로 써요(`home-install-*`, `go-state-*`, `go-hooks-*`, `hooks-*` 등). 이 SPEC 은 그것들을 한 번에 모아 보여 주고 순서대로 부르는 층만 더해요.
- 되돌리기를 위한 삭제. 어떤 단계도 파일을 지우지 않아요(이동·보관만). 이 성질은 지켜요.
- 이전 SPEC 의 변경: HOME 이동 규칙(HOME 합치기 카드), 프로젝트 상태 폴더 이동 규칙(프로젝트 이동 카드), Rust-only 전환(전환 카드), OMP 제거(OMP 카드), 보안 훅(보안 가드 카드). 그 SPEC 들이 같은 판별 코드(`state_root`)와 같은 미리보기·토큰·영수증·되돌리기 모양을 쓰도록 맞추는 일은 각 카드의 몫이에요.
- `workflows/update.md` 와 `SKILL.md` 의 문장 쓰기. 그 파일은 다른 카드가 다시 쓰는 중이에요; 이 SPEC 은 두 곳이 같은 흐름을 말한다는 구조만 고정해요(REQ-10).
- 바이너리 자기 갱신(`self_update`). 연결되어 있지 않아요.

## Requirements

- REQ-1: When the user starts `/moai update` or the shell `moai update`, the update survey shall read in one pass without writing anything every layer the update can change, namely the package skills and agents, the routing and style blocks, the command names and the moai name, the switch step, the Go-generation HOME items, the HOME folder move, the project state-folder move, the earlier-generation project state, the Go hooks and the Rust hooks, and the OMP-managed files, and give every detected item an id, a plain-language title, a layer, a count of files and folders, a guarded flag, the ids it depends on, and either its revert command or the statement that no revert command exists.
- REQ-2 (REQ-1): IF a layer cannot be read, THEN the update survey shall name that layer as unread with its reason, still survey the other layers, and apply nothing that depends on the unread layer.
- REQ-3: The update preview shall group the items into will change, kept as is and needs your pick, and show for each group the item count, the guarded items and the revert command of each item.
- REQ-4: When the preview has been shown on a terminal, the update shall ask one question right after it with three answers, proceed with everything recommended, choose per group, and stop, and write nothing before the answer. · failure: none — a session without a terminal takes the non-interactive path of REQ-11 and never reaches the question
- REQ-5: The update shall apply a guarded item only when the user picks it explicitly and name every guarded item in the question so that the default answer does not cover it.
- REQ-6: The update shall apply the picked items in one fixed order that places the HOME folder retire before the HOME folder move, the HOME move before the other HOME items, the project backup phase before the project install phase, and the project hooks after the switch step, with a new survey after each step whose outcome changes a path or a digest of a later step.
- REQ-7: The update shall print one progress line after every step with the step number out of the total, the step name in plain words and the result done, skipped or failed with its reason.
- REQ-8 (REQ-6, REQ-7): IF a step fails, THEN the update shall stop at that step, run no later step, print the revert command of every step already done in reverse order, and leave what was already written as it is.
- REQ-9: When the last step has run, the update shall print one summary of what changed, what is kept, the revert commands in reverse order and the next command, and end in a state where a second survey reports nothing to apply. · failure: none — a failing step ends in REQ-8 and never reaches the summary
- REQ-10: The update skill documents and the shell terminal screens shall tell the same flow with the same group names, step keys and order, pinned by a test of structure keys and not of sentences.
- REQ-11: The existing update flags shall keep their meaning, namely `--confirm`, `--confirm-cleanup`, `--confirm-home`, `--confirm-switch`, `--revert-switch`, `--allow-older` and `--json`, while a session without a terminal gets the preview with one survey digest that goes stale and writes nothing when an item changes.
- REQ-12: The update shall never delete a file, with every move going to the retire, backup or cleanup folder that its own step already uses.

## Acceptance

- AC-1 (REQ-1, REQ-3): on a temporary HOME with a Go-generation `.moai` folder, an earlier `.moai-cli` folder and a project with `.moai-cli` and Go hooks, the survey names every item of the detection table in the three groups, each with a revert command or the no-revert statement, and writes nothing.
  verify: `cargo test --locked --offline --test update_guided_migration_cli survey_reads_every_layer`
- AC-2 (REQ-2): a layer that cannot be read is named unread with its reason, the other layers are still surveyed and nothing that depends on it is applied.
  verify: `cargo test --locked --offline --test update_guided_migration_cli unread_layer_is_named`
- AC-3 (REQ-4): after the preview on a terminal exactly one question is asked, its answers are the three named ones, and answering stop writes nothing.
  verify: `cargo test --locked --offline --test update_guided_migration_cli one_question_stop_writes_nothing`
- AC-4 (REQ-5): the default answer applies the unguarded items only, and a guarded item is applied only when it is picked by its name.
  verify: `cargo test --locked --offline --test update_guided_migration_cli guarded_only_when_picked`
- AC-5 (REQ-6, REQ-7): accepting migrates every picked item on the temporary HOME in the fixed order and prints one progress line per step with the number out of the total.
  verify: `cargo test --locked --offline --test update_guided_migration_cli accept_migrates_in_order_with_progress`
- AC-6 (REQ-8): a forced failure of one step stops the run, runs no later step, prints the revert commands of the steps done in reverse order, and leaves the written files as they are.
  verify: `cargo test --locked --offline --test update_guided_migration_cli first_failure_stops_with_revert`
- AC-7 (REQ-9): after the migration a second run reports nothing to apply and writes nothing, and the summary names the next command.
  verify: `cargo test --locked --offline --test update_guided_migration_cli rerun_reports_nothing`
- AC-8 (REQ-10): the update workflow document, the skill entry and the terminal screens carry the same group names, step keys and order.
  verify: `cargo test --locked --offline --test update_guided_migration_cli skill_and_screen_share_structure_keys`
- AC-9 (REQ-11): the existing flags behave as in the tip and a digest of the survey that no longer matches refuses to write.
  verify: `cargo test --locked --offline --test update_guided_migration_cli flags_keep_meaning_stale_digest_refuses`
- AC-10 (REQ-12): after the whole migration on the temporary HOME no file of the starting tree is missing, and each moved item is in its own retire, backup or cleanup folder.
  verify: `cargo test --locked --offline --test update_guided_migration_cli nothing_is_deleted`
- AC-11 (REQ-1, REQ-3): the detection table of this SPEC names every step that the tip's update code detects, with the group, the guard, the order and the revert command, and this SPEC passes the spec check.
  verify: `moai spec check SPEC-20261120-update-guided-migration`

## Scope

- slice 1 (this change), write: `.moai/specs/SPEC-20261120-update-guided-migration/**` and nothing else; read-only: every other file, the real HOME.
- later slices (their own card scope, taken when they start): `src/update.rs`, `src/update_tux.rs`, `src/update_survey.rs` (new), `src/main.rs` (flag wiring only), `tests/update_guided_migration_cli.rs` (new), `tests/fixtures/ui/**` (goldens), and after the other card has merged `skills/moai/workflows/update.md` and `skills/moai/SKILL.md` with the source manifest.
- tier 2 (HOME, update): every code slice needs an independent audit.
- depends on, in this order: the HOME-merge SPEC (HOME move), the project state-folder move SPEC, the Rust-only switch SPEC, the OMP removal SPEC, the security guard slice that moves Go hooks to Rust hooks, and the card that rewrites `workflows/update.md`. This SPEC reads their code at the tip; a slice that lands later than them follows their final form.

## Plan

1. Slice 1: this SPEC and `detection-table.md` (written by reading the tip's code; nothing run).
2. Slice 2: RED end-to-end test on a temporary HOME (AC-1, AC-2), then `update_survey`: one read-only call that assembles the existing plans (user layer, Go HOME retire, HOME move, cleanup, switch, project update plan) into items, groups, order and one survey digest, and a `--json` envelope that carries it.
3. Slice 3: the question, the fixed-order apply loop with progress lines and the first-failure stop (AC-3 through AC-7, AC-9, AC-10); the guarded rule (D-3).
4. Slice 4: skill text and terminal screens tell the same flow (AC-8), after the update-workflow card has merged; the source manifest.
5. Slice 5: items that wait for other cards: OMP-managed files listed as no longer managed (after the OMP removal), Go hook entries after the security guard slices; the gaps G4 and G6 of the table.
6. Each code slice: the unnarrowed READY set, then an independent audit.

## Risks

- The HOME move and the project move are not called by `moai update` today (table H2, P2); the HOME move also refuses to write when a live session is registered in this clone, and the session that runs `/moai update` may itself be registered. D-1 decides where the move runs.
- The project layer has no apply command of its own: each step prints its command and a session runs it (G2). A survey that spans projects must not guess the project: it is the working project of the session, as today.
- A step changes a path (HOME move) or a digest (go-state, hooks) that later steps used: the run surveys again at the points REQ-6 names, and a stale digest writes nothing.
- Several steps have no revert command (table G4: skill copies, command names, `codex-config`, `codex-skill`): the summary says so in words and never invents one.
- `go-hooks-revert` can delete a file that had no archived copy, against its own "Deletes nothing." text (table, project part c4): REQ-12 needs this checked in slice 3 before the step is promised as non-deleting.
- The calls the preview reads are many (one per layer); the survey keeps each plan's own digest and binds them (REQ-11), so a plan that changes between the preview and the answer makes the digest stale.
- OMP is still a live host at the tip; the OMP removal card is not in it. The table lists both states.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | proposed: the HOME move runs inside the new apply loop as its own step (H1 retire, then H2 move), through the existing `home-install-apply`; it is guarded | the card asks one flow to cover `~/.moai-cli → ~/.moai`; the move code exists and needs only the loop to call it | leave the move as its own command and list it as "needs your pick" with its command |
| D-2 | proposed: the shell `moai update` covers the HOME layer and the working project's items it can read; project steps that print commands run in the loop one after another, each with a fresh preview | the card says the shell TUI and the skill tell the same story | the shell stops at the HOME layer and prints the project route as next command |
| D-3 | proposed: Enter and the default answer apply the unguarded items; guarded items (HOME retire, HOME move, vault key import, project state-folder move, cleanup) are named in the question and apply only on an explicit pick, either all of them in one named answer or one by one in the per-group choice | the card: a guarded item is covered by one answer only when picked explicitly | make guarded items part of the default |
| D-4 | proposed: the non-interactive path is a new flag that takes the survey digest (working name `--confirm-all`); the old flags stay as they are | REQ-11 keeps existing scripts working | pick another name before slice 3 |
| D-5 | proposed: the progress is a plain list of lines after each step, not a spinner | the update has no progress indicator today (table G5) | add a spinner later |

All five need the leader's confirmation before slice 2 starts; none touches money, law, a contract, a secret or real data, so each is reversible (case 3: the leader decides).

## Result

(written at SYNC)
