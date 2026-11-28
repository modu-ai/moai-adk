---
id: SPEC-20261126-guard-record-repoint
title: "git 가드: 통합 브랜치가 바뀌면 낡은 기록을 미리보기와 사용자 선택으로 다시 맞추고 update 가 가드 훅을 제안하기"
status: draft
revision: 1
created: 2026-11-26
---

# git 가드: 통합 브랜치가 바뀌면 낡은 기록을 미리보기와 사용자 선택으로 다시 맞추고 update 가 가드 훅을 제안하기

## Intent

프로필의 `work_flow.integration_branch` 를 바꿔도 git 가드(통합 브랜치를 손으로 옮기지 못하게 막는 `reference-transaction` 훅)가 읽는 기록 `<공통 git 폴더>/moai/integration-branch` 는 그대로예요. 그래서 새 통합 브랜치는 아무것도 지키지 않고, 옛 브랜치만 계속 지켜져요. 이 SPEC 을 부른 카드가 mo.ai.kr 에서 읽은 모습은 이래요(카드 본문의 글이고 이 SPEC 작성자가 mo.ai.kr 을 직접 읽지는 않았어요, 미확인): `/moai project` 가 `main` 을 `NeXT` 로 바꿨는데(영수증 `project-apply-9cdfc06affdb`) 기록은 `refs/heads/main` 그대로이고 보호 훅도 없어요.

읽은 사실(NeXT 기준 `f48ea9d815`, 이 트리, 코드 읽기만 했고 실행하지 않았어요):

1. 낡는 값은 기록 파일의 브랜치 하나예요. 가드가 저장하는 설치 경로는 없어요. 훅은 실행할 때 `moai` 를 찾아요(`install/git-hooks/reference-transaction`). 카드를 보낸 쪽이 "옛 경로"라고 쓴 것은 이 기록을 뜻한 말이었다고 확인받았어요(리더 메시지, 같은 날). 그래서 이 SPEC 은 기록의 브랜치 값을 다시 맞추는 일을 다뤄요.
2. 기록은 `seed_record`/`protected()` 가 비어 있을 때만 써요(`src/card/guard.rs:229`, `:372-401`). 이미 있으면 프로필과 달라도 `present` 로 답하고 두며, 어긋남은 `card land` 의 답에서 알리기만 해요. 승인된 `SPEC-20261125-integration-guard-install` REQ-2 가 "an existing record is never rewritten" 이고, `SPEC-20261121-ref-guard` 는 "기록을 바꾸는 일"을 나중의 별도 행동으로 남겨 뒀어요(Non-goals, D-11).
3. 기록이 낡았고 훅이 이미 최신이면 설치 계획에 쓸 것이 없어 확인 토큰이 없고, 적용은 거절돼요(`src/git_hooks/mod.rs:471-478`). 그래서 기존 길로는 기록을 맞출 수 없어요.
4. `moai update` 는 `git_hooks::update::plan` 으로 훅 층을 계획하고 가드 값에 항상 거짓을 넘겨요(`src/update.rs:2076`, `:2618`, `:2764`; `src/git_hooks/update.rs:28-29`, `:129-130`). 그래서 update 는 가드 훅을 제안하지 않고, 낡은 기록에 대한 안내도 없어요.
5. `card land`, `finish`, `submit`, `drift` 등은 프로필보다 기록을 먼저 읽어요(`src/card/landed.rs:124-129`). 그래서 기록이 낡으면 착지 명령도 옛 브랜치를 향해요. 한편 Bash 경계(`src/hook_bash_boundary.rs`)는 프로필의 이름을 지켜서, 어긋나 있는 동안 두 층이 서로 다른 브랜치를 지켜요.

이 SPEC 은 두 가지를 해요. (a) 새 행동 `moai card guard-repoint`: 낡은 기록을 미리보기로 보여 주고, 사용자 선택(미리보기의 digest 와 permission reference)이 있을 때만 기록을 프로필 값으로 다시 맞춰요. 영수증과 되돌리기가 있어요. (b) `moai update` 가 가드 훅이 없을 때 그것을 제안하고, 기록이 낡았으면 훅 제안을 보류하며 안내 한 줄을 보여 줘요. 설치기와 `seed_record` 와 `card land` 는 지금처럼 기존 기록을 다시 쓰지 않아요.

## Interview

- Q1: 이 SPEC 의 첫 판(revision 1)을 승인할까요? → A: 아직 답이 없어요(사용자 선택 대기).
- Q2: 아래 Decisions 의 열린 갈림길(D-2 ~ D-9)을 각 기본값대로 확정할까요? → A: 아직 답이 없어요(사용자 선택 대기).

## Non-goals

- `/moai project`, `project-apply` 와 `skills/moai/workflows/project.md` 9단계의 변경. 프로필을 바꾸는 쪽은 이번에 고치지 않고, 낡은 기록을 알리고 맞추는 길을 따로 줘요(이미 착지한 프로젝트 설정 SPEC 의 몫과 겹치지 않아요).
- 보호 훅의 텍스트와 판정 규칙(`SPEC-20261121-ref-guard`). 훅의 바이트는 그대로예요.
- 설치기(`git-hooks-apply`)와 `seed_record` 가 기존 기록을 다시 쓰게 하는 일(승인된 REQ-2 를 바꾸는 일이에요).
- 옛 브랜치를 다시 지키는 두 참조 기록(D-6 의 선택지 b). 훅 텍스트가 바뀌는 더 큰 tier 2 변경이에요.
- 세션 시작 줄이 낡은 기록을 알리는 일(`src/hook_session_start.rs`, D-8 의 선택지).
- 링크 워크트리에서 훅이 그 워크트리 자신의 git 폴더에 설치되는 기존 동작(`core.hooksPath` 카드의 감사가 `pre_existing` 으로 표시한 것, `src/git_hooks/mod.rs:120`).
- 이 저장소의 실제 `.git/hooks`·`.git/moai`, `~/.moai`, 실제 mo.ai.kr 상태를 쓰는 일. 코드와 시험은 임시 HOME 과 임시 저장소에서만 해요.

## Requirements

- REQ-1: When `moai card guard-repoint` runs without `--apply` or `--revert` in the primary worktree of a repository whose profile names an integration branch whose full ref differs from the guard record, the card command shall answer `stale` with the record, the target ref, whether the target exists as a local branch, the hook state, `intent_in_flight`, `unguarded_after`, a plan digest and the exact apply command, and write nothing · failure: none — no record, a record equal to the profile, a profile that names no branch, or a linked worktree answers `no_record`, `current`, `no_profile_branch` or `linked_worktree` with no digest and writes nothing, and a read that fails answers a refusal with its reason and writes nothing.
- REQ-2: When `moai card guard-repoint --apply` is given the preview's `--plan-digest` and a valid `--permission-reference`, the card command shall rewrite the record to the target ref, write a receipt in the common git folder and answer the receipt, the stamp and the revert command.
- REQ-3 (REQ-2): IF the digest differs, the reference is missing or invalid, the record changed since the preview, the target is not an existing local `refs/heads/` ref, the file `moai/land-intent` exists, or the worktree is linked, THEN the card command shall write nothing and name the reason.
- REQ-4 (REQ-2): IF a step fails after the lock is taken, THEN the card command shall remove the lock and the receipt and leave the record as it was.
- REQ-5: When `moai card guard-repoint --revert <stamp>` runs and the record still equals the new bytes of that stamp's receipt, the card command shall restore the old bytes and write `<stamp>.revert.json` beside the receipt.
- REQ-6 (REQ-5): IF the record changed since, the revert receipt already exists, or the receipt is missing or unreadable, THEN the card command shall write nothing and refuse with the reason.
- REQ-7: When `moai update` plans or applies the git-hook layer in a repository that holds the guard record or whose profile names an integration branch, the update shall pass the same guard answer to the git-hook plan and to both apply sites, so the missing `reference-transaction` hook is offered and installed through `--confirm-git-hooks`.
- REQ-8 (REQ-7): IF the guard record is stale, THEN the update shall not offer the guard hook.
- REQ-9: When `moai update` previews in a primary worktree whose guard record is stale, the update shall carry one read-only notice that names the record, the profile branch and `moai card guard-repoint`, outside every digest and regardless of whether the git-hook layer has anything to write.
- REQ-10 (REQ-9): IF the repository, the record or the profile cannot be read, or the worktree is linked, THEN the update shall add no notice.
- REQ-11: When the installer (`git-hooks-apply`), `seed_record` or `card land` meets an existing record, each shall behave as before this SPEC and never rewrite it · failure: none — the four tests that pin this stay unchanged and pass.
- REQ-12: When `guard-repoint` or the update's guard offer changes what a person sees or types, the card operations document and the two git-hook operations documents shall say so · failure: none — the three documents name `guard-repoint`, the receipt folder `moai/record-repoint`, the withheld hook for a stale record, and that the installer and the seed still never rewrite the record.

## Acceptance

- AC-1 (REQ-1): with the record `refs/heads/main` and a profile that names `NeXT` in a temporary primary worktree, the preview answers `stale` with a digest and `unguarded_after`, writes nothing and creates no receipt; `current`, `no_record`, `no_profile_branch` and a linked worktree answer their state without a digest · verify: `cargo test --locked --offline --test integration_guard_repoint ac1_`
- AC-2 (REQ-2, REQ-3, REQ-4): the dangerous path itself, in a temporary HOME and a temporary repository with the real binary: after `--apply` the record names `refs/heads/NeXT`, `moai/record-repoint/<stamp>.json` exists with both byte hashes, a hand `git update-ref refs/heads/NeXT` and `git branch -f NeXT` are refused after `git-hooks-apply` while a move of `main` passes, and a wrong digest, a missing or invalid reference, a changed record, a missing target branch, a `moai/land-intent` file and a linked worktree each write nothing · verify: `cargo test --locked --offline --test integration_guard_repoint ac2_`
- AC-3 (REQ-5, REQ-6): `--revert` restores the old bytes only while the record equals the new value, and a changed record, a second revert and a missing receipt refuse and write nothing · verify: `cargo test --locked --offline --test integration_guard_repoint ac3_`
- AC-4 (REQ-7, REQ-8): `moai update --json` in a temporary repository with a record and no hook lists `reference-transaction` under `git_hooks.install`, `--confirm-git-hooks` installs it and the answer carries `guard`, a stale record offers no guard hook, and the plan and apply tokens match · verify: `cargo test --locked --offline --test update_git_hooks_cli update_offers_the_guard_hook_`
- AC-5 (REQ-9, REQ-10): the `moai update` preview of a stale primary worktree carries the notice with the command and no `git_hooks` part, the notice is outside `plan_digest`, and an unreadable profile or a linked worktree adds none · verify: `cargo test --locked --offline --test update_git_hooks_cli update_notice_names_a_stale_record`
- AC-6 (REQ-11): the installer test that pins "written once, never rewritten" still passes unchanged · verify: `cargo test --locked --offline --test integration_guard_install ac2_`
- AC-7 (REQ-12): the three operations documents name `guard-repoint`, `moai/record-repoint`, the withheld hook for a stale record, and the never-rewrite rule of the installer · verify: `cargo test --locked --offline --test integration_guard_repoint ac7_`
- AC-8 (REQ-11): the existing guard and landing tests of the card target that pin the record (`ac9_`, `ac12_`, `ac13_`, `ac22_`) still pass unchanged · verify: `cargo test --locked --offline --test card -- card_guard_cli card_land_cli`

## Scope

- write: `.moai/specs/SPEC-20261126-guard-record-repoint/spec.md`
- write: `src/card/guard.rs` (한 줄, `pub mod repoint;`), `src/card/guard/repoint.rs` (새 파일), `src/card/drift.rs` (가시성 한 단어, `reference_ok`)
- write: `src/update.rs`, `src/ops/card.rs` (공유 파일, 줄 임대로 `Action::GuardRepoint` 한 변형과 match 한 팔)
- write: `docs/design-review/appendix/operations/card.md` (공유 파일, 줄 임대로 표 한 줄), `docs/design-review/appendix/operations/git_hooks_plan.md`, `docs/design-review/appendix/operations/git_hooks_apply.md`
- write: `tests/integration_guard_repoint.rs` (새 파일), `tests/update_git_hooks_cli.rs`
- write: `tools/repo_scans.json` (공유 파일, 줄 임대 또는 `moai check fix`; 새 시험 파일을 `reads-docs` 묶음에 더해요)
- read-only: `src/git_hooks/mod.rs`, `src/git_hooks/update.rs`, `src/card/landed.rs`, `src/card/land.rs`, `src/profile.rs`, `src/update_tux.rs`, `install/git-hooks/reference-transaction`, `tests/integration_guard_install.rs`, `tests/card/card_guard_cli.rs`, `tests/card/card_land_cli.rs`
- forbidden: 이 저장소의 `.git/hooks`·`.git/moai`, `.moai/project.yaml`, `CLAUDE.md`, `src/card/mod.rs`, `tests/card/main.rs`, `tests/areas.toml`, `src/help.rs`, `src/main.rs`, `src/hook_session_start.rs`, `src/project_update.rs`, `skills/**`

## Plan

1. (지금, SPEC 폴더만) 이 SPEC 을 채우고 `moai spec check` 를 돌려요. SPEC-READY 노트에 digest 와 열린 선택을 `LANE-BLOCKED kind=decision` 으로 올려요. 승인은 사용자의 것이라 이 레인이 기록하지 않아요.
2. 승인 뒤 `moai card scope` 로 위 쓰기 범위를 더해요. 공유 파일 세 개는 `card line-lease` 로 한 번에 쓰고 커밋 직후 돌려줘요.
3. 구현은 `fix/hooks-path-record` 브랜치(`guard.rs:234-247` 을 고쳐요)와 `spec/keep-restore-backup-20261126` 브랜치(`src/update.rs` 를 고쳐요)가 NeXT 에 착지한 tip 에서 새 브랜치로 시작해요. 프로젝트 설정 SPEC 의 묶음은 이미 착지했어요(`2e97d012e5`).
4. 테스트 먼저(RED): `tests/integration_guard_repoint.rs` 의 AC-1·2·3·7 과 `tests/update_git_hooks_cli.rs` 의 AC-4·5 를 추가하고, 변경 전 트리에서 실패하는 것을 읽어요(`moai tdd red`, 슬롯 `--max 2`). 위험 경로(AC-2)는 임시 HOME·임시 저장소·실제 바이너리를 쓰는 Rust 시험이에요. 셸 스크립트는 쓰지 않아요.
5. `src/card/guard/repoint.rs`: `stale(project)`(기록은 `read_record`, 프로필은 `landed::profile_branch`+`branch_ref`, 주 워크트리만), `preview`, `apply`, `revert`, `stale_notice(cwd)`, `update_guard(cwd)`(= `can_act` 이면서 낡지 않음), `plan_update_hooks`, `apply_update_hooks`. 적용 순서는 이래요: 모든 검증을 먼저 하고, 잠금 `moai/integration-branch.lock` 을 `create_new` 로 잡고, 기록을 다시 읽어 비교하고, 영수증을 `create_new` 로 쓰고, 잠금을 기록 위로 rename 해요. rename 이 실패하면 영수증과 잠금을 지워요.
6. 영수증은 `<공통 git 폴더>/moai/record-repoint/<stamp>.json`(stamp 는 `YYYYMMDDTHHMMSSZ`)이고 되돌리기는 옆에 `<stamp>.revert.json` 을 써요. 모두 추가만 해요. 필드는 `schema`, `kind`, `stamp`, `applied_at`, `old_record`, `new_record`, `old_bytes_sha256`, `new_bytes_sha256`, `profile_branch`, `plan_digest`, `permission_reference`, `hook_present` 이고, 되돌리기 영수증은 `stamp`, `reverted_at`, `permission_reference`, `restored_record` 예요. 훅 설치기의 retire 폴더에는 두지 않아요(그 폴더에 모르는 파일이 있으면 `git-hooks-revert` 가 거절해요, `src/git_hooks/mod.rs:650-654`). digest 는 기록 줄과 바이트, 목표 ref, 공통 git 폴더를 묶어요. 목표 브랜치의 끝(tip)은 착지마다 움직여서 묶지 않아요.
7. `src/ops/card.rs`: `Action::GuardRepoint { apply, plan_digest, revert, permission_reference, project }` 와 match 한 팔(명령 이름은 `guard-repoint` 예요. `card guard` 는 모르는 상태를 "허용"으로 답해서 `repoint` 를 받으면 아무 일도 안 해요, `src/ops/card.rs:2298-2303`). `--apply` 와 `--revert` 는 함께 못 써요. 도움말 첫 줄은 "Writes the git folder only with `--apply` or `--revert`." 로 시작해요. `src/card/drift.rs` 의 `reference_ok` 를 `pub(super)` 로 열어 permission reference 검사를 같은 규칙으로 써요.
8. `src/update.rs` 세 곳: `:2076` 은 `repoint::plan_update_hooks(home, &cwd)`, `:2618` 와 `:2764` 는 `repoint::apply_update_hooks(&cwd, token, &utc_now())`, 그리고 `:2071` 뒤에 `notices.extend(repoint::stale_notice(&cwd))`. 안내는 모든 digest 밖이고, 계획과 두 적용이 같은 값을 받아야 토큰이 같아요(`src/git_hooks/mod.rs:310-311`).
9. 문서 세 개(`card.md` 표에 `card guard-repoint` 줄, `git_hooks_plan.md`, `git_hooks_apply.md`). `tests/integration_guard_install.rs:660` 의 `ac7_` 가 읽는 문자열(`is not guarded`, `written`, `present`, `not_written`, `moai/integration-branch`)은 지켜요.
10. 영수증 전 준비: `cargo fmt --all`, `src/ops` rustfmt, 대상 clippy, 새 시험 파일과 새 인자 때문에 `state_folder_literal_scan`, `repo_scans_table`, `source_manifest_facts`, `next_flow_skill_scan`, `test_area_layout`, `ac_command_lint`, `--bin moai-cli` 를 nextest 로 돌려 확인해요. 가드 뮤턴트는 새 가드 시험마다 한 줄씩 Result 에 적어요(편집 한 번, 슬롯 안 nextest, 평범한 `git restore` 한 줄): `guard-mutant: ac2_ | 잠금 아래 기록 비교 건너뜀 | exit 101`, `guard-mutant: ac3_ | 되돌리기 비교 건너뜀 | exit 101`, `guard-mutant: ac1_ | 링크 워크트리에서 digest 를 답함 | exit 101`, `guard-mutant: ac4_ | update 호출 한 곳에서 가드 값 대신 거짓을 넘김 | exit 101`. 그다음 `moai trailer-check`, `moai check affected --card` 한 번, READY 노트.

## Risks

- 다시 맞춘 뒤에는 옛 브랜치가 더는 지켜지지 않아요(기록이 한 ref 만 담아요). 미리보기가 `unguarded_after` 로 미리 말하고, 두 참조 기록은 이 SPEC 밖이에요(D-6).
- 다시 맞춘 뒤 `card land`·`finish`·`submit`·`drift`·`cleanup`·`remote_copy` 는 모두 새 기록을 따라가요(`landed::integration_branch`). 옛 브랜치에서 착지를 시작한 진행 중인 카드가 있으면 어긋나요. 그래서 `moai/land-intent` 가 있으면 거절하고, 원장은 읽지 않아요(git 폴더만 만져요). 착지 없이 옛 브랜치 위에서 작업 중인 카드는 감지하지 못해요(미확인). 이 빈틈은 `docs/later.md` 에 한 줄로 남겨요.
- 어긋나 있는 동안 Bash 경계는 프로필 이름을, git 훅은 기록 이름을 지켜요. 다시 맞추면 두 층이 같은 브랜치를 가리켜요.
- 기록이 낡은 채로 `moai update` 가 가드 훅을 깔면 사용자가 방금 놓으려는 `main` 을 잠가요. 그래서 낡은 동안은 훅 제안을 보류하고 먼저 다시 맞추라고 안내해요(D-2). `git-hooks-plan` 과 `git-hooks-apply` 는 바꾸지 않아요. 전역으로 `can_act` 를 낡음 인식으로 바꾸면 승인된 REQ-2 의 테스트 `ac2_the_record_is_written_once_and_never_rewritten` 이 깨져요.
- 기록 파일이 파일이라 `git update-ref` 의 비교 후 교체가 없어요. 그래서 이 SPEC 이 잠금 파일과 적용 직전 재비교로 직접 만들어요. 프로세스가 죽어 잠금이 남으면 다음 적용이 거절돼요(수동으로 잠금을 지우는 안내가 필요해요, 거절 문장에 잠금 경로를 적어요).
- `git-hooks-revert` 는 자신이 설치하지 않은 관리 훅도 바이트가 같으면 지워요(`src/git_hooks/mod.rs:698-712`). 이 SPEC 의 되돌리기는 그 동작에 기대지 않고, 영수증과 기록 바이트만 비교해요.
- 사용자만 낼 수 있는 permission reference 를 CLI 가 레인의 호출과 구분하지 못해요. 그래서 레인은 `--apply` 를 직접 하지 않고 `LANE-BLOCKED kind=decision` 으로 멈춰요(D-4).
- 이 SPEC 의 코드는 위 두 브랜치가 먼저 착지해야 시작할 수 있어요. 각각 `src/card/guard.rs:234-247` 과 `src/update.rs` 를 고쳐요. 이 SPEC 은 `guard.rs` 에 한 줄만 더해서 병합이 쉬워요.
- mo.ai.kr 의 실제 상태는 카드 본문의 글뿐이에요(미확인). 시험은 임시 HOME·임시 저장소에서만 재현해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 낡은 기록을 맞추는 일은 새 행동 `moai card guard-repoint`(미리보기 → `--apply`+digest+permission reference → `--revert`)로 해요. 설치기, `seed_record`, `card land` 는 계속 기존 기록을 다시 쓰지 않아요. 선택지 b: 맞추는 일을 `--confirm-git-hooks` 토큰에 합쳐요(승인된 REQ-2 를 바꾸는 개정이라 사용자만 고를 수 있어요). 기본값: a(별도 행동). | 승인된 `SPEC-20261125-integration-guard-install` REQ-2 와 `SPEC-20261121-ref-guard` D-11("기록을 바꾸는 일은 나중의 별도 행동")이 자리를 남겨 뒀어요. 기존 시험 네 개가 그대로 통과해서 시험을 약하게 바꾸지 않아요. | 새 행동과 그 시험, 문서 줄을 지우는 판. 기록은 영수증의 되돌리기로 돌려요. |
| D-2 | `moai update` 는 기록이 낡은 동안 가드 훅 제안을 보류하고 안내를 보여 줘요. 선택지 b: 경고와 함께 제안해요. c: `git-hooks-plan` 에서도 보류해요(승인된 시험 `ac2_` 가 깨져요). 기본값: a. | 낡은 기록에 훅을 깔면 사용자가 방금 놓으려는 브랜치를 잠가요. 전역 변경은 승인된 REQ-2 의 시험을 깨요. | update 의 한 호출을 가드 값으로 되돌리는 판. |
| D-3 | 프로필은 주 워크트리에서만 읽고, 링크 워크트리는 `linked_worktree` 로 거절해요. 목표는 로컬 `refs/heads/<이름>` 이 실제로 있을 때만이에요. | `SPEC-20261121-ref-guard` D-11 "a lane's checked-out profile never writes it". 없는 브랜치를 기록하면 `card land` 말고는 그 브랜치를 만들 수 없어요(`read_record` 는 존재를 보지 않아요). | 검사 두 줄을 지우는 판. |
| D-4 | 누가 맞출 수 있나: `--permission-reference`(사용자의 승인을 가리키는 값)가 필수이고 `--by` 는 없어요. 레인은 `--apply` 를 직접 하지 않고 `LANE-BLOCKED kind=decision` 을 써요. 선택지 b: `--by leader`(drift 처럼). c: 대화형만. 기본값: a. | 저장소 보호 설정을 바꾸는 일은 `guarded` 예요. CLI 는 레인과 사용자를 구분하지 못해서 규칙으로 막아요. | 인자 검사를 바꾸는 판. |
| D-5 | 진행 중인 카드: `moai/land-intent` 파일이 있으면 거절해요. 원장은 읽지 않아요. 선택지 b: 원장에서 옛 브랜치 위의 열린 착지를 찾아 거절해요(원장 읽기가 생기고 git 폴더 전용이 아니게 돼요). 기본값: a. | 착지 중 기록이 바뀌면 `card land` 의 비교가 어긋나요. 원장 읽기는 범위를 넓혀요. | 파일 검사 한 줄을 지우는 판. |
| D-6 | 다시 맞춘 뒤 옛 브랜치는 지켜지지 않고, 미리보기가 `unguarded_after` 로 말해요. 선택지 b: 두 참조 기록(훅 텍스트가 바뀌는 더 큰 tier 2 변경). 기본값: a. | 기록은 한 ref 를 담고, 훅은 그 한 줄만 읽어요. 훅 바이트를 바꾸지 않는 것이 ref-guard Non-goal 이에요. | 두 참조 기록은 별도 SPEC. |
| D-7 | 영수증은 `<공통 git 폴더>/moai/record-repoint/<stamp>.json`, 되돌리기는 `<stamp>.revert.json`, 모두 추가만 해요. 훅 설치기의 retire 폴더는 쓰지 않아요. 같은 초에 두 번 적용하면 `create_new` 로 거절돼요. | `git-hooks-revert` 가 retire 폴더의 모르는 파일을 거절해요(`src/git_hooks/mod.rs:650-654`). | 폴더 이름을 바꾸는 판. |
| D-8 | 세션 시작 줄이 낡은 기록을 알리는 일은 이 SPEC 에 넣지 않아요(후속). 선택지 b: `src/hook_session_start.rs` 를 범위에 넣어요. 기본값: a. | 그 파일은 다른 카드의 영역이고, 알림은 `moai update` 와 `guard-repoint` 미리보기가 맡아요. | 후속 카드. |
| D-9 | `moai update` 의 훅 적용도 `git-hooks-apply` 처럼 비어 있는 기록을 씨앗으로 심어요. 미리보기에는 그 문장이 없고, `git_hooks_apply.md` 에 적고 답의 `guard.record` 로 보여요. 선택지 b: 미리보기에 한 줄을 더해요. 기본값: a. | 같은 `apply_guarded` 와 `seed_record` 를 써서 동작이 같아요. | 문서 문장을 바꾸는 판. |
| D-10 | 카드 보낸 쪽의 "옛 경로"는 기록의 낡은 브랜치 값을 뜻해요. 설치 경로는 저장되지 않아요. | 리더가 확인했어요(메시지, 같은 날). 코드 읽기에서도 설치 경로를 저장하는 곳이 없었어요. | 해당 없음. |

## Result

구현 레인(moai-adk-agent-uma)이 2026-11-26T23:52Z(date -u)에 채웠어요. macOS(개발 Mac), 빌드 폴더는 이 워크트리의 `target/`, 무거운 명령은 모두 `moai heavy-slot-acquire --name moai-adk-agent-uma --max 2 -- env -u CLAUDE_CODE_SESSION_ID …` 슬롯 안에서 돌렸어요. 기준 NeXT `0d999775156900bca36e7b6163409fc0973d7b0a`(`src/update.rs` 를 고친 다른 카드가 랜딩된 뒤) 위로 옮겼어요. 승인은 레저 노트 10485(사용자가 직접 고름)이고, 이 파일 머리글은 승인 기록 없이 draft 로 남아 있어요.

변경: `src/card/guard/repoint.rs`(새, 미리보기·`--apply`·`--revert`·`stale()`), `src/card/drift.rs`(`reference_ok` 를 `pub(crate)` 로), `src/card/guard.rs`(`pub mod repoint;` 한 줄), `src/ops/card.rs`(`Action::GuardRepoint`, 라인 리스), `src/update.rs`(`guard_answer`·`apply_git_hooks`·`repoint_notice`, 계획과 두 적용 지점), 문서 3곳(`card.md` 라인 리스), `tools/repo_scans.json`(라인 리스, `moai check fix` 가 PASS), 테스트 두 파일.

테스트 (모두 이 레인이 슬롯에서 직접 돌린 결과예요):

| 명령 | 결과 |
|---|---|
| `cargo test --locked --offline --test integration_guard_repoint` | 26 통과, 0 실패 (ac1_ 7, ac2_ 12, ac3_ 6, ac7_ 1) |
| `cargo test --locked --offline --test update_git_hooks_cli` | 16 통과, 0 실패 (새 7개 포함, 기존 9개) |
| `cargo test --locked --offline --test integration_guard_install ac2_` | 3 통과 (`ac2_the_record_is_written_once_and_never_rewritten` 포함, 수정 없음) |
| `cargo test --locked --offline --test card -- card_guard_cli card_land_cli` | 88 통과 |

RED 는 두 번 기록했어요: 첫 기록(`4decf2c463`)은 여섯 체크가 모두 단언 실패, 두 번째(보안 리뷰 뒤 `revert` 게이트 테스트 둘)는 두 체크가 모두 단언 실패였고 `moai tdd green` 이 둘 다 통과로 재생했어요. 첫 `tdd red` 시도 하나는 여섯째 체크가 환경 실패(종료 코드 없음, 슬롯 대기 추정)로 거절됐고, 같은 체크를 직접 돌려 단언 실패를 확인한 뒤 다시 기록했어요.

- AC-1: met
- AC-2: met
- AC-3: met
- AC-4: met
- AC-5: met
- AC-6: met
- AC-7: met
- AC-8: met

가드 뮤턴트 (편집 한 번, 슬롯 안 `cargo test`, 평범한 `git restore` 한 줄, 매번 트리가 깨끗한 것까지 읽었어요):

guard-mutant: ac2_a_wrong_or_absent_digest_writes_nothing | apply 의 digest 비교를 `if false &&` 로 끔 | exit 101
guard-mutant: ac3_a_changed_record_refuses_the_revert | 잠금 아래 기록 비교가 아무 기록이나 받음 | exit 101
guard-mutant: ac1_a_linked_worktree_answers_linked_worktree | probe 의 주 워크트리 확인을 끔 | exit 101
guard-mutant: ac3_a_linked_worktree_cannot_revert | revert 의 주 워크트리 확인을 끔 | exit 101
guard-mutant: ac3_a_land_intent_file_blocks_the_revert | revert 의 `land-intent` 확인을 끔 | exit 101
guard-mutant: update_offers_the_guard_hook_not_for_a_stale_record | `guard_answer` 가 낡은 기록을 무시함 | exit 101
guard-mutant: none | 나머지 가드 테스트(`ac2_a_land_intent_file_blocks_the_apply`, `ac2_a_leftover_lock_…`, `ac2_a_failing_receipt_write_…`, 참조 검사 셋)는 뮤턴트를 돌리지 않았어요(not_run) | 해당 없음

SPEC 이탈과 추가 (모두 더 보수적인 쪽이에요):

- 영수증에 `old_bytes_hex` 필드를 더했어요. `ac3_revert_restores_the_old_bytes…` 가 CRLF 와 둘째 줄까지 정확히 되돌리기를 요구해서, 해시만으로는 되돌릴 수 없어서예요. 되돌릴 때 이 값을 두 해시와 맞춰 보고 다르면 거절해요.
- 보안 리뷰(자동) 지적으로 `revert` 에도 주 워크트리 확인과 `moai/land-intent` 확인을 더했고, 영수증의 옛 기록이 브랜치 ref 가 아니면 거절해요. REQ-6 의 거절 조건 목록에는 없는 추가예요. 테스트 둘(`ac3_a_linked_worktree_cannot_revert`, `ac3_a_land_intent_file_blocks_the_revert`)이 고정해요.
- `moai update` 의 적용은 `apply_guarded` 뒤에 `seed_record` 를 불러 답의 `guard.record` 를 채워요(D-9, 테스트가 요구해요). 기존 기록은 다시 쓰지 않아요.
- 미리보기는 낯선 `--plan-digest`·`--permission-reference` 를 거절하고, `--revert` 는 `--plan-digest` 를 거절해요. 스탬프는 `utc_stamp` 로 검사해서 `../x` 가 영수증 폴더를 벗어나지 못해요.

알려진 빈틈:

- 이 카드는 리더의 노트 10510 의 두 요청을 하지 않았어요. (1) 프로필을 바꾸는 첫 랜딩의 안내와 순서(`guard-repoint` 미리보기 → 확인 → `--apply` → `card drift` → 랜딩)를 next-flow 와 거절 문구에 적기, (2) 기록이 프로필과 다를 때 `card land` 가 거절하기. 둘 다 승인된 SPEC 개정 1 의 범위 밖이에요(REQ-11 은 `card land` 가 기록을 다시 쓰지 않는다고만 해요). SPEC 개정과 사용자 승인이 필요해요.
- 잠금 아래 비교는 apply 경로에서 digest 검사가 먼저 걸려서 결정적 테스트로 따로 잡히지 않아요(깊은 방어). revert 경로에서는 잡혀요.
- 옛 브랜치 위에서 랜딩 없이 일하는 카드는 감지하지 못해요(Risks).
- 서브 에이전트 기록을 읽어 본 결과: 문서 담당의 마지막 `git status` 한 줄이 훅에 거절됐어요(다시 시도하지 않았고 제가 직접 확인했어요). RED 담당이 쓰기 전에 `moai card check` 를 하지 않았고(뒤에 확인하니 `allowed: true`), RED 담당과 구현 담당이 자기 워크트리의 `target/debug` 를 `rm -rf` 로 지웠어요(상대 경로, 규칙 안, 빌드가 느려졌어요).
