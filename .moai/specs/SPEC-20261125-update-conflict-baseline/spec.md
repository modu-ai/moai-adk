---
id: SPEC-20261125-update-conflict-baseline
title: "moai update 가 패키지와 같은 스킬 파일을 거짓 conflict 로 남기지 않고, 화면이 남은 일을 사실대로 말해요"
status: draft
revision: 3
created: 2026-11-25
branch: spec/update-conflict-baseline-r3-20261125
---

# moai update 가 패키지와 같은 스킬 파일을 거짓 conflict 로 남기지 않고, 화면이 남은 일을 사실대로 말해요

## Intent

사용자가 빌드 26 으로 `moai update` 를 돌렸더니 스킬 파일 네 개(`references/flows/next-flow.md`, `references/source-manifest.json`, `workflows/project.md`, `workflows/work.md`)가 conflict 로 남아 갱신되지 않았어요. 리더의 조사 에이전트가 소스 `310eccd051` 에서 읽은 원인은 이래요(이 일의 배차 노트, 2026-11-24T22:52~22:55Z). 이 SPEC 은 그 위에서 코드를 다시 읽어 확인한 것만 적어요.

- 세 갈래 판정 `apply::decide(baseline, current, desired)`(`src/apply.rs:47`)는 `current == desired || desired == baseline` 이면 그대로 두고, `current == baseline` 이면 덮어쓰고, 나머지는 conflict 예요. 기준선(baseline)은 설치 영수증이 소유한 파일의 다이제스트(`owned`)예요.
- 갱신이 영수증의 `owned` 를 앞으로 옮기는 것은 파일을 쓰거나 입양한 때뿐이에요(`src/update.rs:2916-2921`, `2962`). 디스크가 이미 패키지와 같아서 "unchanged" 가 된 파일은 옛 기준선을 쥐고 있어요. 쓸 파일이 하나도 없으면 그 호스트는 영수증을 건드리지 않고 지나가요(`src/update.rs:2809-2834`). 그래서 다음 세대 패키지가 그 파일을 바꾸면 `current != baseline`, `current != desired`, `desired != baseline` 이 되어 conflict 가 영원히 남아요.
- 이전 세대 패키지와의 대조(`earlier_generations`)는 영수증이 소유하지 않은 파일에만 쓰여요(`src/update.rs:1380-1393`). 소유한 파일이 이전 세대의 패키지 바이트와 같아도 conflict 로 남아요.
- 사용자의 파일은 넷 다 사용자 수정이 아니었어요(둘은 `.24` 패키지와 바이트 동일, 둘은 저장소 커밋과 동일).
- 화면은 conflict 를 호스트 상태가 `conflict` 일 때만 세요(`src/update_tux.rs:2015-2030`). 다른 파일이 써져서 호스트가 `applied` 이면 남은 conflict 는 화면에 없어요.
- 이 밖에 화면의 사실과 맞지 않는 곳이 네 군데 있어요: 링크 폴더의 routing 줄이 "skipped" 로 보이고(`src/update_tux.rs:991-997`, 동작은 `linked`, `src/home_routing/rules.rs:440`), 옛 `rules/moai/flow.md` 의 "직접 지우세요" 안내는 detail 에만 있고 화면에 없고(`src/home_routing/rules.rs:521-534`), 켜지 않은 switch 의 다음 명령이 Applied 에 없고(`src/update_tux.rs:2157` 은 `switch_done` 만), 되돌리기 안내가 `<HOME>` 과 `<reference>` 를 글자 그대로 찍어요(결과 JSON 의 `revert` 문장은 `src/update.rs:2467-2471` 이 만들고, 화면이 그대로 보여 줘요. 그 JSON 문장은 기계가 읽고 테스트가 두 자리표시자를 채워 실행하므로 바꾸지 않고 화면만 풀어 써요).

개정 3 의 이유: 개정 2 를 실제 HOME 에 적용해 보니 결함 둘이 나왔어요. (1) 백업이 퇴역한 `~/.moai-cli/backups/` 아래에 새로 쓰였어요. 쓰는 곳 일곱이 모두 `home_root::path(home)` 에 `backups` 를 붙이는데, 이 기계의 HOME 은 `~/.moai` 가 마커 없이 항목 28개를 가진 `unknown` 이고 `~/.moai-cli` 도 있어서 그 함수가 `~/.moai-cli` 로 풀려요(`home-state-plan` 으로 읽어 확인했어요; 임시 HOME 은 `.moai` 로 풀려서 테스트가 못 잡았어요). (2) 옛 `rules/moai/flow.md` 를 지운 뒤 빈 `rules/moai/` 폴더가 남아요.

## Interview

- Q1: 백업을 어디에 둘지 → A: 사용자가 "백업만 ~/.moai 로 (Recommended)" 를 골랐어요(2026-11-25T05:00:41Z, 선택 질문, 리더가 카드 노트에 기록). 이미 `~/.moai-cli/backups/` 에 있는 것은 그대로 두고 읽기만 해요.
- Q2: 이 기계의 HOME 상태 폴더가 왜 `~/.moai-cli` 로 풀리는지 → A: 읽기 전용 `home-state-plan` 이 `~/.moai` 를 `unknown`(항목 28개, 마커 없음)으로 읽고 `~/.moai-cli` 도 있어서 `resolves_to_moai` 의 표가 `~/.moai-cli` 를 골라요(2026-11-25, 레인이 읽어서 확인).
- Q3: 옛 `rules/moai/flow.md` 를 지운 뒤 남는 빈 폴더를 치울지 → A: 비어 있을 때만 치워요. 리더 지시이고 삭제라서, 사용자 승인은 이 SPEC 의 digest 승인에 포함해 받아요(Decisions D-8).

## Non-goals

- 개정 3: 이미 `~/.moai-cli/backups/` 에 있는 백업을 옮기거나 지우거나 고치지 않아요(읽기만). HOME 상태 폴더 판정(`resolves_to_moai`)은 바꾸지 않아요. `~/.moai-cli` 아래에 새 파일을 만들지 않아요.
- 개정 2: (h) conflict 를 백업 뒤 패키지 것으로 바꾸는 플래그와 (i) 옛 `rules/moai/flow.md` 의 백업 뒤 삭제를 REQ-10…REQ-16 으로 담아요. 사용자가 두 선택을 골랐어요(Decisions D-3, D-4). 되돌리기 통합은 하지 않아요: 백업 사본을 되돌려 놓는 것이 길이고, 화면이 그 폴더를 말해요.
- 사용자가 고친 파일을 플래그 없이 덮는 일은 개정 2 에서도 없어요(REQ-12).
- (j) HOME 상태 폴더 `~/.moai-cli` → `~/.moai` 이름 바꾸기(다른 SPEC 의 일).
- (g) 머리줄에 도는 바이너리 버전 병기: 하지 않아요. 코드를 읽으니 머리줄의 `package` 는 이미 도는 빌드예요(`target_version` 은 `crate::BUILD_VERSION`, `src/update.rs` 의 계획 만드는 곳). 헷갈리는 쪽은 `installed` 인데, 쓴 파일이 없어 영수증이 새로 안 써진 실행에서 마지막 영수증의 버전에 머물러요. 이 SPEC 의 REQ-1·REQ-2 가 baseline 만 움직일 때도 영수증을 써서 그 머무름을 줄여요. 그래도 남는 혼동이 있으면 사용자가 본 화면을 근거로 따로 다뤄요.
- 사용자가 고친 파일을 덮어쓰거나 conflict 를 조용히 푸는 일. 이전 세대 어느 패키지와도 같지 않은 바이트는 conflict 로 남아요.
- Desktop 이 읽는 설정 폴더 조사와 미리보기의 설정 폴더 선택(같은 카드의 기준 1~3). 이 SPEC 은 그 기준을 만들지도 닫지도 않아요.

## Requirements

- REQ-1: When an update plans an owned skill file whose bytes equal the package file but differ from its baseline, the confirmed update shall record those bytes as the file's baseline in the host receipt. · failure: none — 파일 바이트는 한 바이트도 쓰지 않고 영수증의 소유 목록만 앞으로 가요.
- REQ-2: When the only skill files that change a host's receipt are baselines, the confirmed update shall still write that host's receipt. · failure: none — 쓸 파일이 없는 호스트도 영수증이 갱신돼요.
- REQ-3: When an owned skill file differs from its baseline and its bytes equal the same file in an earlier generation's package, the update shall plan the file against that copy as its baseline. · failure: none — 이전 세대의 복사본은 사용자 수정이 아니라 우리가 쓴 파일이에요.
- REQ-4: When an owned skill file matches neither its baseline, the package file nor any earlier generation's file, the update shall plan it as a conflict and write nothing to it. · failure: none — 사용자의 수정은 지금처럼 보존돼요.
- REQ-5: When the human Applied screen is drawn for a host with files in conflict, it shall name the number of conflicts and their paths, whether or not other files of the host were written. · failure: none — 값이 없으면 줄이 없어요.
- REQ-6: When a routing target is a link to the file that moai keeps in the default Claude folder, the human screen shall say the file is linked instead of skipped. · failure: none — 동작 `linked` 의 화면 말이 생겨요.
- REQ-7: When a routing host's flow rule plan carries a detail, the human preview shall show that detail as a line of its own. · failure: none — 옛 flow 규칙이 남았다는 안내(파일 이름과 "left in place for you to delete")가 이 detail 에 들어 있고, detail 이 없으면 줄이 없어요.
- REQ-8: When a planned switch script was not confirmed in this run, the human Applied screen shall show the command that confirms it. · failure: none — 이미 적용된 switch 는 지금처럼 되돌리기 줄만 보여요.
- REQ-9: When the human Applied screen shows a revert command, it shall show the real home folder in place of the home placeholder and a line that says what the permission reference is. · failure: none — 결과 JSON 의 `revert` 문장은 그대로예요.

- REQ-10: When an update plans owned skill files in conflict, the host plan shall carry a conflict replacement offer naming each conflicting file, the backup folder and the flag that confirms the replacement. · failure: none — 충돌이 없으면 제안이 없고 계획의 나머지는 그대로예요.
- REQ-11: When an update is confirmed with the flag of a host's conflict replacement offer, the update shall copy every conflicting file of that host into the backup folder before it writes the package bytes over that file. · failure: none — 복사가 하나라도 실패하거나 계획 뒤 파일이 바뀌었으면 어떤 파일도 덮지 않고 멈추며 모든 파일이 그대로예요.
- REQ-12: When an update is confirmed without the flag of a conflict replacement offer, the update shall leave every conflicting file of that host as it is. · failure: none — 개정 1 의 동작 그대로예요.
- REQ-13: When the human preview shows files in conflict, it shall show the command that replaces them after a backup. · failure: none — 충돌이 없으면 그 줄이 없어요.
- REQ-14: When the routing plan finds the earlier flow rule file at its earlier path and the file's digest equals a digest an earlier routing receipt recorded as written for that path, the plan shall list the file as removed after a backup. · failure: none — 다르면 사용자가 고친 파일이라 지금처럼 이름만 말하고 그대로 둬요.
- REQ-15: When an update with a removal in its plan is confirmed, the update shall copy the earlier flow rule file into a backup folder and then remove it. · failure: none — 복사가 실패하거나 계획 뒤 파일이 바뀌었으면 지우지 않고 파일이 그대로예요.
- REQ-16: When the earlier flow rule file was removed, the Applied screen shall name the backup folder the file was copied to. · failure: none — 지우지 않았으면 그 줄이 없어요.
- REQ-17: When an update plans a backup of conflicting, adopted, replaced or removed files, the host plan shall name as the backup root the `backups` folder of the HOME's `.moai` folder. · failure: none — 미리 보기가 말한 자리가 사용자가 승인한 자리예요.
- REQ-18: When a confirmed update makes a backup, the update shall create the backup folder under the `backups` folder of the HOME's `.moai` folder whichever folder the HOME's state resolves to. · failure: none — 확정한 실행이 `.moai-cli` 아래에 새 파일을 만들지 않아요.
- REQ-19: When the HOME's `.moai` folder is a link or its verdict is go, mixed, interrupted or refuse, the update shall make no backup there and leave every step that needs the backup undone. · failure: none — 백업 없이 파일을 덮거나 지우지 않고, 백업이 필요 없는 단계는 그대로 해요.
- REQ-20: When the update leaves a step undone for want of a usable backup root, the screen shall name the step and the reason. · failure: none — 그 단계가 없으면 줄이 없어요.
- REQ-21: When a backup folder of an earlier update lies under the HOME's earlier state folder, the revert and recovery readers shall read it without changing it. · failure: none — 옛 백업은 옮기지도 지우지도 않아요.
- REQ-22: When the update removes the earlier flow rule file and the folder that held it is empty afterwards, the update shall remove that empty folder. · failure: none — 비어 있지 않으면 그 폴더도 그 안의 파일도 그대로예요(폴더만 지우고 안을 지우지 않아요).
- REQ-23: When the empty folder of the earlier flow rule file was removed, the Applied screen shall name that folder. · failure: none — 지우지 않았으면 그 줄이 없어요.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 HOME 에서 영수증이 옛 파일을 소유하고 디스크는 패키지와 같을 때, 확정한 update 가 스킬 바이트를 쓰지 않고 영수증의 그 파일 소유 다이제스트를 패키지의 것으로 바꾸고, 그다음 패키지가 바뀐 파일은 conflict 가 아니라 update 로 계획돼요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac1_`
- AC-2 (REQ-3): 영수증이 소유한 파일의 디스크 바이트가 이전 세대 버전 폴더의 패키지와 같고 이번 패키지와 다르면, 미리보기의 그 파일은 conflict 가 아니라 update 이고 확정하면 패키지 바이트가 돼요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac2_`
- AC-3 (REQ-4): 기준선과 패키지와 어느 이전 세대와도 다른 바이트의 소유 파일은 conflict 로 남고 확정해도 바이트가 그대로예요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac3_`
- AC-4 (REQ-5): `applied` 호스트에 conflict 파일 둘이 있으면 Applied 화면이 그 수와 두 경로를 한국어와 영어로 말하고, conflict 가 없으면 그 줄이 없어요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac4_`
- AC-5 (REQ-6, REQ-7): 링크된 routing 줄은 "linked" 로, 옛 flow 규칙이 남은 미리보기는 그 파일 이름과 "left in place for you to delete" 가 든 detail 을 화면 줄로 보여 주고 링크된 줄에 "skipped" 는 없어요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac5_`
- AC-6 (REQ-8, REQ-9): 확정하지 않은 switch 의 다음 명령이 Applied 화면에 있고, 되돌리기 줄에 실제 HOME 이 있고 `<HOME>` 글자는 없으며 `<reference>` 가 무엇인지 한 줄로 말해요. 결과 JSON 의 `revert` 문장은 그대로예요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac6_`
- AC-8 (REQ-10, REQ-13): 임시 HOME 에서 사용자가 고친 소유 파일 하나가 conflict 일 때 미리보기 JSON 의 호스트 계획에 그 파일, 백업 폴더, 확정 플래그가 있고, 사람 화면이 확정 명령을 보여 줘요. conflict 가 없으면 제안도 줄도 없어요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac8_`
- AC-9 (REQ-11, REQ-12): 같은 상태에서 플래그 없이 확정하면 그 파일의 바이트가 그대로예요. 플래그와 함께 확정하면 백업 폴더의 사본이 원래 바이트이고, 파일은 패키지 바이트이며, 영수증이 패키지의 다이제스트를 소유해요. 복사가 실패하게 한 실행(`MOAI_TEST_FAULT=conflict_backup_copy`)은 어떤 파일도 덮지 않아요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac9_`
- AC-10 (REQ-14, REQ-15, REQ-16): 임시 HOME 에 옛 라우팅 영수증(옛 경로의 흐름 규칙을 쓴 항목과 그 다이제스트)과 그 바이트 그대로의 `rules/moai/flow.md` 가 있으면 미리보기가 삭제 대상으로 말하고, 확정하면 백업 사본이 생기고 파일이 없어지며 Applied 화면이 백업 폴더를 말해요. 바이트가 다른 파일은 그대로이고 이름만 말해요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac10_`
- AC-11 (REQ-17, REQ-18): HOME 에 `.moai-cli` 폴더가 있어서 HOME 상태 폴더가 그쪽으로 풀리는 임시 HOME 에서, conflict 와 입양과 옛 flow 삭제 각각의 미리 보기 JSON `backup_root` 가 `<HOME>/.moai/backups` 이고, 확정한 실행의 백업 폴더가 그 아래에 생기며 `.moai-cli` 아래에는 `backups` 가 생기지 않아요. 상태 폴더가 `.moai` 로 풀리는 임시 HOME 은 전과 같아요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac11_`
- AC-12 (REQ-19, REQ-20): `.moai` 가 링크인 임시 HOME 과 Go 마커가 있는 임시 HOME 에서 플래그와 함께 확정해도 conflict 파일과 옛 flow 파일의 바이트가 그대로이고 어느 곳에도 백업 폴더가 생기지 않으며, 화면이 그 단계와 이유를 한국어와 영어로 말해요. 백업이 필요 없는 단계는 적용돼요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac12_`
- AC-13 (REQ-21): `.moai-cli/backups/<폴더>` 에 있는 옛 백업을 가리키는 영수증의 되돌리기와 복구가 그 백업을 읽고, 실행 뒤 그 폴더의 모든 파일이 바이트까지 그대로예요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac13_`
- AC-14 (REQ-22, REQ-23): 옛 flow 삭제를 확정한 임시 HOME 에서 `rules/moai/` 가 비면 그 폴더가 없어지고 Applied 화면이 그 폴더를 말해요. 같은 폴더에 다른 파일이 있으면 폴더와 그 파일이 그대로이고 그 줄이 없어요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli ac14_`
- AC-7 (REQ-1…REQ-23): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test update_adopt_cli`
  verify: `cargo test --locked --offline --test update_cli`
  verify: `cargo test --locked --offline --test update_moved_receipts`
  verify: `cargo test --locked --offline --test update_tux_git_hooks`
  verify: `cargo test --locked --offline --test update_linked_config_files`

## Scope

- write: `.moai/specs/SPEC-20261125-update-conflict-baseline/spec.md`, `src/update.rs`, `src/update_tux.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs`, `src/main.rs`(확정 플래그 한 줄), `src/home_root.rs`(개정 3: 백업 뿌리 함수), `src/recovery.rs`(개정 3: 옛 백업 읽기), `tests/update_conflict_baseline_cli.rs`, `tests/fixtures/update_tux/*.json`, `tests/fixtures/ui/**/update-*.golden`
- read-only: `src/apply.rs` (세 갈래 판정은 그대로 둬요), `src/install.rs`
- forbidden: 진짜 HOME 의 Claude 설정 폴더와 HOME 상태 폴더(테스트는 읽기도 임시 HOME 에서만 해요), 설치된 moai 바이너리, 다른 프로젝트, 영수증 스키마, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1…AC-6 을 임시 HOME 과 이 빌드 바이너리의 복사본으로(`tests/update_adopt_cli.rs` 의 `install`·`home` 방식) 돌려요. 화면 테스트는 `tests/fixtures/update_tux/*.json` 을 고쳐 `update_tux::screen` 에 넣어요. 기준에서는 맞는 이유로 실패해야 해요.
2. `src/update.rs`: 계획 단계에서 소유 파일의 baseline 을 이전 세대 복사본과 대조해 올리고(REQ-3), 현재가 패키지와 같은데 baseline 이 다르면 계획에 baseline 과 current 를 그대로 싣고(이미 FilePlan 이 둘을 가져요), 확정 단계가 그런 파일을 `owned` 에 모아 쓸 파일이 없어도 영수증을 한 번 써요(REQ-1, REQ-2).
3. `src/update_tux.rs`: 남은 conflict 이름, linked 낱말, 옛 flow 안내 줄, switch 다음 명령(Applied 결과 JSON 에 `switch` 부분이 함께 실려요), 되돌리기 줄을 풀어 쓰기(REQ-5…REQ-9). `src/update.rs` 의 JSON `revert` 문장은 안 바꿔요.
4. 지킴이마다 뮤턴트, 이웃 테스트, READY 에서 `check affected` 한 번.
5. 개정 2 (h): 계획 단계가 conflict 파일마다 `ConflictReplace` 제안(파일, 계획 때의 다이제스트, 백업 폴더 이름, 토큰 `conflict-replace:<host>:<hex>` 와 `--confirm-conflict-replace <토큰>` 플래그)을 싣고, `src/main.rs` 가 그 플래그를 받아 적용 단계로 넘겨요. 적용은 입양 백업(`adopt_backup`)과 같은 모양의 백업(종류 `conflict`, 폴더 `<UTC>-<host>-conflict`)을 사본 검증까지 만든 뒤에야 파일을 덮고, 영수증은 패키지 다이제스트를 소유해요.
6. 개정 2 (i): 라우팅 계획이 옛 경로 파일의 다이제스트를 옛 영수증의 `flow` 항목이 기록한 `after_sha256` 들과 대조해 삭제 대상으로 올리고, 확정이 백업(`<UTC>-claude-flow-rule`)을 만든 뒤 지워요. 다이제스트가 맞지 않으면 개정 1 처럼 이름만 말해요.
7. 개정 3 (RED 먼저): `~/.moai-cli` 가 있는 임시 HOME(AC-11…AC-14)을 만들어 지금의 코드가 백업을 `.moai-cli/backups` 에 쓰는 것을 맞는 이유로 실패시켜요. 임시 HOME 의 기본 모양은 `.moai` 로 풀려서 이 결함이 안 보이니, 테스트 도우미가 `.moai-cli` 를 먼저 만들어요.
8. 개정 3 `src/home_root.rs`: `backup_root(home) -> Result<PathBuf, String>` 한 곳이 `<HOME>/.moai/backups` 를 돌려주고, `.moai` 가 링크이거나 판정이 go·mixed·interrupted·refuse 면 거절해요(`.moai` 가 없으면 `backups` 만 만들고 마커는 만들지 않아요). `src/update.rs` 의 쓰는 곳 다섯(계획 셋, 쓰기 둘)과 `src/home_routing/rules.rs` 의 둘(계획, 쓰기)이 이 함수만 써요. 거절이면 그 단계를 하지 않고 이유를 결과에 실어 화면이 말해요(REQ-19, REQ-20).
9. 개정 3 `src/recovery.rs`: 백업 폴더의 부모가 새 뿌리이거나 옛 `home_root::path(home)/backups` 일 때 읽어요. 읽기만 하고 쓰지 않아요(REQ-21).
10. 개정 3 `src/home_routing/rules.rs`: 옛 flow 파일을 지운 뒤 `rules/moai/` 가 비어 있으면 그 폴더만 `remove_dir` 로 지워요(재귀 삭제 없음). 결과에 폴더 경로를 실어 화면이 말해요(REQ-22, REQ-23).
11. 개정 3 위험 경로 실행: READY 는 임시 HOME 에서 실제로 돌린 확정 실행(상태 폴더가 `.moai-cli` 로 풀리는 HOME, 링크된 `.moai`, 링크된 백업 폴더, 비어 있지 않은 `rules/moai/`)의 결과를 담아요.

## Risks

- 개정 3: 실제 HOME 의 `~/.moai` 는 마커 없는 `unknown` 이고 다른 도구의 항목이 있어요. 거기에 `backups/<UTC>-claude-…` 폴더를 새로 두는 것이 `state_root` 의 판정(`holds_go_state_receipt` 는 UTC 이름의 폴더 안 `go-state-receipt.json` 만 봐요)과 부딪치지 않는지 AC-11 과 이웃 테스트(`state_root` 의 단위 테스트)로 보여요. `~/.moai/backups` 가 이미 있으면 그 폴더를 그대로 쓰고 안의 것을 건드리지 않아요.
- 개정 3: `~/.moai` 가 링크이거나 Go 의 폴더(go·mixed·interrupted·refuse)면 백업을 거절하고 그 단계를 하지 않아요(D-7). 그러면 입양과 conflict 교체와 옛 flow 삭제가 그 HOME 에서는 안 일어나요. 사용자가 원하면 D-7 의 다른 길을 골라요.
- 개정 3: 한 백업 뿌리의 계획 문구(`backup_root`)가 바뀌면 계획 digest 가 바뀌어요. 같은 입력이면 같은 digest 가 나오는 성질은 유지돼요.

- 영수증의 소유 목록만 바뀌고 파일 바이트는 안 바뀌는 쓰기가 새로 생겨요. 그 쓰기는 영수증을 pending 으로 시작하지 않아도 되는지(쓸 파일이 없으니 되돌릴 것도 없어요)를 구현에서 정하고, 되돌리기(`apply-recover`)가 이 영수증을 파일 쓰기로 읽지 않는지 테스트로 확인해요.
- REQ-3 의 이전 세대 대조는 이 빌드가 버전 폴더 배치 안에 있을 때만 이전 세대를 읽어요(`earlier_generations`). 배치 밖의 빌드는 이전 세대를 빌려오지 않으니 REQ-1 만 작동해요. 지금 사용자의 상태(소유한 옛 기준선, 디스크는 이전 세대와 같음)는 REQ-3 이 풀어요.
- 계획이 FilePlan 의 baseline 값을 바꾸면 `plan_digest` 가 바뀌어요. 같은 입력이면 같은 digest 가 나오는 성질은 유지돼요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 거짓 conflict 는 결정(세 갈래)을 바꾸지 않고 baseline 을 바로잡아 풀어요 | `decide` 는 그대로 옳아요: baseline 이 낡았을 때만 틀려요. 결정을 약하게 하면 진짜 사용자 수정도 덮일 수 있어요 | baseline 보정을 빼요 |
| D-2 | 이전 세대 복사본과 같은 소유 파일은 사용자 수정이 아니에요 | 입양 SPEC(`SPEC-20261123-update-adopt-unowned`)과 같은 증거예요. 우리 도구가 쓴 바이트예요 | REQ-3 을 빼요 |
| D-3 | **(h) 결정됨**: 사용자가 (2) "새 옵션이 있을 때만 conflict 파일을 백업한 뒤 패키지 것으로 덮어요" 를 골랐어요(2026-11-25, 리더 대화의 선택 질문, 리더가 카드 노트에 기록). 구현은 REQ-10…REQ-13. 원래 질문: conflict 를 "패키지 것으로 바꾸고 백업" 하는 선택을 넣을지 | 선택지: (1) 넣지 않아요, 사용자가 직접 지워요(기본, 이번 SPEC); (2) 새 플래그 `--replace-conflicts` 로 conflict 파일을 `backups/<UTC>-claude-conflict/` 에 복사한 뒤 덮어요(백업은 입양 백업과 같은 모양); (3) 미리보기 질문으로 파일마다 고르게 해요. 추천 (2) 는 쓰기 동작이 늘어나 따로 SPEC 이 필요해요 | 기본이라 되돌릴 것이 없어요 |
| D-4 | **(i) 결정됨**: 사용자가 (2) "옛 `rules/moai/flow.md` 가 옛 빌드가 쓴 바이트와 같을 때만 백업하고 지워요" 를 골랐어요(2026-11-25, 같은 선택 질문, 같은 기록). "옛 빌드가 쓴 바이트" 는 옛 라우팅 영수증이 그 경로에 쓴 것으로 기록한 다이제스트예요. 구현은 REQ-14…REQ-16. 원래 질문: update 가 옛 `rules/moai/flow.md` 를 지울지 | 선택지: (1) 지우지 않고 화면에 알려요(기본, REQ-7); (2) 바이트가 옛 빌드가 쓴 것과 같을 때만 백업 뒤 지워요; (3) 지우지 않고 새 규칙 파일이 같은 글을 이미 읽게 하는 것을 막아요. 추천 (2), 단 삭제는 `guarded` 라 사용자 선택이 먼저예요 | 기본이라 되돌릴 것이 없어요 |
| D-5 | 쓸 파일이 없는 호스트의 영수증 쓰기는 소유 목록만 바꾸고 `entries` 는 비워요 | 쓴 파일이 없는데 쓴 기록을 만들지 않아요 | 쓰기를 빼요 |
| D-6 | **(k) 결정됨**: 백업의 새 뿌리는 `<HOME>/.moai/backups` 예요. 사용자가 "백업만 ~/.moai 로 (Recommended)" 를 골랐어요(2026-11-25T05:00:41Z, 리더가 카드 노트에 기록). 구현은 REQ-17, REQ-18, REQ-21 | 선택지: 백업만 `~/.moai` 로(추천, 선택됨); `~/.moai-cli` 에 계속 써요; HOME 상태 전체를 옮겨요(개정 3 의 일이 아니에요). 이미 있는 `~/.moai-cli/backups/` 는 그대로 두고 읽기만 해요 | 새 뿌리 함수가 옛 `home_root::path(home)/backups` 를 돌려주게 해요 |
| D-7 | `<HOME>/.moai` 가 링크이거나 판정이 go·mixed·interrupted·refuse 면 백업을 거절하고 그 단계를 하지 않아요(REQ-19). 기본이에요 | `resolves_to_moai` 의 표는 Go 의 폴더를 쓰지 않고, 링크는 따라가지 않아요. 선택지: (a) 거절하고 그 단계를 하지 않아요(기본); (b) 옛 `~/.moai-cli/backups` 로 되돌려 써요(사용자가 퇴역한 폴더에 새로 쓰지 말라고 했어요); (c) 다른 폴더를 만들어요(새 규칙이 필요해요) | 거절 대신 (b) 로 바꿔요 |
| D-8 | 빈 `rules/moai/` 폴더를 치워요. 옛 flow 파일을 지운 직후에만, 폴더가 비어 있을 때만, `remove_dir` 한 번(재귀 없음)으로 해요(REQ-22). 리더의 지시이고 삭제라서(`guarded`) 사용자 승인이 이 SPEC 의 승인에 들어 있어요 | 빈 폴더는 아무것도 안 담아서 잃는 것이 없어요. 선택지: 치워요(기본); 빈 폴더를 두고 화면에 알려요 | REQ-22, REQ-23 을 빼요 |

## Result

개정 3 (REQ-17…REQ-23, 백업 뿌리와 빈 폴더; digest `sha256:96026db458aed50b2c951ff59e0dd3f7de1f5ad4a5bef1c5fce4fa0c40ec26a9`, 사용자의 "승인, D-7·D-8 기본값 (Recommended)" 선택을 리더가 카드 노트 9045 에 기록한 것을 인용해요) 구현:

- 고치기 전 RED: 새 테스트 ac11~ac14 4개와, 백업 위치를 새 뿌리로 고쳐 쓴 이웃 테스트 4개(`update_adopt_cli` 1, `update_backup_replace` 3)가 옛 코드에서 맞는 이유로 실패했어요(백업이 `.moai-cli/backups` 에 생기고, 화면에 줄이 없고, 빈 폴더가 남았어요). 지킴이 테스트인 ac13 의 거절 쪽(`ac13_recovery_does_not_read_…`)은 코드가 이미 거절해서 RED 가 없고, 뮤턴트가 보증해요.
- 원인과 고친 곳: 백업을 쓰는 곳 일곱이 `home_root::path(home)/backups` 를 써서, `~/.moai-cli` 가 있고 표지가 없는 HOME 은 `~/.moai-cli` 로 풀렸어요. 새 함수 `home_root::backup_root` 하나가 `<HOME>/.moai/backups` 를 돌려주고(`go`·`mixed`·`interrupted`·`refuse` 판정이면 `Err`), 쓰는 곳 다섯(`update.rs` 계획 셋과 쓰기 둘, `rules.rs` 계획과 쓰기)이 그것을 써요. HOME 상태 폴더의 판정(`resolves_to_moai`)은 안 바꿨어요.
- AC-11: met — `cargo test --locked --offline --test update_conflict_baseline_cli ac11_` exit 0. `.moai-cli` 로 풀리는 HOME(남의 `.moai` 가 있는 경우도)에서 conflict 교체, 입양, 옛 flow 삭제의 백업이 모두 `.moai/backups` 에 하나씩 생기고 `.moai-cli/backups` 는 안 생겨요. 상태 폴더 판정은 그대로예요.
- AC-12: met — `… ac12_` exit 0. `.moai` 가 링크이거나 Go 표지(`credentials.yaml`)가 있으면 계획이 `conflict_replace`·`adopt`·`legacy_flow` 를 내지 않고 `backup_blocked` 에 단계와 이유를 담아요. 한국어·영어 화면이 단계와 이유를 말해요. 확정해도 사용자 바이트(충돌 파일, 스킬 폴더, flow 규칙 파일)가 그대로이고 링크 바깥에는 아무것도 안 써요.
- AC-13: met, 한 가지 해석을 적어요 — `… ac13_` 둘 exit 0. 복구가 `.moai-cli/backups` 의 옛 백업을 읽어 원본을 되돌리고, 백업 폴더는 그 자리에 남고 `MANIFEST.json` 바이트와 읽기만 하는 파일이 그대로이며 `.moai` 에는 아무것도 안 생겨요. 거절 쪽: 알려진 백업 폴더 밖을 가리키는 영수증은 복구가 읽지 않고 백업과 스킬을 그대로 둬요. 해석: 복구는 원래 백업 안의 원본 트리(`moai/…`)를 제자리로 옮기므로 AC-13 의 "모든 파일이 바이트까지 그대로" 를 그 트리에는 적용할 수 없어요. 그 트리는 복구의 목적대로 옮겨지고, 나머지는 위처럼 그대로예요. 심사자가 이 해석을 봐 주세요.
- AC-14: met — `… ac14_` exit 0. 옛 flow 파일을 지운 뒤 `rules/moai/` 가 비면 치워지고(부모 `rules/` 는 남아요) 결과 JSON 이 `folder_removed` 를 싣고 Applied 화면(ko/en)이 그 폴더를 말해요. 다른 파일(`mine.md`)이 있으면 폴더와 그 파일이 남고 줄도 없어요.
- 되돌려야 했던 회귀 하나: 백업이 `.moai-cli` 밖으로 나가자 `~/.moai-cli` 자체가 링크일 때의 거절(`a_symlinked_state_directory_is_refused`)이 사라져 영수증이 링크를 따라 HOME 밖에 써졌어요. `update::refuse_backup_roots` 가 백업 폴더와 상태 폴더의 링크를 둘 다 막아요.
- 이웃: `update_conflict_baseline_cli` 14(개정 2 의 ac1~ac10 포함), `update_adopt_cli` 15, `update_backup_replace` 20, `update_route` 9 통과, clippy `-D warnings` 통과. 백업 위치를 고쳐 쓴 이웃 테스트는 `update_adopt_cli`(2곳), `update_backup_replace`(5곳)이고 이 슬라이스의 범위에 넣었어요.
- 지킴이 뮤턴트 12(각각 정확히 한 번 고치고 `git checkout` 으로 복원, 트리 깨끗함 확인, 모두 단언 실패 exit 101): `go` 판정 거절 제거 → ac12, `refuse` 판정 거절 제거 → ac12, 상태 폴더로 되돌린 뿌리 → ac11, 상태 폴더 링크 확인 제거 → `a_symlinked_state_directory_is_refused`, `remove_dir` 을 `remove_dir_all` 로 → ac14, `folder_removed` 키 제거 → ac14, 입양 취소 제거 → ac12, conflict 막힘 기록 제거 → ac12, flow 막힘 단계 이름 바꿈 → ac12, 복구 위치 확인 제거 → 처음엔 살아남아(exit 0) 거절 테스트를 더한 뒤 ac13 실패, 막힘 화면 줄 제거 → ac12, 빈 폴더 화면 줄 제거 → ac14.
- 임시 HOME 에서만 돌렸어요(빌드한 `moai-cli` 를 임시 HOME 에 실행). 진짜 `~/.claude`, `~/.moai`, `~/.moai-cli` 는 읽지도 쓰지도 않았어요.
- 하지 않은 것: `rules/moai` 가 링크일 때의 `remove_dir` 거절(`plain` 확인)은 테스트가 없어 뮤턴트도 안 돌렸어요. `remove_dir` 은 링크를 따라가지 않으므로 방어층이에요. 복구가 HOME 상태가 `.moai` 로 풀리는 HOME 에서 `.moai-cli/backups` 를 읽는 경우는 시험하지 못했어요(영수증 위치까지 옮겨야 해요).

개정 1 은 REQ-1…REQ-9, 개정 2 는 REQ-10…REQ-16 이에요. (j) 는 구현하지 않았고(사용자 선택 대기), (g) 는 근거가 코드와 어긋나 하지 않았어요.

개정 2 (digest `sha256:b2431bc1ce2e59cc10ca8b93a17e6396e8408fd8f1ef140496973bd79518cba4`, 사용자의 "승인 (Recommended)" 선택을 리더가 카드 노트에 기록한 것을 인용해요) 구현:

- AC-8: met — `cargo test --locked --offline --test update_conflict_baseline_cli ac8_` exit 0. 충돌 호스트의 미리보기 JSON 에 `conflict_replace`(파일, 백업 폴더, 토큰, `--confirm-conflict-replace` 플래그)가 있고 한국어·영어 화면이 `moai update --confirm <digest> <플래그>` 를 보여 줘요. 충돌 없는 픽스처에는 제안도 줄도 없어요.
- AC-9: met — `… ac9_` exit 0. 플래그 없이 확정하면 사용자 바이트가 그대로이고 백업 폴더가 없어요. `MOAI_TEST_FAULT=conflict_backup_copy` 실행은 파일을 안 덮고 백업 폴더도 안 남겨요. 플래그와 확정은 사본(원래 바이트)을 먼저 만든 뒤 패키지 바이트를 쓰고 영수증이 그 다이제스트를 소유해요.
- AC-10: met — `… ac10_` exit 0. 옛 영수증이 기록한 다이제스트와 같은 `rules/moai/flow.md` 는 미리보기가 삭제 대상으로 말하고 확정하면 `<stamp>-claude-flow-rule` 사본을 만든 뒤 지워요. Applied 화면이 백업 폴더를 말해요. `MOAI_TEST_FAULT=flow_rule_backup` 실행은 파일을 남겨요. 다른 바이트는 그대로이고 이름만 말해요.
- AC-7: met — 개정 2 코드에서 `update_adopt_cli` 15, `update_cli` 11, `update_moved_receipts` 8, `update_tux_git_hooks` 6, `update_linked_config_files` 9, `help_layout` 11, `update_tux_golden` 16 통과. 골든은 다시 쓰지 않았어요.
- 지킴이 뮤턴트(각각 정확히 한 번 고치고 `git checkout` 으로 복원, 트리 깨끗함 확인, 모두 단언 실패 exit 101): 토큰 확인을 `true` 로 → ac9 실패, 백업 복사 실패 주입점 제거 → ac9 실패, 제안 대상을 `conflict` 가 아닌 파일로 → ac8 실패, 영수증 다이제스트 확인 제거 → ac10 실패, 옛 flow 백업 실패 주입점 제거 → ac10 실패, 옛 flow 사본 쓰기 제거 → ac10 실패.
- 옛 flow 삭제는 영수증 항목을 만들지 않아요(사본이 되돌리는 길이에요). 삭제가 있으면 라우팅 쓰기로 세어 터미널 질문이 떠요. 새 플래그는 `src/main.rs` 8줄, `src/help.rs` 7줄을 바꿔요(리더가 허락했어요).
- 임시 HOME 에서만 돌렸어요. 진짜 `~/.claude` 와 `~/.moai-cli` 는 읽지도 쓰지도 않았어요.

개정 1 결과:

- AC-1…AC-6 met: 각 verify 줄을 적힌 그대로 돌려 모두 exit 0. 고치기 전 RED 는 ac1(baseline 이 안 움직임), ac2(conflict), ac4·ac5·ac6(화면에 줄이 없음)이 맞는 이유로 실패했고 ac3 은 지킴이라 RED 가 없어요(뮤턴트 M2 가 보증해요).
- AC-7 met: `update_adopt_cli` 15, `update_cli` 11, `update_moved_receipts` 8, `update_tux_git_hooks` 6, `update_linked_config_files` 9 통과. 이웃 실행(`update_*`, `home_update*`, `project_update*` 통합 테스트와 lib 의 update·tux 테스트 388개)에서 `update_tux_golden` 4개가 실패했는데, 화면이 되돌리기 줄의 `<HOME>` 을 실제 HOME 으로 바꿔 보여 주기 때문에 그 자리표시자를 고정해 둔 단언과 골든 다섯 개(ko/en `update-applied*.golden`)가 옛 동작을 가리킨 것이었어요. 단언은 뷰의 HOME 으로 바꾸고 골든은 `MOAI_UPDATE_GOLDEN_WRITE=1` 로 다시 만든 뒤 diff 가 그 줄들뿐임을 읽었어요. 그 뒤 22개(새 테스트 6 + 골든 16) 통과, clippy `-D warnings` 통과.
- 지킴이 뮤턴트(각각 `git checkout` 으로 복원, 트리 깨끗함 확인): baseline 이동 조건에 `&& false` → ac1 실패, 이전 세대 확인을 `ours || true` 로 → ac3 실패, 남은 conflict 줄에 `false &&` → ac4 실패, 되돌리기 줄의 `<HOME>` 풀기 제거 → ac6 실패. 모두 exit 100.
- 결과 JSON 의 `revert` 문장은 안 바꿨어요(기계가 읽고 테스트가 채워 실행해요). 바뀐 JSON 은 확정한 실행의 결과에 `switch` 부분이 확정하지 않은 때만 더 실리는 것 하나예요.
- 영수증만 움직이는 쓰기(REQ-1, REQ-2)가 새로 생겼어요. 그 호스트는 `entries` 가 비고 상태가 `unchanged`(또는 `conflict`)로 보고되고 영수증은 `committed` 예요. 플랜이 이를 `baseline N` 으로 세고 미리보기가 확정을 물어요. 계획 뒤 그 파일이 바뀌면 영수증을 쓰기 전에 `a file changed during apply; run recovery-plan` 으로 멈춰요.
- 임시 HOME 에서만 돌렸어요. 진짜 `~/.claude` 와 `~/.moai-cli` 는 읽지도 쓰지도 않았어요.
