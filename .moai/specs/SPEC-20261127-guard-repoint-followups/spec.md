---
id: SPEC-20261127-guard-repoint-followups
title: "guard-repoint 후속: card land 가 낡은 기록을 관찰해 알리고, next-flow 가 repoint 를 알려 주고, update 가 링크 워크트리에서도 훅을 보류"
status: approved
revision: 2
created: 2026-11-27
approval:
  by: user
  at: 2026-11-27
  basis: "사용자가 리더 세션의 질문 도구에서 직접 고른 답(결정 노트 10644, 승인 노트 10647): \"승인, 거절은 관찰부터 (Recommended)\" (D-1 은 b, D-2~D-7 은 a); 이 레인이 기록 직전에 커밋 03411f456a 의 파일에서 moai spec show 로 approval_digest 를 다시 읽어 노트 10647 의 값과 같음을 확인했어요(선택은 리더가 옮긴 글이고 이 레인이 직접 보지는 않았어요)"
  digest: "sha256:904d69c6a26213187d7859053f4113b495bb02816fd034cd0bfba1ca9ca4f4ff"
---

# guard-repoint 후속: card land 가 낡은 기록을 관찰해 알리고, next-flow 가 repoint 를 알려 주고, update 가 링크 워크트리에서도 훅을 보류

## Intent

`SPEC-20261126-guard-record-repoint`(랜딩 완료, 배치 24)는 프로필의 통합 브랜치가 바뀐 뒤 낡은 가드 기록을 `moai card guard-repoint` 로 다시 맞추는 길을 줬어요. 하지만 그 SPEC 의 첫 판이 일부러 비워 둔 것이 셋 남았어요. 이 SPEC 이 그 셋을 채워요.

1. `card land` 는 기록이 프로필과 달라도 기록이 가리키는 브랜치를 그대로 옮겨요(기록이 이긴다는 규칙이에요). mo.ai.kr 에서 `main` 이 옮겨졌어요(이 SPEC 을 부른 카드 본문의 글이고 작성자가 mo.ai.kr 을 직접 읽지는 않았어요, 미확인). 지금은 어긋남을 답의 `profile_disagrees` 로만 알려서 아무도 멈추지 못했어요. 사용자의 선택(D-1 b, 개정 2)에 따라 RC1 에서는 거절하지 않고, 어긋남을 `guard_record_stale` 관찰 필드로 답에 싣고 `guard-repoint` 를 이름으로 알려요. 거절 여부는 다음 릴리즈의 질문으로 남겨요.
2. `skills/moai/references/flows/next-flow.md` 는 낡은 기록을 맞추는 길(`guard-repoint`)을 어디에도 적지 않았어요.
3. 링크 워크트리에서 `moai update` 는 기록이 낡아도 가드 훅을 제안해요. 이유는 `repoint::stale()` 가 링크 워크트리에서 늘 `None` 이라서예요(앞 SPEC 리뷰의 논블로킹 지적, `src/update.rs:2190`).

읽은 사실(NeXT `8933f238cd` 의 이 트리, 코드 읽기만 했고 이 SPEC 을 쓰며 실행한 것은 없어요):

- 브랜치는 `src/card/land.rs:421` 의 `landed::integration_branch` 가 정해요. `src/card/landed.rs:124-129` 는 기록을 먼저, 없을 때만 프로필을 읽어요. 어긋남(`guard::protected`)은 `land.rs:570` 에서 모든 확인이 끝난 뒤 불리고 답의 `profile_disagrees`(`land.rs:686`)에만 실려요.
- `land-batch` 는 `src/card/land/batch.rs:461` 에서 같은 함수로 브랜치를 정하고, `land-sync` 는 `src/card/land/sync.rs:424` 에서 따로 정한 뒤 `landed::move_branch`(`sync.rs:534`)로 브랜치를 옮겨요.
- `repoint::stale()`(`src/card/guard/repoint.rs:67-75`)은 주 워크트리에서만 `Some` 이에요. `probe` 가 기록을 읽기 전에 링크 워크트리를 `Linked` 로 돌려보내요(`repoint.rs:126-130`). `guard_answer`(`src/update.rs:2190-2192`)는 `can_act(cwd) && stale(cwd).is_none()` 이라 링크 워크트리에서는 `can_act` 만 남아요.
- 링크 워크트리의 훅은 그 워크트리 자신의 git 폴더(`<공통>/worktrees/<이름>/hooks`)에 설치돼요(`src/git_hooks/mod.rs:98-120`). `core.hooksPath` 가 없으면 git 이 거기서 훅을 돌리지 않으니, 그 제안은 효과 없는 훅을 계획하는 일이에요(승인된 앞 SPEC 이 `pre_existing` 으로 둔 동작이에요).
- 기존 테스트 `card_land_cli::ac22_the_answer_reports_a_variant_and_a_disagreeing_profile`(`tests/card/card_land_cli.rs:1441-1479`)는 어긋난 프로필에서도 기록의 브랜치가 옮겨지고 답의 `profile_disagrees` 가 프로필 브랜치를 담는 것을 단언해요. 앞 SPEC 의 REQ-11 과 AC-8 이 그 테스트를 "수정 없이 통과"로 묶어 뒀고, 같은 SPEC 의 알려진 빈틈 2 가 "기록이 프로필과 다를 때 `card land` 가 거절하기"를 SPEC 개정과 사용자 승인이 필요한 일로 남겼어요. 사용자가 거절은 관찰부터 하자고 골라서(D-1 b) 이 SPEC 은 그 테스트를 바꾸지 않아요.
- `next-flow.md` 는 8111 바이트이고 세 테스트가 크기를 묶어요: `tests/next_flow_skill_scan.rs:2416` 의 `FLOW_REF_TIP_BYTES = 8121`, 같은 파일 `:91` 의 `FLOW_REF_LIMIT = 8192`, `tests/lane_team_names.rs:279-291` 의 8192.

이 SPEC 은 게이트 동작을 바꾸는 일이라 tier 2 예요. 승인 전에는 코드를 쓰지 않아요.

## Interview

- Q1: 이 SPEC 을 승인할까요, 거절은 관찰부터 할까요? → A: 사용자가 직접 골랐어요(2026-11-27, 리더 세션의 질문 도구, 결정 노트 10644): "승인, 거절은 관찰부터 (Recommended)". 그 선택지의 설명은 "D-2~D-7은 기본안(a)으로 승인하고 D-1만 b로 해요. RC1에서는 어긋남을 기록만 하고(거절 횟수와 오탐을 Result에 적음) 실제 거절은 다음 릴리즈에서 다시 물어요. 11-20 규칙과 C-321 결정에 맞아요. 단, RC1에서는 낡은 기록으로 엉뚱한 브랜치가 옮겨지는 것을 막지 못해요."예요.
- Q2: 아래 Decisions 의 열린 갈림길(D-1 ~ D-7)은 어떻게 확정할까요? → A: D-1 은 b(관찰부터), D-2 ~ D-7 은 기본값 a 로 같은 선택에서 확정했어요(결정 노트 10644). 개정 2 는 이 선택을 글에 옮긴 판이고, 개정 1(바로 거절)의 다른 부분은 바꾸지 않았어요.

## Non-goals

- 기록을 다시 쓰는 일. 기록을 바꾸는 길은 계속 `moai card guard-repoint`(미리보기와 사용자 선택) 하나뿐이에요. `card land` 는 기록을 한 번도 쓰지 않고 답에 관찰 필드를 더하기만 해요.
- RC1 에서 land 계열이 어긋난 기록 앞에서 거절하는 일(사용자가 관찰부터 하자고 골랐어요, D-1 b). 거절로 바꿀지는 다음 릴리즈의 질문이고 이 SPEC 의 Result 에 남겨요.
- `repoint::stale()` 의 판정과 그 소비자(`repoint_notice`, `probe`)의 변경. 링크 워크트리에서 `None` 인 것은 앞 SPEC 의 REQ-10 이에요.
- 가드 훅의 텍스트와 판정 규칙(`SPEC-20261121-ref-guard`), 세션 시작 줄(`src/hook_session_start.rs`), 프로젝트 설정 쪽(`/moai project`).
- 링크 워크트리의 훅이 그 워크트리 자신의 git 폴더에 설치되는 기존 동작.
- `skills/moai/workflows/**` 와 `docs/leader-rules.md`, `CLAUDE.md` 의 문장. `stale_base` 를 이름으로 부르는 다른 문서에 `guard_record_stale` 을 더하는 일은 이번에 하지 않아요(next-flow.md 한 곳과 연산 문서만 해요).
- 이 저장소의 실제 `.git/moai`, `~/.moai`, 실제 mo.ai.kr 상태를 쓰는 일. 코드와 테스트는 임시 HOME 과 임시 저장소에서만 해요.

## Requirements

- REQ-1: When `moai card land` has resolved the integration branch and the guard record names a ref that differs from the ref of the integration branch that the profile of the folder it runs in names, the card command shall land exactly as before, adding the field `guard_record_stale` to its answer with the record, the profile branch, the worktree kind (`primary` or `linked`), `would_refuse: true` and the command `moai card guard-repoint`, and writing no ledger row and no receipt beyond what a landing writes today · failure: none — a repository with no record, an unreadable record, a record equal to the profile, or a profile that names no integration branch adds no field and lands exactly as before this SPEC.
- REQ-2 (REQ-1): IF the observation is made in a linked worktree, THEN the field shall also carry the sentence that `guard-repoint` refuses in a linked worktree and has to be run from the primary checkout.
- REQ-3: When `moai card land-batch` has resolved the integration branch and the same difference exists, the card command shall land the batch exactly as before and add the same field once to the top of its answer · failure: none — as REQ-1, the no-difference cases run as before.
- REQ-4: When `moai card land-sync` is about to move the integration branch and the same difference exists, the card command shall move it exactly as before and add the same field to its answer · failure: none — as REQ-1.
- REQ-5: When `moai update` plans or applies the git-hook layer in any worktree, primary or linked, whose guard record differs from the profile's integration branch as in REQ-1, the update shall withhold the guard hook, giving the plan and both apply sites the same answer so the tokens match · failure: none — a repository with no record or an equal record offers the hook as before.
- REQ-6: When a caller needs to know whether the record and the profile differ, the repoint module shall give it one read-only function that works in a primary and in a linked worktree · failure: none — an absent, unreadable or equal record, and a profile that names no integration branch, answer no difference.
- REQ-9: When `card land`, `land-batch`, `land-sync` or the update's guard answer needs that difference, it shall call the one function of REQ-6, and `repoint::stale()`, `probe` and `repoint_notice` keep their answers · failure: none — the existing guard-repoint tests pass unchanged.
- REQ-7: When a reader opens `skills/moai/references/flows/next-flow.md`, the "Landing" section shall tell when to run `moai card guard-repoint` (a landing answer that carries `guard_record_stale`, which RC1 only reports and never refuses on, or the notice of `moai update`), how (the preview, then `--apply` with the preview's `--plan-digest` and the `--permission-reference` of the user's yes, `--revert <stamp>`), that nobody but the leader with the user's yes applies it, and that the old branch is unguarded afterwards · failure: none — the file stays inside its byte limits (D-5).
- REQ-8: When the operations documents describe `card land`, `card land-batch`, `card land-sync` and the update's guard offer, they shall name the `guard_record_stale` answer field and that a linked worktree withholds the guard hook for a stale record · failure: none — the three documents that already pin `guard-repoint` keep naming it.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary repository with the record `refs/heads/integ` and a profile whose integration branch is another existing branch, `card land` of a passed card lands as before (`integ` moves, the other branch stays) and its answer carries `guard_record_stale` with the record, the profile branch, `would_refuse: true` and the worktree kind, and the same run from a linked worktree adds the primary-checkout sentence; a record equal to the profile, no record and a profile with no branch add no field and land as before · verify: `cargo test --locked --offline --test card -- guard_record_stale_land`
- AC-2 (REQ-1): the test that pins the move and `profile_disagrees` (`ac22_the_answer_reports_a_variant_and_a_disagreeing_profile`) and the tests `ac9_`, `ac12_` and `ac13_` of the same target pass unchanged · verify: `cargo test --locked --offline --test card -- card_land_cli::ac22_`
- AC-3 (REQ-3): `card land-batch` with two audited heads and a stale record lands both as before, drops no head and carries `guard_record_stale` once at the top of its answer · verify: `cargo test --locked --offline --test card -- guard_record_stale_batch`
- AC-4 (REQ-4): `card land-sync` with a stale record moves the branch as before and carries `guard_record_stale` in its answer · verify: `cargo test --locked --offline --test card -- guard_record_stale_sync`
- AC-5 (REQ-5): `moai update --json` in a temporary repository with a stale record offers no guard hook in a primary worktree and in a linked worktree, the plan token and the apply token match in both, and the existing stale-record, notice and offer tests pass unchanged · verify: `cargo test --locked --offline --test update_git_hooks_cli stale_record`
- AC-6 (REQ-6, REQ-9): the new function answers the record and the profile branch for a stale record in a primary and in a linked worktree, and nothing for no record, an equal record, an unreadable record and a profile with no branch; `stale()` still answers `None` in a linked worktree · verify: `cargo test --locked --offline --test integration_guard_repoint disagrees_`
- AC-7 (REQ-7): `next-flow.md` names `guard-repoint`, the `guard_record_stale` answer field, the preview, `--apply` with `--plan-digest` and `--permission-reference`, `--revert` and the unguarded old branch inside the "Landing" section, and the three size tests and the existing scans pass · verify: `cargo test --locked --offline --test next_flow_skill_scan guard_repoint`
- AC-8 (REQ-7): the byte limits hold: the file is at most 8192 bytes and the tip ratchet equals the new size · verify: `cargo test --locked --offline --test next_flow_skill_scan land_scans_the_flow_names_the_scan_tests_and_the_pass_receipt_step`
- AC-9 (REQ-8): the operations documents name the `guard_record_stale` answer field for the three landing commands and the linked-worktree withholding, and `ac7_` of the guard test still passes · verify: `cargo test --locked --offline --test integration_guard_repoint ac7_`

## Scope

- write: `.moai/specs/SPEC-20261127-guard-repoint-followups/spec.md`
- write: `src/card/guard/repoint.rs` (한 함수와 설명), `src/card/land.rs`, `src/card/land/batch.rs`, `src/card/land/sync.rs`, `src/update.rs`
- write: `tests/card/card_land_cli.rs`, `tests/card/card_land_batch_cli.rs`, `tests/card/card_land_sync_cli.rs`, `tests/update_git_hooks_cli.rs`, `tests/integration_guard_repoint.rs`, `tests/next_flow_skill_scan.rs`
- write: `skills/moai/references/flows/next-flow.md`, `skills/moai/references/source-manifest.json` (그 파일의 해시 한 줄만)
- write: `docs/design-review/appendix/operations/card.md` (공유 파일, 줄 임대로 land, land-batch, land-sync 행), `docs/design-review/appendix/operations/git_hooks_plan.md`
- read-only: `src/card/landed.rs`, `src/card/guard.rs`, `src/git_hooks/mod.rs`, `src/contract.rs`, `tests/lane_team_names.rs`, `tests/integration_guard_install.rs`, `tests/card/card_guard_cli.rs`
- forbidden: 이 저장소의 `.git/hooks`·`.git/moai`, `.moai/project.yaml`, `CLAUDE.md`, `install/git-hooks/**`, `src/hook_session_start.rs`, `src/card/landed.rs`(브랜치를 정하는 함수는 바꾸지 않아요), `skills/moai/workflows/**`, `skills/moai/references/internal/**`, `tests/lane_team_names.rs`(한도 8192 는 그대로예요)

## Plan

1. (지금, SPEC 폴더만) 이 SPEC 을 채우고 `moai spec check` 를 돌려요. SPEC-READY 노트에 digest 와 열린 선택을 올리고 `LANE-BLOCKED kind=decision` 으로 사용자의 선택을 기다려요. 승인은 사용자의 것이라 이 레인이 기록하지 않아요.
2. 승인 뒤 `moai card scope` 로 위 쓰기 범위를 더해요. 공유 파일(`card.md`, 크기 한도가 걸린 `next-flow.md`)은 `card line-lease` 로 쓰고 커밋 직후 돌려줘요. 겹치는 열린 브랜치가 `src/update.rs`, `src/card/land.rs` 를 건드리는지 `card take` 가 거절하는지로 확인하고, 거절되면 그 카드에 물어요.
3. 테스트 먼저(RED, 슬롯 `--max 2`): AC-1~AC-6 과 AC-7 의 새 테스트를 추가해요(`ac22_` 는 바꾸지 않아요). 변경 전 트리에서 실패하는 것을 읽어요. AC-1 의 링크 워크트리와 AC-5 의 위험 경로는 임시 HOME·임시 저장소·실제 바이너리를 쓰는 Rust 테스트예요. 셸 스크립트는 쓰지 않아요.
4. `src/card/guard/repoint.rs` 에 `disagrees(project) -> Option<Staleness>` 를 더해요: `common_dir_of(project)` 의 기록(`read_record`)과 `landed::profile_branch(project)` 를 `landed::branch_ref` 로 비교하고, 둘이 모두 있고 다를 때만 `Some` 이에요. 주 워크트리 검사는 하지 않아요. 쓰기는 없어요.
5. `src/card/land.rs` 는 `:421-427` 의 브랜치 결정 뒤에서 `disagrees` 를 불러, 어긋나면 답의 `guard_record_stale` 값(기록, 프로필 브랜치, 워크트리 종류, `would_refuse: true`, 명령 이름, 링크면 REQ-2 의 문장)을 만들어 둬요. 브랜치를 옮기는 코드와 거절 경로(`stale_base:`, `drift:`, `contract.rs`)는 건드리지 않고, 답을 만드는 자리(`land.rs:682-702`)에서 그 값을 더해요. `batch.rs:461` 과 `sync.rs:424` 도 같은 함수로 같은 값을 만들어 답에 더해요(값은 한 곳에서 만들어 셋이 나눠 써요). `profile_disagrees` 답 필드는 그대로 둬요(D-6).
6. `src/update.rs:2190` 의 `guard_answer` 를 `can_act(cwd) && disagrees(cwd).is_none()` 으로 바꿔요. 계획(`:2081`)과 두 적용(`:2658`, `:2804`)이 모두 이 한 곳을 지나서 토큰이 같아요. `repoint_notice` 는 그대로예요.
7. `next-flow.md` "Landing" 에 REQ-7 의 문단을 더해요. 크기는 D-5 의 기본값(기존 문장을 줄여 8192 한도를 지키고 팁 래칫만 새 크기로 올려요). 줄이는 문장은 `next_flow_skill_scan`, `lane_completion_handshake`, `leader_route`, `followup_per_spec_rules`, `batch.rs` 의 테스트가 고정한 구절과 `push`·`fast-forward` 를 함께 쓰지 않는 규칙을 건드리지 않게 골라요. 그다음 `shasum -a 256 skills/moai/references/flows/next-flow.md` 로 `source-manifest.json` 의 해시를 맞춰요(파이썬 없이).
8. `card.md` 의 land 계열 행과 `git_hooks_plan.md` 한 문장(링크 워크트리에서도 낡은 기록이면 훅을 보류해요). `ac7_` 가 읽는 문자열은 지켜요.
9. 체크 결과 전 준비: `cargo fmt --all`, 대상 clippy, `moai check affected --card`(한 번, 감싸지 않고), 새 테스트와 스킬 텍스트 변경 때문에 `next_flow_skill_scan`, `source_manifest_facts`, `repo_scans_table`, `lane_team_names`, `lane_completion_handshake`, `leader_route`, `--bin moai-cli` 를 nextest 로 확인해요. 가드 뮤턴트는 새 가드 테스트마다 한 줄씩 Result 에 적어요(편집 한 번, 슬롯 안 nextest, 평범한 `git restore` 한 줄): AC-1(링크 워크트리 문장 빼기), AC-1(필드를 더하지 않음), AC-3(batch 답에서 필드 빠짐), AC-4(sync 답에서 필드 빠짐), AC-5(`guard_answer` 가 다시 `stale()` 를 씀).
10. 랜딩 뒤 리더가 `moai update` 로 사용자 복사본(`~/.claude/skills/moai`)을 맞추고 `diff -rq skills/moai ~/.claude/skills/moai` 를 읽어요(`CLAUDE.local.md` 의 순서, 이 카드의 일이 아니에요).
11. 관찰 결과: 랜딩을 돌리는 리더가 `guard_record_stale` 필드가 실린 랜딩 답의 건수와 오탐(필드가 있었는데 기록이 낡지 않았던 경우)을 카드 노트로 알리고, 이 SPEC 의 Result 가 그 건수와 "거절로 바꿀지"를 다음 릴리즈의 질문으로 적어요(D-1 b).

## Risks

- RC1 에서는 낡은 기록으로 엉뚱한 브랜치가 옮겨지는 것을 막지 못해요(사용자가 이 위험을 알고 관찰부터를 골랐어요, 결정 노트 10644). 관찰 필드가 답에 실려도 읽는 사람이 없으면 같은 일이 되풀이돼요. 그래서 next-flow.md 문단(REQ-7)과 연산 문서가 필드를 읽는 법과 `guard-repoint` 를 알려요.
- 링크 워크트리는 자기 체크아웃의 프로필을 읽어요. 옛 커밋에서 만든 링크 워크트리에서 `card land` 를 돌리면 기록은 맞는데 프로필이 달라 보여 필드가 붙어요(오탐). 거절하지 않으므로 랜딩은 막히지 않고, 필드가 두 값을 모두 적어 사람이 읽을 수 있어요. 오탐은 리더가 알리는 건수에서 가려내 Result 에 적어요(D-1 b).
- 앞 SPEC 의 REQ-11("`card land` 는 기존 기록을 다시 쓰지 않아요")과 `ac22_` 는 그대로 지켜져요(관찰 필드만 더해요).
- `next-flow.md` 의 크기 한도 셋(8121 팁 래칫, 8192 둘)과 많은 문구 고정 테스트 때문에 기존 문장을 줄이는 일이 어려울 수 있어요. 줄이지 못하면 이 레인은 `LANE-BLOCKED kind=decision` 으로 멈추고 D-5 의 다른 선택지를 사용자에게 물어요(한도를 임의로 올리지 않아요).
- `land-sync` 는 이미 손으로 옮겨진 랜딩을 원장에 맞추는 명령인데 `move_branch` 도 불러요(`sync.rs:534`). 관찰 필드는 그 답에도 실려서 sync 도 같은 안내를 받아요(D-2).
- `update` 의 링크 워크트리 보류는 조용해요(안내는 주 워크트리에만 있어요, 앞 SPEC 의 REQ-10). 링크 워크트리 사용자는 왜 훅 제안이 없는지 모를 수 있어요(D-7).
- mo.ai.kr 의 실제 상태는 카드 본문의 글뿐이에요(미확인). 테스트는 임시 HOME·임시 저장소에서만 재현해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기록이 프로필과 다르면 `card land` 계열은 RC1 에서 거절하지 않고 답의 `guard_record_stale` 필드로 기록만 해요(관찰부터). 선택지 a: 바로 거절해요(`guard_record_stale:` 로 아무것도 옮기지 않아요). 선택: b. | 사용자가 직접 골랐어요(2026-11-27, 결정 노트 10644): "승인, 거절은 관찰부터 (Recommended)". 거절 횟수와 오탐은 Result 에 적고 거절로 바꿀지는 다음 릴리즈의 질문이에요. RC1 에서는 낡은 기록으로 엉뚱한 브랜치가 옮겨지는 것을 막지 못해요(사용자가 알고 골랐어요). | 필드를 더하는 코드와 그 테스트를 지우는 판. 거절로 바꾸는 판은 새 개정이에요. |
| D-2 | 관찰 필드는 브랜치를 옮기는 세 명령의 답에 실어요: `card land`, `land-batch`, `land-sync`. 선택지 b: `card land` 만. 기본값: a(사용자 확정). | 세 명령 모두 `landed::move_branch` 로 통합 브랜치를 옮겨요(`land.rs`, `batch.rs`, `sync.rs:534`). `land` 만 막으면 sync 와 batch 가 같은 일을 해요. | 세 호출 중 하나를 지우는 판. |
| D-3 | 링크 워크트리에서도 그 워크트리의 프로필로 비교하고 관찰 필드에 실어요(문장이 주 체크아웃에서 `guard-repoint` 를 돌리라고 말해요). 선택지 b: 주 워크트리에서만 싣고 링크에서는 지금처럼 필드가 없어요. 기본값: a(사용자 확정). | 링크 워크트리는 리더의 랜딩이 병합 팁에서 돌리는 곳이기도 하고, 거기서만 새는 틈은 앞 SPEC 이 이미 `update` 에서 겪었어요. 오탐은 Risks 첫 줄이에요. | 비교를 주 워크트리로 제한하는 판. |
| D-4 | 비교는 `repoint.rs` 의 새 읽기 전용 함수 하나(`disagrees`)로 하고 `land`, `land-batch`, `land-sync`, `update` 가 모두 써요. `stale()` 는 바꾸지 않아요. 선택지 b: 호출마다 각자 비교해요. 기본값: a. | 같은 판정을 네 곳에 흩으면 어긋나요. `stale()` 를 바꾸면 앞 SPEC 의 REQ-10 테스트가 깨져요. | 함수를 지우고 호출을 되돌리는 판. |
| D-5 | `next-flow.md` 에 문단을 더하고 기존 문장을 줄여 8192 한도 둘은 그대로 지키며, 팁 래칫(`FLOW_REF_TIP_BYTES`)만 새 크기로 올려요. 줄이지 못하면 사용자에게 물어요. 선택지 b: 세 한도를 모두 올려요. c: 문단을 `landing-release.md` 에 두고 `next-flow.md` 에는 한 줄 포인터만 둬요(기준 2 의 글자는 `next-flow.md` 를 가리켜서 사용자의 확인이 필요해요). 기본값: a. | 한도는 스킬 문서를 짧게 지키려는 장치예요. 임의로 올리지 않고 줄여 보는 쪽이 한도를 약하게 하지 않아요. | 문단과 래칫 값을 되돌리는 판. |
| D-6 | `card land` 답의 `profile_disagrees` 필드는 지금 값 그대로 두어요(`guard_record_stale` 필드는 그 옆에 더해요). 선택지 b: 필드를 지워요. 기본값: a(사용자 확정). | 답의 모양을 바꾸지 않아 읽는 쪽(`ac22_`, 리더 도구)이 안 깨져요. | 필드를 지우는 판. |
| D-7 | 링크 워크트리에서 `moai update` 의 훅 보류는 조용히 해요(안내는 주 워크트리에만 있어요). 선택지 b: 링크 워크트리에도 안내 한 줄을 더해요(앞 SPEC 의 REQ-10 을 바꾸는 개정). 기본값: a. | 앞 SPEC 의 REQ-10 과 그 테스트를 건드리지 않아요. | 해당 없음(후속 카드). |

## Result

구현 레인(moai-adk-agent-jade)이 2026-11-27T04:05Z(date -u)에 채웠어요. macOS(개발 Mac), 빌드 폴더는 이 워크트리의 `target/`, 무거운 명령은 모두 `moai heavy-slot-acquire --name moai-adk-agent-jade --max 2 --` 슬롯 안에서 돌렸어요. 승인은 레저 노트 10644·10647(사용자가 리더 세션에서 직접 고른 답 "승인, 거절은 관찰부터 (Recommended)")이고, 이 파일 머리글에 승인 블록이 있어요. 기준은 NeXT `8933f238cd` 예요.

변경: `src/card/guard/repoint.rs`(`disagrees`, `observation`, 비교를 `compare` 로 분리, `stale()`·`probe`·`repoint_notice` 는 그대로), `src/card/land.rs`·`src/card/land/batch.rs`·`src/card/land/sync.rs`(답에 `guard_record_stale` 한 필드), `src/update.rs`(`guard_answer` 가 `disagrees` 를 씀), `skills/moai/references/flows/next-flow.md`(Landing 에 문단 하나, 다른 줄을 줄여 8180 바이트), `skills/moai/references/source-manifest.json`(해시 한 줄, 줄 임대), 문서 둘(`card.md` 줄 임대, `git_hooks_plan.md`), 테스트 다섯 파일과 `tests/next_flow_skill_scan.rs`(새 시험 하나, 팁 래칫 8121 → 8180).

테스트 (이 레인이 슬롯에서 직접 돌린 결과예요):

| 명령 | 결과 |
|---|---|
| `cargo nextest run --test card -- guard_record_stale` | 7 통과 (RED 때는 4 실패 + 가드 3 통과) |
| `cargo nextest run --test card -- card_land` | 176 통과, 0 실패 (`ac22_`, `ac9_`, `ac13_` 포함, 수정 없음) |
| `cargo nextest run --test integration_guard_repoint` | 33 통과 (`disagrees_` 7 포함, `ac7_` 수정 없음) |
| `cargo nextest run --test update_git_hooks_cli` | 17 통과 (새 링크 워크트리 시험 포함) |
| `cargo nextest run` 대상: `next_flow_skill_scan`, `source_manifest_facts`, `repo_scans_table`, `lane_team_names`, `lane_completion_handshake`, `leader_route`, `repo_claude_md`, `integration_guard_repoint` | 224 통과, 0 실패 |
| `cargo nextest run --test card -- followup_per_spec` | 2 통과 |

RED 는 `moai tdd red` 로 기록했어요(커밋 `e9233b60ba`, 체크 둘 모두 assertion 실패 exit 101). `disagrees_` 시험은 함수가 없어 컴파일 실패였고 RED 체크에 넣을 수 없어서(컴파일 실패는 RED 가 아니에요) 두 체크에서 뺐어요; 함수가 생긴 뒤 7개 모두 통과했어요. `moai tdd green` 은 두 체크를 통과로 재생했어요. 새 `next_flow_skill_scan` 시험(`guard_repoint_is_told_in_the_landing_section`)은 문서 수정과 같은 라운드에 써서 문서 수정 전에 실패하는 것을 따로 돌려 보지 않았어요(not_run).

- AC-1: met
- AC-2: met
- AC-3: met
- AC-4: met
- AC-5: met
- AC-6: met
- AC-7: met
- AC-8: met
- AC-9: met

가드 뮤턴트 (편집 한 번, 슬롯 안 nextest, 평범한 `git restore`, 매번 트리가 깨끗한 것까지 읽었어요; 배치·싱크·업데이트 셋은 한 번에 편집해 한 번 빌드하고 시험은 각각의 결과를 읽었어요):

guard-mutant: guard_record_stale_land_in_the_primary_worktree_observes_and_lands_as_before | land 가 `guard_record_stale` 를 답에 더하지 않음 | exit 101
guard-mutant: guard_record_stale_land_in_a_linked_worktree_adds_the_primary_checkout_sentence | `observation` 이 링크 워크트리 문장(`note`)을 빼 버림 | exit 100
guard-mutant: guard_record_stale_batch_lands_both_heads_and_reports_the_record_once | land-batch 가 필드를 더하지 않음 | exit 100
guard-mutant: guard_record_stale_sync_moves_the_recorded_branch_and_reports_the_record | land-sync 가 필드를 더하지 않음 | exit 100
guard-mutant: update_offers_the_guard_hook_not_for_a_stale_record_in_a_linked_worktree | `guard_answer` 가 다시 `stale()`(주 워크트리만)를 씀 | exit 100
guard-mutant: none | `disagrees_*` 일곱 시험은 뮤턴트를 돌리지 않았어요(not_run) | 해당 없음

(exit 100 은 nextest 의 시험 실패 종료 코드이고 101 은 `cargo test` 의 것이에요. 둘 다 시험이 단언에서 실패했어요.)

next-flow.md 한도 (D-5 a): 8111 → 8153 바이트(한도 8192 두 곳은 그대로, 팁 래칫만 8153 으로). 문단을 넣으려고 줄인 곳은 아래 다섯 가지이고 테스트가 고정하지 않는 문장만이에요(고정된 문장은 `GAP`, 원격 복사 줄의 `remote_copy` 구절, 감사 브리프 줄, 재시작 줄의 앞부분, 그리고 `src/card/land/batch.rs` 의 단위 테스트가 읽는 `every five landings` 라 그대로 뒀어요): 새 창 문장 한 줄, "스스로 세션을 비우는 것은 보장이 아니다" 문장, "백그라운드 세션은 푸시도 풀 리퀘스트도 하지 않는다" 문장(같은 줄 앞부분이 이미 푸시와 풀 리퀘스트 없음을 말해요), 원격 복사 줄의 두 꼬리(`a key of the project profile` 과 끝 절), ACK 줄과 재시작 줄의 단어, 넓은 체크 줄의 문장 모양(`the full run comes every five landings`, 뜻은 그대로).

첫 `check affected --card`(헤드 `5841f7e9b7`)는 FAIL 이었어요: `card::land::batch::tests::the_verify_section_and_the_flow_reference_name_the_new_order` 하나가 `every five landings` 를 못 찾았어요(`src/card/land/batch.rs:661`, 이 테스트가 두 문서를 `include_str!` 로 읽어요). 제 고정 문장 검색이 `src/` 를 빼고 `tests/` 만 봤던 탓이에요. 문장을 되살리고 위 백그라운드 문장을 줄여 고쳤어요; 이 줄 아래의 결과는 새 헤드로 다시 돌린 것이에요.

알려진 빈틈과 사용자에게 남는 질문:

- 기준 1(카드 본문의 글: "카드 land 가 낡은 기록이면 거절")은 이 SPEC 이 RC1 에서 관찰만 하므로 충족되지 않았어요(not met). 리더 결정(노트 10647)대로 글은 고치지 않았고, `card land --keep-open` 으로 올려야 해요.
- Plan 11, 관찰 결과: 이 레인은 실제 랜딩을 보지 못해요. 리더가 `guard_record_stale` 필드가 실린 랜딩 답의 건수와 오탐(필드가 있었는데 기록이 낡지 않았던 경우)을 카드 노트로 알리면 여기에 적어요. 관찰 건수: 없음(not_run, 리더 입력 대기). 다음 릴리즈의 질문: "관찰한 건수와 오탐이 이래요. 관찰을 거절로 바꿀까요?"
- 링크 워크트리에서 `moai update` 는 낡은 기록이면 훅을 조용히 보류해요(D-7 a, 안내 없음).
- `src/update.rs` 를 다른 카드 둘(같은 holder, 번호는 카드 노트에 적었어요)이 쥐고 있어요(`card scope` 의 겹침 답). 이 카드는 `guard_answer` 한 줄만 바꿨고, 병합은 리더의 몫이에요.

서브 에이전트 기록: 두 에이전트의 기록에서 도구 거절은 없었어요(오류 결과는 `cargo fmt --check` 와 실패하는 시험의 종료 코드뿐이에요). 첫 에이전트는 자기 워크트리의 `target/debug` 를 `rm -rf`(상대 경로)로 지워서 다음 빌드가 느려졌어요; 둘째에게는 지우지 말라고 했고 지우지 않았어요.
