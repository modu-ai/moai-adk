---
id: SPEC-20261125-land-sync-merge
title: "a leader-only command merges another branch into the integration branch with one guarded move and a wide check"
status: draft
revision: 1
created: 2026-11-25
branch: spec/land-sync-merge-20261125
---

# a leader-only command merges another branch into the integration branch with one guarded move and a wide check

## Intent

통합 브랜치(`NeXT`)에 다른 브랜치(예: 릴리스·hotfix 뒤의 `main`)를 되머지하는 길이 리더에게 없어요. `card land` 는 카드의 감사된 head 만 받고, `card land-batch` 도 카드 head 의 join 만 만들어요. 손 머지(`git merge --no-ff main`)는 `reference-transaction` 보호 훅이 설치되면 막혀요(`.moai/specs/SPEC-20261125-integration-guard-install`). 그러면 "main 에만 있는 커밋을 통합 브랜치로 되돌리는 길"이 흐름 문서와 도구 양쪽에서 불가능해져요. 출처는 mo.ai.kr 리더의 개선 요청 덧붙임 6 항목 E(`~/.moai/reports/mo-ai-kr/leader-20261125/moai-adk-improvement-request-addendum-6.md`)이고, 사용자가 2026-11-25 리더 대화에서 "받아라"로 받기로 했어요(이 일을 맡긴 카드의 본문).

이 SPEC 은 리더 등록 세션만 부를 수 있는 명령 `moai card land-sync <FROM> --spec <SPEC-ID> --by leader` 하나를 더해요. 명령은 통합 팁 위에 `<FROM>` 의 팁을 머지 커밋으로 만들고(첫 부모 = 통합 팁), 넓은 검사를 스스로 돌려 통과해야만, `card land` 와 같은 의도 기록(intent record)을 써서 통합 브랜치를 compare-and-swap 한 번으로 옮겨요. 보호 훅과 `card guard` 는 한 바이트도 바꾸지 않아요.

읽은 사실(NeXT 팁 `59ac3344cd`, 이 트리; 조사는 `moai-research` 두 개, 읽은 시각은 2026-11-25 13:38Z 무렵):

1. 이동의 길은 `landed::move_branch`(`git update-ref -m <msg> <ref> <new> <old>`)이고, `land.rs` 가 `record_land_started` → `guard::write_intent` → `move_branch` → `guard::remove_intent` 순서로 불러요. `card guard`(`guard.rs` `judge_move`)는 의도 기록이 `{ref, old, new}` 와 정확히 맞고 옛 값이 새 값의 조상일 때만 통과시켜요. 새 커밋의 부모, 카드, 판정, `by` 는 읽지 않아요. 그래서 첫 부모가 팁인 머지 커밋은 같은 의도 기록으로 통과해요.
2. `card land-batch` 만 리더 세션을 확인해요(`batch.rs` `confirm_leader`: 레지스트리 이름 `leader`, `mode: leader`, 살아 있는 임대, 같은 세션 id). `card land` 는 `--by leader` 낱말만 봐요.
3. 넓은 검사는 `batch.rs` `run_check` 가 `<프로젝트>/target/land-batch/<12자리>-<n>` 의 떼어낸(detached) 워크트리에서 `check_affected::run`(전체 선택)으로 돌리고, `summary.txt` 한 줄이 `receipt_covers` 를 통과할 때만 스캔을 덮은 것으로 읽어요.
4. `drift::refusal` 은 `card land` 마다 불려요. 통합 팁이 장부의 마지막 `card.land_started`/`card.landed` 짝과 다르면 `drift` 로 거절해요. 그래서 되머지가 장부에 행을 남기지 않으면 다음 `card land` 가 거절돼요. `drift::rows` 는 `card_id` 를 `i64` 로 읽어서, `card.land_started` 에 NULL 카드를 넣으면 이후 모든 착지가 깨져요(읽어서 추론, 실행하지 않음). 카드 없는 이벤트는 이미 있어요(`claude.run`, `headless_run.rs`).
5. 이 SPEC 이 새로 만드는 것은 명령 하나, `card.sync_started` 와 `card.synced` 두 이벤트 종류, `sync_sources` 프로필 키 하나예요.

## Interview

- Q1: `<FROM>` 으로 어떤 브랜치를 받을까요? 이 명령은 감사 판정 없이 통합 브랜치에 커밋을 넣는 길이라서, 받는 브랜치를 막지 않으면 감사 관문의 우회로가 돼요. → A: (대기) 선택지 a(권장): `.moai/project.yaml` 의 `work_flow.sync_sources` 목록에 이름이 정확히 있는 로컬 브랜치만 받아요. 목록이 비면 키 이름을 대고 거절해요. 선택지 b: 아무 로컬 브랜치나 받아요(덧붙임 6 의 글자 그대로지만 관문이 약해져요).
- Q2: 머지 커밋의 `Moai-Spec` 트레일러에 쓸 SPEC 을 `--spec` 으로 꼭 받을까요? → A: (대기) 선택지 a(권장): 꼭 받고, 팁에 있는 승인된 SPEC 이어야 해요(왜 되머지하는지 SPEC 이 남아요). 선택지 b: 트레일러에 이름만 적고 승인 여부는 보지 않아요.
- Q3: 이 SPEC 의 개정 1 을 승인할까요? → A: (대기)

## Non-goals

- `card guard`, 보호 훅의 글, `card land`, `card land-batch` 의 동작을 바꾸는 일. 이 SPEC 은 그들을 부르거나 같은 부분을 베낄 뿐이에요.
- 통합 브랜치가 아닌 브랜치로 되머지하는 일, 여러 브랜치를 한 번에 받는 일, 병합 충돌을 풀어 주는 일. 충돌은 거절해요.
- 원격 태그, 릴리스, `main` 에 쓰는 일. 명령은 로컬 통합 브랜치만 옮기고 `remote_copy` 가 켜져 있으면 `card land` 와 같이 통합 브랜치만 fast-forward 로 본사해요.
- `.moai/project.yaml` 에 `sync_sources` 값을 넣는 일. 키의 읽기만 코드가 해요. 값을 정하는 건 사용자 몫이에요.
- `CLAUDE.md` 의 "`card land` 로만 옮긴다" 문장을 고치는 일. `CLAUDE.md` 는 쓰기 범위에 그 파일을 적은 카드로만 바뀌어요(후속 한 줄).
- 덧붙임 6 의 항목 5(`lander_review` 와 AUDIT 문서가 서로 다른 절차를 가리킨다는 지적). 이 카드의 일이 아니에요.

## Requirements

- REQ-1: When `card land-sync` is called, the card command shall run only for `--by leader` from the session of the live registered leader entry · failure: none — any other caller (no session id, a lane session, no leader entry, an entry without `--mode leader`, an ended lease, another `--by` word) is refused before anything is read from git or written, the answer ends with `nothing was changed`, and the refusal text names the land-sync command, not `land-batch`.
- REQ-2: When `card land-sync` is called with a source branch, the card command shall accept only a local `refs/heads/` branch whose name the profile list `work_flow.sync_sources` holds exactly and that is not the integration branch, together with an approved SPEC given by `--spec` · failure: none — an unlisted source, an empty list, a tag, a remote-tracking name, the integration branch itself, a missing or malformed branch and an unapproved or missing SPEC are each refused with a reason that names the key or the SPEC, with no ledger row, no reference change and no intent file written.
- REQ-3: When every gate of REQ-1 and REQ-2 passes and the tip of the source branch is not inside the integration tip, the card command shall make one merge commit whose first parent is the integration tip and whose second parent is the source tip, with git plumbing and no checkout · failure: none — a conflicting or unrelated history is refused as `conflicts` or `join_failed` with nothing written, a source already inside the tip answers `moved: false` and writes nothing, and a drifted tip is refused as `drift` before the merge is made.
- REQ-4: When the merge commit exists, the card command shall run the wide `check affected` of that commit itself in a private detached worktree and move the branch only after a full-selection PASS · failure: none — a check that cannot start gives `check_not_started`, a red check gives `check_failed`, and a result that is not a full-selection PASS (a replaced check program, a `selection=` word, no `result=PASS`) gives `not_full_pass`, each with the branch, the ledger and the intent file unchanged and the private worktree removed.
- REQ-5: When the wide check passed, the card command shall move the integration branch to the merge commit by one compare-and-swap through the intent record `card land` writes and the protection hook accepts · failure: none — a tip that is not the one the merge was built on, found before the move or lost at the swap, is refused as `stale_base` with nothing moved and the intent file removed, and the guard, the hook text and every hand move of the branch stay refused as before.
- REQ-6: When the card command moves the branch, the card command shall record the card-less ledger rows `card.sync_started` before the move and `card.synced` after it, each naming the merge commit, the source tip and the old tip · failure: none — a sync that stopped after the move is closed by running the same call again, and rows with no card never break `drift`, `card export` or the landing's own reads.
- REQ-7: When the branch moved, the card command shall answer the count `git rev-list --count <branch>..<source>`, the worktree follow and the remote-copy state, and run the remote copy that `card land` runs · failure: none — a checkout that cannot follow (a dirty or busy primary checkout) is refused before the move, and a remote copy that cannot finish is reported as it is without undoing the sync.
- REQ-8: When `card land-sync` changes what a leader types or reads, the leader workflow, the leader rules and the card operations document shall say so · failure: none — the three documents name the command, its gates and its refusal codes, and the existing sentences their tests pin stay as they are.
- REQ-9: When `card land` checks the integration tip after a sync, the drift check shall expect the merge commit of the newest `card.synced` or `card.sync_started` row as the tip · failure: none — the next landing is not refused as `drift`, the refusal text and digest of a landing that follows a landing stay byte-identical, and the newer of a landing and a sync decides.

## Acceptance

- AC-1 (REQ-3, REQ-5, REQ-7): a registered leader merges a listed source into a temporary integration branch, the branch moves once with the old tip as first parent and the source tip as second parent, one reflog entry is added, the answer carries the `rev-list` count `0`, a source already inside the tip answers `moved: false`, and a tip that moved after the merge was built is refused as `stale_base` · verify: `cargo test --locked --offline --test card -- card_land_sync_cli::ac1_`
- AC-2 (REQ-1): a lane session, a call with no session id, a missing or wrong `--by`, no leader entry, an entry without `--mode leader` and an ended lease are each refused with the branch, the references and the ledger unchanged · verify: `cargo test --locked --offline --test card -- card_land_sync_cli::ac2_`
- AC-3 (REQ-2, REQ-3): an unlisted source, an empty `sync_sources`, a tag, the integration branch, a missing branch, a malformed name, an unapproved SPEC, a conflict and a drifted tip are each refused, and a refusal writes no ledger row, no reference and no intent file · verify: `cargo test --locked --offline --test card -- card_land_sync_cli::ac3_`
- AC-4 (REQ-4): a check that cannot start gives `check_not_started`, a red one `check_failed` and a replaced check program `not_full_pass`, the branch is unchanged each time, and a full-selection PASS of the command's own in-process run lets the move happen · verify: `cargo test --locked --offline --test card -- card_land_sync_cli::ac4_`
- AC-5 (REQ-5): in a temporary repository whose hook was installed only through `git-hooks-plan` and `git-hooks-apply` (the dangerous path itself), `git update-ref`, `git branch -f`, `git branch -D`, `git merge --no-ff` in the worktree that has the branch checked out and `git branch -M` leave the branch where it was before and after a sync, and the sync moves it and leaves no intent file · verify: `cargo test --locked --offline --test card -- card_land_sync_cli::ac5_`
- AC-6 (REQ-6, REQ-9): the rows `card.sync_started` and `card.synced` have no card and name the merge commit, the source tip, the old tip and the SPEC, the next `card land` of an audited head is not refused as `drift`, a sync stopped after the move is closed by the same call, and `card export` reads a ledger that holds such rows · verify: `cargo test --locked --offline --test card -- card_land_sync_cli::ac6_` and `cargo test --locked --offline --lib -- card::drift`
- AC-7 (REQ-2): the profile key `work_flow.sync_sources` is a list of at most 8 valid branch names, with no duplicate and not the integration branch, and a missing key is an empty list · verify: `cargo test --locked --offline --lib -- profile::`
- AC-8 (REQ-8): the leader workflow, the leader rules and the card operations document name the command, its gates and its refusal codes, with the pinned sentences and the byte caps intact · verify: `cargo test --locked --offline --test leader_route --test lane_team_names --test next_flow_skill_scan --test repo_claude_md --test leader_reclaim_guidance --test source_manifest_facts`
- AC-9 (REQ-1, REQ-5): the existing landing, batch, guard and drift tests still pass with no change of their expected texts · verify: `cargo test --locked --offline --test card -- card_land_cli card_land_batch_cli card_land_checked_out_cli card_land_audited_merge_cli card_land_scans_cli card_guard_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-land-sync-merge/spec.md`
- write: `src/card/land/sync.rs`, `src/card/land.rs`, `src/card/land/batch.rs`, `src/card/follow.rs`, `src/card/drift.rs`
- write: `tests/card/card_land_sync_cli.rs`, `tests/card/main.rs`, `tests/leader_route.rs`
- write: `skills/moai/workflows/leader.md`, `docs/leader-rules.md`
- write by line lease (shared files): `src/ops/card.rs`, `src/profile.rs`, `docs/design-review/appendix/operations/card.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/card/guard.rs`, `src/card/landed.rs`, `src/card/remote_copy.rs`, `src/card/spec_link.rs`, `install/git-hooks/reference-transaction`, `tests/integration_guard_install.rs`
- forbidden: `src/card/guard.rs` and the hook file (not even a byte), `.moai/project.yaml`, `CLAUDE.md`, the `.git/hooks` and `.git/moai` folders of this repository, `skills/moai/references/flows/next-flow.md` (73 bytes free under its cap; the sentences go to `leader.md` and `leader-rules.md`)

## Plan

1. 테스트 먼저(RED): `src/profile.rs` 의 `sync_sources` 단위 시험, `src/card/drift.rs` 의 단위 시험(완전한 sync 는 기대 팁, 중단된 sync 는 출처 팁과 머지를 받음, 착지와 sync 중 새것이 이김, 착지 거절 문장과 digest 는 그대로), `tests/card/card_land_sync_cli.rs` 의 ac1~ac6(고정물은 `card_land_batch_cli.rs` 의 World·`register_leader`·`script`·`fake_cargo` 를 줄여 베끼고 ac5 는 `integration_guard_install.rs` 의 설치기 고정물을 써요), `tests/card/main.rs` 에 `mod card_land_sync_cli;` 한 줄. 변경 전 트리에서 실패하는 것을 먼저 읽어요.
2. `src/profile.rs`(줄 임대): `work_flow.sync_sources` 키와 검증, `profile.rs:1580` 의 리터럴.
3. `src/card/drift.rs`: `card.sync_started`·`card.synced` 두 종류, `rows()` 에 두 짝의 합집합, `Row.card` 와 `Expected.card` 를 `Option<i64>` 로, 출처 값 `sync`·`interrupted_sync`, 카드 없는 이벤트를 쓰는 작은 함수(`WriteLedger::with_write_tx`, `new_event_key(.., None)`). `ledger/workflow.rs` 는 건드리지 않아요.
4. `src/card/land/batch.rs`: `confirm_leader` 를 `leader_reason(project, session, now, what)` 으로 나누고 `confirm_leader` 는 같은 글을 내는 감싸개로 두며, `run_check` 등을 `pub(super)` 로만 열어요. `run_check` 본문은 바꾸지 않아요(`record_summary` 의 자리를 고정하는 단위 시험이 있어요).
5. `src/card/follow.rs`: `rerun` 에서 `follow::rerun_from(project, old, full_ref, branch, head)` 를 뽑아 중단된 sync 의 워크트리 따라가기가 카드 없이 돌게 해요.
6. `src/card/land/sync.rs`(+ `land.rs` 에 `pub mod sync;` 한 줄): (a) `--by leader`, (b) `leader_reason`, (c) 출처 관문(로컬 브랜치, 목록, 통합 브랜치 아님), (d) `--spec` 승인 확인, (e) 팁 읽기와 `drift::refusal`, `guard::protected`, (f) 출처가 이미 팁 안이면 `moved: false`, (g) `merge-tree --write-tree` 와 `commit-tree -p <팁> -p <출처>`, (h) `follow::plan`, (i) 넓은 검사와 그 뒤 팁 다시 읽기(다르면 `stale_base`), (j) `card.sync_started` 행, `write_intent`, `move_branch`, `remove_intent`, `follow::apply`, `card.synced` 행, (k) 답. 디버그 빌드 전용 시험 이음매 `MOAI_TEST_LAND_SYNC_CHECK` 와 그 고정 시험.
7. `src/ops/card.rs`(줄 임대): `Action::LandSync` 와 match 한 갈래, 이동했을 때만 `remote_copy::copy(&project, "land")`.
8. 문서: `leader.md`(약 1100바이트 여유 안에서 한 문장과 실패 표 한 줄), `leader-rules.md` 한 줄, `operations/card.md` 한 줄(줄 임대), `tests/leader_route.rs` 에 문장 고정, 고친 스킬 파일의 sha256 줄을 `source-manifest.json` 에 `shasum -a 256` 으로(줄 임대).
9. 검증: `cargo fmt --all`, AC 명령, `moai repo-scans --refs <tip> HEAD` 가 고른 것만, 가드 뮤턴트(셸로, 한 슬롯), 위험 경로 직접 실행(`target/` 아래 일회용 저장소에 진짜 훅을 `git-hooks-apply` 로 설치하고 리더로 sync, 손 이동과 비리더 호출), 그다음 `check affected --card` 한 번.

## Risks

- 이 명령은 어떤 판정도 덮지 않은 내용을 통합 브랜치에 넣어요. 그래서 관문이 넷이에요: 등록된 리더 세션, 프로필의 출처 목록, 승인된 SPEC, 전체 선택의 넓은 검사 영수증. 각각이 거절로 시험돼요. 목록이 비면 아무것도 받지 않아요.
- `WorkFlowSettings` 가 `deny_unknown_fields` 라서, 새 키가 든 프로필을 옛 설치본이 거절해요. 리더는 설치본을 새 빌드로 올린 뒤에 값을 넣어요.
- 거절된 호출도 `merge-tree`·`commit-tree` 가 객체 저장소에 닿지 않는 객체를 남겨요(`land-batch` 와 같아요). 시험은 참조와 장부만 봐요.
- 넓은 검사가 느리고, 그동안 팁이 움직이면 `stale_base` 로 거절해요. `land-batch` 와 달리 다시 만들지 않아요; 리더가 다시 불러요.
- 실제 쓰임은 `NeXT` 가 주 체크아웃에 체크아웃된 상태라서 워크트리 따라가기가 보통 길이에요. 더럽거나 바쁜 체크아웃은 이동 전에 거절해요(되돌릴 변경을 지우지 않아요).
- 카드 없는 이벤트가 `recover` 의 이벤트 복사 길에서 어떻게 읽히는지는 읽지 않았어요(미확인). 구현 중 AC-6 의 `card export` 시험과 함께 읽어요.
- 다른 카드가 레지스트리 키(`lane_registry.rs`, `session.rs`)를 바꾸고 있어요. `leader_reason` 한 곳을 쓰면 구현이 하나지만, 그 카드가 착지한 뒤 리더 시험을 다시 돌려요.
- `CLAUDE.md` 는 여전히 `card land` 로만 옮긴다고 적어요. 후속 카드 한 줄이에요(Non-goals).
- tier 2(관문과 통합 브랜치 이동)라서 전체 감사를 저자와 다른 세션이 해요. 스킬 복사본은 착지 뒤 리더가 `moai update` 로 맞추고 `diff -rq` 로 읽어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 명령 이름과 모양은 `moai card land-sync <FROM> --spec <SPEC-ID> --by leader`. | 덧붙임 6 이 제안한 이름 `land-sync` 와 `card land-batch` 의 모양을 따라요. | 이름만 바꾸는 판. |
| D-2 | `<FROM>` 은 프로필 `work_flow.sync_sources` 에 있는 로컬 브랜치만(Interview Q1 의 선택지 a). | 감사 판정이 덮지 않은 내용을 넣는 길이라서 목록 없이는 관문이 약해져요. 사용자가 고르는 항목이에요. | 목록 검사를 지우는 판(선택지 b). |
| D-3 | 중단된 sync 의 워크트리 따라가기를 위해 `follow::rerun_from` 을 더해요. | 이동 뒤 멈춘 호출을 다시 불러 닫으려면 카드 없는 길이 필요해요. 없으면 워크트리가 뒤처진 채 남아요. | `follow.rs` 변경을 지우고 수동 `git read-tree` 안내로. |
| D-4 | `card drift --restore` 는 sync 로 기록된 기대 팁에서도 카드 없는 `card.tip_restored` 행으로 되돌려요. | sync 도 착지처럼 장부가 아는 이동이라서 같은 되돌림 길을 열어 둬요. | sync 출처의 기대에서는 거절하는 판. |
| D-5 | 출처는 로컬 브랜치 이름만 받고, `origin/main` 같은 원격 추적 이름은 받지 않아요. 목록은 정확히 같은 이름만 맞춰요. | 로컬 브랜치의 팁을 읽는 규칙이 `landed::tip_of` 와 같아요. 원격 추적 이름은 읽는 시점마다 움직여요. | 목록 검사에 원격 이름을 더하는 판. |

## Result

- AC-1 met: `card_land_sync_cli::ac1_a_registered_leader_merges_a_listed_source_in_one_guarded_move`(첫 부모가 옛 tip, 둘째 부모가 출처 tip, 이동은 한 번), `ac1_a_source_already_inside_the_tip_answers_moved_false_and_writes_nothing`, `ac1_a_tip_that_moved_after_the_merge_was_built_is_refused_stale_base`.
- AC-2 met: `ac2_a_lane_session_and_a_call_with_no_session_id_are_refused`, `ac2_a_missing_or_wrong_by_word_is_refused`, `ac2_no_entry_no_leader_mode_or_an_ended_lease_confirms_nobody`.
- AC-3 met: `ac3_an_unlisted_source_and_an_empty_or_missing_list_name_the_key`, `ac3_only_a_local_branch_the_list_holds_exactly_is_a_source`, `ac3_an_unapproved_missing_or_malformed_spec_is_refused`, `ac3_a_conflicting_source_is_refused_as_conflicts`, `ac3_an_unrelated_history_is_refused_as_join_failed`, `ac3_a_drifted_tip_is_refused_as_drift`.
- AC-4 met: `ac4_a_check_that_cannot_start_gives_check_not_started`, `ac4_a_red_check_gives_check_failed`, `ac4_a_replaced_check_program_gives_not_full_pass_even_with_a_pass_line`, `ac4_the_wide_run_of_the_command_itself_lets_the_move_happen`, `ac4_the_private_worktree_is_removed_after_every_outcome_of_the_check`.
- AC-5 met (위험 경로 그 자체): `ac5_the_installed_hook_refuses_every_hand_move_before_and_after_a_sync` 가 `git-hooks-plan`·`git-hooks-apply` 로만 깐 진짜 훅 아래에서 손 이동을 sync 전후에 거절하고 sync 는 한 번 옮기며 워크트리를 따라가요. 곁들여 `ac5_a_tip_lost_at_the_swap_is_refused_stale_base_and_removes_the_intent_file`, `ac5_a_dirty_checkout_of_the_branch_is_refused_before_the_check_and_any_write`.
- AC-6 met: `ac6_the_sync_writes_card_less_rows_that_name_the_merge_the_source_the_old_tip_and_the_spec`, `ac6_the_next_landing_of_an_audited_head_is_not_refused_as_drift`, `ac6_the_same_call_again_answers_already_synced_and_writes_nothing`, `ac6_card_export_reads_a_ledger_that_holds_sync_rows`, `ac6_a_sync_stopped_after_the_move_is_closed_by_the_same_call`, 그리고 lib `card::drift` 10개.
- AC-7 met: lib `profile::` 23개(`sync_sources_refusals_name_the_key` 포함).
- AC-8 met: `leader_route`(22), `lane_team_names`(10), `next_flow_skill_scan`(75), `repo_claude_md`(27), `leader_reclaim_guidance`(1), `source_manifest_facts`(8) 통과, `leader.md` 는 10100바이트(상한 10240).
- AC-9 met: `card_land_cli`·`card_land_batch_cli`·`card_land_checked_out_cli`·`card_land_audited_merge_cli`·`card_land_scans_cli`·`card_guard_cli` 141개 통과, 기대 문구는 그대로예요.
- 가드 뮤턴트: 관문마다 코드를 하나씩 지워 죽는 테스트를 확인했어요(READY 노트의 `guard-mutant:` 줄).
- 남긴 것: `CLAUDE.md` 는 아직 "`card land` 만 통합 브랜치를 옮긴다"고 적어요(다른 카드). `run_check` 의 폴더 이름 `land-batch` 는 그대로예요. 승인 규칙은 `spec_link::status` 를 복사했어요. 설치된 `moai` 가 `sync_sources` 키를 알기 전에는 리더가 프로필에 그 키를 쓰면 안 돼요.
